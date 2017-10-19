// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// db.go contains generic functions for parsing data from the database.
// This file is not exhaustive of all database functions, if they pertain to a
// specific property (fingerprinting/priors/parameters), it will instead be in respective file.

package main

import (
	"log"
	"os"
	"path"
	"strings"
	"time"
	"encoding/json"
	"github.com/boltdb/bolt"
	"fmt"
)

// checks is the database file exist or not.
func groupExists(group string) bool {
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, strings.ToLower(group)+".db")); os.IsNotExist(err) {
		return false
	}
	return true
}

// renames the network, then calls savePersistentParameters() function to save ps
func renameNetwork(group string, oldName string, newName string) {
	//todo: It's better to regenerate ps from the modified fingerprints bucket than modifying the current ps
	Debug.Println("Opening parameters")
	ps, _ := openParameters(group)
	Debug.Println("Opening persistent parameters")
	persistentPs, _ := openPersistentParameters(group)
	Debug.Println("Looping network macs")
	for n := range ps.NetworkMacs {
		if n == oldName {
			macs := []string{}
			Debug.Println("Looping macs for ", n)
			for mac := range ps.NetworkMacs[n] {
				macs = append(macs, mac)
			}
			Debug.Println("Adding to persistentPs")
			persistentPs.NetworkRenamed[newName] = macs
			delete(persistentPs.NetworkRenamed, oldName)
			break
		}
	}
	Debug.Println("Saving persistentPs")
	savePersistentParameters(group, persistentPs)
}

// if the users of group are cached, returns them.
// otherwise, read them from database, cache them and return them.
func getUsers(group string) []string {
	val, ok := getUserCache(group)
	if ok {
		return val
	}

	uniqueUsers := []string{}
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints-track"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			if !stringInSlice(v2.Username, uniqueUsers) {
				uniqueUsers = append(uniqueUsers, v2.Username)
			}
		}
		return nil
	})

	go setUserCache(group, uniqueUsers)
	return uniqueUsers
}

// returns MACs from fingerprints bucket
func getUniqueMacs(group string) []string {
	defer timeTrack(time.Now(), "getUniqueMacs")
	uniqueMacs := []string{}

	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			for _, router := range v2.WifiFingerprint {
				if !stringInSlice(router.Mac, uniqueMacs) {
					uniqueMacs = append(uniqueMacs, router.Mac)
				}
			}
		}
		return nil
	})
	return uniqueMacs
}

// returns all locations in a fingerprints bucket
func getUniqueLocations(group string) (uniqueLocs []string) {
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			if !stringInSlice(v2.Location, uniqueLocs) {
				uniqueLocs = append(uniqueLocs, v2.Location)
			}
		}
		return nil
	})
	return uniqueLocs
}

// returns count of each MAC in a fingerprints bucket
func getMacCount(group string) (macCount map[string]int) {
	macCount = make(map[string]int)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			for _, router := range v2.WifiFingerprint {
				if _, ok := macCount[router.Mac]; !ok {
					macCount[router.Mac] = 0
				}
				macCount[router.Mac]++
			}
		}
		return nil
	})
	return
}

// returns count of each MAC in a location
func getMacCountByLoc(group string) (macCountByLoc map[string]map[string]int) {
	macCountByLoc = make(map[string]map[string]int)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			if _, ok := macCountByLoc[v2.Location]; !ok {
				macCountByLoc[v2.Location] = make(map[string]int)
			}
			for _, router := range v2.WifiFingerprint {
				if _, ok := macCountByLoc[v2.Location][router.Mac]; !ok {
					macCountByLoc[v2.Location][router.Mac] = 0
				}
				macCountByLoc[v2.Location][router.Mac]++
			}
		}
		return nil
	})
	return
}

// Return admin users as map style(user:pass)
func getAdminUsers() (map[string]string, error) {
	userList := make(map[string]string)

	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, "users.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("adminList"))
		if len(v) == 0 {
			fmt.Errorf("Admin list is empty")
			return nil
		} else {
			err := json.Unmarshal(v, &userList)
			if err != nil {
				log.Fatal(err)
			}
			return err
		}
	})
	return userList, err
}

// Add an admin user or change his password
func addAdminUser(username string, password string) error {
	userList := make(map[string]string)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, "users.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create users bucket if doesn't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists([]byte("users"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		return err2
	})

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf("Resources don't exist")
		}
		v := b.Get([]byte("adminList"))
		if len(v) == 0 {
			fmt.Errorf("Admin list is empty")
			return nil
		} else {
			err := json.Unmarshal(v, &userList)
			if err != nil {
				log.Fatal(err)
			}
			return err
		}
	})

	if err != nil {
		return err
	}

	// Add an admin user or change his password
	userList[username] = password

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("users"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		marshalledUserList, _ := json.Marshal(userList)
		err2 = bucket.Put([]byte("adminList"), marshalledUserList)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})

	return err
}

// Set macs that to be filtered
func setFilterMacDB(group string, FilterMacs []string) error {
	Warning.Println(FilterMacs)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create filtermacs bucket if doesn't exist & set filtermacs
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))

		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		//Warning.Println(FilterMacs)
		marshalledFilterMacList, _ := json.Marshal(FilterMacs)
		err2 = bucket.Put([]byte("filterMacList"), marshalledFilterMacList)
		//Warning.Println("bucket creation problem :",err2)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		//Warning.Println("setFilterMacDB successfully")
		return err2
	})

	return err
}

// Get macs that to be filtered
func getFilterMacDB(group string) (error, []string) {
	var FilterMacs []string
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources don't exist")
		}
		v := b.Get([]byte("filterMacList"))
		if len(v) == 0 {
			fmt.Errorf("filterMacList list is empty")
			return nil
		} else {
			err := json.Unmarshal(v, &FilterMacs)
			if err != nil {
				log.Fatal(err)
			}
			return err
		}
	})

	return err, FilterMacs
}