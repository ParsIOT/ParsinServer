package dbm

import (
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jinzhu/copier"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
)

// make a db according to group Name
func PutFingerprintIntoDatabase(res parameters.Fingerprint, database string) error {
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, res.Group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err2 := tx.CreateBucketIfNotExists([]byte(database))
		if err2 != nil {
			return fmt.Errorf("create bucket: %s", err2)
		}

		//if res.Timestamp == 0 {
		//	res.Timestamp = time.Now().UnixNano()
		//}
		err2 = bucket.Put([]byte(strconv.FormatInt(res.Timestamp, 10)), parameters.DumpFingerprint(res))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	db.Close()
	return err
}

//returns the filtered macs from macs.json file and remove the other macs from fingerprint
func FilterFingerprint(curFP parameters.Fingerprint) parameters.Fingerprint {

	ok1 := GetRuntimePrf(curFP.Group).NeedToFilter      //check need for filtering
	ok2 := GetRuntimePrf(curFP.Group).NotNullFilterList //check that filterMap is null

	resFP := parameters.Fingerprint{}
	copier.Copy(&resFP, &curFP)

	if ok1 && ok2 {
		//glb.Debug.Println("1")
		//if _, ok := glb.RuntimeArgs.FilterMacsMap[curFP.Group]; !ok {
		//	err, filterMacs := GetFilterMacDB(curFP.Group)
		//
		//	glb.Warning.Println(filterMacs)
		//	if err != nil {
		//		return
		//	}
		//	glb.RuntimeArgs.FilterMacsMap[curFP.Group] = filterMacs
		//	//Rglb.RuntimeArgs.NeedToFilter[curFP.Group] = false //ToDo: filtering in loadfingerprint that was called by scikit.go not working! So i comment this line !
		//}

		filterMacsTemp := GetSharedPrf(curFP.Group).FilterMacsMap
		const (
			Combined int = 0 //Or general mode
			WIFIOnly int = 1
			BLEOnly  int = 2
		)

		technologyFilter := Combined // 0: combined; 1:just wifi; 2:just ble
		bleFoundIndex := glb.FindStringInSlice("BLE", filterMacsTemp)
		wifiFoundIndex := glb.FindStringInSlice("WIFI", filterMacsTemp)

		filterMacs := []string{}
		if wifiFoundIndex != -1 && bleFoundIndex == -1 { //Don't use BLEOnly and WIFIOnly with each other!
			technologyFilter = WIFIOnly
			// Delete WIFI from mac list
			filterMacs = glb.RemoveStringSliceItem(filterMacsTemp, wifiFoundIndex)
		} else if wifiFoundIndex == -1 && bleFoundIndex != -1 {
			technologyFilter = BLEOnly
			// Delete BLE from mac list
			filterMacs = glb.RemoveStringSliceItem(filterMacsTemp, bleFoundIndex)
		}

		//tempRouters1 := make([]parameters.Router, len(curFP.WifiFingerprint))
		tempRouters1 := []parameters.Router{}

		// filter according to technology
		// 1.Just WIFI
		if technologyFilter == WIFIOnly {
			for _, rt := range curFP.WifiFingerprint {
				theMac := rt.Mac
				if theMac[:4] == "WIFI" {
					//glb.Debug.Println("4")
					//Error.Println("filtered mac : ",curFP.WifiFingerprint[i].Mac)
					//tempRouters1[curNum] = curFP.WifiFingerprint[i]
					tempRouters1 = append(tempRouters1, rt)
					//tempRouters1[curNum].Mac = tempRouters1[curNum].Mac[0:len(tempRouters1[curNum].Mac)-1] + "0"
					//curNum++
				}
			}
		} else if technologyFilter == BLEOnly { // 2.Just BLE
			for _, rt := range curFP.WifiFingerprint {
				theMac := rt.Mac
				if theMac[:3] == "BLE" {
					//glb.Debug.Println("4")
					//Error.Println("filtered mac : ",curFP.WifiFingerprint[i].Mac)
					//tempRouters1[curNum] = curFP.WifiFingerprint[i]
					tempRouters1 = append(tempRouters1, rt)
					//tempRouters1[curNum].Mac = tempRouters1[curNum].Mac[0:len(tempRouters1[curNum].Mac)-1] + "0"
					//curNum++
				}
			}
		} else { // Combined mode
			tempRouters1 = curFP.WifiFingerprint
			filterMacs = filterMacsTemp
		}

		tempRouters2 := []parameters.Router{}

		if len(filterMacs) == 0 { // Just filter by WIFIOnly or BLEOnly and any mac isn't montioned in the macfilter list
			resFP.WifiFingerprint = tempRouters1
		} else { // Some macs entered to filter(maybe just in BLEOnly & WIFIOnly or in general(Combined) mode)
			for _, rt := range tempRouters1 {
				for _, mac := range filterMacs {
					if rt.Mac == mac {
						//glb.Debug.Println("4")
						//Error.Println("filtered mac : ",curFP.WifiFingerprint[i].Mac)
						//tempRouters1[curNum] = curFP.WifiFingerprint[i]
						tempRouters2 = append(tempRouters2, rt)
						//tempRouters1[curNum].Mac = tempRouters1[curNum].Mac[0:len(tempRouters1[curNum].Mac)-1] + "0"
						//curNum++
					}
				}
			}

			//glb.Debug.Println(tempRouters1[0:curNum])
			resFP.WifiFingerprint = tempRouters2
		}

	}

	return resFP
}

func LoadFingerprint(jsonByte []byte, doFilter bool) parameters.Fingerprint {
	var fp parameters.Fingerprint
	fp = parameters.LoadRawFingerprint(jsonByte)
	//glb.Debug.Println(fp)
	if len(fp.Group) == 0 {
		glb.Error.Println("fingerprint doesn't have group name:", fp)
		//panic("fingerprint doesn't have group name!")
		return fp
	}
	//t1 := len(fp.WifiFingerprint)
	if doFilter {
		fp = FilterFingerprint(fp)
	}
	//t2 := len(fp.WifiFingerprint)
	//if(t1 != t2 ){
	//	glb.Error.Println("Filtered #############")
	//}else{
	//	glb.Debug.Println("worked")
	//}

	//glb.Debug.Println(res)
	return fp
}

// make a folder that is named dump-groupName and dump track and learn db's data to files
func DumpFingerprints(group string) error {
	// glb.Debug.Println("Making dump-" + group + " directory")
	err := os.MkdirAll(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group), 0777)
	if err != nil {
		return err
	}

	// glb.Debug.Println("Opening db")
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0664, nil)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	// glb.Debug.Println("Opening file for learning fingerprints")
	// glb.Debug.Println(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "learning"))
	f, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "learning.json"), os.O_WRONLY|os.O_CREATE, 0664)
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

	//// glb.Debug.Println("Opening file for tracking fingerprints")
	//f, err = os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "tracking"), os.O_WRONLY|os.O_CREATE, 0664)
	//if err != nil {
	//	return err
	//}
	//// glb.Debug.Println("Writing fingerprints to file")
	//db.View(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("results"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		if _, err = f.WriteString(string(glb.DecompressByte(v)) + "\n"); err != nil {
	//			panic(err)
	//		}
	//	}
	//	return nil
	//})
	//f.Close()
	// glb.Debug.Println("Returning")

	return nil
}

// make a folder that is named dump-groupName and dump track and learn db's data to files
func DumpRawFingerprintsBaseDB(group string) error {
	// glb.Debug.Println("Making dump-" + group + " directory")
	err := os.MkdirAll(path.Join(glb.RuntimeArgs.SourcePath, "dumpraw-"+group), 0777)
	if err != nil {
		return err
	}

	// glb.Debug.Println("Opening db")
	db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0664, nil)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	// glb.Debug.Println("Opening file for learning fingerprints")
	// glb.Debug.Println(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "learning"))
	f, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dumpraw-"+group, "learning.csv"), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	var fingerprints []parameters.Fingerprint

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fp := parameters.LoadRawFingerprint(v)
			fingerprints = append(fingerprints, fp)
		}
		return nil
	})

	var uniqueMacs []string
	firstLine := "x,y,"
	for _, fp := range fingerprints {
		for _, rt := range fp.WifiFingerprint {
			if !glb.StringInSlice(rt.Mac, uniqueMacs) {
				uniqueMacs = append(uniqueMacs, rt.Mac)
				firstLine += rt.Mac + ","
			}
		}
	}

	if _, err = f.WriteString(firstLine); err != nil {
		panic(err)
	}

	// glb.Debug.Println("Writing fingerprints to file")
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fp := parameters.LoadRawFingerprint(v)
			line := fp.Location + ","

			for _, mac := range uniqueMacs {
				found := 0
				for _, rt := range fp.WifiFingerprint {
					if rt.Mac == mac {
						line += fmt.Sprintf("%v", rt.Rssi) + ","
						found = 1
						break
					}
				}
				if found != 1 {
					line += "-100,"
				}
			}

			if _, err = f.WriteString("\n" + line); err != nil {
				panic(err)
			}
		}
		return nil
	})
	f.Close()

	//// glb.Debug.Println("Opening file for tracking fingerprints")
	//f, err = os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "tracking"), os.O_WRONLY|os.O_CREATE, 0664)
	//if err != nil {
	//	return err
	//}
	//// glb.Debug.Println("Writing fingerprints to file")
	//db.View(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("results"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		if _, err = f.WriteString(string(glb.DecompressByte(v)) + "\n"); err != nil {
	//			panic(err)
	//		}
	//	}
	//	return nil
	//})
	//f.Close()
	// glb.Debug.Println("Returning")

	return nil
}

// make a folder that is named dump-groupName and dump track and learn db's data to files
func DumpRawFingerprints(group string) error {
	// glb.Debug.Println("Making dump-" + group + " directory")
	err := os.MkdirAll(path.Join(glb.RuntimeArgs.SourcePath, "dumpraw-"+group), 0777)
	if err != nil {
		return err
	}

	// glb.Debug.Println("Opening file for learning fingerprints")
	// glb.Debug.Println(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "learning"))
	f, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dumpraw-"+group, "learning.csv"), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	//var fingerprints []parameters.Fingerprint
	//
	//db.View(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("fingerprints"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		fp := parameters.LoadRawFingerprint(v)
	//		fingerprints = append(fingerprints, fp)
	//	}
	//	return nil
	//})

	rd := GM.GetGroup(group).Get_RawData()
	fingerprints := rd.Get_Fingerprints()

	var uniqueMacs []string
	firstLine := "x,y,"
	for _, fp := range fingerprints {
		for _, rt := range fp.WifiFingerprint {
			if !glb.StringInSlice(rt.Mac, uniqueMacs) {
				uniqueMacs = append(uniqueMacs, rt.Mac)
				firstLine += rt.Mac + ","
			}
		}
	}

	if _, err = f.WriteString(firstLine); err != nil {
		panic(err)
	}

	// glb.Debug.Println("Writing fingerprints to file")
	//db.View(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("fingerprints"))
	//	c := b.Cursor()a
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		fp := parameters.LoadRawFingerprint(v)
	for _, fp := range fingerprints {

		line := fp.Location + ","

		for _, mac := range uniqueMacs {
			found := 0
			for _, rt := range fp.WifiFingerprint {
				if rt.Mac == mac {
					line += fmt.Sprintf("%v", rt.Rssi) + ","
					found = 1
					break
				}
			}
			if found != 1 {
				line += "-100,"
			}
		}

		if _, err = f.WriteString("\n" + line); err != nil {
			panic(err)
		}
	}
	//return nil
	//})
	f.Close()

	//// glb.Debug.Println("Opening file for tracking fingerprints")
	//f, err = os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+group, "tracking"), os.O_WRONLY|os.O_CREATE, 0664)
	//if err != nil {
	//	return err
	//}
	//// glb.Debug.Println("Writing fingerprints to file")
	//db.View(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("results"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		if _, err = f.WriteString(string(glb.DecompressByte(v)) + "\n"); err != nil {
	//			panic(err)
	//		}
	//	}
	//	return nil
	//})
	//f.Close()
	// glb.Debug.Println("Returning")

	return nil
}

func DumpCalculatedFingerprints(groupName string) error {
	err := os.MkdirAll(path.Join(glb.RuntimeArgs.SourcePath, "dumpcalc-"+groupName), 0777)
	if err != nil {
		return err
	}

	// glb.Debug.Println("Opening file for learning fingerprints")
	// glb.Debug.Println(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+groupName, "learning"))
	fcsv, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dumpcalc-"+groupName, "learning.csv"), os.O_WRONLY|os.O_CREATE, 0664)
	defer fcsv.Close()
	if err != nil {
		return err
	}

	fjson, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dumpcalc-"+groupName, "learning.json"), os.O_WRONLY|os.O_CREATE, 0664)
	defer fjson.Close()
	if err != nil {
		return err
	}

	fTestValidJson, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dumpcalc-"+groupName, "tracking.json"), os.O_WRONLY|os.O_CREATE, 0664)
	defer fTestValidJson.Close()
	if err != nil {
		return err
	}



	var fingerprints []parameters.Fingerprint
	rd := GM.GetGroup(groupName).Get_RawData()
	fingerprintsInMemory := rd.Get_Fingerprints()
	fingerprintsOrdering := rd.Get_FingerprintsOrdering()

	sort.Strings(fingerprintsOrdering)
	for _, fpTime := range fingerprintsOrdering {
		fingerprints = append(fingerprints, fingerprintsInMemory[fpTime])
	}

	var uniqueMacs []string
	firstLine := "x,y,"
	for _, fp := range fingerprints {
		for _, rt := range fp.WifiFingerprint {
			if !glb.StringInSlice(rt.Mac, uniqueMacs) {
				uniqueMacs = append(uniqueMacs, rt.Mac)
			}
		}
	}

	sort.Strings(uniqueMacs)
	for _, mac := range uniqueMacs {
		firstLine += mac + ","
	}

	if _, err = fcsv.WriteString(firstLine); err != nil {
		panic(err)
	}

	for _, fp := range fingerprints {
		line := fp.Location + ","

		for _, mac := range uniqueMacs {
			found := 0
			for _, rt := range fp.WifiFingerprint {
				if rt.Mac == mac {
					line += fmt.Sprintf("%v", rt.Rssi) + ","
					found = 1
					break
				}
			}
			if found != 1 {
				line += "-100,"
			}
		}

		if _, err = fcsv.WriteString("\n" + line); err != nil {
			panic(err)
		}
	}

	for _, fp := range fingerprintsInMemory {
		fpJson, err := json.Marshal(fp)

		if err != nil {
			panic(err)
		}
		if _, err = fjson.WriteString(string(fpJson) + "\n"); err != nil {
			panic(err)
		}
	}

	rsd := GM.GetGroup(groupName).Get_ResultData()
	for _, testValid := range rsd.Get_TestValidTracks() {
		fpJson, err := json.Marshal(testValid.UserPosition.Fingerprint)

		if err != nil {
			panic(err)
		}
		if _, err = fTestValidJson.WriteString(string(fpJson) + "\n"); err != nil {
			panic(err)
		}
	}


	//// glb.Debug.Println("Opening file for tracking fingerprints")
	//f, err = os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, "dump-"+groupName, "tracking"), os.O_WRONLY|os.O_CREATE, 0664)
	//if err != nil {
	//	return err
	//}
	//// glb.Debug.Println("Writing fingerprints to file")
	//db.View(func(tx *bolt.Tx) error {
	//	b := tx.Bucket([]byte("results"))
	//	c := b.Cursor()
	//	for k, v := c.First(); k != nil; k, v = c.Next() {
	//		if _, err = f.WriteString(string(glb.DecompressByte(v)) + "\n"); err != nil {
	//			panic(err)
	//		}
	//	}
	//	return nil
	//})
	//f.Close()
	// glb.Debug.Println("Returning")

	return nil
}
