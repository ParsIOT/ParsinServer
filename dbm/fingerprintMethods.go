package dbm

import (
	"ParsinServer/algorithms/parameters"
	"github.com/boltdb/bolt"
	"path"
	"ParsinServer/glb"
	"log"
	"fmt"
	"time"
	"strconv"
	"os"
)


// make a db according to group name
func PutFingerprintIntoDatabase(res parameters.Fingerprint, database string) error {
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, res.Group+".db"), 0600, nil)
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

func LoadFingerprint(jsonByte []byte, doFilter bool) parameters.Fingerprint{
	var fp parameters.Fingerprint
	fp = parameters.LoadRawFingerprint(jsonByte)
	err, filterMacs := GetFilterMacDB(fp.Group)
	if err != nil {
		glb.Warning.Println("LoadFingerprint function: Error on GetFilterMacDB : ")
		glb.Warning.Println(err)
		return fp
	}
	if(doFilter){
		parameters.FilterRawFingerprint(&fp,filterMacs)
	}
	//glb.Debug.Println(res)
	return fp
}

func FilterFingerprint(fp *parameters.Fingerprint){
	err, filterMacs := GetFilterMacDB(fp.Group)
	glb.Warning.Println(filterMacs)
	if err != nil {
		glb.Error.Println("FilterFingerprint function: Error on GetFilterMacDB")
	}
	parameters.FilterRawFingerprint(fp,filterMacs)
}


// make a folder that is named dump-groupName and dump track and learn db's data to files
func DumpFingerprints(group string) error {
	// glb.Debug.Println("Making dump-" + group + " directory")
	err := os.MkdirAll(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group), 0777)
	if err != nil {
		return err
	}

	// glb.Debug.Println("Opening db")
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0664, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// glb.Debug.Println("Opening file for learning fingerprints")
	// glb.Debug.Println(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "learning"))
	f, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "learning"), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	// glb.Debug.Println("Writing fingerprints to file")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if _, err = f.WriteString(string(glb.DecompressByte(v)) + "\n"); err != nil {
				panic(err)
			}
		}
		return nil
	})
	f.Close()

	// glb.Debug.Println("Opening file for tracking fingerprints")
	f, err = os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "tracking"), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	// glb.Debug.Println("Writing fingerprints to file")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints-track"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if _, err = f.WriteString(string(glb.DecompressByte(v)) + "\n"); err != nil {
				panic(err)
			}
		}
		return nil
	})
	f.Close()
	// glb.Debug.Println("Returning")

	return nil
}