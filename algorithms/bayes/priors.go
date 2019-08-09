//// Copyright 2015-2016 Zack Scholl. All rights reserved.
//// Use of this source code is governed by a AGPL
//// license that can be found in the LICENSE file.
//
//// priors.go contains variables for calculating priors.
//
package bayes

//
//import (
//	"ParsinServer/glb"
//	"ParsinServer/algorithms/parameters"
//	"ParsinServer/algorithms/clustering"
//	"ParsinServer/dbm"
//)
//
//// PdfType dictates the width of gaussian smoothing
//var PdfType []float32
//
//
//// RssiPartitions are the calculated number of partitions from MinRssi and MaxRssi
//var RssiPartitions int
//
//// Absentee is the base level of probability for any signal
//var Absentee float32
//
//
//// FoldCrossValidation is the amount of data left out during learning to be used in cross validation
//var FoldCrossValidation float64
//
//// Variables initialization
//func init() {
//	//todo:what is PdfType and how to find the values
//	PdfType = []float32{.1995, .1760, .1210, .0648, .027, 0.005}
//	Absentee = 1e-6
//
//	RssiPartitions = glb.MaxRssi - glb.MinRssi + 1
//	glb.RssiRange = make([]float32, RssiPartitions)
//	for i := 0; i < len(glb.RssiRange); i++ {
//		glb.RssiRange[i] = float32(glb.MinRssi + i)
//	}
//	FoldCrossValidation = 5
//}
//
//// deprecated
//func optimizePriors(group string) {
//	// generate the fingerprintsInMemory
//	//var gp = dbm.NewGroup(group)
//	//defer dbm.GM.GetGroup(group).Set(gp)
//	gp1 := dbm.GM.GetGroup(group)
//
//	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
//	var fingerprintsOrdering []string
//	var err error
//
//	fingerprintsOrdering,fingerprintsInMemory,err = dbm.GetLearnFingerPrints(group,true)
//	if err != nil{
//		return
//	}
//
//	//var ps = *parameters.NewFullParameters()
//	GetParameters(group, gp, fingerprintsInMemory, fingerprintsOrdering)
//	if glb.RuntimeArgs.GaussianDist {
//		calculateGaussianPriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
//	} else {
//		calculatePriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
//	}
//
//	// fmt.Println(string(dumpParameters(ps)))
//	// ps, _ = openParameters("findtest")
//	var results = *parameters.NewResultsParameters()
//	for n := range gp.BayesPriors {
//		gp.BayesResults[n] = results
//	}
//	// fmt.Println(ps.BayesResults)
//	// ps.BayesPriors["0"].Special["MixIn"] = 1.0
//	// fmt.Println(crossValidation(group, "0", &ps))
//	// fmt.Println(ps.BayesResults)
//
//	// loop through these parameters
//	mixins := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
//	cutoffs := []float64{0.005}
//
//	for n := range gp.BayesPriors {
//		bestResult := float64(0)
//		bestMixin := float64(0)
//		bestCutoff := float64(0)
//		for _, cutoff := range cutoffs {
//			for _, mixin := range mixins {
//				gp.BayesPriors[n].Special["MixIn"] = mixin
//				gp.BayesPriors[n].Special["VarabilityCutoff"] = cutoff
//				avgAccuracy := crossValidation(group, n, gp, fingerprintsInMemory, fingerprintsOrdering)
//				// avgAccuracy := crossValidation(group, n, &ps)
//				if avgAccuracy > bestResult {
//					bestResult = avgAccuracy
//					bestCutoff = cutoff
//					bestMixin = mixin
//				}
//			}
//		}
//		gp.BayesPriors[n].Special["MixIn"] = bestMixin
//		gp.BayesPriors[n].Special["VarabilityCutoff"] = bestCutoff
//		// Final validation
//		crossValidation(group, n, gp, fingerprintsInMemory, fingerprintsOrdering)
//		// crossValidation(group, n, &ps)
//	}
//
//	//go dbm.SaveParameters(group, ps)
//	//go dbm.SetPsCache(group, ps)
//}
//
//func regenerateEverything(group string) {
//	// generate the fingerprintsInMemory
//	var gp = dbm.GM.GetGroup(group).Get()
//	defer dbm.GM.GetGroup(group).Set(gp)
//
//	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
//	var fingerprintsOrdering []string
//	var err error
//
//	fingerprintsOrdering,fingerprintsInMemory,err = dbm.GetLearnFingerPrints(group,true)
//	if err != nil{
//		return
//	}
//
//	//var ps = *parameters.NewFullParameters()
//	//ps, _ = dbm.OpenParameters(group)
//	GetParameters(group, gp, fingerprintsInMemory, fingerprintsOrdering)//openParameters is only called here.
//	if glb.RuntimeArgs.GaussianDist {
//		calculateGaussianPriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
//	} else {
//		calculatePriors(group, gp, fingerprintsInMemory, fingerprintsOrdering)
//	}
//	var results = *parameters.NewResultsParameters()
//	for n := range gp.BayesPriors {
//		gp.BayesResults[n] = results
//	}
//	for n := range gp.BayesPriors {
//		crossValidation(group, n, gp, fingerprintsInMemory, fingerprintsOrdering)
//	}
//	//dbm.SaveParameters(group, ps)
//}
//
//// (1/FoldCrossValidation) of the learned fingerprints are predicted with ps data, then results are wrote in ps.BayesResults
//func crossValidation(group string, n string,gp *dbm.Group, fingerprintsInMemory map[string]parameters.Fingerprint, fingerprintsOrdering []string) float64 {
//	//mainGp := dbm.GM.GethGroup(group)
//	//gp := dbm.GM.GetGroup(group).Get()
//	//defer mainGp
//
//	for loc := range gp.Get_NetworkLocs()[n] {
//		gp.BayesResults[n].TotalLocations[loc] = 0
//		gp.BayesResults[n].CorrectLocations[loc] = 0
//		gp.BayesResults[n].Accuracy[loc] = 0
//		gp.BayesResults[n].Guess[loc] = make(map[string]int)
//	}
//
//	for _, v1 := range fingerprintsOrdering {
//
//		v2 := fingerprintsInMemory[v1]
//		if len(v2.WifiFingerprint) == 0 {
//			continue
//		}
//		if _, ok := gp.NetworkLocs[n][v2.Location]; ok {
//			locationGuess, _ := CalculatePosterior(v2, gp)
//			gp.BayesResults[n].TotalLocations[v2.Location]++ //set TotalLocations
//			if locationGuess == v2.Location {
//				gp.BayesResults[n].CorrectLocations[v2.Location]++ //set CorrectLocations
//			}
//			if _, ok := gp.BayesResults[n].Guess[v2.Location]; !ok {
//				gp.BayesResults[n].Guess[v2.Location] = make(map[string]int)
//			}
//			if _, ok := gp.BayesResults[n].Guess[v2.Location][locationGuess]; !ok {
//				gp.BayesResults[n].Guess[v2.Location][locationGuess] = 0
//			}
//			gp.BayesResults[n].Guess[v2.Location][locationGuess]++ //set Guess
//		}
//
//	}
//
//	average := float64(0)
//	for loc := range gp.NetworkLocs[n] {
//		if gp.BayesResults[n].TotalLocations[loc] > 0 {
//			// fmt.Println(ps.BayesResults[n].CorrectLocations[loc], ps.BayesResults[n].TotalLocations[loc])
//			// set Accuracy
//			gp.BayesResults[n].Accuracy[loc] = int(100.0 * gp.BayesResults[n].CorrectLocations[loc] / gp.BayesResults[n].TotalLocations[loc])
//			average += float64(gp.BayesResults[n].Accuracy[loc])
//		}
//	}
//	average = average / float64(len(gp.NetworkLocs[n]))
//
//	return average
//}
//
//// calculatePriors generates the ps.Prior(P,NP,MacFreq,NMacFreq) data and ps.MacVariability for Naive-Bayes classification. Now deprecated, use calculatePriorsThreaded instead.
////todo: write calculatePriorsThreaded function
//func calculatePriors(fingerprintsInMemory map[string]parameters.Fingerprint, fingerprintsOrdering []string, md *dbm.MiddleDataStruct, bayesPriors *map[string]parameters.PriorParameters) {
//	// defer timeTrack(time.Now(), "calculatePriors")
//	BayesPriors := *bayesPriors
//
//
//	for n := range md.NetworkLocs {
//		var newPrior = *parameters.NewPriorParameters()
//		BayesPriors[n] = newPrior
//	}
//
//	// Initialization
//	md.MacVariability = make(map[string]float32)
//	for n := range BayesPriors {
//		BayesPriors[n].Special["MacFreqMin"] = float64(100)
//		BayesPriors[n].Special["NMacFreqMin"] = float64(100)
//		for loc := range md.NetworkLocs[n] {
//			BayesPriors[n].P[loc] = make(map[string][]float32)
//			BayesPriors[n].NP[loc] = make(map[string][]float32)
//			BayesPriors[n].MacFreq[loc] = make(map[string]float32)
//			BayesPriors[n].NMacFreq[loc] = make(map[string]float32)
//			for mac := range md.NetworkMacs[n] {
//				BayesPriors[n].P[loc][mac] = make([]float32, RssiPartitions)
//				BayesPriors[n].NP[loc][mac] = make([]float32, RssiPartitions)
//			}
//		}
//	}
//
//	//create gaussian distribution for every mac in every location
//
//	for _, v1 := range fingerprintsOrdering {
//
//		v2 := fingerprintsInMemory[v1]
//		macs := []string{}
//		for _, router := range v2.WifiFingerprint {
//			macs = append(macs, router.Mac)
//		}
//
//		// todo: gp is set in the getParameters function (getParameters is called before calculatePriors), so calling the hasNetwork function returns true
//		networkName, inNetwork := clustring.HasNetwork(md.NetworkMacs, macs)
//		if inNetwork {
//			for _, router := range v2.WifiFingerprint {
//				if router.Rssi > glb.MinRssiOpt {
//					//fmt.Println(router.Rssi)
//					BayesPriors[networkName].P[v2.Location][router.Mac][router.Rssi-glb.MinRssi] += PdfType[0]
//					//make the real probability of the rssi distribution
//					for i, val := range PdfType {
//						if i > 0 {
//							//if (router.Rssi-MinRssi-i<2) {
//							//	fmt.Println("i=", i)
//							//	fmt.Println("router.Rssi=", router.Rssi)
//							//	fmt.Println("router.rssi-MinRSSi-i=", router.Rssi-MinRssi-i)
//							//}
//							if (router.Rssi-glb.MinRssi-i > 0 && router.Rssi-glb.MinRssi+i < RssiPartitions) {
//								BayesPriors[networkName].P[v2.Location][router.Mac][router.Rssi-glb.MinRssi-i] += val
//								BayesPriors[networkName].P[v2.Location][router.Mac][router.Rssi-glb.MinRssi+i] += val
//							}
//
//						}
//					}
//					//} else {
//					//	Warning.Println(router.Rssi)
//				}
//			}
//		}
//
//	}
//
//	// Calculate the nP
//	for n := range BayesPriors {
//		for locN := range md.NetworkLocs[n] {
//			for loc := range md.NetworkLocs[n] {
//				if loc != locN {
//					for mac := range md.NetworkMacs[n] {
//						for i := range BayesPriors[n].P[locN][mac] {
//							//i is rssi
//							if BayesPriors[n].P[loc][mac][i] > 0 {
//								BayesPriors[n].NP[locN][mac][i] += BayesPriors[n].P[loc][mac][i]
//							}
//						}
//					}
//				}
//			}
//		}
//	}
//
//	// Add in absentee, normalize P and nP and determine MacVariability
//
//	for n := range BayesPriors {
//		macAverages := make(map[string][]float32)
//
//		for loc := range md.NetworkLocs[n] {
//			for mac := range md.NetworkMacs[n] {
//				for i := range BayesPriors[n].P[loc][mac] { //i is rssi
//					//why using Absentee instead of 0
//					BayesPriors[n].P[loc][mac][i] += Absentee
//					BayesPriors[n].NP[loc][mac][i] += Absentee
//				}
//				total := float32(0) //total = sum of probabilities(P) of all rssi for a specific mac and location
//				for _, val := range BayesPriors[n].P[loc][mac] {
//					total += val
//				}
//				averageMac := float32(0)
//				for i, val := range BayesPriors[n].P[loc][mac] {
//					if val > float32(0) { //val is always => Absentee >0 --> it is required in normalization
//						BayesPriors[n].P[loc][mac][i] = val / total                    //normalizing P
//						averageMac += glb.RssiRange[i] * BayesPriors[n].P[loc][mac][i] // RssiRange[i] equals to rssi.
//						//todo: average mac is not valid if the probability distribution (P) is not a standard gaussian function,e.g. has two peaks
//					}
//				}
//				//why checking is required?
//				if averageMac < float32(0) {
//					if _, ok := macAverages[mac]; !ok {
//						macAverages[mac] = []float32{}
//					}
//					macAverages[mac] = append(macAverages[mac], averageMac) // averageMac of each mac in every locations
//				}
//
//				//normalizing NP
//				total = float32(0)
//				for i := range BayesPriors[n].NP[loc][mac] {
//					total += BayesPriors[n].NP[loc][mac][i]
//				}
//				if total > 0 {
//					for i := range BayesPriors[n].NP[loc][mac] {
//						BayesPriors[n].NP[loc][mac][i] = BayesPriors[n].NP[loc][mac][i] / total
//					}
//				}
//			}
//		}
//
//		// Determine MacVariability
//		for mac := range macAverages {
//			//todo: why 2?
//			if len(macAverages[mac]) <= 2 {
//				md.MacVariability[mac] = float32(1)
//			} else {
//				maxVal := float32(-10000)
//				for _, val := range macAverages[mac] {
//					if val > maxVal {
//						maxVal = val
//					}
//				}
//				for i, val := range macAverages[mac] {
//					//todo: why not using the actual values of macAverages instead of the normalized values?
//					macAverages[mac][i] = maxVal / val // normalization(because val is < 0, we use maxVal/val instead of val /maxVal)
//				}
//				// MacVariability shows the standard deviation of a specific AP in all locations
//				md.MacVariability[mac] = glb.StandardDeviation(macAverages[mac]) //refer to line 300 todo
//			}
//		}
//	}
//
//	// Determine mac frequencies and normalize
//	for n := range BayesPriors {
//		for loc := range md.NetworkLocs[n] {
//			maxCount := 0
//			for mac := range md.MacCountByLoc[loc] {
//				if md.MacCountByLoc[loc][mac] > maxCount {
//					maxCount = md.MacCountByLoc[loc][mac] //maxCount:repeat number of the most seen mac in a location
//
//				}
//			}
//			//fmt.Println("MAX COUNT:", maxCount)
//			for mac := range md.MacCountByLoc[loc] {
//				//if a mac is not seen in a location, the macFreq of that mac equals to 0 (gp.MacCountByLoc[loc][mac]).
//				//todo: Does the above mentioned 0 value make some error in the bayesian function?
//				BayesPriors[n].MacFreq[loc][mac] = float32(md.MacCountByLoc[loc][mac]) / float32(maxCount)
//				//fmt.Println("mac freq:", gp.BayesPriors[n].MacFreq[loc][mac])
//				if float64(BayesPriors[n].MacFreq[loc][mac]) < BayesPriors[n].Special["MacFreqMin"] {
//					BayesPriors[n].Special["MacFreqMin"] = float64(BayesPriors[n].MacFreq[loc][mac])
//				}
//			}
//		}
//	}
//
//	// Determine negative mac frequencies and normalize
//	for n := range BayesPriors {
//		for loc1 := range BayesPriors[n].MacFreq {
//			sum := float32(0)
//			for loc2 := range BayesPriors[n].MacFreq {
//				if loc2 != loc1 {
//					for mac := range BayesPriors[n].MacFreq[loc2] {
//						BayesPriors[n].NMacFreq[loc1][mac] += BayesPriors[n].MacFreq[loc2][mac]
//					}
//					sum++
//				}
//			}
//			// sum = i(i-1); i = gp.NetworkLocs[n]
//			// Normalize
//			//Done: it seems that sum is not calculated correctly. It should be equals to "number of locations-1"
//			if sum > 0 {
//				for mac := range BayesPriors[n].MacFreq[loc1] {
//					BayesPriors[n].NMacFreq[loc1][mac] = BayesPriors[n].NMacFreq[loc1][mac] / sum
//					if float64(BayesPriors[n].NMacFreq[loc1][mac]) < BayesPriors[n].Special["NMacFreqMin"] {
//						BayesPriors[n].Special["NMacFreqMin"] = float64(BayesPriors[n].NMacFreq[loc1][mac])
//					}
//				}
//			}
//		}
//	}
//	//todo: the default values for MixIn and Cutoff should be set as initial values not hardcoded values
//	for n := range BayesPriors {
//		BayesPriors[n].Special["MixIn"] = 0.5
//		//todo: spell check for Varability!
//		BayesPriors[n].Special["VarabilityCutoff"] = 0
//	}
//
//	*bayesPriors = BayesPriors
//	////#cache
//	//gp := dbm.GM.GetGroup(group)
//	//
//	//gp.Set_Priors(gp.BayesPriors)
//	//gp.Set_MacVariability(gp.MacVariability)
//}
//
//func calculateGaussianPriors(fingerprintsInMemory map[string]parameters.Fingerprint, fingerprintsOrdering []string, md *dbm.MiddleDataStruct, bayesPriors *map[string]parameters.PriorParameters) {
//	// defer timeTrack(time.Now(), "calculatePriors")
//	BayesPriors := *bayesPriors
//
//	for n := range md.NetworkLocs {
//		var newPrior = *parameters.NewPriorParameters()
//		BayesPriors[n] = newPrior
//	}
//
//	// Initialization
//	Rssies := make(map[string]map[string][]float64)
//	RssiesVariance := make(map[string]map[string]float64)
//	RssiesAvg := make(map[string]map[string]float64)
//
//	md.MacVariability = make(map[string]float32)
//	for n := range BayesPriors {
//		BayesPriors[n].Special["MacFreqMin"] = float64(100)
//		BayesPriors[n].Special["NMacFreqMin"] = float64(100)
//		for loc := range md.NetworkLocs[n] {
//			BayesPriors[n].P[loc] = make(map[string][]float32)
//
//			Rssies[loc] = make(map[string][]float64)
//			RssiesVariance[loc] = make(map[string]float64)
//			RssiesAvg[loc] = make(map[string]float64)
//
//			BayesPriors[n].NP[loc] = make(map[string][]float32)
//			BayesPriors[n].MacFreq[loc] = make(map[string]float32)
//			BayesPriors[n].NMacFreq[loc] = make(map[string]float32)
//			for mac := range md.NetworkMacs[n] {
//				BayesPriors[n].P[loc][mac] = make([]float32, RssiPartitions)
//
//				Rssies[loc][mac] = make([]float64, 0)
//				RssiesVariance[loc][mac] = float64(0)
//				RssiesAvg[loc][mac] = float64(0)
//
//				BayesPriors[n].NP[loc][mac] = make([]float32, RssiPartitions)
//			}
//		}
//	}
//
//	//create gaussian distribution for every mac in every location
//
//	// create list of collected rssi according to the locations and MACs
//	for _, v1 := range fingerprintsOrdering {
//		v2 := fingerprintsInMemory[v1]
//		macs := []string{}
//		for _, router := range v2.WifiFingerprint {
//			macs = append(macs, router.Mac)
//		}
//		_, inNetwork := clustring.HasNetwork(md.NetworkMacs, macs)
//		if inNetwork {
//			for _, router := range v2.WifiFingerprint {
//				if router.Rssi > glb.MinRssiOpt {
//					//fmt.Println(router.Rssi)
//					Rssies[v2.Location][router.Mac] = append(Rssies[v2.Location][router.Mac], float64(router.Rssi-glb.MinRssi))
//				}
//			}
//
//		}
//	}
//
//	// Calculate average and variance of a rssi list of a mac in a location
//	for loc := range Rssies {
//		for mac := range Rssies[loc] {
//			//fmt.Println("RSSIes for loc:",loc,"& mac:",mac)
//			//fmt.Println(Rssies[loc][mac])
//			//fmt.Println("######")
//			RssiesAvg[loc][mac] = glb.Average64(Rssies[loc][mac])
//			RssiesVariance[loc][mac] = glb.Variance64(Rssies[loc][mac])
//		}
//	}
//
//	//fmt.Println("###")
//
//	g := NewGaussian(0, 1)
//
//	// Create gaussian distribution; set probability for each rssi of each mac in each location
//	for n := range BayesPriors {
//		for loc := range md.NetworkLocs[n] {
//			for mac := range md.NetworkMacs[n] {
//				for rssi := 0; rssi < len(glb.RssiRange); rssi++ {
//					if (RssiesVariance[loc][mac] == 0) {
//						g = NewGaussian(RssiesAvg[loc][mac], 1)
//					} else {
//						g = NewGaussian(RssiesAvg[loc][mac], RssiesVariance[loc][mac])
//					}
//					//fmt.Println(float32(g.Pdf(float64(rssi))))
//					//fmt.Println(loc)
//					//fmt.Println(mac)
//					//fmt.Println(rssi)
//					BayesPriors[n].P[loc][mac][rssi] = float32(g.Pdf(float64(rssi)))
//				}
//			}
//		}
//	}
//
//	// Calculate the nP
//	for n := range BayesPriors {
//		for locN := range md.NetworkLocs[n] {
//			for loc := range md.NetworkLocs[n] {
//				if loc != locN {
//					for mac := range md.NetworkMacs[n] {
//						for i := range BayesPriors[n].P[locN][mac] {
//							//i is rssi
//							if BayesPriors[n].P[loc][mac][i] > 0 {
//								BayesPriors[n].NP[locN][mac][i] += BayesPriors[n].P[loc][mac][i]
//							}
//						}
//					}
//				}
//			}
//		}
//	}
//
//	// Add in absentee, normalize P and nP and determine MacVariability
//
//	for n := range BayesPriors {
//		macAverages := make(map[string][]float32)
//
//		for loc := range md.NetworkLocs[n] {
//			for mac := range md.NetworkMacs[n] {
//				for i := range BayesPriors[n].P[loc][mac] { //i is rssi
//					//why using Absentee instead of 0
//					BayesPriors[n].P[loc][mac][i] += Absentee
//					BayesPriors[n].NP[loc][mac][i] += Absentee
//				}
//				total := float32(0) //total = sum of probabilities(P) of all rssi for a specific mac and location
//				for _, val := range BayesPriors[n].P[loc][mac] {
//					total += val
//				}
//				averageMac := float32(0)
//				for i, val := range BayesPriors[n].P[loc][mac] {
//					if val > float32(0) { //val is always => Absentee >0 --> it is required in normalization
//						BayesPriors[n].P[loc][mac][i] = val / total                    //normalizing P
//						averageMac += glb.RssiRange[i] * BayesPriors[n].P[loc][mac][i] // RssiRange[i] equals to rssi.
//						//todo: average mac is not valid if the probability distribution (P) is not a standard gaussian function,e.g. has two peaks
//					}
//				}
//				//why checking is required?
//				if averageMac < float32(0) {
//					if _, ok := macAverages[mac]; !ok {
//						macAverages[mac] = []float32{}
//					}
//					macAverages[mac] = append(macAverages[mac], averageMac) // averageMac of each mac in every locations
//				}
//
//				//normalizing NP
//				total = float32(0)
//				for i := range BayesPriors[n].NP[loc][mac] {
//					total += BayesPriors[n].NP[loc][mac][i]
//				}
//				if total > 0 {
//					for i := range BayesPriors[n].NP[loc][mac] {
//						BayesPriors[n].NP[loc][mac][i] = BayesPriors[n].NP[loc][mac][i] / total
//					}
//				}
//			}
//		}
//
//		// Determine MacVariability
//		for mac := range macAverages {
//			//todo: why 2?
//			if len(macAverages[mac]) <= 2 {
//				md.MacVariability[mac] = float32(1)
//			} else {
//				maxVal := float32(-10000)
//				for _, val := range macAverages[mac] {
//					if val > maxVal {
//						maxVal = val
//					}
//				}
//				for i, val := range macAverages[mac] {
//					//todo: why not using the actual values of macAverages instead of the normalized values?
//					macAverages[mac][i] = maxVal / val // normalization(because val is < 0, we use maxVal/val instead of val /maxVal)
//				}
//				// MacVariability shows the standard deviation of a specific AP in all locations
//				md.MacVariability[mac] = glb.StandardDeviation(macAverages[mac]) //refer to line 300 todo
//			}
//		}
//	}
//
//	// Determine mac frequencies and normalize
//	for n := range BayesPriors {
//		for loc := range md.NetworkLocs[n] {
//			maxCount := 0
//			for mac := range md.MacCountByLoc[loc] {
//				if md.MacCountByLoc[loc][mac] > maxCount {
//					maxCount = md.MacCountByLoc[loc][mac] //maxCount:repeat number of the most seen mac in a location
//
//				}
//			}
//			//fmt.Println("MAX COUNT:", maxCount)
//			for mac := range md.MacCountByLoc[loc] {
//				//if a mac is not seen in a location, the macFreq of that mac equals to 0 (ps.MacCountByLoc[loc][mac]).
//				//todo: Does the above mentioned 0 value make some error in the bayesian function?
//				BayesPriors[n].MacFreq[loc][mac] = float32(md.MacCountByLoc[loc][mac]) / float32(maxCount)
//				//fmt.Println("mac freq:", ps.BayesPriors[n].MacFreq[loc][mac])
//				if float64(BayesPriors[n].MacFreq[loc][mac]) < BayesPriors[n].Special["MacFreqMin"] {
//					BayesPriors[n].Special["MacFreqMin"] = float64(BayesPriors[n].MacFreq[loc][mac])
//				}
//			}
//		}
//	}
//
//	// Determine negative mac frequencies and normalize
//	for n := range BayesPriors {
//		for loc1 := range BayesPriors[n].MacFreq {
//			sum := float32(0)
//			for loc2 := range BayesPriors[n].MacFreq {
//				if loc2 != loc1 {
//					for mac := range BayesPriors[n].MacFreq[loc2] {
//						BayesPriors[n].NMacFreq[loc1][mac] += BayesPriors[n].MacFreq[loc2][mac]
//					}
//					sum++
//				}
//			}
//			// sum = i(i-1); i = ps.NetworkLocs[n]
//			// Normalize
//			//Done: it seems that sum is not calculated correctly. It should be equals to "number of locations-1"
//			if sum > 0 {
//				for mac := range BayesPriors[n].MacFreq[loc1] {
//					BayesPriors[n].NMacFreq[loc1][mac] = BayesPriors[n].NMacFreq[loc1][mac] / sum
//					if float64(BayesPriors[n].NMacFreq[loc1][mac]) < BayesPriors[n].Special["NMacFreqMin"] {
//						BayesPriors[n].Special["NMacFreqMin"] = float64(BayesPriors[n].NMacFreq[loc1][mac])
//					}
//				}
//			}
//		}
//	}
//	//todo: the default values for MixIn and Cutoff should be set as initial values not hardcoded values
//	for n := range BayesPriors {
//		BayesPriors[n].Special["MixIn"] = 0.5
//		//todo: spell check for Varability!
//		BayesPriors[n].Special["VarabilityCutoff"] = 0
//	}
//	*bayesPriors = BayesPriors
//}
//
