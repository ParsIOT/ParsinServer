// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// db.go contains generic functions for parsing data from the database.
// This file is not exhaustive of all database functions, if they pertain to a
// specific property (fingerprinting/priors/parameters), it will instead be in respective file.

package dbm

import (
	"log"
	"os"
	"path"
	"strings"
	"time"
	"encoding/json"
	"github.com/boltdb/bolt"
	"fmt"
	"ParsinServer/glb"
	"strconv"
	"errors"
	"ParsinServer/algorithms/parameters"
)


func boltOpen(path string, mode os.FileMode, options *bolt.Options) (*bolt.DB, error) {
	// Works before db open
	//file := filepath.Base(path)
	//group := strings.TrimSuffix(file, filepath.Ext(file))
	glb.Debug.Println("db Opened and locked")
	blt, err := bolt.Open(path, mode, options)
	return blt, err
}

// checks is the database file exist or not.
func GroupExists(group string) bool {
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, strings.ToLower(group)+".db")); os.IsNotExist(err) {
		return false
	}
	return true
}

// renames the network, then calls savePersistentParameters() function to save ps
func RenameNetwork(group string, oldName string, newName string) error {
	//todo: It's better to regenerate ps from the modified fingerprints bucket than modifying the current ps
	//glb.Debug.Println("Opening parameters")
	var err error
	////ps, _ := OpenParameters(group)
	//gp := GM.GetGroup(group)
	//
	////glb.Debug.Println("Opening persistent parameters")
	//persistentPs, _ := OpenPersistentParameters(group)
	////glb.Debug.Println("Looping network macs")
	//for n := range gp.Get_NetworkMacs() {
	//	if n == oldName {
	//		macs := []string{}
	//		glb.Debug.Println("Looping macs for ", n)
	//		for mac := range gp.Get_NetworkMacs()[n] {
	//			macs = append(macs, mac)
	//		}
	//		glb.Debug.Println("Adding to persistentPs")
	//		persistentPs.NetworkRenamed[newName] = macs
	//		delete(persistentPs.NetworkRenamed, oldName)
	//		break
	//	}
	//}
	//glb.Debug.Println("Saving persistentPs")
	//err = SavePersistentParameters(group, persistentPs)
	return err
}

func GetRecentUsers(groupName string) []string {
	userHistories := GM.GetGroup(groupName).Get_ResultData().Get_AllUserResults()
	var users []string
	for user, _ := range userHistories {
		users = append(users, user)
	}
	return users
}

// if the users of group are cached, returns them.
// otherwise, read them from database, cache them and return them.
func GetUsers(group string) []string {
	//val, ok := GetUserCache(group)
	//if ok {
	//	return val
	//}

	uniqueUsers := []string{}
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}


	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Results"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := LoadFingerprint(v, false)
			if !glb.StringInSlice(v2.Username, uniqueUsers) {
				//glb.Debug.Println(v2)
				uniqueUsers = append(uniqueUsers, v2.Username)
			}
		}
		return nil
	})

	//go SetUserCache(group, uniqueUsers)
	return uniqueUsers
}

// returns MACs from fingerprints bucket
func GetUniqueMacs(group string) []string {
	uniqueMacs := []string{}

	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return uniqueMacs
	}
	for _,fp := range fingerprintInMemory{
		for _, router := range fp.WifiFingerprint {
			if !glb.StringInSlice(router.Mac, uniqueMacs) {
				uniqueMacs = append(uniqueMacs, router.Mac)
			}
		}
	}

	return uniqueMacs
}

// returns all locations in a fingerprints bucket
func GetUniqueLocations(group string) []string {

	var uniqueLocs []string
	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return uniqueLocs
	}
	for _,fp := range fingerprintInMemory{
		if !glb.StringInSlice(fp.Location, uniqueLocs) {
			uniqueLocs = append(uniqueLocs, fp.Location)
		}
	}
	return uniqueLocs
}

// returns count of each MAC in a fingerprints bucket
func GetMacCount(group string) (macCount map[string]int) {
	macCount = make(map[string]int)

	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return macCount
	}
	for _,fp := range fingerprintInMemory{
		for _, router := range fp.WifiFingerprint {
			if _, ok := macCount[router.Mac]; !ok {
				macCount[router.Mac] = 0
			}
			macCount[router.Mac]++
		}
	}
	return macCount
}

// returns count of each MAC in a location
func GetMacCountByLoc(group string) (macCountByLoc map[string]map[string]int) {
	macCountByLoc = make(map[string]map[string]int)

	_,fingerprintInMemory,err := GetLearnFingerPrints(group,true)
	if err!=nil{
		return macCountByLoc
	}
	for _,fp := range fingerprintInMemory{
		if _, ok := macCountByLoc[fp.Location]; !ok {
			macCountByLoc[fp.Location] = make(map[string]int)
		}
		for _, router := range fp.WifiFingerprint {
			if _, ok := macCountByLoc[fp.Location][router.Mac]; !ok {
				macCountByLoc[fp.Location][router.Mac] = 0
			}
			macCountByLoc[fp.Location][router.Mac]++
		}
	}

	return macCountByLoc
}

// Return admin users as map style(user:pass)
//func GetAdminUsers() (map[string]string, error) {
//	userList := make(map[string]string)
//
//	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, "users.db"), 0600, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	err = db.View(func(tx *bolt.Tx) error {
//		// Assume bucket exists and has keys
//		b := tx.Bucket([]byte("users"))
//		if b == nil {
//			glb.Error.Println("Resources dont exist")
//			return errors.New("")
//		}
//		v := b.Get([]byte("adminList"))
//		if len(v) == 0 {
//			fmt.Errorf("Admin list is empty")
//			return nil
//		} else {
//			err := json.Unmarshal(v, &userList)
//			if err != nil {
//				log.Fatal(err)
//			}
//			return err
//		}
//	})
//	return userList, err
//}

// Add an admin user or change his password
//func AddAdminUser(username string, password string) error {
//	userList := make(map[string]string)
//	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, "users.db"), 0600, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Create users bucket if doesn't exist
//	err = db.Update(func(tx *bolt.Tx) error {
//		_, err2 := tx.CreateBucketIfNotExists([]byte("users"))
//		if err2 != nil {
//			return fmt.Errorf("create bucket: %s", err2)
//		}
//		return err2
//	})
//
//	err = db.View(func(tx *bolt.Tx) error {
//		// Assume bucket exists and has keys
//		b := tx.Bucket([]byte("users"))
//		if b == nil {
//			return fmt.Errorf("Resources don't exist")
//		}
//		v := b.Get([]byte("adminList"))
//		if len(v) == 0 {
//			fmt.Errorf("Admin list is empty")
//			return nil
//		} else {
//			err := json.Unmarshal(v, &userList)
//			if err != nil {
//				log.Fatal(err)
//			}
//			return err
//		}
//	})
//
//	if err != nil {
//		return err
//	}
//
//	// Add an admin user or change his password
//	userList[username] = password
//
//	err = db.Update(func(tx *bolt.Tx) error {
//		bucket, err2 := tx.CreateBucketIfNotExists([]byte("users"))
//		if err2 != nil {
//			return fmt.Errorf("create bucket: %s", err2)
//		}
//		marshalledUserList, _ := json.Marshal(userList)
//		err2 = bucket.Put([]byte("adminList"), marshalledUserList)
//		if err2 != nil {
//			return fmt.Errorf("could add to bucket: %s", err2)
//		}
//		return err2
//	})
//
//	return err
//}


func GetLearnFingerPrints(group string,doFilter bool)([]string,map[string]parameters.Fingerprint,error){
	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var fingerprintsOrdering []string
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println("Can't get learn fingerprints.")
		return fingerprintsOrdering, fingerprintsInMemory, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		//gets the fingerprint bucket
		b := tx.Bucket([]byte("fingerprints"))
		if b == nil {
			glb.Error.Println("No fingerprint bucket")
			return fmt.Errorf("No fingerprint bucket")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fingerprintsInMemory[string(k)] = LoadFingerprint(v,doFilter)

			fingerprintsOrdering = append(fingerprintsOrdering, string(k))
		}
		return nil
	})
	if err != nil {
		glb.Debug.Println(group)
		glb.Error.Println("Can't get learn fingerprints.")
		return fingerprintsOrdering,fingerprintsInMemory,err
	}
	return fingerprintsOrdering,fingerprintsInMemory,nil
}


func PutDataIntoDatabase(res parameters.Fingerprint, database string) error {
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, res.Group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte(database))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		if res.Timestamp == 0 {
			res.Timestamp = time.Now().UnixNano()
		}
		err2 = bucket.Put([]byte(strconv.FormatInt(res.Timestamp, 10)), parameters.DumpFingerprint(res))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	db.Close()
	return err
}


func loadSharedPreferences(group string) (RawSharedPreferences,error) {
	//glb.Debug.Println(group)
	tempSharedPreferences := NewRawSharedPreferences()
	//glb.Debug.Println(path.Join(glb.RuntimeArgs.SourcePath, group+".db"))
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0755, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return tempSharedPreferences,errors.New("Can't reset shared preferences")
	}


	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return errors.New("Resources dont exist")
		}
		temp := b.Get([]byte("sharedPreferences"))
		if len(temp) == 0{
			glb.Error.Println("Empty sharedPreferences")
			return nil
		}
		err = json.Unmarshal(temp,&tempSharedPreferences)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		glb.Error.Println(err)
		return tempSharedPreferences,errors.New("Can't reset shared preferences")
	}
	return tempSharedPreferences,nil
}

func initializeSharedPreferences(group string) error {
	tempSharedPreferences := NewRawSharedPreferences()
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return errors.New("Can't reset shared preferences")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		tempSharedPreferencesJson, err3 := json.Marshal(tempSharedPreferences)
		if err3 != nil {
			return fmt.Errorf("Can't marshal : %s", err2)
		}

		err2 = bucket.Put([]byte("sharedPreferences"), tempSharedPreferencesJson)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	if err != nil {
		glb.Error.Println(err)
		return errors.New("Can't reset shared preferences")
	}
	return nil
}


func putSharedPreferences(group string, prf RawSharedPreferences) error {
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return errors.New("Can't set shared preferences")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		tempSharedPreferencesJson, err3 := json.Marshal(prf)
		if err3 != nil {
			return fmt.Errorf("Can't marshal : %s", err2)
		}

		err2 = bucket.Put([]byte("sharedPreferences"), tempSharedPreferencesJson)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	if err != nil {
		glb.Error.Println(err)
		return errors.New("Can't set shared preferences")
	}
	return nil
}
