// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// posteriors.go contains variables for calculating Naive-Bayes posteriors.

package main

import "math"

// calculatePosterior takes a Fingerprint and a Parameter set and returns the noramlized Bayes probabilities of possible locations
func calculatePosterior(res Fingerprint, ps FullParameters) (string, map[string]float64) {
	if !ps.Loaded {
		ps, _ = openParameters(res.Group)
	}
	macs := []string{}
	W := make(map[string]int)
	for v2 := range res.WifiFingerprint {
		macs = append(macs, res.WifiFingerprint[v2].Mac)
		W[res.WifiFingerprint[v2].Mac] = res.WifiFingerprint[v2].Rssi
	}
	n, inNetworkAlready := hasNetwork(ps.NetworkMacs, macs)
	// Debug.Println(n, inNetworkAlready, ps.NetworkLocs[n])
	if !inNetworkAlready {
		Warning.Println("Not in network")
		Debug.Println(n, inNetworkAlready, ps.NetworkLocs[n], res)
	}

	if len(ps.NetworkLocs[n]) == 1 {
		for key := range ps.NetworkLocs[n] {
			PBayesMix := make(map[string]float64)
			PBayesMix[key] = 1
			return key, PBayesMix
		}
	}

	PBayes1 := make(map[string]float64)
	PBayes2 := make(map[string]float64)
	PA := 1.0 / float64(len(ps.NetworkLocs[n]))
	PnA := (float64(len(ps.NetworkLocs[n])) - 1.0) / float64(len(ps.NetworkLocs[n]))
	for loc := range ps.NetworkLocs[n] {
		PBayes1[loc] = float64(0)
		PBayes2[loc] = float64(0)
		for mac := range W {
			weight := float64(0)
			nweight := float64(0)
			if _, ok := ps.Priors[n].MacFreq[loc][mac]; ok {
				weight = float64(ps.Priors[n].MacFreq[loc][mac])
			} else {
				weight = float64(ps.Priors[n].Special["MacFreqMin"])
			}
			if _, ok := ps.Priors[n].NMacFreq[loc][mac]; ok {
				nweight = float64(ps.Priors[n].NMacFreq[loc][mac])
			} else {
				nweight = float64(ps.Priors[n].Special["NMacFreqMin"])
			}
			PBayes1[loc] += math.Log(weight*PA) - math.Log(weight*PA+PnA*nweight)

			if float64(ps.MacVariability[mac]) >= ps.Priors[n].Special["VarabilityCutoff"] && W[mac] > MinRssiOpt {
				ind := int(W[mac] - MinRssi)
				if len(ps.Priors[n].P[loc][mac]) > 0 {
					PBA := float64(ps.Priors[n].P[loc][mac][ind])
					PBnA := float64(ps.Priors[n].NP[loc][mac][ind])
					if PBA > 0 {
						PBayes2[loc] += (math.Log(PBA*PA) - math.Log(PBA*PA+PBnA*PnA))
					} else {
						PBayes2[loc] += -1
					}
				}
			}
		}
	}
	PBayes1 = normalizeBayes(PBayes1)
	PBayes2 = normalizeBayes(PBayes2)
	PBayesMix := make(map[string]float64)
	bestLocation := ""
	maxVal := float64(-100)
	for key := range PBayes1 { //key = loc
		PBayesMix[key] = ps.Priors[n].Special["MixIn"]*PBayes1[key] + (1-ps.Priors[n].Special["MixIn"])*PBayes2[key]
		if PBayesMix[key] > maxVal {
			maxVal = PBayesMix[key]
			bestLocation = key
		}
	}
	return bestLocation, PBayesMix
}

// calculatePosteriorThreadSafe is exactly the same as calculatePosterior except it does not do the mixin calculation
// as it is used for optimizing priors.
func calculatePosteriorThreadSafe(res Fingerprint, ps FullParameters, cutoff float64) (map[string]float64, map[string]float64) {
	if !ps.Loaded {
		ps, _ = openParameters(res.Group)
	}
	macs := []string{}
	//Done: rename W
	resRoutes := make(map[string]int) //a map from mac to rssi
	for v2 := range res.WifiFingerprint {
		macs = append(macs, res.WifiFingerprint[v2].Mac)
		resRoutes[res.WifiFingerprint[v2].Mac] = res.WifiFingerprint[v2].Rssi
	}
	n, inNetworkAlready := hasNetwork(ps.NetworkMacs, macs)
	// Debug.Println(n, inNetworkAlready, ps.NetworkLocs[n])
	if !inNetworkAlready {
		Warning.Println("Not in network")
		Debug.Println(n, inNetworkAlready, ps.NetworkLocs[n], res)
	}

	PBayes1 := make(map[string]float64)
	PBayes2 := make(map[string]float64)
	PA := 1.0 / float64(len(ps.NetworkLocs[n]))//the real prior !
	PnA := (float64(len(ps.NetworkLocs[n])) - 1.0) / float64(len(ps.NetworkLocs[n])) // 1.0 - PA
	for loc := range ps.NetworkLocs[n] {
		PBayes1[loc] = float64(0)
		PBayes2[loc] = float64(0)
		for mac := range resRoutes {
			weight := float64(0)
			nweight := float64(0)

			// todo: The condition should be like this: ok && ps.Priors[n].MacFreq[loc][mac]!=0
			if _, ok := ps.Priors[n].MacFreq[loc][mac]; ok {
				//if the MacFreq (or weight) of a Mac in a location is high, it means that the location is near the AP(Mac).
				weight = float64(ps.Priors[n].MacFreq[loc][mac])
			} else {
				//todo:why not using 0 instead of ps.Priors[n].Special["MacFreqMin"]?
				weight = float64(ps.Priors[n].Special["MacFreqMin"])
			}
			if _, ok := ps.Priors[n].NMacFreq[loc][mac]; ok {
				// nweight determines presence of the AP signals in other locations
				nweight = float64(ps.Priors[n].NMacFreq[loc][mac])
			} else {
				nweight = float64(ps.Priors[n].Special["NMacFreqMin"])
			}
			//fmt.Println("bayes", (weight*PA)/(weight*PA+PnA*nweight))
			//PBayes1 is used for proximity purposes.
			PBayes1[loc] += math.Log(weight*PA) - math.Log(weight*PA+PnA*nweight)

			// todo: why not verifying the (W[mac] > MinRssi) & (ps.MacVariability[mac]) >= cutoff) conditions while calculation PBayes1
			// cutoffs is a number which is compared with the standard deviation of a specific AP in all locations(MacVariability)
			// if macVariability is lower than cutoff it is ignored in PBayes2 calculation.
			if float64(ps.MacVariability[mac]) >= cutoff && resRoutes[mac] > MinRssiOpt { //TODO: why calculating the mac variability of a mac not a location?
				ind := int(resRoutes[mac] - MinRssi) //same as what is done in P calculation
				if len(ps.Priors[n].P[loc][mac]) > 0 {
					PBA := float64(ps.Priors[n].P[loc][mac][ind])
					PBnA := float64(ps.Priors[n].NP[loc][mac][ind])
					if PBA > 0 {
						//todo: replace the PBayes2 calculation with the standard bayesian calculation
						PBayes2[loc] += (math.Log(PBA*PA) - math.Log(PBA*PA+PBnA*PnA)) //todo: what is this?
					} else {
						PBayes2[loc] += -1
					}

				}
			}
		}
	}
	PBayes1 = normalizeBayes(PBayes1) //todo : what is the phase of him?!!!
	PBayes2 = normalizeBayes(PBayes2)
	return PBayes1, PBayes2
}

// normalizeBayes takes the bayes map and normalizes to standard normal.
func normalizeBayes(bayes map[string]float64) map[string]float64 {
	vals := make([]float64, len(bayes))
	i := 0
	for _, val := range bayes {
		vals[i] = val
		i++
	}
	mean := average64(vals)
	sd := standardDeviation64(vals)
	for key := range bayes {
		// todo: why 1e-5?
		if sd < 1e-5 {
			bayes[key] = 0
		} else {
			bayes[key] = (bayes[key] - mean) / sd
		}
		if math.IsNaN(bayes[key]) {
			bayes[key] = 0
		}
	}
	return bayes
}