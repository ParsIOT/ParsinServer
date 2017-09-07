// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// priors.go contains variables for calculating priors.

package main

import (
	"log"

	"path"

	"github.com/boltdb/bolt"
	//"fmt"
	//"fmt"
)

// PdfType dictates the width of gaussian smoothing
var PdfType []float32


// MaxRssi is the maximum level of signal
var MaxRssi int

// MinRssi is the minimum level of signal that can save to db
var MinRssi int

// MinRssiOpt is the minimum level of signal for learning and tracking
var MinRssiOpt int

// RssiPartitions are the calculated number of partitions from MinRssi and MaxRssi
var RssiPartitions int

// Absentee is the base level of probability for any signal
var Absentee float32

// RssiRange is the calculated partitions in array form
var RssiRange []float32

// FoldCrossValidation is the amount of data left out during learning to be used in cross validation
var FoldCrossValidation float64

// Variables initialization
func init() {
	//todo:what is PdfType and how to find the values
	PdfType = []float32{.1995, .1760, .1210, .0648, .027, 0.005}
	Absentee = 1e-6
	MinRssi = -110 //default:-110,ble=-80,wifi=-75
	MinRssiOpt = -100
	MaxRssi = 5
	RssiPartitions = MaxRssi - MinRssi + 1
	RssiRange = make([]float32, RssiPartitions)
	for i := 0; i < len(RssiRange); i++ {
		RssiRange[i] = float32(MinRssi + i)
	}
	FoldCrossValidation = 5
}

// deprecated
func optimizePriors(group string) {
	// generate the fingerprintsInMemory
	fingerprintsInMemory := make(map[string]Fingerprint)
	var fingerprintsOrdering []string
	//opening the db
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fingerprintsInMemory[string(k)] = loadFingerprint(v)
			// fmt.Println(fingerprintsInMemory[string(k)].Location, string(k))
			fingerprintsOrdering = append(fingerprintsOrdering, string(k))
		}
		return nil
	})
	db.Close()

	var ps = *NewFullParameters()
	getParameters(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	if RuntimeArgs.GaussianDist {
		calculateGaussianPriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	} else {
		calculatePriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	}

	// fmt.Println(string(dumpParameters(ps)))
	// ps, _ = openParameters("findtest")
	var results = *NewResultsParameters()
	for n := range ps.Priors {
		ps.Results[n] = results
	}
	// fmt.Println(ps.Results)
	// ps.Priors["0"].Special["MixIn"] = 1.0
	// fmt.Println(crossValidation(group, "0", &ps))
	// fmt.Println(ps.Results)

	// loop through these parameters
	mixins := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
	cutoffs := []float64{0.005}

	for n := range ps.Priors {
		bestResult := float64(0)
		bestMixin := float64(0)
		bestCutoff := float64(0)
		for _, cutoff := range cutoffs {
			for _, mixin := range mixins {
				ps.Priors[n].Special["MixIn"] = mixin
				ps.Priors[n].Special["VarabilityCutoff"] = cutoff
				avgAccuracy := crossValidation(group, n, &ps, fingerprintsInMemory, fingerprintsOrdering)
				// avgAccuracy := crossValidation(group, n, &ps)
				if avgAccuracy > bestResult {
					bestResult = avgAccuracy
					bestCutoff = cutoff
					bestMixin = mixin
				}
			}
		}
		ps.Priors[n].Special["MixIn"] = bestMixin
		ps.Priors[n].Special["VarabilityCutoff"] = bestCutoff
		// Final validation
		crossValidation(group, n, &ps, fingerprintsInMemory, fingerprintsOrdering)
		// crossValidation(group, n, &ps)
	}

	go saveParameters(group, ps)
	go setPsCache(group, ps)
}

func regenerateEverything(group string) {
	// generate the fingerprintsInMemory
	fingerprintsInMemory := make(map[string]Fingerprint)
	var fingerprintsOrdering []string
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fingerprintsInMemory[string(v)] = loadFingerprint(v)
			fingerprintsOrdering = append(fingerprintsOrdering, string(v))
		}
		return nil
	})
	db.Close()

	var ps = *NewFullParameters()
	ps, _ = openParameters(group)
	getParameters(group, &ps, fingerprintsInMemory, fingerprintsOrdering)//openParameters is only called here.
	if RuntimeArgs.GaussianDist {
		calculateGaussianPriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	} else {
		calculatePriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	}
	var results = *NewResultsParameters()
	for n := range ps.Priors {
		ps.Results[n] = results
	}
	for n := range ps.Priors {
		crossValidation(group, n, &ps, fingerprintsInMemory, fingerprintsOrdering)
	}
	saveParameters(group, ps)
}

// (1/FoldCrossValidation) of the learned fingerprints are predicted with ps data, then results are wrote in ps.Results
func crossValidation(group string, n string, ps *FullParameters, fingerprintsInMemory map[string]Fingerprint, fingerprintsOrdering []string) float64 {
	for loc := range ps.NetworkLocs[n] {
		ps.Results[n].TotalLocations[loc] = 0
		ps.Results[n].CorrectLocations[loc] = 0
		ps.Results[n].Accuracy[loc] = 0
		ps.Results[n].Guess[loc] = make(map[string]int)
	}

	for _, v1 := range fingerprintsOrdering {

		v2 := fingerprintsInMemory[v1]
		if len(v2.WifiFingerprint) == 0 {
			continue
		}
		if _, ok := ps.NetworkLocs[n][v2.Location]; ok {
			locationGuess, _ := calculatePosterior(v2, *ps)
			ps.Results[n].TotalLocations[v2.Location]++ //set TotalLocations
			if locationGuess == v2.Location {
				ps.Results[n].CorrectLocations[v2.Location]++ //set CorrectLocations
			}
			if _, ok := ps.Results[n].Guess[v2.Location]; !ok {
				ps.Results[n].Guess[v2.Location] = make(map[string]int)
			}
			if _, ok := ps.Results[n].Guess[v2.Location][locationGuess]; !ok {
				ps.Results[n].Guess[v2.Location][locationGuess] = 0
			}
			ps.Results[n].Guess[v2.Location][locationGuess]++ //set Guess
		}

	}

	average := float64(0)
	for loc := range ps.NetworkLocs[n] {
		if ps.Results[n].TotalLocations[loc] > 0 {
			// fmt.Println(ps.Results[n].CorrectLocations[loc], ps.Results[n].TotalLocations[loc])
			// set Accuracy
			ps.Results[n].Accuracy[loc] = int(100.0 * ps.Results[n].CorrectLocations[loc] / ps.Results[n].TotalLocations[loc])
			average += float64(ps.Results[n].Accuracy[loc])
		}
	}
	average = average / float64(len(ps.NetworkLocs[n]))

	return average
}

// calculatePriors generates the ps.Prior(P,NP,MacFreq,NMacFreq) data and ps.MacVariability for Naive-Bayes classification. Now deprecated, use calculatePriorsThreaded instead.
//todo: write calculatePriorsThreaded function
func calculatePriors(group string, ps *FullParameters, fingerprintsInMemory map[string]Fingerprint, fingerprintsOrdering []string) {
	// defer timeTrack(time.Now(), "calculatePriors")
	ps.Priors = make(map[string]PriorParameters)
	for n := range ps.NetworkLocs {
		var newPrior = *NewPriorParameters()
		ps.Priors[n] = newPrior
	}

	// Initialization
	ps.MacVariability = make(map[string]float32)
	for n := range ps.Priors {
		ps.Priors[n].Special["MacFreqMin"] = float64(100)
		ps.Priors[n].Special["NMacFreqMin"] = float64(100)
		for loc := range ps.NetworkLocs[n] {
			ps.Priors[n].P[loc] = make(map[string][]float32)
			ps.Priors[n].NP[loc] = make(map[string][]float32)
			ps.Priors[n].MacFreq[loc] = make(map[string]float32)
			ps.Priors[n].NMacFreq[loc] = make(map[string]float32)
			for mac := range ps.NetworkMacs[n] {
				ps.Priors[n].P[loc][mac] = make([]float32, RssiPartitions)
				ps.Priors[n].NP[loc][mac] = make([]float32, RssiPartitions)
			}
		}
	}

	//create gaussian distribution for every mac in every location

	for _, v1 := range fingerprintsOrdering {

		v2 := fingerprintsInMemory[v1]
		macs := []string{}
		for _, router := range v2.WifiFingerprint {
			macs = append(macs, router.Mac)
		}

		// todo: ps is set in the getParameters function (getParameters is called before calculatePriors), so calling the hasNetwork function returns true
		networkName, inNetwork := hasNetwork(ps.NetworkMacs, macs)
		if inNetwork {
			for _, router := range v2.WifiFingerprint {
				if router.Rssi > MinRssiOpt {
					//fmt.Println(router.Rssi)
					ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi] += PdfType[0]
					//make the real probability of the rssi distribution
					for i, val := range PdfType {
						if i > 0 {
							//if (router.Rssi-MinRssi-i<2) {
							//	fmt.Println("i=", i)
							//	fmt.Println("router.Rssi=", router.Rssi)
							//	fmt.Println("router.rssi-MinRSSi-i=", router.Rssi-MinRssi-i)
							//}
							if (router.Rssi-MinRssi-i > 0 && router.Rssi-MinRssi+i < RssiPartitions) {
								ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi-i] += val
								ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi+i] += val
							}

						}
					}
					//} else {
					//	Warning.Println(router.Rssi)
				}
			}
		}

	}

	// Calculate the nP
	for n := range ps.Priors {
		for locN := range ps.NetworkLocs[n] {
			for loc := range ps.NetworkLocs[n] {
				if loc != locN {
					for mac := range ps.NetworkMacs[n] {
						for i := range ps.Priors[n].P[locN][mac] {
							//i is rssi
							if ps.Priors[n].P[loc][mac][i] > 0 {
								ps.Priors[n].NP[locN][mac][i] += ps.Priors[n].P[loc][mac][i]
							}
						}
					}
				}
			}
		}
	}

	// Add in absentee, normalize P and nP and determine MacVariability

	for n := range ps.Priors {
		macAverages := make(map[string][]float32)

		for loc := range ps.NetworkLocs[n] {
			for mac := range ps.NetworkMacs[n] {
				for i := range ps.Priors[n].P[loc][mac] { //i is rssi
					//why using Absentee instead of 0
					ps.Priors[n].P[loc][mac][i] += Absentee
					ps.Priors[n].NP[loc][mac][i] += Absentee
				}
				total := float32(0) //total = sum of probabilities(P) of all rssi for a specific mac and location
				for _, val := range ps.Priors[n].P[loc][mac] {
					total += val
				}
				averageMac := float32(0)
				for i, val := range ps.Priors[n].P[loc][mac] {
					if val > float32(0) { //val is always => Absentee >0 --> it is required in normalization
						ps.Priors[n].P[loc][mac][i] = val / total                //normalizing P
						averageMac += RssiRange[i] * ps.Priors[n].P[loc][mac][i] // RssiRange[i] equals to rssi.
						//todo: average mac is not valid if the probability distribution (P) is not a standard gaussian function,e.g. has two peaks
					}
				}
				//why checking is required?
				if averageMac < float32(0) {
					if _, ok := macAverages[mac]; !ok {
						macAverages[mac] = []float32{}
					}
					macAverages[mac] = append(macAverages[mac], averageMac) // averageMac of each mac in every locations
				}

				//normalizing NP
				total = float32(0)
				for i := range ps.Priors[n].NP[loc][mac] {
					total += ps.Priors[n].NP[loc][mac][i]
				}
				if total > 0 {
					for i := range ps.Priors[n].NP[loc][mac] {
						ps.Priors[n].NP[loc][mac][i] = ps.Priors[n].NP[loc][mac][i] / total
					}
				}
			}
		}

		// Determine MacVariability
		for mac := range macAverages {
			//todo: why 2?
			if len(macAverages[mac]) <= 2 {
				ps.MacVariability[mac] = float32(1)
			} else {
				maxVal := float32(-10000)
				for _, val := range macAverages[mac] {
					if val > maxVal {
						maxVal = val
					}
				}
				for i, val := range macAverages[mac] {
					//todo: why not using the actual values of macAverages instead of the normalized values?
					macAverages[mac][i] = maxVal / val // normalization(because val is < 0, we use maxVal/val instead of val /maxVal)
				}
				// MacVariability shows the standard deviation of a specific AP in all locations
				ps.MacVariability[mac] = standardDeviation(macAverages[mac]) //refer to line 300 todo
			}
		}
	}

	// Determine mac frequencies and normalize
	for n := range ps.Priors {
		for loc := range ps.NetworkLocs[n] {
			maxCount := 0
			for mac := range ps.MacCountByLoc[loc] {
				if ps.MacCountByLoc[loc][mac] > maxCount {
					maxCount = ps.MacCountByLoc[loc][mac] //maxCount:repeat number of the most seen mac in a location

				}
			}
			//fmt.Println("MAX COUNT:", maxCount)
			for mac := range ps.MacCountByLoc[loc] {
				//if a mac is not seen in a location, the macFreq of that mac equals to 0 (ps.MacCountByLoc[loc][mac]).
				//todo: Does the above mentioned 0 value make some error in the bayesian function?
				ps.Priors[n].MacFreq[loc][mac] = float32(ps.MacCountByLoc[loc][mac]) / float32(maxCount)
				//fmt.Println("mac freq:", ps.Priors[n].MacFreq[loc][mac])
				if float64(ps.Priors[n].MacFreq[loc][mac]) < ps.Priors[n].Special["MacFreqMin"] {
					ps.Priors[n].Special["MacFreqMin"] = float64(ps.Priors[n].MacFreq[loc][mac])
				}
			}
		}
	}

	// Determine negative mac frequencies and normalize
	for n := range ps.Priors {
		for loc1 := range ps.Priors[n].MacFreq {
			sum := float32(0)
			for loc2 := range ps.Priors[n].MacFreq {
				if loc2 != loc1 {
					for mac := range ps.Priors[n].MacFreq[loc2] {
						ps.Priors[n].NMacFreq[loc1][mac] += ps.Priors[n].MacFreq[loc2][mac]
					}
					sum++
				}
			}
			// sum = i(i-1); i = ps.NetworkLocs[n]
			// Normalize
			//Done: it seems that sum is not calculated correctly. It should be equals to "number of locations-1"
			if sum > 0 {
				for mac := range ps.Priors[n].MacFreq[loc1] {
					ps.Priors[n].NMacFreq[loc1][mac] = ps.Priors[n].NMacFreq[loc1][mac] / sum
					if float64(ps.Priors[n].NMacFreq[loc1][mac]) < ps.Priors[n].Special["NMacFreqMin"] {
						ps.Priors[n].Special["NMacFreqMin"] = float64(ps.Priors[n].NMacFreq[loc1][mac])
					}
				}
			}
		}
	}
	//todo: the default values for MixIn and Cutoff should be set as initial values not hardcoded values
	for n := range ps.Priors {
		ps.Priors[n].Special["MixIn"] = 0.5
		//todo: spell check for Varability!
		ps.Priors[n].Special["VarabilityCutoff"] = 0
	}
}

func calculateGaussianPriors(group string, ps *FullParameters, fingerprintsInMemory map[string]Fingerprint, fingerprintsOrdering []string) {
	// defer timeTrack(time.Now(), "calculatePriors")
	ps.Priors = make(map[string]PriorParameters)
	for n := range ps.NetworkLocs {
		var newPrior = *NewPriorParameters()
		ps.Priors[n] = newPrior
	}

	// Initialization
	Rssies := make(map[string]map[string][]float64)
	RssiesVariance := make(map[string]map[string]float64)
	RssiesAvg := make(map[string]map[string]float64)

	ps.MacVariability = make(map[string]float32)
	for n := range ps.Priors {
		ps.Priors[n].Special["MacFreqMin"] = float64(100)
		ps.Priors[n].Special["NMacFreqMin"] = float64(100)
		for loc := range ps.NetworkLocs[n] {
			ps.Priors[n].P[loc] = make(map[string][]float32)

			Rssies[loc] = make(map[string][]float64)
			RssiesVariance[loc] = make(map[string]float64)
			RssiesAvg[loc] = make(map[string]float64)

			ps.Priors[n].NP[loc] = make(map[string][]float32)
			ps.Priors[n].MacFreq[loc] = make(map[string]float32)
			ps.Priors[n].NMacFreq[loc] = make(map[string]float32)
			for mac := range ps.NetworkMacs[n] {
				ps.Priors[n].P[loc][mac] = make([]float32, RssiPartitions)

				Rssies[loc][mac] = make([]float64, 0)
				RssiesVariance[loc][mac] = float64(0)
				RssiesAvg[loc][mac] = float64(0)

				ps.Priors[n].NP[loc][mac] = make([]float32, RssiPartitions)
			}
		}
	}

	//create gaussian distribution for every mac in every location

	// create list of collected rssi according to the locations and MACs
	for _, v1 := range fingerprintsOrdering {
		v2 := fingerprintsInMemory[v1]
		macs := []string{}
		for _, router := range v2.WifiFingerprint {
			macs = append(macs, router.Mac)
		}
		_, inNetwork := hasNetwork(ps.NetworkMacs, macs)
		if inNetwork {
			for _, router := range v2.WifiFingerprint {
				if router.Rssi > MinRssiOpt {
					//fmt.Println(router.Rssi)
					Rssies[v2.Location][router.Mac] = append(Rssies[v2.Location][router.Mac], float64(router.Rssi-MinRssi))
				}
			}

		}
	}

	// Calculate average and variance of a rssi list of a mac in a location
	for loc := range Rssies {
		for mac := range Rssies[loc] {
			//fmt.Println("RSSIes for loc:",loc,"& mac:",mac)
			//fmt.Println(Rssies[loc][mac])
			//fmt.Println("######")
			RssiesAvg[loc][mac] = average64(Rssies[loc][mac])
			RssiesVariance[loc][mac] = variance64(Rssies[loc][mac])
		}
	}

	//fmt.Println("###")

	g := NewGaussian(0, 1)

	// Create gaussian distribution; set probability for each rssi of each mac in each location
	for n := range ps.Priors {
		for loc := range ps.NetworkLocs[n] {
			for mac := range ps.NetworkMacs[n] {
				for rssi := 0; rssi < len(RssiRange); rssi++ {
					if (RssiesVariance[loc][mac] == 0) {
						g = NewGaussian(RssiesAvg[loc][mac], 1)
					} else {
						g = NewGaussian(RssiesAvg[loc][mac], RssiesVariance[loc][mac])
					}
					//fmt.Println(float32(g.Pdf(float64(rssi))))
					//fmt.Println(loc)
					//fmt.Println(mac)
					//fmt.Println(rssi)
					ps.Priors[n].P[loc][mac][rssi] = float32(g.Pdf(float64(rssi)))
				}
			}
		}
	}

	// Calculate the nP
	for n := range ps.Priors {
		for locN := range ps.NetworkLocs[n] {
			for loc := range ps.NetworkLocs[n] {
				if loc != locN {
					for mac := range ps.NetworkMacs[n] {
						for i := range ps.Priors[n].P[locN][mac] {
							//i is rssi
							if ps.Priors[n].P[loc][mac][i] > 0 {
								ps.Priors[n].NP[locN][mac][i] += ps.Priors[n].P[loc][mac][i]
							}
						}
					}
				}
			}
		}
	}

	// Add in absentee, normalize P and nP and determine MacVariability

	for n := range ps.Priors {
		macAverages := make(map[string][]float32)

		for loc := range ps.NetworkLocs[n] {
			for mac := range ps.NetworkMacs[n] {
				for i := range ps.Priors[n].P[loc][mac] { //i is rssi
					//why using Absentee instead of 0
					ps.Priors[n].P[loc][mac][i] += Absentee
					ps.Priors[n].NP[loc][mac][i] += Absentee
				}
				total := float32(0) //total = sum of probabilities(P) of all rssi for a specific mac and location
				for _, val := range ps.Priors[n].P[loc][mac] {
					total += val
				}
				averageMac := float32(0)
				for i, val := range ps.Priors[n].P[loc][mac] {
					if val > float32(0) { //val is always => Absentee >0 --> it is required in normalization
						ps.Priors[n].P[loc][mac][i] = val / total                //normalizing P
						averageMac += RssiRange[i] * ps.Priors[n].P[loc][mac][i] // RssiRange[i] equals to rssi.
						//todo: average mac is not valid if the probability distribution (P) is not a standard gaussian function,e.g. has two peaks
					}
				}
				//why checking is required?
				if averageMac < float32(0) {
					if _, ok := macAverages[mac]; !ok {
						macAverages[mac] = []float32{}
					}
					macAverages[mac] = append(macAverages[mac], averageMac) // averageMac of each mac in every locations
				}

				//normalizing NP
				total = float32(0)
				for i := range ps.Priors[n].NP[loc][mac] {
					total += ps.Priors[n].NP[loc][mac][i]
				}
				if total > 0 {
					for i := range ps.Priors[n].NP[loc][mac] {
						ps.Priors[n].NP[loc][mac][i] = ps.Priors[n].NP[loc][mac][i] / total
					}
				}
			}
		}

		// Determine MacVariability
		for mac := range macAverages {
			//todo: why 2?
			if len(macAverages[mac]) <= 2 {
				ps.MacVariability[mac] = float32(1)
			} else {
				maxVal := float32(-10000)
				for _, val := range macAverages[mac] {
					if val > maxVal {
						maxVal = val
					}
				}
				for i, val := range macAverages[mac] {
					//todo: why not using the actual values of macAverages instead of the normalized values?
					macAverages[mac][i] = maxVal / val // normalization(because val is < 0, we use maxVal/val instead of val /maxVal)
				}
				// MacVariability shows the standard deviation of a specific AP in all locations
				ps.MacVariability[mac] = standardDeviation(macAverages[mac]) //refer to line 300 todo
			}
		}
	}

	// Determine mac frequencies and normalize
	for n := range ps.Priors {
		for loc := range ps.NetworkLocs[n] {
			maxCount := 0
			for mac := range ps.MacCountByLoc[loc] {
				if ps.MacCountByLoc[loc][mac] > maxCount {
					maxCount = ps.MacCountByLoc[loc][mac] //maxCount:repeat number of the most seen mac in a location

				}
			}
			//fmt.Println("MAX COUNT:", maxCount)
			for mac := range ps.MacCountByLoc[loc] {
				//if a mac is not seen in a location, the macFreq of that mac equals to 0 (ps.MacCountByLoc[loc][mac]).
				//todo: Does the above mentioned 0 value make some error in the bayesian function?
				ps.Priors[n].MacFreq[loc][mac] = float32(ps.MacCountByLoc[loc][mac]) / float32(maxCount)
				//fmt.Println("mac freq:", ps.Priors[n].MacFreq[loc][mac])
				if float64(ps.Priors[n].MacFreq[loc][mac]) < ps.Priors[n].Special["MacFreqMin"] {
					ps.Priors[n].Special["MacFreqMin"] = float64(ps.Priors[n].MacFreq[loc][mac])
				}
			}
		}
	}

	// Determine negative mac frequencies and normalize
	for n := range ps.Priors {
		for loc1 := range ps.Priors[n].MacFreq {
			sum := float32(0)
			for loc2 := range ps.Priors[n].MacFreq {
				if loc2 != loc1 {
					for mac := range ps.Priors[n].MacFreq[loc2] {
						ps.Priors[n].NMacFreq[loc1][mac] += ps.Priors[n].MacFreq[loc2][mac]
					}
					sum++
				}
			}
			// sum = i(i-1); i = ps.NetworkLocs[n]
			// Normalize
			//Done: it seems that sum is not calculated correctly. It should be equals to "number of locations-1"
			if sum > 0 {
				for mac := range ps.Priors[n].MacFreq[loc1] {
					ps.Priors[n].NMacFreq[loc1][mac] = ps.Priors[n].NMacFreq[loc1][mac] / sum
					if float64(ps.Priors[n].NMacFreq[loc1][mac]) < ps.Priors[n].Special["NMacFreqMin"] {
						ps.Priors[n].Special["NMacFreqMin"] = float64(ps.Priors[n].NMacFreq[loc1][mac])
					}
				}
			}
		}
	}
	//todo: the default values for MixIn and Cutoff should be set as initial values not hardcoded values
	for n := range ps.Priors {
		ps.Priors[n].Special["MixIn"] = 0.5
		//todo: spell check for Varability!
		ps.Priors[n].Special["VarabilityCutoff"] = 0
	}

}