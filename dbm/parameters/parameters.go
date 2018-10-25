// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// parameters.go contains structures and functions for setting and getting Naive-Bayes parameters.

package parameters

import "ParsinServer/glb"

// PersistentParameters are not reloaded each time
//type PersistentParameters struct {
//	NetworkRenamed map[string][]string // key:networkName, value:mac list; e.g.: {"1":["mac1","mac2"]}
//}

// Constant parameters that set manually are in KnnParameters
type KnnParameters struct {
	MaxEuclideanRssDist int
}

func NewKnnParameters() KnnParameters {
	return KnnParameters{
		MaxEuclideanRssDist: glb.DefaultMaxEuclideanRssDist,
	}
}

// Constant parameters that set by cross-validation are in KnnHyperParameters
type KnnHyperParameters struct {
	K             int       `json:"K"`
	MinClusterRss int       `json:"MinClusterRss"`
	GraphFactors  []float64 `json:"GraphFactors"`
}

func NewKnnHyperParameters() KnnHyperParameters {
	return KnnHyperParameters{
		K:             glb.DefaultKnnKRange[0],
		MinClusterRss: glb.DefaultKnnMinCRssRange[0],
		GraphFactors:  glb.DefaultGraphFactorsRange[0],
	}
}

type KnnFingerprints struct {
	FingerprintsInMemory map[string]Fingerprint `json:"FingerprintsInMemory"`
	FingerprintsOrdering []string               `json:"FingerprintsOrdering"`
	Clusters             map[string][]string    `json:"Clusters"`
	HyperParameters      KnnHyperParameters     `json:"HyperParameters"`
	Node2FPs             map[string][]string    `json:"Node2FPs"`
}

func NewKnnFingerprints() KnnFingerprints {
	return KnnFingerprints{
		FingerprintsInMemory: make(map[string]Fingerprint),
		FingerprintsOrdering: []string{},
		Clusters:             make(map[string][]string),
		HyperParameters:      NewKnnHyperParameters(),
		Node2FPs:             make(map[string][]string),
	}
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
//
//// FullParameters is the full parameter set for a given group
//type FullParameters struct {
//	NetworkMacs    map[string]map[string]bool   // map of networks and then the associated macs in each
//	NetworkLocs    map[string]map[string]bool   // map of the networks, and then the associated locations in each
//	MacVariability map[string]float32           // variability of macs
//	MacCount       map[string]int               // number of fingerprints of a AP in all data, regardless of the location; e.g. 10 of AP1, 12 of AP2, ...
//	MacCountByLoc  map[string]map[string]int    // number of fingerprints of a AP in a location; e.g. in location A, 10 of AP1, 12 of AP2, ...
//	UniqueLocs     []string                     // a list of all unique locations e.g. {P1,P2,P3}
//	UniqueMacs     []string                     // a list of all unique APs
//	BayesPriors         map[string]PriorParameters   // generate priors for each network
//	BayesResults        map[string]ResultsParameters // generate results for each network
//	Loaded         bool                         // flag to determine if parameters have been loaded
//}
//
//// NewFullParameters generates a blank FullParameters
//func NewFullParameters() *FullParameters {
//	return &FullParameters{
//		//todo: networkMacs difference with UniqueMacs
//		//todo: NetworkLocs difference with UniqueLocs
//		//todo: in networkMacs and networkLocs what is the purpose of true values? Could it be false?
//		NetworkMacs:    make(map[string]map[string]bool), //e.g.: {"0":["MAC1":true,"MAC2":true,...]}
//		NetworkLocs:    make(map[string]map[string]bool), //e.g.: {"0":["P1":true,"P2":true,...]}
//		MacCount:       make(map[string]int),             //number of fingerprints of an AP(mac) in all locations; e.g. : {"MAC1":10,"Mac2":15,...}
//		MacCountByLoc:  make(map[string]map[string]int),  //e.g.: {"P1":{"MAC1":10,"MAC2":14},"P2":{MacCount2},}
//		UniqueMacs:     []string{},                       //UniqueMacs is an array of AP's macs
//		UniqueLocs:     []string{},                       //UniqueLocs is an array of map's locations e.g.: ["P1","P2","P3",...]
//		BayesPriors:         make(map[string]PriorParameters),
//		MacVariability: make(map[string]float32), //the standard deviation of rssi of each mac
//		BayesResults:        make(map[string]ResultsParameters),
//		Loaded:         false, //is true if ps was created and save in resources
//	}
//}

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

type UserPositionJSON struct {
	Time        int64              `json:"time"`
	Location    string             `json:"location"`
	BayesGuess  string             `json:"bayesguess"`
	BayesData   map[string]float64 `json:"bayesdata"`
	SvmGuess    string             `json:"svmguess"`
	SvmData     map[string]float64 `json:"svmdata"`
	ScikitData  map[string]string  `json:"rfdata"`
	KnnGuess    string             `json:"knnguess"`
	KnnData     map[string]float64 `json:"knndata"` // fpTime --> 1/(1+RssVectordistance) or weight
	PDRLocation string             `json:"pdrlocation"`
	Fingerprint Fingerprint        `json:"fingerprint"` // raw fingerprint data
}

type TestValidTrack struct {
	TrueLocation string           `json:"truelocation"`
	UserPosition UserPositionJSON `json:"userposition"`
}

//	filterMacs is used for set filtermacs
type FilterMacs struct {
	Group string   `json:"group"`
	Macs  []string `json:"macs"`
}

//// NewPersistentParameters returns the peristent parameters initialization
//func NewPersistentParameters() *PersistentParameters {
//	return &PersistentParameters{
//		NetworkRenamed: make(map[string][]string),
//	}
//}

// returns compress state of res.MarshalJSON
//func DumpParameters(res FullParameters) []byte {
//	jsonByte, _ := res.MarshalJSON()
//	return glb.CompressByte(jsonByte)
//}

//// UnmarshalJson a FullParameters
//func LoadParameters(jsonByte []byte) FullParameters {
//	var res2 FullParameters
//	res2.UnmarshalJSON(glb.DecompressByte(jsonByte))
//	return res2
//}




