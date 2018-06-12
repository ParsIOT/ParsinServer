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
	Socket            string
	Cwd               string
	MqttServer        string
	MqttAdmin         string
	MosquittoPID      string
	MqttAdminPassword string
	Dump              string
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

type UserPositionJSON struct {
	Time        int64              `json:"time"`
	Location    string             `json:"Location"`
	BayesGuess  string             `json:"bayesguess"`
	BayesData   map[string]float64 `json:"bayesdata"`
	SvmGuess    string             `json:"svmguess"`
	SvmData     map[string]float64 `json:"svmdata"`
	ScikitData  map[string]string  `json:"rfdata"`
	KnnGuess    string             `json:"knnguess"`
	KnnData     map[string]float64 `json:"knndata"`
	PDRLocation string             `json:"pdrlocation"`
}

//	filterMacs is used for set filtermacs
type FilterMacs struct {
	Group string   `json:"group"`
	Macs  []string `json:"macs"`
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
	//RuntimeArgs.SourcePath
	RuntimeArgs.Message = ""
	//RuntimeArgs.FilterMacsMap = make(map[string][]string)
	//RuntimeArgs.NeedToFilter = make(map[string]bool)
	//RuntimeArgs.NotNullFilterList = make(map[string]bool)
}


