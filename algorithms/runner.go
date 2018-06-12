package algorithms

import (
	"ParsinServer/glb"
	"strings"

	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/dbm"
	"ParsinServer/algorithms/clustering"
	"sort"
	"runtime"
)

var threadedCross bool = false //don't use, it's not safe now!

type KnnJob struct{
	gp							*dbm.Group
	K							int
	MinClusterRss 				int
	crossValidationPartsList 	[]crossValidationParts
}

type KnnJobResult struct{
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
				for algorithm, valXY := range scikitData{
					resJsonMap[algorithm]=valXY
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
	gp.Get_ResultData().Append(curFingerprint)
	//go dbm.PutFingerprintIntoDatabase(fullFingerprint, "fingerprints-track")


	// Calculating Scikit
	//
	//if glb.RuntimeArgs.Scikit {
	//	scikitData = ScikitClassify(strings.ToLower(curFingerprint.Group), curFingerprint)
	//	glb.Debug.Println(scikitData)
	//	for algorithm, valueXY := range scikitData{
	//		message += " "+algorithm+":v" + valueXY
	//	}
	//
	//}

	// Send out the final responses
	var userJSON glb.UserPositionJSON
	userJSON.Time = curFingerprint.Timestamp
	userJSON.BayesGuess = bayesGuess
	userJSON.BayesData = bayesData
	userJSON.SvmGuess = svmGuess
	userJSON.SvmData = svmData
	userJSON.ScikitData = scikitData
	userJSON.KnnGuess = knnGuess
	userJSON.KnnData = knnData
	userJSON.PDRLocation = pdrLocation

	go dbm.SetUserPositionCache(strings.ToLower(curFingerprint.Group)+strings.ToLower(curFingerprint.Username), userJSON)
	go gp.Get_ResultData().Append_UserHistory(curFingerprint.Username, userJSON)

	// User history effect
	location = knnGuess
	userHistory := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)
	location, accuracyCircleRadius = HistoryEffect(userJSON, userHistory)

	userJSON.Location = location
	glb.Debug.Println("Knn guess: ", knnGuess)
	glb.Debug.Println("location: ", location)

	//location = userJSON.KnnGuess
	//userJSON.KnnGuess = location //todo: must add location as seprated variable from knnguess in glb.UserPositionJSON

	//glb.Debug.Println(len(gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)))

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
		//go routes.SendMQTTLocation(string(mqttMessage), curFingerprint.Group, curFingerprint.Username)
	}

	glb.Debug.Println(userJSON)

	return message, true, location, bayesGuess, bayesData, svmGuess, svmData, knnGuess, scikitData, accuracyCircleRadius

}

func CalculateLearn(groupName string) {
	// Now performance isn't important in learning, just care about performance on track (it helps to code easily!)
	groupName = strings.ToLower(groupName)
	gp := dbm.GM.GetGroup(groupName)

	gp.Set_Permanent(false) //for crossvalidation
	rd := gp.Get_RawData_Filtered_Val()

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
					learnedKnnData.K = K
					learnedKnnData.MinClusterRss = minClusterRss

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
	glb.Debug.Println()
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
		learnedKnnData.K = bestK
		learnedKnnData.MinClusterRss = bestMinClusterRss

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
	// select best algo config

	// learn algorithm
	ad := gp.Get_AlgoData()
	gp.Set_Permanent(true)
	bestHyperParameters := []interface{}{bestK,bestMinClusterRss}
	learnedKnnData,_:= LearnKnn(md,rd,bestHyperParameters)

	// Set best hyper parameter values
	learnedKnnData.K = bestK
	learnedKnnData.MinClusterRss = bestMinClusterRss

	ad.Set_KnnFPs(learnedKnnData)
	gp.Set_AlgoData(ad)

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
	//if glb.RuntimeArgs.Scikit {
	//	ScikitLearn(groupName)
	//}
	//runnerLock.Unlock()
}


func calcKNN(id int, knnJobs <-chan KnnJob, knnJobResults chan<- KnnJobResult) {
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
			learnedKnnData.K = knnJob.K
			learnedKnnData.MinClusterRss = knnJob.MinClusterRss

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


type crossValidationParts struct{
	trainSet			[]dbm.RawDataStruct
	testSet				dbm.RawDataStruct
}
func (cvParts *crossValidationParts) GetTrainSet(gp *dbm.Group) dbm.RawDataStruct {
	resRD := *gp.NewRawDataStruct()
	for _,rd := range cvParts.trainSet{
		for _,index := range rd.FingerprintsOrdering{
			resRD.FingerprintsOrdering = append(resRD.FingerprintsOrdering,index)
			resRD.Fingerprints[index] = rd.Fingerprints[index]
		}
	}
	return resRD
}

func GetCrossValidationParts(gp *dbm.Group,rd dbm.RawDataStruct) []crossValidationParts{
	var CVPartsList []crossValidationParts
	var tempCVParts crossValidationParts
	locRDMap := make(map[string]dbm.RawDataStruct)

	for _,index := range rd.FingerprintsOrdering{
		fp := rd.Fingerprints[index]

		if fpRD, ok := locRDMap[fp.Location]; ok {
			fpRD.Fingerprints[index] = fp
			fpRD.FingerprintsOrdering = append(fpRD.FingerprintsOrdering, index)
			locRDMap[fp.Location] = fpRD
		}else{
			fpM := make(map[string]parameters.Fingerprint)
			fpM[index] = fp
			fpO := []string{index}
			templocRDMap:= *gp.NewRawDataStruct()
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

	for locMain,_ := range locRDMap {
		tempCVParts.testSet = dbm.RawDataStruct{}
		tempCVParts.trainSet = []dbm.RawDataStruct{}
		for loc,RD := range locRDMap {
				if(loc==locMain){
					// add to test set
					tempCVParts.testSet = RD
				}else{
					// add to train set
					tempCVParts.trainSet = append(tempCVParts.trainSet,RD)
				}
		}
		CVPartsList = append(CVPartsList, tempCVParts)
	}
	return CVPartsList
}


//group: group
//ps:
//fingerprintsInMemory:
//fingerprintsOrdering:
//updates ps with the new fingerprint.
//(The Parameters which are manipulated: NetworkMacs,NetworkLocs,UniqueMacs,UniqueLocs,MacCount,MacCountByLoc and Loaded)
func GetParameters(md *dbm.MiddleDataStruct,rd dbm.RawDataStruct) {
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
	for _, fpIndex:= range fingerprintsOrdering {
		fp := fingerprints[fpIndex]
		locations = append(locations,fp.Location)
	}
	md.LocCount = glb.DuplicateCountString(locations)

}
