package parameters

import (
	"path"
	"testing"
	"github.com/boltdb/bolt"
	"ParsinServer/glb"
	"fmt"
	"strconv"
	"os"
	"os/exec"
	"log"
	"github.com/gin-gonic/gin"
	"reflect"
	"strings"
	"sync"
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
func BenchmarkLoadParameters(b *testing.B) {
	testdb := gettestdbName()
	defer freedb(testdb)

	var ps FullParameters = *NewFullParameters()
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, testdb+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte("resources"))
			if b == nil {
				glb.Error.Println("Resources dont exist")
				return fmt.Errorf("")
			}
			v := b.Get([]byte("fullParameters"))
			ps = LoadParameters(v)
			return nil
		})
		if err != nil {
			glb.Error.Println(err)
		}

	}
}

