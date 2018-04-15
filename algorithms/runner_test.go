package algorithms

import (
	"testing"
	"github.com/gin-gonic/gin"
	"bytes"
	"net/http/httptest"
	"net/http"
)

func BenchmarkTrackFingerprintRoute(b *testing.B) {
	//jsonTest := `{"username": "zack", "group": "testdb", "wifi-fingerprint": [{"rssi": -45, "mac": "80:37:73:ba:f7:d8"}, {"rssi": -58, "mac": "80:37:73:ba:f7:dc"}, {"rssi": -61, "mac": "a0:63:91:2b:9e:65"}, {"rssi": -68, "mac": "a0:63:91:2b:9e:64"}, {"rssi": -70, "mac": "70:73:cb:bd:9f:b5"}, {"rssi": -75, "mac": "d4:05:98:57:b3:10"}, {"rssi": -75, "mac": "00:23:69:d4:47:9f"}, {"rssi": -76, "mac": "30:46:9a:a0:28:c4"}, {"rssi": -81, "mac": "2c:b0:5d:36:e3:b8"}, {"rssi": -82, "mac": "00:1a:1e:46:cd:10"}, {"rssi": -82, "mac": "20:aa:4b:b8:31:c8"}, {"rssi": -83, "mac": "e8:ed:05:55:21:10"}, {"rssi": -83, "mac": "ec:1a:59:4a:9c:ed"}, {"rssi": -88, "mac": "b8:3e:59:78:35:99"}, {"rssi": -84, "mac": "e0:46:9a:6d:02:ea"}, {"rssi": -84, "mac": "00:1a:1e:46:cd:11"}, {"rssi": -84, "mac": "f8:35:dd:0a:da:be"}, {"rssi": -84, "mac": "b4:75:0e:03:cd:69"}], "location": "zakhome floor 2 office", "time": 1439596533831, "password": "frusciante_0128"}`
	 jsonTest := `{ "group":"arman_20_7_96_ble_2", "username":"hadi", "time":12309123, "wifi-fingerprint":[ { "mac":"b4:52:7d:26:e3:f3", "rssi":-21 }, { "mac":"b4:52:7d:26:e3:f4", "rssi":-71 }, { "mac":"01:17:C5:97:5B:1D", "rssi":-62 }, { "mac":"01:17:C5:97:E7:B3", "rssi":-85 }, { "mac":"01:17:C5:97:1B:44", "rssi":-78 }, { "mac":"01:17:C5:97:B5:70", "rssi":-82 }, { "mac":"01:17:C5:97:87:84", "rssi":-84 }, { "mac":"01:17:C5:97:58:C3", "rssi":-88 } ] }`

	router := gin.New()
	router.POST("/foo", TrackFingerprintPOST)

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/foo?group=arman_20_7_96_ble_2", bytes.NewBufferString(jsonTest))
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
}
