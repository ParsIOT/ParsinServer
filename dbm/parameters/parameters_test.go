package parameters

import (
	"ParsinServer/glb"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
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

// It's using bolt to check LoadParameters
//func BenchmarkLoadParameters(b *testing.B) {
//	testdb := gettestdbName()
//	defer freedb(testdb)
//
//	var ps FullParameters = *NewFullParameters()
//	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, testdb+".db"), 0600, nil)
//	defer db.Close()
//	if err != nil {
//		glb.Error.Println(err)
//	}
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		err = db.View(func(tx *bolt.Tx) error {
//			// Assume bucket exists and has keys
//			b := tx.Bucket([]byte("resources"))
//			if b == nil {
//				glb.Error.Println("Resources dont exist")
//				return fmt.Errorf("")
//			}
//			v := b.Get([]byte("fullParameters"))
//			ps = LoadParameters(v)
//			return nil
//		})
//		if err != nil {
//			glb.Error.Println(err)
//		}
//
//	}
//}

func TestConvertSharpToSemiColonInFP(t *testing.T) {
	fps := []Router{
		Router{
			Mac:  "WIFI#b4:52:7d:26:e3:f3",
			Rssi: -45,
		},
		Router{
			Mac:  "BLE#14:51:7E:22:A1:E4",
			Rssi: -50,
		},
	}

	fpsRes := []Router{
		Router{
			Mac:  "WIFI;b4:52:7d:26:e3:f3",
			Rssi: -45,
		},
		Router{
			Mac:  "BLE;14:51:7E:22:A1:E4",
			Rssi: -50,
		},
	}

	res := ConvertSharpToUnderlineInFP(fps)
	glb.Debug.Println(res)
	assert.Equal(t, fpsRes, res)
}
