// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

//cache.go handles the global variables for caching and the clearing(With setter and getter functions).
//Each global variable as a shared memory has a RWMutex to protect it.
// [functionName]Threaded() means that this function used cached variables
package dbm

import (
	"ParsinServer/dbm/parameters"
	"strings"
	"sync"
	"time"
)

//Todo: These time can be dynamic for each group and each variable
//Todo: is it necessary to clear cache ?!!!!

var CacheResetFastPeriod time.Duration = 30 //second
var CacheResetPeriod time.Duration = 600    //second

//Containing a map: key=group Name, value= a FullParameters instance
// if there is psCache in memory, ps isn't got from db
//var psCache = struct {
//	sync.RWMutex
//	m map[string]parameters.FullParameters
//}{m: make(map[string]parameters.FullParameters)}

//var knnFPCache = struct {
//	sync.RWMutex
//	m map[string]parameters.KnnFingerprints
//}{m: make(map[string]parameters.KnnFingerprints)}

//List of users that is tracked.
//Containing a map : key= group Name, value= users list
var usersCache = struct {
	sync.RWMutex
	m map[string][]string
}{m: make(map[string][]string)}

//Containing a user position:
// key= concatenation of group Name and user Name, value= a UserPositionJSON instance
var userPositionCache = struct {
	sync.RWMutex
	m map[string]parameters.UserPositionJSON
}{m: make(map[string]parameters.UserPositionJSON)}

//It's used to understand that new Fingerprint was added to db.
//Containing a map : key= group Name, value= is new fingerprint added to db?
var isLearning = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

//Running clearCache and clearCacheFast
func init() {
	go ClearCache()
	go ClearCacheFast()
}

//resetCache userCache variable
func ClearCacheFast() {
	for {
		go ResetCache("userCache")
		time.Sleep(time.Second * CacheResetFastPeriod)
	}
}

//resetCache isLearning,psCache,userPositionCache variables
func ClearCache() {
	for {
		//Debug.Println("Resetting cache")
		go ResetCache("isLearning")
		go ResetCache("psCache")
		go ResetCache("userPositionCache")
		go ResetCache("knnFPCache")
		time.Sleep(time.Second * CacheResetPeriod)
	}
}

//Initializing the variable
func ResetCache(cache string) {
	if cache == "userCache" {
		usersCache.Lock()
		usersCache.m = make(map[string][]string)
		usersCache.Unlock()
	} else if cache == "userPositionCache" {
		userPositionCache.Lock()
		userPositionCache.m = make(map[string]parameters.UserPositionJSON)
		userPositionCache.Unlock()
		//} else if cache == "psCache" {
		//	psCache.Lock()
		//	psCache.m = make(map[string]parameters.FullParameters)
		//	psCache.Unlock()
	} else if cache == "isLearning" {
		isLearning.Lock()
		isLearning.m = make(map[string]bool)
		isLearning.Unlock()
		//} else if cache == "knnFPCache" {
		//	knnFPCache.Lock()
		//	knnFPCache.m = make(map[string]parameters.KnnFingerprints)
		//	knnFPCache.Unlock()
	}
}

//isLearning variable getter function
func GetLearningCache(group string) (bool, bool) {
	//Debug.Println("getLearningCache")
	isLearning.RLock()
	cached, ok := isLearning.m[group]
	isLearning.RUnlock()
	return cached, ok
}

//isLearning variable setter function
func SetLearningCache(group string, val bool) {
	isLearning.Lock()
	isLearning.m[group] = val
	isLearning.Unlock()
}

//usersCache variable getter fucntion
func GetUserCache(group string) ([]string, bool) {
	//Debug.Println("Getting userCache")
	usersCache.RLock()
	cached, ok := usersCache.m[group]
	usersCache.RUnlock()
	return cached, ok
}

//userCache variable setter fucntion
func SetUserCache(group string, users []string) {
	usersCache.Lock()
	usersCache.m[group] = users
	usersCache.Unlock()
}

//AppendResult a user to the user list of a group (in usersCache variable)
func AppendUserCache(group string, user string) {
	usersCache.Lock()
	if _, ok := usersCache.m[group]; ok {
		if len(usersCache.m[group]) == 0 {
			usersCache.m[group] = append([]string{}, strings.ToLower(user))
		}
	}
	usersCache.Unlock()
}

////psCache variable getter function
//func GetPsCache(group string) (parameters.FullParameters, bool) {
//	//Debug.Println("Getting pscache")
//	psCache.RLock()
//	psCached, ok := psCache.m[group]
//	psCache.RUnlock()
//	return psCached, ok
//}
//
////psCache variable setter function
//func SetPsCache(group string, ps parameters.FullParameters) {
//	//Debug.Println("Setting pscache")
//	psCache.Lock()
//	psCache.m[group] = ps
//	psCache.Unlock()
//	return
//}

//psCache variable getter function
//func GetKnnFPCache(group string) (parameters.KnnFingerprints, bool) {
//	//Debug.Println("Getting pscache")
//	knnFPCache.RLock()
//	knnFPCached, ok := knnFPCache.m[group]
//	knnFPCache.RUnlock()
//	return knnFPCached, ok
//}
//
////psCache variable setter function
//func SetKnnFPCache(group string, knnFP parameters.KnnFingerprints) {
//	//Debug.Println("Setting pscache")
//	knnFPCache.Lock()
//	knnFPCache.m[group] = knnFP
//	knnFPCache.Unlock()
//	return
//}

//userPositionCache variable getter function
func GetUserPositionCache(group_user string) (parameters.UserPositionJSON, bool) {
	//Debug.Println("getUserPositionCache")
	userPositionCache.RLock()
	cached, ok := userPositionCache.m[group_user]
	userPositionCache.RUnlock()
	return cached, ok
}

//userPositionCache variable setter function
func SetUserPositionCache(group_user string, p parameters.UserPositionJSON) {
	//Debug.Println("setUserPositionCache")
	userPositionCache.Lock()
	userPositionCache.m[group_user] = p
	userPositionCache.Unlock()
	return
}
