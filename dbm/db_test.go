package dbm

import (
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"ParsinServer/glb"
	"sync"
	"strconv"
	"os/exec"
	"log"
	"github.com/gin-gonic/gin"
	"reflect"
	"strings"
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
	//initializeSharedPreferences(testdbName)
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

// This function use user.db database
//func TestGetAdminUsers(t *testing.T) {
//	//adminList, err := GetAdminUsers("test")
//
//	adminList, err := GetAdminUsers()
//	if err != nil {
//		t.Errorf("Can't Unmarshal admin list")
//		return
//	}
//	adminListJson, err := json.Marshal(adminList)
//	if err != nil {
//		t.Errorf("Can't remarshal admin list!")
//	}
//	response := `{"admin":"admin"}`
//	assert.Equal(t, adminListJson, response)
//}
//
//func TestAddAdminUser(t *testing.T){
//	// copy current user.db to user.db.backup.temp
//	lock.Lock()
//	_, err := exec.Command("cp", []string{path.Join(DataPath, "users.db"),path.Join(DataPath, "users.db.backup.temp")}...).Output()
//	if err != nil {
//		t.Errorf("Can't copy user.db to user.db.backup.temp")
//		log.Fatal(err)
//	}
//	lock.Unlock()
//
//	// set user.db.backup as user.db
//	lock.Lock()
//	_, err = exec.Command("cp", []string{path.Join(DataPath, "users.db.backup"),path.Join(DataPath, "users.db")}...).Output()
//	if err != nil {
//		t.Errorf("Can't copy user.db.backup to user.db")
//		log.Fatal(err)
//	}
//	lock.Unlock()
//
//	glb.Debug.Println("user db copied")
//
//	addAdminErr := AddAdminUser("test","test")
//	glb.Debug.Println("AddAdminUser executed ")
//
//	if addAdminErr != nil{
//		t.Errorf(addAdminErr.Error())
//	}else{
//		adminList, err := GetAdminUsers()
//		if err != nil {
//			t.Errorf("Can't Unmarshal admin list")
//		}else{
//			glb.Debug.Println(adminList)
//			adminListJson, err := json.Marshal(adminList)
//			if err != nil {
//				t.Errorf("Can't remarshal admin list!")
//			}
//			response := `{"admin":"admin","test":"test"}`
//			assert.Equal(t, adminListJson, response)
//		}
//	}
//
//	// recover user.db from user.db.backup.temp
//	lock.Lock()
//	_, err = exec.Command("mv", []string{path.Join(DataPath, "users.db.backup.temp"),path.Join(DataPath, "users.db")}...).Output()
//	if err != nil {
//		t.Errorf("Can't copy user.db.backup.temp to user.db")
//		log.Fatal(err)
//	}
//	lock.Unlock()
//}

func TestBackup(t *testing.T) {
	assert.Equal(t, DumpFingerprints("testdb"), nil)
}

func TestGroupExists(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	isExist := GroupExists(testdb)
	assert.Equal(t, isExist, true)

}

func TestGetUsers(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	users := GetUsers(testdb)
	assert.Equal(t, users, []string{"user","test","hadi"})
}

func TestGetUniqueMacs(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	//var UniqueMacs []string
	//defer glb.TimeTrack(time.Now(), "getUniqueMacs1")
	//if true{
	//	defer glb.TimeTrack(time.Now(), "getUniqueMacs2")
	//	UniqueMacs = GetUniqueMacs(testdb)
	//}

	uniqueMacs := GetUniqueMacs(testdb)
	//responseList := []string{"6c:19:8f:50:c6:a5", "b4:52:7d:26:e3:f3", "6c:3b:6b:09:da:6f", "9c:d6:43:72:0e:83", "02:1a:11:f5:6c:03", "34:97:f6:63:bd:94", "00:23:f8:91:be:43", "58:6d:8f:2b:26:42", "c4:6e:1f:d7:2e:de", "98:42:46:00:99:eb", "c4:12:f5:01:89:70", "4c:5e:0c:ec:85:85", "4c:5e:0c:40:1c:77", "b4:75:0e:e1:39:1a", "50:67:f0:7b:02:c7", "00:23:f8:91:c5:27", "6c:fd:b9:8c:fa:9b", "e4:8d:8c:15:e1:6f", "c4:e9:84:98:cb:ed", "40:4a:03:ad:17:ae"}
	responseList := []string{"b4:52:7d:26:e3:f3","50:67:f0:7b:02:c7","6c:19:8f:50:c6:a5","9c:d6:43:72:0e:83","6c:3b:6b:09:da:6f","4c:5e:0c:ec:85:85","02:1a:11:f5:6c:03","34:97:f6:63:bd:94","58:6d:8f:2b:26:42","b4:75:0e:e1:39:1a","98:42:46:00:99:eb"}

	listEqual := true
	itemFound := false
	for _,item1 := range responseList{
		for _,item2 := range uniqueMacs{
			if (item1 == item2){
				itemFound = true
				break
			}
		}
		if (!itemFound){
			listEqual = false
			break
		}
		itemFound = false
	}

	if (!listEqual){
		glb.Debug.Println("UniqueMacs: ")
		glb.Debug.Println(uniqueMacs)
		glb.Debug.Println("responseList: ")
		glb.Debug.Println(responseList)
	}
	assert.Equal(t, listEqual, true)
}

func TestGetUniqueLocations(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	uniqueLocs := GetUniqueLocations(testdb)
	responseList := []string{"100,100", "300,300"}

	listEqual := true
	itemFound := false
	for _,item1 := range responseList{
		for _,item2 := range uniqueLocs{
			if (item1 == item2){
				itemFound = true
				break
			}
		}
		if (!itemFound){
			listEqual = false
			break
		}
		itemFound = false
	}

	if (!listEqual){
		glb.Debug.Println("UniqueLocs: ")
		glb.Debug.Println(uniqueLocs)
		glb.Debug.Println("responseList: ")
		glb.Debug.Println(responseList)
	}
	assert.Equal(t, listEqual, true)
}

func TestGetMacCount(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	macs := GetMacCount(testdb)
	//jsonTest := "{\"4c:5e:0c:40:1c:77\":3,\"e4:8d:8c:15:e1:6f\":2,\"c4:e9:84:98:cb:ed\":3,\"b4:75:0e:e1:39:1a\":16,\"00:23:f8:91:be:43\":14,\"6c:19:8f:50:c6:a5\":100,\"b4:52:7d:26:e3:f3\":99,\"34:97:f6:63:bd:94\":90,\"02:1a:11:f5:6c:03\":100,\"c4:12:f5:01:89:70\":50,\"6c:fd:b9:8c:fa:9b\":2,\"50:67:f0:7b:02:c7\":8,\"40:4a:03:ad:17:ae\":1,\"00:23:f8:91:c5:27\":1,\"6c:3b:6b:09:da:6f\":100,\"98:42:46:00:99:eb\":97,\"4c:5e:0c:ec:85:85\":78,\"c4:6e:1f:d7:2e:de\":13,\"58:6d:8f:2b:26:42\":100,\"9c:d6:43:72:0e:83\":65}"
	jsonTest :="{\"b4:75:0e:e1:39:1a\":1,\"6c:19:8f:50:c6:a5\":1,\"9c:d6:43:72:0e:83\":1,\"58:6d:8f:2b:26:42\":1,\"4c:5e:0c:ec:85:85\":1,\"02:1a:11:f5:6c:03\":1,\"34:97:f6:63:bd:94\":1,\"98:42:46:00:99:eb\":1,\"b4:52:7d:26:e3:f3\":99,\"50:67:f0:7b:02:c7\":8,\"6c:3b:6b:09:da:6f\":1}"
	tempMac := make(map[string]int)
	json.Unmarshal([]byte(jsonTest), &tempMac)
	//glb.Debug.Println(tempMac)
	//glb.Debug.Println(macs)
	isEqual := glb.MapLike(macs,tempMac)
	assert.Equal(t, isEqual, true)
}

func TestRenameNetwork(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	err := RenameNetwork(testdb,"0","1")
	assert.Equal(t, err, nil)
}

func TestSetFilterMacDB(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	filterMacs := []string{"6c:19:8f:50:c6:a5","4c:5e:0c:40:1c:77"}
	err := SetSharedPrf(testdb,"FilterMacsMap", filterMacs)
	assert.Equal(t, err, nil)
}

func TestGetFilterMacDB(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)


	filterMacs := []string{"6c:19:8f:50:c6:a5","4c:5e:0c:40:1c:77"}
	//err := SetFilterMacDB(testdb,  filterMacs)

	err := SetSharedPrf(testdb,"FilterMacsMap", filterMacs)
	assert.Equal(t, err, nil)

	//err, filterMacs := GetFilterMacDB(testdb)
	filterMacs = GetSharedPrf(testdb).FilterMacsMap

	//if err != nil {
	//	t.Errorf("Can't get macs list")
	//}
	filterMacsRes := []string{"6c:19:8f:50:c6:a5","4c:5e:0c:40:1c:77"}
	assert.Equal(t, filterMacs, filterMacsRes)
}
//
//func TestGetMixinOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	mixinOverride, err := GetMixinOverride(testdb)
//	if err != nil{
//		t.Errorf("Can't get from db")
//	}else{
//		assert.Equal(t, mixinOverride, glb.DefaultMixin)
//	}
//}
//
//func TestGetCutoffOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	mixinOverride, err := GetCutoffOverride(testdb)
//	if err != nil{
//		t.Errorf("Can't get from db")
//	}else{
//		assert.Equal(t, mixinOverride, glb.DefaultCutoff)
//	}
//}
//
//func TestGetKnnKOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	mixinOverride, err := GetKnnKOverride(testdb)
//	if err != nil{
//		t.Errorf("Can't get from db")
//	}else{
//		assert.Equal(t, mixinOverride, glb.DefaultKnnK)
//	}
//}
//
//func TestGetMinRSSOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	mixinOverride, err := GetMinRSSOverride(testdb)
//	if err != nil{
//		t.Errorf("Can't get from db")
//	}else{
//		assert.Equal(t, mixinOverride, glb.MinRssi)
//	}
//}
//
//func TestSetMixinOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	err := SetMixinOverride(testdb, glb.DefaultMixin/100)
//	if err != nil{
//		t.Errorf("Can't set to db")
//	}else{
//		mixinOverride, err := GetMixinOverride(testdb)
//		if err != nil{
//			t.Errorf("Can't get from db")
//		}else{
//			assert.Equal(t, mixinOverride, glb.DefaultMixin/100)
//		}
//	}
//}
//
//func TestSetCutoffOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	err := SetCutoffOverride(testdb, glb.DefaultCutoff/100)
//	if err != nil{
//		t.Errorf("Can't set to db")
//	}else{
//		mixinOverride, err := GetCutoffOverride(testdb)
//		if err != nil{
//			t.Errorf("Can't get from db")
//		}else{
//			assert.Equal(t, mixinOverride, glb.DefaultCutoff/100)
//		}
//	}
//}
//
//func TestSetKnnKOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	err := SetKnnKOverride(testdb, glb.DefaultKnnK*2)
//	if err != nil{
//		t.Errorf("Can't set to db")
//	}else{
//		mixinOverride, err := GetKnnKOverride(testdb)
//		if err != nil{
//			t.Errorf("Can't get from db")
//		}else{
//			assert.Equal(t, mixinOverride, glb.DefaultKnnK*2)
//		}
//	}
//}
//
//func TestSetMinRSSOverride(t *testing.T) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	err := SetMinRSSOverride(testdb, glb.MinRssi-20)
//	if err != nil{
//		t.Errorf("Can't set to db")
//	}else{
//		mixinOverride, err := GetMinRSSOverride(testdb)
//		if err != nil{
//			t.Errorf("Can't get from db")
//		}else{
//			assert.Equal(t, mixinOverride, glb.MinRssi-20)
//		}
//	}
//}

func TestGetLearnFingerPrints(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	fingerprintsOrdering1,fingerprintsInMemory1,err1 := GetLearnFingerPrints(testdb,false)
	fingerprintsOrdering2,fingerprintsInMemory2,err2 := GetLearnFingerPrints(testdb,true)

	//glb.Debug.Println(fingerprintsInMemory1)
	//
	//glb.Debug.Println(fingerprintsInMemory2)

	glb.Debug.Println(len(fingerprintsInMemory1[fingerprintsOrdering1[10]].WifiFingerprint))
	glb.Debug.Println(len(fingerprintsInMemory2[fingerprintsOrdering2[10]].WifiFingerprint))
	if err1!=nil || err2!=nil{
		t.Errorf(err1.Error())
		t.Errorf(err2.Error())
	}else{
		result := []int{len(fingerprintsOrdering1),len(fingerprintsInMemory1),len(fingerprintsOrdering2),len(fingerprintsInMemory2)}
		assert.Equal(t, result, []int{100,100,100,100})
	}

}

func TestLoadSharedPreferences(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	shrPrf,err := loadSharedPreferences(testdb)
	assert.Equal(t, err, nil)


	shrPrfRes := RawSharedPreferences{
		Mixin:     			float64(0.1),
		Cutoff:    			float64(0.01),
		MinRss:    			int(-110),
		MinRssOpt: 			int(-100),
		FilterMacsMap: 		[]string{"b4:52:7d:26:e3:f3","50:67:f0:7b:02:c7"},
	}
	//glb.Debug.Println(shrPrf)
	//glb.Debug.Println(shrPrfRes)
	assert.Equal(t, shrPrf, shrPrfRes)
}

func TestPutSharedPreferences(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	shrPrf := RawSharedPreferences{
		Mixin:     			float64(0.01),
		Cutoff:    			float64(0.001),
		MinRss:    			int(-130),
		MinRssOpt: 			int(-110),
		FilterMacsMap: 		[]string{"b4:52:7d:26:e2:c3","50:67:f0:7c:01:a1"},
	}
	err := putSharedPreferences(testdb, shrPrf)
	assert.Equal(t, err, nil)

	shrPrfRes,err := loadSharedPreferences(testdb)
	assert.Equal(t, err, nil)

	//glb.Debug.Println(shrPrf)
	//glb.Debug.Println(shrPrfRes)
	assert.Equal(t, shrPrfRes, shrPrf)
}

func TestInitializeSharedPreferences(t *testing.T){
	testdb := gettestdbName()
	defer freedb(testdb)

	initializeSharedPreferences(testdb)

	shrPrf,err := loadSharedPreferences(testdb)
	assert.Equal(t, err, nil)

	shrPrfRes := NewRawSharedPreferences()
	//glb.Debug.Println(shrPrf)
	//glb.Debug.Println(shrPrfRes)
	assert.Equal(t, shrPrf, shrPrfRes)
}

func BenchmarkGetUniqueLocations(b *testing.B){
	testdb := gettestdbName()
	defer freedb(testdb)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uniqueLocs := GetUniqueLocations(testdb)
		glb.Debug.Println(uniqueLocs)
	}
}

//func BenchmarkTrackFingerprintPOST(b *testing.B){
//	//testdb := gettestdbName()
//	//defer freedb(testdb)
//
//
//	router := gin.New()
//	router.POST("/foo", algorithms.TrackFingerprintPOST)
//	jsonStr := []byte("{\"group\":\""+testdb+"\",\"macs\":[\"6c:19:8f:50:c6:a5\",\"b4:52:7d:26:e3:f3\"]}")
//	req, _ := http.NewRequest("POST", "/foo", bytes.NewBuffer(jsonStr))
//	resp := httptest.NewRecorder()
//	router.ServeHTTP(resp, req)
//	response := "{\"message\":\"MacFilter set successfully\",\"success\":true}"
//
//	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
//}
//
//
//router := gin.New()
//router.PUT("/foo", PutMixinOverride)
//
//req, _ := http.NewRequest("PUT", "/foo?group="+testdb+"&mixin=100", nil)
//resp := httptest.NewRecorder()
//router.ServeHTTP(resp, req)
//response := `{"message":"mixin must be between 0 and 1","success":false}`
//assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)

//func BenchmarkPutFingerprintInDatabase(b *testing.B) {
//	jsonTest := `{"username": "zack", "group": "testdbfoo", "wifi-fingerprint": [{"rssi": -45, "mac": "80:37:73:ba:f7:d8"}, {"rssi": -58, "mac": "80:37:73:ba:f7:dc"}, {"rssi": -61, "mac": "a0:63:91:2b:9e:65"}, {"rssi": -68, "mac": "a0:63:91:2b:9e:64"}, {"rssi": -70, "mac": "70:73:cb:bd:9f:b5"}, {"rssi": -75, "mac": "d4:05:98:57:b3:10"}, {"rssi": -75, "mac": "00:23:69:d4:47:9f"}, {"rssi": -76, "mac": "30:46:9a:a0:28:c4"}, {"rssi": -81, "mac": "2c:b0:5d:36:e3:b8"}, {"rssi": -82, "mac": "00:1a:1e:46:cd:10"}, {"rssi": -82, "mac": "20:aa:4b:b8:31:c8"}, {"rssi": -83, "mac": "e8:ed:05:55:21:10"}, {"rssi": -83, "mac": "ec:1a:59:4a:9c:ed"}, {"rssi": -88, "mac": "b8:3e:59:78:35:99"}, {"rssi": -84, "mac": "e0:46:9a:6d:02:ea"}, {"rssi": -84, "mac": "00:1a:1e:46:cd:11"}, {"rssi": -84, "mac": "f8:35:dd:0a:da:be"}, {"rssi": -84, "mac": "b4:75:0e:03:cd:69"}], "location": "zakhome floor 2 office", "time": 1439596533831, "password": "frusciante_0128"}`
//	res := parameters.Fingerprint{}
//	json.Unmarshal([]byte(jsonTest), &res)
//	os.Remove(path.Join(glb.RuntimeArgs.SourcePath, "testdbfoo.db"))
//	PutFingerprintIntoDatabase(res, "fingerprints") // create it for the first time
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		PutFingerprintIntoDatabase(res, "fingerprints")
//	}
//}
//
//func BenchmarkGetFingerprintInDatabase(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		group := "testdb"
//		db, _ := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
//		db.View(func(tx *bolt.Tx) error {
//			// Assume bucket exists and has keys
//			b := tx.Bucket([]byte("fingerprints"))
//			c := b.Cursor()
//			for k, v := c.First(); k != nil; k, v = c.Next() {
//				LoadFingerprint(v,true)
//				break
//			}
//			return nil
//		})
//		db.Close()
//	}
//}
//

