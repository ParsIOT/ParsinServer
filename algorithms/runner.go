package algorithms

import (
	"ParsinServer/glb"
	"time"
	"strings"

	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"ParsinServer/algorithms/bayes"
	"math"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/dbm"
)


// track api that calls trackFingerprint() function
func TrackFingerprintPOST(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	var jsonFingerprint parameters.Fingerprint
	//glb.Info.Println(jsonFingerprint)

	if glb.BindJSON(&jsonFingerprint, c) == nil {
		if (len(jsonFingerprint.WifiFingerprint) >= glb.MinApNum) {
			glb.Debug.Println("Track json: ",jsonFingerprint)
			message, success, bayesGuess, _, svmGuess, _, knnGuess, scikitData  := TrackFingerprint(jsonFingerprint)
			if success {
				scikitDataStr := glb.StringMap2String(scikitData)
				resJsonMap := gin.H{"message": message, "success": true, "bayes": bayesGuess, "svm": svmGuess, "knn": knnGuess}
				for algorithm, valXY := range scikitData{
					resJsonMap[algorithm]=valXY
				}

				glb.Debug.Println("message", message, " success", true, " bayes", bayesGuess, " svm", svmGuess, scikitDataStr, " knn", knnGuess)
				c.JSON(http.StatusOK, resJsonMap)
			} else {
				glb.Debug.Println(message)
				c.JSON(http.StatusOK, gin.H{"message": message, "success": false})
			}
		} else {
			glb.Warning.Println("Nums of AP must be greater than 3")
			c.JSON(http.StatusOK, gin.H{"message": "Nums of AP must be greater than 3", "success": false})
		}
	} else {
		glb.Warning.Println("Could not bind JSON")
		c.JSON(http.StatusOK, gin.H{"message": "Could not bind JSON", "success": false})
	}
}

// call leanFingerprint() function
func LearnFingerprintPOST(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	var jsonFingerprint parameters.Fingerprint
	glb.Info.Println(jsonFingerprint)
	if glb.BindJSON(&jsonFingerprint, c) == nil {
		message, success := LearnFingerprint(jsonFingerprint)
		glb.Debug.Println(message)
		if !success {
			glb.Debug.Println(jsonFingerprint)
		}
		c.JSON(http.StatusOK, gin.H{"message": message, "success": success})
	} else {
		glb.Warning.Println("Could not bind JSON")
		c.JSON(http.StatusOK, gin.H{"message": "Could not bind JSON", "success": false})
	}
}

// call leanFingerprint() function for each fp of BulkFingerprint
func BulkLearnFingerprintPOST(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	var bulkJsonFingerprint parameters.BulkFingerprint
	var returnMessage string
	var returnSuccess string
	if glb.BindJSON(&bulkJsonFingerprint, c) == nil {
		glb.Debug.Println("BulkFingerPrints:")
		glb.Debug.Println(bulkJsonFingerprint)
		for i, jsonFingerprint := range bulkJsonFingerprint.Fingerprints {
			message, success := LearnFingerprint(jsonFingerprint)
			glb.Debug.Println(i, " th fingerprint saving process: ", message)
			if success {
				glb.Debug.Println(i, " th fingerprint data: ", jsonFingerprint)
				returnSuccess = returnSuccess + strconv.Itoa(i) + " th success: true\n"
				returnMessage = returnMessage + strconv.Itoa(i) + " th message: " + message + "\n"
			} else {
				returnSuccess = returnSuccess + strconv.Itoa(i) + " th success: false\n"
				returnMessage = returnMessage + strconv.Itoa(i) + " th message: " + message + "\n"
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": returnMessage, "success": returnSuccess})
	} else {
		glb.Warning.Println("Could not bind to BulkFingerprint")
		c.JSON(http.StatusOK, gin.H{"message": "Could not bind JSON", "success": false})
	}
}

// cleanFingerPrint and save the Fingerprint to db
func LearnFingerprint(jsonFingerprint parameters.Fingerprint) (string, bool) {
	parameters.CleanFingerprint(&jsonFingerprint)
	glb.Info.Println(jsonFingerprint)
	if len(jsonFingerprint.Group) == 0 {
		return "Need to define your group name in request, see API", false
	}
	if len(jsonFingerprint.WifiFingerprint) == 0 {
		return "No fingerprints found to insert, see API", false
	}
	dbm.PutFingerprintIntoDatabase(jsonFingerprint, "fingerprints")
	go dbm.SetLearningCache(strings.ToLower(jsonFingerprint.Group), true)
	message := "Inserted fingerprint containing " + strconv.Itoa(len(jsonFingerprint.WifiFingerprint)) + " APs for " + jsonFingerprint.Username + " (" + jsonFingerprint.Group + ") at " + jsonFingerprint.Location
	return message, true
}

// call leanFingerprint(),calculateSVM() and rfLearn() functions after that call prediction functions and return the estimation location
func TrackFingerprint(jsonFingerprint parameters.Fingerprint) (string, bool, string, map[string]float64, string, map[string]float64, string, map[string]string) {
	// Classify with filter fingerprint
	fullFingerprint := jsonFingerprint
	dbm.FilterFingerprint(&jsonFingerprint)

	bayesGuess := ""
	bayesData := make(map[string]float64)
	svmGuess := ""
	svmData := make(map[string]float64)
	scikitData := make(map[string]string)
	knnGuess := ""

	parameters.CleanFingerprint(&jsonFingerprint)
	if !dbm.GroupExists(jsonFingerprint.Group) || len(jsonFingerprint.Group) == 0 {
		return "You should insert fingerprints before tracking", false, "", bayesData, "", make(map[string]float64), "", make(map[string]string)
	}
	if len(jsonFingerprint.WifiFingerprint) == 0 {
		return "No fingerprints found to track, see API", false, "", bayesData, "", make(map[string]float64), "", make(map[string]string)
	}
	if len(jsonFingerprint.Username) == 0 {
		return "No username defined, see API", false, "", bayesData, "", make(map[string]float64), "", make(map[string]string)
	}
	wasLearning, ok := dbm.GetLearningCache(strings.ToLower(jsonFingerprint.Group))
	if ok {
		if wasLearning {
			glb.Debug.Println("Was learning, calculating priors")
			group := strings.ToLower(jsonFingerprint.Group)
			go dbm.SetLearningCache(group, false)
			bayes.OptimizePriorsThreaded(group)
			if glb.RuntimeArgs.Svm {
				DumpFingerprintsSVM(group)
				CalculateSVM(group)
			}
			if glb.RuntimeArgs.Scikit {
				ScikitLearn(group)
			}
			LearnKnn(group)
			go dbm.AppendUserCache(group, jsonFingerprint.Username)
		}
	}
	glb.Info.Println(jsonFingerprint)
	bayesGuess, bayesData = bayes.CalculatePosterior(jsonFingerprint, nil)
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
	go dbm.PutFingerprintIntoDatabase(fullFingerprint, "fingerprints-track")

	message := ""
	glb.Debug.Println("Tracking fingerprint containing " + strconv.Itoa(len(jsonFingerprint.WifiFingerprint)) + " APs for " + jsonFingerprint.Username + " (" + jsonFingerprint.Group + ") at " + jsonFingerprint.Location + " (guess)")
	message += " BayesGuess: " + bayesGuess //+ " (" + strconv.Itoa(int(percentGuess1)) + "% confidence)"

	// Process SVM if needed
	//if glb.RuntimeArgs.Svm {
	//	svmGuess, svmData := SvmClassify(jsonFingerprint)
	//	percentSvmGuess := int(100 * math.Exp(svmData[svmGuess]))
	//	if percentSvmGuess > 100 {
	//		//todo: wtf? \/ \/ why is could be more than 100
	//		percentSvmGuess = percentSvmGuess / 10
	//	}
	//	message += " svmGuess: " + svmGuess
	//	//message = "NB: " + locationGuess1 + " (" + strconv.Itoa(int(percentGuess1)) + "%)" + ", SVM: " + locationGuess2 + " (" + strconv.Itoa(int(percentGuess2)) + "%)"
	//}

	// Calculating KNN
	err, knnGuess := TrackKnn(jsonFingerprint)
	if err != nil {
		glb.Error.Println(err)
	}
	message += " knnGuess: " + knnGuess

	// Calculating RF

	if glb.RuntimeArgs.Scikit {
		scikitData = ScikitClassify(strings.ToLower(jsonFingerprint.Group), jsonFingerprint)
		glb.Debug.Println(scikitData)
		for algorithm, valueXY := range scikitData{
			message += " "+algorithm+":v" + valueXY
		}

	}

	// Send out the final responses
	var userJSON glb.UserPositionJSON
	userJSON.Time = time.Now().String()
	userJSON.BayesGuess = bayesGuess
	userJSON.BayesData = bayesData
	userJSON.SvmGuess = svmGuess
	userJSON.SvmData = svmData
	userJSON.ScikitData = scikitData
	userJSON.KnnGuess = knnGuess

	go dbm.SetUserPositionCache(strings.ToLower(jsonFingerprint.Group)+strings.ToLower(jsonFingerprint.Username), userJSON)

	// Send MQTT if needed
	if glb.RuntimeArgs.Mqtt {
		type FingerprintResponse struct {
			Timestamp  int64  `json:"time"`
			BayesGuess string `json:"bayesguess"`
			SvmGuess   string `json:"svmguess"`
			ScikitData    map[string]string `json:"scikitdata"`
			KnnGuess   string `json:"knnguess"`
		}
		//mqttMessage, _ := json.Marshal(FingerprintResponse{
		//	Timestamp:  time.Now().UnixNano(),
		//	BayesGuess: bayesGuess,
		//	SvmGuess:   svmGuess,
		//	ScikitData:    scikitData,
		//	KnnGuess:   knnGuess,
		//})
		//go routes.SendMQTTLocation(string(mqttMessage), jsonFingerprint.Group, jsonFingerprint.Username)
	}



	return message, true, bayesGuess, bayesData, svmGuess, svmData, knnGuess, scikitData

}

