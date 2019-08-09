package algorithms

import (
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

// # sudo apt-get install g++
// # wget http://www.csie.ntu.edu.tw/~cjlin/cgi-bin/libsvm.cgi?+http://www.csie.ntu.edu.tw/~cjlin/libsvm+tar.gz
// # tar -xvf libsvm-3.18.tar.gz
// # cd libsvm-3.18
// # make
//
// cp ~/Documents/find/svm ./
// cat svm | shuf > svm.shuffled
// ./svm-scale -l 0 -u 1 svm.shuffled > svm.shuffled.scaled
// head -n 500 svm.shuffled.scaled > learning
// tail -n 1500 svm.shuffled.scaled > testing
// ./svm-train -s 0 -t 0 -b 1 learning > /dev/null
// ./svm-predict -b 1 testing learning.model out

type Svm struct {
	Data     string
	Mac      map[string]string
	Location map[string]string
}

//gets the fingerprints data from db and make them in the format of LIBSVM.
func DumpFingerprintsSVM(group string) {
	//macs: every mac as key and the value is a unique id.
	macs := make(map[string]int)
	//locations: every location as key and the value is a unique id.
	locations := make(map[string]int)
	//macsFromID: reverse map of macs
	macsFromID := make(map[string]string)
	//locationsFromID: reverse map of locations
	locationsFromID := make(map[string]string)
	macI := 1
	locationI := 1

	_, fingerprintsInMemory, err := dbm.GetLearnFingerPrints(group, true)
	if err != nil {
		glb.Error.Println(err)
	}
	for _, fp := range fingerprintsInMemory {
		for _, fingerprint := range fp.WifiFingerprint {
			if _, ok := macs[fingerprint.Mac]; !ok {

				macs[fingerprint.Mac] = macI

				macsFromID[strconv.Itoa(macI)] = fingerprint.Mac
				macI++
			}
		}
		if _, ok := locations[fp.Location]; !ok {

			locations[fp.Location] = locationI

			locationsFromID[strconv.Itoa(locationI)] = fp.Location
			locationI++
		}
	}

	//loop through fingerprints bucket and make the svmData string
	svmData := ""
	for _, fp := range fingerprintsInMemory {
		svmData = svmData + makeSVMLine(fp, macs, locations)
		/*
			svmData: 	loc mac1:rss1 mac2:rss2 ...
				1 	1	:-52 	2 :-76 	3:-78 4:-88 5:-80
				2 	1	:-53 	2 :-69 	3:-76 4:-88 5:-80
		*/
	}

	var errs []error
	errs = append(errs, dbm.SetResourceInBucket("svmData", svmData, "svmresources", group))
	errs = append(errs, dbm.SetResourceInBucket("macsFromID", macsFromID, "svmresources", group))
	errs = append(errs, dbm.SetResourceInBucket("locationsFromID", locationsFromID, "svmresources", group))
	errs = append(errs, dbm.SetResourceInBucket("macs", macs, "svmresources", group))
	errs = append(errs, dbm.SetResourceInBucket("locations", locations, "svmresources", group))

	for _, err := range errs {
		if err != nil {
			glb.Error.Println(err)
		}
	}
}

//uses the LIBSVM to generate svm model and test for accuracy
func CalculateSVM(group string) error {
	defer glb.TimeTrack(time.Now(), "TIMEING")
	var err error
	//opening the database

	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0755, nil)
	//defer db.Close()
	// if err != nil {
	//	panic(err)
	//}
	//
	//svmData := ""
	//err = db.View(func(tx *bolt.Tx) error {
	//	// Assume bucket exists and has keys
	//	b := tx.Bucket([]byte("svmresources"))
	//	if b == nil {
	//		return fmt.Errorf("Resources dont exist")
	//	}
	//	//get svmData from db
	//	v := b.Get([]byte("svmData"))
	//	svmData = string(v)
	//	return err
	//})
	//if err != nil {
	//	panic(err)
	//}

	svmData := ""

	err = dbm.GetResourceInBucket("svmData", &svmData, "svmresources", group)
	if err != nil {
		panic(err)
	}

	if len(svmData) == 0 {
		return fmt.Errorf("No data")
	}

	lines := strings.Split(svmData, "\n")
	//rand.Perm(n) returns a random slice of length n; e.g. rand.Perm(5) = {1,3,2,0,4}
	list := rand.Perm(len(lines))
	learningSet := ""
	testingSet := ""
	fullSet := ""

	//split svmData to training and testing data set
	//todo: use libsvm subset.py tool instead
	for i := range list {
		if len(lines[list[i]]) == 0 {
			continue
		}
		if i < len(list)/2 {
			learningSet = learningSet + lines[list[i]] + "\n"
			fullSet = fullSet + lines[list[i]] + "\n"
		} else {
			testingSet = testingSet + lines[list[i]] + "\n"
			fullSet = fullSet + lines[list[i]] + "\n"
		}
	}

	tempFileFull := glb.RandStringBytesMaskImprSrc(16) + ".full"
	tempFileTrain := glb.RandStringBytesMaskImprSrc(16) + ".learning"
	tempFileTest := glb.RandStringBytesMaskImprSrc(16) + ".testing"
	tempFileOut := glb.RandStringBytesMaskImprSrc(16) + ".out"
	d1 := []byte(learningSet)
	//todo: replace random string with epoch time

	err = ioutil.WriteFile(tempFileTrain, d1, 0644)
	if err != nil {
		panic(err)
	}

	d1 = []byte(testingSet)
	err = ioutil.WriteFile(tempFileTest, d1, 0644)
	if err != nil {
		panic(err)
	}

	d1 = []byte(fullSet)
	err = ioutil.WriteFile(tempFileFull, d1, 0644)
	if err != nil {
		panic(err)
	}

	// cmd := "svm-scale"
	// args := "-l 0 -u 1 " + tempFileTrain
	// Debug.Println(cmd, args)
	// outCmd, err := exec.Command(cmd, strings.Split(args, " ")...).Output()
	// if err != nil {
	// 	panic(err)
	// }
	// err = ioutil.WriteFile(tempFileTrain+".scaled", outCmd, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// cmd = "svm-scale"
	// args = "-l 0 -u 1 " + tempFileTest
	// Debug.Println(cmd, args)
	// outCmd, err = exec.Command(cmd, strings.Split(args, " ")...).Output()
	// if err != nil {
	// 	panic(err)
	// }
	// err = ioutil.WriteFile(tempFileTest+".scaled", outCmd, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// cmd = "svm-scale"
	// args = "-l 0 -u 1 " + tempFileFull
	// Debug.Println(cmd, args)
	// outCmd, err = exec.Command(cmd, strings.Split(args, " ")...).Output()
	// if err != nil {
	// 	panic(err)
	// }
	// err = ioutil.WriteFile(tempFileFull+".scaled", outCmd, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	//todo: test some other libsvm options here to find the best model for svm. the results could be checked with the console outCmd variable log
	//todo: make a function to rename the mac of an AP in db. The solution should be compatible with bayesian solution.
	cmd := "svm-train"
	args := "-s 0 -t 0 -b 1 " + tempFileFull + " data/" + group + ".model"
	glb.Debug.Println(cmd, args)
	if _, err = exec.Command(cmd, strings.Split(args, " ")...).Output(); err != nil {
		panic(err)
	}

	cmd = "svm-train"
	args = "-s 0 -t 0 -b 1 " + tempFileTrain + " " + tempFileTrain + ".model"
	glb.Debug.Println(cmd, args)
	if _, err = exec.Command(cmd, strings.Split(args, " ")...).Output(); err != nil {
		panic(err)
	}

	cmd = "svm-predict"
	args = "-b 1 " + tempFileTest + " " + tempFileTrain + ".model " + tempFileOut
	glb.Debug.Println(cmd, args)
	outCmd, err := exec.Command(cmd, strings.Split(args, " ")...).Output()
	if err != nil {
		panic(err)
	}
	glb.Debug.Printf("%s SVM: %s", group, strings.TrimSpace(string(outCmd)))

	//os.Remove(tempFileTrain + ".scaled")
	//os.Remove(tempFileTest + ".scaled")
	//os.Remove(tempFileFull + ".scaled")
	os.Remove(tempFileTrain)
	os.Remove(tempFileTrain + ".model")
	os.Remove(tempFileTest)
	os.Remove(tempFileFull)
	os.Remove(tempFileOut)
	return nil
}

//uses the LIBSVM to predict and classify the incoming fingerprint
func SvmClassify(jsonFingerprint parameters.Fingerprint) (string, map[string]float64) {
	var err error
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, strings.ToLower(jsonFingerprint.Group)+".model")); os.IsNotExist(err) {
		return "", make(map[string]float64)
	}
	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, jsonFingerprint.Group+".db"), 0755, nil)
	//defer db.Close()
	// if err != nil {
	//	panic(err)
	//}

	var locations map[string]int
	var macs map[string]int
	var locationsFromID map[string]string
	//err = db.View(func(tx *bolt.Tx) error {
	//	// Assume bucket exists and has keys
	//	b := tx.Bucket([]byte("svmresources"))
	//	if b == nil {
	//		return fmt.Errorf("Resources dont exist")
	//	}
	//	//gets some data from db
	//	v := b.Get([]byte("locations"))
	//	json.Unmarshal(v, &locations)
	//	v = b.Get([]byte("locationsFromID"))
	//	json.Unmarshal(v, &locationsFromID)
	//	v = b.Get([]byte("macs"))
	//	json.Unmarshal(v, &macs)
	//	return err
	//})

	err = dbm.GetResourceInBucket("locations", &locations, "svmresources", jsonFingerprint.Group)
	if err != nil {
		glb.Error.Println(err)
		return "", nil
	}
	err = dbm.GetResourceInBucket("locationsFromID", &locationsFromID, "svmresources", jsonFingerprint.Group)
	if err != nil {
		glb.Error.Println(err)
		return "", nil
	}
	err = dbm.GetResourceInBucket("macs", &macs, "svmresources", jsonFingerprint.Group)
	if err != nil {
		glb.Error.Println(err)
		return "", nil
	}

	svmData := makeSVMLine(jsonFingerprint, macs, locations)
	if len(svmData) < 5 {
		glb.Warning.Println(svmData)
		return "", make(map[string]float64)
	}
	// Debug.Println(svmData)

	tempFileTest := glb.RandStringBytesMaskImprSrc(6) + ".testing"
	tempFileOut := glb.RandStringBytesMaskImprSrc(6) + ".out"
	d1 := []byte(svmData)
	err = ioutil.WriteFile(tempFileTest, d1, 0644)
	if err != nil {
		panic(err)
	}

	// cmd := "svm-scale"
	// args := "-l 0 -u 1 " + tempFileTest
	// outCmd, err := exec.Command(cmd, strings.Split(args, " ")...).Output()
	// if err != nil {
	// 	panic(err)
	// }
	// err = ioutil.WriteFile(tempFileTest+".scaled", outCmd, 0644)
	// if err != nil {
	// 	panic(err)
	// }

	cmd := "svm-predict"
	args := "-b 1 " + tempFileTest + " data/" + jsonFingerprint.Group + ".model " + tempFileOut
	_, err = exec.Command(cmd, strings.Split(args, " ")...).Output()
	if err != nil {
		panic(err)
	}

	dat, err := ioutil.ReadFile(tempFileOut)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(dat), "\n")
	labels := strings.Split(lines[0], " ")
	probabilities := strings.Split(lines[1], " ")
	P := make(map[string]float64)
	bestLocation := ""
	bestP := float64(0)
	for i := range labels {
		if i == 0 {
			continue
		}
		Pval, _ := strconv.ParseFloat(probabilities[i], 64)
		//if the best probability is more than the current probability, then change best location.
		if Pval > bestP {
			bestLocation = locationsFromID[labels[i]]
			bestP = Pval
		}
		P[locationsFromID[labels[i]]] = math.Log(float64(Pval))
	}
	os.Remove(tempFileTest)
	// os.Remove(tempFileTest + ".scaled")
	os.Remove(tempFileOut)
	// Debug.Println(P)
	return bestLocation, P
}

func makeSVMLine(v2 parameters.Fingerprint, macs map[string]int, locations map[string]int) string {
	m := make(map[int]int)
	for _, fingerprint := range v2.WifiFingerprint {
		if _, ok := macs[fingerprint.Mac]; ok && fingerprint.Rssi > glb.MinRssiOpt {
			m[macs[fingerprint.Mac]] = fingerprint.Rssi
		}
	}
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	svmData := ""
	// for i := 0; i < 3; i++ {
	if _, ok := locations[v2.Location]; ok {
		svmData = svmData + strconv.Itoa(locations[v2.Location]) + " "
	} else {
		svmData = svmData + "1 "
	}
	for _, k := range keys {
		svmData = svmData + strconv.Itoa(k) + ":" + strconv.Itoa(m[k]) + " "
	}
	svmData = svmData + "\n"
	// }

	return svmData

}

// cp ~/Documents/find/svm ./
// cat svm | shuf > svm.shuffled
// ./svm-scale -l 0 -u 1 svm.shuffled > svm.shuffled.scaled
// head -n 500 svm.shuffled.scaled > learning
// tail -n 1500 svm.shuffled.scaled > testing
// ./svm-train -s 0 -t 0 -b 1 learning > /dev/null
// ./svm-predict -b 1 testing learning.model out
