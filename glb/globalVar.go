package glb

import (
	"github.com/MA-Heshmatkhah/SimpleAuth"
	"os"
	"path"
	"strings"
	"reflect"
)

var SessionManager SimpleAuth.Manager

// There are 3 type of global var :
//	1.Constants: in globalVar.go and algoVar.go in glb package
//  2.Share variables(in db or runtime): in cache.go
//  3.Runtime arguments: in RuntimeArgs struct in globalVar.go

// RuntimeArgs contains all runtime
// arguments available
// Todo: Just add runtime variable here. Add shared variable in cache.go. Add constant in another fields and in algoVar.go
var RuntimeArgs struct {
	ScikitPort        string
	ExternalIP        string
	Port              string
	ServerCRT         string
	ServerKey         string
	SourcePath        string
	MapPath			  string
	MapDirectory	  string
	Socket            string
	Cwd               string
	MqttServer        string
	MqttAdmin         string
	MosquittoPID      string
	MqttAdminPassword string
	Dump              string
	DumpRaw           string
	DumpCalc          string
	Message           string
	Mqtt              bool
	MqttExisting      bool
	Svm               bool
	Scikit            bool
	//NeedToFilter      map[string]bool //check needing for filtering
	//NotNullFilterList map[string]bool //check that filterMap is null(used to avoid filter fingerprint with null map)
	//FilterMacsMap     map[string][]string
	AdminAdd          string
	GaussianDist      bool
	MinRssOpt         int
	KNN               bool
}


type Empty struct{}

func init(){
	cwd, _ := os.Getwd()
	pkgName := reflect.TypeOf(Empty{}).PkgPath()
	
	projName := strings.Split(pkgName,"/")[0]
	for _,p := range strings.Split(cwd,"/") {
		if p == projName {
			RuntimeArgs.Cwd += p +"/"
			break
		}
		RuntimeArgs.Cwd += p +"/"
	}
	RuntimeArgs.SourcePath = path.Join(RuntimeArgs.Cwd, "data")
	//MapPath := filepath.FromSlash("res/static/map")
	MapPath := "static/map"
	FullMapDirectory := "res/static/map"
	//tempMapPath := filepath.FromSlash("res/static/map/")
	//fmt.Println("map path from slash: ",tempMapPath)
	RuntimeArgs.MapDirectory = path.Join(RuntimeArgs.Cwd, FullMapDirectory)
	RuntimeArgs.MapPath = path.Join("/", MapPath)
	//RuntimeArgs.MapPath = path.Join(RuntimeArgs.Cwd, MapPath)
	//RuntimeArgs.SourcePath
	RuntimeArgs.Message = ""
	//RuntimeArgs.FilterMacsMap = make(map[string][]string)
	//RuntimeArgs.NeedToFilter = make(map[string]bool)
	//RuntimeArgs.NotNullFilterList = make(map[string]bool)
}


