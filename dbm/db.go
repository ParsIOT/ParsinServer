// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// db.go contains generic functions for parsing data from the database.
// This file is not exhaustive of all database functions, if they pertain to a
// specific property (fingerprinting/priors/parameters), it will instead be in respective file.

package dbm

import (
	"log"
	"os"
	"path"
	"strings"
	"time"
	"encoding/json"
	"github.com/boltdb/bolt"
	"fmt"
	"ParsinServer/glb"
	"strconv"
	"errors"
	"ParsinServer/algorithms/parameters"
)

// checks is the database file exist or not.
func GroupExists(group string) bool {
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, strings.ToLower(group)+".db")); os.IsNotExist(err) {
		return false
	}
	return true
}

// renames the network, then calls savePersistentParameters() function to save ps
func RenameNetwork(group string, oldName string, newName string) {
	//todo: It's better to regenerate ps from the modified fingerprints bucket than modifying the current ps
	//glb.Debug.Println("Opening parameters")
	ps, _ := OpenParameters(group)
	//glb.Debug.Println("Opening persistent parameters")
	persistentPs, _ := OpenPersistentParameters(group)
	//glb.Debug.Println("Looping network macs")
	for n := range ps.NetworkMacs {
		if n == oldName {
			macs := []string{}
			glb.Debug.Println("Looping macs for ", n)
			for mac := range ps.NetworkMacs[n] {
				macs = append(macs, mac)
			}
			glb.Debug.Println("Adding to persistentPs")
			persistentPs.NetworkRenamed[newName] = macs
			delete(persistentPs.NetworkRenamed, oldName)
			break
		}
	}
	//glb.Debug.Println("Saving persistentPs")
	SavePersistentParameters(group, persistentPs)
}

// if the users of group are cached, returns them.
// otherwise, read them from database, cache them and return them.
func GetUsers(group string) []string {
	val, ok := GetUserCache(group)
	if ok {
		return val
	}

	uniqueUsers := []string{}
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints-track"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := LoadFingerprint(v, true)
			if !glb.StringInSlice(v2.Username, uniqueUsers) {
				uniqueUsers = append(uniqueUsers, v2.Username)
			}
		}
		return nil
	})

	go SetUserCache(group, uniqueUsers)
	return uniqueUsers
}

// returns MACs from fingerprints bucket
func GetUniqueMacs(group string) []string {
	defer glb.TimeTrack(time.Now(), "getUniqueMacs")
	uniqueMacs := []string{}

	//db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()
	//
	//db.View(func(tx *bolt.Tx) error {
	//	// Assume bucket exists and has keys
	//	b := tx.Bucket([]byte("fingerprints"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		v2 := LoadFingerprint(v, true)
	//		for _, router := range v2.WifiFingerprint {
	//			if !glb.StringInSlice(router.Mac, uniqueMacs) {
	//				uniqueMacs = append(uniqueMacs, router.Mac)
	//			}
	//		}
	//	}
	//	return nil
	//})

	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return uniqueMacs
	}
	for _,fp := range fingerprintInMemory{
		for _, router := range fp.WifiFingerprint {
			if !glb.StringInSlice(router.Mac, uniqueMacs) {
				uniqueMacs = append(uniqueMacs, router.Mac)
			}
		}
	}

	return uniqueMacs
}

// returns all locations in a fingerprints bucket
func GetUniqueLocations(group string) []string {
	//db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()
	//db.View(func(tx *bolt.Tx) error {
	//	// Assume bucket exists and has keys
	//	b := tx.Bucket([]byte("fingerprints"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		v2 := LoadFingerprint(v, true)
	//		if !glb.StringInSlice(v2.Location, uniqueLocs) {
	//			uniqueLocs = append(uniqueLocs, v2.Location)
	//		}
	//	}
	//	return nil
	//})

	var uniqueLocs []string
	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return uniqueLocs
	}
	for _,fp := range fingerprintInMemory{
		if !glb.StringInSlice(fp.Location, uniqueLocs) {
			uniqueLocs = append(uniqueLocs, fp.Location)
		}
	}
	return uniqueLocs
}

// returns count of each MAC in a fingerprints bucket
func GetMacCount(group string) (macCount map[string]int) {
	macCount = make(map[string]int)
	//db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()
	//db.View(func(tx *bolt.Tx) error {
	//	// Assume bucket exists and has keys
	//	b := tx.Bucket([]byte("fingerprints"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		v2 := LoadFingerprint(v, true)
	//		for _, router := range v2.WifiFingerprint {
	//			if _, ok := macCount[router.Mac]; !ok {
	//				macCount[router.Mac] = 0
	//			}
	//			macCount[router.Mac]++
	//		}
	//	}
	//	return nil
	//})

	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return macCount
	}
	for _,fp := range fingerprintInMemory{
		for _, router := range fp.WifiFingerprint {
			if _, ok := macCount[router.Mac]; !ok {
				macCount[router.Mac] = 0
			}
			macCount[router.Mac]++
		}
	}
	return macCount
}

// returns count of each MAC in a location
func GetMacCountByLoc(group string) (macCountByLoc map[string]map[string]int) {
	macCountByLoc = make(map[string]map[string]int)
	//db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()
	//db.View(func(tx *bolt.Tx) error {
	//	// Assume bucket exists and has keys
	//	b := tx.Bucket([]byte("fingerprints"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		v2 := LoadFingerprint(v, true)
	//		if _, ok := macCountByLoc[v2.Location]; !ok {
	//			macCountByLoc[v2.Location] = make(map[string]int)
	//		}
	//		for _, router := range v2.WifiFingerprint {
	//			if _, ok := macCountByLoc[v2.Location][router.Mac]; !ok {
	//				macCountByLoc[v2.Location][router.Mac] = 0
	//			}
	//			macCountByLoc[v2.Location][router.Mac]++
	//		}
	//	}
	//	return nil
	//})


	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return macCountByLoc
	}
	for _,fp := range fingerprintInMemory{
		if _, ok := macCountByLoc[fp.Location]; !ok {
			macCountByLoc[fp.Location] = make(map[string]int)
		}
		for _, router := range fp.WifiFingerprint {
			if _, ok := macCountByLoc[fp.Location][router.Mac]; !ok {
				macCountByLoc[fp.Location][router.Mac] = 0
			}
			macCountByLoc[fp.Location][router.Mac]++
		}
	}

	return macCountByLoc
}

// Return admin users as map style(user:pass)
func GetAdminUsers() (map[string]string, error) {
	userList := make(map[string]string)

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, "users.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("users"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		v := b.Get([]byte("adminList"))
		if len(v) == 0 {
			fmt.Errorf("Admin list is empty")
			return nil
		} else {
			err := json.Unmarshal(v, &userList)
			if err != nil {
				log.Fatal(err)
			}
			return err
		}
	})
	return userList, err
}

// Add an admin user or change his password
func AddAdminUser(username string, password string) error {
	userList := make(map[string]string)
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, "users.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create users bucket if doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists([]byte("users"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		return err2
	})

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf("Resources don't exist")
		}
		v := b.Get([]byte("adminList"))
		if len(v) == 0 {
			fmt.Errorf("Admin list is empty")
			return nil
		} else {
			err := json.Unmarshal(v, &userList)
			if err != nil {
				log.Fatal(err)
			}
			return err
		}
	})

	if err != nil {
		return err
	}

	// Add an admin user or change his password
	userList[username] = password

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("users"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		marshalledUserList, _ := json.Marshal(userList)
		err2 = bucket.Put([]byte("adminList"), marshalledUserList)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})

	return err
}

// Set macs that to be filtered
func SetFilterMacDB(group string, FilterMacs []string) error {
	glb.Warning.Println(FilterMacs)
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	glb.Debug.Println(FilterMacs)
	// Create filtermacs bucket if doesn't exist & set filtermacs
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))

		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		//Warning.Println(FilterMacs)
		marshalledFilterMacList, _ := json.Marshal(FilterMacs)
		err2 = bucket.Put([]byte("filterMacList"), marshalledFilterMacList)
		//Warning.Println("bucket creation problem :",err2)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		//Warning.Println("setFilterMacDB successfully")
		return err2
	})

	return err
}

// Get macs that to be filtered
func GetFilterMacDB(group string) (error, []string) {
	var FilterMacs []string
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		v := b.Get([]byte("filterMacList"))
		if len(v) == 0 {
			fmt.Errorf("filterMacList list is empty")
			return nil
		} else {
			err := json.Unmarshal(v, &FilterMacs)
			if err != nil {
				log.Fatal(err)
				fmt.Println("hi")
			}
			return err
		}
	})

	return err, FilterMacs
}


// return mixinOverride value from resources bucket in db
func GetMixinOverride(group string) (float64, error) {
	group = strings.ToLower(group)
	override := float64(-1)
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
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
func GetCutoffOverride(group string) (float64, error) {
	group = strings.ToLower(group)
	override := float64(-1)
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
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
func GetKnnKOverride(group string) (int, error) {
	group = strings.ToLower(group)
	override := int(-1)
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		v := b.Get([]byte("knnKOverride"))
		if len(v) == 0 {
			return fmt.Errorf("No mixin override")
		}
		override, err = strconv.Atoi(string(v))
		return err
	})
	if (override == 0) {
		err := errors.New("invalid knnOverride")
		return glb.DefaultKnnK, err
	}
	return override, err
}

// return KNN K Override value from resources bucket in db
func GetMinRSSOverride(group string) (int, error) {
	group = strings.ToLower(group)
	override := int(-1)
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
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
func SetMixinOverride(group string, mixin float64) error {
	if (mixin < 0 || mixin > 1) && mixin != -1 {
		return fmt.Errorf("mixin must be between 0 and 1")
	}
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
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
func SetCutoffOverride(group string, cutoff float64) error {
	if (cutoff < 0 || cutoff > 1) && cutoff != -1 {
		return fmt.Errorf("cutoff must be between 0 and 1")
	}
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
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
func SetKnnK(group string, knnk int) error {
	if knnk <= 0 && knnk != -1 {
		return fmt.Errorf("knnk must be greater than 0")
	}
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
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
func SetMinRSS(group string, minRss int) error {
	if minRss < glb.MaxRssi && minRss > glb.MinRssi {
		return fmt.Errorf("minRss must be greater than 0")
	}
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
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


func GetLearnFingerPrints(group string,doFilter bool)([]string,map[string]parameters.Fingerprint,error){
	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var fingerprintsOrdering []string
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println("Can't get learn fingerprints.")
		return fingerprintsOrdering,fingerprintsInMemory,err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		//gets the fingerprint bucket
		b := tx.Bucket([]byte("fingerprints"))
		if b == nil {
			return fmt.Errorf("No fingerprint bucket")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fingerprintsInMemory[string(k)] = LoadFingerprint(v,doFilter)
			fingerprintsOrdering = append(fingerprintsOrdering, string(k))
		}
		return nil
	})
	if err != nil {
		glb.Error.Println("Can't get learn fingerprints.")
		return fingerprintsOrdering,fingerprintsInMemory,err
	}
	return fingerprintsOrdering,fingerprintsInMemory,nil
}


func PutDataIntoDatabase(res parameters.Fingerprint, database string) error {
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, res.Group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte(database))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		if res.Timestamp == 0 {
			res.Timestamp = time.Now().UnixNano()
		}
		err2 = bucket.Put([]byte(strconv.FormatInt(res.Timestamp, 10)), parameters.DumpFingerprint(res))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	db.Close()
	return err
}
