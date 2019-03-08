package algorithms

import (
	"ParsinServer/algorithms/clustering"
	"ParsinServer/algorithms/particlefilter"
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

//
const (
	AUC              string = "AUC"
	FalseRate        string = "FalseRate"
	Mean             string = "Mean"
	LatterPercentile string = "LatterPercentile"
)

// Trackfingerprint options
const (
	IgnoreSimpleHistoryEffect string = "IgnoreSimpleHistoryEffect"
)

var threadedCross bool = false //don't use, it's not safe now!

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
		glb.Debug.Println("Track json: ", fp)
		if !glb.IsValidXY(fp.Location) {
			// Mobile PDR data isn't available
			fp.Location = "" // this value check in TrackFingerprint, non-empty fp.Location is considered as PDR data
		}
		userPosJson, success, message := TrackOnlineFingerprint(fp)

		knnGuess := userPosJson.Location // todo: this is done for compatability ,use 'location' instead of knn
		if success {
			scikitDataStr := glb.StringMap2String(userPosJson.ScikitData)
			resJsonMap := gin.H{"message": message, "success": true, "location": userPosJson.Location, "bayes": userPosJson.BayesGuess, "svm": userPosJson.SvmGuess, "knn": knnGuess, "accuracyCircleRadius": userPosJson.Confidentiality}
			for algorithm, valXY := range userPosJson.ScikitData {
				resJsonMap[algorithm] = valXY
			}

			glb.Debug.Println("message", message, " success", true, " location", userPosJson.Location, " bayes", userPosJson.BayesGuess, " svm", userPosJson.SvmGuess, scikitDataStr, " knn", knnGuess, " accuracyCircleRadius", userPosJson.Confidentiality)
			c.JSON(http.StatusOK, resJsonMap)
		} else {
			glb.Warning.Println(message)
			c.JSON(http.StatusOK, gin.H{"message": message, "success": false})
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
/*func TrackFingerprint(curFingerprint parameters.Fingerprint) (string, bool, string, string, map[string]float64, string, map[string]float64, string, map[string]string, float64) {
	// Classify with filter curFingerprint
	//fullFingerprint := curFingerprint
	parameters.CleanFingerprint(&curFingerprint)
	filteredCurFingerprint := dbm.FilterFingerprint(curFingerprint) // Use this as algorithm inputs, also curFingerprint saved as main fingerprint in UserJson


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

	if !dbm.GroupExists(curFingerprint.Group) || len(curFingerprint.Group) == 0 {
		return "You should insert fingerprints before tracking", false, "", "", bayesData, "", make(map[string]float64), "", make(map[string]string), float64(0)
	}
	if len(curFingerprint.WifiFingerprint) == 0 || len(filteredCurFingerprint.WifiFingerprint) == 0 {
		return "No fingerprints found to track, see API", false, "", "", bayesData, "", make(map[string]float64), "", make(map[string]string), float64(0)
	}
	if len(curFingerprint.Username) == 0 {
		return "No username defined, see API", false, "", "", bayesData, "", make(map[string]float64), "", make(map[string]string), float64(0)
	}

	//wasLearning, ok := dbm.GetLearningCache(strings.ToLower(curFingerprint.Group))
	//	if ok {
	//		if wasLearning {
	//			glb.Debug.Println("Was learning, calculating priors")
	//
	//			go dbm.SetLearningCache(groupName, false)
	//			//bayes.OptimizePriorsThreaded(groupName)
	//			//if glb.RuntimeArgs.Svm {
	//			//	DumpFingerprintsSVM(groupName)
	//			//	CalculateSVM(groupName)
	//			//}
	//			//if glb.RuntimeArgs.Scikit {
	//			//	ScikitLearn(groupName)
	//			//}
	//			//LearnKnn(groupName)
	//			CalculateLearn(groupName)
	//			go dbm.AppendUserCache(groupName, curFingerprint.Username)
	//		}
	//	}

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
	cd := gp.Get_ConfigData()
	//knnConfig := cd.Get_KnnConfig()
	otherGpConfig := cd.Get_OtherGroupConfig()
	// Calculating KNN
	//glb.Debug.Println(curFingerprint)
	err, knnGuess, knnData := TrackKnn(gp, filteredCurFingerprint, true)
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
		scikitData = ScikitClassify(groupName, filteredCurFingerprint)
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
	location = knnGuess
	if otherGpConfig.SimpleHistoryEnabled {
		userHistory := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)
		location, accuracyCircleRadius = SimpleHistoryEffect(userJSON, userHistory)
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
*/

func TrackFingerprint(curFingerprint parameters.Fingerprint, options ...string) (parameters.UserPositionJSON, bool, string) {
	doAllAlgorithm := true
	mainAlgo := glb.MainPositioningAlgo

	if len(options) > 0 {
		for _, ALGO := range glb.ALLALGORITHMS { // Set mainAlgo if was set by options
			// First algorithm is set as mainAlgo
			if options[0] == ALGO {
				mainAlgo = ALGO
				doAllAlgorithm = false
				break
			}
		}
	}

	// Classify with filter curFingerprint
	filteredCurFingerprint := dbm.FilterFingerprint(curFingerprint) // Use this as algorithm inputs, also curFingerprint saved as main fingerprint in UserJson

	if (len(curFingerprint.WifiFingerprint) < glb.MinApNum) {
		glb.Error.Println("For testValidTrack:", curFingerprint.Timestamp, " there is no fingerprint that its number of APs be more than", glb.MinApNum)
		return parameters.UserPositionJSON{}, false, "NumofAP_lowerThan_MinApNum"
	}
	if !dbm.GroupExists(curFingerprint.Group) || len(curFingerprint.Group) == 0 {
		glb.Error.Println("You should insert fingerprints before tracking")
		return parameters.UserPositionJSON{}, false, "You should insert fingerprints before tracking"
	}
	if len(curFingerprint.WifiFingerprint) == 0 || len(filteredCurFingerprint.WifiFingerprint) == 0 {
		glb.Error.Println("No fingerprints(maybe after filtering) were found to track, see API")
		return parameters.UserPositionJSON{}, false, "No fingerprints(maybe after filtering) were found to track, see API"
	}
	if len(curFingerprint.Username) == 0 {
		glb.Error.Println("No username was defined, see API")
		return parameters.UserPositionJSON{}, false, "No username was defined, see API"
	}

	groupName := curFingerprint.Group
	gp := dbm.GM.GetGroup(groupName)
	cd := gp.Get_ConfigData()
	otherGpConfig := cd.Get_OtherGroupConfig()


	bayesGuess := ""
	bayesData := make(map[string]float64)
	svmGuess := ""
	svmData := make(map[string]float64)
	scikitData := make(map[string]string)
	knnGuess := ""
	knnData := make(map[string]float64)
	message := ""
	rawLocation := ""
	location := ""
	accuracyCircleRadius := float64(0)

	// Calculating KNN
	if doAllAlgorithm || glb.StringInSlice(glb.KNN, options) {
		var err error
		err, knnGuess, knnData = TrackKnn(gp, filteredCurFingerprint, true)
		if err != nil {
			glb.Error.Println(err)
		}
		message += " knnGuess: " + knnGuess
	}

	// Calculating Scikit

	scikitSelected := glb.StringInSlice(glb.SCIKIT_CLASSIFICATION, options) || glb.StringInSlice(glb.SCIKIT_REGRESSION, options)
	if glb.RuntimeArgs.Scikit && (doAllAlgorithm || scikitSelected) {
		scikitData = ScikitClassify(groupName, filteredCurFingerprint)
		glb.Debug.Println(scikitData)
		for algorithm, valueXY := range scikitData {
			message += " " + algorithm + ":" + valueXY
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
	userPosJson.Fingerprint = curFingerprint

	// Set location to the main algorithm result
	switch mainAlgo {
	case glb.KNN:
		rawLocation = knnGuess
	case glb.BAYES:
		rawLocation = bayesGuess
	case glb.SVM:
		rawLocation = svmGuess
	case glb.SCIKIT_REGRESSION:
		rawLocation = scikitData[glb.SCIKIT_REGRESSION]
	case glb.SCIKIT_CLASSIFICATION:
		rawLocation = scikitData[glb.SCIKIT_CLASSIFICATION]
	}
	userPosJson.RawLocation = rawLocation


	// User history effect
	location = rawLocation
	if otherGpConfig.SimpleHistoryEnabled && !glb.StringInSlice(IgnoreSimpleHistoryEffect, options) {
		userHistory := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)
		location, accuracyCircleRadius = SimpleHistoryEffect(userPosJson, userHistory)
		//location, accuracyCircleRadius = HistoryEffectStaticFactors(userPosJson, userHistory)
	}

	userPosJson.Location = location
	userPosJson.Confidentiality = accuracyCircleRadius

	glb.Debug.Println(userPosJson)
	return userPosJson, true, message
}

// When online track fingerprint was sent to server(from a cellphone) this function calculate the result and returns it instantly
func TrackOnlineFingerprint(curFingerprint parameters.Fingerprint) (parameters.UserPositionJSON, bool, string) {
	parameters.CleanFingerprint(&curFingerprint)

	pdrLocation := curFingerprint.Location // When raw fingerprint is got from cellphone, pdrLocation is in location field
	userPosJson, success, message := TrackFingerprint(curFingerprint)
	userPosJson.PDRLocation = pdrLocation

	if success {
		gp := dbm.GM.GetGroup(curFingerprint.Group)
		go dbm.SetUserPositionCache(strings.ToLower(curFingerprint.Group)+strings.ToLower(curFingerprint.Username), userPosJson)
		go gp.Get_ResultData().Append_UserHistory(curFingerprint.Username, userPosJson)
		go gp.Get_ResultData().Append_UserResults(curFingerprint.Username, userPosJson)

		// Add fingerprint as a test-valid fp to db
		if curFingerprint.TestValidation && curFingerprint.Username == glb.TesterUsername {
			glb.Debug.Println("TestValidTrack added ")
			tempTestValidTrack := parameters.TestValidTrack{TrueLocation: userPosJson.Fingerprint.Location, UserPosition: userPosJson}
			go gp.Get_ResultData().Append_TestValidTracks(tempTestValidTrack)
		}
	}
	return userPosJson, success, message
}

func Fusion(curUsrPos parameters.UserPositionJSON, gpUsrHistory, coGpUsrHistory []parameters.UserPositionJSON) string {

	if len(coGpUsrHistory) == 0 {
		return curUsrPos.Location
	}
	atmostTimeDiff := int64(6000)
	coGpLastPos := coGpUsrHistory[len(coGpUsrHistory)-1]
	timestampDiff := curUsrPos.Time - coGpLastPos.Time
	//glb.Error.Println(curUsrPos.Time)
	//glb.Error.Println(coGpLastPos.Time)

	if -1*atmostTimeDiff < timestampDiff && timestampDiff < atmostTimeDiff {
		glb.Error.Println(timestampDiff)
		loc1 := curUsrPos.Location
		loc2 := coGpLastPos.Location
		x1, y1 := glb.GetDotFromString(loc1)
		x2, y2 := glb.GetDotFromString(loc2)
		xt := (x1 + x2) / 2
		yt := (y1 + y2) / 2
		return glb.FloatToString(xt) + "," + glb.FloatToString(yt)
		//return coGpLastPos.Location
	} else {
		return curUsrPos.Location
	}
}

func ParticleFilterFusion(curUsrPos parameters.UserPositionJSON, timeDiffWithLastFP int64, CoGroupMode int) string {
	//Todo: Now you need to swap x and y(solve this problem)
	glb.Debug.Println("Using Particle filter ...")
	groupName := curUsrPos.Fingerprint.Group

	timestamp := curUsrPos.Time
	loc := curUsrPos.Location

	x, y := glb.GetDotFromString(loc)

	//CoGroupMode

	if timeDiffWithLastFP == 0 {
		glb.Debug.Println("Initialization Particle filter")

		// Providing map graph
		masterGroupName := groupName
		if CoGroupMode == parameters.CoGroupState_Slave { // if group is slave get its master group name
			coGpName := dbm.GM.GetGroup(groupName).Get_ConfigData().Get_OtherGroupConfig().CoGroup
			masterGroupName = coGpName
		}

		// Get graph
		mapGraph := dbm.GM.GetGroup(masterGroupName).Get_ConfigData().Get_GroupGraph()

		// Get all lines(walls)
		floatAllLines := glb.Convert2DimStringSliceTo3DFloat32(mapGraph.AllLines())

		// Convet map(3DFloat slice) to protobuf format
		mapGraphProtobufStyle := particlefilter.GetMapGraph(floatAllLines)

		// Initialize Particle filter
		particlefilter.Initialize(timestamp, []float32{float32(y), float32(x)}, mapGraphProtobufStyle)

		return curUsrPos.Location
	} else {
		diffTime := timeDiffWithLastFP
		lastPredictionTimestamp := timestamp - timeDiffWithLastFP

		MaxTimeDiffForUpdate := int64(2000)
		PredictionPerdiod := int64(1000)
		for diffTime > MaxTimeDiffForUpdate {
			diffTime -= PredictionPerdiod
			lastPredictionTimestamp += PredictionPerdiod
			resultXY := particlefilter.Predict(lastPredictionTimestamp)
			glb.Debug.Println("Only Prediction result:", resultXY)
		}

		// GET TEST VALID TRUE LOCATION:
		//////////////////////////////////////////////////
		// // JUST FOR TEST
		trueLocX, trueLocY := float64(0), float64(0)
		gp := dbm.GM.GetGroup(curUsrPos.Fingerprint.Group)
		// 1.FOR UWB Based TRUE LOCATIONS
		rd := gp.Get_RawData()
		allLocationLogs := rd.Get_TestValidTrueLocations()
		if len(allLocationLogs) != 0 {

			glb.Debug.Println(allLocationLogs)
			allLocationLogsOrdering := rd.Get_TestValidTrueLocationsOrdering()
			//glb.Error.Println("TrueLocationLog :", allLocationLogsOrdering)
			TrueLoc, _, err := FindTrueFPloc(curUsrPos.Fingerprint, allLocationLogs, allLocationLogsOrdering)
			if err != nil {
				glb.Error.Println(err)
				trueLocX, trueLocY = -100000000, -100000000
			} else {
				trueLocX, trueLocY = glb.GetDotFromString(TrueLoc)
			}
		} else {
		// 2. FOR CELLPHONE Based TRUE LOCATIONS
			rsd := gp.Get_ResultData()
			//trueLocX, trueLocY := 0.0, 0.0
			testValidTracks := rsd.Get_TestValidTracks()
			for _, testvalidTrack := range testValidTracks {
				if testvalidTrack.UserPosition.Time == curUsrPos.Time {
					glb.Debug.Println("####################")
					glb.Debug.Println(curUsrPos.Time)
					if testvalidTrack.TrueLocation == "" {
						glb.Error.Println(testvalidTrack.UserPosition.Fingerprint.Group)
						glb.Error.Println(testvalidTrack.UserPosition.Fingerprint.Location)
						trueLocX, trueLocY = glb.GetDotFromString(testvalidTrack.UserPosition.Fingerprint.Location)
					} else {
						glb.Error.Println(testvalidTrack.TrueLocation)
						trueLocX, trueLocY = glb.GetDotFromString(testvalidTrack.TrueLocation)
					}
					//glb.Error.Println(testvalidTrack.UserPosition.Fingerprint.Location)
					//trueLocX, trueLocY = glb.GetDotFromString(testvalidTrack.TrueLocation)
				}
			}
		}
		//////////////////////////////////////////////

		//resultXY := particlefilter.Update(timestamp, []float32{float32(y), float32(x), float32(trueLocY), float32(trueLocX)})
		//trueLocY, trueLocX := float32(0),float32(0)

		resultXY := []float32{}
		if CoGroupMode == parameters.CoGroupState_Master || CoGroupMode == parameters.CoGroupState_None {
			glb.Debug.Println("Updating by measurement 	... ")
			resultXY = particlefilter.Update(timestamp, []float32{float32(y), float32(x)}, []float32{}, []float32{float32(trueLocY), float32(trueLocX)})
		} else if CoGroupMode == parameters.CoGroupState_Slave {
			glb.Debug.Println("Updating by Co-Group measurement ... ")
			resultXY = particlefilter.Update(timestamp, []float32{}, []float32{float32(y), float32(x)}, []float32{float32(trueLocY), float32(trueLocX)})
		}

		if len(resultXY) == 0 {
			glb.Error.Println("Particle filter update not working well")
		}

		glb.Error.Println("################")
		glb.Error.Println(loc)
		glb.Error.Println(glb.FloatToString(float64(int(resultXY[1]))) + "," + glb.FloatToString(float64(int(resultXY[0]))))
		return glb.FloatToString(float64(int(resultXY[1]))) + "," + glb.FloatToString(float64(int(resultXY[0])))
	}

	//if len(coGpUsrHistory) == 0 {
	//	return curUsrPos.Location
	//}
	//atmostTimeDiff := int64(6000)
	//coGpLastPos := coGpUsrHistory[len(coGpUsrHistory)-1]
	//timestampDiff := curUsrPos.Time - coGpLastPos.Time
	////glb.Error.Println(curUsrPos.Time)
	////glb.Error.Println(coGpLastPos.Time)
	//
	//if -1*atmostTimeDiff < timestampDiff && timestampDiff < atmostTimeDiff {
	//	glb.Error.Println(timestampDiff)
	//	loc1 := curUsrPos.Location
	//	loc2 := coGpLastPos.Location
	//	x1, y1 := glb.GetDotFromString(loc1)
	//	x2, y2 := glb.GetDotFromString(loc2)
	//	xt := (x1 + x2) / 2
	//	yt := (y1 + y2) / 2
	//	return glb.FloatToString(xt) + "," + glb.FloatToString(yt)
	//	//return coGpLastPos.Location
	//} else {
	//	return curUsrPos.Location
	//}
}

//////////////////////////////////////
// JUSTS FOR TEST : It's in apimethods
// find best fp location according to
// then return that location and the index of timestamp in allLocationLogsOrdering list that in that time the correct location is found(it's used for optimized search in allLocationLogs)
func FindTrueFPloc(fp parameters.Fingerprint, allLocationLogs map[int64]string, allLocationLogsOrdering []int64) (string, int, error) {
	fpTimeStamp := fp.Timestamp
	//newLoc := ""

	//timeStamps := []int64{}
	//for timestamp, _ := range allLocationLogs {
	//	timeStamps = glb.SortedInsertInt64(timeStamps, timestamp)
	//}
	lessUntil := 0
	//stopit := true
	for i, timeStamp := range allLocationLogsOrdering {

		/*	if timeStamp < int64(1537973812090) && fp.Location=="-22.0,39.0" && stopit{
				glb.Error.Println("Found it ",allLocationLogs[timeStamp])
				stopit = false
			}*/
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
						return "", -1, errors.New("Converting string 2 float problem")
					}
					return glb.IntToString(int(y)) + ".0," + glb.IntToString(int(x)) + ".0", lessUntil, nil
				} else {
					timeStamp1 := allLocationLogsOrdering[i-1]
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
							return "", -1, errors.New("Converting string 2 float problem")
						}
						return glb.IntToString(int(y)) + ".0," + glb.IntToString(int(x)) + ".0", lessUntil, nil
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
							return "", -1, errors.New("Converting string 2 float problem")
						}
						return glb.IntToString(int(y)) + ".0," + glb.IntToString(int(x)) + ".0", lessUntil, nil
					}
				}
				break
			} else {
				//glb.Error.Println("FP timestamp is before the uwb log timestamps")
			}
		}
	}
	glb.Error.Println("FindTrueFPloc on ", fp.Location, ":", fp.Timestamp, " ended but timestamp ranges doesn't match to relocate any fp")
	glb.Error.Println("UWB timestamp range is :", allLocationLogsOrdering[0], " to ", allLocationLogsOrdering[len(allLocationLogsOrdering)-1])
	return "", -1, errors.New("Timestamp range problem")

}

/////////////////////////////////////


// call leanFingerprint(),calculateSVM() and rfLearn() functions after that call prediction functions and return the estimation location
func RecalculateTrackFingerprint(curFingerprint parameters.Fingerprint, timeDiffWithLastFP int64) (parameters.UserPositionJSON, error) {
	userPosJson, success, message := TrackFingerprint(curFingerprint)

	if success {
		gp := dbm.GM.GetGroup(curFingerprint.Group)
		coGp, CoGroupMode, coGpExistErr := gp.Get_CoGroup()

		if gp.Get_ConfigData().Get_OtherGroupConfig().ParticleFilterEnabled && glb.ParticleFilterEnabled {
			//gpUsrHistory := gp.Get_ResultData().Get_UserHistory(glb.TesterUsername)
			//coGpUsrHisotry := coGp.Get_ResultData().Get_UserHistory(glb.TesterUsername)
			//if len(gpUsrHistory) == 0 {
			resultXY := ParticleFilterFusion(userPosJson, timeDiffWithLastFP, CoGroupMode)
			glb.Error.Println("Particle filter result: ", resultXY)
			userPosJson.Location = resultXY
			//if timeDiffWithLastFP == 0 {
			//	userPosJson.Location = ParticleFilterFusion(userPosJson, gpUsrHistory,coGpUsrHisotry, true)
			//} else {
			//	userPosJson.Location = ParticleFilterFusion(userPosJson, gpUsrHistory,coGpUsrHisotry, false)
			//}
		} else {
			if coGpExistErr == nil {
				gpUsrHistory := gp.Get_ResultData().Get_UserHistory(glb.TesterUsername)
				coGpUsrHisotry := coGp.Get_ResultData().Get_UserHistory(glb.TesterUsername)
				userPosJson.Location = Fusion(userPosJson, gpUsrHistory, coGpUsrHisotry)
			}
		}


		//userPosJson.KnnGuess = userPosJson.Location //todo: must add location as seprated variable( from knnguess) in parameters.UserPositionJSON
		gp.Get_ResultData().Append_UserHistory(curFingerprint.Username, userPosJson)
		return userPosJson, nil
	} else {
		return userPosJson, errors.New(message)
	}
}

func RecalculateTrackFingerprintKnnCrossValidation(curFingerprint parameters.Fingerprint) string {
	//userPosJson, success, message := TrackFingerprint(curFingerprint, KNN, IgnoreSimpleHistoryEffect)
	userPosJson, success, _ := TrackFingerprint(curFingerprint, glb.KNN)

	if success {
		gp := dbm.GM.GetGroup(curFingerprint.Group)
		//userPosJson.KnnGuess = userPosJson.Location //todo: must add location as seprated variable( from knnguess) in parameters.UserPositionJSON
		gp.Get_ResultData().Append_UserHistory(curFingerprint.Username, userPosJson)
	}

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
	knnKRange := dbm.GetSharedPrf(gp.Get_Name()).KRange
	if len(knnKRange) == 1{
		validKs = glb.MakeRange(knnKRange[0],knnKRange[0])
	}else if len(knnKRange) == 2{
		validKs = glb.MakeRange(knnKRange[0],knnKRange[1])
	}else{
		glb.Error.Println("Can't set valid Knn K values")
	}
		//2.MinClusterRSS
	validMinClusterRSSs := glb.MakeRange(glb.DefaultKnnMinClusterRssRange[0],glb.DefaultKnnMinClusterRssRange[1])

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
		fp = dbm.FilterFingerprint(fp)
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
	knnConfig := cd.Get_KnnConfig()

	//totalErrorList := []int{}
	//knnErrHyperParameters := make(map[int]parameters.KnnHyperParameters)

	allHyperParamDetails := make(map[int]parameters.KnnHyperParameters)
	allErrDetails := make(map[int][]int)

	//bestResult := -1

	//Set algorithm parameters range:

	// KNN:
	// Parameters list creation
	// 1.K
	validKs := []int{glb.DefaultKnnKRange[0]}
	if len(glb.DefaultKnnKRange) > 1 {
		validKs = glb.MakeRange(glb.DefaultKnnKRange[0], glb.DefaultKnnKRange[1])
	}
	knnKRange := knnConfig.KRange
	if len(knnKRange) == 1 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[0])
	} else if len(knnKRange) == 2 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[1])
	} else {
		glb.Error.Println("knnKRange:", knnKRange)
		glb.Error.Println("Can't set valid Knn K values")
	}
	//2.MinClusterRSS
	validMinClusterRSSs := []int{glb.DefaultKnnMinClusterRssRange[0]}
	if len(glb.DefaultKnnMinClusterRssRange) > 1 {
		validMinClusterRSSs = glb.MakeRange(glb.DefaultKnnMinClusterRssRange[0], glb.DefaultKnnMinClusterRssRange[1])
	}

	minClusterRSSRange := knnConfig.MinClusterRssRange
	if len(minClusterRSSRange) == 1 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[0])
	} else if len(minClusterRSSRange) == 2 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[1])
		validMinClusterRSSs = append(validMinClusterRSSs, 0)
	} else {
		glb.Error.Println("minClusterRSSRange:", minClusterRSSRange)
		glb.Error.Println("Can't set valid min cluster rss values")
	}

	//3.MaxEuclideanRssDist
	validMaxEuclideanRssDists := []int{glb.DefaultMaxEuclideanRssDistRange[0]}
	if len(glb.DefaultMaxEuclideanRssDistRange) > 1 {
		validMaxEuclideanRssDists = glb.MakeRange(glb.DefaultMaxEuclideanRssDistRange[0], glb.DefaultMaxEuclideanRssDistRange[1])
	}
	maxEuclideanRssDistRange := knnConfig.MaxEuclideanRssDistRange
	if len(maxEuclideanRssDistRange) == 1 {
		validMaxEuclideanRssDists = glb.MakeRange(maxEuclideanRssDistRange[0], maxEuclideanRssDistRange[0])
	} else if len(maxEuclideanRssDistRange) == 2 {
		validMaxEuclideanRssDists = glb.MakeRange(maxEuclideanRssDistRange[0], maxEuclideanRssDistRange[1])
	} else {
		glb.Error.Println("maxEuclideanRssDistRange:", maxEuclideanRssDistRange)
		glb.Error.Println("Can't set valid maxEuclidean Rss Dist Range values")
	}

	//4.BLEFactor
	validBLEFactors := []float64{glb.DefaultBLEFactorRange[0]}
	if len(glb.DefaultBLEFactorRange) > 1 {
		validBLEFactors = glb.MakeRangeFloat(glb.DefaultBLEFactorRange[0], glb.DefaultBLEFactorRange[1])
	}
	bleFactorRange := knnConfig.BLEFactorRange
	if len(bleFactorRange) == 1 {
		validBLEFactors = glb.MakeRangeFloat(bleFactorRange[0], bleFactorRange[0])
	} else if len(bleFactorRange) == 2 {
		validBLEFactors = glb.MakeRangeFloat(bleFactorRange[0], bleFactorRange[1], float64(1))
	} else if len(bleFactorRange) == 3 {
		validBLEFactors = glb.MakeRangeFloat(bleFactorRange[0], bleFactorRange[1], bleFactorRange[2])
	} else {
		glb.Error.Println("bleFactorRange:", bleFactorRange)
		glb.Error.Println("Can't set valid bleFactor Range values")
	}

	// Set length of calculation progress bar
	// This is shared between all threads, so it's invalid when two calculateLearn thread run
	calculationLen := len(validMinClusterRSSs) * len(validKs) * len(validMaxEuclideanRssDists) * len(validMaxEuclideanRssDists) * len(validBLEFactors)
	glb.ProgressBarLength = calculationLen

	adTemp := tempGp.NewAlgoDataStruct()
	rdTemp := tempGp.Get_RawData()

	//allErrDetailsList = make([][]int,calculationLen)

	paramUniqueKey := 0 // just creating unique key for each possible the parameters permutation
	for i1, maxEuclideanRssDist := range validMaxEuclideanRssDists {
		for i2, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
			for i3, K := range validKs { // for over KnnK
				for i4, bleFactor := range validBLEFactors {

					glb.ProgressBarCurLevel = i1*len(validMinClusterRSSs) + i2*len(validKs) + i3*len(validBLEFactors) + i4
					totalDistError := 0

					tempHyperParameters := parameters.NewKnnHyperParameters()
					//glb.Error.Println(tempHyperParameters)
					tempHyperParameters.K = K
					tempHyperParameters.MinClusterRss = minClusterRss
					tempHyperParameters.MaxEuclideanRssDist = maxEuclideanRssDist
					tempHyperParameters.BLEFactor = bleFactor

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

						PreProcess(rdTemp, shprf.NeedToRelocateFP)
						GetParametersWithGP(tempGp)

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
					glb.Debug.Printf("Knn error (minClusterRss=%d,K=%d,maxEuclideanRssDist=%d) = %d \n", minClusterRss, K, maxEuclideanRssDist, totalDistError)

					//if(bestResult==-1 || totalDistError<bestResult){
					//	bestResult = totalDistError
					//	bestK = K
					//}
					//totalErrorList = append(totalErrorList, totalDistError)
					//knnErrHyperParameters[totalDistError] = tempHyperParameters
				}
			}
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
	knnConfig := cd.Get_KnnConfig()
	tempGp.Set_ConfigData(cd)

	totalErrorList := []int{}
	knnErrHyperParameters := make(map[int]parameters.KnnHyperParameters)

	bestResult := -1

	//Set algorithm parameters range:

	// KNN:
	// Parameters list creation
	// 1.K
	validKs := []int{glb.DefaultKnnKRange[0]}
	if len(glb.DefaultKnnKRange) > 1 {
		validKs = glb.MakeRange(glb.DefaultKnnKRange[0], glb.DefaultKnnKRange[1])
	}
	//knnKRange := shprf.KRange
	knnKRange := knnConfig.KRange
	if len(knnKRange) == 1 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[0])
	} else if len(knnKRange) == 2 {
		validKs = glb.MakeRange(knnKRange[0], knnKRange[1])
	} else {
		glb.Error.Println("knnKRange:", knnKRange)
		glb.Error.Println("Can't set valid Knn K values")
	}
	//2.MinClusterRSS
	validMinClusterRSSs := []int{glb.DefaultKnnMinClusterRssRange[0]}
	if len(glb.DefaultKnnMinClusterRssRange) > 1 {
		validMinClusterRSSs = glb.MakeRange(glb.DefaultKnnMinClusterRssRange[0], glb.DefaultKnnMinClusterRssRange[1])
	}
	minClusterRSSRange := knnConfig.MinClusterRssRange
	if len(minClusterRSSRange) == 1 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[0])
	} else if len(minClusterRSSRange) == 2 {
		validMinClusterRSSs = glb.MakeRange(minClusterRSSRange[0], minClusterRSSRange[1])
		validMinClusterRSSs = append(validMinClusterRSSs, 0)
	} else {
		glb.Error.Println("minClusterRSSRange:", minClusterRSSRange)
		glb.Error.Println("Can't set valid min cluster rss values")
	}

	//3.MaxEuclideanRssDist
	validMaxEuclideanRssDists := []int{glb.DefaultMaxEuclideanRssDistRange[0]}
	if len(glb.DefaultMaxEuclideanRssDistRange) > 1 {
		validMaxEuclideanRssDists = glb.MakeRange(glb.DefaultMaxEuclideanRssDistRange[0], glb.DefaultMaxEuclideanRssDistRange[1])
	}
	maxEuclideanRssDistRange := knnConfig.MaxEuclideanRssDistRange
	if len(maxEuclideanRssDistRange) == 1 {
		validMaxEuclideanRssDists = glb.MakeRange(maxEuclideanRssDistRange[0], maxEuclideanRssDistRange[0])
	} else if len(maxEuclideanRssDistRange) == 2 {
		validMaxEuclideanRssDists = glb.MakeRange(maxEuclideanRssDistRange[0], maxEuclideanRssDistRange[1])
	} else {
		glb.Error.Println("maxEuclideanRssDistRange:", maxEuclideanRssDistRange)
		glb.Error.Println("Can't set valid maxEuclidean Rss Dist Range values")
	}

	//4.BLEFactor
	validBLEFactors := []float64{glb.DefaultBLEFactorRange[0]}
	if len(glb.DefaultBLEFactorRange) > 1 {
		validBLEFactors = glb.MakeRangeFloat(glb.DefaultBLEFactorRange[0], glb.DefaultBLEFactorRange[1])
	}
	bleFactorRange := knnConfig.BLEFactorRange
	if len(bleFactorRange) == 1 {
		validBLEFactors = glb.MakeRangeFloat(bleFactorRange[0], bleFactorRange[0])
	} else if len(bleFactorRange) == 2 {
		validBLEFactors = glb.MakeRangeFloat(bleFactorRange[0], bleFactorRange[1], float64(1))
	} else if len(bleFactorRange) == 3 {
		validBLEFactors = glb.MakeRangeFloat(bleFactorRange[0], bleFactorRange[1], bleFactorRange[2])
	} else {
		glb.Error.Println("bleFactorRange:", bleFactorRange)
		glb.Error.Println("Can't set valid bleFactor Range values")
	}

	// Set length of calculation progress bar
	// This is shared between all threads, so it's invalid when two calculateLearn thread run
	glb.ProgressBarLength = len(validMinClusterRSSs) * len(validKs) * len(validMaxEuclideanRssDists) * len(validMaxEuclideanRssDists) * len(validBLEFactors)

	adTemp := tempGp.NewAlgoDataStruct()
	rdTemp := tempGp.Get_RawData()
	for i1, maxEuclideanRssDist := range validMaxEuclideanRssDists {
		for i2, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
			for i3, K := range validKs { // for over KnnK
				for i4, bleFactor := range validBLEFactors {
					glb.ProgressBarCurLevel = i1*len(validMinClusterRSSs) + i2*len(validKs) + i3*len(validBLEFactors) + i4

					totalDistError := 0

					tempHyperParameters := parameters.NewKnnHyperParameters()
					//glb.Error.Println(tempHyperParameters)
					tempHyperParameters.K = K
					tempHyperParameters.MinClusterRss = minClusterRss
					tempHyperParameters.MaxEuclideanRssDist = maxEuclideanRssDist
					tempHyperParameters.BLEFactor = bleFactor

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

						PreProcess(rdTemp, shprf.NeedToRelocateFP)
						GetParametersWithGP(tempGp)

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

					glb.Debug.Printf("Knn error (minClusterRss=%d,K=%d,maxEuclideanRssDist=%d) = %d \n", minClusterRss, K, maxEuclideanRssDist, totalDistError)

					//if(bestResult==-1 || totalDistError<bestResult){
					//	bestResult = totalDistError
					//	bestK = K
					//}
					totalErrorList = append(totalErrorList, totalDistError)
					knnErrHyperParameters[totalDistError] = tempHyperParameters
				}
			}
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
	glb.Debug.Println("Best MaxEuclideanRssDist : ", bestErrHyperParameters.MaxEuclideanRssDist)
	glb.Debug.Println("Best BLEFactor : ", bestErrHyperParameters.BLEFactor)
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

func DisableHistoryConsideredMethodTemprorary(cd *dbm.ConfigDataStruct) parameters.KnnConfig {
	lastKnnConfig := cd.Get_KnnConfig()
	knnConfig := parameters.KnnConfig{}
	copier.Copy(&knnConfig, &lastKnnConfig)
	if knnConfig.GraphEnabled { // Todo: graphEnabled must be declared in sharedPrf
		knnConfig.GraphEnabled = false
	}
	if knnConfig.DSAEnabled {
		knnConfig.DSAEnabled = false
	}
	cd.Set_KnnConfig(knnConfig)
	glb.Debug.Println("Disabling history considered methods, learning with knnconfig:", knnConfig)
	return lastKnnConfig
}

func EnableHistoryConsideredMethodTemprorary(cd *dbm.ConfigDataStruct, lastKnnConfig parameters.KnnConfig) {
	glb.Debug.Println("Enabling history considered methods, learning with knnconfig:", lastKnnConfig)
	cd.Set_KnnConfig(lastKnnConfig)
}

func CalculateLearn(groupName string) {
	// Now performance isn't important in learning, just care about performance on track (it helps to code easily!)

	glb.Debug.Println("################### CalculateLearn ##################")
	gp := dbm.GM.GetGroup(groupName)
	gp.Set_Permanent(false) //for crossvalidation

	rd := gp.Get_RawData()
	mainFPData := rd.Get_FingerprintsBackup()
	mainFPOrdering := rd.Get_FingerprintsOrderingBackup()
	cd := gp.Get_ConfigData()
	ad := gp.Get_AlgoData()
	rs := gp.Get_ResultData()

	//knnConfig := cd.Get_KnnConfig()

	lastKnnConfigs := DisableHistoryConsideredMethodTemprorary(cd)    // disable graph
	defer EnableHistoryConsideredMethodTemprorary(cd, lastKnnConfigs) // enabled graph at end

	knnLocAccuracy := make(map[string]int)
	var crossValidationPartsList []crossValidationParts
	//glb.Debug.Println(mainFPOrdering)
	//glb.Debug.Println(mainFPData)
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

		PreProcess(rd, shprf.NeedToRelocateFP)
		GetParametersWithGP(gp)

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

	rs.Set_ALL_AlgoLocAccuracy(make(map[string]map[string]int))
	for loc, accuracy := range knnLocAccuracy {
		rs.Set_AlgoLocAccuracy("knn", loc, accuracy)
	}
	glb.Debug.Println(dbm.GetCVResults(groupName))

	// Set main parameters: Note: Don't change mainFpData
	rd.Set_Fingerprints(mainFPData)
	rd.Set_FingerprintsOrdering(mainFPOrdering)
	rd.Set_FingerprintsBackup(mainFPData)
	rd.Set_FingerprintsOrderingBackup(mainFPOrdering)

	gp.GMutex.Lock()

	PreProcess(rd, shprf.NeedToRelocateFP)

	GetParametersWithGP(gp)
	glb.Debug.Println(gp.Get_MiddleData().Get_UniqueMacs())

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

	//runnerLock.Unlock()
}

func CalculateByTestValidTracks(groupName string) {
	gp := dbm.GM.GetGroup(groupName)
	rsd := gp.Get_ResultData()
	cd := gp.Get_ConfigData()
	ad := gp.Get_AlgoData()
	knnConfig := gp.Get_ConfigData().Get_KnnConfig()
	knnFPs := ad.Get_KnnFPs()

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

	glb.Debug.Println("bestKnnHyperParams before CalculateByTestValidTracks :", knnHyperParams)
	// 1. TestValid without graph:
	/*	testValidTrueLoc := []string{}
		testvalidGuessLoc := []string{}*/
	//glb.DefaultGraphEnabled = false
	lastKnnConfigs := DisableHistoryConsideredMethodTemprorary(cd)
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
	EnableHistoryConsideredMethodTemprorary(cd, lastKnnConfigs)
	/*	glb.Error.Println("testValidTrueLoc",testValidTrueLoc)
		glb.Error.Println("testvalidGuessLoc",testvalidGuessLoc)*/

	// 2. Testvalid with graph:
	//totalErrorList := []int{}
	//knnErrHyperParameters := make(map[int]parameters.KnnHyperParameters)
	//bestResult := -1

	// graphfactor used in online phase so learning must be done one time
	var bestErrHyperParameters parameters.KnnHyperParameters

	rsd.Set_AlgoAccuracy("knn_testvalid_graph", 0)
	rsd.Set_AlgoAccuracy("knn_testvalid_dsa", 0)

	if knnConfig.GraphEnabled {
		glb.Debug.Println("Selecting best graph factors by test-valid tracks ...")
		bestErrHyperParameters = SelectBestGraphFactorsByTestValidTracks(gp, testValidTracks, knnConfig, knnHyperParams)
	} else if knnConfig.DSAEnabled {
		glb.Debug.Println("Selecting best max movement factor by test-valid tracks ...")
		bestErrHyperParameters = SelectBestMaxMovementByTestValidTracks(gp, testValidTracks, knnConfig, knnHyperParams)
	} else {
		bestErrHyperParameters = knnHyperParams
	}

	knnFPs.HyperParameters = bestErrHyperParameters
	ad.Set_KnnFPs(knnFPs)
}

func SelectBestGraphFactorsByTestValidTracks(gp *dbm.Group, testValidTracks []parameters.TestValidTrack, knnConfig parameters.KnnConfig, knnHyperParams parameters.KnnHyperParameters) parameters.KnnHyperParameters {
	learnedKnnData, _ := LearnKnn(gp, knnHyperParams)
	ad := gp.Get_AlgoData()
	rsd := gp.Get_ResultData()
	// GraphFactors range:
	//validGraphFactorsRange = [][]float64{}
	//validGraphFactors := [][]float64{{0.5, 0.5, 1, 1, 1}, {2, 1, 1, 1}, {100, 100, 100, 100, 3, 2, 1}, {10, 10, 10, 10, 3, 2, 1}, {8, 8, 8, 8, 3, 2, 1}, {1, 1, 1, 1}}

	//beginSlice := []float64{1, 1, 1, 1, 1, 1, 1}
	//endSlice := []float64{10, 10, 10, 10, 3, 2, 1}
	graphFactorRange := knnConfig.GraphFactorRange
	var validGraphFactors [][]float64
	if len(graphFactorRange) == 0 {
		glb.Debug.Println("graphFactorRange is empty, CalculateByTestValidTracks is ignored!")
		return knnHyperParams
	} else if len(graphFactorRange) == 1 {
		validGraphFactors = graphFactorRange[:1]
	} else if len(graphFactorRange) == 2 {
		validGraphFactors = glb.GetGraphSlicesRangeRecursive(graphFactorRange[0], graphFactorRange[1], glb.DefaultGraphStep)
		validGraphFactors = append(validGraphFactors, []float64{1, 1, 1, 1})
	} else if len(graphFactorRange) == 3 {
		validGraphFactors = glb.GetGraphSlicesRangeRecursive(graphFactorRange[0], graphFactorRange[1], graphFactorRange[2][0])
		validGraphFactors = append(validGraphFactors, []float64{1, 1, 1, 1})
	} else {
		glb.Error.Println("graphFactorRange length must be lower than 3(now range created by first and second members)")
		validGraphFactors = glb.GetGraphSlicesRangeRecursive(graphFactorRange[0], graphFactorRange[1], graphFactorRange[2][0])
		validGraphFactors = append(validGraphFactors, []float64{1, 1, 1, 1})
	}

	////Note: Delete it!
	//validGraphFactors = [][]float64{{10,0,10,0,10,10,10},{10,0,0,10,10,0,0},{10,0,0,0,0,10,10},{10,10,10,0,0,10,10},{10,0,10,10,10,10,0},{10,10,0,0,0,10,0},{10,0,0,10,10,10,10},{10,10,0,10,10,0,10},{10,10,10,0,0,0,0},{10,10,10,10,10,0,0},{10,0,0,10,0,0,10},{10,10,0,10,10,10,0},{10,0,10,10,0,10,10},{10,0,0,10,0,10,0},{10,10,0,0,10,0,0},{10,0,0,0,10,10,0},{10,10,10,0,10,10,0},{10,0,0,0,0,0,0},{10,0,10,0,10,0,0},{10,10,0,10,0,10,10},{10,10,0,10,0,0,0},{10,10,10,10,10,10,10},{10,0,10,0,0,0,10},{10,10,10,0,10,0,10},{10,10,0,0,10,10,10},{10,0,10,10,10,0,10},{10,0,0,0,10,0,10},{10,0,10,0,0,10,0},{10,10,0,0,0,0,10},{10,10,10,10,0,10,0},{10,10,10,10,0,0,10}}

	// Iterate over all of test-valid tracks
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
	glb.Debug.Println("%%%%%%% best result %%%%%%", bestResult)

	for _, i := range sortedErrDetails {
		glb.Debug.Println("-----------------------------")
		glb.Debug.Println("Hyper Params:", allHyperParamDetails[i])
		glb.Debug.Println("Error:", newErrorMap[i])
	}

	for _, i := range sortedErrDetails {
		glb.Debug.Println(allHyperParamDetails[i], " ", newErrorMap[i])
	}
	glb.Debug.Println("Best HyperParameters: ", bestErrHyperParameters)
	glb.Debug.Println("Best GraphFactors : ", bestErrHyperParameters.GraphFactors)
	rsd.Set_AlgoAccuracy("knn_testvalid_graph", bestResult)

	return bestErrHyperParameters
}

func SelectBestMaxMovementByTestValidTracks(gp *dbm.Group, testValidTracks []parameters.TestValidTrack, knnConfig parameters.KnnConfig, knnHyperParams parameters.KnnHyperParameters) parameters.KnnHyperParameters {
	learnedKnnData, _ := LearnKnn(gp, knnHyperParams)
	ad := gp.Get_AlgoData()
	rsd := gp.Get_ResultData()

	// Maxmovement range
	validMaxMovements := []int{glb.DefaultMaxMovementRange[0]}
	if len(glb.DefaultMaxMovementRange) > 1 {
		validMaxMovements = glb.MakeRange(glb.DefaultMaxMovementRange[0], glb.DefaultMaxMovementRange[1])
	}
	maxMovementRange := knnConfig.MaxMovementRange
	if len(maxMovementRange) == 1 {
		validMaxMovements = glb.MakeRange(maxMovementRange[0], maxMovementRange[0])
	} else if len(maxMovementRange) == 2 {
		validMaxMovements = glb.MakeRange(maxMovementRange[0], maxMovementRange[1])
	} else if len(maxMovementRange) == 3 {
		validMaxMovements = glb.MakeRange(maxMovementRange[0], maxMovementRange[1], maxMovementRange[2])
	} else {
		glb.Error.Println("maxMovementRange:", maxMovementRange)
		glb.Error.Println("Can't set valid maxMovement Range values")
	}

	glb.Debug.Println(validMaxMovements)

	// Iterate over all of test-valid tracks
	allHyperParamDetails := make(map[int]parameters.KnnHyperParameters)
	allErrDetails := make(map[int][]int)
	paramUniqueKey := 0 // just creating unique key for each possible the parameters permutation
	for _, maxMovement := range validMaxMovements { // for over the validMaxMovements
		gp.Get_ResultData().Set_UserHistory(glb.TesterUsername, []parameters.UserPositionJSON{}) // clear userhistory to check knn error with new graphfactor

		allHyperParamDetails[0] = learnedKnnData.HyperParameters

		tempHyperParameters := knnHyperParams
		tempHyperParameters.MaxMovement = maxMovement
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
		glb.Debug.Println("Calculation for maxMovement=", maxMovement, " ended ", )
	}

	/*	sort.Ints(totalErrorList)
		bestResult = totalErrorList[0]
		bestErrHyperParameters := knnErrHyperParameters[bestResult]*/

	bestKey, sortedErrDetails, newErrorMap := SelectBestFromErrMap(allErrDetails)
	bestErrHyperParameters := allHyperParamDetails[bestKey]
	bestResult := newErrorMap[bestKey]
	glb.Debug.Println("%%%%%%% best result %%%%%%", bestResult)

	for _, i := range sortedErrDetails {
		glb.Debug.Println("-----------------------------")
		glb.Debug.Println("Hyper Params:", allHyperParamDetails[i])
		glb.Debug.Println("Error:", newErrorMap[i])
	}

	for _, i := range sortedErrDetails {
		glb.Debug.Println(allHyperParamDetails[i], " ", newErrorMap[i])
	}
	glb.Debug.Println("Best HyperParameters: ", bestErrHyperParameters)
	glb.Debug.Println("Best MaxMovement : ", bestErrHyperParameters.MaxMovement)
	rsd.Set_AlgoAccuracy("knn_testvalid_dsa", bestResult)

	return bestErrHyperParameters
}

// Error calculation using mean as error algorithm
func CalculateByTestValidTracksLegacy(groupName string) {
	gp := dbm.GM.GetGroup(groupName)
	rsd := gp.Get_ResultData()
	cd := gp.Get_ConfigData()
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

	glb.Debug.Println("bestKnnHyperParams before CalculateByTestValidTracks :", knnHyperParams)
	// 1. TestValid without graph:
	/*	testValidTrueLoc := []string{}
		testvalidGuessLoc := []string{}*/
	//glb.DefaultGraphEnabled = false
	lastKnnConfigs := DisableHistoryConsideredMethodTemprorary(cd)
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
	//glb.DefaultGraphEnabled = true
	EnableHistoryConsideredMethodTemprorary(cd, lastKnnConfigs)
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
