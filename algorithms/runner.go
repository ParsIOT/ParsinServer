package algorithms

import (
	"ParsinServer/algorithms/clustering"
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"errors"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

const (
	AUC              string = "AUC"
	FalseRate        string = "FalseRate"
	Mean             string = "Mean"
	LatterPercentile string = "LatterPercentile"
)

var threadedCross bool = false //don't use, it's not safe now!

type KnnJob struct {
	gp                       *dbm.Group
	K                        int
	MinClusterRss            int
	crossValidationPartsList []crossValidationParts
}

type KnnJobResult struct {
	TotalError            int
	KnnErrHyperParameters []interface{}
}

// track api that calls trackFingerprint() function
func TrackFingerprintPOST(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	var fp parameters.Fingerprint
	//glb.Info.Println(fp)

	if glb.BindJSON(&fp, c) == nil {
		if (len(fp.WifiFingerprint) >= glb.MinApNum) {
			glb.Debug.Println("Track json: ", fp)
			if !glb.IsValidXY(fp.Location) {
				// Mobile PDR data isn't available
				fp.Location = "" // this value check in TrackFingerprint, non-empty fp.Location is considered as PDR data
			}
			message, success, location, bayesGuess, _, svmGuess, _, knnGuess, scikitData, accuracyCircleRadius := TrackFingerprint(fp)

			knnGuess = location // todo: this is done for compatability ,use 'location' instead of knn
			if success {
				scikitDataStr := glb.StringMap2String(scikitData)
				resJsonMap := gin.H{"message": message, "success": true, "location": location, "bayes": bayesGuess, "svm": svmGuess, "knn": knnGuess, "accuracyCircleRadius": accuracyCircleRadius}
				for algorithm, valXY := range scikitData {
					resJsonMap[algorithm] = valXY
				}

				glb.Debug.Println("message", message, " success", true, " location", location, " bayes", bayesGuess, " svm", svmGuess, scikitDataStr, " knn", knnGuess, " accuracyCircleRadius", accuracyCircleRadius)
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
	//glb.Debug.Println(c)
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
func TrackFingerprint(curFingerprint parameters.Fingerprint) (string, bool, string, string, map[string]float64, string, map[string]float64, string, map[string]string, float64) {
	// Classify with filter curFingerprint
	//fullFingerprint := curFingerprint

	dbm.FilterFingerprint(&curFingerprint)

	groupName := strings.ToLower(curFingerprint.Group)

	bayesGuess := ""
	bayesData := make(map[string]float64)
	svmGuess := ""
	svmData := make(map[string]float64)
	scikitData := make(map[string]string)
	knnGuess := ""
	message := ""
	location := ""
	pdrLocation := curFingerprint.Location
	accuracyCircleRadius := float64(0)

	parameters.CleanFingerprint(&curFingerprint)
	if !dbm.GroupExists(curFingerprint.Group) || len(curFingerprint.Group) == 0 {
		return "You should insert fingerprints before tracking", false, "", "", bayesData, "", make(map[string]float64), "", make(map[string]string), float64(0)
	}
	if len(curFingerprint.WifiFingerprint) == 0 {
		return "No fingerprints found to track, see API", false, "", "", bayesData, "", make(map[string]float64), "", make(map[string]string), float64(0)
	}
	if len(curFingerprint.Username) == 0 {
		return "No username defined, see API", false, "", "", bayesData, "", make(map[string]float64), "", make(map[string]string), float64(0)
	}

	wasLearning, ok := dbm.GetLearningCache(strings.ToLower(curFingerprint.Group))
	if ok {
		if wasLearning {
			glb.Debug.Println("Was learning, calculating priors")

			go dbm.SetLearningCache(groupName, false)
			//bayes.OptimizePriorsThreaded(groupName)
			//if glb.RuntimeArgs.Svm {
			//	DumpFingerprintsSVM(groupName)
			//	CalculateSVM(groupName)
			//}
			//if glb.RuntimeArgs.Scikit {
			//	ScikitLearn(groupName)
			//}
			//LearnKnn(groupName)
			CalculateLearn(groupName)
			go dbm.AppendUserCache(groupName, curFingerprint.Username)
		}
	}
	//glb.Info.Println(curFingerprint)
	//bayesGuess, bayesData = bayes.CalculatePosterior(curFingerprint, nil)
	//percentBayesGuess := float64(0)
	//total := float64(0)
	//for _, locBayes := range bayesData {
	//	total += math.Exp(locBayes)
	//	if locBayes > percentBayesGuess {
	//		percentBayesGuess = locBayes
	//	}
	//}
	//percentBayesGuess = math.Exp(bayesData[bayesGuess]) / total * 100.0

	//// todo: add abitlity to save rf, knn, svm guess
	//curFingerprint.Location = bayesGuess

	//message := ""
	//glb.Debug.Println("Tracking curFingerprint containing " + strconv.Itoa(len(curFingerprint.WifiFingerprint)) + " APs for " + curFingerprint.Username + " (" + curFingerprint.Group + ") at " + curFingerprint.Location + " (guess)")
	//message += " BayesGuess: " + bayesGuess //+ " (" + strconv.Itoa(int(percentGuess1)) + "% confidence)"
	//
	//// Process SVM if needed
	//if glb.RuntimeArgs.Svm {
	//	svmGuess, svmData := SvmClassify(curFingerprint)
	//	percentSvmGuess := int(100 * math.Exp(svmData[svmGuess]))
	//	if percentSvmGuess > 100 {
	//		//todo: wtf? \/ \/ why is could be more than 100
	//		percentSvmGuess = percentSvmGuess / 10
	//	}
	//	message += " svmGuess: " + svmGuess
	//	//message = "NB: " + locationGuess1 + " (" + strconv.Itoa(int(percentGuess1)) + "%)" + ", SVM: " + locationGuess2 + " (" + strconv.Itoa(int(percentGuess2)) + "%)"
	//}

	gp := dbm.GM.GetGroup(groupName)
	// Calculating KNN
	//glb.Debug.Println(curFingerprint)
	err, knnGuess, knnData := TrackKnn(gp, curFingerprint, true)
	if err != nil {
		glb.Error.Println(err)
	}
	message += " knnGuess: " + knnGuess

	//curFingerprint.Location = knnGuess
	// Insert full curFingerprint
	//glb.Debug.Println(curFingerprint)
	//go dbm.PutFingerprintIntoDatabase(fullFingerprint, "fingerprints-track")

	// Calculating Scikit
	if glb.RuntimeArgs.Scikit {
		scikitData = ScikitClassify(groupName, curFingerprint)
		glb.Debug.Println(scikitData)
		for algorithm, valueXY := range scikitData {
			message += " " + algorithm + ":" + valueXY
		}

	}

	// Send out the final responses
	var userJSON parameters.UserPositionJSON
	userJSON.Time = curFingerprint.Timestamp
	userJSON.BayesGuess = bayesGuess
	userJSON.BayesData = bayesData
	userJSON.SvmGuess = svmGuess
	userJSON.SvmData = svmData
	userJSON.ScikitData = scikitData
	userJSON.KnnGuess = knnGuess
	userJSON.KnnData = knnData
	userJSON.PDRLocation = pdrLocation
	userJSON.Fingerprint = curFingerprint

	// User history effect
	//location = knnGuess
	userHistory := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)
	location, accuracyCircleRadius = HistoryEffect(userJSON, userHistory)

	if glb.GraphEnabled {
		location = knnGuess
	}
	userJSON.Location = location
	glb.Debug.Println("Knn guess: ", knnGuess)
	glb.Debug.Println("location: ", location)

	//location = userJSON.KnnGuess
	userJSON.KnnGuess = location //todo: must add location as seprated variable from knnguess in parameters.UserPositionJSON
	go dbm.SetUserPositionCache(strings.ToLower(curFingerprint.Group)+strings.ToLower(curFingerprint.Username), userJSON)
	go gp.Get_ResultData().Append_UserHistory(curFingerprint.Username, userJSON)
	go gp.Get_ResultData().Append_UserResults(curFingerprint.Username, userJSON)

	if curFingerprint.TestValidation && curFingerprint.Username == glb.TesterUsername {
		glb.Debug.Println("TestValidTrack added ")
		tempTestValidTrack := parameters.TestValidTrack{UserPosition: userJSON}
		go gp.Get_ResultData().Append_TestValidTracks(tempTestValidTrack)
	}

	//glb.Debug.Println(len(gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)))

	// Send MQTT if needed
	if glb.RuntimeArgs.Mqtt {
		type FingerprintResponse struct {
			Timestamp  int64             `json:"time"`
			BayesGuess string            `json:"bayesguess"`
			SvmGuess   string            `json:"svmguess"`
			ScikitData map[string]string `json:"scikitdata"`
			KnnGuess   string            `json:"knnguess"`
		}
		//mqttMessage, _ := json.Marshal(FingerprintResponse{
		//	Timestamp:  time.Now().UnixNano(),
		//	BayesGuess: bayesGuess,
		//	SvmGuess:   svmGuess,
		//	ScikitData:    scikitData,
		//	KnnGuess:   knnGuess,
		//})
		//go routes.SendMQTTLocation(string(mqttMessage), curFingerprint.Group, curFingerprint.Username)
	}

	glb.Debug.Println(userJSON)

	return message, true, location, bayesGuess, bayesData, svmGuess, svmData, knnGuess, scikitData, accuracyCircleRadius

}

// call leanFingerprint(),calculateSVM() and rfLearn() functions after that call prediction functions and return the estimation location
func RecalculateTrackFingerprint(curFingerprint parameters.Fingerprint) parameters.UserPositionJSON {
	// Classify with filter curFingerprint
	//fullFingerprint := curFingerprint

	dbm.FilterFingerprint(&curFingerprint)

	groupName := strings.ToLower(curFingerprint.Group)

	bayesGuess := ""
	bayesData := make(map[string]float64)
	svmGuess := ""
	svmData := make(map[string]float64)
	scikitData := make(map[string]string)
	knnGuess := ""
	location := ""
	pdrLocation := curFingerprint.Location
	accuracyCircleRadius := float64(0)
	glb.Debug.Println("accuracy circle:", accuracyCircleRadius)
	//parameters.CleanFingerprint(&curFingerprint)

	/*	wasLearning, ok := dbm.GetLearningCache(strings.ToLower(curFingerprint.Group))
		if ok {
			if wasLearning {
				glb.Debug.Println("Was learning, calculating priors")

				go dbm.SetLearningCache(groupName, false)
				//bayes.OptimizePriorsThreaded(groupName)
				//if glb.RuntimeArgs.Svm {
				//	DumpFingerprintsSVM(groupName)
				//	CalculateSVM(groupName)
				//}
				//if glb.RuntimeArgs.Scikit {
				//	ScikitLearn(groupName)
				//}
				//LearnKnn(groupName)
				CalculateLearn(groupName)
				go dbm.AppendUserCache(groupName, curFingerprint.Username)
			}
		}*/
	//glb.Info.Println(curFingerprint)
	//bayesGuess, bayesData = bayes.CalculatePosterior(curFingerprint, nil)
	//percentBayesGuess := float64(0)
	//total := float64(0)
	//for _, locBayes := range bayesData {
	//	total += math.Exp(locBayes)
	//	if locBayes > percentBayesGuess {
	//		percentBayesGuess = locBayes
	//	}
	//}
	//percentBayesGuess = math.Exp(bayesData[bayesGuess]) / total * 100.0

	//// todo: add abitlity to save rf, knn, svm guess
	//curFingerprint.Location = bayesGuess

	//message := ""
	//glb.Debug.Println("Tracking curFingerprint containing " + strconv.Itoa(len(curFingerprint.WifiFingerprint)) + " APs for " + curFingerprint.Username + " (" + curFingerprint.Group + ") at " + curFingerprint.Location + " (guess)")
	//message += " BayesGuess: " + bayesGuess //+ " (" + strconv.Itoa(int(percentGuess1)) + "% confidence)"
	//
	//// Process SVM if needed
	//if glb.RuntimeArgs.Svm {
	//	svmGuess, svmData := SvmClassify(curFingerprint)
	//	percentSvmGuess := int(100 * math.Exp(svmData[svmGuess]))
	//	if percentSvmGuess > 100 {
	//		//todo: wtf? \/ \/ why is could be more than 100
	//		percentSvmGuess = percentSvmGuess / 10
	//	}
	//	message += " svmGuess: " + svmGuess
	//	//message = "NB: " + locationGuess1 + " (" + strconv.Itoa(int(percentGuess1)) + "%)" + ", SVM: " + locationGuess2 + " (" + strconv.Itoa(int(percentGuess2)) + "%)"
	//}

	gp := dbm.GM.GetGroup(groupName)
	// Calculating KNN
	//glb.Debug.Println(curFingerprint)
	err, knnGuess, knnData := TrackKnn(gp, curFingerprint, true)
	if err != nil {
		glb.Error.Println(err)
	}

	//curFingerprint.Location = knnGuess
	// Insert full curFingerprint
	//glb.Debug.Println(curFingerprint)
	//go dbm.PutFingerprintIntoDatabase(fullFingerprint, "fingerprints-track")

	// Calculating Scikit
	if glb.RuntimeArgs.Scikit {
		scikitData = ScikitClassify(groupName, curFingerprint)
		glb.Debug.Println(scikitData)
		for algorithm, valueXY := range scikitData {
			glb.Debug.Println(algorithm + ":" + valueXY)
		}

	}

	// Send out the final responses
	var userPosJson parameters.UserPositionJSON
	userPosJson.Time = curFingerprint.Timestamp
	userPosJson.BayesGuess = bayesGuess
	userPosJson.BayesData = bayesData
	userPosJson.SvmGuess = svmGuess
	userPosJson.SvmData = svmData
	userPosJson.ScikitData = scikitData
	userPosJson.KnnGuess = knnGuess
	userPosJson.KnnData = knnData
	userPosJson.PDRLocation = pdrLocation
	userPosJson.Fingerprint = curFingerprint

	// User history effect
	//location = knnGuess
	//userHistory := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)
	//location, accuracyCircleRadius = HistoryEffect(userPosJson, userHistory)

	//if glb.GraphEnabled {
		location = knnGuess
	//}
	userPosJson.Location = location
	glb.Debug.Println("Knn guess: ", knnGuess)
	glb.Debug.Println("location: ", location)

	//location = userPosJson.KnnGuess
	userPosJson.KnnGuess = location //todo: must add location as seprated variable from knnguess in parameters.UserPositionJSON
	//go dbm.SetUserPositionCache(strings.ToLower(curFingerprint.Group)+strings.ToLower(curFingerprint.Username), userPosJson)
	gp.Get_ResultData().Append_UserHistory(curFingerprint.Username, userPosJson) //Todo: if we use this as goroutine the delay until running effect next track

	//go gp.Get_ResultData().Append_UserResults(curFingerprint.Username, userPosJson)

	//glb.Debug.Println(len(gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)))

	glb.Debug.Println(userPosJson)

	return userPosJson

}

func RecalculateTrackFingerprintKnnCrossValidation(curFingerprint parameters.Fingerprint) string {
	// Classify with filter curFingerprint
	gp := dbm.GM.GetGroup(curFingerprint.Group)
	dbm.FilterFingerprint(&curFingerprint)
	parameters.CleanFingerprint(&curFingerprint)

	// Calculating KNN
	err, knnGuess, knnData := TrackKnn(gp, curFingerprint, true)
	if err != nil {
		glb.Error.Println(err)
	}

	// Send out the final responses
	var userPosJson parameters.UserPositionJSON
	userPosJson.Time = curFingerprint.Timestamp
	userPosJson.KnnGuess = knnGuess
	userPosJson.KnnData = knnData
	userPosJson.Fingerprint = curFingerprint

	// User history effect

	//userHistory := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)
	//location, _ := HistoryEffect(userPosJson, userHistory)

	userPosJson.Location = knnGuess
	//glb.Debug.Println("Knn guess: ", knnGuess)
	//glb.Debug.Println("location: ", location)

	gp.Get_ResultData().Append_UserHistory(curFingerprint.Username, userPosJson) //Todo: if we use this as goroutine the delay until running effect next track

	//glb.Debug.Println(userPosJson)
	/*	glb.Error.Println("###############")
		glb.Error.Println(userPosJson.KnnGuess)
		glb.Error.Println(userPosJson.KnnData)*/

	return userPosJson.Location
}

/*func CalculateLearn2(groupName string) {
	// Now performance isn't important in learning, just care about performance on track (it helps to code easily!)

	// Preprocess
	PreProcess(groupName)

	glb.Debug.Println("################### enetered CalculateLearn ##################")
	groupName = strings.ToLower(groupName)
	gp := dbm.GM.GetGroup(groupName)

	gp.Set_Permanent(false) //for crossvalidation

	//rd := gp.Get_RawData_Filtered_Val()  //Todo: instead of local rd , preprocess rd and save that
	rd := gp.Get_RawData_Val()

	var crossValidationPartsList []crossValidationParts
	crossValidationPartsList = GetCrossValidationParts(gp,rd)
	// ToDo: Need to learn algorithms concurrently

	// CrossValidation

	totalErrorList := []int{}
	knnErrHyperParameters := make(map[int][]interface{})

	bestK := 1
	bestMinClusterRss:= 1
	bestResult := -1


	//Set algorithm parameters range:

	// KNN:
	// Parameters list creation
		// 1.K
	validKs := glb.MakeRange(glb.DefaultKnnKRange[0],glb.DefaultKnnKRange[1])
	knnKRange := dbm.GetSharedPrf(gp.Get_Name()).KnnKRange
	if len(knnKRange) == 1{
		validKs = glb.MakeRange(knnKRange[0],knnKRange[0])
	}else if len(knnKRange) == 2{
		validKs = glb.MakeRange(knnKRange[0],knnKRange[1])
	}else{
		glb.Error.Println("Can't set valid Knn K values")
	}
		//2.MinClusterRSS
	validMinClusterRSSs := glb.MakeRange(glb.DefaultKnnMinCRssRange[0],glb.DefaultKnnMinCRssRange[1])

	minClusterRSSRange := dbm.GetSharedPrf(gp.Get_Name()).KnnMinCRssRange
	if len(minClusterRSSRange) == 1{
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0],minClusterRSSRange[0])
	}else if len(knnKRange) == 2{
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0],minClusterRSSRange[1])
	}else{
		glb.Error.Println("Can't set valid Knn K values")
	}


	// Set length of calculation progress bar
	// This is shared between all threads, so it's invalid when two calculateLearn thread run
	glb.ProgressBarLength = len(validMinClusterRSSs) * len(validKs)

	knnLocAccuracy := make(map[string]int)

	if threadedCross{
		numKnnJobs := len(validMinClusterRSSs) * len(validKs)
		runtime.GOMAXPROCS(glb.MaxParallelism())
		chanKnnJobs := make(chan KnnJob, 1+numKnnJobs)
		chanKnnJobResults := make(chan KnnJobResult, 1+numKnnJobs)

		// running calculator thread
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcKNN(id, chanKnnJobs, chanKnnJobResults)
		}



		for _, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
			//glb.Debug.Println("KNN minClusterRss :", minClusterRss)
			for _, K := range validKs { // for over KnnK
				chanKnnJobs <- KnnJob{
					gp:gp,
					K: K,
					MinClusterRss: minClusterRss,
					crossValidationPartsList: crossValidationPartsList}
			}
		}
		close(chanKnnJobs)

		// Get calculated resuls
		for i := 1; i <= numKnnJobs; i++ {
			res := <-chanKnnJobResults
			totalErrorList = append(totalErrorList, res.TotalError)
			knnErrHyperParameters[res.TotalError] = res.KnnErrHyperParameters
		}
		close(chanKnnJobResults)

	}else{
		adTemp := gp.NewAlgoDataStruct()

		for i, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
			for j, K := range validKs { // for over KnnK
				glb.ProgressBarCurLevel = i*len(validKs)+j
				totalDistError := 0

				// 1-foldCrossValidation (each round one location select as test set)
				for _,CVParts := range crossValidationPartsList{

					//glb.Debug.Println(CVNum)

					// Learn:
					mdTemp := gp.NewMiddleDataStruct()
					rdTemp := CVParts.GetTrainSet(gp)
					testFPs := CVParts.testSet.Fingerprints
					testFPsOrdering := CVParts.testSet.FingerprintsOrdering
					GetParameters(mdTemp, rdTemp)
					tempHyperParameters := []interface{}{K,minClusterRss}
					learnedKnnData,_:= LearnKnn(mdTemp,rdTemp,tempHyperParameters)

					// Set hyper parameters
					learnedKnnData.HyperParameters = parameters.KnnHyperParameters{K:K,MinClusterRss:minClusterRss}

					adTemp.Set_KnnFPs(learnedKnnData)

					// Set to main group
					gp.GMutex.Lock() //For each group there's a lock to avoid race between concurrent calculateLearn s
					gp.Set_AlgoData(adTemp)

					// Error calculation for this round
					distError := 0
					trackedPointsNum := 0
					testLocation := testFPs[testFPsOrdering[0]].Location
					for _,index := range testFPsOrdering{
						fp := testFPs[index]

						resultDot := ""
						var err error
						err, resultDot, _ = TrackKnn(gp, fp, false)
						if err != nil{
							if err.Error() == "NumofAP_lowerThan_MinApNum"{
								continue
							} else if err.Error() == "NoValidFingerprints" {
								continue
							}
						}else{
							trackedPointsNum++
						}

						//glb.Debug.Println(fp.Timestamp)
						resx, resy := glb.GetDotFromString(resultDot)
						x, y := glb.GetDotFromString(testLocation)
						//if fp.Timestamp==int64(1516794991872647445){
						//	glb.Error.Println("ResultDot = ",resultDot)
						//	glb.Error.Println("DistError = ",int(calcDist(x,y,resx,resy)))
						//}
						distError += int(glb.CalcDist(x, y, resx, resy))
						if distError < 0 { //print if distError is lower than zero(it's for error detection)
							glb.Error.Println(fp)
							glb.Error.Println(resultDot)
							_, resultDot, _ = TrackKnn(gp, fp, false)
							glb.Error.Println(x,y)
							glb.Error.Println(resx,resy)
						}
					}
					if trackedPointsNum==0{
						glb.Error.Println("For loc:",testLocation," there is no fingerprint that its number of APs be more than",glb.MinApNum)
					}else{
						distError = distError/trackedPointsNum
						totalDistError += distError
					}

					gp.GMutex.Unlock()
				}

				glb.Debug.Printf("Knn error (minClusterRss=%d,K=%d) = %d \n", minClusterRss,K,totalDistError)

				//if(bestResult==-1 || totalDistError<bestResult){
				//	bestResult = totalDistError
				//	bestK = K
				//}
				totalErrorList = append(totalErrorList,totalDistError)
				knnErrHyperParameters[totalDistError] = []interface{}{K, minClusterRss}

			}
		}
	}

	glb.ProgressBarCurLevel = 0 // reset progressBar level

	// Select best hyperParameters
	//glb.Debug.Println(totalErrorList)
	sort.Ints(totalErrorList)
	bestResult = totalErrorList[0]
	bestErrHyperParameters := knnErrHyperParameters[bestResult]
	bestK = bestErrHyperParameters[0].(int)
	bestMinClusterRss = bestErrHyperParameters[1].(int)

	glb.Debug.Println("CrossValidation resuts:")
	for _, res := range totalErrorList {
		glb.Debug.Println(knnErrHyperParameters[res], " : ", res)
	}
	//glb.Debug.Println()
	glb.Debug.Println("Best K : ",bestK)
	glb.Debug.Println("Best MinClusterRss : ",bestMinClusterRss)
	glb.Debug.Println("Minimum error = ",bestResult)

	// Calculating each location detection accuracy with best hyperParameters:
	for _,CVParts := range crossValidationPartsList{

		//glb.Debug.Println(CVNum)
		mdTemp := gp.NewMiddleDataStruct()
		adTemp := gp.NewAlgoDataStruct()
		rdTemp := CVParts.GetTrainSet(gp)
		testFPs := CVParts.testSet.Fingerprints

		// Learn
		testFPsOrdering := CVParts.testSet.FingerprintsOrdering
		GetParameters(mdTemp, rdTemp)
		tempHyperParameters := []interface{}{bestK,bestMinClusterRss}
		learnedKnnData,_:= LearnKnn(mdTemp,rdTemp,tempHyperParameters)

		// Set hyper parameters
		learnedKnnData.HyperParameters = parameters.KnnHyperParameters{K:bestK,MinClusterRss:bestMinClusterRss}

		adTemp.Set_KnnFPs(learnedKnnData)

		gp.GMutex.Lock()
		gp.Set_AlgoData(adTemp)

		// Error calculation for each location with best hyperParameters
		distError := 0
		trackedPointsNum := 0
		testLocation := testFPs[testFPsOrdering[0]].Location
		for _,index := range testFPsOrdering{

			fp := testFPs[index]
			resultDot := ""
			var err error
			err, resultDot, _ = TrackKnn(gp, fp, false)

			if err != nil {
				if err.Error() == "NumofAP_lowerThan_MinApNum" {
					continue
				} else if err.Error() == "NoValidFingerprints" {
					continue
				}
			}else{
				trackedPointsNum++
			}
			//glb.Debug.Println(fp.Location, " ==== ", resultDot)

			resx, resy := glb.GetDotFromString(resultDot)
			x, y := glb.GetDotFromString(testLocation) // testLocation is fp.Location
			distError += int(glb.CalcDist(x, y, resx, resy))
			if distError < 0{
				glb.Error.Println(fp)
				glb.Error.Println(resultDot)
				_, resultDot, _ = TrackKnn(gp, fp, false)
				glb.Error.Println(x,y)
				glb.Error.Println(resx,resy)
			}
		}
		gp.GMutex.Unlock()

		if trackedPointsNum==0{
			glb.Error.Println("For loc:",testLocation," there is no fingerprint that its number of APs be more than",glb.MinApNum)
			knnLocAccuracy[testLocation] = -1
		}else{
			distError = distError/trackedPointsNum
			knnLocAccuracy[testLocation] = distError
		}



	}

	// Set CrossValidation results
	rs := gp.Get_ResultData()
	glb.Debug.Println(gp.Get_Name())
	rs.Set_AlgoAccuracy("knn",bestResult)
	for loc,accuracy := range knnLocAccuracy{
		rs.Set_AlgoLocAccuracy("knn",loc,accuracy)
	}
	glb.Debug.Println(dbm.GetCVResults(gp.Get_Name()))

	// Set main parameters
	md := gp.NewMiddleDataStruct()
	GetParameters(md, rd)

	gp.GMutex.Lock()
	gp.Set_MiddleData(md)
	glb.Debug.Println(md.UniqueMacs)
	// select best algo config

	// learn algorithm
	ad := gp.Get_AlgoData()
	gp.Set_Permanent(true)
	bestHyperParameters := []interface{}{bestK,bestMinClusterRss}
	learnedKnnData,_:= LearnKnn(md,rd,bestHyperParameters)

	// Set best hyper parameter values
	learnedKnnData.HyperParameters = parameters.KnnHyperParameters{K:bestK,MinClusterRss:bestMinClusterRss}

	ad.Set_KnnFPs(learnedKnnData)
	gp.Set_AlgoData(ad)

	if glb.RuntimeArgs.Scikit {
		ScikitLearn(groupName)
	}

	gp.GMutex.Unlock()

	glb.Debug.Println("Calculation finished.")
	//if glb.RuntimeArgs.Svm {
	//	DumpFingerprintsSVM(groupName)
	//	err := CalculateSVM(groupName)
	//	if err != nil {
	//		glb.Warning.Println("Encountered error when calculating SVM")
	//		glb.Warning.Println(err)
	//	}
	//}

	//runnerLock.Unlock()
}

*/
/*func calcKNN(id int, knnJobs <-chan KnnJob, knnJobResults chan<- KnnJobResult) {
	totalDistError := 0

	//glb.Debug.Println("KNN K :",K)
	//temptemptemp := make(map[string]float64)

	for knnJob := range knnJobs {
		adTemp := knnJob.gp.NewAlgoDataStruct() //todo: gp.algo are lock here !!!! i use newalgodatastruct() instead of get_algo
		//glb.Debug.Println(len(crossValidationPartsList))
		for _,CVParts := range knnJob.crossValidationPartsList{
			//glb.Debug.Println(CVNum)
			mdTemp := knnJob.gp.NewMiddleDataStruct()
			rdTemp := CVParts.GetTrainSet(knnJob.gp)
			testFPs := CVParts.testSet.Fingerprints


			testFPsOrdering := CVParts.testSet.FingerprintsOrdering
			GetParameters(mdTemp, rdTemp)
			tempHyperParameters := []interface{}{knnJob.K,knnJob.MinClusterRss}
			learnedKnnData,_:= LearnKnn(mdTemp,rdTemp,tempHyperParameters)

			// Set hyper parameters
			learnedKnnData.HyperParameters = parameters.KnnHyperParameters{K:knnJob.K,MinClusterRss:knnJob.MinClusterRss}
			adTemp.Set_KnnFPs(learnedKnnData)
			knnJob.gp.Set_AlgoData(adTemp)

			distError := 0
			//FPtEMP := parameters.Fingerprint{}

			for _,index := range testFPsOrdering{
				fp := testFPs[index]
				//FPtEMP = fp
				//if(fp.Location =="-165.000000,-1295.000000"){
				//glb.Warning.Println(index)
				_, resultDot, _ := TrackKnn(knnJob.gp, fp, false)
				//glb.Debug.Println(fp.Location," ==== ",resultDot)
				//glb.Debug.Println(fp)

				resx, resy := glb.GetDotFromString(resultDot)
				x, y := glb.GetDotFromString(fp.Location)
				distError += int(glb.CalcDist(x, y, resx, resy))
				//}
			}
			totalDistError += distError
		}

		glb.Debug.Printf("Knn error (minClusterRss=%d,K=%d) = %f \n", knnJob.MinClusterRss,knnJob.K,totalDistError)
		//if(bestResult==-1 || totalDistError<bestResult){
		//	bestResult = totalDistError
		//	bestK = K
		//}

		knnJobResults <- KnnJobResult{
			TotalError: totalDistError,
			KnnErrHyperParameters: []interface{}{knnJob.K, knnJob.MinClusterRss},
			}
	}



}
*/

type crossValidationParts struct {
	trainSet []dbm.RawDataStruct
	testSet  dbm.RawDataStruct
}

func (cvParts *crossValidationParts) GetTrainSet(gp *dbm.Group) dbm.RawDataStruct {
	resRD := *gp.NewRawDataStruct()
	for _, rd := range cvParts.trainSet {
		for _, index := range rd.FingerprintsOrdering {
			resRD.FingerprintsOrdering = append(resRD.FingerprintsOrdering, index)
			resRD.Fingerprints[index] = rd.Fingerprints[index]
		}
	}
	return resRD
}

func GetCrossValidationParts(gp *dbm.Group, fpOrdering []string, fpData map[string]parameters.Fingerprint) []crossValidationParts {
	var CVPartsList []crossValidationParts
	var tempCVParts crossValidationParts
	locRDMap := make(map[string]dbm.RawDataStruct)

	for _, index := range fpOrdering {
		fp := fpData[index]

		if fpRD, ok := locRDMap[fp.Location]; ok {
			fpRD.Fingerprints[index] = fp
			fpRD.FingerprintsOrdering = append(fpRD.FingerprintsOrdering, index)
			locRDMap[fp.Location] = fpRD
		} else {
			fpM := make(map[string]parameters.Fingerprint)
			fpM[index] = fp
			fpO := []string{index}
			templocRDMap := *gp.NewRawDataStruct()
			templocRDMap.Fingerprints = fpM
			templocRDMap.FingerprintsOrdering = fpO
			locRDMap[fp.Location] = templocRDMap
			//
			//locRDMap[fp.Location] = dbm.RawDataStruct{
			//	Fingerprints:	fpM,
			//	FingerprintsOrdering: fpO,
			//}
		}
	}

	//glb.Error.Println(len(locRDMap))
	//glb.Error.Println("###############")
	//for loc,RD := range locRDMap{
	//	glb.Error.Println(len(RD.Fingerprints))
	//	for _,fp := range RD.Fingerprints{
	//		if loc != fp.Location {
	//			glb.Error.Println("FOUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUND")
	//		}
	//	}
	//}

	tempCVParts.testSet = dbm.RawDataStruct{}
	tempCVParts.trainSet = []dbm.RawDataStruct{}

	for locMain, _ := range locRDMap {
		tempCVParts.testSet = dbm.RawDataStruct{}
		tempCVParts.trainSet = []dbm.RawDataStruct{}
		for loc, RD := range locRDMap {
			if (loc == locMain) {
				// add to test set
				tempCVParts.testSet = RD
			} else {
				// add to train set
				tempCVParts.trainSet = append(tempCVParts.trainSet, RD)
			}
		}
		CVPartsList = append(CVPartsList, tempCVParts)
	}
	return CVPartsList
}

func GetParameters(md *dbm.MiddleDataStruct, rd dbm.RawDataStruct) {
	//
	//persistentPs, err := dbm.OpenPersistentParameters(group) //persistentPs is just like ps but with renamed network name; e.g.: "0" -> "1"
	//if err != nil {
	//	//log.Fatal(err)
	//	glb.Error.Println(err)
	//}
	fingerprints := rd.Fingerprints
	fingerprintsOrdering := rd.FingerprintsOrdering

	//glb.Error.Println("d")
	md.NetworkMacs = make(map[string]map[string]bool)
	md.NetworkLocs = make(map[string]map[string]bool)
	md.UniqueMacs = []string{}
	md.UniqueLocs = []string{}
	md.MacCount = make(map[string]int)
	md.MacCountByLoc = make(map[string]map[string]int)
	//md.Loaded = true
	//opening the db
	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//defer db.Close()
	// if err != nil {
	//	log.Fatal(err)
	//}

	macs := []string{}

	// Get all parameters that don't need a network graph (?)
	for _, v1 := range fingerprintsOrdering {

		//log.Println("calculateResults=true")
		v2 := fingerprints[v1]

		// append the fingerprint location to UniqueLocs array if doesn't exist in it.
		if !glb.StringInSlice(v2.Location, md.UniqueLocs) {
			md.UniqueLocs = append(md.UniqueLocs, v2.Location)
		}

		// MacCountByLoc initialization for new location
		if _, ok := md.MacCountByLoc[v2.Location]; !ok {
			md.MacCountByLoc[v2.Location] = make(map[string]int)
		}

		//// building network
		//macs := []string{}

		for _, router := range v2.WifiFingerprint {
			// building network
			macs = append(macs, router.Mac)

			// append the fingerprint mac to UniqueMacs array if doesn't exist in it.
			if !glb.StringInSlice(router.Mac, md.UniqueMacs) {
				md.UniqueMacs = append(md.UniqueMacs, router.Mac)
			}

			// mac count
			if _, ok := md.MacCount[router.Mac]; !ok {
				md.MacCount[router.Mac] = 0
			}
			md.MacCount[router.Mac]++

			// mac by location count
			if _, ok := md.MacCountByLoc[v2.Location][router.Mac]; !ok {
				md.MacCountByLoc[v2.Location][router.Mac] = 0
			}
			md.MacCountByLoc[v2.Location][router.Mac]++
		}

		// building network
		//ps.NetworkMacs = buildNetwork(ps.NetworkMacs, macs)
	}
	// todo: network definition and buildNetwork() must be redefined
	md.NetworkMacs = clustring.BuildNetwork(md.NetworkMacs, macs)
	md.NetworkMacs = clustring.MergeNetwork(md.NetworkMacs)

	//Error.Println("ps.Networkmacs", ps.NetworkMacs)
	// Rename the NetworkMacs
	//if len(persistentPs.NetworkRenamed) > 0 {
	//	newNames := []string{}
	//	for k := range persistentPs.NetworkRenamed {
	//		newNames = append(newNames, k)
	//
	//	}
	//	//todo: \/ wtf? Rename procedure could be redefined better.
	//	for n := range md.NetworkMacs {
	//		renamed := false
	//		for mac := range md.NetworkMacs[n] {
	//			for renamedN := range persistentPs.NetworkRenamed {
	//				if glb.StringInSlice(mac, persistentPs.NetworkRenamed[renamedN]) && !glb.StringInSlice(n, newNames) {
	//					md.NetworkMacs[renamedN] = make(map[string]bool)
	//					for k, v := range md.NetworkMacs[n] {
	//						md.NetworkMacs[renamedN][k] = v //copy ps.NetworkMacs[n] to ps.NetworkMacs[renamedN]
	//					}
	//					delete(md.NetworkMacs, n)
	//					renamed = true
	//				}
	//				if renamed {
	//					break
	//				}
	//			}
	//			if renamed {
	//				break
	//			}
	//		}
	//	}
	//}

	// Get the locations for each graph (Has to have network built first)

	for _, v1 := range fingerprintsOrdering {

		v2 := fingerprints[v1]
		//todo: Make the macs array just once for each fingerprint instead of repeating the process

		macs := []string{}
		for _, router := range v2.WifiFingerprint {
			macs = append(macs, router.Mac)
		}
		//todo: ps.NetworkMacs is created from mac array; so it seems that hasNetwork function doesn't do anything useful!
		networkName, inNetwork := clustring.HasNetwork(md.NetworkMacs, macs)
		if inNetwork {
			if _, ok := md.NetworkLocs[networkName]; !ok {
				md.NetworkLocs[networkName] = make(map[string]bool)
			}
			if _, ok := md.NetworkLocs[networkName][v2.Location]; !ok {
				md.NetworkLocs[networkName][v2.Location] = true
			}
		}
	}

	//calculate locCount
	locations := []string{}
	for _, fpIndex := range fingerprintsOrdering {
		fp := fingerprints[fpIndex]
		locations = append(locations, fp.Location)
	}
	md.LocCount = glb.DuplicateCountString(locations)
}

// Set middleData in gp
func GetParametersWithGP(gp *dbm.Group) {
	//
	//persistentPs, err := dbm.OpenPersistentParameters(group) //persistentPs is just like ps but with renamed network name; e.g.: "0" -> "1"
	//if err != nil {
	//	//log.Fatal(err)
	//	glb.Error.Println(err)
	//}
	rd := gp.Get_RawData_Val()
	md := gp.NewMiddleDataStruct()
	fingerprints := rd.Fingerprints
	fingerprintsOrdering := rd.FingerprintsOrdering

	//glb.Error.Println("d")
	md.NetworkMacs = make(map[string]map[string]bool)
	md.NetworkLocs = make(map[string]map[string]bool)
	md.UniqueMacs = []string{}
	md.UniqueLocs = []string{}
	md.MacCount = make(map[string]int)
	md.MacCountByLoc = make(map[string]map[string]int)
	//md.Loaded = true
	//opening the db
	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	//defer db.Close()
	// if err != nil {
	//	log.Fatal(err)
	//}

	macs := []string{}

	// Get all parameters that don't need a network graph (?)
	for _, v1 := range fingerprintsOrdering {

		//log.Println("calculateResults=true")
		v2 := fingerprints[v1]

		// append the fingerprint location to UniqueLocs array if doesn't exist in it.
		if !glb.StringInSlice(v2.Location, md.UniqueLocs) {
			md.UniqueLocs = append(md.UniqueLocs, v2.Location)
		}

		// MacCountByLoc initialization for new location
		if _, ok := md.MacCountByLoc[v2.Location]; !ok {
			md.MacCountByLoc[v2.Location] = make(map[string]int)
		}

		//// building network
		//macs := []string{}

		for _, router := range v2.WifiFingerprint {
			// building network
			macs = append(macs, router.Mac)

			// append the fingerprint mac to UniqueMacs array if doesn't exist in it.
			if !glb.StringInSlice(router.Mac, md.UniqueMacs) {
				md.UniqueMacs = append(md.UniqueMacs, router.Mac)
			}

			// mac count
			if _, ok := md.MacCount[router.Mac]; !ok {
				md.MacCount[router.Mac] = 0
			}
			md.MacCount[router.Mac]++

			// mac by location count
			if _, ok := md.MacCountByLoc[v2.Location][router.Mac]; !ok {
				md.MacCountByLoc[v2.Location][router.Mac] = 0
			}
			md.MacCountByLoc[v2.Location][router.Mac]++
		}

		// building network
		//ps.NetworkMacs = buildNetwork(ps.NetworkMacs, macs)
	}
	// todo: network definition and buildNetwork() must be redefined
	md.NetworkMacs = clustring.BuildNetwork(md.NetworkMacs, macs)
	md.NetworkMacs = clustring.MergeNetwork(md.NetworkMacs)

	//Error.Println("ps.Networkmacs", ps.NetworkMacs)
	// Rename the NetworkMacs
	//if len(persistentPs.NetworkRenamed) > 0 {
	//	newNames := []string{}
	//	for k := range persistentPs.NetworkRenamed {
	//		newNames = append(newNames, k)
	//
	//	}
	//	//todo: \/ wtf? Rename procedure could be redefined better.
	//	for n := range md.NetworkMacs {
	//		renamed := false
	//		for mac := range md.NetworkMacs[n] {
	//			for renamedN := range persistentPs.NetworkRenamed {
	//				if glb.StringInSlice(mac, persistentPs.NetworkRenamed[renamedN]) && !glb.StringInSlice(n, newNames) {
	//					md.NetworkMacs[renamedN] = make(map[string]bool)
	//					for k, v := range md.NetworkMacs[n] {
	//						md.NetworkMacs[renamedN][k] = v //copy ps.NetworkMacs[n] to ps.NetworkMacs[renamedN]
	//					}
	//					delete(md.NetworkMacs, n)
	//					renamed = true
	//				}
	//				if renamed {
	//					break
	//				}
	//			}
	//			if renamed {
	//				break
	//			}
	//		}
	//	}
	//}

	// Get the locations for each graph (Has to have network built first)

	for _, v1 := range fingerprintsOrdering {

		v2 := fingerprints[v1]
		//todo: Make the macs array just once for each fingerprint instead of repeating the process

		macs := []string{}
		for _, router := range v2.WifiFingerprint {
			macs = append(macs, router.Mac)
		}
		//todo: ps.NetworkMacs is created from mac array; so it seems that hasNetwork function doesn't do anything useful!
		networkName, inNetwork := clustring.HasNetwork(md.NetworkMacs, macs)
		if inNetwork {
			if _, ok := md.NetworkLocs[networkName]; !ok {
				md.NetworkLocs[networkName] = make(map[string]bool)
			}
			if _, ok := md.NetworkLocs[networkName][v2.Location]; !ok {
				md.NetworkLocs[networkName][v2.Location] = true
			}
		}
	}

	//calculate locCount
	locations := []string{}
	for _, fpIndex := range fingerprintsOrdering {
		fp := fingerprints[fpIndex]
		locations = append(locations, fp.Location)
	}
	md.LocCount = glb.DuplicateCountString(locations)

	gp.Set_MiddleData(md)
}

//Note: Use it just one time(not use it in calculatelearn, use it in buildgroup)
func PreProcess(rd *dbm.RawDataStruct, needToRelocateFP bool) {

	fingerprintsOrdering := rd.Get_FingerprintsOrderingBackup()
	fingerprintsData := rd.Get_FingerprintsBackup()

	// filter fingerprints according to filtermac list
	for _, fpIndex := range fingerprintsOrdering {
		fp := fingerprintsData[fpIndex]
		dbm.FilterFingerprint(&fp)
		fingerprintsData[fpIndex] = fp
	}

	// Regulation rss data(outline detection)
	// Grouping fingerprints by location
	locRDMap := make(map[string]dbm.RawDataStruct)
	for _, index := range fingerprintsOrdering {
		fp := fingerprintsData[index]

		if fpRD, ok := locRDMap[fp.Location]; ok {
			fpRD.Fingerprints[index] = fp
			fpRD.FingerprintsOrdering = append(fpRD.FingerprintsOrdering, index)
			locRDMap[fp.Location] = fpRD
		} else {
			fpM := make(map[string]parameters.Fingerprint)
			fpM[index] = fp
			fpO := []string{index}
			templocRDMap := dbm.RawDataStruct{}
			templocRDMap.Fingerprints = fpM
			templocRDMap.FingerprintsOrdering = fpO
			locRDMap[fp.Location] = templocRDMap
			//
			//locRDMap[fp.Location] = dbm.RawDataStruct{
			//	Fingerprints:	fpM,
			//	FingerprintsOrdering: fpO,
			//}
		}
	}

	if !glb.FastLearn {
		//regulating rss data
		if glb.RssRegulation {
			for loc, rdData := range locRDMap {
				locRDMap[loc] = RemoveOutlines(rdData)
			}
		}
	}

	// converting locRDMap to rd
	tempFingerprintsData := make(map[string]parameters.Fingerprint)
	tempFingerprintsOrdering := []string{}
	for _, rdData := range locRDMap {
		for _, fpTime := range rdData.FingerprintsOrdering {
			tempFingerprintsOrdering = append(tempFingerprintsOrdering, fpTime)
			tempFingerprintsData[fpTime] = rdData.Fingerprints[fpTime]
		}

	}

	//Average Rss vector of adjacent fingerprints
	if !glb.FastLearn {
		if needToRelocateFP {
			if glb.AvgRSSAdjacentDots {
				maxValidFPDistAVG := float64(100); // 100 cm

				tempFingerprintsData2 := make(map[string]parameters.Fingerprint)
				for fpOMain, fpMain := range tempFingerprintsData {
					adjacentFPs := []parameters.Fingerprint{}
					for fpO, fp := range tempFingerprintsData {
						if (fpO == fpOMain) {
							continue
						}
						xyMain := strings.Split(fpMain.Location, ",")
						xy := strings.Split(fp.Location, ",")
						if len(xyMain) != 2 || len(xy) != 2 {
							glb.Error.Println("location value doesn't have x,y format: xyMain:" + fpMain.Location + " & xy:" + fp.Location)
							break
						}

						xMain, err1 := glb.StringToFloat(xyMain[0])
						yMain, err2 := glb.StringToFloat(xyMain[1])
						x, err3 := glb.StringToFloat(xy[0])
						y, err4 := glb.StringToFloat(xy[1])
						if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
							glb.Error.Println(err1)
							glb.Error.Println(err2)
							glb.Error.Println(err3)
							glb.Error.Println(err4)
							break
						}
						//glb.Debug.Println(x,",",y)
						//glb.Debug.Println(xMain,",",yMain)

						dist := glb.CalcDist(x, y, xMain, yMain)

						//glb.Debug.Println(dist)
						if dist < maxValidFPDistAVG {
							adjacentFPs = append(adjacentFPs, fp)
						}
					}
					//glb.Error.Println(len(adjacentFPs))

					//Average rss
					newRouteWithAvgRss := []parameters.Router{}
					mac2RssList := make(map[string][]int) // todo: problem with fingerprints that doesn't have some mac in their list
					for _, rt := range fpMain.WifiFingerprint {
						mac2RssList[rt.Mac] = []int{rt.Rssi}
					}

					if len(adjacentFPs) != 0 {
						for _, adjfp := range adjacentFPs {
							for _, rt := range adjfp.WifiFingerprint {
								if rssList, ok := mac2RssList[rt.Mac]; ok {
									mac2RssList[rt.Mac] = append(rssList, rt.Rssi)
								}
							}
						}

						//glb.Debug.Println(mac2RssList)
						for mac, rssList := range mac2RssList {
							avgRss := 0
							for _, rss := range rssList {
								avgRss += rss
							}
							avgRss /= len(rssList)
							//glb.Debug.Println(avgRss)
							rt := parameters.Router{Mac: mac, Rssi: avgRss}
							newRouteWithAvgRss = append(newRouteWithAvgRss, rt)
						}
						fpMainTemp := fpMain
						fpMainTemp.WifiFingerprint = newRouteWithAvgRss //change mac,rss list
						tempFingerprintsData2[fpOMain] = fpMainTemp
					} else {
						tempFingerprintsData2[fpOMain] = fpMain
					}

				}
				tempFingerprintsData = tempFingerprintsData2
			}
		}
	}

	//// save processed data
	rd.Set_Fingerprints(tempFingerprintsData)
	rd.Set_FingerprintsOrdering(tempFingerprintsOrdering)
}

func RemoveOutlines(rd dbm.RawDataStruct) dbm.RawDataStruct {

	macs := []string{}
	for _, fp := range rd.Fingerprints {
		for _, rt := range fp.WifiFingerprint {
			if !glb.StringInSlice(rt.Mac, macs) {
				macs = append(macs, rt.Mac)
			}
		}
	}

	//rd.Fingerprints[rd.FingerprintsOrdering[0]].Location == "999.0,645.0"
	//Converting fp.wififingerprints to map
	fpMap := make(map[string]map[string]int)
	for _, fpTime := range rd.FingerprintsOrdering {
		fpMap[fpTime] = make(map[string]int)
		for _, rt := range rd.Fingerprints[fpTime].WifiFingerprint {
			fpMap[fpTime][rt.Mac] = rt.Rssi
		}
	}

	fpNum := len(rd.Fingerprints)

	maxOutlineNum := int(float64(fpNum) * glb.PreprocessOutlinePercent)

	minNonOutline := int(fpNum / 2)
	//glb.Error.Println(macs)
	for _, mac := range macs {
		nonOutlineCount := 0
		validFP := []string{}
		rssVals := []int{}
		for _, fpTime := range rd.FingerprintsOrdering {
			//fp := rd.Fingerprints[fpTime]
			//for _,rt := range fp.WifiFingerprint{
			//	if rt.Mac == mac {

			//glb.Error.Println(fpMap[fpTime][mac])
			if rss, ok := fpMap[fpTime][mac]; ok {
				nonOutlineCount++
				validFP = append(validFP, fpTime)
				rssVals = append(rssVals, rss)
			}

			//	}
			//}
		}

		//glb.Error.Println(minNonOutline)
		//glb.Error.Println(nonOutlineCount)
		if nonOutlineCount < minNonOutline { //ignore this mac
			continue
		}

		// todo: instead of sorting, use proper parameters or algorithm to choose rsss

		substRssVal := glb.Median(rssVals)
		//glb.Error.Println(substRssVal)

		for _, fpTime := range rd.FingerprintsOrdering { //substitude outlines with substRssVal
			if !glb.StringInSlice(fpTime, validFP) {
				//for i,rt := range rd.Fingerprints[fpTime].WifiFingerprint{
				//	if rt.Mac == mac{
				//		rd.Fingerprints[fpTime].WifiFingerprint[i].Rssi = substRssVal
				//	}
				//}
				//rd.Fingerprints[fpTime].Location=="988.0,1441.0" && mac=="01:17:C5:97:58:C3"
				fpMap[fpTime][mac] = substRssVal
			}
		}

		if fpNum-nonOutlineCount >= maxOutlineNum { //loss rss are outlines, no need to calculate outlines
			continue
		}

		remainOutlineNum := maxOutlineNum - (fpNum - nonOutlineCount)
		//Calculate outlines
		rssDist := make(map[string]float64)
		for _, fpTime := range validFP {
			rss := fpMap[fpTime][mac]
			rssDist[fpTime] = math.Abs(float64(substRssVal - rss))
		}

		sortedFP := glb.SortFPByRSS(rssDist)

		outlineChangeCount := 0
		for _, fpTime := range sortedFP {
			//glb.Error.Println(fpMap[fpTime][mac])
			//glb.Error.Println(substRssVal)
			//glb.Error.Println()

			if math.Abs(float64(fpMap[fpTime][mac]-substRssVal)) < float64(glb.NormalRssDev+1) { // Avoid changing too close rss
				break
			}
			fpMap[fpTime][mac] = substRssVal
			outlineChangeCount++
			if outlineChangeCount == remainOutlineNum {
				break
			}
		}

	}

	// Converting fpMap to fingerprint.Wififingerpint
	//for _,fpTime := range rd.FingerprintsOrdering{
	//	for i,rt := range rd.Fingerprints[fpTime].WifiFingerprint{
	//		rd.Fingerprints[fpTime].WifiFingerprint[i].Rssi = fpMap[fpTime][rt.Mac]
	//	}
	//}

	for _, fpTime := range rd.FingerprintsOrdering {
		fp := rd.Fingerprints[fpTime]
		tempRouteList := []parameters.Router{}
		for mac, rss := range fpMap[fpTime] {
			tempRouteList = append(tempRouteList, parameters.Router{Mac: mac, Rssi: rss})
		}
		newFP := parameters.Fingerprint{
			Timestamp:       fp.Timestamp,
			Location:        fp.Location,
			Username:        fp.Username,
			Group:           fp.Group,
			WifiFingerprint: tempRouteList,
		}
		rd.Fingerprints[fpTime] = newFP
		//for i,rt := range rd.Fingerprints[fpTime].WifiFingerprint{
		//	rd.Fingerprints[fpTime].WifiFingerprint[i].Rssi = fpMap[fpTime][rt.Mac]
		//}
	}

	return rd
}

// Error calculation using mean as error algorithm
func GetBestKnnHyperParams(groupName string, shprf dbm.RawSharedPreferences, cd *dbm.ConfigDataStruct, crossValidationPartsList []crossValidationParts) parameters.KnnHyperParameters {
	// CrossValidation
	tempGp := dbm.GM.GetGroup(groupName, false) //permanent:false
	//tempGp.Set_Permanent(false)
	tempGp.Set_ConfigData(cd)

	//totalErrorList := []int{}
	//knnErrHyperParameters := make(map[int]parameters.KnnHyperParameters)

	allHyperParamDetails := make(map[int]parameters.KnnHyperParameters)
	allErrDetails := make(map[int][]int)

	//bestResult := -1

	//Set algorithm parameters range:

	// KNN:
	// Parameters list creation
	// 1.K
	validKs := glb.MakeRange(glb.DefaultKnnKRange[0], glb.DefaultKnnKRange[1])
	knnKRange := shprf.KnnKRange
	if len(knnKRange) == 1 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[0])
	} else if len(knnKRange) == 2 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[1])
	} else {
		glb.Error.Println("knnKRange:", knnKRange)
		glb.Error.Println("Can't set valid Knn K values")
	}
	//2.MinClusterRSS
	validMinClusterRSSs := glb.MakeRange(glb.DefaultKnnMinCRssRange[0], glb.DefaultKnnMinCRssRange[1])

	minClusterRSSRange := shprf.KnnMinCRssRange
	if len(minClusterRSSRange) == 1 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[0])
	} else if len(minClusterRSSRange) == 2 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[1])
		validMinClusterRSSs = append(validMinClusterRSSs, 0)
	} else {
		glb.Error.Println("minClusterRSSRange:", minClusterRSSRange)
		glb.Error.Println("Can't set valid min cluster rss values")
	}

	// Set length of calculation progress bar
	// This is shared between all threads, so it's invalid when two calculateLearn thread run
	calculationLen := len(validMinClusterRSSs) * len(validKs)
	glb.ProgressBarLength = calculationLen

	adTemp := tempGp.NewAlgoDataStruct()
	rdTemp := tempGp.Get_RawData()

	//allErrDetailsList = make([][]int,calculationLen)

	paramUniqueKey := 0 // just creating unique key for each possible the parameters permutation
	for i, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
		for j, K := range validKs { // for over KnnK
			glb.ProgressBarCurLevel = i*len(validKs) + j
			totalDistError := 0

			tempHyperParameters := parameters.NewKnnHyperParameters()
			//glb.Error.Println(tempHyperParameters)
			tempHyperParameters.K = K
			tempHyperParameters.MinClusterRss = minClusterRss

			paramUniqueKey++
			allHyperParamDetails[paramUniqueKey] = tempHyperParameters
			tempAllErrDetailList := []int{}
			//tempAllErrDetailList := make([]int,len(crossValidationPartsList))

			// 1-foldCrossValidation (each round one location select as test set)
			for CVNum, CVParts := range crossValidationPartsList {
				glb.Debug.Println("CrossValidation Part num :", CVNum)

				// Learn:
				trainSetTemp := CVParts.GetTrainSet(tempGp)
				rdTemp.Set_Fingerprints(trainSetTemp.Fingerprints)
				rdTemp.Set_FingerprintsOrdering(trainSetTemp.FingerprintsOrdering)
				rdTemp.Set_FingerprintsBackup(trainSetTemp.Fingerprints)
				rdTemp.Set_FingerprintsOrderingBackup(trainSetTemp.FingerprintsOrdering)

				testFPs := CVParts.testSet.Fingerprints
				testFPsOrdering := CVParts.testSet.FingerprintsOrdering

				GetParametersWithGP(tempGp)

				PreProcess(rdTemp, shprf.NeedToRelocateFP)

				learnedKnnData, _ := LearnKnn(tempGp, tempHyperParameters)

				/*			// Set hyper parameters
							learnedKnnData.HyperParameters = parameters.NewKnnHyperParameters()
							learnedKnnData.HyperParameters.K = K
							learnedKnnData.HyperParameters.MinClusterRss = minClusterRss*/

				//learnedKnnData.HyperParameters = parameters.KnnHyperParameters{K: K, MinClusterRss: minClusterRss}

				adTemp.Set_KnnFPs(learnedKnnData)

				// Set to main group
				tempGp.GMutex.Lock() //For each group there's a lock to avoid race between concurrent calculateLearn s
				tempGp.Set_AlgoData(adTemp)

				// Error calculation for this round
				distError := 0
				trackedPointsNum := 0
				testLocation := testFPs[testFPsOrdering[0]].Location
				for _, index := range testFPsOrdering {
					fp := testFPs[index]

					resultDot := ""
					var err error
					err, resultDot, _ = TrackKnn(tempGp, fp, false)
					if err != nil {
						if err.Error() == "NumofAP_lowerThan_MinApNum" {
							glb.Error.Println("NumofAP_lowerThan_MinApNum")
							continue
						} else if err.Error() == "NoValidFingerprints" {
							glb.Error.Println("NoValidFingerprints")
							continue
						}
					} else {
						trackedPointsNum++
					}

					//glb.Debug.Println(fp.Timestamp)
					resx, resy := glb.GetDotFromString(resultDot)
					x, y := glb.GetDotFromString(testLocation)
					//if fp.Timestamp==int64(1516794991872647445){
					//	glb.Error.Println("ResultDot = ",resultDot)
					//	glb.Error.Println("DistError = ",int(calcDist(x,y,resx,resy)))
					//}
					tempDistError := int(glb.CalcDist(x, y, resx, resy))
					if tempDistError < 0 { //print if distError is lower than zero(it's for error detection)
						glb.Error.Println(fp)
						glb.Error.Println(resultDot)
						_, resultDot, _ = TrackKnn(tempGp, fp, false)
						glb.Error.Println(x, y)
						glb.Error.Println(resx, resy)
					} else {
						distError += tempDistError
						tempAllErrDetailList = append(tempAllErrDetailList, tempDistError)
						//tempAllErrDetailList[CVNum] = tempDistError
						//allErrDetails[paramUniqueKey] = append(allErrDetails[paramUniqueKey],tempDistError)
					}
				}
				if trackedPointsNum == 0 {
					glb.Error.Println("For loc:", testLocation, " there is no fingerprint that its number of APs be more than", glb.MinApNum)
				} else {
					//distError = distError / trackedPointsNum
					//totalDistError += distError
				}
				tempGp.GMutex.Unlock()
			}

			allErrDetails[paramUniqueKey] = tempAllErrDetailList
			//allErrDetailsList[paramUniqueKey] = tempAllErrDetailList
			glb.Debug.Printf("Knn error (minClusterRss=%d,K=%d) = %d \n", minClusterRss, K, totalDistError)

			//if(bestResult==-1 || totalDistError<bestResult){
			//	bestResult = totalDistError
			//	bestK = K
			//}
			//totalErrorList = append(totalErrorList, totalDistError)
			//knnErrHyperParameters[totalDistError] = tempHyperParameters
		}
	}

	glb.Debug.Println((allErrDetails))

	glb.ProgressBarCurLevel = 0 // reset progressBar level

	// Select best hyperParameters
	//glb.Debug.Println(totalErrorList)

	//bestKey, sortedErrDetails, newErrorMap, err := SelectLowestError(allErrDetails,"FalseRate")
	bestKey, sortedErrDetails, newErrorMap := SelectBestFromErrMap(allErrDetails)
	bestErrHyperParameters := allHyperParamDetails[bestKey]

	for _, i := range sortedErrDetails {
		glb.Debug.Println("-----------------------------")
		glb.Debug.Println("Hyper Params:", allHyperParamDetails[i])
		glb.Debug.Println("Error:", newErrorMap[i])
	}

	for _, i := range sortedErrDetails {
		glb.Debug.Println(allHyperParamDetails[i], " ", newErrorMap[i])
	}

	glb.Debug.Println("Best HyperParameters: ", bestErrHyperParameters)

	return bestErrHyperParameters

	/*sort.Ints(totalErrorList)
	bestResult = totalErrorList[0]
	bestErrHyperParameters := knnErrHyperParameters[bestResult]

	glb.Debug.Println("CrossValidation resuts:")
	for _, res := range totalErrorList {
		glb.Debug.Println(knnErrHyperParameters[res], " : ", res)
	}
	//glb.Debug.Println()
	glb.Debug.Println("Best K : ", bestErrHyperParameters.K)
	glb.Debug.Println("Best MinClusterRss : ", bestErrHyperParameters.MinClusterRss)
	glb.Debug.Println("Minimum error = ", bestResult)

	return bestErrHyperParameters*/
}

func GetBestKnnHyperParamsLegacy(groupName string, shprf dbm.RawSharedPreferences, cd *dbm.ConfigDataStruct, crossValidationPartsList []crossValidationParts) parameters.KnnHyperParameters {
	// CrossValidation
	tempGp := dbm.GM.GetGroup(groupName, false) //permanent:false
	tempGp.Set_ConfigData(cd)

	totalErrorList := []int{}
	knnErrHyperParameters := make(map[int]parameters.KnnHyperParameters)

	bestResult := -1

	//Set algorithm parameters range:

	// KNN:
	// Parameters list creation
	// 1.K
	validKs := glb.MakeRange(glb.DefaultKnnKRange[0], glb.DefaultKnnKRange[1])
	knnKRange := shprf.KnnKRange
	if len(knnKRange) == 1 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[0])
	} else if len(knnKRange) == 2 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[1])
	} else {
		glb.Error.Println("knnKRange:", knnKRange)
		glb.Error.Println("Can't set valid Knn K values")
	}
	//2.MinClusterRSS
	validMinClusterRSSs := glb.MakeRange(glb.DefaultKnnMinCRssRange[0], glb.DefaultKnnMinCRssRange[1])

	minClusterRSSRange := shprf.KnnMinCRssRange
	if len(minClusterRSSRange) == 1 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[0])
	} else if len(minClusterRSSRange) == 2 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[1])
		validMinClusterRSSs = append(validMinClusterRSSs, 0)
	} else {
		glb.Error.Println("minClusterRSSRange:", minClusterRSSRange)
		glb.Error.Println("Can't set valid min cluster rss values")
	}

	// Set length of calculation progress bar
	// This is shared between all threads, so it's invalid when two calculateLearn thread run
	glb.ProgressBarLength = len(validMinClusterRSSs) * len(validKs)

	adTemp := tempGp.NewAlgoDataStruct()
	rdTemp := tempGp.Get_RawData()

	for i, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
		for j, K := range validKs { // for over KnnK
			glb.ProgressBarCurLevel = i*len(validKs) + j
			totalDistError := 0

			tempHyperParameters := parameters.NewKnnHyperParameters()
			//glb.Error.Println(tempHyperParameters)
			tempHyperParameters.K = K
			tempHyperParameters.MinClusterRss = minClusterRss

			// 1-foldCrossValidation (each round one location select as test set)
			for CVNum, CVParts := range crossValidationPartsList {
				glb.Debug.Println("CrossValidation Part num :", CVNum)

				// Learn:
				trainSetTemp := CVParts.GetTrainSet(tempGp)
				rdTemp.Set_Fingerprints(trainSetTemp.Fingerprints)
				rdTemp.Set_FingerprintsOrdering(trainSetTemp.FingerprintsOrdering)
				rdTemp.Set_FingerprintsBackup(trainSetTemp.Fingerprints)
				rdTemp.Set_FingerprintsOrderingBackup(trainSetTemp.FingerprintsOrdering)

				testFPs := CVParts.testSet.Fingerprints
				testFPsOrdering := CVParts.testSet.FingerprintsOrdering

				GetParametersWithGP(tempGp)

				PreProcess(rdTemp, shprf.NeedToRelocateFP)

				learnedKnnData, _ := LearnKnn(tempGp, tempHyperParameters)

				/*			// Set hyper parameters
							learnedKnnData.HyperParameters = parameters.NewKnnHyperParameters()
							learnedKnnData.HyperParameters.K = K
							learnedKnnData.HyperParameters.MinClusterRss = minClusterRss*/

				//learnedKnnData.HyperParameters = parameters.KnnHyperParameters{K: K, MinClusterRss: minClusterRss}

				adTemp.Set_KnnFPs(learnedKnnData)

				// Set to main group
				tempGp.GMutex.Lock() //For each group there's a lock to avoid race between concurrent calculateLearn s
				tempGp.Set_AlgoData(adTemp)

				// Error calculation for this round
				distError := 0
				trackedPointsNum := 0
				testLocation := testFPs[testFPsOrdering[0]].Location
				for _, index := range testFPsOrdering {
					fp := testFPs[index]

					resultDot := ""
					var err error
					err, resultDot, _ = TrackKnn(tempGp, fp, false)
					if err != nil {
						if err.Error() == "NumofAP_lowerThan_MinApNum" {
							glb.Error.Println("NumofAP_lowerThan_MinApNum")
							continue
						} else if err.Error() == "NoValidFingerprints" {
							glb.Error.Println("NoValidFingerprints")
							continue
						}
					} else {
						trackedPointsNum++
					}

					//glb.Debug.Println(fp.Timestamp)
					resx, resy := glb.GetDotFromString(resultDot)
					x, y := glb.GetDotFromString(testLocation)
					//if fp.Timestamp==int64(1516794991872647445){
					//	glb.Error.Println("ResultDot = ",resultDot)
					//	glb.Error.Println("DistError = ",int(calcDist(x,y,resx,resy)))
					//}
					distError += int(glb.CalcDist(x, y, resx, resy))
					if distError < 0 { //print if distError is lower than zero(it's for error detection)
						glb.Error.Println(fp)
						glb.Error.Println(resultDot)
						_, resultDot, _ = TrackKnn(tempGp, fp, false)
						glb.Error.Println(x, y)
						glb.Error.Println(resx, resy)
					}
				}
				if trackedPointsNum == 0 {
					glb.Error.Println("For loc:", testLocation, " there is no fingerprint that its number of APs be more than", glb.MinApNum)
				} else {
					distError = distError / trackedPointsNum
					totalDistError += distError
				}

				tempGp.GMutex.Unlock()
			}

			glb.Debug.Printf("Knn error (minClusterRss=%d,K=%d) = %d \n", minClusterRss, K, totalDistError)

			//if(bestResult==-1 || totalDistError<bestResult){
			//	bestResult = totalDistError
			//	bestK = K
			//}
			totalErrorList = append(totalErrorList, totalDistError)
			knnErrHyperParameters[totalDistError] = tempHyperParameters
		}
	}

	glb.ProgressBarCurLevel = 0 // reset progressBar level

	// Select best hyperParameters
	//glb.Debug.Println(totalErrorList)
	sort.Ints(totalErrorList)
	bestResult = totalErrorList[0]
	bestErrHyperParameters := knnErrHyperParameters[bestResult]

	glb.Debug.Println("CrossValidation resuts:")
	for _, res := range totalErrorList {
		glb.Debug.Println(knnErrHyperParameters[res], " : ", res)
	}
	//glb.Debug.Println()
	glb.Debug.Println("Best K : ", bestErrHyperParameters.K)
	glb.Debug.Println("Best MinClusterRss : ", bestErrHyperParameters.MinClusterRss)
	glb.Debug.Println("Minimum error = ", bestResult)

	return bestErrHyperParameters
}

// return best index from error map
// bestkey is key of minimum error
// sortedErrDetails is sorted keys according to error(first one is bestkey)
// newErrorMap is map of int to int that originate from allErrDetails map according to error method(auc,mean, ...)
func SelectBestFromErrMap(allErrDetails map[int][]int) (int, []int, map[int]int) {
	MainErrAlgorithm := Mean

	bestKey, sortedErrDetails, newErrorMap, err := SelectLowestError(allErrDetails, MainErrAlgorithm)
	if err != nil { // Use Mean algorithm
		glb.Error.Println(err)

		// find best mean error
		bestKey, sortedErrDetails, newErrorMap, err := SelectLowestError(allErrDetails, Mean)
		if err != nil {
			glb.Error.Println(err)
		}
		//bestResult := sortedErrDetails[0]
		return bestKey, sortedErrDetails, newErrorMap
		/*bestErrHyperParameters := knnErrHyperParameters[bestResult]

		glb.Debug.Println("CrossValidation resuts:")
		for _, res := range totalErrorList {
			glb.Debug.Println(knnErrHyperParameters[res], " : ", res)
		}
		//glb.Debug.Println()
		glb.Debug.Println("Best K : ", bestErrHyperParameters.K)
		glb.Debug.Println("Best MinClusterRss : ", bestErrHyperParameters.MinClusterRss)
		glb.Debug.Println("Minimum error = ", bestResult)

		return bestErrHyperParameters*/
	}
	glb.Debug.Println(newErrorMap)
	return bestKey, sortedErrDetails, newErrorMap
	/*	for _, i := range sortedErrDetails {
			glb.Debug.Println("-----------------------------")
			glb.Debug.Println("Hyper Params:", allHyperParamDetails[i])
			glb.Debug.Println("Error:", newErrorMap[i])
		}

		for _, i := range sortedErrDetails {
			glb.Debug.Println(allHyperParamDetails[i], " ", newErrorMap[i])
		}*/
}

func CalculateLearn(groupName string) {
	// Now performance isn't important in learning, just care about performance on track (it helps to code easily!)
	graphDisabled := false
	if glb.GraphEnabled { // Todo: graphEnabled must be declared in sharedPrf
		glb.GraphEnabled = false
		graphDisabled = true
	}

	glb.Debug.Println("################### CalculateLearn ##################")
	gp := dbm.GM.GetGroup(groupName)
	gp.Set_Permanent(false) //for crossvalidation

	rd := gp.Get_RawData()
	mainFPData := rd.Get_FingerprintsBackup()
	mainFPOrdering := rd.Get_FingerprintsOrderingBackup()
	cd := gp.Get_ConfigData()
	ad := gp.Get_AlgoData()
	rs := gp.Get_ResultData()

	knnLocAccuracy := make(map[string]int)
	var crossValidationPartsList []crossValidationParts
	glb.Debug.Println(mainFPOrdering)
	glb.Debug.Println(mainFPData)
	crossValidationPartsList = GetCrossValidationParts(gp, mainFPOrdering, mainFPData)
	// ToDo: Need to learn algorithms concurrently

	glb.Debug.Println("crossValidationPartsList length:", len(crossValidationPartsList))
	shprf := dbm.GetSharedPrf(groupName)

	//bestKnnHyperParams := GetBestKnnHyperParams(groupName+"_KNNTemp", shprf, cd, crossValidationPartsList)
	bestKnnHyperParams := GetBestKnnHyperParamsLegacy(groupName+"_KNNTemp", shprf, cd, crossValidationPartsList)

	glb.Debug.Println("BestHyperParameters before testvalid track recalculation :", bestKnnHyperParams)

	// Calculating each location detection accuracy with best hyperParameters: //todo:Avoid this extra level,do it in GetBestKnnHyperParams
	knnDistError := 0
	numLocCrossed := 0
	for CVNum, CVParts := range crossValidationPartsList {
		glb.Debug.Println("CrossValidation Part num :", CVNum)

		// Learn
		trainSetTemp := CVParts.GetTrainSet(gp)
		rd.Set_Fingerprints(trainSetTemp.Fingerprints)
		rd.Set_FingerprintsOrdering(trainSetTemp.FingerprintsOrdering)
		rd.Set_FingerprintsBackup(trainSetTemp.Fingerprints)
		rd.Set_FingerprintsOrderingBackup(trainSetTemp.FingerprintsOrdering)
		testFPs := CVParts.testSet.Fingerprints
		testFPsOrdering := CVParts.testSet.FingerprintsOrdering
		GetParametersWithGP(gp)

		PreProcess(rd, shprf.NeedToRelocateFP)

		learnedKnnData, _ := LearnKnn(gp, bestKnnHyperParams)
		// Set hyper parameters
		learnedKnnData.HyperParameters = bestKnnHyperParams
		gp.GMutex.Lock()
		ad.Set_KnnFPs(learnedKnnData)
		gp.Set_AlgoData(ad)

		// Error calculation for each location with best hyperParameters
		distError := 0
		trackedPointsNum := 0
		testLocation := testFPs[testFPsOrdering[0]].Location
		for _, index := range testFPsOrdering {

			fp := testFPs[index]
			resultDot := ""
			var err error
			err, resultDot, _ = TrackKnn(gp, fp, false)

			if err != nil {
				if err.Error() == "NumofAP_lowerThan_MinApNum" {
					continue
				} else if err.Error() == "NoValidFingerprints" {
					continue
				}
			}
			//glb.Debug.Println(fp.Location, " ==== ", resultDot)

			resx, resy := glb.GetDotFromString(resultDot)
			x, y := glb.GetDotFromString(testLocation) // testLocation is fp.Location
			distErrorTemp := int(glb.CalcDist(x, y, resx, resy))
			if distErrorTemp < 0 {
				glb.Error.Println(fp)
				glb.Error.Println(resultDot)
				//_, resultDot, _ = TrackKnn(gp, fp, false)
				glb.Error.Println(x, y)
				glb.Error.Println(resx, resy)
			} else {
				distError += distErrorTemp
				trackedPointsNum++
			}
		}

		if trackedPointsNum == 0 {
			glb.Error.Println("For loc:", testLocation, " there is no fingerprint that its number of APs be more than", glb.MinApNum)
			knnLocAccuracy[testLocation] = -1
		} else {
			distError = distError / trackedPointsNum
			knnLocAccuracy[testLocation] = distError
			knnDistError += distError
			numLocCrossed++
		}
		gp.GMutex.Unlock()
	}

	// Set CrossValidation results
	if numLocCrossed == 0 {
		glb.Error.Println("numLocCrossed is zero ")
	} else {
		rs.Set_AlgoAccuracy("knn", knnDistError/numLocCrossed)
	}
	for loc, accuracy := range knnLocAccuracy {
		rs.Set_AlgoLocAccuracy("knn", loc, accuracy)
	}
	glb.Debug.Println(dbm.GetCVResults(groupName))

	// Set main parameters: Note: Don't change mainFpData
	rd.Set_Fingerprints(mainFPData)
	rd.Set_FingerprintsOrdering(mainFPOrdering)
	rd.Set_FingerprintsBackup(mainFPData)
	rd.Set_FingerprintsOrderingBackup(mainFPOrdering)
	GetParametersWithGP(gp)

	gp.GMutex.Lock()
	glb.Debug.Println(gp.Get_MiddleData().Get_UniqueMacs())

	PreProcess(rd, shprf.NeedToRelocateFP)

	// learn algorithm
	learnedKnnData, _ := LearnKnn(gp, bestKnnHyperParams)
	//learnedKnnData.HyperParameters = bestKnnHyperParams

	// Set best hyperparameter values
	ad.Set_KnnFPs(learnedKnnData)
	gp.Set_Permanent(true)
	gp.GMutex.Unlock()

	if glb.RuntimeArgs.Scikit {
		ScikitLearn(groupName)
	}

	glb.Debug.Println("Calculation finished.")
	//if glb.RuntimeArgs.Svm {
	//	DumpFingerprintsSVM(groupName)
	//	err := CalculateSVM(groupName)
	//	if err != nil {
	//		glb.Warning.Println("Encountered error when calculating SVM")
	//		glb.Warning.Println(err)
	//	}
	//}
	if graphDisabled {
		glb.GraphEnabled = true
	}
	//runnerLock.Unlock()
}

func CalculateGraphFactor(groupName string) {
	gp := dbm.GM.GetGroup(groupName)
	rsd := gp.Get_ResultData()
	ad := gp.Get_AlgoData()
	knnFPs := ad.Get_KnnFPs()

	//GraphFactors range:
	//validGraphFactorsRange = [][]float64{}
	//validGraphFactors := [][]float64{{1, 0, 0, 0}}
	//validGraphFactors := [][]float64{{2, 1, 1, 1}, {100, 100, 100, 100, 3, 2, 1}, {10, 10, 10, 10, 3, 2, 1}, {8, 8, 8, 8, 3, 2, 1}, {1, 1, 1, 1}}

	beginSlice := []float64{1, 0, 0, 0,0,0,0}
	endSlice := []float64{10,10,10,10,10,10,10}
	step := 1.0
	validGraphFactors := glb.GetGraphSlicesRangeRecursive(beginSlice, endSlice, step)



	//if len(validGraphFactorsRange) == 1{
	//
	//}else if len(validGraphFactorsRange) == 2{
	//
	//}else{
	//	glb.Error.Println("Can't set valid graph factors values")
	//}

	testValidTracks := rsd.Get_TestValidTracks()

	// Set parameters with testvalid tracks
	//reset temp params
	knnHyperParams := knnFPs.HyperParameters

	glb.Debug.Println("bestKnnHyperParams before CalculateGraphFactor :", knnHyperParams)
	// 1. TestValid without graph:
	/*	testValidTrueLoc := []string{}
		testvalidGuessLoc := []string{}*/
	glb.GraphEnabled = false
	{
		learnedKnnData, _ := LearnKnn(gp, knnHyperParams)
		ad.Set_KnnFPs(learnedKnnData)

		allErrDetails := make(map[int][]int)

		tempAllErrDetailList := []int{}
		/*totalDistError := 0
		trackedPointsNum := 0*/
		gp.Get_ResultData().Set_UserHistory(glb.TesterUsername, []parameters.UserPositionJSON{}) // clear userhistory to check knn error with new graphfactor

		for _, testValidTrack := range testValidTracks {
			fp := testValidTrack.UserPosition.Fingerprint
			testLocation := testValidTrack.TrueLocation

			resultDot := RecalculateTrackFingerprintKnnCrossValidation(fp)

			/*			testValidTrueLoc = append(testValidTrueLoc, testLocation)
						testvalidGuessLoc = append(testvalidGuessLoc, resultDot)*/

			resx, resy := glb.GetDotFromString(resultDot)
			x, y := glb.GetDotFromString(testLocation)
			tempDistError := int(glb.CalcDist(x, y, resx, resy))
			if tempDistError < 0 { //print if totalDistError is lower than zero(it's for error detection)
				glb.Error.Println(fp)
				glb.Error.Println(resultDot)
				glb.Error.Println("totalDistError is negetive!")
			} else {
				/*trackedPointsNum++
				totalDistError += tempDistError*/
				tempAllErrDetailList = append(tempAllErrDetailList, tempDistError)

				//glb.Debug.Println("###########")
				//glb.Debug.Println(fp)
				//glb.Debug.Println(testLocation)
				//glb.Debug.Println(resultDot)
				//glb.Debug.Println(tempDistError)
			}
		}
		allErrDetails[0] = tempAllErrDetailList
		_, _, newErrorMap := SelectBestFromErrMap(allErrDetails)
		glb.Debug.Println("testvalid error without graph:", newErrorMap[0])
		rsd.Set_AlgoAccuracy("knn_testvalid", newErrorMap[0])
	}
	glb.GraphEnabled = true
	/*	glb.Error.Println("testValidTrueLoc",testValidTrueLoc)
		glb.Error.Println("testvalidGuessLoc",testvalidGuessLoc)*/

	// 2. Testvalid with graph:
	//totalErrorList := []int{}
	//knnErrHyperParameters := make(map[int]parameters.KnnHyperParameters)
	//bestResult := -1

	// graphfactor used in online phase so learning must be done one time
	learnedKnnData, _ := LearnKnn(gp, knnHyperParams)

	allHyperParamDetails := make(map[int]parameters.KnnHyperParameters)
	allErrDetails := make(map[int][]int)
	paramUniqueKey := 0 // just creating unique key for each possible the parameters permutation
	for _, graphFactor := range validGraphFactors { // for over the validGraphFactors
		gp.Get_ResultData().Set_UserHistory(glb.TesterUsername, []parameters.UserPositionJSON{}) // clear userhistory to check knn error with new graphfactor

		allHyperParamDetails[0] = learnedKnnData.HyperParameters

		tempHyperParameters := knnHyperParams
		tempHyperParameters.GraphFactors = graphFactor
		learnedKnnData.HyperParameters = tempHyperParameters
		ad.Set_KnnFPs(learnedKnnData)

		paramUniqueKey++
		allHyperParamDetails[paramUniqueKey] = tempHyperParameters
		tempAllErrDetailList := []int{}

		//totalDistError := 0
		trackedPointsNum := 0
		for _, testValidTrack := range testValidTracks {
			fp := testValidTrack.UserPosition.Fingerprint
			testLocation := testValidTrack.TrueLocation
			resultDot := RecalculateTrackFingerprintKnnCrossValidation(fp)

			/*	if (testLocation=="-428.0,906.0"){
					glb.Debug.Println("-33.0,946.0 resultdot:",resultDot)
					glb.Debug.Println(fp)
				}
	*/
			resx, resy := glb.GetDotFromString(resultDot)
			x, y := glb.GetDotFromString(testLocation)
			tempDistError := int(glb.CalcDist(x, y, resx, resy))
			if tempDistError < 0 { //print if totalDistError is lower than zero(it's for error detection)
				glb.Error.Println(fp)
				glb.Error.Println(resultDot)
				glb.Error.Println("totalDistError is negetive!")
			} else {
				trackedPointsNum++
				//totalDistError += tempDistError
				tempAllErrDetailList = append(tempAllErrDetailList, tempDistError)
			}
		}

		if trackedPointsNum == 0 {
			glb.Error.Println("trackedPointsNum is zero!")
		} else {
			/*			totalDistError = totalDistError / trackedPointsNum
						totalErrorList = append(totalErrorList, totalDistError)
						knnErrHyperParameters[totalDistError] = tempHyperParameters*/
		}
		allErrDetails[paramUniqueKey] = tempAllErrDetailList
		glb.Debug.Println("Calculation for graphfactor=", graphFactor, " ended ", )
	}

	/*	sort.Ints(totalErrorList)
		bestResult = totalErrorList[0]
		bestErrHyperParameters := knnErrHyperParameters[bestResult]*/

	bestKey, sortedErrDetails, newErrorMap := SelectBestFromErrMap(allErrDetails)
	bestErrHyperParameters := allHyperParamDetails[bestKey]
	bestResult := newErrorMap[bestKey]
	glb.Debug.Println("%%%%%%% best result %%%%%%",bestResult)

	for _, i := range sortedErrDetails {
		glb.Debug.Println("-----------------------------")
		glb.Debug.Println("Hyper Params:", allHyperParamDetails[i])
		glb.Debug.Println("Error:", newErrorMap[i])
	}

	for _, i := range sortedErrDetails {
		glb.Debug.Println(allHyperParamDetails[i], " ", newErrorMap[i])
	}

	glb.Debug.Println("Best HyperParameters: ", bestErrHyperParameters)

	knnFPs.HyperParameters = bestErrHyperParameters
	glb.Debug.Println("Best GraphFactors : ", bestErrHyperParameters.GraphFactors)

	ad.Set_KnnFPs(knnFPs)
	rsd.Set_AlgoAccuracy("knn_testvalid_graph", bestResult)

}

// Error calculation using mean as error algorithm
func CalculateGraphFactorLegacy(groupName string) {
	gp := dbm.GM.GetGroup(groupName)
	rsd := gp.Get_ResultData()
	ad := gp.Get_AlgoData()
	knnFPs := ad.Get_KnnFPs()

	//GraphFactors range:
	//validGraphFactorsRange = [][]float64{}
	validGraphFactors := [][]float64{{0.5, 0.5, 1, 1, 1}, {2, 1, 1, 1}, {100, 100, 100, 100, 3, 2, 1}, {10, 10, 10, 10, 3, 2, 1}, {8, 8, 8, 8, 3, 2, 1}, {1, 1, 1, 1}}

	//if len(validGraphFactorsRange) == 1{
	//
	//}else if len(validGraphFactorsRange) == 2{
	//
	//}else{
	//	glb.Error.Println("Can't set valid graph factors values")
	//}

	testValidTracks := rsd.Get_TestValidTracks()

	// Set parameters with testvalid tracks
	//reset temp params
	knnHyperParams := knnFPs.HyperParameters

	glb.Debug.Println("bestKnnHyperParams before CalculateGraphFactor :", knnHyperParams)
	// 1. TestValid without graph:
	/*	testValidTrueLoc := []string{}
		testvalidGuessLoc := []string{}*/
	glb.GraphEnabled = false
	{
		learnedKnnData, _ := LearnKnn(gp, knnHyperParams)
		ad.Set_KnnFPs(learnedKnnData)

		totalDistError := 0
		trackedPointsNum := 0
		gp.Get_ResultData().Set_UserHistory(glb.TesterUsername, []parameters.UserPositionJSON{}) // clear userhistory to check knn error with new graphfactor

		for _, testValidTrack := range testValidTracks {
			fp := testValidTrack.UserPosition.Fingerprint
			testLocation := testValidTrack.TrueLocation

			resultDot := RecalculateTrackFingerprintKnnCrossValidation(fp)

			/*			testValidTrueLoc = append(testValidTrueLoc, testLocation)
						testvalidGuessLoc = append(testvalidGuessLoc, resultDot)*/

			resx, resy := glb.GetDotFromString(resultDot)
			x, y := glb.GetDotFromString(testLocation)
			distErrorTemp := int(glb.CalcDist(x, y, resx, resy))
			if distErrorTemp < 0 { //print if totalDistError is lower than zero(it's for error detection)
				glb.Error.Println(fp)
				glb.Error.Println(resultDot)
				glb.Error.Println("totalDistError is negetive!")
			} else {
				trackedPointsNum++
				totalDistError += distErrorTemp
				//glb.Debug.Println("###########")
				//glb.Debug.Println(fp)
				//glb.Debug.Println(testLocation)
				//glb.Debug.Println(resultDot)
				//glb.Debug.Println(distErrorTemp)
			}
		}
		glb.Error.Println("testvalid error without graph:", totalDistError/trackedPointsNum)
		rsd.Set_AlgoAccuracy("knn_testvalid", totalDistError/trackedPointsNum)
	}
	glb.GraphEnabled = true
	/*	glb.Error.Println("testValidTrueLoc",testValidTrueLoc)
		glb.Error.Println("testvalidGuessLoc",testvalidGuessLoc)*/

	// 2. Testvalid with graph:
	totalErrorList := []int{}
	knnErrHyperParameters := make(map[int]parameters.KnnHyperParameters)
	bestResult := -1

	// graphfactor used in online phase so learning must be done one time
	learnedKnnData, _ := LearnKnn(gp, knnHyperParams)

	for _, graphFactor := range validGraphFactors { // for over the validGraphFactors
		gp.Get_ResultData().Set_UserHistory(glb.TesterUsername, []parameters.UserPositionJSON{}) // clear userhistory to check knn error with new graphfactor

		tempHyperParameters := knnHyperParams
		tempHyperParameters.GraphFactors = graphFactor

		learnedKnnData.HyperParameters = tempHyperParameters
		ad.Set_KnnFPs(learnedKnnData)

		totalDistError := 0
		trackedPointsNum := 0
		for _, testValidTrack := range testValidTracks {
			fp := testValidTrack.UserPosition.Fingerprint
			testLocation := testValidTrack.TrueLocation
			resultDot := RecalculateTrackFingerprintKnnCrossValidation(fp)

			/*	if (testLocation=="-428.0,906.0"){
					glb.Debug.Println("-33.0,946.0 resultdot:",resultDot)
					glb.Debug.Println(fp)
				}
	*/
			resx, resy := glb.GetDotFromString(resultDot)
			x, y := glb.GetDotFromString(testLocation)
			distErrorTemp := int(glb.CalcDist(x, y, resx, resy))
			if distErrorTemp < 0 { //print if totalDistError is lower than zero(it's for error detection)
				glb.Error.Println(fp)
				glb.Error.Println(resultDot)
				glb.Error.Println("totalDistError is negetive!")
			} else {
				trackedPointsNum++
				totalDistError += distErrorTemp
			}
		}

		if trackedPointsNum == 0 {
			glb.Error.Println("trackedPointsNum is zero!")
		} else {
			totalDistError = totalDistError / trackedPointsNum
			totalErrorList = append(totalErrorList, totalDistError)
			knnErrHyperParameters[totalDistError] = tempHyperParameters
		}
		glb.Debug.Println("Knn error (graphfactor=", graphFactor, ") =", totalDistError)
	}

	sort.Ints(totalErrorList)
	bestResult = totalErrorList[0]
	bestErrHyperParameters := knnErrHyperParameters[bestResult]

	knnFPs.HyperParameters = bestErrHyperParameters
	glb.Debug.Println("Best GraphFactors : ", bestErrHyperParameters.GraphFactors)

	ad.Set_KnnFPs(knnFPs)
	rsd.Set_AlgoAccuracy("knn_testvalid_graph", bestResult)

}

// Find best key that has lowest error list:
// method : "FalseRate", "AUC" ,"LatterPercentile","Mean"
func SelectLowestError(errMap map[int][]int, method string) (int, []int, map[int]int, error) {
	if method == FalseRate {
		trueMaxErr := 100 // 100 cm
		falseCountMap := make(map[int]int)

		for key, errList := range errMap {
			falseCount := 0
			sort.Ints(errList)
			for _, err := range errList {
				if err > trueMaxErr {
					falseCount++
				}
			}
			falseCountMap[key] = falseCount
		}
		sortedKey := glb.SortIntKeyDictByIntVal(falseCountMap)
		return sortedKey[0], sortedKey, falseCountMap, nil
	} else if method == AUC {
		falseCountMap := make(map[int]int)

		//latterPercentile := 0.90
		trueMaxErr := 300 // 100 cm
		trueMinErr := 50  // 100 cm
		step := 10

		for key, errList := range errMap {
			allStepFalseCount := 0
			sort.Ints(errList)

			//latterPercentileIndex := int(float64(errListLen)*latterPercentile)
			//latterPercentileVal := errList[latterPercentileIndex]
			for tempTrueMaxErr := trueMinErr; tempTrueMaxErr <= trueMaxErr; tempTrueMaxErr += step {
				eachStepFalseCount := 0
				for _, err := range errList {
					if err > tempTrueMaxErr {
						eachStepFalseCount++
					}
				}
				allStepFalseCount += eachStepFalseCount
			}
			falseCountMap[key] = allStepFalseCount
		}
		sortedKey := glb.SortIntKeyDictByIntVal(falseCountMap)
		return sortedKey[0], sortedKey, falseCountMap, nil
	} else if method == LatterPercentile {
		latterPercentile := 0.75

		percentileMap := make(map[int]int)
		errListLen := 0

		// find error list length
		for key := range errMap {
			if errListLen == 0 {
				errListLen = len(errMap[key])
			} else {
				break
			}
		}
		for key, errList := range errMap {
			sort.Ints(errList)
			latterPercentileIndex := int(float64(errListLen) * latterPercentile)
			latterPercentileVal := errList[latterPercentileIndex]
			percentileMap[key] = latterPercentileVal
		}
		sortedKey := glb.SortIntKeyDictByIntVal(percentileMap)
		return sortedKey[0], sortedKey, percentileMap, nil
	} else if method == Mean {
		meanErrMap := make(map[int]int)
		for key, errList := range errMap {
			mean := 0
			for _, err := range errList {
				mean += err
			}
			mean /= len(errList)
			meanErrMap[key] = mean
		}
		sortedKey := glb.SortIntKeyDictByIntVal(meanErrMap)
		return sortedKey[0], sortedKey, meanErrMap, nil
	}
	return -1, []int{}, make(map[int]int), errors.New("Invalid method parameters")
}
