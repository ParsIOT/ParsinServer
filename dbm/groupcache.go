package dbm

import (
	"fmt"
	"sync"
	"time"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"reflect"
	"errors"
)

//Todo: After each update in groupcache.go, rebuild the group (use /buildgroup)
//Todo: After any change in structs rerun "easyjson -all groupcache.go" in dbm directory
var GM GroupManger



var Wg sync.WaitGroup
var FlushDelay time.Duration = 20


type RawDataStruct struct{
	mutex *sync.RWMutex
	group *Group
	//Learned data:
	Fingerprints			map[string]parameters.Fingerprint
	FingerprintsOrdering 	[]string

	LearnTrueLocations     map[int64]string
	TestValidTrueLocations map[int64]string //timestamp:location
	//Note: Run easyjson.sh after editing
}

func (st *RawDataStruct) Lock() {
	//glb.Debug.Println("RawDataStruct Lock")
	st.mutex.Lock()
}
func (st *RawDataStruct) Unlock() {
	//glb.Debug.Println("RawDataStruct UnLock")
	st.mutex.Unlock()
}
func (st *RawDataStruct) RLock() {
	//glb.Debug.Println("RawDataStruct RLock")
	st.mutex.RLock()
}
func (st *RawDataStruct) RUnlock() {
	//glb.Debug.Println("RawDataStruct RUnLock")
	st.mutex.RUnlock()
}

type ConfigDataStruct struct {
	mutex *sync.RWMutex
	group *Group
	//Learned data:
	Test int
	GroupGraph  			parameters.Graph

	//Note: Run easyjson.sh after editing
}

func (st *ConfigDataStruct) Lock() {
	//glb.Debug.Println("RawDataStruct Lock")
	st.mutex.Lock()
}
func (st *ConfigDataStruct) Unlock() {
	//glb.Debug.Println("RawDataStruct UnLock")
	st.mutex.Unlock()
}
func (st *ConfigDataStruct) RLock() {
	//glb.Debug.Println("RawDataStruct RLock")
	st.mutex.RLock()
}
func (st *ConfigDataStruct) RUnlock() {
	//glb.Debug.Println("RawDataStruct RUnLock")
	st.mutex.RUnlock()
}

type MiddleDataStruct struct{
	mutex *sync.RWMutex
	group *Group
	//Midlle data:
	NetworkMacs    			map[string]map[string]bool             // map of networks and then the associated macs in each
	NetworkLocs    			map[string]map[string]bool             // map of the networks, and then the associated locations in each
	MacVariability 			map[string]float32                     // variability of macs
	MacCount      			map[string]int                          // number of fingerprints of an AP in all data, regardless of the location; e.g. 10 of AP1, 12 of AP2, ...
	MacCountByLoc 			map[string]map[string]int               // number of fingerprints of an AP in a location; e.g. in location A, 10 of AP1, 12 of AP2, ...
	UniqueLocs    			[]string                                // a list of all unique locations e.g. {P1,P2,P3}
	UniqueMacs    			[]string                                // a list of all unique APs
	LocCount				map[string]int							// number of fp that its Location equals to loc
	//Note: Run easyjson.sh after editing
}

func (st *MiddleDataStruct) Lock() {
	//glb.Debug.Println("MiddleDataStruct Lock")
	st.mutex.Lock()
}
func (st *MiddleDataStruct) Unlock() {
	//glb.Debug.Println("MiddleDataStruct UnLock")
	st.mutex.Unlock()
}
func (st *MiddleDataStruct) RLock() {
	//glb.Debug.Println("MiddleDataStruct RLock")
	st.mutex.RLock()
}
func (st *MiddleDataStruct) RUnlock() {
	//glb.Debug.Println("MiddleDataStruct RUnLock")
	st.mutex.RUnlock()
}

// Assume learned model not to be changed or improved (if there is one algorithm that need it, add new struct near rawdata,middledata and ...)
// this is because all AlgoDataStruct rewrite completely to db if it chanfluges.


type AlgoDataStruct struct{
	mutex *sync.RWMutex
	group *Group
	////Algorithm Data:
	//BayesPriors   			map[string]parameters.PriorParameters   // generate BayesPriors for each network
	//BayesResults  			map[string]parameters.ResultsParameters // generate BayesResults for each network
	KnnFPs        			parameters.KnnFingerprints
	//GroupGraph  			parameters.Graph
	//Note: Run easyjson.sh after editing
}

func (st *AlgoDataStruct) Lock() {
	//glb.Debug.Println("AlgoDataStruct Lock")
	st.mutex.Lock()
}
func (st *AlgoDataStruct) Unlock() {
	//glb.Debug.Println("AlgoDataStruct ULock")
	st.mutex.Unlock()
}
func (st *AlgoDataStruct) RLock() {
	//glb.Debug.Println("AlgoDataStruct RLock")
	st.mutex.RLock()
}
func (st *AlgoDataStruct) RUnlock() {
	//glb.Debug.Println("AlgoDataStruct RUnLock")
	st.mutex.RUnlock()
}



type ResultDataStruct struct{
	mutex           *sync.RWMutex
	group           *Group
	//Results         map[string]parameters.Fingerprint
	AlgoAccuracy    map[string]int            // crossValidation results
	AlgoAccuracyLoc map[string]map[string]int // from crossValidation results

	UserHistory     map[string][]parameters.UserPositionJSON // it's temprary
	UserResults     map[string][]parameters.UserPositionJSON // it saves in db

	// TestValid FPs field:
	AlgoTestErrorAccuracy map[string]int              // algorithmName --> error
	TestValidTracks       []parameters.TestValidTrack // keeps testValid userPosition and the true location
	//Note: Run easyjson.sh after editing
}

func (st *ResultDataStruct) Lock() {
	//glb.Debug.Println("ResultDataStruct Lock")
	st.mutex.Lock()
}
func (st *ResultDataStruct) Unlock() {
	//glb.Debug.Println("ResultDataStruct UnLock")
	st.mutex.Unlock()
}
func (st *ResultDataStruct) RLock() {
	//glb.Debug.Println("ResultDataStruct RLock")
	st.mutex.RLock()
}
func (st *ResultDataStruct) RUnlock() {
	//glb.Debug.Println("ResultDataStruct RUnLock")
	st.mutex.RUnlock()
}

//parameters Name must be lowercase that can't be access out of cachelib(must provide set&get func for each and provide locking mutex for each one)
type Group struct {
	mutex      *sync.RWMutex
	GMutex     *sync.RWMutex
	Name       string
	Permanent  bool // Some group doesn't need to be saved
	RawData    *RawDataStruct
	ConfigData *ConfigDataStruct
	MiddleData *MiddleDataStruct
	AlgoData   *AlgoDataStruct
	ResultData *ResultDataStruct

	RawDataChanged    bool
	ConfigDataChanged bool
	MiddleDataChanged bool
	AlgoDataChanged   bool
	ResultDataChanged bool
	//learnDB	   map[string]map[string]{}interface	// group-->algorithm-->learnedData
}

func (st *Group) Lock() {
	//glb.Debug.Println("Group Lock")
	st.mutex.Lock()
}
func (st *Group) Unlock() {
	//glb.Debug.Println("Group UnLock")
	st.mutex.Unlock()
}
func (st *Group) RLock() {
	//glb.Debug.Println("Group RLock")
	st.mutex.RLock()
}
func (st *Group) RUnlock() {
	//glb.Debug.Println("Group RUnLock")
	st.mutex.RUnlock()
}

func NewGroup(groupName string) *Group {
	gp := &Group{
		mutex:             &sync.RWMutex{},
		GMutex:            &sync.RWMutex{},
		Name:              groupName,
		Permanent:         true,
		RawDataChanged:    false,
		ConfigDataChanged: false,
		MiddleDataChanged: false,
		AlgoDataChanged:   false,
		ResultDataChanged: false,
	}
	gp.Lock()
	gp.RawData = gp.NewRawDataStruct()
	gp.ConfigData = gp.NewConfigDataStruct()
	gp.MiddleData = gp.NewMiddleDataStruct()
	gp.AlgoData = gp.NewAlgoDataStruct()
	gp.ResultData = gp.NewResultDataStruct()
	gp.Unlock()
	return gp
}



// parameters
//func  (gp *Group) GetParameters(){
//	d := gp.Get_d()
//	//some works
//	gp.Set_d(d+1)
//}
//func GetParameters1(gp *Group){
//	d := gp.Get_d()
//	//some works
//	gp.Set_d(d+1)
//}


//Access to db must be done over GM (for consistency issue)
type GroupManger struct {
	mutex    *sync.RWMutex
	isLoad   map[string]bool
	dbLock   map[string]*sync.RWMutex
	dirtyBit map[string]bool
	groups   map[string]*Group
}

func (st *GroupManger) Lock() {
	//glb.Debug.Println("GroupManger Lock")
	st.mutex.Lock()
}
func (st *GroupManger) Unlock() {
	//glb.Debug.Println("GroupManger UnLock")
	st.mutex.Unlock()
}
func (st *GroupManger) RLock() {
	//glb.Debug.Println("GroupManger RLock")
	st.mutex.RLock()
}
func (st *GroupManger) RUnlock() {
	//glb.Debug.Println("GroupManger RUnLock")
	st.mutex.RUnlock()
}

func init(){
	GM = GroupManger{
		mutex:    &sync.RWMutex{},
		isLoad:   make(map[string]bool),
		dbLock:   make(map[string]*sync.RWMutex),
		dirtyBit: make(map[string]bool),
		groups:   make(map[string]*Group),
	}

	//GM.NewGroup("t1")
	//GM.groups["t2"] = NewGroup("t1")
	//GM.isLoad["t2"] = true
	////wg.Add(1)
	//go GM.Flusher()
	////wg.Wait()

	// Must run on server.go
}

// Note: NewGroup created raw group that has new rawdata and configdata
func (gm *GroupManger) NewGroup(groupName string) *Group {
	GM.RLock()
	groups := GM.groups
	GM.RUnlock()
	for gpName,gp := range groups{
		if(groupName == gpName){
			glb.Error.Println("There is a group exists with same Name:" + groupName)
			return gp
		}
	}

	gp := NewGroup(groupName)
	GM.Lock()
	GM.groups[groupName] = gp
	GM.isLoad[groupName] = true
	GM.dbLock[groupName] = &sync.RWMutex{}
	GM.dirtyBit[groupName] = false
	GM.Unlock()
	return gp
}

func (gm *GroupManger) GetGroup(groupName string) *Group {
	gm.LoadGroup(groupName)
	gm.RLock()
	group := gm.groups[groupName]
	gm.RUnlock()
	return group
}


//func (gm *GroupManger) SetGroup(gp *Group) {
//	//gm.LoadGroup(groupName)
//	gm.Lock()
//	gm.groups[gp.Name] = gp
//	gm.Unlock()
//	GM.SetDirtyBit(gp.Get_Name())
//}

func (gm *GroupManger) LoadGroup(groupName string){
	gm.RLock()
	loaded := gm.isLoad[groupName]
	dblock := gm.dbLock[groupName]
	gm.RUnlock()
	if dblock == nil{
		dblock = &sync.RWMutex{}
		gm.Lock()
		gm.dbLock[groupName] = dblock
		gm.Unlock()
	}

	if loaded{
		return
	}else {
		gp := NewGroup(groupName)
		rawData := gp.NewRawDataStruct()
		configData := gp.NewConfigDataStruct()
		middleData := gp.NewMiddleDataStruct()
		algoData := gp.NewAlgoDataStruct()
		resultData := gp.NewResultDataStruct()

		//{
			//dblock.dbLock()
			//GM.DBlock(groupName)
			dblock.Lock()

			//db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
			//if err != nil {
			//	log.Fatal(err)
			//	gp = GM.NewGroup(groupName)
			//}
			//db.View(func(tx *bolt.Tx) error {
			//	b := tx.Bucket([]byte("resources"))
			//	v := b.Get([]byte("parameters"))
			//	gp.UnmarshalJSON(v)
			//	return nil
			//})
			//db.Close
		rawDataBytes, err1 := GetBytejsonResourceInBucket("rawData", "resources", groupName)
		configDataBytes, err2 := GetBytejsonResourceInBucket("configData", "resources", groupName)
		middleDataBytes, err3 := GetBytejsonResourceInBucket("middleData", "resources", groupName)
		algoDataBytes, err4 := GetBytejsonResourceInBucket("algoData", "resources", groupName)
		resultDataBytes, err5 := GetBytejsonResourceInBucket("resultData", "resources", groupName)
			dblock.Unlock()

		//}


		if err1 != nil {
			glb.Error.Println(err1.Error())
		}else{
			rawData.UnmarshalJSON(rawDataBytes)
		}

		if err2 != nil {
			glb.Error.Println(err2.Error())
		}else{
			configData.UnmarshalJSON(configDataBytes)
		}


		if err3 != nil {
			glb.Error.Println(err3.Error())
		}else{
			middleData.UnmarshalJSON(middleDataBytes)
		}


		if err4 != nil {
			glb.Error.Println(err4.Error())
		}else{
			algoData.UnmarshalJSON(algoDataBytes)
		}

		if err5 != nil {
			glb.Error.Println(err5.Error())
		} else {
			resultData.UnmarshalJSON(resultDataBytes)
		}
		//bytes,err1 := GetBytejsonResourceInBucket("parameters","resources",groupName)
		//if err1 != nil {
		//	//log.Fatal(err1)
		//	glb.Error.Println(err1.Error())
		//	gp = GM.NewGroup(groupName)
		//}else{
		//	//jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bytes,&gp)
		//	gp.UnmarshalJSON(bytes)
		//}
		//dblock.Unlock()
		//dblock.dbUnlock()
		//GM.DBUnlock(groupName)
		//glb.Debug.Println(err1!=nil)
		//glb.Debug.Println(err2!=nil)
		//glb.Debug.Println(err3!=nil)
		//glb.Debug.Println(err4!=nil)
		if err1 != nil && err2 != nil && err3 != nil && err4 != nil && err5 != nil {
			gp = GM.NewGroup(groupName)
			glb.Debug.Println("Raw group created")
			//glb.Debug.Println(gp)
		}else{
			//glb.Error.Println(err1)
			//glb.Error.Println(err2)
			//glb.Error.Println(err3)
			//glb.Error.Println(err4)
			gp.Lock()
			gp.RawData = rawData
			gp.RawData.mutex = &sync.RWMutex{}
			gp.ConfigData = configData
			gp.ConfigData.mutex = &sync.RWMutex{}
			gp.MiddleData = middleData
			gp.MiddleData.mutex = &sync.RWMutex{}
			gp.AlgoData = algoData
			gp.AlgoData.mutex = &sync.RWMutex{}
			gp.ResultData = resultData
			gp.ResultData.mutex = &sync.RWMutex{}
			//gp.ResultData.Results = make(map[string]parameters.Fingerprint)

			gp.Unlock()
		}

		//gp.GetParameters()
		gp.Set_Name(groupName) //Some times need to reset Name!
		gm.Lock()
		gm.isLoad[groupName] = true
		gm.groups[groupName] = gp
		gm.Unlock()
	}
}

func (gm *GroupManger) SetDirtyBit(groupName string){
	gm.Lock()
	gm.dirtyBit[groupName] = true
	gm.Unlock()
}


func (gm *GroupManger) FlushDB(groupName string, gp *Group){
	//go func(groupName string){
	gm.Lock()
	dirtyBit := gm.dirtyBit[groupName]
	gm.dirtyBit[groupName] = false
	loaded := gm.isLoad[groupName]
	dblock := gm.dbLock[groupName]
	gm.Unlock()

	if dblock == nil{
		dblock = &sync.RWMutex{}
		gm.Lock()
		gm.dbLock[groupName] = dblock
		gm.Unlock()
	}

	if(dirtyBit) {
		//glb.Debug.Println("Dirtybit is true")
		if !loaded {
			glb.Error.Println("DB isn't loaded!")
			return
		} else {
			glb.Debug.Println("Flushing :", groupName)
			//glb.Debug.Println("DB is loaded")
			//DBLock.Lock()
			dbData := make(map[string][]byte)

			gp.RLock()
			rdChanged := gp.RawDataChanged
			cdChanged := gp.ConfigDataChanged
			mdChanged := gp.MiddleDataChanged
			adChanged := gp.AlgoDataChanged
			rsChanged := gp.ResultDataChanged
			gp.RUnlock()

			if rdChanged {
				gp.RLock()
				rawData := gp.RawData
				gp.RUnlock()
				rawData.RLock()
				v, err := rawData.MarshalJSON()
				rawData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["rawData"] = v
				gp.Lock()
				gp.RawDataChanged = false
				gp.Unlock()
			}

			if cdChanged {
				gp.RLock()
				configData := gp.ConfigData
				gp.RUnlock()
				configData.RLock()
				v, err := configData.MarshalJSON()
				configData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["configData"] = v
				gp.Lock()
				gp.ConfigDataChanged = false
				gp.Unlock()
			}

			if mdChanged {
				gp.RLock()
				middleData := gp.MiddleData
				gp.RUnlock()
				middleData.RLock()
				v, err := middleData.MarshalJSON()
				middleData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["middleData"] = v

				gp.Lock()
				gp.MiddleDataChanged = false
				gp.Unlock()
			}

			if adChanged {
				gp.RLock()
				algoData := gp.AlgoData
				gp.RUnlock()
				algoData.RLock()
				v, err := algoData.MarshalJSON()
				algoData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["algoData"] = v

				gp.Lock()
				gp.AlgoDataChanged = false
				gp.Unlock()
			}

			resultDataList := make(map[string]parameters.Fingerprint)
			if rsChanged {
				gp.RLock()
				resultData := gp.ResultData
				gp.RUnlock()

				//resultData.RLock()
				//resultDataList = resultData.Results
				//resultData.RUnlock()
				//resultData.Lock()
				//resultData.Results = make(map[string]parameters.Fingerprint)  // delete trackresults data
				////resultData.UserHistory = make(map[string][]string)
				//resultData.Unlock()
				//glb.Error.Println(resultData)
				resultData.Lock()

				tempUserHistory := resultData.UserHistory
				resultData.UserHistory = make(map[string][]parameters.UserPositionJSON)
				v, err := resultData.MarshalJSON()
				resultData.UserHistory = tempUserHistory

				resultData.Unlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}

				dbData["resultData"] = v
				dbData["Results"] = []byte{}

				gp.Lock()
				gp.ResultDataChanged = false
				gp.Unlock()


			}


			{
				dblock.Lock()
				defer dblock.Unlock()

				for key,val := range dbData{
					if(key == "Results"){
						for timeStamp,fp := range resultDataList{ // must put the list to db instantly
							err1 := SetByteResourceInBucket(parameters.DumpFingerprint(fp),timeStamp,"Results",groupName)
							if err1 != nil{
								glb.Error.Println(err1.Error())
							}
						}
					}else{
						err1 := SetByteResourceInBucket(val,key,"resources",groupName)
						if err1 != nil{
							glb.Error.Println(err1.Error())
						}
					}
				}
			}

		}

	}
}

func (gm *GroupManger) Flusher(){
	defer Wg.Done()
	for{
		//fmt.Println("Flushing DB ...")
		time.Sleep(FlushDelay * time.Second)
		glb.Debug.Println("Flushing DBs ...")
		var groups map[string]*Group

		GM.RLock()
		groups = GM.groups
		GM.RUnlock()
		for groupName,gp := range groups{
			gp.RLock()
			Permanent := gp.Permanent
			gp.RUnlock()
			if Permanent {
				//glb.Debug.Println("Flushing: ",groupName)
				GM.FlushDB(groupName,gp)
			}
		}
	}
}

func (gm *GroupManger) InstantFlushDB(groupName string){
	Wg.Add(1)
	go func(groupName string){
		gm.RLock()
		loaded := gm.isLoad[groupName]
		gp := gm.groups[groupName]
		dblock := gm.dbLock[groupName]
		gm.RUnlock()
		if !loaded {
			glb.Error.Println("DB isn't loaded!")
			return
		} else{
			//DBLock.Lock()
			dbData := make(map[string][]byte)

			gp.RLock()
			rdChanged := gp.RawDataChanged
			cdChanged := gp.ConfigDataChanged
			mdChanged := gp.MiddleDataChanged
			adChanged := gp.AlgoDataChanged
			rsChanged := gp.ResultDataChanged
			gp.RUnlock()

			if rdChanged {
				gp.RLock()
				rawData := gp.RawData
				gp.RUnlock()
				rawData.RLock()
				v, err := rawData.MarshalJSON()
				rawData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["rawData"] = v
				gp.Lock()
				gp.RawDataChanged = false
				gp.Unlock()
			}

			if cdChanged {
				gp.RLock()
				configData := gp.ConfigData
				gp.RUnlock()
				configData.RLock()
				v, err := configData.MarshalJSON()
				configData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["configData"] = v
				gp.Lock()
				gp.ConfigDataChanged = false
				gp.Unlock()
			}


			if mdChanged {
				gp.RLock()
				middleData := gp.MiddleData
				gp.RUnlock()
				middleData.RLock()
				v, err := middleData.MarshalJSON()
				middleData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["middleData"] = v

				gp.Lock()
				gp.MiddleDataChanged = false
				gp.Unlock()
			}

			if adChanged {
				gp.RLock()
				algoData := gp.AlgoData
				gp.RUnlock()
				algoData.RLock()
				v, err := algoData.MarshalJSON()
				algoData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}
				dbData["algoData"] = v

				gp.Lock()
				gp.AlgoDataChanged = false
				gp.Unlock()
			}

			resultDataList := make(map[string]parameters.Fingerprint)
			if rsChanged {
				gp.RLock()
				resultData := gp.ResultData
				gp.RUnlock()



				//
				//resultData.RLock()
				//resultDataList = resultData.Results
				//resultData.RUnlock()
				//resultData.Lock()
				//resultData.Results = make(map[string]parameters.Fingerprint) // delete trackresults data
				//resultData.Unlock()
				resultData.RLock()
				v, err := resultData.MarshalJSON()
				resultData.RUnlock()
				if err != nil {
					fmt.Errorf(err.Error())
				}

				dbData["resultData"] = v
				dbData["Results"] = []byte{}

				gp.Lock()
				gp.ResultDataChanged = false
				gp.Unlock()


			}


			{
				dblock.Lock()
				defer dblock.Unlock()


				for key,val := range dbData{
					if(key == "Results"){
						err1 := SetByteResourceInBucket(val,key,"Results",groupName)
						if err1 != nil{
							fmt.Errorf(err1.Error())
						}
					}else{
						for timeStamp,fp := range resultDataList{ // must put the list to db instantly

							err1 := SetByteResourceInBucket(parameters.DumpFingerprint(fp),timeStamp,"resources",groupName)
							if err1 != nil{
								fmt.Errorf(err1.Error())
							}
						}
					}
				}


			}

		}
		Wg.Done()
	}(groupName)
}


//func main(){
//
//	// Multi thread testing :
//Wwg.Add(1)
//	//	go func(gp *Group,num int){
//	//		gp.Set_d(1)
//	//		if(gp.Get_d()==0) {
//	//			fmt.Println(num, ": ", gp.Get_d())
//	//		}
//	//		wg.Done()
//	//	}(gp1,i)
//	//	wg.Add(1)
//	//	go func(gp *Group,num int){
//	//		gp.Set_d(0)
//	//		if(gp.Get_d()==1) {
//	//			fmt.Println(num, ": ", gp.Get_d())
//	//		}
//	//		wg.Done()
//	//	}(gp2,i)
//	//}
//	//
//	//wg.Wait()
//
//	//##################################
//
//	wg.Add(1)
//	go GM.Flusher()
//	//GM.GetGroup("t1").Set_d(1)
//	//fmt.Println(GM.GetGroup("t1").Get_d())
//
//	// main thread must wait for all thread to be done.
//	wg.Wait()
//}



// Setter & Getters APIs
// To access to each group it's better to use GM & groupName instead of group pointer

// Two usage forms :
// 1:(use it when many properties are needed and you want to set inner object properties line prop1.innerProp.aList[n])
// 		gp := dbm.GM.GetGroup(groupName).Get()
//		defer dbm.GM.GetGroup(groupName).Set(gp)
// 2:
//		gp := dbm.GM.GetGroup(groupName)
//		gp.Set_<property>(new value)
//		gp.Get_<property>()

func (gp *Group) NewRawDataStruct() *RawDataStruct {
	return &RawDataStruct{
		mutex:                  &sync.RWMutex{},
		group:                  gp,
		Fingerprints:           make(map[string]parameters.Fingerprint),
		FingerprintsOrdering:   []string{},
		LearnTrueLocations:     make(map[int64]string),
		TestValidTrueLocations: make(map[int64]string),
	}
}
func (gp *Group) NewConfigDataStruct() *ConfigDataStruct {
	return &ConfigDataStruct{
		mutex: &sync.RWMutex{},
		group: gp,
		Test:  1,
		GroupGraph:   			parameters.NewGraph(),
	}
}


func (gp *Group) NewMiddleDataStruct() *MiddleDataStruct {
	return &MiddleDataStruct{
		mutex:          &sync.RWMutex{},
		group:          gp,
		NetworkMacs:    make(map[string]map[string]bool),
		NetworkLocs:    make(map[string]map[string]bool),
		MacVariability: make(map[string]float32),
		MacCount:       make(map[string]int),
		MacCountByLoc:  make(map[string]map[string]int),
		UniqueLocs:     []string{},
		UniqueMacs:     []string{},
		LocCount:       make(map[string]int),
	}
}

func (gp *Group) NewAlgoDataStruct() *AlgoDataStruct {
	return &AlgoDataStruct{
		mutex: &sync.RWMutex{},
		group: gp,
		//BayesPriors:    		make(map[string]parameters.PriorParameters),
		//BayesResults:   		make(map[string]parameters.ResultsParameters),
		KnnFPs:         		parameters.NewKnnFingerprints(),
		//GroupGraph:   			parameters.NewGraph(),
	}
}

func (gp *Group) NewResultDataStruct() *ResultDataStruct {
	return &ResultDataStruct{
		mutex:           &sync.RWMutex{},
		group:           gp,
		//Results:         make(map[string]parameters.Fingerprint),
		AlgoAccuracy:          make(map[string]int),
		AlgoAccuracyLoc:       make(map[string]map[string]int),
		UserHistory:           make(map[string][]parameters.UserPositionJSON),
		UserResults:           make(map[string][]parameters.UserPositionJSON),
		AlgoTestErrorAccuracy: make(map[string]int),
		TestValidTracks:       []parameters.TestValidTrack{},
	}
}

// Return a copy of group
// Use it when many fields are needed (use Get_<property>() functions instead)
//func (gp *Group) Get() *Group {
//	newGp := &Group{}
//	gp.RLock()
//	*newGp = *gp
//	gp.RUnlock()
//	return newGp
//}

// Set all of group properties
// Use it when many fields are needed (use Set_<property>() functions instead)
//func (gp *Group) Set(newGp *Group) {
//	gp.Lock()
//	newGp.RLock()
//
//	elmNew := reflect.ValueOf(newGp).Elem()
//	elm := reflect.ValueOf(gp).Elem()
//	gpType := elmNew.Type()
//	fmt.Println(elmNew.NumField())
//	for i := 0; i < elmNew.NumField(); i++ {
//		fieldNew := elmNew.Field(i)
//		field := elm.Field(i)
//		//fmt.Println(gpType.Field(i).Name)
//		//fmt.Println(fieldNew.Type())
//		if(gpType.Field(i).Name!="RWMutex"){
//			field.Set(reflect.Value(fieldNew))
//		}
//	}
//	newGp.RUnlock()
//	gp.Unlock()
//	GM.SetDirtyBit(gp.Get_Name())
//}
// Return a copy of group
//x Use it when many fields are needed (use Get_<property>() functions instead)
func (gp *Group) Get() *Group {
	newGp := &Group{}
	gp.RLock()
	*newGp = *gp
	gp.RUnlock()
	return newGp
}

// Set all of group properties
// Use it when many fields are needed (use Set_<property>() functions instead)
//func (gp *Group) Set(newGp *Group) {
//	gp.Lock()
//	newGp.RLock()
//
//	elmNew := reflect.ValueOf(newGp).Elem()
//	elm := reflect.ValueOf(gp).Elem()
//	gpType := elmNew.Type()
//	fmt.Println(elmNew.NumField())
//	for i := 0; i < elmNew.NumField(); i++ {
//		fieldNew := elmNew.Field(i)
//		field := elm.Field(i)
//		//fmt.Println(gpType.Field(i).Name)
//		//fmt.Println(fieldNew.Type())
//		if(gpType.Field(i).Name!="RWMutex"){
//			field.Set(reflect.Value(fieldNew))
//		}
//	}
//	newGp.RUnlock()
//	gp.Unlock()
//	GM.SetDirtyBit(gp.Get_Name())
//}

func (gp *Group) Get_Name() string {
	gp.RLock()
	item := gp.Name
	gp.RUnlock()
	return item
}
func (gp *Group) Set_Name(new_item string){
	gp.Lock()
	gp.Name = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_Permanent() bool {
	gp.RLock()
	item := gp.Permanent
	gp.RUnlock()
	return item
}
func (gp *Group) Set_Permanent(new_item bool){
	gp.Lock()
	gp.Permanent = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}


func (rd *RawDataStruct) SetDirtyBit(){
	rd.RLock()
	gp := rd.group
	rd.RUnlock()

	gp.Lock()
	gpName := rd.group.Name
	gp.RawDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gpName)
}
func (gp *Group) Get_RawData() *RawDataStruct {
	gp.RLock()
	item := gp.RawData
	gp.RUnlock()
	return item
}

func (gp *Group) Get_RawData_Val() RawDataStruct {
	gp.RLock()
	item := *gp.RawData
	gp.RUnlock()
	return item
}

func (gp *Group) Get_RawData_Filtered_Val() RawDataStruct {
	gp.RLock()
	item := *gp.RawData
	gp.RUnlock()

	for _,fpIndex := range item.FingerprintsOrdering{
		fp := item.Fingerprints[fpIndex]
		FilterFingerprint(&fp)
		item.Fingerprints[fpIndex] = fp
	}

	return item
}

//Note: Use it just in buildgroup
func (gp *Group) Set_RawData(newItem *RawDataStruct) {
	gp.RLock()
	item := gp.RawData
	gp.RUnlock()
	//item.Lock()
	//newItem.RLock()

	elmNew := reflect.ValueOf(newItem).Elem()
	elm := reflect.ValueOf(item).Elem()
	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(itemType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.RawDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

//Note: Use it just in buildgroup
func (gp *Group) Set_RawData_Val(newItem RawDataStruct) {
	gp.RLock()
	item := gp.RawData
	gp.RUnlock()
	//item.Lock()
	//newItem.RLock()

	elmNew := reflect.ValueOf(newItem).Elem()
	elm := reflect.ValueOf(item).Elem()
	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(itemType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.RawDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}
////func (gp *Group) Set_RawData(new_item *RawDataStruct){
////	gp.Lock()
////	gp.RawData = new_item
////	//gp.RawDataChanged = true
////	gp.Unlock()
////	//GM.SetDirtyBit(gp.Get_Name())
////}
//
//func (md *MiddleDataStruct) Lock1(){
//	glb.Debug.Println("lock")
//	md.Lock()
//}
//func (md *MiddleDataStruct) Unlock1(){
//	glb.Debug.Println("unlock")
//	md.Unlock()
//}

func (cd *ConfigDataStruct) SetDirtyBit() {
	cd.RLock()
	gp := cd.group
	cd.RUnlock()

	gp.Lock()
	gpName := cd.group.Name
	gp.ConfigDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gpName)
}
func (gp *Group) Get_ConfigData() *ConfigDataStruct {
	gp.RLock()
	item := gp.ConfigData
	gp.RUnlock()
	return item
}

func (gp *Group) Get_ConfigData_Val() ConfigDataStruct {
	gp.RLock()
	item := *gp.ConfigData
	gp.RUnlock()
	return item
}

func (gp *Group) Set_ConfigData(newItem *ConfigDataStruct) {
	gp.RLock()
	item := gp.ConfigData
	gp.RUnlock()
	//item.Lock()
	//newItem.RLock()

	elmNew := reflect.ValueOf(newItem).Elem()
	elm := reflect.ValueOf(item).Elem()
	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(itemType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.ConfigDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Set_ConfigData_Val(newItem ConfigDataStruct) {
	gp.RLock()
	item := gp.ConfigData
	gp.RUnlock()
	//item.Lock()
	//newItem.RLock()

	elmNew := reflect.ValueOf(newItem).Elem()
	elm := reflect.ValueOf(item).Elem()
	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(itemType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.ConfigDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}


func (rd *MiddleDataStruct) SetDirtyBit(){
	rd.RLock()
	gp := rd.group
	rd.RUnlock()
	gp.Lock()
	gpName := rd.group.Name
	rd.group.MiddleDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gpName)
}
func (gp *Group) Get_MiddleData() *MiddleDataStruct {
	gp.RLock()
	item := gp.MiddleData
	gp.RUnlock()
	return item
}

func (gp *Group) Get_MiddleData_Val() MiddleDataStruct {
	gp.RLock()
	item := *gp.MiddleData
	gp.RUnlock()
	return item
}

func (gp *Group) Set_MiddleData(newItem *MiddleDataStruct) {
	gp.RLock()
	item := gp.MiddleData
	gp.RUnlock()
	//item.Lock()
	//newItem.RLock()

	elmNew := reflect.ValueOf(newItem).Elem()
	elm := reflect.ValueOf(item).Elem()
	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(itemType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.MiddleDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Set_MiddleData_Val(newItem MiddleDataStruct) {
	gp.RLock()
	item := gp.MiddleData
	gp.RUnlock()
	//item.Lock()
	//newItem.RLock()

	elmNew := reflect.ValueOf(newItem).Elem()
	elm := reflect.ValueOf(item).Elem()
	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(itemType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.MiddleDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

//func (gp *Group) Set_MiddleData(new_item *MiddleDataStruct){
//	gp.Lock()
//	gp.MiddleData = new_item
//	gp.MiddleDataChanged = true
//	gp.Unlock()
//	GM.SetDirtyBit(gp.Get_Name())
//}

func (ad *AlgoDataStruct) SetDirtyBit(){
	ad.RLock()
	gp := ad.group
	ad.RUnlock()

	gp.Lock()
	gpName := ad.group.Name
	gp.AlgoDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gpName)
}
func (gp *Group) Get_AlgoData() *AlgoDataStruct {
	gp.RLock()
	item := gp.AlgoData
	gp.RUnlock()
	return item
}

func (gp *Group) Get_AlgoData_Val() AlgoDataStruct {
	gp.RLock()
	item := *gp.AlgoData
	gp.RUnlock()
	return item
}

func (gp *Group) Set_AlgoData(newItem *AlgoDataStruct) {
	gp.RLock()
	item := gp.AlgoData
	gp.RUnlock()

	//newItem.RLock()
	elmNew := reflect.ValueOf(newItem).Elem()
	//item.Lock()
	elm := reflect.ValueOf(item).Elem()
	//item.Unlock()

	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())

	//item.Lock()
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		//item.Lock()
		field := elm.Field(i)
		//item.Unlock()
		//glb.Debug.Println(itemType.Field(i).Name)
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//item.Unlock()
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.AlgoDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Set_AlgoData_Val(newItemRaw AlgoDataStruct) {
	gp.RLock()
	item := gp.AlgoData
	gp.RUnlock()
	//item.Lock()
	newItem := &newItemRaw
	//newItem.RLock()

	elmNew := reflect.ValueOf(newItem).Elem()
	elm := reflect.ValueOf(item).Elem()
	itemType := elmNew.Type()
	//fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(itemType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if (itemType.Field(i).Name != "mutex" && itemType.Field(i).Name != "group") {
			glb.Debug.Println(itemType.Field(i).Name)
			//newItem.RLock()
			val := reflect.Value(fieldNew)
			//newItem.RUnlock()
			item.Lock()
			field.Set(val)
			item.Unlock()
		}
	}
	//newItem.RUnlock()
	//item.Unlock()
	gp.Lock()
	gp.AlgoDataChanged = true
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (rs *ResultDataStruct) SetDirtyBit(){
	rs.group.Lock()
	gpName := rs.group.Name
	rs.group.ResultDataChanged = true
	rs.group.Unlock()
	GM.SetDirtyBit(gpName)
}
func (gp *Group) Get_ResultData() *ResultDataStruct {
	gp.RLock()
	item := gp.ResultData
	gp.RUnlock()
	return item
}


//func (gp *Group) Set_ResultData(new_item *ResultDataStruct){
//	gp.Lock()
//	gp.ResultData = new_item
//	//gp.ResultDataChanged = true
//	gp.Unlock()
//	//GM.SetDirtyBit(gp.Get_Name())
//}




func (rd *RawDataStruct) Get_Fingerprints() map[string]parameters.Fingerprint {
	rd.RLock()
	item := rd.Fingerprints
	rd.RUnlock()

	return item
}

//Note: Use it just in buildgroup
func (rd *RawDataStruct) Set_Fingerprints(new_item map[string]parameters.Fingerprint){
	defer rd.SetDirtyBit()

	rd.Lock()
	rd.Fingerprints = new_item
	rd.Unlock()
}

func (rd *RawDataStruct) Get_FingerprintsOrdering() []string {
	rd.RLock()
	item := rd.FingerprintsOrdering
	rd.RUnlock()
	return item
}

//Note: Use it just in buildgroup
func (rd *RawDataStruct) Set_FingerprintsOrdering(new_item []string){
	defer rd.SetDirtyBit()

	rd.Lock()
	rd.FingerprintsOrdering = new_item
	rd.Unlock()
}
func (rd *RawDataStruct) Get_FPs() ([]string, map[string]parameters.Fingerprint) {
	rd.RLock()
	item1 := rd.FingerprintsOrdering
	item2 := rd.Fingerprints
	rd.RUnlock()
	return item1, item2
}

//Note: Use it just in buildgroup
func (rd *RawDataStruct) Set_FPs(new_item1 []string, new_item2 map[string]parameters.Fingerprint) {
	defer rd.SetDirtyBit()

	rd.Lock()
	rd.FingerprintsOrdering = new_item1
	rd.Fingerprints = new_item2
	rd.Unlock()
}


func (rd *RawDataStruct) Get_LearnTrueLocations() map[int64]string {
	rd.RLock()
	item := rd.LearnTrueLocations
	rd.RUnlock()
	return item
}
func (rd *RawDataStruct) Set_LearnTrueLocations(new_item map[int64]string) {
	defer rd.SetDirtyBit()

	rd.Lock()
	rd.LearnTrueLocations = new_item
	rd.Unlock()
}

func (rd *RawDataStruct) Get_TestValidTrueLocations() map[int64]string {
	rd.RLock()
	item := rd.TestValidTrueLocations
	rd.RUnlock()
	return item
}
func (rd *RawDataStruct) Set_TestValidTrueLocations(new_item map[int64]string) {
	defer rd.SetDirtyBit()

	rd.Lock()
	rd.TestValidTrueLocations = new_item
	rd.Unlock()
}


func (md *MiddleDataStruct) Get_NetworkMacs()  map[string]map[string]bool {
	md.RLock()
	item := md.NetworkMacs
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_NetworkMacs(new_item  map[string]map[string]bool){
	defer md.SetDirtyBit()

	md.Lock()
	md.NetworkMacs = new_item
	md.Unlock()
}

func (md *MiddleDataStruct) Get_NetworkLocs()  map[string]map[string]bool {
	md.RLock()
	item := md.NetworkLocs
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_NetworkLocs(new_item  map[string]map[string]bool){
	defer md.SetDirtyBit()

	md.Lock()
	md.NetworkLocs = new_item
	md.Unlock()
}

func (md *MiddleDataStruct) Get_MacVariability() map[string]float32  {
	md.RLock()
	item := md.MacVariability
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_MacVariability(new_item  map[string]float32 ){
	defer md.SetDirtyBit()

	md.Lock()
	md.MacVariability = new_item
	md.Unlock()
}

func (md *MiddleDataStruct) Get_MacCount() map[string]int   {
	md.RLock()
	item := md.MacCount
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_MacCount(new_item  map[string]int ){
	defer md.SetDirtyBit()

	md.Lock()
	md.MacCount = new_item
	md.Unlock()
}

func (md *MiddleDataStruct) Get_MacCountByLoc() map[string]map[string]int {
	md.RLock()
	item := md.MacCountByLoc
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_MacCountByLoc(new_item map[string]map[string]int ){
	defer md.SetDirtyBit()

	md.Lock()
	md.MacCountByLoc = new_item
	md.Unlock()
}

func (md *MiddleDataStruct) Get_UniqueLocs() []string {
	md.RLock()
	item := md.UniqueLocs
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_UniqueLocs(new_item []string ){
	defer md.SetDirtyBit()

	md.Lock()
	md.UniqueLocs = new_item
	md.Unlock()
}

func (md *MiddleDataStruct) Get_UniqueMacs() []string {
	md.RLock()
	item := md.UniqueMacs
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_UniqueMacs(new_item []string ){
	defer md.SetDirtyBit()

	md.Lock()
	md.UniqueMacs = new_item
	md.Unlock()
}

func (md *MiddleDataStruct) Get_LocCount() map[string]int {
	md.RLock()
	item := md.LocCount
	md.RUnlock()
	return item
}
func (md *MiddleDataStruct) Set_LocCount(new_item map[string]int ){
	defer md.SetDirtyBit()

	md.Lock()
	md.LocCount = new_item
	md.Unlock()
}

//
//func (ad *AlgoDataStruct) Get_BayesPriors() map[string]parameters.PriorParameters {
//	ad.RLock()
//	item := ad.BayesPriors
//	ad.RUnlock()
//	return item
//}
//func (ad *AlgoDataStruct) Set_BayesPriors(new_item map[string]parameters.PriorParameters){
//	defer ad.SetDirtyBit()
//
//	ad.Lock()
//	ad.BayesPriors = new_item
//	ad.Unlock()
//}
//
//func (ad *AlgoDataStruct) Get_BayesResults()  map[string]parameters.ResultsParameters  {
//	ad.RLock()
//	item := ad.BayesResults
//	ad.RUnlock()
//	return item
//}
//func (ad *AlgoDataStruct) Set_BayesResults(new_item  map[string]parameters.ResultsParameters ){
//	defer ad.SetDirtyBit()
//
//	ad.Lock()
//	ad.BayesResults = new_item
//	ad.Unlock()
//}

func (ad *AlgoDataStruct) Get_KnnFPs() parameters.KnnFingerprints  {
	ad.RLock()
	item := ad.KnnFPs
	ad.RUnlock()
	return item
}
func (ad *AlgoDataStruct) Set_KnnFPs(new_item  parameters.KnnFingerprints){
	defer ad.SetDirtyBit()

	ad.Lock()
	ad.KnnFPs = new_item
	ad.Unlock()
}

func (confdata *ConfigDataStruct) Get_GroupGraph() parameters.Graph  {
	confdata.RLock()
	item := confdata.GroupGraph
	confdata.RUnlock()
	return item
}
func (confdata *ConfigDataStruct) Set_GroupGraph(new_item  parameters.Graph){
	defer confdata.SetDirtyBit()

	glb.Debug.Println("Set_GroupGraph")

	confdata.Lock()
	confdata.GroupGraph = new_item
	confdata.Unlock()
}
//func (rs *ResultDataStruct) AppendResult(fp parameters.Fingerprint){
//	defer rs.SetDirtyBit()
//	rs.Lock()
//	rs.Results[strconv.FormatInt(fp.Timestamp, 10)] = fp
//	rs.Unlock()
//}
//func (rs *ResultDataStruct) GetUserResult(user string,n int)map[string]parameters.Fingerprint{
//	userResults := make(map[string]string)
//	rs.RLock()
//	results = rs.Results
//	rs.RUnlock()
//	count := 0
//	for fi
//
//	return results
//}

func (rs *ResultDataStruct) Get_AlgoAccuracy() map[string]int {
	rs.RLock()
	item := rs.AlgoAccuracy
	rs.RUnlock()
	return item
}
func (rs *ResultDataStruct) Set_AlgoAccuracy(algoName string, distError int){
	defer rs.SetDirtyBit()

	rs.Lock()

	if _,ok := rs.AlgoAccuracy[algoName];ok{
		rs.AlgoAccuracy[algoName] = distError
	}else{
		rs.AlgoAccuracy = make(map[string]int)
		rs.AlgoAccuracy[algoName] = distError
	}
	rs.Unlock()
}

func (rs *ResultDataStruct) Get_AlgoLocAccuracy() map[string]map[string]int  {
	rs.RLock()
	item := rs.AlgoAccuracyLoc
	rs.RUnlock()
	return item
}
func (rs *ResultDataStruct) Set_AlgoLocAccuracy(algoName string,loc string, distError int){
	//defer rs.SetDirtyBit()

	//glb.Error.Println(algoName," ",loc," ",distError)
	rs.Lock()
	if _,ok := rs.AlgoAccuracyLoc[algoName];ok{
		rs.AlgoAccuracyLoc[algoName][loc] = distError
	}else{
		rs.AlgoAccuracyLoc = make(map[string]map[string]int)
		rs.AlgoAccuracyLoc[algoName] = make(map[string]int)
		rs.AlgoAccuracyLoc[algoName][loc] = distError
	}
	rs.Unlock()
}

func (rs *ResultDataStruct) Append_UserHistory(user string, userPos parameters.UserPositionJSON) {
	//defer rs.SetDirtyBit()

	rs.Lock()
	if _, ok := rs.UserHistory[user]; ok {
		if len(rs.UserHistory[user]) < glb.MaxUserHistoryLen {
			rs.UserHistory[user] = append(rs.UserHistory[user], userPos)
		} else {
			tempUserHistory := []parameters.UserPositionJSON{}
			tempUserHistory = append(tempUserHistory, rs.UserHistory[user][1:]...)
			tempUserHistory = append(tempUserHistory, userPos)
			rs.UserHistory[user] = tempUserHistory
		}
	} else {
		//Todo: must provide standard way when new item added to groupcache structs 
		if rs.UserHistory == nil { // in old db there is now userHistory
			rs.UserHistory = make(map[string][]parameters.UserPositionJSON)
		}
		rs.UserHistory[user] = []parameters.UserPositionJSON{userPos}
		//glb.Debug.Println(rs.UserHistory[user])
	}
	rs.Unlock()
}
func (rs *ResultDataStruct) Get_UserHistory(user string) []parameters.UserPositionJSON {
	//defer rs.SetDirtyBit()

	history := []parameters.UserPositionJSON{}
	rs.RLock()
	if userHistory, ok := rs.UserHistory[user]; ok {
		history = userHistory
	} else {
		history = []parameters.UserPositionJSON{}
	}
	rs.RUnlock()
	return history
}
func (rs *ResultDataStruct) Get_AllHistory() map[string][]parameters.UserPositionJSON {
	//defer rs.SetDirtyBit()

	history := make(map[string][]parameters.UserPositionJSON)
	rs.RLock()
	history = rs.UserHistory
	rs.RUnlock()
	return history
}


func (rs *ResultDataStruct) Append_UserResults(user string, userPos parameters.UserPositionJSON) {
	defer rs.SetDirtyBit()

	rs.Lock()
	if _, ok := rs.UserResults[user]; ok {
		if len(rs.UserResults[user]) < glb.MaxUserResultsLen {
			rs.UserResults[user] = append(rs.UserResults[user], userPos)
		} else {
			tempUserResults := []parameters.UserPositionJSON{}
			tempUserResults = append(tempUserResults, rs.UserHistory[user][1:]...)
			tempUserResults = append(tempUserResults, userPos)
			rs.UserResults[user] = tempUserResults
		}
	} else {
		//Todo: must provide standard way when new item added to groupcache structs
		if rs.UserResults == nil { // in old db there is now userHistory
			rs.UserResults = make(map[string][]parameters.UserPositionJSON)
		}
		rs.UserResults[user] = []parameters.UserPositionJSON{userPos}
		//glb.Debug.Println(rs.UserHistory[user])
	}
	rs.Unlock()
}
func (rs *ResultDataStruct) Get_UserResults(user string) []parameters.UserPositionJSON {
	//defer rs.SetDirtyBit()

	results := []parameters.UserPositionJSON{}
	rs.RLock()
	if userResults, ok := rs.UserResults[user]; ok {
		results = userResults
	} else {
		results = []parameters.UserPositionJSON{}
	}
	rs.RUnlock()
	return results
}
func (rs *ResultDataStruct) Get_AllUserResults() map[string][]parameters.UserPositionJSON {
	//defer rs.SetDirtyBit()

	results := make(map[string][]parameters.UserPositionJSON)
	rs.RLock()
	results = rs.UserResults
	rs.RUnlock()
	return results
}
func (rs *ResultDataStruct) Clear_UserResults(user string) error {
	defer rs.SetDirtyBit()

	rs.Lock()
	if val, ok := rs.UserResults[user]; (ok && len(val) != 0) {
		rs.UserResults[user] = []parameters.UserPositionJSON{}
		rs.Unlock()
		return nil
	} else {
		rs.Unlock()
		return errors.New("There are no results for this user")
	}
}

func (rs *ResultDataStruct) Get_AlgoTestErrorAccuracy() map[string]int {
	rs.RLock()
	item := rs.AlgoTestErrorAccuracy
	rs.RUnlock()
	return item
}
func (rs *ResultDataStruct) Set_AlgoTestErrorAccuracy(algoName string, distError int) {
	defer rs.SetDirtyBit()

	rs.Lock()

	if _, ok := rs.AlgoTestErrorAccuracy[algoName]; ok {
		rs.AlgoTestErrorAccuracy[algoName] = distError
	} else {
		rs.AlgoTestErrorAccuracy = make(map[string]int)
		rs.AlgoTestErrorAccuracy[algoName] = distError
	}
	rs.Unlock()
}

func (rs *ResultDataStruct) Get_TestValidTracks() []parameters.TestValidTrack {
	rs.RLock()
	item := rs.TestValidTracks
	rs.RUnlock()
	return item
}
func (rs *ResultDataStruct) Set_TestValidTracks(new_item []parameters.TestValidTrack) {
	defer rs.SetDirtyBit()

	rs.Lock()
	rs.TestValidTracks = new_item
	rs.Unlock()
}
func (rs *ResultDataStruct) Append_TestValidTracks(new_item parameters.TestValidTrack) {
	defer rs.SetDirtyBit()

	rs.Lock()
	rs.TestValidTracks = append(rs.TestValidTracks, new_item)
	rs.Unlock()
}
