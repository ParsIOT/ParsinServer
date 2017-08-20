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

	"github.com/boltdb/bolt"
)

// checks is the database file exist or not.
func groupExists(group string) bool {
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, strings.ToLower(group)+".db")); os.IsNotExist(err) {
		return false
	}
	return true
}

// renames the network.
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
