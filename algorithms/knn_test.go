package algorithms

import (
	"testing"
	"ParsinServer/dbm/parameters"
	"ParsinServer/dbm"
	"ParsinServer/glb"
	"github.com/stretchr/testify/assert"
)

func TestTrackKnn(t *testing.T){
	gp := dbm.GM.GetGroup("arman_20_7_96_ble_2")

	//glb.Debug.Println(len(fingerprintsInMemory1))

	fp := parameters.Fingerprint{
		Group:           "arman_20_7_96_ble_2",
		Username:        "hadi",
		Location:        "-12.000000,1412.000000",
		Timestamp:       1507803686841705438,
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
	for i:=0;i<10000;i++{
		_, resultDot, _ := TrackKnn(gp, fp, false)

		if (resultDot != "-12.000004,1411.999993"){

			glb.Debug.Println(resultDot)
			assert.Equal(t, resultDot, "-12.000004,1411.999993")
		}
	}



}