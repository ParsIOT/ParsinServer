package dbm

import (
	"testing"
	"ParsinServer/algorithms/parameters"
	"github.com/stretchr/testify/assert"
)

func TestPutFingerprintIntoDatabase(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	_,fingerprintsInMemory1,_ := GetLearnFingerPrints(testdb,false)
	//glb.Debug.Println(len(fingerprintsInMemory1))

	fp := parameters.Fingerprint{
		Group:           testdb,
		Username:        "hadi",
		Location:        "NaN,NaN",
		Timestamp:       123456789,
		WifiFingerprint: []parameters.Router{
			parameters.Router{
				Mac:  "b4:52:7d:26:e3:f3",
				Rssi: -45,
			},
			parameters.Router{
				Mac:  "14:51:7e:22:a1:e4",
				Rssi: -50,
			},
		},
	}
	PutFingerprintIntoDatabase(fp,"fingerprints")
	_,fingerprintsInMemory2,_ := GetLearnFingerPrints(testdb,false)
	//glb.Debug.Println(len(fingerprintsInMemory2))

	assert.Equal(t, len(fingerprintsInMemory2), len(fingerprintsInMemory1)+1)
}

func TestFilterFingerprint(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	fp := parameters.Fingerprint{
		Group:           testdb,
		Username:        "hadi",
		Location:        "NaN,NaN",
		Timestamp:       123456789,
		WifiFingerprint: []parameters.Router{
			parameters.Router{
				Mac:  "b4:52:7d:26:e3:f3",
				Rssi: -45,
			},
			parameters.Router{
				Mac:  "14:51:7e:22:a1:e4",
				Rssi: -50,
			},
		},
	}

	fpRes := parameters.Fingerprint{
		Group:           testdb,
		Username:        "hadi",
		Location:        "NaN,NaN",
		Timestamp:       123456789,
		WifiFingerprint: []parameters.Router{
			parameters.Router{
				Mac:  "b4:52:7d:26:e3:f3",
				Rssi: -45,
			},
		},
	}

	//glb.Debug.Println(fpRes)
	FilterFingerprint(&fp)
	//glb.Debug.Println(fp)
	assert.Equal(t, fp, fpRes)
}

func TestLoadFingerprint(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	fpByte := []byte{84,207,205,234,219,48,16,4,240,119,217,179,3,171,175,149,52,175,82,122,144,100,41,8,106,39,216,14,61,132,188,123,9,20,164,255,117,249,177,51,243,166,251,241,120,61,9,116,213,243,90,51,45,244,58,235,177,167,173,254,191,209,66,127,30,37,93,253,177,19,72,49,47,138,153,22,186,250,86,207,43,109,79,130,178,193,43,239,196,7,97,237,156,119,204,11,253,237,173,223,90,223,239,245,120,30,125,191,8,191,222,180,165,66,32,41,80,17,161,193,49,138,32,57,90,232,56,207,78,184,137,124,150,137,153,12,201,224,136,53,65,218,196,226,96,217,194,105,248,21,90,80,13,154,25,204,241,96,182,192,85,112,65,45,8,14,97,10,245,102,48,99,17,61,154,64,12,242,138,104,39,54,125,99,13,149,160,20,154,131,20,240,20,106,213,96,46,64,214,239,82,157,191,245,172,158,38,132,193,98,193,42,176,6,94,131,43,194,244,45,186,193,138,133,210,223,68,86,8,17,158,167,110,19,139,1,86,195,10,152,17,35,106,254,193,126,127,254,5,0,0,255,255}

	fpRes := parameters.Fingerprint{
		Group:           "testdb",
		Username:        "test",
		Location:        "100,100",
		Timestamp:       1487175678602557500,
		WifiFingerprint: []parameters.Router{
			parameters.Router{
				Mac:  "6c:19:8f:50:c6:a5",
				Rssi: -66,
			},
			parameters.Router{
				Mac:  "6c:3b:6b:09:da:6f",
				Rssi: -69,
			},
			parameters.Router{
				Mac:  "b4:52:7d:26:e3:f3",
				Rssi: -50,
			},
			parameters.Router{
				Mac:  "4c:5e:0c:ec:85:85",
				Rssi: -73,
			},
			parameters.Router{
				Mac:  "34:97:f6:63:bd:94",
				Rssi: -70,
			},
			parameters.Router{
				Mac:  "02:1a:11:f5:6c:03",
				Rssi: -41,
			},
			parameters.Router{
				Mac:  "58:6d:8f:2b:26:42",
				Rssi: -68,
			},
			parameters.Router{
				Mac:  "9c:d6:43:72:0e:83",
				Rssi: -95,
			},
			parameters.Router{
				Mac:  "c4:12:f5:01:89:70",
				Rssi: -75,
			},
			parameters.Router{
				Mac:  "98:42:46:00:99:eb",
				Rssi: -75,
			},
		},
	}

	fpResFiltered := parameters.Fingerprint{
		Group:           "testdb",
		Username:        "test",
		Location:        "100,100",
		Timestamp:       1487175678602557500,
		WifiFingerprint: []parameters.Router{
			parameters.Router{
				Mac:  "b4:52:7d:26:e3:f3",
				Rssi: -50,
			},
		},
	}

	fp := LoadFingerprint(fpByte, false)
	fpFiltered := LoadFingerprint(fpByte, true)

	//glb.Debug.Println(fp)
	//glb.Debug.Println(fpRes)

	assert.Equal(t, fp, fpRes)

	//glb.Debug.Println(fpFiltered)
	//glb.Debug.Println(fpResFiltered)

	assert.Equal(t, fpFiltered, fpResFiltered)
}