// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

//cache.go handles the global variables for caching and the clearing(With setter and getter functions).
//Each global variable as a shared memory has a RWMutex to protect it.
package main

import (
	"strings"
	"sync"
	"time"
)

//Containing a map: key=group name, value= a FullParameters instance
var psCache = struct {
	sync.RWMutex
	m map[string]FullParameters
}{m: make(map[string]FullParameters)}

//List of users that is tracked.
//Containing a map : key= group name, value= users list
var usersCache = struct {
	sync.RWMutex
	m map[string][]string
}{m: make(map[string][]string)}

//Containing a user position:
// key= concatenation of group name and user name, value= a UserPositionJSON instance
var userPositionCache = struct {
	sync.RWMutex
	m map[string]UserPositionJSON
}{m: make(map[string]UserPositionJSON)}

//It's used to understand that new Fingerprint was added to db.
//Containing a map : key= group name, value= is new fingerprint added to db?
var isLearning = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

//Running clearCache and clearCacheFast
func init() {
	go clearCache()
	go clearCacheFast()
}

//resetCache userCache variable
func clearCacheFast() {
	for {
		go resetCache("userCache")
		time.Sleep(time.Second * 30)
	}
}

//resetCache isLearning,psCache,userPositionCache variables
func clearCache() {
	for {
		//Debug.Println("Resetting cache")
		go resetCache("isLearning")
		go resetCache("psCache")
		go resetCache("userPositionCache")
		time.Sleep(time.Second * 600)
	}
}

//Initializing the variable
func resetCache(cache string) {
	if cache == "userCache" {
		usersCache.Lock()
		usersCache.m = make(map[string][]string)
		usersCache.Unlock()
	} else if cache == "userPositionCache" {
		userPositionCache.Lock()
		userPositionCache.m = make(map[string]UserPositionJSON)
		userPositionCache.Unlock()
	} else if cache == "psCache" {
		psCache.Lock()
		psCache.m = make(map[string]FullParameters)
		psCache.Unlock()
	} else if cache == "isLearning" {
		isLearning.Lock()
		isLearning.m = make(map[string]bool)
		isLearning.Unlock()
	}
}

//isLearning variable getter function
func getLearningCache(group string) (bool, bool) {
	//Debug.Println("getLearningCache")
	isLearning.RLock()
	cached, ok := isLearning.m[group]
	isLearning.RUnlock()
	return cached, ok
}

//isLearning variable setter function
func setLearningCache(group string, val bool) {
	isLearning.Lock()
	isLearning.m[group] = val
	isLearning.Unlock()
}

//usersCache variable getter fucntion
func getUserCache(group string) ([]string, bool) {
	//Debug.Println("Getting userCache")
	usersCache.RLock()
	cached, ok := usersCache.m[group]
	usersCache.RUnlock()
	return cached, ok
}

//userCache variable setter fucntion
func setUserCache(group string, users []string) {
	usersCache.Lock()
	usersCache.m[group] = users
	usersCache.Unlock()
}

//Append a user to the user list of a group (in usersCache variable)
func appendUserCache(group string, user string) {
	usersCache.Lock()
	if _, ok := usersCache.m[group]; ok {
		if len(usersCache.m[group]) == 0 {
			usersCache.m[group] = append([]string{}, strings.ToLower(user))
		}
	}
	usersCache.Unlock()
}

//psCache variable getter function
func getPsCache(group string) (FullParameters, bool) {
	//Debug.Println("Getting pscache")
	psCache.RLock()
	psCached, ok := psCache.m[group]
	psCache.RUnlock()
	return psCached, ok
}

//psCache variable setter function
func setPsCache(group string, ps FullParameters) {
	//Debug.Println("Setting pscache")
	psCache.Lock()
	psCache.m[group] = ps
	psCache.Unlock()
	return
}

//userPositionCache variable getter function
func getUserPositionCache(group_user string) (UserPositionJSON, bool) {
	//Debug.Println("getUserPositionCache")
	userPositionCache.RLock()
	cached, ok := userPositionCache.m[group_user]
	userPositionCache.RUnlock()
	return cached, ok
}

//userPositionCache variable setter function
func setUserPositionCache(group_user string, p UserPositionJSON) {
	//Debug.Println("setUserPositionCache")
	userPositionCache.Lock()
	userPositionCache.m[group_user] = p
	userPositionCache.Unlock()
	return
}
