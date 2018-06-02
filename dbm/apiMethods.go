package dbm

import (
	"strconv"
	"time"
	"fmt"
	"github.com/boltdb/bolt"
	"strings"
	"path"
	"ParsinServer/glb"
	"ParsinServer/algorithms/parameters"
	"encoding/json"
	"errors"
)

func TrackFingerprintsEmptyPosition(group string)(map[string]glb.UserPositionJSON,map[string]parameters.Fingerprint,error){
	userPositions := make(map[string]glb.UserPositionJSON)
	userFingerprints := make(map[string]parameters.Fingerprint)

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return userPositions,userFingerprints,err
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
				foo := glb.UserPositionJSON{Time: timestampUnixNano}
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
		return userPositions,userFingerprints,err
	}
	return userPositions,userFingerprints,nil
}

func TrackFingeprintEmptyPosition(user string, group string)(glb.UserPositionJSON,parameters.Fingerprint,error){
	var userJSON glb.UserPositionJSON
	var userFingerprint parameters.Fingerprint

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return userJSON,userFingerprint,err
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
		return userJSON,userFingerprint,err
	}
	return userJSON,userFingerprint,nil
}


func TrackFingerprints(user string,n int, group string) ([]parameters.Fingerprint,error){

	var fingerprints []parameters.Fingerprint

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
		return fingerprints,err
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
		if numFound == 0{
			return fmt.Errorf("User " + user + " not found")
		}else{
			return nil
		}
	})
	if err != nil {
		glb.Error.Println(err)
		return fingerprints,err
	}
	return fingerprints,nil
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
	_,fingerprintsInMemory,err := GetLearnFingerPrints(group,true)
	if err != nil {
		return ""
	}
	for fpTime,fp := range fingerprintsInMemory{
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

func MigrateDatabaseDB(fromDB string,toDB string){
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


func EditLocDB(oldloc string, newloc string, groupName string) int{
	toUpdate := make(map[string]parameters.Fingerprint)
	numChanges := 0
	//glb.Debug.Println(groupName)
	rd := GM.GetGroup(groupName).Get_RawData()
	fingerprintInMemory := rd.Get_Fingerprints()
	//if err!= nil{r
	//	return 0
	//}
	for fpTime,fp := range fingerprintInMemory{
		if fp.Location == oldloc {
			tempFp := fp
			tempFp.Location = newloc
			toUpdate[fpTime] = tempFp
		}
	}
	//glb.Debug.Println(fingerprintInMemory)
	for fpTime,fp := range toUpdate{
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


// Direct access to db to change Mac names in fingerprints
func EditMacDB(oldmac string, newmac string, groupName string) int{
	toUpdate := make(map[string]parameters.Fingerprint)
	numChanges := 0
	//_,fingerprintInMemory,err := GetLearnFingerPrints(groupName,false)
	rd := GM.GetGroup(groupName).Get_RawData()
	fingerprintInMemory := rd.Get_Fingerprints()
	//if err!= nil{
	//	return 0
	//}
	for fpTime,fp := range fingerprintInMemory{
		for i, rt := range fp.WifiFingerprint {
			if rt.Mac == oldmac {
				tempFp := fp
				tempFp.WifiFingerprint[i].Mac = newmac
				toUpdate[fpTime] = tempFp
			}
		}
	}

	for fpTime,fp := range toUpdate{
		fingerprintInMemory[fpTime] = fp
	}

	numChanges += len(toUpdate)

	//fingerprintInMemory is map(pointer and no need to save)
	rd.SetDirtyBit()
	GM.InstantFlushDB(groupName)

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	defer db.Close()
	if err != nil {
		glb.Error.Println(err)
	}
	//
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

func EditUserNameDB(user string, newname string, groupName string) int{
	toUpdate := make(map[string]parameters.Fingerprint)
	numChanges := 0

	rd := GM.GetGroup(groupName).Get_RawData()
	fingerprintInMemory := rd.Get_Fingerprints()
	//_,fingerprintInMemory,err := GetLearnFingerPrints(groupName,false)
	//if err!= nil{
	//	return 0
	//}
	for fpTime,fp := range fingerprintInMemory{
		if fp.Username == user {
			tempFp := fp
			tempFp.Username = newname
			toUpdate[fpTime] = tempFp
		}
	}

	for fpTime,fp := range toUpdate{
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

func DeleteLocationDB(location string, groupName string)int {
	numChanges := 0


	rd := GM.GetGroup(groupName).Get_RawData_Val()

	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	//defer db.Close()
	//if err != nil {
	//	glb.Error.Println(err)
	//}

	for fpTime,fp := range rd.Fingerprints{
		if fp.Location == location{
			delete(rd.Fingerprints,fpTime)
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


func DeleteLocationsDB(locations []string, groupName string)int {
	numChanges := 0


	rd := GM.GetGroup(groupName).Get_RawData_Val()

	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"), 0600, nil)
	//defer db.Close()
	//if err != nil {
	//	glb.Error.Println(err)
	//}

	for _,location := range locations{
		for fpTime,fp := range rd.Fingerprints{
			if fp.Location == location{
				delete(rd.Fingerprints,fpTime)
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

func DeleteUser(user string, group string)int{
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
	if err != nil{
		glb.Error.Println(err)
	}
	return numChanges

}

func ReformDBDB(group string)int{
	toUpdate := make(map[string]string)
	numChanges := 0

	_,fingerprintInMemory,err := GetLearnFingerPrints(group,false)
	//glb.Warning.Println(fingerprintInMemory)
	if err!= nil{
		return 0
	}
	for fpTime,fp := range fingerprintInMemory{
		tempFp := fp
		tempFp.Group = group
		tempFp.Location = glb.RoundLocationDim(tempFp.Location)
		toUpdate[fpTime] = string(parameters.DumpFingerprint(tempFp))
	}

	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
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

				v2.Group = group
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

func GetCVResults(groupName string) map[string]int{
	gp := GM.GetGroup(groupName)
	glb.Debug.Println(gp.Get_Name())
	return gp.Get_ResultData().Get_AlgoAccuracy()
}

func GetCalcCompletionLevel() float64{
	level := float64(glb.ProgressBarCurLevel) / float64(glb.ProgressBarLength)
	return level
}

func BuildGroupDB(groupName string){
	fingerprintOrdering,fingerprintInMemoryRaw,_ := GetLearnFingerPrints(groupName,false)
	fingerprintInMemory := make(map[string]parameters.Fingerprint)
	for key,fp := range fingerprintInMemoryRaw{
		fp.Location = glb.RoundLocationDim(fp.Location)
		//glb.Debug.Println(fp.Location)
		fingerprintInMemory[key] = fp
	}
	//glb.Debug.Println(fingerprintOrdering)
	//glb.Debug.Println(fingerprintInMemory[fingerprintOrdering[0]])
	//glb.Debug.Println(groupName)
	gp := GM.GetGroup(groupName)
	rd := gp.Get_RawData()
	rd.Set_Fingerprints(fingerprintInMemory)
	rd.Set_FingerprintsOrdering(fingerprintOrdering)
	//glb.Debug.Println(GM.isLoad[groupName])
	//GM.InstantFlushDB(groupName)
	//glb.Debug.Println(gp.Get_RawData_Val().FingerprintsOrdering)
}

