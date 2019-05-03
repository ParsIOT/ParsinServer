package glb

// MinRssiOpt is the minimum level of signal for learning and tracking
var MinRssiOpt int

// MaxRssi is the maximum level of signal
var MaxRssi int

// MinRssi is the minimum level of signal that can save to db
var MinRssi int

// RssiRange is the calculated partitions in array form
var RssiRange []float32

var DefaultCutoff,DefaultMixin float64
var MinApNum int

var ParticleFilterEnabled bool

var PreprocessOutlinePercent float64 // percentage of a location fingerprints that its received rss considered as outline
var NormalRssDev int                 // Normal deviation rss from median
var RssRegulation bool               // permit to rss outlines deleting
var AvgRSSAdjacentDots bool          // permit to set average of rss of adjacent dots instead of raw rss; according to heatmap it's not good to do this!

var ProgressBarLength,ProgressBarCurLevel int
var MinRssClustringEnabled bool

var MaxUserHistoryLen int
var MaxUserResultsLen int
var UserHistoryEffectFactors []float64

var UserHistoryGaussVariance float64
var UserHistoryTimeDelayFactor float64
// Default K in KNN algorithm
var DefaultKnnMinClusterRssRange, DefaultKnnKRange, DefaultMaxEuclideanRssDistRange, DefaultMaxMovementRange []int
var DefaultBLEFactorRange []float64
var DefaultGraphFactorsRange [][]float64
var DefaultRPFRadius float64

var DefaultMapName string
var DefaultMapDimensions []int
var DefaultMapWidth int
var DefaultMapHeight int

var PDREnabledForDynamicSubareaMethod bool

var DefaultGraphEnabled, DefaultDSAEnabled, DefaultRPFEnabled bool
var DefaultSimpleHistoryEnabled bool //Enable SimpleHistoryEffect after running ML algo (for details See SimpleHistoryEffect function)
var DefaultGraphStep float64

var FastLearn bool //ignore some crossvalidation calculation(rss regulating & get rss avg of adjacency dots) to learn fast

var NewDistAlgo string

var TesterUsername string

const (
	KNN                   string = "KNN"
	BAYES                 string = "BAYES"
	SVM                   string = "SVM"
	SCIKIT_REGRESSION     string = "SCIKIT_REGRESSION"
	SCIKIT_CLASSIFICATION string = "SCIKIT_CLASSIFICATION"
)

var ALLALGORITHMS = []string{KNN, BAYES, SVM, SCIKIT_CLASSIFICATION, SCIKIT_REGRESSION}

var MainPositioningAlgo string

func init() {
	DefaultMixin = float64(0.1)
	DefaultCutoff = float64(0.01)
	MinApNum = 3
	MinRssi = -100
	MaxRssi = 5
	ProgressBarLength = 0
	ProgressBarCurLevel = 0
	DefaultKnnKRange = []int{5}                 //{10,30}
	DefaultKnnMinClusterRssRange = []int{-65}   //{-60,-90}
	DefaultMaxEuclideanRssDistRange = []int{16} // wifi:50, ble:30
	DefaultMaxMovementRange = []int{100, 1000}
	DefaultGraphFactorsRange = [][]float64{{1, 1, 1, 1}, {2, 2, 2, 1}}
	DefaultBLEFactorRange = []float64{1} //{1.0, 1.2, 0.1}
	DefaultGraphStep = 1.0
	DefaultRPFRadius = 100

	//MinClusterRss = -75
	MaxUserHistoryLen = 6 //Note: I don't know why i change it to 3, location not changed very well in live location map!
	MaxUserResultsLen = 100

	UserHistoryEffectFactors = []float64{0.01, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.5, 0.7, 0.8, 1}
	UserHistoryGaussVariance = 0.15
	UserHistoryTimeDelayFactor = 10000

	PreprocessOutlinePercent = float64(0.333) // third part of finger	prints are considered as outline
	NormalRssDev = 5
	RssRegulation = true
	AvgRSSAdjacentDots = true

	DefaultMapName = "ArmanExactMap.png"
	DefaultMapDimensions = []int{3400,3600}
	DefaultMapHeight = 3400
	DefaultMapWidth = 3600
	DefaultGraphEnabled = false
	DefaultDSAEnabled = false
	DefaultRPFEnabled = false
	PDREnabledForDynamicSubareaMethod = false

	DefaultSimpleHistoryEnabled = false
	FastLearn = false
	TesterUsername = "tester"
	MinRssClustringEnabled = true

	MainPositioningAlgo = KNN // see runner.go

	ParticleFilterEnabled = true
}