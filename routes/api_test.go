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
	"reflect"
	"ParsinServer/glb"
)

type Empty struct{}

var DataPath string

func init() {
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
	fmt.Println(DataPath)
	_, err := exec.Command("cp", []string{path.Join(DataPath, "testdb.db.backup"),path.Join(DataPath, "testdb.db")}...).Output()
	if err != nil {
		log.Fatal(err)
	}
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
	router := gin.New()
	router.PUT("/foo", MigrateDatabase)

	req, _ := http.NewRequest("PUT", "/foo?from=testdb&to=newdb", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), "{\"message\":\"Successfully migrated testdb to newdb\",\"success\":true}")
	fmt.Println(DataPath)
	os.Remove(path.Join(DataPath,"newdb.db"))
}

func TestDeleteDatabase(t *testing.T) {
	glb.CopyFile(path.Join(DataPath,"testdb.db"), path.Join(DataPath,"deleteme.db"))

	router := gin.New()
	router.DELETE("/foo", DeleteDatabase)

	req, _ := http.NewRequest("DELETE", "/foo?group=deleteme", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), "{\"message\":\"Successfully deleted deleteme\",\"success\":true}")
}

func TestCalculate(t *testing.T) {
	router := gin.New()
	router.GET("/foo", Calculate)

	req, _ := http.NewRequest("GET", "/foo?group=testdb", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, strings.TrimSpace(resp.Body.String()), "{\"message\":\"Parameters optimized.\",\"success\":true}")
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
	router := gin.New()
	router.GET("/foo", GetUserLocations)

	req, _ := http.NewRequest("GET", "/foo?group=testdb&user=hadi", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, strings.Contains(resp.Body.String(), "{\"message\":\"Correctly found locations.\""), true)
}

func TestPutMixinOverrideBad(t *testing.T) {
	router := gin.New()
	router.PUT("/foo", PutMixinOverride)

	req, _ := http.NewRequest("PUT", "/foo?group=testdb&mixin=100", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := `{"message":"mixin must be between 0 and 1","success":false}`
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestPutMixinOverrideGood(t *testing.T) {
	router := gin.New()
	router.PUT("/foo", PutMixinOverride)

	req, _ := http.NewRequest("PUT", "/foo?group=testdb&mixin=0", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := `{"message":"Overriding mixin for testdb, now set to 0","success":true}`
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestEditNetworkName(t *testing.T) {
	router := gin.New()
	router.GET("/foo", EditNetworkName)
	req, _ := http.NewRequest("GET", "/foo?group=testdb&oldname=0&newname=home", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Finished\",\"success\":true}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestEditName(t *testing.T) {
	router := gin.New()
	router.GET("/foo", EditName)

	req, _ := http.NewRequest("GET", "/foo?group=testdb&location=t1&newname=t2", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Changed name of 50 things\",\"success\":true}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestEditUserName(t *testing.T) {
	router := gin.New()
	router.GET("/foo", EditUserName)

	req, _ := http.NewRequest("GET", "/foo?group=testdb&user=hadi1&newname=hadi1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Changed name of 6 things\",\"success\":true}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}

func TestDeleteUser(t *testing.T) {
	router := gin.New()
	router.DELETE("/foo", DeleteUser)
	req, _ := http.NewRequest("DELETE", "/foo?group=testdb&user=hadi", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	response := "{\"message\":\"Deletes 6 things  with user hadi\",\"success\":true}"
	assert.Equal(t, strings.TrimSpace(resp.Body.String()), response)
}
