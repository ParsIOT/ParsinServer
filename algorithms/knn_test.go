package algorithms

import (
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"fmt"
	"testing"
	"time"
)

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}

func TestTrackKnn(t *testing.T) {
	//gp := dbm.GM.GetGroup("arman_20_7_96_ble_2")

	//glb.Debug.Println(len(fingerprintsInMemory1))
	defer elapsed("test ddf")
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
	fpTemp := fp
	fpTemp.WifiFingerprint = nil

	glb.Debug.Println(fp.WifiFingerprint)
	glb.Debug.Println(fpTemp.WifiFingerprint)
	for i := 0; i < len(fp.WifiFingerprint)-1; i++ {
		currRouter := fp.WifiFingerprint[i]
		for j := i + 1; j < len(fp.WifiFingerprint); j++ {
			var r parameters.Router
			nextRouter := fp.WifiFingerprint[j]
			r.Mac = fmt.Sprintf("%v#%v", currRouter.Mac, nextRouter.Mac)
			r.Rssi = currRouter.Rssi - nextRouter.Rssi
			fpTemp.WifiFingerprint = append(fpTemp.WifiFingerprint, r)
		}

	}
	/*	if strings.Contains(fpTemp.WifiFingerprint[0].Mac, "#") {
		fmt.Println("string contains #")
	}*/
	//fmt.Println(len(fp.WifiFingerprint))
	/*	sort.Slice(fpTemp.WifiFingerprint, func(i, j int) bool {
		return fpTemp.WifiFingerprint[i].Mac < fpTemp.WifiFingerprint[j].Mac
	})*/
	//fmt.Println(fpTemp.WifiFingerprint)
	/*	for i := 0; i < 10000; i++ {
			_, resultDot, _ := TrackKnn(gp, fp, false)

			if (resultDot != "-12.000004,1411.999993") {

				glb.Debug.Println(resultDot)
				assert.Equal(t, resultDot, "-12.000004,1411.999993")
			}
		}
	*/
}
