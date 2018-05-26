package algorithms

import (
	"ParsinServer/glb"
	"strings"

	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"math"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/dbm"
	"ParsinServer/algorithms/clustering"
	"sort"
	"runtime"
	"time"
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
	//fullFingerprint := jsonFingerprint

	dbm.FilterFingerprint(&jsonFingerprint)

	groupName := strings.ToLower(jsonFingerprint.Group)

	bayesGuess := ""
	bayesData := make(map[string]float64)
	svmGuess := ""
	svmData := make(map[string]float64)
	scikitData := make(map[string]string)
	knnGuess := ""
	message := ""

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
			go dbm.AppendUserCache(groupName, jsonFingerprint.Username)
		}
	}
	glb.Info.Println(jsonFingerprint)
	//bayesGuess, bayesData = bayes.CalculatePosterior(jsonFingerprint, nil)
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
	//jsonFingerprint.Location = bayesGuess

	//message := ""
	//glb.Debug.Println("Tracking fingerprint containing " + strconv.Itoa(len(jsonFingerprint.WifiFingerprint)) + " APs for " + jsonFingerprint.Username + " (" + jsonFingerprint.Group + ") at " + jsonFingerprint.Location + " (guess)")
	//message += " BayesGuess: " + bayesGuess //+ " (" + strconv.Itoa(int(percentGuess1)) + "% confidence)"
	//
	//// Process SVM if needed
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

	gp := dbm.GM.GetGroup(groupName)
	// Calculating KNN
	glb.Debug.Println(jsonFingerprint)
	err, knnGuess := TrackKnn(gp,jsonFingerprint)
	if err != nil {
		glb.Error.Println(err)
	}
	message += " knnGuess: " + knnGuess

	jsonFingerprint.Location = knnGuess
	// Insert full fingerprint
	//glb.Debug.Println(jsonFingerprint)
	gp.Get_ResultData().Append(jsonFingerprint)
	//go dbm.PutFingerprintIntoDatabase(fullFingerprint, "fingerprints-track")


	// Calculating Scikit
	//
	//if glb.RuntimeArgs.Scikit {
	//	scikitData = ScikitClassify(strings.ToLower(jsonFingerprint.Group), jsonFingerprint)
	//	glb.Debug.Println(scikitData)
	//	for algorithm, valueXY := range scikitData{
	//		message += " "+algorithm+":v" + valueXY
	//	}
	//
	//}

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

func CalculateLearn(groupName string) {
	// Now performance isn't important in learning, just care about performance on track (it helps to code easily!)
	//var runnerLock sync.Mutex
	//runnerLock.Lock()
	groupName = strings.ToLower(groupName)
	//var gpMain = dbm.NewGroup(groupName)
	//defer dbm.GM.GetGroup(groupName).Set(gpMain)
	gp := dbm.GM.GetGroup(groupName)

	gp.Set_Permanent(false) //for crossvalidation
	rd := gp.Get_RawData_Filtered_Val()

	//glb.Debug.Println(1)
	var crossValidationPartsList []crossValidationParts
	//glb.Debug.Println(rd)
	crossValidationPartsList = GetCrossValidationParts(gp,rd)
	//glb.Debug.Println(2)
	// ToDo: Need to learn algorithms concurrently



	// CrossValidation

	//crossHistory := map[string]float64{"-1064.000000,-240.000000":9033.65681, "-18.000000,1408.000000":1207.539336, "-676.000000,-216.000000":1963.042125, "-140.000000,1660.000000":2708.496748, "11.000000,-432.000000":6478.9375660000005, "-11.000000,-164.000000":5596.867363, "-128.000000,1676.000000":2903.8254620000002, "-346.000000,-140.000000":3914.7535380000004, "-983.000000,-1543.000000":2686.4171349999997, "-72.000000,-1655.000000":2635.7918250000002, "426.000000,1697.000000":3777.554719, "-334.000000,-124.000000":4608.481008999999, "275.000000,-1564.000000":4734.017012, "41.000000,895.000000":5279.052433999999, "-7.000000,294.000000":2250.712678, "-10.000000,1416.000000":1276.908034, "-965.000000,1427.000000":4044.6845869999997, "71.000000,-844.000000":3105.23509, "265.000000,-1560.000000":2028.976247, "-997.000000,1331.000000":6004.292162000001, "-15.000000,809.000000":4580.49267, "-992.000000,-756.000000":4737.604829999999, "-636.000000,-1232.000000":2362.0894990000006, "67.000000,-428.000000":3094.2482190000005, "17.000000,899.000000":3153.974731, "-443.000000,-1151.000000":1570.7646770000001, "-472.000000,-832.000000":4996.788302999999, "-992.000000,-796.000000":3914.063738, "-391.000000,-1143.000000":3249.982455, "406.000000,1663.000000":2947.5144950000004, "-500.000000,-760.000000":2409.1792899999996, "141.000000,-1629.000000":1589.6683010000002, "37.000000,334.000000":5706.589000999999, "-82.000000,1650.000000":2304.5169760000003, "-258.000000,-128.000000":2463.538787, "-160.000000,-1140.000000":3955.944411, "398.000000,1379.000000":3133.744625, "-445.000000,1158.000000":2524.50515, "-1040.000000,-240.000000":3057.8319759999995, "-469.000000,1137.000000":2881.7311130000003, "-165.000000,-1295.000000":1508.556196, "-453.000000,1209.000000":2978.8128530000004, "13.000000,-132.000000":1793.5983840000001, "-652.000000,-208.000000":1950.6050560000003, "43.000000,-840.000000":5260.662135999999, "-155.000000,-1248.000000":1932.6672510000003, "-688.000000,-228.000000":3195.3973159999996, "-27.000000,-188.000000":3926.2190110000006, "352.000000,1389.000000":2747.951709, "-995.000000,-1519.000000":2474.1054220000005, "402.000000,1377.000000":2332.6383920000003, "7.000000,-784.000000":4342.899576999999, "-604.000000,-1360.000000":1489.126892, "-177.000000,-1355.000000":937.836732, "-492.000000,-820.000000":3778.0793289999997, "-208.000000,1128.000000":3530.605291, "420.000000,1671.000000":4577.632777, "-326.000000,-1142.000000":2502.278241, "307.000000,-1562.000000":3563.419985, "-165.000000,-1355.000000":609.349328, "-1005.000000,1383.000000":6216.601672000001, "-1011.000000,-1531.000000":5511.334352000001, "-988.000000,-776.000000":3556.838922, "-188.000000,1120.000000":2916.510908, "13.000000,302.000000":3158.929238, "-12.000000,1412.000000":1037.613669, "-204.000000,1228.000000":1258.8875859999998, "19.000000,-432.000000":4624.825684, "-623.000000,-1211.000000":1314.497231}
	//CVResults := make(map[int]float64)
	//glb.Debug.Println(3)

	totalErrorList := []int{}
	knnErrHyperParameters := make(map[int][]interface{})

	bestK := 1
	bestMinClusterRss:= 1
	bestResult := -1


	//Set algorithm parameters range:

		// KNN:
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
			//glb.Debug.Println("KNN minClusterRss :", minClusterRss)
			for j, K := range validKs { // for over KnnK
				glb.ProgressBarCurLevel = i*len(validKs)+j
				totalDistError := 0

				//glb.Debug.Println("KNN K :",K)
				//temptemptemp := make(map[string]float64)
				//glb.Debug.Println(len(crossValidationPartsList))
				for _,CVParts := range crossValidationPartsList{
					//glb.Debug.Println(CVNum)
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
					gp.Set_AlgoData(adTemp)

					distError := 0
					//FPtEMP := parameters.Fingerprint{}

					trackedPointsNum := 0
					testLocation := testFPs[testFPsOrdering[0]].Location
					for _,index := range testFPsOrdering{
						fp := testFPs[index]

						//FPtEMP = fp
						//if(fp.Location =="-165.000000,-1295.000000"){
						//glb.Warning.Println(index)
						resultDot := ""
						var err error
						err,resultDot = TrackKnn(gp, fp)
						if err != nil{
							if err.Error() == "NumofAP_lowerThan_MinApNum"{
								continue
							}
						}else{
							trackedPointsNum++
						}

						//glb.Debug.Println(fp.Location," ==== ",resultDot)
						//glb.Debug.Println(fp)

						resx,resy := getDotFromString(resultDot)
						x,y := getDotFromString(testLocation)
						//if fp.Timestamp==int64(1516794991872647445){
						//	glb.Error.Println("ResultDot = ",resultDot)
						//	glb.Error.Println("DistError = ",int(calcDist(x,y,resx,resy)))
						//}
						distError += int(calcDist(x,y,resx,resy))
						if distError < 0{
							glb.Error.Println(fp)
							glb.Error.Println(resultDot)
							_,resultDot = TrackKnn(gp, fp)
							glb.Error.Println(x,y)
							glb.Error.Println(resx,resy)
						}
						//}
					}
					if trackedPointsNum==0{
						glb.Error.Println("For loc:",testLocation," there is no fingerprint that its number of APs be more than",glb.MinApNum)
					}else{
						distError = distError/trackedPointsNum
						totalDistError += distError
					}
					//glb.Debug.Println(distError)
					//if totalDistError >0{
					//	glb.Debug.Println(totalDistError)
					//}else{
					//	glb.Error.Println(totalDistError)
					//}

					//CVResults[CVNum] = distError
					//if val,ok := crossHistory[FPtEMP.Location];ok{
					//	if val != distError{
					//		glb.Error.Println("Errrror!")
					//
					//		glb.Error.Println(FPtEMP.Location)
					//
					//		glb.Error.Println(val)
					//
					//		glb.Error.Println(len(learnedKnnData.FingerprintsOrdering))
					//		glb.Error.Println(distError)
					//
					//		distError1 := float64(0)
					//		glb.Debug.Println("1111111111111111111111111111111")
					//		for _,index := range testFPsOrdering{
					//			fp := testFPs[index]
					//			//if(fp.Location =="-165.000000,-1295.000000"){
					//			//glb.Warning.Println(index)
					//			_,resultDot := TrackKnn(gp, fp)
					//			glb.Debug.Println(fp.Location," ==== ",resultDot)
					//			glb.Debug.Println(fp)
					//
					//			resx,resy := getDotFromString(resultDot)
					//			x,y := getDotFromString(fp.Location)
					//			distError1 += calcDist(x,y,resx,resy)
					//			//}
					//		}
					//		glb.Debug.Println("2222222222222222222222222222222222222")
					//		distError2 := float64(0)
					//		for _,index := range testFPsOrdering{
					//			fp := testFPs[index]
					//
					//			_,resultDot := TrackKnn(gp, fp)
					//			glb.Debug.Println(fp.Location," ==== ",resultDot)
					//			glb.Debug.Println(fp)
					//
					//			resx,resy := getDotFromString(resultDot)
					//			x,y := getDotFromString(fp.Location)
					//			distError2 += calcDist(x,y,resx,resy)
					//			//}
					//		}
					//		glb.Error.Println(distError1)
					//		glb.Error.Println(distError2)
					//		glb.Debug.Println("3333333333333333333333333333333333333")
					//	}
					//}

					//glb.Debug.Println(CVNum)
					//glb.Debug.Println(distError)
					//temptemptemp[FPtEMP.Location] = distError

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

	glb.ProgressBarCurLevel = 0
	glb.Debug.Println(totalErrorList)
	sort.Ints(totalErrorList)
	bestResult = totalErrorList[0]
	bestErrHyperParameters := knnErrHyperParameters[bestResult]
	bestK = bestErrHyperParameters[0].(int)
	bestMinClusterRss = bestErrHyperParameters[1].(int)

	glb.Debug.Println("Best K : ",bestK)
	glb.Debug.Println("Best MinClusterRss : ",bestMinClusterRss)
	glb.Debug.Println("Minimum error = ",bestResult)




	//glb.Debug.Println("KNN K :",K)
	//temptemptemp := make(map[string]float64)
	//glb.Debug.Println(len(crossValidationPartsList))

	// Calculating each location detection accuracy :
	for _,CVParts := range crossValidationPartsList{
		//glb.Debug.Println(CVNum)
		mdTemp := gp.NewMiddleDataStruct()
		adTemp := gp.NewAlgoDataStruct()
		rdTemp := CVParts.GetTrainSet(gp)
		testFPs := CVParts.testSet.Fingerprints


		testFPsOrdering := CVParts.testSet.FingerprintsOrdering
		GetParameters(mdTemp, rdTemp)
		tempHyperParameters := []interface{}{bestK,bestMinClusterRss}
		learnedKnnData,_:= LearnKnn(mdTemp,rdTemp,tempHyperParameters)

		// Set hyper parameters
		learnedKnnData.K = bestK
		learnedKnnData.MinClusterRss = bestMinClusterRss

		adTemp.Set_KnnFPs(learnedKnnData)
		gp.Set_AlgoData(adTemp)

		distError := 0
		//FPtEMP := parameters.Fingerprint{}
		trackedPointsNum := 0
		testLocation := testFPs[testFPsOrdering[0]].Location
		for _,index := range testFPsOrdering{

			fp := testFPs[index]

			//FPtEMP = fp
			//if(fp.Location =="-165.000000,-1295.000000"){
			//glb.Warning.Println(index)
			resultDot := ""
			var err error
			err,resultDot = TrackKnn(gp, fp)

			if err != nil {
				if err.Error() == "NumofAP_lowerThan_MinApNum" {
					continue
				}
			}else{
				trackedPointsNum++
			}
			//glb.Debug.Println(fp.Location," ==== ",resultDot)
			//glb.Debug.Println(fp)

			resx,resy := getDotFromString(resultDot)
			x,y := getDotFromString(testLocation) // testLocation is fp.Location
			distError += int(calcDist(x,y,resx,resy))
			if distError < 0{
				glb.Error.Println(fp)
				glb.Error.Println(resultDot)
				_,resultDot = TrackKnn(gp, fp)
				glb.Error.Println(x,y)
				glb.Error.Println(resx,resy)
			}
			//}
		}
		if trackedPointsNum==0{
			glb.Error.Println("For loc:",testLocation," there is no fingerprint that its number of APs be more than",glb.MinApNum)
			knnLocAccuracy[testLocation] = -1
		}else{
			distError = distError/trackedPointsNum
			knnLocAccuracy[testLocation] = distError
		}



	}

	//glb.Debug.Println(temptemptemp)

	//glb.Debug.Println(crossValidationPartsList)

	//for _,c := range crossValidationPartsList{
	//	glb.Debug.Println("######################### dot:")
	//	for _,dot := range c.testSet.Fingerprints{
	//		glb.Debug.Println(dot)
	//	}
	//	glb.Debug.Println("---------------------_")
	//	//for _,ci := range c.trainSet{
	//	//	glb.Debug.Println(ci)
	//	//	glb.Debug.Println("%%%%%%%%%%")
	//	//}
	//	glb.Debug.Println("#########################")
	//}
	// set crossvalidation results
	rs := gp.Get_ResultData()
	glb.Debug.Println(gp.Get_Name())
	rs.Set_AlgoAccuracy("knn",bestResult)
	for loc,accuracy := range knnLocAccuracy{
		rs.Set_AlgoLocAccuracy("knn",loc,accuracy)
	}
	glb.Debug.Println(dbm.GetCVResults(gp.Get_Name()))

	// set main parameters
	md := gp.NewMiddleDataStruct()
	GetParameters(md, rd)

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
				_,resultDot := TrackKnn(knnJob.gp, fp)
				//glb.Debug.Println(fp.Location," ==== ",resultDot)
				//glb.Debug.Println(fp)

				resx,resy := getDotFromString(resultDot)
				x,y := getDotFromString(fp.Location)
				distError += int(calcDist(x,y,resx,resy))
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

func getDotFromString(dotStr string) (float64,float64){
	x_y := strings.Split(dotStr, ",")
	locXstr := x_y[0]
	locYstr := x_y[1]
	locX, _ := strconv.ParseFloat(locXstr, 64)
	locY, _ := strconv.ParseFloat(locYstr, 64)
	return locX,locY
}
func calcDist(x1,y1,x2,y2 float64) float64{
	return math.Pow(math.Pow(float64(x1-x2), 2)+math.Pow(float64(y1-y2), 2),0.5)
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
