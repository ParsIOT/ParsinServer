package dbm

import (
	"fmt"
	"ParsinServer/algorithms/parameters"
	"github.com/boltdb/bolt"
	"path"
	"ParsinServer/glb"
	"encoding/json"
	"errors"
)


//return cached ps(a FullParameters instance) or get it from db then return
func OpenParameters(group string) (parameters.FullParameters, error) {

	psCached, ok := GetPsCache(group)
	if ok {
		return psCached, nil
	}

	var ps = *parameters.NewFullParameters()
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		v := b.Get([]byte("fullParameters"))
		ps = parameters.LoadParameters(v)
		return nil
	})

	go SetPsCache(group, ps)
	return ps, err
}

//save ps(a FullParameters instance) to db
func SaveParameters(group string, res parameters.FullParameters) error {
	//todo: why we should save ps in database? It can be regenerated from fingerprints bucket in db.
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		err2 = bucket.Put([]byte("fullParameters"), parameters.DumpParameters(res))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	return err
}



// Get persistentParameters from resources bucket in db
func OpenPersistentParameters(group string) (parameters.PersistentParameters, error) {
	var persistentPs = *parameters.NewPersistentParameters()
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		v := b.Get([]byte("persistentParameters"))
		json.Unmarshal(v, &persistentPs)
		return nil
	})
	return persistentPs, err
}


// Set persistentParameters to resources bucket in db (it's used in remednetwork() function)
func SavePersistentParameters(group string, res parameters.PersistentParameters) error {
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		glb.Error.Println(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("resources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		jsonByte, _ := json.Marshal(res)
		err2 = bucket.Put([]byte("persistentParameters"), jsonByte)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err)
		}
		return err2
	})
	glb.Debug.Println("Saved")
	return err
}

func SetKnnFingerprints(tempKnnFingerprints parameters.KnnFingerprints, group string) error {
	// Set KnnFingerprints to db
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte("knnresources"))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}
		//Debug.Println(tempKnnFingerprints)
		KnnFingerprintsJson, err3 := json.Marshal(tempKnnFingerprints)
		if err3 != nil {
			return fmt.Errorf("Can't marshal : %s", err2)
		}

		err2 = bucket.Put([]byte("knnFingerprints"), KnnFingerprintsJson)
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	if err != nil {
		return err
	}
	return nil
}

func GetKnnFingerprints(group string) (parameters.KnnFingerprints,error){
	var tempKnnFingerprints parameters.KnnFingerprints
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		return tempKnnFingerprints,err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("knnresources"))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		KnnFingerprintsJson := b.Get([]byte("knnFingerprints"))
		err = json.Unmarshal(KnnFingerprintsJson,&tempKnnFingerprints)
		if err != nil {
			glb.Error.Println(err)
		}

		return nil
	})

	if err != nil {
		return tempKnnFingerprints,err
	}
	return tempKnnFingerprints,nil
}

func SetResourceInBucket(keyName string,input interface{},bucketName string,group string) error {
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0755, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	//open the database and save the previously generated variables to database
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		tempInput,_ := json.Marshal(input)
		err = bucket.Put([]byte(keyName), []byte(tempInput)) //why svmData is not marshal?
		if err != nil {
			return fmt.Errorf("could add to bucket: %s", err)
		}
		return err
	})
	return err
}


func GetResourceInBucket(keyName string,input interface{},bucketName string,group string) error {
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0755, nil)
	if err != nil {
		glb.Error.Println(err)
		return err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		//gets some data from db
		v := b.Get([]byte(keyName))
		json.Unmarshal(v, &input)
		return err
	})
	if err != nil {
		glb.Error.Println(err)
		return err
	}
	return nil
}

func GetCompressedResourceInBucket(keyName string,input interface{},bucketName string,group string) error {
	db, err := bolt.Open(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0755, nil)
	if err != nil {
		glb.Error.Println(err)
		return err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			glb.Error.Println("Resources dont exist")
			return errors.New("")
		}
		//gets some data from db
		v := b.Get([]byte(keyName))
		//json.Unmarshal(v, &input)
		tempPs := parameters.LoadParameters(v)
		input = &tempPs
		return err
	})
	if err != nil {
		glb.Error.Println(err)
		return err
	}
	return nil
}