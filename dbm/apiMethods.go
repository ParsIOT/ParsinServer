package dbm

import (
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

func TrackFingerprintsEmptyPosition(group string) (map[string]parameters.UserPositionJSON, map[string]parameters.Fingerprint, error) {
	userPositions := make(map[string]parameters.UserPositionJSON)
	userFingerprints := make(map[string]parameters.Fingerprint)

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return userPositions, userFingerprints, err
	}

	numUsersFound := 0
	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Results"))
		if b == nil {
			return fmt.Errorf("Database not found")
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, true)
			if _, ok := userPositions[v2.Username]; !ok {
				timestampString := string(k)
				timestampUnixNano, _ := strconv.ParseInt(timestampString, 10, 64)
				//UTCfromUnixNano := time.Unix(0, timestampUnixNano)
				foo := parameters.UserPositionJSON{Time: timestampUnixNano}
				userPositions[v2.Username] = foo
				userFingerprints[v2.Username] = v2
				numUsersFound++
			}
			if numUsersFound > 40 {
				return nil
			}
		}
		return nil
	})
	if err != nil {
		glb.Error.Println(err)
		return userPositions, userFingerprints, err
	}
	return userPositions, userFingerprints, nil
}

func TrackFingeprintEmptyPosition(user string, group string) (parameters.UserPositionJSON, parameters.Fingerprint, error) {
	var userJSON parameters.UserPositionJSON
	var userFingerprint parameters.Fingerprint

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return userJSON, userFingerprint, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Results"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, true)
			i++
			if i > 10000 {
				return fmt.Errorf("Too deep!")
			}
			if v2.Username == user {
				timestampString := string(k)
				timestampUnixNano, _ := strconv.ParseInt(timestampString, 10, 64)
				//UTCfromUnixNano := time.Unix(0, timestampUnixNano)
				//userJSON.Time = UTCfromUnixNano.String()
				userJSON.Time = timestampUnixNano
				userFingerprint = v2
				return nil
			}
		}
		return fmt.Errorf("User " + user + " not found")
	})

	if err != nil {
		glb.Error.Println(err)
		return userJSON, userFingerprint, err
	}
	return userJSON, userFingerprint, nil
}

func TrackFingerprints(user string, n int, group string) ([]parameters.Fingerprint, error) {

	var fingerprints []parameters.Fingerprint

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return fingerprints, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Results"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		numFound := 0
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, true)

			//glb.Debug.Println(v2)
			//glb.Debug.Println(user,v2.Username)
			if v2.Username == user {
				timestampString := string(k)
				timestampUnixNano, _ := strconv.ParseInt(timestampString, 10, 64)
				UTCfromUnixNano := time.Unix(0, timestampUnixNano)
				v2.Timestamp = UTCfromUnixNano.UnixNano()
				fingerprints = append(fingerprints, v2)
				numFound++
				if numFound >= n {
					return nil
				}
			}
		}
		if numFound == 0 {
			return fmt.Errorf("User " + user + " not found")
		} else {
			return nil
		}
	})
	if err != nil {
		glb.Error.Println(err)
		return fingerprints, err
	}
	return fingerprints, nil
}

// Returns the last location of a user and the last fingerprint that was sent
//Done: fingerprints-learn bucket isn't set but is used here! Returning the last learn fingerprint must be defined
func LastFingerprint(group string, user string) string {
	group = strings.ToLower(group)
	user = strings.ToLower(user)
	sentAs := ""

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	var tempFp parameters.Fingerprint

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Results"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v3 := LoadFingerprint(v, true)
			if v3.Username == user {
				tempFp = v3
				timestampString := string(k)
				timestampUnixNano, _ := strconv.ParseInt(timestampString, 10, 64)
				UTCfromUnixNano := time.Unix(0, timestampUnixNano)
				tempFp.Timestamp = UTCfromUnixNano.UnixNano()
				sentAs = "sent as /track\n"
				break
			}
		}
		return fmt.Errorf("User " + user + " not found")
	})
	db.Close()
	_, fingerprintsInMemory, err := GetLearnFingerPrints(group, true)
	if err != nil {
		return ""
	}
	for fpTime, fp := range fingerprintsInMemory {
		timestampString := fpTime
		timestampUnixNano, _ := strconv.ParseInt(timestampString, 10, 64)
		UTCfromUnixNano := time.Unix(0, timestampUnixNano).UnixNano()
		if UTCfromUnixNano < tempFp.Timestamp {
			break
		}
		if tempFp.Username == user {
			tempFp = fp
			tempFp.Timestamp = UTCfromUnixNano
			sentAs = "sent as /learn\n"
			break
		}
		glb.Error.Println("User " + user + " not found")
	}

	bJson, _ := json.MarshalIndent(tempFp, "", " ")
	return sentAs + string(bJson)
}

func MigrateDatabaseDB(fromDB string, toDB string) {
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, fromDB+".db"), 0664, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	db2, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, toDB+".db"), 0664, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db2.Close()

	db2.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("fingerprints"))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				bucket.Put(k, v)
			}
			return nil
		})
		return nil
	})

	db2.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Results"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Results"))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				bucket.Put(k, v)
			}
			return nil
		})
		return nil
	})
}

func EditLocDB(oldloc string, newloc string, groupName string) int {
	toUpdate := make(map[string]parameters.Fingerprint)
	numChanges := 0
	//glb.Debug.Println(groupName)
	rd := GM.GetGroup(groupName).Get_RawData()
	fingerprintInMemory := rd.Get_Fingerprints()
	//if err!= nil{r
	//	return 0
	//}
	for fpTime, fp := range fingerprintInMemory {
		if fp.Location == oldloc {
			tempFp := fp
			tempFp.Location = newloc
			toUpdate[fpTime] = tempFp
		}
	}
	//glb.Debug.Println(fingerprintInMemory)
	for fpTime, fp := range toUpdate {
		fingerprintInMemory[fpTime] = fp
	}

	numChanges += len(toUpdate)

	rd.SetDirtyBit()
	GM.InstantFlushDB(groupName)

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}
	//db.Update(func(tx *bolt.Tx) error {
	//	bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
	//	if err != nil {
	//		return fmt.Errorf("create bucket: %s", err)
	//	}
	//
	//	for k, v := range toUpdate {
	//		bucket.Put([]byte(k), []byte(v))
	//	}
	//	return nil
	//})

	toUpdateRes := make(map[string]string)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Results"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				if v2.Location == oldloc {
					v2.Location = newloc
					toUpdateRes[string(k)] = string(parameters.DumpFingerprint(v2))
				}
			}
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Results"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdateRes {
			bucket.Put([]byte(k), []byte(v))
		}
		return nil
	})
	numChanges += len(toUpdateRes)

	//return numChanges,toUpdate
	return numChanges
}

func EditLocBaseDB(oldloc string, newloc string, groupName string) int {
	toUpdate := make(map[string]parameters.Fingerprint)
	numChanges := 0
	//glb.Debug.Println(groupName)
	_, fingerprintInMemoryRaw, _ := GetLearnFingerPrints(groupName, false)
	//if err!= nil{r
	//	return 0
	//}
	//glb.Debug.Println(oldloc)
	//glb.Debug.Println(newloc)
	for fpTime, fp := range fingerprintInMemoryRaw {
		if fp.Location == oldloc {
			tempFp := fp
			tempFp.Location = newloc
			toUpdate[fpTime] = tempFp
		}
	}
	//glb.Debug.Println(fingerprintInMemory)
	//for fpTime,fp := range toUpdate{
	//	fingerprintInMemoryRaw[fpTime] = fp
	//}

	numChanges += len(toUpdate)

	//rd.SetDirtyBit()
	//GM.InstantFlushDB(groupName)

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdate {
			bucket.Put([]byte(k), []byte(string(parameters.DumpFingerprint(v))))
		}
		return nil
	})

	//glb.Debug.Println(numChanges)

	//
	//toUpdateRes := make(map[string]string)
	//
	//db.View(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("Results"))
	//	if b != nil {
	//		c := b.Cursor()
	//		for k, v := c.Last(); k != nil; k, v = c.Prev() {
	//			v2 := LoadFingerprint(v, false)
	//			if v2.Location == oldloc {
	//				v2.Location = newloc
	//				toUpdateRes[string(k)] = string(parameters.DumpFingerprint(v2))
	//			}
	//		}
	//	}
	//	return nil
	//})
	//
	//db.Update(func(tx *bolt.Tx) error {
	//	bucket, err := tx.CreateBucketIfNotExists([]byte("Results"))
	//	if err != nil {
	//		return fmt.Errorf("create bucket: %s", err)
	//	}
	//
	//	for k, v := range toUpdateRes {
	//		bucket.Put([]byte(k), []byte(v))
	//	}
	//	return nil
	//})
	//numChanges += len(toUpdateRes)

	//return numChanges,toUpdate
	return numChanges
}

// Direct access to db to change Mac names in fingerprints
func EditMacDB(oldmac string, newmac string, groupName string) int {
	toUpdate := make(map[string]parameters.Fingerprint)
	numChanges := 0
	//_,fingerprintInMemory,err := GetLearnFingerPrints(groupName,false)
	rd := GM.GetGroup(groupName).Get_RawData()
	fingerprintInMemory := rd.Get_Fingerprints()
	//if err!= nil{
	//	return 0
	//}
	for fpTime, fp := range fingerprintInMemory {
		for i, rt := range fp.WifiFingerprint {
			if rt.Mac == oldmac {
				tempFp := fp
				tempFp.WifiFingerprint[i].Mac = newmac
				toUpdate[fpTime] = tempFp
			}
		}
	}

	for fpTime, fp := range toUpdate {
		fingerprintInMemory[fpTime] = fp
	}

	numChanges += len(toUpdate)

	//fingerprintInMemory is map(pointer and no need to save)
	rd.Set_Fingerprints(fingerprintInMemory)
	//rd.SetDirtyBit()
	//GM.InstantFlushDB(groupName)

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdate {
			bucket.Put([]byte(k), parameters.DumpFingerprint(v))
		}
		return nil
	})

	toUpdateRes := make(map[string]string)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Results"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				for i, rt := range v2.WifiFingerprint {
					if rt.Mac == oldmac {
						v2.WifiFingerprint[i].Mac = newmac
						toUpdateRes[string(k)] = string(parameters.DumpFingerprint(v2))
					}
				}
			}
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Results"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdateRes {
			bucket.Put([]byte(k), []byte(v))
		}
		return nil
	})

	numChanges += len(toUpdateRes)

	return numChanges
}

func EditUserNameDB(user string, newname string, groupName string) int {
	toUpdate := make(map[string]parameters.Fingerprint)
	numChanges := 0

	rd := GM.GetGroup(groupName).Get_RawData()
	fingerprintInMemory := rd.Get_Fingerprints()
	//_,fingerprintInMemory,err := GetLearnFingerPrints(groupName,false)
	//if err!= nil{
	//	return 0
	//}
	for fpTime, fp := range fingerprintInMemory {
		if fp.Username == user {
			tempFp := fp
			tempFp.Username = newname
			toUpdate[fpTime] = tempFp
		}
	}

	for fpTime, fp := range toUpdate {
		fingerprintInMemory[fpTime] = fp
	}

	numChanges += len(toUpdate)

	rd.SetDirtyBit()
	GM.InstantFlushDB(groupName)

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}
	//db.Update(func(tx *bolt.Tx) error {
	//	bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
	//	if err != nil {
	//		return fmt.Errorf("create bucket: %s", err)
	//	}
	//
	//	for k, v := range toUpdate {
	//		bucket.Put([]byte(k), []byte(v))
	//	}
	//	return nil
	//})

	toUpdateRes := make(map[string]string)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Results"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				if v2.Username == user {
					v2.Username = newname
					toUpdateRes[string(k)] = string(parameters.DumpFingerprint(v2))
				}
			}
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Results"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdateRes {
			bucket.Put([]byte(k), []byte(v))
		}
		return nil
	})

	numChanges += len(toUpdate)

	return numChanges
}

func DeleteLocationDB(location string, groupName string) int {
	numChanges := 0

	rd := GM.GetGroup(groupName).Get_RawData_Val()

	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	//defer db.Close()
	//if err != nil {
	//	glb.Error.Println(err)
	//}

	for fpTime, fp := range rd.Fingerprints {
		if fp.Location == location {
			delete(rd.Fingerprints, fpTime)
			rd.FingerprintsOrdering = glb.DeleteSliceItemStr(rd.FingerprintsOrdering, fpTime)
			numChanges++
		}
	}
	rd.SetDirtyBit()
	glb.Debug.Println(numChanges)
	GM.InstantFlushDB(groupName)
	//GM.InstantFlushDB(groupName)
	//err = db.Update(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("fingerprints"))
	//	if b == nil {
	//		return errors.New("fingerprints dont exist")
	//	}
	//	c := b.Cursor()
	//	for k, v := c.Last(); k != nil; k, v = c.Prev() {
	//		v2 := LoadFingerprint(v, false)
	//		if v2.Location == location {
	//			b.Delete(k)
	//			numChanges++
	//		}
	//	}
	//	return nil
	//})
	//
	//if err != nil{
	//	glb.Error.Println(err)
	//}
	return numChanges
}

func DeleteLocationBaseDB(location string, group string) int {
	numChanges := 0

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		if b == nil {
			return errors.New("fingerprints dont exist")
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, false)
			if v2.Location == location {
				b.Delete(k)
				numChanges++
			}
		}
		return nil
	})

	if err != nil {
		glb.Error.Println(err)
	}
	return numChanges
}

func DeleteLocationsDB(locations []string, groupName string) int {
	numChanges := 0

	rd := GM.GetGroup(groupName).Get_RawData_Val()

	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	//defer db.Close()
	//if err != nil {
	//	glb.Error.Println(err)
	//}

	for _, location := range locations {
		for fpTime, fp := range rd.Fingerprints {
			if fp.Location == location {
				delete(rd.Fingerprints, fpTime)
				rd.FingerprintsOrdering = glb.DeleteSliceItemStr(rd.FingerprintsOrdering, fpTime)
				numChanges++
			}
		}
	}

	rd.SetDirtyBit()
	glb.Debug.Println(numChanges)
	GM.InstantFlushDB(groupName)
	//GM.InstantFlushDB(groupName)
	//err = db.Update(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("fingerprints"))
	//	if b == nil {
	//		return errors.New("fingerprints dont exist")
	//	}
	//	c := b.Cursor()
	//	for k, v := c.Last(); k != nil; k, v = c.Prev() {
	//		v2 := LoadFingerprint(v, false)
	//		if v2.Location == location {
	//			b.Delete(k)
	//			numChanges++
	//		}
	//	}
	//	return nil
	//})
	//
	//if err != nil{
	//	glb.Error.Println(err)
	//}
	return numChanges
}

func DeleteLocationsBaseDB(locations []string, groupName string) int {
	numChanges := 0

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		if b == nil {
			return errors.New("fingerprints dont exist")
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, false)
			if glb.StringInSlice(v2.Location, locations) {
				b.Delete(k)
				numChanges++
			}
		}
		return nil
	})

	if err != nil {
		glb.Error.Println(err)
	}
	return numChanges
}

//func DeleteLocationsDB(locations []string,group string) int{
//	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
//	defer db.Close()
//	if err != nil {
//		glb.Error.Println(err)
//	}
//	numChanges := 0
//	err = db.Update(func(tx *bolt.Tx) error {
//		b := tx.Bucket([]byte("fingerprints"))
//		if b == nil {
//			return errors.New("fingerprints dont exist")
//		}
//		c := b.Cursor()
//		for k, v := c.Last(); k != nil; k, v = c.Prev() {
//			v2 := LoadFingerprint(v, false)
//			for _, location := range locations {
//				if v2.Location == location {
//					b.Delete(k)
//					numChanges++
//					break
//				}
//			}
//		}
//
//		return nil
//	})
//
//	if err != nil{
//		glb.Error.Println(err)
//	}
//	return numChanges
//}

func DeleteUser(user string, group string) int {
	numChanges := 0

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Results"))
		if b == nil {
			return errors.New("fingerprints-track dont exist")
		}

		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, false)
			if v2.Username == user {
				b.Delete(k)
				numChanges++
			}
		}
		return nil

	})
	if err != nil {
		glb.Error.Println(err)
	}
	return numChanges

}

func ReformDBDB(groupName string) int {
	toUpdate := make(map[string]string)
	numChanges := 0

	gp := GM.GetGroup(groupName)
	rsd := gp.Get_ResultData()

	// rename group name in testvalid tracks
	testVallidTracks := rsd.Get_TestValidTracks()
	for i, _ := range testVallidTracks {
		testVallidTracks[i].UserPosition.Fingerprint.Group = groupName
	}
	rsd.Set_TestValidTracks(testVallidTracks)

	// rename group name in user results
	userResults := rsd.Get_AllUserResults()
	for user, results := range userResults {
		for i, _ := range results {
			results[i].Fingerprint.Group = groupName
		}
		userResults[user] = results
	}
	rsd.Set_AllUserResults(userResults)

	_, fingerprintInMemory, err := GetLearnFingerPrints(groupName, false)
	//glb.Warning.Println(fingerprintInMemory)
	if err != nil {
		return 0
	}
	for fpTime, fp := range fingerprintInMemory {
		tempFp := fp
		tempFp.Group = groupName
		tempFp.Location = glb.RoundLocationDim(tempFp.Location)
		toUpdate[fpTime] = string(parameters.DumpFingerprint(tempFp))
	}

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdate {
			bucket.Put([]byte(k), []byte(v))
		}
		return nil
	})

	numChanges += len(toUpdate)

	toUpdate = make(map[string]string)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Results"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)

				v2.Group = groupName
				toUpdate[string(k)] = string(parameters.DumpFingerprint(v2))

			}
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Results"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdate {
			bucket.Put([]byte(k), []byte(v))
		}
		return nil
	})

	numChanges += len(toUpdate)
	return numChanges
}

func GetCVResults(groupName string) map[string]int {
	gp := GM.GetGroup(groupName)
	glb.Debug.Println(gp.Get_Name())
	return gp.Get_ResultData().Get_AlgoAccuracy()
}

func GetCalcCompletionLevel() float64 {
	level := float64(glb.ProgressBarCurLevel) / float64(glb.ProgressBarLength)
	return level
}

func BuildGroupDB(groupName string) { //Todo: After each update in groupcache.go rerun this function
	fingerprintOrdering, fingerprintInMemoryRaw, _ := GetLearnFingerPrints(groupName, false)
	fingerprintInMemory := make(map[string]parameters.Fingerprint)
	for key, fp := range fingerprintInMemoryRaw {
		fp.Location = glb.RoundLocationDim(fp.Location)
		//glb.Debug.Println(fp.Location)
		fingerprintInMemory[key] = fp
	}
	//glb.Debug.Println(fingerprintOrdering)
	//glb.Debug.Println(fingerprintInMemory[fingerprintOrdering[0]])
	//glb.Debug.Println(groupName)
	//gp := GM.GetGroup(groupName)

	gp := GM.GetGroup(groupName)
	rd := gp.Get_RawData()
	rd.Set_Fingerprints(fingerprintInMemory)
	rd.Set_FingerprintsOrdering(fingerprintOrdering)
	rd.Set_FingerprintsBackup(fingerprintInMemory)
	rd.Set_FingerprintsOrderingBackup(fingerprintOrdering)

	//glb.Debug.Println(GM.isLoad[groupName])
	//GM.InstantFlushDB(groupName)
	//glb.Debug.Println(gp.Get_RawData_Val().FingerprintsOrdering)
}

func FingerprintLikeness(groupName string, loc string, maxFPDist float64) (map[string][]string, [][]string) {
	resultWithMainFP := []parameters.Fingerprint{}

	gp := GM.GetGroup(groupName)
	rd := gp.Get_RawData()
	md := gp.Get_MiddleData()

	FingerprintsOrdering := rd.Get_FingerprintsOrdering()
	FingerprintsData := rd.Get_Fingerprints()

	locFingerprintsOrdering := []string{}
	locFingerprintsData := make(map[string]parameters.Fingerprint)

	//locCalculatedDistance := []float64{} // a final distance used to sort and choose the one that should be deleted. =avg(physicalDistance/knnDistance)
	// its size must be the size of locFingerprintOrdering which is the number of fingerprints in each location
	totalFingerprintsOrdering := []string{}
	totalFingerprintsData := make(map[string]parameters.Fingerprint)

	CalculatedDistanceOverall := make(map[string][]float64)

	//glb.Debug.Println(len(FingerprintsOrdering))
	for _, fpTime := range FingerprintsOrdering {
		if FingerprintsData[fpTime].Location == loc {
			//glb.Debug.Println("format of loc: ",FingerprintsData[fpTime].Location ) // komeil, Just for test
			locFingerprintsOrdering = append(locFingerprintsOrdering, fpTime)
			locFingerprintsData[fpTime] = FingerprintsData[fpTime]
			resultWithMainFP = append(resultWithMainFP, FingerprintsData[fpTime])
		} else {
			totalFingerprintsOrdering = append(totalFingerprintsOrdering, fpTime)
			totalFingerprintsData[fpTime] = FingerprintsData[fpTime]
		}
	}

	//Distance calculating
	resultFPs := make(map[string]parameters.Fingerprint)
	uniqueMacs := md.Get_UniqueMacs()

	sort.Strings(uniqueMacs)
	for _, fpMain := range locFingerprintsData { // it loops over each fingerprint in selected location
		mac2RssMain := make(map[string]int)
		mainMacs := []string{}
		for _, rt := range fpMain.WifiFingerprint {
			mac2RssMain[rt.Mac] = rt.Rssi
			mainMacs = append(mainMacs, rt.Mac)
		}
		for _, mac := range uniqueMacs {
			if !glb.StringInSlice(mac, mainMacs) {
				mac2RssMain[mac] = glb.MinRssiOpt
			}
		}

		for _, fpTime := range totalFingerprintsOrdering { // loops over fingerprints of other locations all.
			//glb.Debug.Println(totalFingerprintsData[fpTime])
			fp := totalFingerprintsData[fpTime]
			mac2Rss := make(map[string]int)
			macs := []string{}

			// here we want to calculate physical distance between current locFingerprint and all other fingerprints in every location
			otherLocX, otherLocY := glb.GetLocationOfFingerprint(fp.Location)
			mainLocX, mainLocY := glb.GetLocationOfFingerprint(fpMain.Location)

			for _, rt := range fp.WifiFingerprint {
				mac2Rss[rt.Mac] = rt.Rssi
				macs = append(macs, rt.Mac)
			}
			for _, mac := range uniqueMacs {
				if !glb.StringInSlice(mac, macs) {
					mac2Rss[mac] = glb.MinRssiOpt
				}
			}

			distance := float64(0)
			knnParams := gp.Get_ConfigData().Get_KnnParameters()
			MaxEuclideanRssDist := knnParams.MaxEuclideanRssDist

			for mainMac, mainRssi := range mac2RssMain {
				if fpRss, ok := mac2Rss[mainMac]; ok {
					distance = distance + math.Pow(float64(mainRssi-fpRss), 2)
				} else {
					distance = distance + math.Pow(float64(MaxEuclideanRssDist), 2)
				}
			}
			distance = distance / float64(len(mac2RssMain))
			//if(distance==float64(0)){
			//	glb.Error.Println("###@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
			//}
			distance = math.Pow(distance, float64(1.0)/2)
			precision := 10

			/*glb.Debug.Println("testDistance",testDistance)
			precision := 3
			distance = glb.Round(math.Pow(distance, float64(1.0)/2), precision)
			glb.Debug.Println("distance with 3 precision: ", distance)
			precision = 5
			distance = glb.Round(math.Pow(distance, float64(1.0)/2), precision)
			glb.Debug.Println("distance with 5 precision: ", distance)
			precision = 8
			distance = glb.Round(math.Pow(distance, float64(1.0)/2), precision)
			glb.Debug.Println("distance with 8 precision: ", distance)*/

			if distance == float64(0) {
				//glb.Error.Println("Distance zero")
				//glb.Error.Println(job.mac2RssCur)
				//glb.Error.Println(job.mac2RssFP)
				distance = math.Pow(10, -1*float64(precision))
				//distance = maxDist
			}
			if distance <= maxFPDist {
				//glb.Debug.Println(fp)
				resultFPs[fpTime] = fp

				physicalDistance := glb.CalcDist(mainLocX, mainLocY, otherLocX, otherLocY)
				//glb.Debug.Println("### physical: ",physicalDistance,"knndistance:",distance)
				CalculatedDistanceOverall [fpMain.Location] = append(CalculatedDistanceOverall [fpMain.Location], physicalDistance/distance)
			}
		}
	}
	//glb.Debug.Println("**** calculated distance overall:\n", CalculatedDistanceOverall)
	//sortedCalculatedDistanceOverall := glb.SortReverseDictByVal(CalculatedDistanceOverall)
	//glb.Debug.Println("**** calculated distance overall:\n", sortedCalculatedDistanceOverall)
	resultMap := make(map[string][]string)
	for _, fp := range resultFPs {
		resultWithMainFP = append(resultWithMainFP, fp)
		//glb.Debug.Println(fp)
		if list, ok := resultMap[fp.Location]; ok {
			list = append(list, fp.GetTimestamp())
			resultMap[fp.Location] = list
		} else {
			resultMap[fp.Location] = []string{fp.GetTimestamp()}
		}
	}
	fingerprintRssDetails := [][]string{}

	//var uniqueMacs []string
	firstLine := []string{"x,y"}
	for _, fp := range resultWithMainFP {
		for _, rt := range fp.WifiFingerprint {
			if !glb.StringInSlice(rt.Mac, uniqueMacs) {
				uniqueMacs = append(uniqueMacs, rt.Mac)
			}
		}
	}
	sort.Strings(uniqueMacs)
	for _, mac := range uniqueMacs {
		firstLine = append(firstLine, mac)
	}
	fingerprintRssDetails = append(fingerprintRssDetails, firstLine)

	fingerprintRssRawDetails := make(map[string][][]string)
	locs := []string{}
	for _, fp := range resultWithMainFP {
		if !glb.StringInSlice(fp.Location, locs) {
			locs = append(locs, fp.Location)
		}
		line := []string{fp.Location}
		for _, mac := range uniqueMacs {
			macFound := false
			for _, rt := range fp.WifiFingerprint {
				if rt.Mac == mac {
					line = append(line, strconv.Itoa(rt.Rssi))
					macFound = true
					break;
				}
			}
			if !macFound {
				line = append(line, "")
			}
		}

		if val, ok := fingerprintRssRawDetails[fp.Location]; ok {
			val = append(val, line)
			fingerprintRssRawDetails[fp.Location] = val
		} else {
			fingerprintRssRawDetails[fp.Location] = [][]string{line}
		}
	}
	//sort dict by loc
	sort.Strings(locs)

	//create sorted fingerprintRssDetails
	for _, loc := range locs {
		for _, line := range fingerprintRssRawDetails[loc] {
			fingerprintRssDetails = append(fingerprintRssDetails, line)
		}
	}

	//for _,fpRSSs := range fingerprintRssDetails{
	//	line := ""
	//	for _,rss := range fpRSSs{
	//		line += rss
	//		line += ","
	//	}
	//	glb.Debug.Println(line)
	//}

	//glb.Debug.Println("$$$ check:",FingerprintsData["1501761048281042197"].Location)
	//fingerprintRssDetails = append(fingerprintRssDetails,sortedCalculatedDistanceOverall)

	return resultMap, fingerprintRssDetails
}

func GetMostSeenMacs(groupName string) []string {
	macCount := make(map[string]float64)

	rd := GM.GetGroup(groupName).Get_RawData()
	fpData := rd.Fingerprints

	for _, fp := range fpData {
		for _, rt := range fp.WifiFingerprint {
			if val, ok := macCount[rt.Mac]; ok {
				macCount[rt.Mac] = val + 1
			} else {
				macCount[rt.Mac] = 1
			}
		}
	}

	macSorted := glb.SortReverseDictByVal(macCount)

	// get N of most seen macs
	NumOfMustSeenMacs := 40

	glb.Debug.Println(macSorted)

	if (len(macSorted) < NumOfMustSeenMacs) {
		return macSorted
	} else {
		return macSorted[:NumOfMustSeenMacs]
	}
}

//Note:Use it before Preprocess
func RelocateFPLoc(groupName string) error {
	gp := GM.GetGroup(groupName)
	rd := gp.Get_RawData()
	// Get True Location logs from db
	allLocationLogs := rd.Get_LearnTrueLocations()
	glb.Debug.Println("TrueLocationLog :", allLocationLogs)
	// Get fingerprints from db
	fpData := rd.Get_Fingerprints()

	tempfpO := []string{}
	tempfpData := make(map[string]parameters.Fingerprint)

	for fpTime, fp := range fpData {
		//correct fp location
		newLoc, err := FindTrueFPloc(fp, allLocationLogs)
		if err == nil {
			tempfpO = append(tempfpO, fpTime)
			glb.Debug.Println(fp.Location, "-->", newLoc)
			fp.Location = newLoc
			tempfpData[fpTime] = fp
		} else {
			glb.Error.Println(err)
		}
	}

	rd.Set_FingerprintsOrdering(tempfpO)
	rd.Set_Fingerprints(tempfpData)
	rd.Set_FingerprintsOrderingBackup(tempfpO)
	rd.Set_FingerprintsBackup(tempfpData)

	glb.Debug.Println("RelocateFPLoc ended!")

	return nil
}

// find best fp location according to
func FindTrueFPloc(fp parameters.Fingerprint, allLocationLogs map[int64]string) (string, error) {
	fpTimeStamp := fp.Timestamp
	//newLoc := ""

	timeStamps := []int64{}
	for timestamp, _ := range allLocationLogs {
		timeStamps = glb.SortedInsert(timeStamps, timestamp)
	}
	lessUntil := 0
	for i, timeStamp := range timeStamps {
		//glb.Debug.Println(timeStamp-fpTimeStamp)
		if fpTimeStamp > timeStamp {
			lessUntil = i
			//glb.Debug.Println(i)
		} else {
			//glb.Debug.Println("ok ",i)
			if lessUntil != 0 {
				//	xy := allLocationLogs[timeStamp][:2]
				//newLoc = xy[1] + "," + xy[0]
				if timeStamp == fpTimeStamp {
					xy := strings.Split(allLocationLogs[timeStamp], ",")
					if !(len(xy) == 2) {
						err := errors.New("Location names aren't in the format of x,y")
						glb.Error.Println(err)
					}

					x, err1 := glb.StringToFloat(xy[0])
					y, err2 := glb.StringToFloat(xy[1])
					if err1 != nil || err2 != nil {
						glb.Error.Println(err1)
						glb.Error.Println(err2)
						return "", errors.New("Converting string 2 float problem")
					}
					return glb.IntToString(int(y)) + ".0," + glb.IntToString(int(x)) + ".0", nil
				} else {
					timeStamp1 := timeStamps[i-1]
					timeStamp2 := timeStamp
					if (timeStamp2-fpTimeStamp > int64(1*math.Pow(10, 9))) && (fpTimeStamp-timeStamp1 > int64(1*math.Pow(10, 9))) {
						break
					}
					if timeStamp2-fpTimeStamp > fpTimeStamp-timeStamp1 { // set first timestamp location
						//xy := allLocationLogs[timeStamp1][:2]

						xy := strings.Split(allLocationLogs[timeStamp1], ",")
						if !(len(xy) == 2) {
							err := errors.New("Location names aren't in the format of x,y")
							glb.Error.Println(err)
						}

						x, err1 := glb.StringToFloat(xy[0])
						y, err2 := glb.StringToFloat(xy[1])
						if err1 != nil || err2 != nil {
							glb.Error.Println(err1)
							glb.Error.Println(err2)
							return "", errors.New("Converting string 2 float problem")
						}
						return glb.IntToString(int(y)) + ".0," + glb.IntToString(int(x)) + ".0", nil
						//glb.Debug.Println(newLoc)
					} else { //set second timestamp location
						//xy := allLocationLogs[timeStamp2][:2]
						xy := strings.Split(allLocationLogs[timeStamp2], ",")
						if !(len(xy) == 2) {
							err := errors.New("Location names aren't in the format of x,y")
							glb.Error.Println(err)
						}

						x, err1 := glb.StringToFloat(xy[0])
						y, err2 := glb.StringToFloat(xy[1])
						if err1 != nil || err2 != nil {
							glb.Error.Println(err1)
							glb.Error.Println(err2)
							return "", errors.New("Converting string 2 float problem")
						}
						return glb.IntToString(int(y)) + ".0," + glb.IntToString(int(x)) + ".0", nil
					}
				}
				break
			} else {
				//glb.Error.Println("FP timestamp is before the uwb log timestamps")
			}
		}
	}
	glb.Error.Println("FindTrueFPloc on " + fp.Location + " ended but timestamp ranges doesn't match to relocate any fp")
	return "", errors.New("Timestamp range problem")

}

func GetRSSData(groupName string, mac string) [][]int {
	gp := GM.GetGroup(groupName)
	fpData := gp.Get_RawData().Get_Fingerprints()

	LatLngRSS := [][]int{}

	for _, fp := range fpData {
		for _, rt := range fp.WifiFingerprint {
			if (rt.Mac == mac) {
				xy := strings.Split(fp.Location, ",")
				x, err1 := glb.StringToFloat(xy[0])
				y, err2 := glb.StringToFloat(xy[1])
				if err1 != nil || err2 != nil {
					glb.Error.Println(err1)
					glb.Error.Println(err2)
				}
				LatLngRSS = append(LatLngRSS, []int{int(x), int(y), rt.Rssi})
			}
		}
	}
	return LatLngRSS

}

func CalculateTestError(groupName string, testValidTracks []parameters.TestValidTrack) (error, [][]string, []parameters.TestValidTrack) { //todo: create a page to show test-valid test fingerprint on map
	details := [][]string{}

	gp := GM.GetGroup(groupName)
	rsd := gp.Get_ResultData()

	// Get True Location logs from db
	allLocationLogs := gp.Get_RawData().Get_TestValidTrueLocations()

	//glb.Debug.Println("TrueLocationLog :", allLocationLogs)

	tempTestValidTracks := []parameters.TestValidTrack{}
	//glb.Debug.Println(testValidTracks)
	AlgoError := make(map[string]float64)
	numValidTestFPs := 0
	AlgoError["knn"] = float64(0)
	//AlgoError["bayes"] = float64(0)
	//AlgoError["svm"] = float64(0)

	for _, testValidTrack := range testValidTracks {
		userPos := testValidTrack.UserPosition
		//correct fp location
		TrueLoc := ""
		var err error
		if len(allLocationLogs) > 0 {
			TrueLoc, err = FindTrueFPloc(userPos.Fingerprint, allLocationLogs)
			if err != nil {
				glb.Error.Println(err)
				continue
			}
		} else {
			TrueLoc = userPos.Fingerprint.Location
			if TrueLoc == "" {
				continue
			}
		}

		testValidTrack.TrueLocation = TrueLoc
		numValidTestFPs++
		trueX, trueY := glb.GetDotFromString(TrueLoc)

		// Knn guess error:
		fpKnnX, fpKnnY := glb.GetDotFromString(userPos.KnnGuess)
		AlgoError["knn"] += glb.CalcDist(fpKnnX, fpKnnY, trueX, trueY)

		//fpBayesX, fpBayesY := glb.GetDotFromString(userPos.BayesGuess)
		//AlgoError["bayes"] += glb.CalcDist(fpBayesX, fpBayesY, trueX, trueY)

		//fpSvmX, fpSvmY := glb.GetDotFromString(userPos.SvmGuess)
		//AlgoError["svm"] += glb.CalcDist(fpSvmX, fpSvmY, trueX, trueY)
		details = append(details, []string{TrueLoc, userPos.KnnGuess})
		//details = append(details,[]string{TrueLoc,userPos.KnnGuess,userPos.SvmGuess})

		tempTestValidTracks = append(tempTestValidTracks, testValidTrack)
	}

	// Set new TestValidTracks list that true locations was set
	//rsd.Set_TestValidTracks(tempTestValidTracks)

	// Set AlgoTestError Accuracy
	glb.Debug.Println("Num of test-valid fp:", numValidTestFPs)
	for algoName, algoError := range AlgoError {
		glb.Debug.Println("TestValid "+algoName+" Error = ", algoError/float64(numValidTestFPs))
		rsd.Set_AlgoTestErrorAccuracy(algoName, int(algoError/float64(numValidTestFPs)))
	}
	return nil, details, tempTestValidTracks
}

func SetTrueLocationFromLog(groupName string, method string) error {
	if method != "learn" && method != "test" {
		return errors.New("runnig SetTrueLocationFromLog with invalid method")
	}

	file, err := os.Open(path.Join(glb.RuntimeArgs.SourcePath, "TrueLocationLogs/"+method+"/"+groupName+".log"))
	defer file.Close()
	if err != nil {
		glb.Error.Println(err)
		return errors.New("no such file or directory")
	}

	// Get location logs from uploaded true location logs
	locationLogs := make(map[string]map[int64][]string) // tag:timestamp:location(x,y,z)
	allLocationLogs := make(map[int64]string)           // timestamp:location(x,y,z)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		locLogStr := scanner.Text()

		// spliting the line
		locLog := strings.Split(locLogStr, ",")
		for i, item := range locLog {
			locLog[i] = strings.TrimSpace(item)
		}

		if (len(locLog) != 5) {
			return errors.New("Uploaded file doesn't have true location log format(timestamp,tag_name,x,y,z)")
		}
		tagName := locLog[1]
		if (tagName == "None") { // x,y,z are None too.
			glb.Debug.Println("None location")
			continue
		}

		// converting timestamp from string to int64
		timeStamp, err := strconv.ParseInt(locLog[0], 10, 64)
		if err != nil {
			glb.Error.Println(err)
		}

		xy := locLog[2:][:2] // get x,y from line str

		if log, ok := locationLogs[tagName]; ok {
			log[timeStamp] = xy
			locationLogs[tagName] = log
		} else {
			locationLogs[tagName] = make(map[int64][]string)
			locationLogs[tagName][timeStamp] = xy
		}

		// add to allLocationLogs

		allLocationLogs[timeStamp] = strings.Join(xy, ",")
	}
	if err := scanner.Err(); err != nil {
		glb.Error.Println(err)
		return err
	}

	rd := GM.GetGroup(groupName).Get_RawData()
	if method == "test" {
		rd.Set_TestValidTrueLocations(allLocationLogs)
		glb.Debug.Println(allLocationLogs)
	} else { //learn
		rd.Set_LearnTrueLocations(allLocationLogs)
		glb.Debug.Println(allLocationLogs)
	}

	return nil
}
