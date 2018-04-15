package dbm

import (
	"fmt"
	"sync"
	"time"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/glb"
	"reflect"
)

var GM GroupManger



var Wg sync.WaitGroup
var FlushDelay time.Duration = 8



//parameters Name must be lowercase that can't be access out of cachelib(must provide set&get func for each and provide locking mutex for each one)
type Group struct {
	sync.RWMutex
	Name           string
	Permanent      bool                                   // Some group doesn't need to be saved
	NetworkMacs    map[string]map[string]bool             // map of networks and then the associated macs in each
	NetworkLocs    map[string]map[string]bool             // map of the networks, and then the associated locations in each
	MacVariability map[string]float32                     // variability of macs
	MacCount      map[string]int                          // number of fingerprints of a AP in all data, regardless of the location; e.g. 10 of AP1, 12 of AP2, ...
	MacCountByLoc map[string]map[string]int               // number of fingerprints of a AP in a location; e.g. in location A, 10 of AP1, 12 of AP2, ...
	UniqueLocs    []string                                // a list of all unique locations e.g. {P1,P2,P3}
	UniqueMacs    []string                                // a list of all unique APs
	Priors        map[string]parameters.PriorParameters   // generate Priors for each network
	Results       map[string]parameters.ResultsParameters // generate Results for each network
	KnnFPs		  parameters.KnnFingerprints
	//learnDB	   map[string]map[string]{}interface	// group-->algorithm-->learnedData
}

func NewGroup(groupName string) *Group {
	return &Group{
		Name:           groupName,
		Permanent:      true,
		NetworkMacs:    make(map[string]map[string]bool),
		NetworkLocs:    make(map[string]map[string]bool),
		MacVariability: make(map[string]float32),
		MacCount:       make(map[string]int),
		MacCountByLoc:  make(map[string]map[string]int),
		UniqueLocs:     []string{},
		UniqueMacs:     []string{},
		Priors:         make(map[string]parameters.PriorParameters),
		Results:        make(map[string]parameters.ResultsParameters),
		KnnFPs:         parameters.NewKnnFingerprints(),
	}
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
	sync.RWMutex
	isLoad map[string]bool
	dbLock map[string]*sync.RWMutex
	dirtyBit map[string]bool
	groups map[string]*Group
}


func init(){
	GM = GroupManger{
		isLoad:			make(map[string]bool),
		dbLock:			make(map[string]*sync.RWMutex),
		dirtyBit:       make(map[string]bool),
		groups:         make(map[string]*Group),
	}

	//GM.NewGroup("t1")
	//GM.groups["t2"] = NewGroup("t1")
	//GM.isLoad["t2"] = true
	////wg.Add(1)
	//go GM.Flusher()
	////wg.Wait()

	// Must run on server.go
}

func (gm *GroupManger) NewGroup(groupName string) *Group {
	GM.RLock()
	groups := GM.groups
	GM.RUnlock()
	for gpName,gp := range groups{
		if(groupName == gpName){
			fmt.Errorf("There is a group exists with same Name.")
			return gp
		}
	}

	gp := NewGroup(groupName)
	GM.Lock()
	GM.groups[groupName] = gp
	GM.isLoad[groupName] = true
	GM.dbLock[groupName] = &sync.RWMutex{}
	GM.dirtyBit[groupName] = false
	GM.Lock()
	return gp
}

func (gm *GroupManger) GetGroup(groupName string) *Group {
	gm.LoadGroup(groupName)
	gm.RLock()
	group := gm.groups[groupName]
	gm.RUnlock()
	return group
}


func (gm *GroupManger) SetGroup(gp *Group) {
	//gm.LoadGroup(groupName)
	gm.Lock()
	gm.groups[gp.Name] = gp
	gm.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}


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

		{
			//dblock.dbLock()
			//GM.DBlock(groupName)
			dblock.Lock()
			defer dblock.Unlock()

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
			//db.Close()

			bytes,err1 := GetBytejsonResourceInBucket("parameters","resources",groupName)
			if err1 != nil {
				//log.Fatal(err1)
				glb.Error.Println(err1.Error())
				gp = GM.NewGroup(groupName)
			}else{
				//jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(bytes,&gp)
				gp.UnmarshalJSON(bytes)
			}
			//dblock.Unlock()
			//dblock.dbUnlock()
			//GM.DBUnlock(groupName)
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
	gm.RLock()
	dirtyBit := gm.dirtyBit[groupName]
	loaded := gm.isLoad[groupName]
	dblock := gm.dbLock[groupName]
	if dblock == nil{
		dblock = &sync.RWMutex{}
		gm.Lock()
		gm.dbLock[groupName] = dblock
		gm.Unlock()
	}

	gm.RUnlock()
	if(dirtyBit) {
		if !loaded {
			fmt.Errorf("DB isn't loaded!")
			return
		} else {
			//DBLock.Lock()
			{
				dblock.Lock()
				defer dblock.Unlock()
				//db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
				//if err != nil {
				//	log.Fatal(err)
				//}
				//db.Update(func(tx *bolt.Tx) error {
				//	b, _ := tx.CreateBucketIfNotExists([]byte("resources"))
				////v, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(&gp)
				v, err := gp.MarshalJSON()
				if err != nil {
					fmt.Errorf(err.Error())
					//return err
				}
				//	b.Put([]byte("parameters"), v)
				//	return nil
				//})
				//db.Close()
				err1 := SetByteResourceInBucket(v,"parameters","resources",groupName)
				if err1 != nil{
					fmt.Errorf(err.Error())
				}
			}

		}
		gm.Lock()
		gm.dirtyBit[groupName] = false
		gm.Unlock()
	}
}

func (gm *GroupManger) Flusher(){
	defer Wg.Done()
	for{
		//fmt.Println("Flushing DB ...")
		time.Sleep(FlushDelay * time.Second)
		glb.Debug.Println("Flushing DBs ...")
		GM.RLock()
		groups := GM.groups
		GM.RUnlock()
		for groupName,gp := range groups{
			if gp.Permanent {
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
		if !loaded{
			fmt.Errorf("DB isn't loaded!")
			return
		}else{
			//DBLock.Lock()
			{
				dblock.Lock()
				defer dblock.Unlock()
				//db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
				//if err != nil {
				//	log.Fatal(err)
				//}
				//db.Update(func(tx *bolt.Tx) error {
				//	b, _ := tx.CreateBucketIfNotExists([]byte("resources"))

				//v, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(&gp)
				v, err := gp.MarshalJSON()
				if err != nil {
					fmt.Errorf(err.Error())
					//return err
				}
				//	b.Put([]byte("parameters"), v)
				//	return nil
				//})
				//db.Close()
				err1 := SetByteResourceInBucket(v,"parameters","resources",groupName)
				if err1 != nil{
					fmt.Errorf(err.Error())
				}
			}
		}
		Wg.Done()
	}(groupName)
}


//func main(){
//
//	// Multi thread testing :
//
//
//
//
//
//
//
//	//gp1 := GM.GetGroup("t1")
//	//gp2 := GM.GetGroup("t2")
//	//for i:=1000;i<10000;i++{
//	//	wg.Add(1)
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



// Setter & Getters
// To access to each group it's better to use GM & groupName instead of group pointer

// Two usage forms :
// 1:(use it when many properties are needed and you want to set inner object properties line prop1.innerProp.aList[n])
// 		gp := dbm.GM.GetGroup(groupName).Get()
//		defer dbm.GM.GetGroup(groupName).Set(gp)
// 2:
//		gp := dbm.GM.GetGroup(groupName)
//		gp.Set_<property>(new value)
//		gp.Get_<property>()


// Return a copy of group
// Use it when many fields are needed (use Get_<property>() functions instead)
func (gp *Group) Get() *Group {
	newGp := &Group{}
	gp.RLock()
	*newGp = *gp
	gp.RUnlock()
	return newGp
}

// Set all of group properties
// Use it when many fields are needed (use Set_<property>() functions instead)
func (gp *Group) Set(newGp *Group) {
	gp.Lock()
	newGp.RLock()

	elmNew := reflect.ValueOf(newGp).Elem()
	elm := reflect.ValueOf(gp).Elem()
	gpType := elmNew.Type()
	fmt.Println(elmNew.NumField())
	for i := 0; i < elmNew.NumField(); i++ {
		fieldNew := elmNew.Field(i)
		field := elm.Field(i)
		//fmt.Println(gpType.Field(i).Name)
		//fmt.Println(fieldNew.Type())
		if(gpType.Field(i).Name!="RWMutex"){
			field.Set(reflect.Value(fieldNew))
		}
	}
	newGp.RUnlock()
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

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

func (gp *Group) Get_Temporary() bool {
	gp.RLock()
	item := gp.Permanent
	gp.RUnlock()
	return item
}
func (gp *Group) Set_Temporary(new_item bool){
	gp.Lock()
	gp.Permanent = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}


func (gp *Group) Get_NetworkMacs()  map[string]map[string]bool {
	gp.RLock()
	item := gp.NetworkMacs
	gp.RUnlock()
	return item
}
func (gp *Group) Set_NetworkMacs(new_item  map[string]map[string]bool){
	gp.Lock()
	gp.NetworkMacs = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_NetworkLocs()  map[string]map[string]bool {
	gp.RLock()
	item := gp.NetworkLocs
	gp.RUnlock()
	return item
}
func (gp *Group) Set_NetworkLocs(new_item  map[string]map[string]bool){
	gp.Lock()
	gp.NetworkLocs = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_MacVariability() map[string]float32  {
	gp.RLock()
	item := gp.MacVariability
	gp.RUnlock()
	return item
}
func (gp *Group) Set_MacVariability(new_item  map[string]float32 ){
	gp.Lock()
	gp.MacVariability = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_MacCount() map[string]int   {
	gp.RLock()
	item := gp.MacCount
	gp.RUnlock()
	return item
}
func (gp *Group) Set_MacCount(new_item  map[string]int ){
	gp.Lock()
	gp.MacCount = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_MacCountByLoc() map[string]map[string]int {
	gp.RLock()
	item := gp.MacCountByLoc
	gp.RUnlock()
	return item
}
func (gp *Group) Set_MacCountByLoc(new_item map[string]map[string]int ){
	gp.Lock()
	gp.MacCountByLoc = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_UniqueLocs() []string {
	gp.RLock()
	item := gp.UniqueLocs
	gp.RUnlock()
	return item
}
func (gp *Group) Set_UniqueLocs(new_item []string ){
	gp.Lock()
	gp.UniqueLocs = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_UniqueMacs() []string {
	gp.RLock()
	item := gp.UniqueMacs
	gp.RUnlock()
	return item
}
func (gp *Group) Set_UniqueMacs(new_item []string ){
	gp.Lock()
	gp.UniqueMacs = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_Priors() map[string]parameters.PriorParameters {
	gp.RLock()
	item := gp.Priors
	gp.RUnlock()
	return item
}
func (gp *Group) Set_Priors(new_item map[string]parameters.PriorParameters){
	gp.Lock()
	gp.Priors = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_Results()  map[string]parameters.ResultsParameters  {
	gp.RLock()
	item := gp.Results
	gp.RUnlock()
	return item
}
func (gp *Group) Set_Results(new_item  map[string]parameters.ResultsParameters ){
	gp.Lock()
	gp.Results = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}

func (gp *Group) Get_KnnFPs() parameters.KnnFingerprints  {
	gp.RLock()
	item := gp.KnnFPs
	gp.RUnlock()
	return item
}
func (gp *Group) Set_KnnFPs(new_item  parameters.KnnFingerprints){
	gp.Lock()
	gp.KnnFPs = new_item
	gp.Unlock()
	GM.SetDirtyBit(gp.Get_Name())
}
