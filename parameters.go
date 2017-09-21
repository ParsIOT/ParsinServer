// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// parameters.go contains structures and functions for setting and getting Naive-Bayes parameters.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
)

// PersistentParameters are not reloaded each time
type PersistentParameters struct {
	NetworkRenamed map[string][]string // key:networkName, value:mac list; e.g.: {"1":["mac1","mac2"]}
}

// PriorParameters contains the network-specific bayesian priors and Mac frequency, as well as special variables
type PriorParameters struct {
	P map[string]map[string][]float32 // probability of each mac's rssi for each location;e.g.:P["P1"]["MAC1"][-50] = 0.1
	// this probability value is made from the PdfType(priors.go) values
	// P is equals to probability distribution(gaussian)
	NP map[string]map[string][]float32 // sum of probability of each mac's rssi in every locations except for an specific location
	//NP["P1"]["MAC1"][-50] = SUM(P[Pi]["MAC1"][-50]);i!=P1
	MacFreq  map[string]map[string]float32 // Frequency of a mac in a certain location(macCountByLoc/max of macCountByLoc of a specific mac for every location)
	NMacFreq map[string]map[string]float32 // Frequency of a mac, in everywhere BUT a certain location
	Special  map[string]float64            //a map with keys:mixin,variabilityCutoff,macFreqMin,NmacFreqMin
}

// ResultsParameters contains the information about the accuracy from crossValidation
type ResultsParameters struct {
	Accuracy         map[string]int            // accuracy measurement for a given location
	TotalLocations   map[string]int            // number of locations
	CorrectLocations map[string]int            // number of times guessed correctly
	Guess            map[string]map[string]int // correct(real location) -> guess -> times
}

// FullParameters is the full parameter set for a given group
type FullParameters struct {
	NetworkMacs    map[string]map[string]bool   // map of networks and then the associated macs in each
	NetworkLocs    map[string]map[string]bool   // map of the networks, and then the associated locations in each
	MacVariability map[string]float32           // variability of macs
	MacCount       map[string]int               // number of fingerprints of a AP in all data, regardless of the location; e.g. 10 of AP1, 12 of AP2, ...
	MacCountByLoc  map[string]map[string]int    // number of fingerprints of a AP in a location; e.g. in location A, 10 of AP1, 12 of AP2, ...
	UniqueLocs     []string                     // a list of all unique locations e.g. {P1,P2,P3}
	UniqueMacs     []string                     // a list of all unique APs
	Priors         map[string]PriorParameters   // generate priors for each network
	Results        map[string]ResultsParameters // generate results for each network
	Loaded         bool                         // flag to determine if parameters have been loaded
}

// NewFullParameters generates a blank FullParameters
func NewFullParameters() *FullParameters {
	return &FullParameters{
		//todo: networkMacs difference with UniqueMacs
		//todo: NetworkLocs difference with UniqueLocs
		//todo: in networkMacs and networkLocs what is the purpose of true values? Could it be false?
		NetworkMacs:    make(map[string]map[string]bool), //e.g.: {"0":["MAC1":true,"MAC2":true,...]}
		NetworkLocs:    make(map[string]map[string]bool), //e.g.: {"0":["P1":true,"P2":true,...]}
		MacCount:       make(map[string]int),             //number of fingerprints of an AP(mac) in all locations; e.g. : {"MAC1":10,"Mac2":15,...}
		MacCountByLoc:  make(map[string]map[string]int),  //e.g.: {"P1":{"MAC1":10,"MAC2":14},"P2":{MacCount2},}
		UniqueMacs:     []string{},                       //UniqueMacs is an array of AP's macs
		UniqueLocs:     []string{},                       //UniqueLocs is an array of map's locations e.g.: ["P1","P2","P3",...]
		Priors:         make(map[string]PriorParameters),
		MacVariability: make(map[string]float32), //the standard deviation of rssi of each mac
		Results:        make(map[string]ResultsParameters),
		Loaded:         false, //is true if ps was created and save in resources
	}
}

// NewPriorParameters generates a blank PriorParameters
func NewPriorParameters() *PriorParameters {
	return &PriorParameters{
		P:        make(map[string]map[string][]float32),
		NP:       make(map[string]map[string][]float32),
		MacFreq:  make(map[string]map[string]float32),
		NMacFreq: make(map[string]map[string]float32),
		Special:  make(map[string]float64),
	}
}

// NewResultsParameters generates a blank ResultsParameters
func NewResultsParameters() *ResultsParameters {
	return &ResultsParameters{
		Accuracy:         make(map[string]int),
		TotalLocations:   make(map[string]int),
		CorrectLocations: make(map[string]int),
		Guess:            make(map[string]map[string]int),
	}
}

// NewPersistentParameters returns the peristent parameters initialization
func NewPersistentParameters() *PersistentParameters {
	return &PersistentParameters{
		NetworkRenamed: make(map[string][]string),
	}
}

// returns compress state of res.MarshalJSON
func dumpParameters(res FullParameters) []byte {
	jsonByte, _ := res.MarshalJSON()
	return compressByte(jsonByte)
}

// UnmarshalJson a FullParameters
func loadParameters(jsonByte []byte) FullParameters {
	var res2 FullParameters
	res2.UnmarshalJSON(decompressByte(jsonByte))
	return res2
}

//save ps(a FullParameters instance) to db
func saveParameters(group string, res FullParameters) error {
	//todo: why we should save ps in database? It can be regenerated from fingerprints bucket in db.
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		err2 = bucket.Put([]byte("fullParameters"), dumpParameters(res))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	return err
}

//return cached ps(a FullParameters instance) or get it from db then return
func openParameters(group string) (FullParameters, error) {

	psCached, ok := getPsCache(group)
	if ok {
		return psCached, nil
	}

	var ps = *NewFullParameters()
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("fullParameters"))
		ps = loadParameters(v)
		return nil
	})

	go setPsCache(group, ps)
	return ps, err
}

// Get persistentParameters from resources bucket in db
func openPersistentParameters(group string) (PersistentParameters, error) {
	var persistentPs = *NewPersistentParameters()
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("persistentParameters"))
		json.Unmarshal(v, &persistentPs)
		return nil
	})
	return persistentPs, err
}

// Set persistentParameters to resources bucket in db (it's used in remednetwork() function)
func savePersistentParameters(group string, res PersistentParameters) error {
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		jsonByte, _ := json.Marshal(res)
		err2 = bucket.Put([]byte("persistentParameters"), jsonByte)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err)
		}
		return err2
	})
	Debug.Println("Saved")
	return err
}

//group: group
//ps:
//fingerprintsInMemory:
//fingerprintsOrdering:
//updates ps with the new fingerprint.
//(The Parameters which are manipulated: NetworkMacs,NetworkLocs,UniqueMacs,UniqueLocs,MacCount,MacCountByLoc and Loaded)
func getParameters(group string, ps *FullParameters, fingerprintsInMemory map[string]Fingerprint, fingerprintsOrdering []string) {

	persistentPs, err := openPersistentParameters(group) //persistentPs is just like ps but with renamed network name; e.g.: "0" -> "1"
	ps.NetworkMacs = make(map[string]map[string]bool)
	ps.NetworkLocs = make(map[string]map[string]bool)
	ps.UniqueMacs = []string{}
	ps.UniqueLocs = []string{}
	ps.MacCount = make(map[string]int)
	ps.MacCountByLoc = make(map[string]map[string]int)
	ps.Loaded = true
	//opening the db
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get all parameters that don't need a network graph (?)
	for _, v1 := range fingerprintsOrdering {

		//log.Println("calculateResults=true")
		v2 := fingerprintsInMemory[v1]

		// append the fingerprint location to UniqueLocs array if doesn't exist in it.
		if !stringInSlice(v2.Location, ps.UniqueLocs) {
			ps.UniqueLocs = append(ps.UniqueLocs, v2.Location)
		}

		// MacCountByLoc initialization for new location
		if _, ok := ps.MacCountByLoc[v2.Location]; !ok {
			ps.MacCountByLoc[v2.Location] = make(map[string]int)
		}

		// building network
		macs := []string{}

		for _, router := range v2.WifiFingerprint {
			// building network
			macs = append(macs, router.Mac)

			// append the fingerprint mac to UniqueMacs array if doesn't exist in it.
			if !stringInSlice(router.Mac, ps.UniqueMacs) {
				ps.UniqueMacs = append(ps.UniqueMacs, router.Mac)
			}

			// mac count
			if _, ok := ps.MacCount[router.Mac]; !ok {
				ps.MacCount[router.Mac] = 0
			}
			ps.MacCount[router.Mac]++

			// mac by location count
			if _, ok := ps.MacCountByLoc[v2.Location][router.Mac]; !ok {
				ps.MacCountByLoc[v2.Location][router.Mac] = 0
			}
			ps.MacCountByLoc[v2.Location][router.Mac]++
		}

		// building network
		ps.NetworkMacs = buildNetwork(ps.NetworkMacs, macs)
	}

	ps.NetworkMacs = mergeNetwork(ps.NetworkMacs)

	// Rename the NetworkMacs
	if len(persistentPs.NetworkRenamed) > 0 {
		newNames := []string{}
		for k := range persistentPs.NetworkRenamed {
			newNames = append(newNames, k)

		}
		//todo: \/ wtf? Rename procedure could be redefined better.
		for n := range ps.NetworkMacs {
			renamed := false
			for mac := range ps.NetworkMacs[n] {
				for renamedN := range persistentPs.NetworkRenamed {
					if stringInSlice(mac, persistentPs.NetworkRenamed[renamedN]) && !stringInSlice(n, newNames) {
						ps.NetworkMacs[renamedN] = make(map[string]bool)
						for k, v := range ps.NetworkMacs[n] {
							ps.NetworkMacs[renamedN][k] = v //copy ps.NetworkMacs[n] to ps.NetworkMacs[renamedN]
						}
						delete(ps.NetworkMacs, n)
						renamed = true
					}
					if renamed {
						break
					}
				}
				if renamed {
					break
				}
			}
		}
	}

	// Get the locations for each graph (Has to have network built first)

	for _, v1 := range fingerprintsOrdering {

		v2 := fingerprintsInMemory[v1]
		//todo: Make the macs array just once for each fingerprint instead of repeating the process

		macs := []string{}
		for _, router := range v2.WifiFingerprint {
			macs = append(macs, router.Mac)
		}
		//todo: ps.NetworkMacs is created from mac array; so it seems that hasNetwork function doesn't do anything useful!
		networkName, inNetwork := hasNetwork(ps.NetworkMacs, macs)
		if inNetwork {
			if _, ok := ps.NetworkLocs[networkName]; !ok {
				ps.NetworkLocs[networkName] = make(map[string]bool)
			}
			if _, ok := ps.NetworkLocs[networkName][v2.Location]; !ok {
				ps.NetworkLocs[networkName][v2.Location] = true
			}
		}
	}

}

// return mixinOverride value from resources bucket in db
func getMixinOverride(group string) (float64, error) {
	group = strings.ToLower(group)
	override := float64(-1)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("mixinOverride"))
		if len(v) == 0 {
			return fmt.Errorf("No mixin override")
		}
		override, err = strconv.ParseFloat(string(v), 64)
		return err
	})
	return override, err
}

// return cutoffOverride value from resources bucket in db
func getCutoffOverride(group string) (float64, error) {
	group = strings.ToLower(group)
	override := float64(-1)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("cutoffOverride"))
		if len(v) == 0 {
			return fmt.Errorf("No mixin override")
		}
		override, err = strconv.ParseFloat(string(v), 64)
		return err
	})
	return override, err
}

// return KNN K Override value from resources bucket in db
func getKnnKOverride(group string) (int, error) {
	group = strings.ToLower(group)
	override := int(-1)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("knnKOverride"))
		if len(v) == 0 {
			return fmt.Errorf("No mixin override")
		}
		override, err = strconv.Atoi(string(v))
		return err
	})
	return override, err
}

// return KNN K Override value from resources bucket in db
func getMinRSSOverride(group string) (int, error) {
	group = strings.ToLower(group)
	override := int(-1)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("minRSSOverride"))
		if len(v) == 0 {
			return fmt.Errorf("No mixin override")
		}
		override, err = strconv.Atoi(string(v))
		return err
	})
	return override, err
}

// Set mixinOverride value to resources bucket in db
func setMixinOverride(group string, mixin float64) error {
	if (mixin < 0 || mixin > 1) && mixin != -1 {
		return fmt.Errorf("mixin must be between 0 and 1")
	}
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		err2 = bucket.Put([]byte("mixinOverride"), []byte(strconv.FormatFloat(mixin, 'E', -1, 64)))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	return err
}

// Set cutoffOverride value to resources bucket in db
func setCutoffOverride(group string, cutoff float64) error {
	if (cutoff < 0 || cutoff > 1) && cutoff != -1 {
		return fmt.Errorf("cutoff must be between 0 and 1")
	}
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		err2 = bucket.Put([]byte("cutoffOverride"), []byte(strconv.FormatFloat(cutoff, 'E', -1, 64)))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	return err
}

// Set KNN K Override value to resources bucket in db
func setKnnK(group string, knnk int) error {
	if knnk < 0 && knnk != -1 {
		return fmt.Errorf("knnk must be greater than 0")
	}
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		err2 = bucket.Put([]byte("knnKOverride"), []byte(strconv.Itoa(knnk)))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	return err
}

// Set KNN K Override value to resources bucket in db
func setMinRSS(group string, minRss int) error {
	if minRss < MaxRssi && minRss > MinRssi {
		return fmt.Errorf("minRss must be greater than 0")
	}
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		Error.Println(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		err2 = bucket.Put([]byte("minRSSOverride"), []byte(strconv.Itoa(minRss)))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	return err
}
