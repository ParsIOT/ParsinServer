// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// posteriors.go contains variables for calculating Naive-Bayes posteriors.

package bayes

import (
	"math"
	"ParsinServer/glb"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/algorithms/clustering"
	"ParsinServer/dbm"
)

// calculatePosterior takes a parameters.Fingerprint and a Parameter set and returns the noramlized Bayes probabilities of possible locations
func CalculatePosterior(res parameters.Fingerprint, gp *dbm.Group) (string, map[string]float64) {
	//if !ps.Loaded {
	//	ps, _ = dbm.OpenParameters(res.Group)
	//}
	if gp == nil{
		gp = dbm.GM.GetGroup(res.Group).Get()
		//glb.Debug.Println(gp)
		//defer dbm.GM.GetGroup(res.Group).Set(gp)
		// No need to write back
	}

	macs := []string{}
	W := make(map[string]int)
	for v2 := range res.WifiFingerprint {
		macs = append(macs, res.WifiFingerprint[v2].Mac)
		W[res.WifiFingerprint[v2].Mac] = res.WifiFingerprint[v2].Rssi
	}
	n, inNetworkAlready := clustring.HasNetwork(gp.NetworkMacs, macs)
	//Debug.Println(ps.NetworkMacs)
	//Debug.Println(macs)
	//Debug.Println(n, inNetworkAlready, ps.NetworkLocs[n])
	if !inNetworkAlready {
		glb.Warning.Println("Not in network")
		glb.Debug.Println(n, inNetworkAlready, gp.NetworkLocs[n], res)
	}

	if len(gp.NetworkLocs[n]) == 1 {
		for key := range gp.NetworkLocs[n] {
			PBayesMix := make(map[string]float64)
			PBayesMix[key] = 1
			return key, PBayesMix
		}
	}

	PBayes1 := make(map[string]float64)
	PBayes2 := make(map[string]float64)
	PA := 1.0 / float64(len(gp.NetworkLocs[n]))
	PnA := (float64(len(gp.NetworkLocs[n])) - 1.0) / float64(len(gp.NetworkLocs[n]))
	for loc := range gp.NetworkLocs[n] {
		PBayes1[loc] = float64(0)
		PBayes2[loc] = float64(0)
		for mac := range W {
			weight := float64(0)
			nweight := float64(0)
			if _, ok := gp.Priors[n].MacFreq[loc][mac]; ok {
				weight = float64(gp.Priors[n].MacFreq[loc][mac])
			} else {
				weight = float64(gp.Priors[n].Special["MacFreqMin"])
			}
			if _, ok := gp.Priors[n].NMacFreq[loc][mac]; ok {
				nweight = float64(gp.Priors[n].NMacFreq[loc][mac])
			} else {
				nweight = float64(gp.Priors[n].Special["NMacFreqMin"])
			}
			PBayes1[loc] += math.Log(weight*PA) - math.Log(weight*PA+PnA*nweight)

			if float64(gp.MacVariability[mac]) >= gp.Priors[n].Special["VarabilityCutoff"] && W[mac] > glb.MinRssiOpt {
				ind := int(W[mac] - glb.MinRssi)
				if len(gp.Priors[n].P[loc][mac]) > 0 {
					PBA := float64(gp.Priors[n].P[loc][mac][ind])
					//PBnA := float64(ps.Priors[n].NP[loc][mac][ind])
					if PBA > 0 {
						//PBayes2[loc] += (math.Log(PBA*PA) - math.Log(PBA*PA+PBnA*PnA))
						PBayes2[loc] += math.Log(PBA)
					} else {
						PBayes2[loc] += -1
					}
				}
			}
		}
	}
	PBayes1 = NormalizeBayes(PBayes1)
	PBayes2 = NormalizeBayes(PBayes2)
	PBayesMix := make(map[string]float64)
	bestLocation := ""
	maxVal := float64(-100)
	for key := range PBayes1 { //key = loc
		PBayesMix[key] = gp.Priors[n].Special["MixIn"]*PBayes1[key] + (1-gp.Priors[n].Special["MixIn"])*PBayes2[key]
		if PBayesMix[key] > maxVal {
			maxVal = PBayesMix[key]
			bestLocation = key
		}
	}
	return bestLocation, PBayesMix
}

// calculatePosteriorThreadSafe is exactly the same as calculatePosterior except it does not do the mixin calculation
// as it is used for optimizing priors.
func CalculatePosteriorThreadSafe(groupName string,gp *dbm.Group, res parameters.Fingerprint, cutoff float64) (map[string]float64, map[string]float64) {
	//#cache
	//gp := dbm.GM.GetGroup(groupName).Get()

	//if !ps.Loaded {
	//	ps, _ = dbm.OpenParameters(res.Group)
	//}
	macs := []string{}
	//Done: rename W
	resRoutes := make(map[string]int) //a map from mac to rssi
	for v2 := range res.WifiFingerprint {
		macs = append(macs, res.WifiFingerprint[v2].Mac)
		resRoutes[res.WifiFingerprint[v2].Mac] = res.WifiFingerprint[v2].Rssi
	}
	//n, inNetworkAlready := clustring.HasNetwork(ps.NetworkMacs, macs)
	//#cache
	n, inNetworkAlready := clustring.HasNetwork(gp.NetworkMacs, macs)

	// Debug.Println(n, inNetworkAlready, ps.NetworkLocs[n])
	if !inNetworkAlready {
		glb.Warning.Println("Not in network")
		//glb.Debug.Println(n, inNetworkAlready, ps.NetworkLocs[n], res)
		//#cache
		glb.Debug.Println(n, inNetworkAlready, gp.NetworkLocs[n], res)

	}

	PBayes1 := make(map[string]float64)
	PBayes2 := make(map[string]float64)
	//PA := 1.0 / float64(len(ps.NetworkLocs[n]))//the real prior !
	//PnA := (float64(len(ps.NetworkLocs[n])) - 1.0) / float64(len(ps.NetworkLocs[n])) // 1.0 - PA
	//#cache
	PA := 1.0 / float64(len(gp.NetworkLocs[n]))//the real prior !
	PnA := (float64(len(gp.NetworkLocs[n])) - 1.0) / float64(len(gp.NetworkLocs[n])) // 1.0 - PA

	//for loc := range ps.NetworkLocs[n] {
	//#cache
	for loc := range gp.NetworkLocs[n] {
		PBayes1[loc] = float64(0)
		PBayes2[loc] = float64(0)
		for mac := range resRoutes {
			weight := float64(0)
			nweight := float64(0)

			// todo: The condition should be like this: ok && ps.Priors[n].MacFreq[loc][mac]!=0
			//if _, ok := ps.Priors[n].MacFreq[loc][mac]; ok {
			//#cache
			if _, ok := gp.Priors[n].MacFreq[loc][mac]; ok {
				//if the MacFreq (or weight) of a Mac in a location is high, it means that the location is near the AP(Mac).
				weight = float64(gp.Priors[n].MacFreq[loc][mac])
			} else {
				//todo:why not using 0 instead of ps.Priors[n].Special["MacFreqMin"]?
				weight = float64(gp.Priors[n].Special["MacFreqMin"])
			}
			if _, ok := gp.Priors[n].NMacFreq[loc][mac]; ok {
				// nweight determines presence of the AP signals in other locations
				nweight = float64(gp.Priors[n].NMacFreq[loc][mac])
			} else {
				nweight = float64(gp.Priors[n].Special["NMacFreqMin"])
			}
			//fmt.Println("bayes", (weight*PA)/(weight*PA+PnA*nweight))
			//PBayes1 is used for proximity purposes.
			PBayes1[loc] += math.Log(weight*PA) - math.Log(weight*PA+PnA*nweight)

			// todo: why not verifying the (resRoutes[mac] > MinRssi) & (ps.MacVariability[mac]) >= cutoff) conditions while calculation PBayes1?
			// cutoffs is a number which is compared with the standard deviation of a specific AP in all locations(MacVariability)
			// if macVariability is lower than cutoff it is ignored in PBayes2 calculation.
			if float64(gp.MacVariability[mac]) >= cutoff && resRoutes[mac] > glb.MinRssiOpt { //TODO: why calculating the mac variability of a mac not a location?
				ind := int(resRoutes[mac] - glb.MinRssi) //same as what is done in P calculation
				if len(gp.Priors[n].P[loc][mac]) > 0 {
					PBA := float64(gp.Priors[n].P[loc][mac][ind])
					PBnA := float64(gp.Priors[n].NP[loc][mac][ind])
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
	PBayes1 = NormalizeBayes(PBayes1) //todo : what is the phase of him?!!!
	PBayes2 = NormalizeBayes(PBayes2)
	return PBayes1, PBayes2
}

// normalizeBayes takes the bayes map and normalizes to standard normal.
func NormalizeBayes(bayes map[string]float64) map[string]float64 {
	vals := make([]float64, len(bayes))
	i := 0
	for _, val := range bayes {
		vals[i] = val
		i++
	}
	mean := glb.Average64(vals)
	sd := glb.StandardDeviation64(vals)
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
