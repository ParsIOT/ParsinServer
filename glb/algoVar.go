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

var ProgressBarLength,ProgressBarCurLevel int

var MaxUserHistoryLen int
var MaxUserResultsLen int
var UserHistoryEffectFactors []float64

var UserHistoryGaussVariance float64
var UserHistoryTimeDelayFactor float64
// Default K in KNN algorithm
var DefaultKnnMinCRssRange,DefaultKnnKRange []int

var MaxMovement float64

func init() {
	DefaultMixin = float64(0.1)
	DefaultCutoff = float64(0.01)
	MinApNum = 1
	MinRssi = -110
	MaxRssi = 5
	ProgressBarLength = 0
	ProgressBarCurLevel = 0
	DefaultKnnKRange = []int{25, 26}         //{10,30}
	DefaultKnnMinCRssRange = []int{-75, -76} //{-60,-90}
	//MinClusterRss = -75
	MaxUserHistoryLen = 2
	MaxUserResultsLen = 100

	UserHistoryEffectFactors = []float64{0.01, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.5, 0.7, 0.8, 1}
	UserHistoryGaussVariance = 0.15
	UserHistoryTimeDelayFactor = 10000
	MaxMovement = float64(400)

	PreprocessOutlinePercent = float64(0.333) // third part of fingerprints are considered as outline
	NormalRssDev = 5
}