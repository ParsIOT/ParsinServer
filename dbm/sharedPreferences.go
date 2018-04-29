package dbm

import (
	"sync"
	"ParsinServer/glb"
	"errors"
)



type RawSharedPreferences struct {
	Mixin			  float64 				`json:"Mixin"`
	Cutoff			  float64 				`json:"Cutoff"`
	KnnK			  int 					`json:"KnnK"`
	MinRss			  int 					`json:"MinRss"`
	MinRssOpt		  int					`json:"MinRssOpt"`
	FilterMacsMap     []string				`json:"FilterMacsMap"`
}

func (shPrf *RawSharedPreferences) setPreference(prfName string, val interface{}) error{
	switch prfName {
	case "Mixin":
		shPrf.Mixin = val.(float64)
	case "Cutoff":
		shPrf.Cutoff = val.(float64)
	case "KnnK":
		shPrf.KnnK = val.(int)
	case "MinRss":
		shPrf.MinRss = val.(int)
	case "MinRssOpt":
		shPrf.MinRssOpt = val.(int)
	case "FilterMacsMap":
		shPrf.FilterMacsMap = val.([]string)
	default:
		return errors.New("Invalid RawSharedPreferences field")
	}
	return nil
}
//func (shPrf *RawSharedPreferences) getPreference(prfName string) interface{}{
//	switch prfName {
//		case "Mixin":
//			return shPrf.Mixin
//		case "Cutoff":
//			return shPrf.Cutoff
//		case "KnnK":
//			return shPrf.KnnK
//		case "MinRss":
//			return shPrf.MinRss
//		case "MinRssOpt":
//			return shPrf.MinRssOpt
//		case "FilterMacsMap":
//			return shPrf.FilterMacsMap
//		default:
//			return nil
//	}
//}

func NewRawSharedPreferences() RawSharedPreferences {
	return RawSharedPreferences{
		Mixin:     			float64(glb.DefaultMixin),
		Cutoff:    			float64(glb.DefaultCutoff),
		KnnK:      			int(glb.DefaultKnnK),
		MinRss:    			int(glb.MinRssi),
		MinRssOpt: 			int(glb.RuntimeArgs.MinRssOpt),
		FilterMacsMap: 		[]string{},
	}
}

type RawRuntimeSharedPreferences struct {
	NeedToFilter       bool 		`json:"NeedToFilter"`//check needing for filtering
	NotNullFilterList  bool			`json:"NotNullFilterList"` //check that filterMap is null(used to avoid filter fingerprint with null map)
}

func (shPrf *RawRuntimeSharedPreferences) setPreference(prfName string, val interface{}) error{
	switch prfName {
	case "NeedToFilter":
		shPrf.NeedToFilter = val.(bool)
	case "NotNullFilterList":
		shPrf.NotNullFilterList = val.(bool)
	default:
		return errors.New("Invalid RawRuntimeSharedPreferences field")
	}
	return nil
}

var SavedSharedPreferencesCache = struct {
	sync.RWMutex
	isLoad map[string]bool
	dbFields map[string]RawSharedPreferences
}{
	isLoad:			   make(map[string]bool),
	dbFields:          make(map[string]RawSharedPreferences),
}

var RuntimeSharedPreferencesCache = struct {
	sync.RWMutex
	isChangedShrPrf map[string]bool
	runtimeFields map[string]RawRuntimeSharedPreferences
}{
	isChangedShrPrf: make(map[string]bool),
	runtimeFields:     make(map[string]RawRuntimeSharedPreferences),
}

func NewRuntimeSharedPreferences() RawRuntimeSharedPreferences{
	return RawRuntimeSharedPreferences{
		NeedToFilter:     		false,
		NotNullFilterList:    	false,
	}
}


func GetSharedPrf(group string) RawSharedPreferences{
	//SavedSharedPreferencesCache.Lock()
	SavedSharedPreferencesCache.RLock()
	loaded := SavedSharedPreferencesCache.isLoad[group]
	//SavedSharedPreferencesCache.Unlock()
	SavedSharedPreferencesCache.RUnlock()

	if loaded{ // the group was loaded
		SavedSharedPreferencesCache.RLock()
		sharedPrf := SavedSharedPreferencesCache.dbFields[group]
		SavedSharedPreferencesCache.RUnlock()
		return sharedPrf
	}else{ // load shared preferences

		tempSharedPreferences,err := loadSharedPreferences(group)
		if err != nil{
			glb.Error.Println(err.Error())
			//panic(err.Error())
			return NewRawSharedPreferences()

		}

		SavedSharedPreferencesCache.Lock()
		SavedSharedPreferencesCache.dbFields[group] = tempSharedPreferences
		SavedSharedPreferencesCache.isLoad[group] = true
		SavedSharedPreferencesCache.Unlock()
		RuntimeSharedPreferencesCache.Lock()
		RuntimeSharedPreferencesCache.isChangedShrPrf[group] = true
		RuntimeSharedPreferencesCache.Unlock()
		return tempSharedPreferences
	}
}

func SetSharedPrf(group string, prfName string, val interface{}) error {
	SavedSharedPreferencesCache.RLock()
	loaded := SavedSharedPreferencesCache.isLoad[group]
	SavedSharedPreferencesCache.RUnlock()

	if loaded{
		SavedSharedPreferencesCache.RLock()
		sharedPrf := SavedSharedPreferencesCache.dbFields[group]
		SavedSharedPreferencesCache.RUnlock()
		//sharedPrf[prfName] = val
		sharedPrf.setPreference(prfName, val)
		SavedSharedPreferencesCache.Lock()
		SavedSharedPreferencesCache.dbFields[group] = sharedPrf
		SavedSharedPreferencesCache.Unlock()
		err := putSharedPreferences(group, sharedPrf)
		if err != nil{
			glb.Error.Println(err)
			return errors.New("Can't reset shared preferences")
		}
	}else{
		err := initializeSharedPreferences(group)
		if err != nil{
			glb.Error.Println(err)
			return errors.New("Can't reset shared preferences")
		}
		sharedPrf := GetSharedPrf(group)
		//sharedPrf, err := GetSharedPrf(group)
		//if err != nil{
		//	glb.Error.Println(err)
		//	return errors.New("Can't reset shared preferences")
		//}
		//sharedPrf[prfName] = val
		sharedPrf.setPreference(prfName, val)
		SavedSharedPreferencesCache.Lock()
		SavedSharedPreferencesCache.dbFields[group] = sharedPrf
		SavedSharedPreferencesCache.Unlock()

		err = loadRuntimePreferences(group)
		if err != nil{
			glb.Error.Println("Problem in loadRuntimePreferences")
			return errors.New("Problem in loadRuntimePreferences")
		}

		SavedSharedPreferencesCache.Lock()
		SavedSharedPreferencesCache.isLoad[group] = true
		SavedSharedPreferencesCache.Unlock()
		err = putSharedPreferences(group, sharedPrf)
		if err != nil{
			glb.Error.Println(err)
			return errors.New("Can't reset shared preferences")
		}
	}
	RuntimeSharedPreferencesCache.Lock()
	RuntimeSharedPreferencesCache.isChangedShrPrf[group] = true
	RuntimeSharedPreferencesCache.Unlock()
	return nil
}

// Set runtime preferences values
func loadRuntimePreferences(group string) error {
	SavedSharedPreferencesCache.RLock()
	shrPrf := SavedSharedPreferencesCache.dbFields[group]
	SavedSharedPreferencesCache.RUnlock()

	// Set NotNullFilterList and NeedToFilter
	filterMacsList := shrPrf.FilterMacsMap
	if(len(filterMacsList) != 0){
		RuntimeSharedPreferencesCache.RLock()
		runtimePreferences := RuntimeSharedPreferencesCache.runtimeFields[group]
		RuntimeSharedPreferencesCache.RUnlock()

		runtimePreferences.NeedToFilter = true
		runtimePreferences.NotNullFilterList = true

		RuntimeSharedPreferencesCache.Lock()
		RuntimeSharedPreferencesCache.runtimeFields[group] = runtimePreferences
		RuntimeSharedPreferencesCache.Unlock()
	}
	return nil
}

func GetRuntimePrf(group string) RawRuntimeSharedPreferences{
	SavedSharedPreferencesCache.RLock()
	loaded := SavedSharedPreferencesCache.isLoad[group]
	SavedSharedPreferencesCache.RUnlock()
	if !loaded{
		GetSharedPrf(group) //load SavedSharedPreferences
	}

	RuntimeSharedPreferencesCache.RLock()
	changedShrPrf := RuntimeSharedPreferencesCache.isChangedShrPrf[group]
	RuntimeSharedPreferencesCache.RUnlock()

	if changedShrPrf {
		err := loadRuntimePreferences(group)
		if err != nil {
			panic("Problem in loadRuntimePreferences")
			// glb.Error.Println("Problem in loadRuntimePreferences")
			return NewRuntimeSharedPreferences()
		}
		RuntimeSharedPreferencesCache.Lock()
		RuntimeSharedPreferencesCache.isChangedShrPrf[group] = false
		runtimePreferences := RuntimeSharedPreferencesCache.runtimeFields[group]
		RuntimeSharedPreferencesCache.Unlock()
		return runtimePreferences
	}else{
		RuntimeSharedPreferencesCache.RLock()
		runtimePreferences := RuntimeSharedPreferencesCache.runtimeFields[group]
		RuntimeSharedPreferencesCache.RUnlock()
		return runtimePreferences
	}

}

func SetRuntimePrf(group string, prfName string, val interface{}) error {
	SavedSharedPreferencesCache.RLock()
	loaded := SavedSharedPreferencesCache.isLoad[group]
	SavedSharedPreferencesCache.RUnlock()
	if loaded{
		RuntimeSharedPreferencesCache.RLock()
		runtimePrf := RuntimeSharedPreferencesCache.runtimeFields[group]
		RuntimeSharedPreferencesCache.RUnlock()
		//sharedPrf[prfName] = val
		err := runtimePrf.setPreference(prfName,val)
		if err != nil{
			glb.Error.Println("Problem to Runtime setPreference")
			return errors.New("Problem to Runtime setPreference")
		}

		RuntimeSharedPreferencesCache.Lock()
		RuntimeSharedPreferencesCache.runtimeFields[group] = runtimePrf
		RuntimeSharedPreferencesCache.Unlock()
	}else{
		GetSharedPrf(group) //load SavedSharedPreferences
		//if err != nil{
		//	glb.Error.Println("Problem to GetSharedPrf")
		//	return errors.New("Problem to GetSharedPrf")
		//}

		RuntimeSharedPreferencesCache.RLock()
		runtimePrf := RuntimeSharedPreferencesCache.runtimeFields[group]
		RuntimeSharedPreferencesCache.RUnlock()
		//sharedPrf[prfName] = val
		err := runtimePrf.setPreference(prfName,val)
		if err != nil{
			glb.Error.Println("Problem to Runtime setPreference")
			return errors.New("Problem to Runtime setPreference")
		}

		RuntimeSharedPreferencesCache.Lock()
		RuntimeSharedPreferencesCache.runtimeFields[group] = runtimePrf
		RuntimeSharedPreferencesCache.Unlock()
	}

	return nil
}