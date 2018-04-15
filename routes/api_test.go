package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"strings"
	"path"
	"os"
	"fmt"
	"os/exec"
	"log"
	"ParsinServer/glb"
	"reflect"
	"sync"
	"strconv"
	"bytes"
	"encoding/json"
)

type Empty struct{}

var DataPath string

type Lock struct {
	sync.RWMutex
}

var lock Lock

var testCount int

func getTestCount() int {
	testCount += 1
	return testCount
}

func gettestdbName() string{
	testCount := getTestCount()
	initRaw(testCount)
	testdbName := "testdb"+strconv.Itoa(testCount)
	return testdbName
}

func freedb(testdb string){
	os.Remove(path.Join(DataPath,testdb+".db"))
}

func initRaw(testCount int){
	lock.Lock()
	newName := "testdb"+strconv.Itoa(testCount)+".db"
	_, err := exec.Command("cp", []string{path.Join(DataPath, "testdb.db.backup"),path.Join(DataPath, newName)}...).Output()
	if err != nil {
		log.Fatal(err)
	}
	lock.Unlock()
}
func init() {
	testCount = 0
	DataPath = ""
	gin.SetMode(gin.ReleaseMode)
	cwd, _ := os.Getwd()
	pkgName := reflect.TypeOf(Empty{}).PkgPath()
	projName := strings.Split(pkgName,"/")[0]
	for _,p := range strings.Split(cwd,"/") {
		if p == projName {
			DataPath += p+"/"
			break
		}
		DataPath += p +"/"
	}
	DataPath = path.Join(DataPath, "data")
	glb.Debug.Println(DataPath)
}



func TestGetStatus(t *testing.T) {
	router := gin.New()
	router.PUT("/foo", GetStatus)
	req, _ := http.NewRequest("PUT", "/foo", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, strings.Contains(resp.Body.String(), "\"success\":true"), true)
}

func TestMigrateDatabase(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.PUT("/foo", MigrateDatabase)

	req, _ := http.NewRequest("PUT", "/foo?from="+testdb+"&to=newdb", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), "{\"message\":\"Successfully migrated "+testdb+" to newdb\",\"success\":true}")
	fmt.Println(DataPath)
	os.Remove(path.Join(DataPath,"newdb.db"))

}

func TestDeleteDatabase(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	glb.CopyFile(path.Join(DataPath,testdb+".db"), path.Join(DataPath,"deleteme.db"))

	router := gin.New()
	router.DELETE("/foo", DeleteDatabase)

	req, _ := http.NewRequest("DELETE", "/foo?group=deleteme", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), "{\"message\":\"Successfully deleted deleteme\",\"success\":true}")
}

func TestCalculate(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", Calculate)

	req, _ := http.NewRequest("GET", "/foo?group="+testdb, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), "{\"message\":\"Parameters optimized.\",\"success\":true}")
	os.Remove(path.Join(DataPath,testdb+".db"))
}

func TestGetLocationList(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", GetLocationList)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	respond := "{\"locations\":{\"p1\":{\"accuracy\":54,\"count\":11},\"p3\":{\"accuracy\":88,\"count\":9}},\"message\":\"Found 2 unique locations in group "+testdb+"\",\"success\":true}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), respond)
}

func TestGetLastFingerprint(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", GetLastFingerprint)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb+"&user=hadi", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	respond := "sent as /track\n{\n \"group\": \"testdb\",\n \"username\": \"hadi\",\n \"location\": \"location\",\n \"timestamp\": 1504523576678432200,\n \"wifi-fingerprint\": [\n  {\n   \"mac\": \"58:6d:8f:2b:29:6c\",\n   \"rssi\": -70\n  },\n  {\n   \"mac\": \"bc:85:56:64:4e:e6\",\n   \"rssi\": -79\n  },\n  {\n   \"mac\": \"d2:13:fd:56:8a:ed\",\n   \"rssi\": -81\n  },\n  {\n   \"mac\": \"6c:3b:6b:9e:5a:69\",\n   \"rssi\": -54\n  },\n  {\n   \"mac\": \"e4:8d:8c:c1:f2:a9\",\n   \"rssi\": -52\n  },\n  {\n   \"mac\": \"6c:3b:6b:09:da:6f\",\n   \"rssi\": -95\n  },\n  {\n   \"mac\": \"18:d6:c7:78:ec:9b\",\n   \"rssi\": -53\n  }\n ]\n}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), respond)
}


func TestGetHistoricalUserPositions(t *testing.T) {
	// todo : Problem!!!
	testdb := gettestdbName()
	defer freedb(testdb)

	output := GetHistoricalUserPositions(testdb,"hadi",4)
	out, err := json.Marshal(&output)
	if err != nil {
		panic (err)
	}
	outStr := string(out)
	respond := "[{\"time\":\"2018-02-10 20:08:54.108990031 +0330 +0330\",\"bayesguess\":\"100,100\",\"bayesdata\":{\"100,100\":0.35355339059327373,\"300,300\":-0.35355339059327373},\"svmguess\":null,\"svmdata\":null,\"rfdata\":null,\"knnguess\":null},{\"time\":\"2018-02-10 20:08:53.545325469 +0330 +0330\",\"bayesguess\":\"100,100\",\"bayesdata\":{\"100,100\":0.35355339059327373,\"300,300\":-0.35355339059327373},\"svmguess\":null,\"svmdata\":null,\"rfdata\":null,\"knnguess\":null},{\"time\":\"2018-02-10 20:08:53.092894361 +0330 +0330\",\"bayesguess\":\"100,100\",\"bayesdata\":{\"100,100\":0.35355339059327373,\"300,300\":-0.35355339059327373},\"svmguess\":null,\"svmdata\":null,\"rfdata\":null,\"knnguess\":null},{\"time\":\"2018-02-10 20:08:52.589899518 +0330 +0330\",\"bayesguess\":\"100,100\",\"bayesdata\":{\"100,100\":0.35355339059327373,\"300,300\":-0.35355339059327373},\"svmguess\":null,\"svmdata\":null,\"rfdata\":null,\"knnguess\":null}]373},\"svmguess\":null,\"svmdata\":null,\"rfdata\":null,\"knnguess\":null}]"
	assert.Equal(t, outStr, respond)
}


func TestGetUserLocations(t *testing.T) {
	router := gin.New()
	router.GET("/foo", GetUserLocations)

	req, _ := http.NewRequest("GET", "/foo", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), "{\"message\":\"Error parsing request\",\"success\":false}")
}

func TestGetUserLocations2(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", GetUserLocations)

	req, _ := http.NewRequest("GET", "/foo?group="+testdb+"&user=hadi", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, strings.Contains(resp.Body.String(), "{\"message\":\"Correctly found locations.\""), true)
}

func TestPutMixinOverrideBad(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.PUT("/foo", PutMixinOverride)

	req, _ := http.NewRequest("PUT", "/foo?group="+testdb+"&mixin=100", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := `{"message":"mixin must be between 0 and 1","success":false}`
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestPutMixinOverrideGood(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.PUT("/foo", PutMixinOverride)
	req, _ := http.NewRequest("PUT", "/foo?group="+testdb+"&mixin=0", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Overriding mixin for "+testdb+", now set to 0\",\"success\":true}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}


func TestPutCutoffOverrideBad(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.PUT("/foo", PutCutoffOverride)

	req, _ := http.NewRequest("PUT", "/foo?group="+testdb+"&cutoff=100", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := `{"message":"cutoff must be between 0 and 1","success":false}`
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestPutCutoffOverrideGood(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.PUT("/foo", PutCutoffOverride)
	req, _ := http.NewRequest("PUT", "/foo?group="+testdb+"&cutoff=0", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Overriding cutoff for "+testdb+", now set to 0\",\"success\":true}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}



func TestPutKnnK(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.PUT("/foo", PutKnnK)
	req, _ := http.NewRequest("PUT", "/foo?group="+testdb+"&knnK=10", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Overriding KNN K for "+testdb+", now set to 10\",\"success\":true}"


	freedb(testdb)
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}



func TestPutMinRss(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.PUT("/foo", PutMinRss)
	req, _ := http.NewRequest("PUT", "/foo?group="+testdb+"&minRss=-100", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Overriding Minimum RSS for "+testdb+", now set to -100\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}


func TestEditNetworkName(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", EditNetworkName)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb+"&oldname=0&newname=home", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Finished\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestEditName(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", EditName)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb+"&location=p1&newname=p2", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Changed name of 50 things\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestEditMac(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", EditMac)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb+"&oldmac=b4:52:7d:26:e3:f3&newmac=b4:52:7d:26:e3:f4", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Changed name of 99 things\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestEditUserName(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", EditUserName)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb+"&user=hadi&newname=hadi", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Changed name of 6 things\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestDeleteUser(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.DELETE("/foo", DeleteUser)
	req, _ := http.NewRequest("DELETE", "/foo?group="+testdb+"&user=hadi", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Deletes 25 things  with user hadi\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestDeleteLocation(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.DELETE("/foo", DeleteLocation)
	req, _ := http.NewRequest("DELETE", "/foo?group="+testdb+"&location=100,100", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Deleted 50 locations\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestDeleteLocations(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.DELETE("/foo", DeleteLocations)
	req, _ := http.NewRequest("DELETE", "/foo?group="+testdb+"&names=100,100,200,200", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Deleted 100 locations\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}


func TestSetfiltermacs(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.POST("/foo", Setfiltermacs)
	jsonStr := []byte("{\"group\":\""+testdb+"\",\"macs\":[\"6c:19:8f:50:c6:a5\",\"b4:52:7d:26:e3:f3\"]}")
	req, _ := http.NewRequest("POST", "/foo", bytes.NewBuffer(jsonStr))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"MacFilter set successfully\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestGetfiltermacs(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", Getfiltermacs)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":[\"6c:19:8f:50:c6:a5\",\"b4:52:7d:26:e3:f3\"],\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestReformDB(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", ReformDB)
	req, _ := http.NewRequest("GET", "/foo?group="+testdb, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Changed name of 137 things\",\"success\":true}"

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}


func BenchmarkGetUserLocations(b *testing.B) {
	//jsonTest := `{"username": "zack", "group": "testdb", "wifi-fingerprint": [{"rssi": -45, "mac": "80:37:73:ba:f7:d8"}, {"rssi": -58, "mac": "80:37:73:ba:f7:dc"}, {"rssi": -61, "mac": "a0:63:91:2b:9e:65"}, {"rssi": -68, "mac": "a0:63:91:2b:9e:64"}, {"rssi": -70, "mac": "70:73:cb:bd:9f:b5"}, {"rssi": -75, "mac": "d4:05:98:57:b3:10"}, {"rssi": -75, "mac": "00:23:69:d4:47:9f"}, {"rssi": -76, "mac": "30:46:9a:a0:28:c4"}, {"rssi": -81, "mac": "2c:b0:5d:36:e3:b8"}, {"rssi": -82, "mac": "00:1a:1e:46:cd:10"}, {"rssi": -82, "mac": "20:aa:4b:b8:31:c8"}, {"rssi": -83, "mac": "e8:ed:05:55:21:10"}, {"rssi": -83, "mac": "ec:1a:59:4a:9c:ed"}, {"rssi": -88, "mac": "b8:3e:59:78:35:99"}, {"rssi": -84, "mac": "e0:46:9a:6d:02:ea"}, {"rssi": -84, "mac": "00:1a:1e:46:cd:11"}, {"rssi": -84, "mac": "f8:35:dd:0a:da:be"}, {"rssi": -84, "mac": "b4:75:0e:03:cd:69"}], "location": "zakhome floor 2 office", "time": 1439596533831, "password": "frusciante_0128"}`
	//jsonTest := `{ "group":"arman_20_7_96_ble_2", "username":"hadi", "time":12309123, "wifi-fingerprint":[ { "mac":"b4:52:7d:26:e3:f3", "rssi":-21 }, { "mac":"b4:52:7d:26:e3:f4", "rssi":-71 }, { "mac":"01:17:C5:97:5B:1D", "rssi":-62 }, { "mac":"01:17:C5:97:E7:B3", "rssi":-85 }, { "mac":"01:17:C5:97:1B:44", "rssi":-78 }, { "mac":"01:17:C5:97:B5:70", "rssi":-82 }, { "mac":"01:17:C5:97:87:84", "rssi":-84 }, { "mac":"01:17:C5:97:58:C3", "rssi":-88 } ] }`
	testdb := gettestdbName()
	defer freedb(testdb)

	router := gin.New()
	router.GET("/foo", GetUserLocations)

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/foo?group="+testdb+"&user=hadi", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
}
