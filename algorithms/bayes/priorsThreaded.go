// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// priorsThreaded.go contains the main Prior-calculation function which is multi-threaded

package bayes

import (
	"math"
	"runtime"
	"ParsinServer/glb"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/dbm"
)

//following this:https://play.golang.org/p/hK2h-irKyz
// Result of worker()
type resultA struct {
	mixin         float64
	locationGuess string
	locationTrue  string
	n             string
}

// Job that is got to worker()
type jobA struct {
	mixin        float64
	locs         []string
	bayes1       []float64
	bayes2       []float64
	n            string
	locationTrue string
}



// Calculate the value of mixed pbayes1 and pbayes2 and find the location with the maximum probability
func worker(id int, jobs <-chan jobA, results chan<- resultA) {
	for j := range jobs {
		maxVal := float64(-1)
		locationGuess := ""
		for i, loc := range j.locs {
			PBayesMix := j.mixin*j.bayes1[i] + (1-j.mixin)*j.bayes2[i]
			if PBayesMix > maxVal {
				maxVal = PBayesMix
				locationGuess = loc
			}
		}
		results <- resultA{locationGuess: locationGuess,
			locationTrue: j.locationTrue,
			mixin: j.mixin,
			n: j.n}
	}
}

// optimizePriorsThreaded generates the optimized prior data for Naive-Bayes classification.
func OptimizePriorsThreaded(group string) error {
	// generate the fingerprintsInMemory

	//var psMain = *parameters.NewFullParameters()
	var gpMain = dbm.NewGroup(group)
	defer dbm.GM.GetGroup(group).Set(gpMain)

	// Get real PS from raw fingerprint data
	fingerprintsInMemoryMain := make(map[string]parameters.Fingerprint)
	var fingerprintsOrderingMain []string
	var err error
	//opening the db
	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//defer db.Close()
	// if err != nil {
	//	log.Fatal(err)
	//	return err
	//}

	fingerprintsOrderingMain,fingerprintsInMemoryMain,err = dbm.GetLearnFingerPrints(group,true)
	if err != nil {
		return err
	}

	GetParameters(group, gpMain, fingerprintsInMemoryMain, fingerprintsOrderingMain)

	//glb.Debug.Println(fingerprintsInMemory)
	//Info.Println("Running calculatePriors")
	if glb.RuntimeArgs.GaussianDist {
		calculateGaussianPriors(group, gpMain, fingerprintsInMemoryMain, fingerprintsOrderingMain)
	} else {
		calculatePriors(group, gpMain, fingerprintsInMemoryMain, fingerprintsOrderingMain)
	}

	//fmt.Println(ps1)

	fpInMemoryCrossTest := make(map[string]parameters.Fingerprint)
	fpInMemoryCrossTrain := make(map[string]parameters.Fingerprint)

	var fpOrderingCrossTest []string
	var fpOrderingCrossTrain []string

	//opening the db

	it := float64(-1)
	for fpTime,fp := range fingerprintsInMemoryMain{
		it++
		if math.Mod(it, FoldCrossValidation) == 0 {
			fpInMemoryCrossTest[fpTime] = fp
			//fingerprintsOrdering is an array of fingerprintsInMemory keys
			fpOrderingCrossTest = append(fpOrderingCrossTest, fpTime)
		} else {
			//if fpCounter*((FoldCrossValidation-float64(1))/FoldCrossValidation) >= it {
			fpInMemoryCrossTrain[fpTime] = fp
			//fingerprintsOrdering is an array of fingerprintsInMemory keys
			fpOrderingCrossTrain = append(fpOrderingCrossTrain, fpTime)

		}
	}


	//var ps = *parameters.NewFullParameters()
	var gpCross = dbm.NewGroup(group)

	//GetParameters(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	//#cache
	GetParameters(group, gpCross, fpInMemoryCrossTrain, fpOrderingCrossTrain)

	//Info.Println("Running calculatePriors")
	if glb.RuntimeArgs.GaussianDist {
		calculateGaussianPriors(group, gpCross, fpInMemoryCrossTrain, fpOrderingCrossTrain)
	} else {
		calculatePriors(group, gpCross, fpInMemoryCrossTrain, fpOrderingCrossTrain)
	}

	var results = *parameters.NewResultsParameters()
	for n := range gpCross.Priors {
		gpCross.Results[n] = results //Results is shared between all networks
	}

	//glb.Debug.Println("########################################")
	//glb.Debug.Println(group)
	//glb.Debug.Println(len(fingerprintsInMemoryMain))
	//glb.Debug.Println(len(fingerprintsInMemoryCross))
	//glb.Debug.Println(len(fingerprintsInMemory))
	//glb.Debug.Println("########################################")


	// Mixin is ration of pbayes1(proximity to AP) to pbayes2(normal bayesian classifier result)
	mixins := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}
	//mixinOverride, _ := dbm.GetMixinOverride(group)
	mixinOverride := dbm.GetSharedPrf(group).Mixin
	if mixinOverride >= 0 && mixinOverride <= 1 {
		mixins = []float64{mixinOverride}
	}

	// cutoffs is a number which is compared with the standard deviation of a specific AP in all locations(MacVariability)
	// if macVariability is lower than cutoff it is ignored in PBayes2 calculation.
	cutoffs := []float64{0.005, 0.05, 0.1}
	//cutoffOverride, _ := dbm.GetCutoffOverride(group)
	cutoffOverride := dbm.GetSharedPrf(group).Cutoff
	if cutoffOverride >= 0 && cutoffOverride <= 1 {
		cutoffs = []float64{cutoffOverride}
	}

	bestMixin := make(map[string]float64)
	bestResult := make(map[string]float64)
	bestCutoff := make(map[string]float64)
	for n := range gpCross.Priors {
		bestResult[n] = 0
		bestMixin[n] = 0
		bestCutoff[n] = 0
	}

	for _, cutoff := range cutoffs {

		//                  network    id         loc    value
		PBayes1 := make(map[string]map[string]map[string]float64)
		PBayes2 := make(map[string]map[string]map[string]float64)
		totalJobs := 0
		for n := range gpCross.Priors {
			it := float64(-1)
			PBayes1[n] = make(map[string]map[string]float64)
			PBayes2[n] = make(map[string]map[string]float64)

			for _, v1 := range fpOrderingCrossTest {
				it++
				// call calculatePosteriorThreadSafe function every 3 times in 4 times
				//if math.Mod(it, FoldCrossValidation) != 0 {
				//_, ok := ps.NetworkLocs[n][fingerprintsInMemoryCross[v1].Location]
				_, ok := gpCross.NetworkLocs[n][fpInMemoryCrossTest[v1].Location]
					// Check if the fingerprint's location exists in ps
					// todo: ps is made from the fingerprintsInMemory; so there is no need for bellow if!
				if len(fpInMemoryCrossTest[v1].WifiFingerprint) == 0 || !ok {
						continue
					}
					totalJobs++
					//todo: in the following line, fingerprintsInMemory[v1] is included in ps; so the test set(fingerprintsInMemory[v1]) is included in the training set(ps)
					//
				PBayes1[n][v1], PBayes2[n][v1] = CalculatePosteriorThreadSafe(group,gpCross,fpInMemoryCrossTest[v1], cutoff)
				//}
			}
		}

		numJobs := len(mixins) * totalJobs
		runtime.GOMAXPROCS(glb.MaxParallelism())
		chanJobs := make(chan jobA, 1+numJobs)
		chanResults := make(chan resultA, 1+numJobs)
		for w := 1; w <= glb.MaxParallelism(); w++ {
			go worker(w, chanJobs, chanResults)
		}

		finalResults := make(map[string]map[float64]parameters.ResultsParameters)

		// Fill chanJobs with Jobs
		for n := range gpCross.Get_Priors() {
			finalResults[n] = make(map[float64]parameters.ResultsParameters)
			for _, mixin := range mixins {

				finalResults[n][mixin] = *parameters.NewResultsParameters()
				for loc := range gpCross.NetworkLocs[n] {
					finalResults[n][mixin].TotalLocations[loc] = 0
					finalResults[n][mixin].CorrectLocations[loc] = 0
					finalResults[n][mixin].Accuracy[loc] = 0
					finalResults[n][mixin].Guess[loc] = make(map[string]int)
				}
				// Loop through fingerprints which their posterior is included in PBayes1

				for id := range PBayes1[n] { //id = FG timestamps = fingerprint ordering members
					locs := []string{}
					bayes1 := []float64{}
					bayes2 := []float64{}
					for key := range PBayes1[n][id] { //key = locations
						locs = append(locs, key)
						bayes1 = append(bayes1, PBayes1[n][id][key])
						bayes2 = append(bayes2, PBayes2[n][id][key]) //length of PBayes1 array equals to length of PBayes2
					}
					trueLoc := fpInMemoryCrossTest[id].Location
					chanJobs <- jobA{n: n,
						mixin: mixin,
						locs: locs,
						locationTrue: trueLoc,
						bayes1: bayes1,
						bayes2: bayes2}
				}
			}
		}
		close(chanJobs)

		for a := 1; a <= numJobs; a++ {
			t := <-chanResults
			// ps.Results isn't set here(finalResults is a temporary struct), it is set in crossValidation() function
			finalResults[t.n][t.mixin].TotalLocations[t.locationTrue]++ //num of location estimations
			if t.locationGuess == t.locationTrue {
				finalResults[t.n][t.mixin].CorrectLocations[t.locationTrue]++ // num of correct estimations
			}
			//init
			if _, ok := finalResults[t.n][t.mixin].Guess[t.locationTrue]; !ok {
				finalResults[t.n][t.mixin].Guess[t.locationTrue] = make(map[string]int)
			}
			//init
			if _, ok := finalResults[t.n][t.mixin].Guess[t.locationTrue][t.locationGuess]; !ok {
				finalResults[t.n][t.mixin].Guess[t.locationTrue][t.locationGuess] = 0
			}
			finalResults[t.n][t.mixin].Guess[t.locationTrue][t.locationGuess]++
		}

		for n := range gpCross.Priors {
			for mixin := range finalResults[n] {
				average := float64(0)
				it := 0
				for loc := range finalResults[n][mixin].TotalLocations {
					if finalResults[n][mixin].TotalLocations[loc] > 0 { //todo: finalResults[n][mixin].TotalLocations[loc] is always greater than 0
						finalResults[n][mixin].Accuracy[loc] = int(100.0 * finalResults[n][mixin].CorrectLocations[loc] / finalResults[n][mixin].TotalLocations[loc])
						// Debug.Println(n, mixin, cutoff, loc, finalResults[n][mixin].Accuracy[loc])
						average += float64(finalResults[n][mixin].Accuracy[loc])
						it++
					}
				}
				average = average / float64(it)
				// fmt.Println(mixin, average)

				// todo: choose a better algorithm to select the best MixIn rather than average
				if average > bestResult[n] {
					bestResult[n] = average
					bestMixin[n] = mixin
					bestCutoff[n] = cutoff
				}
			}
		}

	}

	// Load new priors and calculate new cross Validation
	for n := range gpCross.Priors {
		gpCross.Priors[n].Special["MixIn"] = bestMixin[n]
		gpCross.Priors[n].Special["VarabilityCutoff"] = bestCutoff[n]
		// (1-1/FoldCrossValidation) of the learned fingerprints are used to set the best mixin and cutoff,
		// 	then (1/FoldCrossValidation) of remained fingerprints are used to set ps.Results(Accuracy,TotalLocations,CorrectLocations,Guess)
		crossValidation(group, n, gpCross, fpInMemoryCrossTest, fpOrderingCrossTest)
	}

	//fmt.Println(ps)

	for n := range gpMain.Priors {
		gpMain.Priors[n].Special["MixIn"] = bestMixin[n]
		gpMain.Priors[n].Special["VarabilityCutoff"] = bestCutoff[n]
	}

	gpMain.Results = gpCross.Results

	// Debug.Println(getUsers(group))
	go dbm.ResetCache("usersCache")

	//go dbm.SaveParameters(group, psMain)
	//go dbm.SetPsCache(group, psMain)

	return nil
}

// Not threaded version of optimizePriorsThreaded() function
func OptimizePriorsThreadedNot(group string) {
	// generate the fingerprintsInMemory
	// Debug.Println("Optimizing priors for " + group)

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

	var results = *parameters.NewResultsParameters()
	for n := range gp.Priors {
		gp.Results[n] = results
	}

	// loop through these parameters
	mixins := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}
	// cutoff := 0.1
	cutoffs := []float64{0.005, 0.05, 0.1}
	bestMixin := make(map[string]float64)
	bestResult := make(map[string]float64)
	bestCutoff := make(map[string]float64)
	for n := range gp.Priors {
		bestResult[n] = 0
		bestMixin[n] = 0
		bestCutoff[n] = 0
	}

	for _, cutoff := range cutoffs {

		//                 network      id      loc    value
		PBayes1 := make(map[string]map[string]map[string]float64)
		PBayes2 := make(map[string]map[string]map[string]float64)
		totalJobs := 0
		for n := range gp.Priors {
			it := float64(-1)
			PBayes1[n] = make(map[string]map[string]float64)
			PBayes2[n] = make(map[string]map[string]float64)

			for _, v1 := range fingerprintsOrdering {
				it++
				if math.Mod(it, FoldCrossValidation) == 0 {
					_, ok := gp.NetworkLocs[n][fingerprintsInMemory[v1].Location]
					if len(fingerprintsInMemory[v1].WifiFingerprint) == 0 || !ok {
						continue
					}
					totalJobs++
					PBayes1[n][v1], PBayes2[n][v1] = CalculatePosteriorThreadSafe(group,gp,fingerprintsInMemory[v1], cutoff)
				}
			}
		}

		finalResults := make(map[string]map[float64]parameters.ResultsParameters)
		bestMixin := make(map[string]float64)
		bestResult := make(map[string]float64)
		for n := range gp.Priors {
			bestResult[n] = 0
			bestMixin[n] = 0
			finalResults[n] = make(map[float64]parameters.ResultsParameters)
			for _, mixin := range mixins {

				finalResults[n][mixin] = *parameters.NewResultsParameters()
				for loc := range gp.NetworkLocs[n] {
					finalResults[n][mixin].TotalLocations[loc] = 0
					finalResults[n][mixin].CorrectLocations[loc] = 0
					finalResults[n][mixin].Accuracy[loc] = 0
					finalResults[n][mixin].Guess[loc] = make(map[string]int)
				}
				// Loop through each fingerprint
				for id := range PBayes1[n] {
					maxVal := float64(-1)
					locationGuess := ""
					for key := range PBayes1[n][id] {
						PBayesMix := mixin*PBayes1[n][id][key] + (1-mixin)*PBayes2[n][id][key]
						if PBayesMix > maxVal {
							maxVal = PBayesMix
							locationGuess = key
						}
						locationTrue := fingerprintsInMemory[id].Location
						finalResults[n][mixin].TotalLocations[locationTrue]++
						if locationGuess == locationTrue {
							finalResults[n][mixin].CorrectLocations[locationTrue]++
						}
						if _, ok := finalResults[n][mixin].Guess[locationTrue]; !ok {
							finalResults[n][mixin].Guess[locationTrue] = make(map[string]int)
						}
						if _, ok := finalResults[n][mixin].Guess[locationTrue][locationGuess]; !ok {
							finalResults[n][mixin].Guess[locationTrue][locationGuess] = 0
						}
						finalResults[n][mixin].Guess[locationTrue][locationGuess]++
					}
				}
				average := float64(0)
				it := 0
				for loc := range finalResults[n][mixin].TotalLocations {
					if finalResults[n][mixin].TotalLocations[loc] > 0 {
						finalResults[n][mixin].Accuracy[loc] = int(100.0 * finalResults[n][mixin].CorrectLocations[loc] / finalResults[n][mixin].TotalLocations[loc])
						average += float64(finalResults[n][mixin].Accuracy[loc])
						it++
					}
				}
				average = average / float64(it)
				// fmt.Println(mixin, average, a)
				if average > bestResult[n] {
					bestResult[n] = average
					bestMixin[n] = mixin
					bestCutoff[n] = cutoff
				}
			}
		}
	}

	// Load new priors and calculate new cross Validation
	for n := range gp.Priors {
		gp.Priors[n].Special["MixIn"] = bestMixin[n]
		gp.Priors[n].Special["VarabilityCutoff"] = bestCutoff[n]
		crossValidation(group, n, gp, fingerprintsInMemory, fingerprintsOrdering)
	}
	//go dbm.SaveParameters(group, ps)
	//go dbm.SetPsCache(group, ps)
	// Debug.Println("Analyzed ", totalJobs, " fingerprints")
}
