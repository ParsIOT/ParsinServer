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

var ProgressBarLength,ProgressBarCurLevel int

// Default K in KNN algorithm
var DefaultKnnK int

func init() {
	DefaultMixin = float64(0.1)
	DefaultCutoff = float64(0.01)
	MinApNum = 3
	MinRssi = -110
	MaxRssi = 5
	DefaultKnnK = 10
	ProgressBarLength = 0
	ProgressBarCurLevel = 0
	//MinClusterRss = -75
}