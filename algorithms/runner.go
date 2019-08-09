package algorithms

import (
	"ParsinServer/algorithms/particlefilter"
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"math"
	"net/http"
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

	if len(curFingerprint.WifiFingerprint) < glb.MinApNum {
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

	//glb.Debug.Println(userPosJson)
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
			userPosJsonTest := parameters.UserPositionJSON{}
			copier.Copy(&userPosJsonTest, &userPosJson)
			userPosJsonTest.Fingerprint = curFingerprint // avoid filtering and edition of the main fingerprint
			tempTestValidTrack := parameters.TestValidTrack{TrueLocation: userPosJson.Fingerprint.Location, UserPosition: userPosJsonTest}
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
