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
)

func TrackFingerprintsEmptyPosition(group string)(map[string]glb.UserPositionJSON,map[string]parameters.Fingerprint,error){
	userPositions := make(map[string]glb.UserPositionJSON)
	userFingerprints := make(map[string]parameters.Fingerprint)

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
		return userPositions,userFingerprints,err
	}

	defer db.Close()
	numUsersFound := 0
	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints-track"))
		if b == nil {
			return fmt.Errorf("Database not found")
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, true)
			if _, ok := userPositions[v2.Username]; !ok {
				timestampString := string(k)
				timestampUnixNano, _ := strconv.ParseInt(timestampString, 10, 64)
				UTCfromUnixNano := time.Unix(0, timestampUnixNano)
				foo := glb.UserPositionJSON{Time: UTCfromUnixNano.String()}
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

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
		return userJSON,userFingerprint,err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints-track"))
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
				UTCfromUnixNano := time.Unix(0, timestampUnixNano)
				userJSON.Time = UTCfromUnixNano.String()
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

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
		return fingerprints,err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints-track"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		numFound := 0
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			v2 := LoadFingerprint(v, true)
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
		return fmt.Errorf("User " + user + " not found")
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

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	var tempFp parameters.Fingerprint

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints-track"))
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
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, fromDB+".db"), 0664, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	db2, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, toDB+".db"), 0664, nil)
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
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints-track"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("fingerprints-track"))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				bucket.Put(k, v)
			}
			return nil
		})
		return nil
	})
}


func EditNameDB(location string, newname string, group string) int{
	toUpdate := make(map[string]string)
	numChanges := 0
	//glb.Debug.Println(group)
	_,fingerprintInMemory,err := GetLearnFingerPrints(group,false)
	if err!= nil{
		return 0
	}
	for fpTime,fp := range fingerprintInMemory{
		if fp.Location == location {
			tempFp := fp
			tempFp.Location = newname
			toUpdate[fpTime] = string(parameters.DumpFingerprint(tempFp))
		}
	}
	//glb.Debug.Println(fingerprintInMemory)

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()
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
		b := tx.Bucket([]byte("fingerprints-track"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				if v2.Location == location {
					v2.Location = newname
					toUpdate[string(k)] = string(parameters.DumpFingerprint(v2))
				}
			}
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints-track"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for k, v := range toUpdate {
			bucket.Put([]byte(k), []byte(v))
		}
		return nil
	})
	numChanges += len(toUpdate)

	//return numChanges,toUpdate
	return numChanges
}


func EditMacDB(oldmac string, newmac string, group string) int{
	toUpdate := make(map[string]string)
	numChanges := 0
	_,fingerprintInMemory,err := GetLearnFingerPrints(group,false)
	if err!= nil{
		return 0
	}
	for fpTime,fp := range fingerprintInMemory{
		for i, rt := range fp.WifiFingerprint {
			if rt.Mac == oldmac {
				tempFp := fp
				tempFp.WifiFingerprint[i].Mac = newmac
				toUpdate[fpTime] = string(parameters.DumpFingerprint(tempFp))
			}
		}
	}

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()
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
		b := tx.Bucket([]byte("fingerprints-track"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				for i, rt := range v2.WifiFingerprint {
					if rt.Mac == oldmac {
						v2.WifiFingerprint[i].Mac = newmac
						toUpdate[string(k)] = string(parameters.DumpFingerprint(v2))
					}
				}
			}
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints-track"))
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

func EditUserNameDB(user string, newname string, group string) int{
	toUpdate := make(map[string]string)
	numChanges := 0

	_,fingerprintInMemory,err := GetLearnFingerPrints(group,false)
	if err!= nil{
		return 0
	}
	for fpTime,fp := range fingerprintInMemory{
		if fp.Username == user {
			tempFp := fp
			tempFp.Username = newname
			toUpdate[fpTime] = string(parameters.DumpFingerprint(tempFp))
		}
	}


	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()
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
		b := tx.Bucket([]byte("fingerprints-track"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				if v2.Username == user {
					v2.Username = newname
					toUpdate[string(k)] = string(parameters.DumpFingerprint(v2))
				}
			}
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints-track"))
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

func DeleteLocationDB(location string,group string)int {
	numChanges := 0

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				if v2.Location == location {
					b.Delete(k)
					numChanges++
				}
			}
		}
		return nil
	})

	return numChanges
}


func DeleteLocationsDB(locations []string,group string) int{
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()
	numChanges := 0
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				for _, location := range locations {
					if v2.Location == location {
						b.Delete(k)
						numChanges++
						break
					}
				}
			}
		}
		return nil
	})

	return numChanges
}

func DeleteUser(user string, group string)int{
	numChanges := 0

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints-track"))
		if b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := LoadFingerprint(v, false)
				if v2.Username == user {
					b.Delete(k)
					numChanges++
				}
			}
		}
		return nil
	})

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
		toUpdate[fpTime] = string(parameters.DumpFingerprint(tempFp))
	}

	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

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
		b := tx.Bucket([]byte("fingerprints-track"))
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
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints-track"))
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