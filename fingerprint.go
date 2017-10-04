// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// fingerprint.go contains structures and functions for handling fingerprints.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	"net/http"
	"path"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	//"google.golang.org/genproto/googleapis/api/serviceconfig"
)

//Fingerprint is the prototypical information from the fingerprinting device
//IF you change Fingerprint, follow these steps to re-generate fingerprint_ffjson.go
//find ./ -name "*.go" -type f | xargs sed -i  's/package main/package main/g'
//Uncomment json.Marshal/Unmarshal functions
//$GOPATH/bin/ffjson fingerprint.go
//find ./ -name "*.go" -type f | xargs sed -i  's/package main/package main/g'
//Comment json.Marshal/Unmarshal functions
type Fingerprint struct {
	Group           string   `json:"group"`
	Username        string   `json:"username"`
	Location        string   `json:"location"`
	Timestamp       int64    `json:"timestamp"`
	WifiFingerprint []Router `json:"wifi-fingerprint"`
}

type BulkFingerprint struct {
	Fingerprints []Fingerprint    `json:"fingerprints"`
}

// Router is the router information for each individual mac address
type Router struct {
	Mac  string `json:"mac"`
	Rssi int    `json:"rssi"`
}

var jsonExample = `{
	"group": "whatevergroup",
	"username": "iamauser",
	"location": null,
	"wififingerprint": [{
		"mac": "AA:AA:AA:AA:AA:AA",
		"rssi": -45
	}, {
		"mac": "BB:BB:BB:BB:BB:BB",
		"rssi": -55
	}]
}`

// compression 9 us -> 900 us
// Marsahal and compress a fingerprint
func dumpFingerprint(res Fingerprint) []byte {
	dumped, _ := res.MarshalJSON()
	//dumped, _ := json.Marshal(res)
	return compressByte(dumped)
}

// compression 30 us -> 600 us
//loadFingerprint returns a fingerprint from given jsonByte input
func loadFingerprint(jsonByte []byte) Fingerprint {
	res := Fingerprint{}
	//json.Unmarshal(decompressByte(jsonByte), res)
	res.UnmarshalJSON(decompressByte(jsonByte))
	filterFingerprint(&res)
	return res
}

//returns the filtered macs from macs.json file and remove the other macs from fingerprint
func filterFingerprint(res *Fingerprint) {
	if RuntimeArgs.Filtering {
		newFingerprint := make([]Router, len(res.WifiFingerprint))
		curNum := 0
		for i := range res.WifiFingerprint {
			if ok2, ok := RuntimeArgs.FilterMacs[res.WifiFingerprint[i].Mac]; ok && ok2 {
				newFingerprint[curNum] = res.WifiFingerprint[i]
				//todo: why "0" is added at the end?
				newFingerprint[curNum].Mac = newFingerprint[curNum].Mac[0:len(newFingerprint[curNum].Mac)-1] + "0"
				curNum++
			}
		}
		newFingerprint = newFingerprint[0:curNum]
		res.WifiFingerprint = newFingerprint
	}
}

// convert quality (0 to 100) to rss(-100 to -50) and delete the records that their mac are "00:00:00:00:00"
func cleanFingerprint(res *Fingerprint) {
	res.Group = strings.TrimSpace(strings.ToLower(res.Group))
	res.Location = strings.TrimSpace(strings.ToLower(res.Location))
	res.Username = strings.TrimSpace(strings.ToLower(res.Username))
	deleteIndex := -1
	for r := range res.WifiFingerprint {
		if res.WifiFingerprint[r].Rssi >= 0 { // https://stackoverflow.com/questions/15797920/how-to-convert-wifi-signal-strength-from-quality-percent-to-rssi-dbm
			res.WifiFingerprint[r].Rssi = int(res.WifiFingerprint[r].Rssi/2) - 100
		}
		if res.WifiFingerprint[r].Mac == "00:00:00:00:00:00" {
			deleteIndex = r
		}
	}
	// delete res.WifiFingerprint[deleteIndex]
	if deleteIndex > -1 {
		res.WifiFingerprint[deleteIndex] = res.WifiFingerprint[len(res.WifiFingerprint)-1]
		res.WifiFingerprint = res.WifiFingerprint[:len(res.WifiFingerprint)-1]
	}
}

// make a db according to group name
func putFingerprintIntoDatabase(res Fingerprint, database string) error {
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, res.Group+".db"), 0600, nil)
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
		err2 = bucket.Put([]byte(strconv.FormatInt(res.Timestamp, 10)), dumpFingerprint(res))
		if err2 != nil {
			return fmt.Errorf("could add to bucket: %s", err2)
		}
		return err2
	})
	db.Close()
	return err
}

// track api that calls trackFingerprint() function
func trackFingerprintPOST(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	var jsonFingerprint Fingerprint
	//Info.Println(jsonFingerprint)

	if c.BindJSON(&jsonFingerprint) == nil {
		if (len(jsonFingerprint.WifiFingerprint) >= 3) {
			message, success, bayesGuess, _, svmGuess, _, rfGuess, _, knnGuess := trackFingerprint(jsonFingerprint)
			if success {
				c.JSON(http.StatusOK, gin.H{"message": message, "success": true, "bayes": bayesGuess, "svm": svmGuess, "rf": rfGuess, "knn": knnGuess})
			} else {
				c.JSON(http.StatusOK, gin.H{"message": message, "success": false})
			}
		} else {
			Warning.Println("Nums of AP must be greater than 3")
			c.JSON(http.StatusOK, gin.H{"message": "Nums of AP must be greater than 3", "success": false})
		}
	} else {
		Warning.Println("Could not bind JSON")
		c.JSON(http.StatusOK, gin.H{"message": "Could not bind JSON", "success": false})
	}
}

// call leanFingerprint() function
func learnFingerprintPOST(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	var jsonFingerprint Fingerprint
	if c.BindJSON(&jsonFingerprint) == nil {
		message, success := learnFingerprint(jsonFingerprint)
		Debug.Println(message)
		if !success {
			Debug.Println(jsonFingerprint)
		}
		c.JSON(http.StatusOK, gin.H{"message": message, "success": success})
	} else {
		Warning.Println("Could not bind JSON")
		c.JSON(http.StatusOK, gin.H{"message": "Could not bind JSON", "success": false})
	}
}

// call leanFingerprint() function for each fp of BulkFingerprint
func bulkLearnFingerprintPOST(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	var bulkJsonFingerprint BulkFingerprint

	var returnMessage string
	var returnSuccess string
	if c.BindJSON(&bulkJsonFingerprint) == nil {
		for i, jsonFingerprint := range bulkJsonFingerprint.Fingerprints {
			message, success := learnFingerprint(jsonFingerprint)
			Debug.Println(i, " th fingerprint saving process: ", message)
			if success {
				Debug.Println(i, " th fingerprint data: ", jsonFingerprint)
				returnSuccess = returnSuccess + strconv.Itoa(i) + " th success: true\n"
				returnMessage = returnMessage + strconv.Itoa(i) + " th message: " + message + "\n"
			} else {
				returnSuccess = returnSuccess + strconv.Itoa(i) + " th success: false\n"
				returnMessage = returnMessage + strconv.Itoa(i) + " th message: " + message + "\n"
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": returnMessage, "success": returnSuccess})
	} else {
		Warning.Println("Could not bind to BulkFingerprint")
		c.JSON(http.StatusOK, gin.H{"message": "Could not bind JSON", "success": false})
	}
}



// cleanFingerPrint and save the Fingerprint to db
func learnFingerprint(jsonFingerprint Fingerprint) (string, bool) {
	cleanFingerprint(&jsonFingerprint)
	Info.Println(jsonFingerprint)
	if len(jsonFingerprint.Group) == 0 {
		return "Need to define your group name in request, see API", false
	}
	if len(jsonFingerprint.WifiFingerprint) == 0 {
		return "No fingerprints found to insert, see API", false
	}
	putFingerprintIntoDatabase(jsonFingerprint, "fingerprints")
	go setLearningCache(strings.ToLower(jsonFingerprint.Group), true)
	message := "Inserted fingerprint containing " + strconv.Itoa(len(jsonFingerprint.WifiFingerprint)) + " APs for " + jsonFingerprint.Username + " (" + jsonFingerprint.Group + ") at " + jsonFingerprint.Location
	return message, true
}

// call leanFingerprint(),calculateSVM() and rfLearn() functions after that call prediction functions and return the estimation location
func trackFingerprint(jsonFingerprint Fingerprint) (string, bool, string, map[string]float64, string, map[string]float64, string, map[string]float64, string) {
	// Classify with filter fingerprint
	fullFingerprint := jsonFingerprint
	filterFingerprint(&jsonFingerprint)

	bayesGuess := ""
	bayesData := make(map[string]float64)
	svmGuess := ""
	svmData := make(map[string]float64)
	rfGuess := ""
	rfData := make(map[string]float64)
	knnGuess := ""

	cleanFingerprint(&jsonFingerprint)
	if !groupExists(jsonFingerprint.Group) || len(jsonFingerprint.Group) == 0 {
		return "You should insert fingerprints before tracking", false, "", bayesData, "", make(map[string]float64), "", make(map[string]float64), ""
	}
	if len(jsonFingerprint.WifiFingerprint) == 0 {
		return "No fingerprints found to track, see API", false, "", bayesData, "", make(map[string]float64), "", make(map[string]float64), ""
	}
	if len(jsonFingerprint.Username) == 0 {
		return "No username defined, see API", false, "", bayesData, "", make(map[string]float64), "", make(map[string]float64), ""
	}
	wasLearning, ok := getLearningCache(strings.ToLower(jsonFingerprint.Group))
	if ok {
		if wasLearning {
			Debug.Println("Was learning, calculating priors")
			group := strings.ToLower(jsonFingerprint.Group)
			go setLearningCache(group, false)
			optimizePriorsThreaded(group)
			if RuntimeArgs.Svm {
				dumpFingerprintsSVM(group)
				calculateSVM(group)
			}
			if RuntimeArgs.RandomForests {
				rfLearn(group)
			}
			go appendUserCache(group, jsonFingerprint.Username)
		}
	}
	Info.Println(jsonFingerprint)
	bayesGuess, bayesData = calculatePosterior(jsonFingerprint, *NewFullParameters())
	percentBayesGuess := float64(0)
	total := float64(0)
	for _, locBayes := range bayesData {
		total += math.Exp(locBayes)
		if locBayes > percentBayesGuess {
			percentBayesGuess = locBayes
		}
	}
	percentBayesGuess = math.Exp(bayesData[bayesGuess]) / total * 100.0

	// todo: add abitlity to save rf, knn, svm guess
	jsonFingerprint.Location = bayesGuess

	// Insert full fingerprint
	putFingerprintIntoDatabase(fullFingerprint, "fingerprints-track")

	message := ""
	Debug.Println("Tracking fingerprint containing " + strconv.Itoa(len(jsonFingerprint.WifiFingerprint)) + " APs for " + jsonFingerprint.Username + " (" + jsonFingerprint.Group + ") at " + jsonFingerprint.Location + " (guess)")
	message += " BayesGuess: " + bayesGuess //+ " (" + strconv.Itoa(int(percentGuess1)) + "% confidence)"

	// Process SVM if needed
	if RuntimeArgs.Svm {
		svmGuess, svmData := classify(jsonFingerprint)
		percentSvmGuess := int(100 * math.Exp(svmData[svmGuess]))
		if percentSvmGuess > 100 {
			//todo: wtf? \/ \/ why is could be more than 100
			percentSvmGuess = percentSvmGuess / 10
		}
		message += " svmGuess: " + svmGuess
		//message = "NB: " + locationGuess1 + " (" + strconv.Itoa(int(percentGuess1)) + "%)" + ", SVM: " + locationGuess2 + " (" + strconv.Itoa(int(percentGuess2)) + "%)"
	}

	// Calculating KNN
	err, knnGuess := calculateKnn(jsonFingerprint)
	if err != nil {
		Error.Println(err)
	}
	message += " knnGuess: " + knnGuess

	// Calculating RF

	if RuntimeArgs.RandomForests {
		rfGuess, rfData = rfClassify(strings.ToLower(jsonFingerprint.Group), jsonFingerprint)
		message += " rfGuess: " + rfGuess
	}

	// Send out the final responses
	var userJSON UserPositionJSON
	userJSON.Time = time.Now().String()
	userJSON.BayesGuess = bayesGuess
	userJSON.BayesData = bayesData
	userJSON.SvmGuess = svmGuess
	userJSON.SvmData = svmData
	userJSON.RfGuess = rfGuess
	userJSON.RfData = rfData
	userJSON.KnnGuess = knnGuess

	go setUserPositionCache(strings.ToLower(jsonFingerprint.Group)+strings.ToLower(jsonFingerprint.Username), userJSON)

	// Send MQTT if needed
	if RuntimeArgs.Mqtt {
		type FingerprintResponse struct {
			Timestamp  int64  `json:"time"`
			BayesGuess string `json:"bayesguess"`
			SvmGuess   string `json:"svmguess"`
			RfGuess    string `json:"rfguess"`
			KnnGuess   string `json:"knnguess"`
		}
		mqttMessage, _ := json.Marshal(FingerprintResponse{
			Timestamp:  time.Now().UnixNano(),
			BayesGuess: bayesGuess,
			SvmGuess:   svmGuess,
			RfGuess:    rfGuess,
			KnnGuess:   knnGuess,
		})
		go sendMQTTLocation(string(mqttMessage), jsonFingerprint.Group, jsonFingerprint.Username)
	}

	return message, true, bayesGuess, bayesData, svmGuess, svmData, rfGuess, rfData, knnGuess

}