// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// parameters.go contains structures and functions for setting and getting Naive-Bayes parameters.

package parameters

import (
	"ParsinServer/glb"
	"strings"
)

// PersistentParameters are not reloaded each time
//type PersistentParameters struct {
//	NetworkRenamed map[string][]string // key:networkName, value:mac list; e.g.: {"1":["mac1","mac2"]}
//}

type Dot struct {
	X    float64
	Y    float64
	XStr string
	YStr string
}

func NewDot(xStr string, yStr string) Dot {
	x, _ := glb.StringToFloat(xStr)
	y, _ := glb.StringToFloat(yStr)
	return Dot{
		X:    x,
		Y:    y,
		XStr: xStr,
		YStr: yStr,
	}
}
func NewDotFromString(xyStr string) Dot {
	xyStrSplited := strings.Split(xyStr, ",")
	if !(len(xyStrSplited) == 2) {
		glb.Error.Println("Location names aren't in the format of x,y")
	}
	xStr := xyStrSplited[0]
	yStr := xyStrSplited[1]
	x, _ := glb.StringToFloat(xStr)
	y, _ := glb.StringToFloat(yStr)
	return Dot{
		X:    x,
		Y:    y,
		XStr: xStr,
		YStr: yStr,
	}
}

// Constant parameters that set manually by user are in KnnConfig
type KnnConfig struct {
	KRange                   []int
	MinClusterRssRange       []int
	MaxEuclideanRssDistRange []int

	GraphEnabled     bool
	GraphFactorRange [][]float64

	DSAEnabled       bool
	MaxMovementRange []int
	BLEFactorRange   []float64

	RPFEnabled     bool
	RPFRadiusRange []float64
}

func NewKnnConfig() KnnConfig {
	return KnnConfig{
		KRange:                   glb.DefaultKnnKRange,
		MinClusterRssRange:       glb.DefaultKnnMinClusterRssRange,
		MaxEuclideanRssDistRange: glb.DefaultMaxEuclideanRssDistRange,

		GraphEnabled:     glb.DefaultGraphEnabled,
		GraphFactorRange: glb.DefaultGraphFactorsRange,

		DSAEnabled:       glb.DefaultDSAEnabled,
		MaxMovementRange: glb.DefaultMaxMovementRange,
		BLEFactorRange:   glb.DefaultBLEFactorRange,

		RPFEnabled:     glb.DefaultRPFEnabled,
		RPFRadiusRange: glb.DefaultRPFRadiusRange,
	}
}

const (
	CoGroupState_None   int = 0
	CoGroupState_Master int = 1
	CoGroupState_Slave  int = 2
)

// Other group configs that aren't in knnconfig and ...
type OtherGroupConfig struct {
	CoGroup               string
	CoGroupMode           int // 0: none, 1:master(main) group , 2: slave group
	SimpleHistoryEnabled  bool
	ParticleFilterEnabled bool
}

func NewOtherGroupConfig() OtherGroupConfig {
	return OtherGroupConfig{
		CoGroup:               "",
		CoGroupMode:           CoGroupState_None,
		SimpleHistoryEnabled:  glb.DefaultSimpleHistoryEnabled,
		ParticleFilterEnabled: false,
	}
}

// Constant parameters that set by cross-validation are in KnnHyperParameters
type KnnHyperParameters struct {
	K                   int       `json:"K"`
	MinClusterRss       int       `json:"MinClusterRss"`
	MaxEuclideanRssDist int       `json:"MaxEuclideanRssDist"`
	MaxMovement         int       `json:"MaxMovement"`
	GraphFactors        []float64 `json:"GraphFactors"`
	RPFRadius           float64   `json:"RPFRadius"`
	BLEFactor           float64   `json:"BLEFactor"`
}

func NewKnnHyperParameters() KnnHyperParameters {
	return KnnHyperParameters{
		K:                   glb.DefaultKnnKRange[0],
		MinClusterRss:       glb.DefaultKnnMinClusterRssRange[0],
		MaxEuclideanRssDist: glb.DefaultMaxEuclideanRssDistRange[0],
		MaxMovement:         glb.DefaultMaxMovementRange[0],
		GraphFactors:        glb.DefaultGraphFactorsRange[0],
		RPFRadius:           glb.DefaultRPFRadiusRange[0],
		BLEFactor:           glb.DefaultBLEFactorRange[0],
	}
}

type KnnFingerprints struct {
	FingerprintsInMemory map[string]Fingerprint `json:"FingerprintsInMemory"`
	FingerprintsOrdering []string               `json:"FingerprintsOrdering"`
	Clusters             map[string][]string    `json:"Clusters"`
	HyperParameters      KnnHyperParameters     `json:"HyperParameters"`
	Node2FPs             map[string][]string    `json:"Node2FPs"`
	RPFs                 map[string]float64     `json:"RPFs"`
}

func NewKnnFingerprints() KnnFingerprints {
	return KnnFingerprints{
		FingerprintsInMemory: make(map[string]Fingerprint),
		FingerprintsOrdering: []string{},
		Clusters:             make(map[string][]string),
		HyperParameters:      NewKnnHyperParameters(),
		Node2FPs:             make(map[string][]string),
		RPFs:                 make(map[string]float64),
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
	Time            int64              `json:"time"`
	Location        string             `json:"location"`    // it contains historyeffected main algorithm guess
	RawLocation     string             `json:"rawlocation"` // it contains raw main algorithm guess
	BayesGuess      string             `json:"bayesguess"`
	BayesData       map[string]float64 `json:"bayesdata"`
	SvmGuess        string             `json:"svmguess"`
	SvmData         map[string]float64 `json:"svmdata"`
	ScikitData      map[string]string  `json:"rfdata"`
	KnnGuess        string             `json:"knnguess"`
	KnnData         map[string]float64 `json:"knndata"` // fpTime --> 1/(1+RssVectordistance) or weight
	Confidentiality float64            `json:"confidentiality"`
	PDRLocation     string             `json:"pdrlocation"`
	Fingerprint     Fingerprint        `json:"fingerprint"` // raw fingerprint data
}

type TestValidTrack struct {
	TrueLocation string           `json:"truelocation"`
	UserPosition UserPositionJSON `json:"userposition"`
}

type TrueLocation struct {
	Timestamp int64  `json:"timestamp"`
	Location  string `json:"location"`
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

func ConvertSharpToUnderlineInFP(routers []Router) []Router {
	newRouters := []Router{}
	for _, rt := range routers {
		rt.Mac = strings.Replace(rt.Mac, "#", "_", -1)
		newRouters = append(newRouters, rt)
	}
	return newRouters
}
