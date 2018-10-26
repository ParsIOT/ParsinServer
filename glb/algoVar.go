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
var DefaultGraphFactorsRange [][]float64

var DefaultMapName string
var DefaultMapDimensions []int
var DefaultMapWidth int
var DefaultMapHeight int


var DefaultMaxMovement float64

var PDREnabledForDynamicSubareaMethod bool

var DefaultGraphEnabled, DefaultDSAEnabled bool
var DefaultMaxEuclideanRssDist int // must be deprecated

var FastLearn bool //ignore some crossvalidation calculation(rss regulating & get rss avg of adjacency dots) to learn fast

var NewDistAlgo bool

var TesterUsername string

func init() {
	DefaultMixin = float64(0.1)
	DefaultCutoff = float64(0.01)
	MinApNum = 1
	MinRssi = -110
	MaxRssi = 5
	DefaultMaxEuclideanRssDist = 30 //=ble, 50=wifi
	ProgressBarLength = 0
	ProgressBarCurLevel = 0
	DefaultKnnKRange = []int{25, 26}                //{10,30}
	DefaultKnnMinClusterRssRange = []int{-75, -76}  //{-60,-90}
	DefaultMaxEuclideanRssDistRange = []int{30, 50} // wifi:50, ble:30
	DefaultMaxMovementRange = []int{10, 100}
	DefaultGraphFactorsRange = [][]float64{{1, 1, 1, 1}, {2, 2, 2, 1}}

	//MinClusterRss = -75
	MaxUserHistoryLen = 2
	MaxUserResultsLen = 1000

	UserHistoryEffectFactors = []float64{0.01, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.5, 0.7, 0.8, 1}
	UserHistoryGaussVariance = 0.15
	UserHistoryTimeDelayFactor = 10000
	DefaultMaxMovement = float64(10000)

	PreprocessOutlinePercent = float64(0.333) // third part of fingerprints are considered as outline
	NormalRssDev = 5
	RssRegulation = true
	AvgRSSAdjacentDots = true

	DefaultMapName = "DefaultMap.png"
	DefaultMapDimensions = []int{3400,3600}
	DefaultMapHeight = 3400
	DefaultMapWidth = 3600
	DefaultGraphEnabled = true
	DefaultDSAEnabled = false
	PDREnabledForDynamicSubareaMethod = false

	FastLearn = false
	NewDistAlgo = false
	TesterUsername = "tester"
	MinRssClustringEnabled = true
}