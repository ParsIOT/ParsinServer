// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// priors.go contains variables for calculating priors.

package bayes

import (
	"ParsinServer/glb"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/algorithms/clustering"
	"ParsinServer/dbm"
)

// PdfType dictates the width of gaussian smoothing
var PdfType []float32


// RssiPartitions are the calculated number of partitions from MinRssi and MaxRssi
var RssiPartitions int

// Absentee is the base level of probability for any signal
var Absentee float32


// FoldCrossValidation is the amount of data left out during learning to be used in cross validation
var FoldCrossValidation float64

// Variables initialization
func init() {
	//todo:what is PdfType and how to find the values
	PdfType = []float32{.1995, .1760, .1210, .0648, .027, 0.005}
	Absentee = 1e-6

	RssiPartitions = glb.MaxRssi - glb.MinRssi + 1
	glb.RssiRange = make([]float32, RssiPartitions)
	for i := 0; i < len(glb.RssiRange); i++ {
		glb.RssiRange[i] = float32(glb.MinRssi + i)
	}
	FoldCrossValidation = 5
}

// deprecated
func optimizePriors(group string) {
	// generate the fingerprintsInMemory
	var gp = dbm.NewGroup(group)
	defer dbm.GM.GetGroup(group).Set(gp)

	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var fingerprintsOrdering []string
	var err error

	fingerprintsOrdering,fingerprintsInMemory,err = dbm.GetLearnFingerPrints(group,true)
	if err != nil{
		return
	}

	//var ps = *parameters.NewFullParameters()
	GetParameters(group, gp, fingerprintsInMemory, fingerprintsOrdering)
	if glb.RuntimeArgs.GaussianDist {
		calculateGaussianPriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
	} else {
		calculatePriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
	}

	// fmt.Println(string(dumpParameters(ps)))
	// ps, _ = openParameters("findtest")
	var results = *parameters.NewResultsParameters()
	for n := range gp.Priors {
		gp.Results[n] = results
	}
	// fmt.Println(ps.Results)
	// ps.Priors["0"].Special["MixIn"] = 1.0
	// fmt.Println(crossValidation(group, "0", &ps))
	// fmt.Println(ps.Results)

	// loop through these parameters
	mixins := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
	cutoffs := []float64{0.005}

	for n := range gp.Priors {
		bestResult := float64(0)
		bestMixin := float64(0)
		bestCutoff := float64(0)
		for _, cutoff := range cutoffs {
			for _, mixin := range mixins {
				gp.Priors[n].Special["MixIn"] = mixin
				gp.Priors[n].Special["VarabilityCutoff"] = cutoff
				avgAccuracy := crossValidation(group, n, gp, fingerprintsInMemory, fingerprintsOrdering)
				// avgAccuracy := crossValidation(group, n, &ps)
				if avgAccuracy > bestResult {
					bestResult = avgAccuracy
					bestCutoff = cutoff
					bestMixin = mixin
				}
			}
		}
		gp.Priors[n].Special["MixIn"] = bestMixin
		gp.Priors[n].Special["VarabilityCutoff"] = bestCutoff
		// Final validation
		crossValidation(group, n, gp, fingerprintsInMemory, fingerprintsOrdering)
		// crossValidation(group, n, &ps)
	}

	//go dbm.SaveParameters(group, ps)
	//go dbm.SetPsCache(group, ps)
}

func regenerateEverything(group string) {
	// generate the fingerprintsInMemory
	var gp = dbm.GM.GetGroup(group).Get()
	defer dbm.GM.GetGroup(group).Set(gp)

	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var fingerprintsOrdering []string
	var err error

	fingerprintsOrdering,fingerprintsInMemory,err = dbm.GetLearnFingerPrints(group,true)
	if err != nil{
		return
	}

	//var ps = *parameters.NewFullParameters()
	//ps, _ = dbm.OpenParameters(group)
	GetParameters(group, gp, fingerprintsInMemory, fingerprintsOrdering)//openParameters is only called here.
	if glb.RuntimeArgs.GaussianDist {
		calculateGaussianPriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
	} else {
		calculatePriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
	}
	var results = *parameters.NewResultsParameters()
	for n := range gp.Priors {
		gp.Results[n] = results
	}
	for n := range gp.Priors {
		crossValidation(group, n, gp, fingerprintsInMemory, fingerprintsOrdering)
	}
	//dbm.SaveParameters(group, ps)
}

// (1/FoldCrossValidation) of the learned fingerprints are predicted with ps data, then results are wrote in ps.Results
func crossValidation(group string, n string,gp *dbm.Group, fingerprintsInMemory map[string]parameters.Fingerprint, fingerprintsOrdering []string) float64 {
	//mainGp := dbm.GM.GetGroup(group)
	//gp := dbm.GM.GetGroup(group).Get()
	//defer mainGp

	for loc := range gp.Get_NetworkLocs()[n] {
		gp.Results[n].TotalLocations[loc] = 0
		gp.Results[n].CorrectLocations[loc] = 0
		gp.Results[n].Accuracy[loc] = 0
		gp.Results[n].Guess[loc] = make(map[string]int)
	}

	for _, v1 := range fingerprintsOrdering {

		v2 := fingerprintsInMemory[v1]
		if len(v2.WifiFingerprint) == 0 {
			continue
		}
		if _, ok := gp.NetworkLocs[n][v2.Location]; ok {
			locationGuess, _ := CalculatePosterior(v2, gp)
			gp.Results[n].TotalLocations[v2.Location]++ //set TotalLocations
			if locationGuess == v2.Location {
				gp.Results[n].CorrectLocations[v2.Location]++ //set CorrectLocations
			}
			if _, ok := gp.Results[n].Guess[v2.Location]; !ok {
				gp.Results[n].Guess[v2.Location] = make(map[string]int)
			}
			if _, ok := gp.Results[n].Guess[v2.Location][locationGuess]; !ok {
				gp.Results[n].Guess[v2.Location][locationGuess] = 0
			}
			gp.Results[n].Guess[v2.Location][locationGuess]++ //set Guess
		}

	}

	average := float64(0)
	for loc := range gp.NetworkLocs[n] {
		if gp.Results[n].TotalLocations[loc] > 0 {
			// fmt.Println(ps.Results[n].CorrectLocations[loc], ps.Results[n].TotalLocations[loc])
			// set Accuracy
			gp.Results[n].Accuracy[loc] = int(100.0 * gp.Results[n].CorrectLocations[loc] / gp.Results[n].TotalLocations[loc])
			average += float64(gp.Results[n].Accuracy[loc])
		}
	}
	average = average / float64(len(gp.NetworkLocs[n]))

	return average
}

// calculatePriors generates the ps.Prior(P,NP,MacFreq,NMacFreq) data and ps.MacVariability for Naive-Bayes classification. Now deprecated, use calculatePriorsThreaded instead.
//todo: write calculatePriorsThreaded function
func calculatePriors(group string, gp *dbm.Group, fingerprintsInMemory map[string]parameters.Fingerprint, fingerprintsOrdering []string) {
	// defer timeTrack(time.Now(), "calculatePriors")
	gp.Priors = make(map[string]parameters.PriorParameters)
	for n := range gp.NetworkLocs {
		var newPrior = *parameters.NewPriorParameters()
		gp.Priors[n] = newPrior
	}

	// Initialization
	gp.MacVariability = make(map[string]float32)
	for n := range gp.Priors {
		gp.Priors[n].Special["MacFreqMin"] = float64(100)
		gp.Priors[n].Special["NMacFreqMin"] = float64(100)
		for loc := range gp.NetworkLocs[n] {
			gp.Priors[n].P[loc] = make(map[string][]float32)
			gp.Priors[n].NP[loc] = make(map[string][]float32)
			gp.Priors[n].MacFreq[loc] = make(map[string]float32)
			gp.Priors[n].NMacFreq[loc] = make(map[string]float32)
			for mac := range gp.NetworkMacs[n] {
				gp.Priors[n].P[loc][mac] = make([]float32, RssiPartitions)
				gp.Priors[n].NP[loc][mac] = make([]float32, RssiPartitions)
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

		// todo: gp is set in the getParameters function (getParameters is called before calculatePriors), so calling the hasNetwork function returns true
		networkName, inNetwork := clustring.HasNetwork(gp.NetworkMacs, macs)
		if inNetwork {
			for _, router := range v2.WifiFingerprint {
				if router.Rssi > glb.MinRssiOpt {
					//fmt.Println(router.Rssi)
					gp.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-glb.MinRssi] += PdfType[0]
					//make the real probability of the rssi distribution
					for i, val := range PdfType {
						if i > 0 {
							//if (router.Rssi-MinRssi-i<2) {
							//	fmt.Println("i=", i)
							//	fmt.Println("router.Rssi=", router.Rssi)
							//	fmt.Println("router.rssi-MinRSSi-i=", router.Rssi-MinRssi-i)
							//}
							if (router.Rssi-glb.MinRssi-i > 0 && router.Rssi-glb.MinRssi+i < RssiPartitions) {
								gp.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-glb.MinRssi-i] += val
								gp.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-glb.MinRssi+i] += val
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
	for n := range gp.Priors {
		for locN := range gp.NetworkLocs[n] {
			for loc := range gp.NetworkLocs[n] {
				if loc != locN {
					for mac := range gp.NetworkMacs[n] {
						for i := range gp.Priors[n].P[locN][mac] {
							//i is rssi
							if gp.Priors[n].P[loc][mac][i] > 0 {
								gp.Priors[n].NP[locN][mac][i] += gp.Priors[n].P[loc][mac][i]
							}
						}
					}
				}
			}
		}
	}

	// Add in absentee, normalize P and nP and determine MacVariability

	for n := range gp.Priors {
		macAverages := make(map[string][]float32)

		for loc := range gp.NetworkLocs[n] {
			for mac := range gp.NetworkMacs[n] {
				for i := range gp.Priors[n].P[loc][mac] { //i is rssi
					//why using Absentee instead of 0
					gp.Priors[n].P[loc][mac][i] += Absentee
					gp.Priors[n].NP[loc][mac][i] += Absentee
				}
				total := float32(0) //total = sum of probabilities(P) of all rssi for a specific mac and location
				for _, val := range gp.Priors[n].P[loc][mac] {
					total += val
				}
				averageMac := float32(0)
				for i, val := range gp.Priors[n].P[loc][mac] {
					if val > float32(0) { //val is always => Absentee >0 --> it is required in normalization
						gp.Priors[n].P[loc][mac][i] = val / total                    //normalizing P
						averageMac += glb.RssiRange[i] * gp.Priors[n].P[loc][mac][i] // RssiRange[i] equals to rssi.
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
				for i := range gp.Priors[n].NP[loc][mac] {
					total += gp.Priors[n].NP[loc][mac][i]
				}
				if total > 0 {
					for i := range gp.Priors[n].NP[loc][mac] {
						gp.Priors[n].NP[loc][mac][i] = gp.Priors[n].NP[loc][mac][i] / total
					}
				}
			}
		}

		// Determine MacVariability
		for mac := range macAverages {
			//todo: why 2?
			if len(macAverages[mac]) <= 2 {
				gp.MacVariability[mac] = float32(1)
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
				gp.MacVariability[mac] = glb.StandardDeviation(macAverages[mac]) //refer to line 300 todo
			}
		}
	}

	// Determine mac frequencies and normalize
	for n := range gp.Priors {
		for loc := range gp.NetworkLocs[n] {
			maxCount := 0
			for mac := range gp.MacCountByLoc[loc] {
				if gp.MacCountByLoc[loc][mac] > maxCount {
					maxCount = gp.MacCountByLoc[loc][mac] //maxCount:repeat number of the most seen mac in a location

				}
			}
			//fmt.Println("MAX COUNT:", maxCount)
			for mac := range gp.MacCountByLoc[loc] {
				//if a mac is not seen in a location, the macFreq of that mac equals to 0 (gp.MacCountByLoc[loc][mac]).
				//todo: Does the above mentioned 0 value make some error in the bayesian function?
				gp.Priors[n].MacFreq[loc][mac] = float32(gp.MacCountByLoc[loc][mac]) / float32(maxCount)
				//fmt.Println("mac freq:", gp.Priors[n].MacFreq[loc][mac])
				if float64(gp.Priors[n].MacFreq[loc][mac]) < gp.Priors[n].Special["MacFreqMin"] {
					gp.Priors[n].Special["MacFreqMin"] = float64(gp.Priors[n].MacFreq[loc][mac])
				}
			}
		}
	}

	// Determine negative mac frequencies and normalize
	for n := range gp.Priors {
		for loc1 := range gp.Priors[n].MacFreq {
			sum := float32(0)
			for loc2 := range gp.Priors[n].MacFreq {
				if loc2 != loc1 {
					for mac := range gp.Priors[n].MacFreq[loc2] {
						gp.Priors[n].NMacFreq[loc1][mac] += gp.Priors[n].MacFreq[loc2][mac]
					}
					sum++
				}
			}
			// sum = i(i-1); i = gp.NetworkLocs[n]
			// Normalize
			//Done: it seems that sum is not calculated correctly. It should be equals to "number of locations-1"
			if sum > 0 {
				for mac := range gp.Priors[n].MacFreq[loc1] {
					gp.Priors[n].NMacFreq[loc1][mac] = gp.Priors[n].NMacFreq[loc1][mac] / sum
					if float64(gp.Priors[n].NMacFreq[loc1][mac]) < gp.Priors[n].Special["NMacFreqMin"] {
						gp.Priors[n].Special["NMacFreqMin"] = float64(gp.Priors[n].NMacFreq[loc1][mac])
					}
				}
			}
		}
	}
	//todo: the default values for MixIn and Cutoff should be set as initial values not hardcoded values
	for n := range gp.Priors {
		gp.Priors[n].Special["MixIn"] = 0.5
		//todo: spell check for Varability!
		gp.Priors[n].Special["VarabilityCutoff"] = 0
	}


	////#cache
	//gp := dbm.GM.GetGroup(group)
	//
	//gp.Set_Priors(gp.Priors)
	//gp.Set_MacVariability(gp.MacVariability)
}

func calculateGaussianPriors(group string, gp *dbm.Group, fingerprintsInMemory map[string]parameters.Fingerprint, fingerprintsOrdering []string) {
	// defer timeTrack(time.Now(), "calculatePriors")
	gp.Priors = make(map[string]parameters.PriorParameters)
	for n := range gp.NetworkLocs {
		var newPrior = *parameters.NewPriorParameters()
		gp.Priors[n] = newPrior
	}

	// Initialization
	Rssies := make(map[string]map[string][]float64)
	RssiesVariance := make(map[string]map[string]float64)
	RssiesAvg := make(map[string]map[string]float64)

	gp.MacVariability = make(map[string]float32)
	for n := range gp.Priors {
		gp.Priors[n].Special["MacFreqMin"] = float64(100)
		gp.Priors[n].Special["NMacFreqMin"] = float64(100)
		for loc := range gp.NetworkLocs[n] {
			gp.Priors[n].P[loc] = make(map[string][]float32)

			Rssies[loc] = make(map[string][]float64)
			RssiesVariance[loc] = make(map[string]float64)
			RssiesAvg[loc] = make(map[string]float64)

			gp.Priors[n].NP[loc] = make(map[string][]float32)
			gp.Priors[n].MacFreq[loc] = make(map[string]float32)
			gp.Priors[n].NMacFreq[loc] = make(map[string]float32)
			for mac := range gp.NetworkMacs[n] {
				gp.Priors[n].P[loc][mac] = make([]float32, RssiPartitions)

				Rssies[loc][mac] = make([]float64, 0)
				RssiesVariance[loc][mac] = float64(0)
				RssiesAvg[loc][mac] = float64(0)

				gp.Priors[n].NP[loc][mac] = make([]float32, RssiPartitions)
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
		_, inNetwork := clustring.HasNetwork(gp.NetworkMacs, macs)
		if inNetwork {
			for _, router := range v2.WifiFingerprint {
				if router.Rssi > glb.MinRssiOpt {
					//fmt.Println(router.Rssi)
					Rssies[v2.Location][router.Mac] = append(Rssies[v2.Location][router.Mac], float64(router.Rssi-glb.MinRssi))
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
			RssiesAvg[loc][mac] = glb.Average64(Rssies[loc][mac])
			RssiesVariance[loc][mac] = glb.Variance64(Rssies[loc][mac])
		}
	}

	//fmt.Println("###")

	g := NewGaussian(0, 1)

	// Create gaussian distribution; set probability for each rssi of each mac in each location
	for n := range gp.Priors {
		for loc := range gp.NetworkLocs[n] {
			for mac := range gp.NetworkMacs[n] {
				for rssi := 0; rssi < len(glb.RssiRange); rssi++ {
					if (RssiesVariance[loc][mac] == 0) {
						g = NewGaussian(RssiesAvg[loc][mac], 1)
					} else {
						g = NewGaussian(RssiesAvg[loc][mac], RssiesVariance[loc][mac])
					}
					//fmt.Println(float32(g.Pdf(float64(rssi))))
					//fmt.Println(loc)
					//fmt.Println(mac)
					//fmt.Println(rssi)
					gp.Priors[n].P[loc][mac][rssi] = float32(g.Pdf(float64(rssi)))
				}
			}
		}
	}

	// Calculate the nP
	for n := range gp.Priors {
		for locN := range gp.NetworkLocs[n] {
			for loc := range gp.NetworkLocs[n] {
				if loc != locN {
					for mac := range gp.NetworkMacs[n] {
						for i := range gp.Priors[n].P[locN][mac] {
							//i is rssi
							if gp.Priors[n].P[loc][mac][i] > 0 {
								gp.Priors[n].NP[locN][mac][i] += gp.Priors[n].P[loc][mac][i]
							}
						}
					}
				}
			}
		}
	}

	// Add in absentee, normalize P and nP and determine MacVariability

	for n := range gp.Priors {
		macAverages := make(map[string][]float32)

		for loc := range gp.NetworkLocs[n] {
			for mac := range gp.NetworkMacs[n] {
				for i := range gp.Priors[n].P[loc][mac] { //i is rssi
					//why using Absentee instead of 0
					gp.Priors[n].P[loc][mac][i] += Absentee
					gp.Priors[n].NP[loc][mac][i] += Absentee
				}
				total := float32(0) //total = sum of probabilities(P) of all rssi for a specific mac and location
				for _, val := range gp.Priors[n].P[loc][mac] {
					total += val
				}
				averageMac := float32(0)
				for i, val := range gp.Priors[n].P[loc][mac] {
					if val > float32(0) { //val is always => Absentee >0 --> it is required in normalization
						gp.Priors[n].P[loc][mac][i] = val / total                //normalizing P
						averageMac += glb.RssiRange[i] * gp.Priors[n].P[loc][mac][i] // RssiRange[i] equals to rssi.
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
				for i := range gp.Priors[n].NP[loc][mac] {
					total += gp.Priors[n].NP[loc][mac][i]
				}
				if total > 0 {
					for i := range gp.Priors[n].NP[loc][mac] {
						gp.Priors[n].NP[loc][mac][i] = gp.Priors[n].NP[loc][mac][i] / total
					}
				}
			}
		}

		// Determine MacVariability
		for mac := range macAverages {
			//todo: why 2?
			if len(macAverages[mac]) <= 2 {
				gp.MacVariability[mac] = float32(1)
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
				gp.MacVariability[mac] = glb.StandardDeviation(macAverages[mac]) //refer to line 300 todo
			}
		}
	}

	// Determine mac frequencies and normalize
	for n := range gp.Priors {
		for loc := range gp.NetworkLocs[n] {
			maxCount := 0
			for mac := range gp.MacCountByLoc[loc] {
				if gp.MacCountByLoc[loc][mac] > maxCount {
					maxCount = gp.MacCountByLoc[loc][mac] //maxCount:repeat number of the most seen mac in a location

				}
			}
			//fmt.Println("MAX COUNT:", maxCount)
			for mac := range gp.MacCountByLoc[loc] {
				//if a mac is not seen in a location, the macFreq of that mac equals to 0 (ps.MacCountByLoc[loc][mac]).
				//todo: Does the above mentioned 0 value make some error in the bayesian function?
				gp.Priors[n].MacFreq[loc][mac] = float32(gp.MacCountByLoc[loc][mac]) / float32(maxCount)
				//fmt.Println("mac freq:", ps.Priors[n].MacFreq[loc][mac])
				if float64(gp.Priors[n].MacFreq[loc][mac]) < gp.Priors[n].Special["MacFreqMin"] {
					gp.Priors[n].Special["MacFreqMin"] = float64(gp.Priors[n].MacFreq[loc][mac])
				}
			}
		}
	}

	// Determine negative mac frequencies and normalize
	for n := range gp.Priors {
		for loc1 := range gp.Priors[n].MacFreq {
			sum := float32(0)
			for loc2 := range gp.Priors[n].MacFreq {
				if loc2 != loc1 {
					for mac := range gp.Priors[n].MacFreq[loc2] {
						gp.Priors[n].NMacFreq[loc1][mac] += gp.Priors[n].MacFreq[loc2][mac]
					}
					sum++
				}
			}
			// sum = i(i-1); i = ps.NetworkLocs[n]
			// Normalize
			//Done: it seems that sum is not calculated correctly. It should be equals to "number of locations-1"
			if sum > 0 {
				for mac := range gp.Priors[n].MacFreq[loc1] {
					gp.Priors[n].NMacFreq[loc1][mac] = gp.Priors[n].NMacFreq[loc1][mac] / sum
					if float64(gp.Priors[n].NMacFreq[loc1][mac]) < gp.Priors[n].Special["NMacFreqMin"] {
						gp.Priors[n].Special["NMacFreqMin"] = float64(gp.Priors[n].NMacFreq[loc1][mac])
					}
				}
			}
		}
	}
	//todo: the default values for MixIn and Cutoff should be set as initial values not hardcoded values
	for n := range gp.Priors {
		gp.Priors[n].Special["MixIn"] = 0.5
		//todo: spell check for Varability!
		gp.Priors[n].Special["VarabilityCutoff"] = 0
	}

}


//group: group
//ps:
//fingerprintsInMemory:
//fingerprintsOrdering:
//updates ps with the new fingerprint.
//(The Parameters which are manipulated: NetworkMacs,NetworkLocs,UniqueMacs,UniqueLocs,MacCount,MacCountByLoc and Loaded)
func GetParameters(group string, gp *dbm.Group, fingerprintsInMemory map[string]parameters.Fingerprint, fingerprintsOrdering []string) {

	persistentPs, err := dbm.OpenPersistentParameters(group) //persistentPs is just like ps but with renamed network name; e.g.: "0" -> "1"
	if err != nil {
		//log.Fatal(err)
		glb.Error.Println(err)
	}



	//glb.Error.Println("d")
	gp.NetworkMacs = make(map[string]map[string]bool)
	gp.NetworkLocs = make(map[string]map[string]bool)
	gp.UniqueMacs = []string{}
	gp.UniqueLocs = []string{}
	gp.MacCount = make(map[string]int)
	gp.MacCountByLoc = make(map[string]map[string]int)
	//gp.Loaded = true
	//opening the db
	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//defer db.Close()
	// if err != nil {
	//	log.Fatal(err)
	//}


	macs := []string{}

	// Get all parameters that don't need a network graph (?)
	for _, v1 := range fingerprintsOrdering {

		//log.Println("calculateResults=true")
		v2 := fingerprintsInMemory[v1]

		// append the fingerprint location to UniqueLocs array if doesn't exist in it.
		if !glb.StringInSlice(v2.Location, gp.UniqueLocs) {
			gp.UniqueLocs = append(gp.UniqueLocs, v2.Location)
		}

		// MacCountByLoc initialization for new location
		if _, ok := gp.MacCountByLoc[v2.Location]; !ok {
			gp.MacCountByLoc[v2.Location] = make(map[string]int)
		}

		//// building network
		//macs := []string{}

		for _, router := range v2.WifiFingerprint {
			// building network
			macs = append(macs, router.Mac)

			// append the fingerprint mac to UniqueMacs array if doesn't exist in it.
			if !glb.StringInSlice(router.Mac, gp.UniqueMacs) {
				gp.UniqueMacs = append(gp.UniqueMacs, router.Mac)
			}

			// mac count
			if _, ok := gp.MacCount[router.Mac]; !ok {
				gp.MacCount[router.Mac] = 0
			}
			gp.MacCount[router.Mac]++

			// mac by location count
			if _, ok := gp.MacCountByLoc[v2.Location][router.Mac]; !ok {
				gp.MacCountByLoc[v2.Location][router.Mac] = 0
			}
			gp.MacCountByLoc[v2.Location][router.Mac]++
		}

		// building network
		//ps.NetworkMacs = buildNetwork(ps.NetworkMacs, macs)
	}
	// todo: network definition and buildNetwork() must be redefined
	gp.NetworkMacs = clustring.BuildNetwork(gp.NetworkMacs, macs)
	gp.NetworkMacs = clustring.MergeNetwork(gp.NetworkMacs)

	//Error.Println("ps.Networkmacs", ps.NetworkMacs)
	// Rename the NetworkMacs
	if len(persistentPs.NetworkRenamed) > 0 {
		newNames := []string{}
		for k := range persistentPs.NetworkRenamed {
			newNames = append(newNames, k)

		}
		//todo: \/ wtf? Rename procedure could be redefined better.
		for n := range gp.NetworkMacs {
			renamed := false
			for mac := range gp.NetworkMacs[n] {
				for renamedN := range persistentPs.NetworkRenamed {
					if glb.StringInSlice(mac, persistentPs.NetworkRenamed[renamedN]) && !glb.StringInSlice(n, newNames) {
						gp.NetworkMacs[renamedN] = make(map[string]bool)
						for k, v := range gp.NetworkMacs[n] {
							gp.NetworkMacs[renamedN][k] = v //copy ps.NetworkMacs[n] to ps.NetworkMacs[renamedN]
						}
						delete(gp.NetworkMacs, n)
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
		networkName, inNetwork := clustring.HasNetwork(gp.NetworkMacs, macs)
		if inNetwork {
			if _, ok := gp.NetworkLocs[networkName]; !ok {
				gp.NetworkLocs[networkName] = make(map[string]bool)
			}
			if _, ok := gp.NetworkLocs[networkName][v2.Location]; !ok {
				gp.NetworkLocs[networkName][v2.Location] = true
			}
		}
	}

	////#cache
	//gp := dbm.GM.GetGroup(group)
	//
	//gp.Set_NetworkMacs(ps.NetworkMacs)
	//gp.Set_NetworkLocs(ps.NetworkLocs)
	//gp.Set_UniqueLocs(ps.UniqueLocs)
	//gp.Set_UniqueMacs(ps.UniqueMacs)
	//gp.Set_MacCount(ps.MacCount)
	//gp.Set_MacCountByLoc(ps.MacCountByLoc)

}
