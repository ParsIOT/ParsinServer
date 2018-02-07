package glb

import (
	"github.com/MA-Heshmatkhah/SimpleAuth"
	"os"
	"path"
	"strings"
	"reflect"
)

var SessionManager SimpleAuth.Manager


// RuntimeArgs contains all runtime
// arguments available
// Todo: May be shared value, so must use lock to edit it.
var RuntimeArgs struct {
	ScikitPort            string
					  //FilterMacFile     string
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
	Scikit	          bool
	NeedToFilter      map[string]bool //check needing for filtering
	NotNullFilterMap  map[string]bool //check that filterMap is null(used to avoid filter fingerprint with null map)
	FilterMacsMap     map[string][]string
	AdminAdd          string
	GaussianDist      bool
	MinRssOpt         int
	KNN               bool
}

type UserPositionJSON struct {
	Time       interface{}        `json:"time"`
	BayesGuess interface{}        `json:"bayesguess"`
	BayesData  map[string]float64 `json:"bayesdata"`
	SvmGuess   interface{}        `json:"svmguess"`
	SvmData    map[string]float64 `json:"svmdata"`
	ScikitData     map[string]string `json:"rfdata"`
	KnnGuess   interface{}        `json:"knnguess"`
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
}