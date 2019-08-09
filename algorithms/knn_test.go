package algorithms

import (
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestTrackKnn(t *testing.T) {
	gp := dbm.GM.GetGroup("arman_20_7_96_ble_2")

	//glb.Debug.Println(len(fingerprintsInMemory1))

	fp := parameters.Fingerprint{
		Group:     "arman_20_7_96_ble_2",
		Username:  "hadi",
		Location:  "-12.000000,1412.000000",
		Timestamp: 1507803686841705438,
		WifiFingerprint: []parameters.Router{

			parameters.Router{
				Mac:  "01:17:C5:97:87:84",
				Rssi: -75,
			},
			parameters.Router{
				Mac:  "01:17:C5:97:B5:70",
				Rssi: -87,
			},
			parameters.Router{
				Mac:  "01:17:C5:97:1B:44",
				Rssi: -84,
			},
			parameters.Router{
				Mac:  "01:17:C5:97:E7:B3",
				Rssi: -92,
			},
			parameters.Router{
				Mac:  "01:17:C5:97:5B:1D",
				Rssi: -89,
			},
			parameters.Router{
				Mac:  "01:17:C5:97:58:C3",
				Rssi: -63,
			},
			parameters.Router{
				Mac:  "01:17:C5:97:44:BE",
				Rssi: -78,
			},
			parameters.Router{
				Mac:  "01:17:C5:97:DE:E8",
				Rssi: -96,
			},
		},
	}
	for i := 0; i < 10000; i++ {
		_, resultDot, _ := TrackKnn(gp, fp, false)

		if resultDot != "-12.000004,1411.999993" {

			glb.Debug.Println(resultDot)
			assert.Equal(t, resultDot, "-12.000004,1411.999993")
		}
	}

}

func TestConvertDist2Wigth(t *testing.T) {
	glb.Debug.Println("123")
	distMap := make(map[string]float64)
	for i := 15; i < 20; i++ {
		distMap[strconv.Itoa(i)] = float64(i)
	}
	fmt.Println(distMap)
	distMapResult := ConvertDist2Wigth(distMap)
	fmt.Println(distMapResult)
	//assert(t,distMapResult,distMap)

}

func TestTriangulateWith3Point(t *testing.T) {
	x1 := float64(0)
	y1 := float64(0)
	r1 := float64(1)

	x2 := float64(1)
	y2 := float64(1)
	r2 := float64(1)

	x3 := float64(3)
	y3 := float64(0)
	r3 := float64(1.5)

	traingulationVals := []float64{x1, y1, r1, x2, y2, r2, x3, y3, r3}

	mid12X, mid12Y := GetMiddleOfLine([]float64{x1, y1, 1 / r1, x2, y2, 1 / r2})
	mid23X, mid23Y := GetMiddleOfLine([]float64{x2, y2, 1 / r2, x3, y3, 1 / r3})
	mid13X, mid13Y := GetMiddleOfLine([]float64{x1, y1, 1 / r1, x3, y3, 1 / r3})

	glb.Debug.Println(mid12X, ",", mid12Y)
	glb.Debug.Println(mid23X, ",", mid23Y)
	glb.Debug.Println(mid13X, ",", mid13Y)

	resultX, resultY := TriangulateWith3Point(traingulationVals)
	glb.Debug.Println(resultX, ",", resultY)

	distFromMid12 := glb.CalcDist(mid12X, mid12Y, resultX, resultY)
	distFromMid23 := glb.CalcDist(mid23X, mid23Y, resultX, resultY)
	distFromMid13 := glb.CalcDist(mid13X, mid13Y, resultX, resultY)

	glb.Debug.Println(distFromMid12)
	glb.Debug.Println(distFromMid23)
	glb.Debug.Println(distFromMid13)

	//glb.Debug.Println((math.Pow(r1,2)+math.Pow(r2,2)+math.Pow(r3,2))/3)
	assert.Equal(t, true, false)
}

func TestGetMiddleOfLine(t *testing.T) {
	x1 := float64(0)
	y1 := float64(0)
	w1 := float64(2)

	x2 := float64(1)
	y2 := float64(1)
	w2 := float64(4000)

	dotVals := []float64{x1, y1, w1, x2, y2, w2}

	resultX, resultY := GetMiddleOfLine(dotVals)
	glb.Debug.Println(resultX, ",", resultY)
	assert.Equal(t, true, false)
}
