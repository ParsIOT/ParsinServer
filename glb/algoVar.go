package glb

// MinRssiOpt is the minimum level of signal for learning and tracking
var MinRssiOpt int

// MaxRssi is the maximum level of signal
var MaxRssi int

// MinRssi is the minimum level of signal that can save to db
var MinRssi int

// RssiRange is the calculated partitions in array form
var RssiRange []float32


var MinApNum int

// Default K in KNN algorithm
var DefaultKnnK,MinClusterRss int

func init() {
	MinApNum = 3
	MinRssi = -110
	MaxRssi = 5
	DefaultKnnK = 60
	MinClusterRss = -70
}