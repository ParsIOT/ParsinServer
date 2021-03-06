package algorithms

import (
	"ParsinServer/algorithms/clustering"
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"math"
	"sort"
	"strings"
)

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
			if loc == locMain {
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

//func GetParameters(md *dbm.MiddleDataStruct, rd dbm.RawDataStruct) {
//	//
//	//persistentPs, err := dbm.OpenPersistentParameters(group) //persistentPs is just like ps but with renamed network name; e.g.: "0" -> "1"
//	//if err != nil {
//	//	//log.Fatal(err)
//	//	glb.Error.Println(err)
//	//}
//	fingerprints := rd.Fingerprints
//	fingerprintsOrdering := rd.FingerprintsOrdering
//
//	//glb.Error.Println("d")
//	md.NetworkMacs = make(map[string]map[string]bool)
//	md.NetworkLocs = make(map[string]map[string]bool)
//	md.UniqueMacs = []string{}
//	md.UniqueLocs = []string{}
//	md.MacCount = make(map[string]int)
//	md.MacCountByLoc = make(map[string]map[string]int)
//	//md.Loaded = true
//	//opening the db
//	//db, err := boltOpen(path.Join(glb.RuntimeArgs.SourcePath, group+".db"), 0600, nil)
//	//defer db.Close()
//	// if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	macs := []string{}
//
//	// Get all parameters that don't need a network graph (?)
//	for _, v1 := range fingerprintsOrdering {
//
//		//log.Println("calculateResults=true")
//		v2 := fingerprints[v1]
//
//		// append the fingerprint location to UniqueLocs array if doesn't exist in it.
//		if !glb.StringInSlice(v2.Location, md.UniqueLocs) {
//			md.UniqueLocs = append(md.UniqueLocs, v2.Location)
//		}
//
//		// MacCountByLoc initialization for new location
//		if _, ok := md.MacCountByLoc[v2.Location]; !ok {
//			md.MacCountByLoc[v2.Location] = make(map[string]int)
//		}
//
//		//// building network
//		//macs := []string{}
//
//		for _, router := range v2.WifiFingerprint {
//			// building network
//			macs = append(macs, router.Mac)
//
//			// append the fingerprint mac to UniqueMacs array if doesn't exist in it.
//			if !glb.StringInSlice(router.Mac, md.UniqueMacs) {
//				md.UniqueMacs = append(md.UniqueMacs, router.Mac)
//			}
//
//			// mac count
//			if _, ok := md.MacCount[router.Mac]; !ok {
//				md.MacCount[router.Mac] = 0
//			}
//			md.MacCount[router.Mac]++
//
//			// mac by location count
//			if _, ok := md.MacCountByLoc[v2.Location][router.Mac]; !ok {
//				md.MacCountByLoc[v2.Location][router.Mac] = 0
//			}
//			md.MacCountByLoc[v2.Location][router.Mac]++
//		}
//
//		// building network
//		//ps.NetworkMacs = buildNetwork(ps.NetworkMacs, macs)
//	}
//	// todo: network definition and buildNetwork() must be redefined
//	md.NetworkMacs = clustring.BuildNetwork(md.NetworkMacs, macs)
//	md.NetworkMacs = clustring.MergeNetwork(md.NetworkMacs)
//
//	//Error.Println("ps.Networkmacs", ps.NetworkMacs)
//	// Rename the NetworkMacs
//	//if len(persistentPs.NetworkRenamed) > 0 {
//	//	newNames := []string{}
//	//	for k := range persistentPs.NetworkRenamed {
//	//		newNames = append(newNames, k)
//	//
//	//	}
//	//	//todo: \/ wtf? Rename procedure could be redefined better.
//	//	for n := range md.NetworkMacs {
//	//		renamed := false
//	//		for mac := range md.NetworkMacs[n] {
//	//			for renamedN := range persistentPs.NetworkRenamed {
//	//				if glb.StringInSlice(mac, persistentPs.NetworkRenamed[renamedN]) && !glb.StringInSlice(n, newNames) {
//	//					md.NetworkMacs[renamedN] = make(map[string]bool)
//	//					for k, v := range md.NetworkMacs[n] {
//	//						md.NetworkMacs[renamedN][k] = v //copy ps.NetworkMacs[n] to ps.NetworkMacs[renamedN]
//	//					}
//	//					delete(md.NetworkMacs, n)
//	//					renamed = true
//	//				}
//	//				if renamed {
//	//					break
//	//				}
//	//			}
//	//			if renamed {
//	//				break
//	//			}
//	//		}
//	//	}
//	//}
//
//	// Get the locations for each graph (Has to have network built first)
//
//	for _, v1 := range fingerprintsOrdering {
//
//		v2 := fingerprints[v1]
//		//todo: Make the macs array just once for each fingerprint instead of repeating the process
//
//		macs := []string{}
//		for _, router := range v2.WifiFingerprint {
//			macs = append(macs, router.Mac)
//		}
//		//todo: ps.NetworkMacs is created from mac array; so it seems that hasNetwork function doesn't do anything useful!
//		networkName, inNetwork := clustring.HasNetwork(md.NetworkMacs, macs)
//		if inNetwork {
//			if _, ok := md.NetworkLocs[networkName]; !ok {
//				md.NetworkLocs[networkName] = make(map[string]bool)
//			}
//			if _, ok := md.NetworkLocs[networkName][v2.Location]; !ok {
//				md.NetworkLocs[networkName][v2.Location] = true
//			}
//		}
//	}
//
//	//calculate locCount
//	locations := []string{}
//	for _, fpIndex := range fingerprintsOrdering {
//		fp := fingerprints[fpIndex]
//		locations = append(locations, fp.Location)
//	}
//	md.LocCount = glb.DuplicateCountString(locations)
//}

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

//Note: Use it just one time(not use it in calculatelearn, use it in buildGroup)
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
				maxValidFPDistAVG := float64(100) // 100 cm

				tempFingerprintsData2 := make(map[string]parameters.Fingerprint)
				for fpOMain, fpMain := range tempFingerprintsData {
					adjacentFPs := []parameters.Fingerprint{}
					for fpO, fp := range tempFingerprintsData {
						if fpO == fpOMain {
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

	//4.RPFRadius
	validRPFRadius := []float64{glb.DefaultRPFRadiusRange[0]}
	if len(glb.DefaultRPFRadiusRange) > 1 {
		validRPFRadius = glb.MakeRangeFloat(glb.DefaultRPFRadiusRange[0], glb.DefaultRPFRadiusRange[1], glb.DefaultRPFRadiusRange[2])
	}
	rpfRadiusRange := knnConfig.RPFRadiusRange
	if len(rpfRadiusRange) == 1 {
		validRPFRadius = glb.MakeRangeFloat(rpfRadiusRange[0], rpfRadiusRange[0])
	} else if len(rpfRadiusRange) == 2 {
		validRPFRadius = glb.MakeRangeFloat(rpfRadiusRange[0], rpfRadiusRange[1], float64(1))
	} else if len(rpfRadiusRange) == 3 {
		validRPFRadius = glb.MakeRangeFloat(rpfRadiusRange[0], rpfRadiusRange[1], rpfRadiusRange[2])
	} else {
		glb.Error.Println("validRPFRadius:", validRPFRadius)
		glb.Error.Println("Can't set valid RPFRadiusRange values")
	}

	// Set length of calculation progress bar
	// This is shared between all threads, so it's invalid when two calculateLearn thread run
	calculationLen := len(validMinClusterRSSs) * len(validKs) * len(validMaxEuclideanRssDists) *
		len(validMaxEuclideanRssDists) * len(validBLEFactors) * len(validRPFRadius)
	glb.ProgressBarLength = calculationLen

	adTemp := tempGp.NewAlgoDataStruct()
	rdTemp := tempGp.Get_RawData()

	//allErrDetailsList = make([][]int,calculationLen)

	paramUniqueKey := 0 // just creating unique key for each possible the parameters permutation
	for i1, maxEuclideanRssDist := range validMaxEuclideanRssDists {
		for i2, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
			for i3, K := range validKs { // for over KnnK
				for i4, bleFactor := range validBLEFactors {
					for i5, rpfRadius := range validRPFRadius {

						glb.ProgressBarCurLevel = i1*len(validMinClusterRSSs) + i2*len(validKs) +
							i3*len(validBLEFactors) + i4*len(validRPFRadius) + i5
						totalDistError := 0

						tempHyperParameters := parameters.NewKnnHyperParameters()
						//glb.Error.Println(tempHyperParameters)
						tempHyperParameters.K = K
						tempHyperParameters.MinClusterRss = minClusterRss
						tempHyperParameters.MaxEuclideanRssDist = maxEuclideanRssDist
						tempHyperParameters.BLEFactor = bleFactor
						tempHyperParameters.RPFRadius = rpfRadius

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
	}

	glb.Debug.Println((allErrDetails))

	glb.ProgressBarCurLevel = 0 // reset progressBar level

	// Select best hyperParameters
	//glb.Debug.Println(totalErrorList)

	//bestKey, sortedErrDetails, newErrorMap, err := SelectLowestError(allErrDetails,"FalseRate")
	bestKey, sortedErrDetails, newErrorMap := SelectBestFromErrMapByParameterPriority(allErrDetails, allHyperParamDetails)
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

	//4.RPFRadius
	validRPFRadius := []float64{glb.DefaultRPFRadiusRange[0]}
	if len(glb.DefaultRPFRadiusRange) > 1 {
		validRPFRadius = glb.MakeRangeFloat(glb.DefaultRPFRadiusRange[0], glb.DefaultRPFRadiusRange[1], glb.DefaultRPFRadiusRange[2])
	}
	rpfRadiusRange := knnConfig.RPFRadiusRange
	if len(rpfRadiusRange) == 1 {
		validRPFRadius = glb.MakeRangeFloat(rpfRadiusRange[0], rpfRadiusRange[0])
	} else if len(rpfRadiusRange) == 2 {
		validRPFRadius = glb.MakeRangeFloat(rpfRadiusRange[0], rpfRadiusRange[1], float64(1))
	} else if len(rpfRadiusRange) == 3 {
		validRPFRadius = glb.MakeRangeFloat(rpfRadiusRange[0], rpfRadiusRange[1], rpfRadiusRange[2])
	} else {
		glb.Error.Println("validRPFRadius:", validRPFRadius)
		glb.Error.Println("Can't set valid RPFRadiusRange values")
	}

	// Set length of calculation progress bar
	// This is shared between all threads, so it's invalid when two calculateLearn thread run
	glb.ProgressBarLength = len(validMinClusterRSSs) * len(validKs) * len(validMaxEuclideanRssDists) *
		len(validMaxEuclideanRssDists) * len(validBLEFactors) * len(validRPFRadius)

	adTemp := tempGp.NewAlgoDataStruct()
	rdTemp := tempGp.Get_RawData()
	for i1, maxEuclideanRssDist := range validMaxEuclideanRssDists {
		for i2, minClusterRss := range validMinClusterRSSs { // for over minClusterRss
			for i3, K := range validKs { // for over KnnK
				for i4, bleFactor := range validBLEFactors {
					for i5, rpfRadius := range validRPFRadius {
						glb.ProgressBarCurLevel = i1*len(validMinClusterRSSs) + i2*len(validKs) +
							i3*len(validBLEFactors) + i4*len(validRPFRadius) + i5

						totalDistError := 0

						tempHyperParameters := parameters.NewKnnHyperParameters()
						//glb.Error.Println(tempHyperParameters)
						tempHyperParameters.K = K
						tempHyperParameters.MinClusterRss = minClusterRss
						tempHyperParameters.MaxEuclideanRssDist = maxEuclideanRssDist
						tempHyperParameters.BLEFactor = bleFactor
						tempHyperParameters.RPFRadius = rpfRadius

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

						glb.Debug.Printf("Knn error (minClusterRss=%d,K=%d,maxEuclideanRssDist=%d,rpfRadius=%d) = %d \n", minClusterRss, K, maxEuclideanRssDist, rpfRadius, totalDistError)

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
func SelectBestFromErrMap(allErrDetails map[int][]int) ([]int, []int, map[int]int) {
	MainErrAlgorithm := Mean

	bestKeys, sortedErrDetails, newErrorMap, err := SelectLowestError(allErrDetails, MainErrAlgorithm)
	if err != nil { // Use Mean algorithm
		glb.Error.Println(err)

		// find best mean error
		bestKeys, sortedErrDetails, newErrorMap, err = SelectLowestError(allErrDetails, Mean)
		if err != nil {
			glb.Error.Println(err)
		}
		//bestResult := sortedErrDetails[0]
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
	//glb.Debug.Println(newErrorMap)

	return bestKeys, sortedErrDetails, newErrorMap
	/*	for _, i := range sortedErrDetails {
			glb.Debug.Println("-----------------------------")
			glb.Debug.Println("Hyper Params:", allHyperParamDetails[i])
			glb.Debug.Println("Error:", newErrorMap[i])
		}

		for _, i := range sortedErrDetails {
			glb.Debug.Println(allHyperParamDetails[i], " ", newErrorMap[i])
		}*/
}

// return sorted list and indexes(according to main list) of  KnnHyperParameters list according to KnnHyperParameters priority
func SortKnnHyperParametersList(rawKnnHyperParamList []parameters.KnnHyperParameters) ([]int, []parameters.KnnHyperParameters) {

	//K_Priority              int = 0
	//MinCluster_Priority     int = 1
	//MaxEculidDist_Priority  int = 2
	//BLEFactor_Priority      int = 3
	//GraphFactor_Priority    int = 4
	//DSAMaxMovement_Priority int = 5

	const (
		ascnd      int = 0 //lowest value is better
		dscnd      int = 1 //most value is better
		mddleAscnd int = 2 // near middle is better
		mddleDscnd int = 3 // far from middle is better
		sumAscnd   int = 4 // near middle is better
		sumDscnd   int = 5 // far from middle is better
	)
	type ParamOrder struct {
		K                   int
		MinClusterRss       int
		MaxEuclideanRssDist int
		BLEFactor           int
		GraphFactor         int
		MaxMovement         int
	}
	paramOrder := ParamOrder{
		K:                   ascnd,
		MinClusterRss:       dscnd,
		MaxEuclideanRssDist: ascnd,
		BLEFactor:           mddleAscnd,
		GraphFactor:         sumAscnd,
		MaxMovement:         ascnd,
	}

	type IndexedKnnHyperParam struct {
		Index              int
		KnnHyperParameters parameters.KnnHyperParameters
	}

	indexedKnnHyperParamList := []IndexedKnnHyperParam{}
	for i, knnHyperParam := range rawKnnHyperParamList {
		indexedKnnHyperParam := IndexedKnnHyperParam{
			Index:              i,
			KnnHyperParameters: knnHyperParam,
		}
		indexedKnnHyperParamList = append(indexedKnnHyperParamList, indexedKnnHyperParam)
	}

	sort.Slice(indexedKnnHyperParamList, func(i, j int) bool {
		rawKnnHyperParamListi := indexedKnnHyperParamList[i].KnnHyperParameters
		rawKnnHyperParamListj := indexedKnnHyperParamList[j].KnnHyperParameters

		Ki, Kj := rawKnnHyperParamListi.K, rawKnnHyperParamListj.K
		MinClusterRssi, MinClusterRssj := rawKnnHyperParamListi.MinClusterRss, rawKnnHyperParamListj.MinClusterRss
		MaxEuclideanRssDisti, MaxEuclideanRssDistj := rawKnnHyperParamListi.MaxEuclideanRssDist, rawKnnHyperParamListj.MaxEuclideanRssDist
		BLEFactori, BLEFactorj := rawKnnHyperParamListi.BLEFactor, rawKnnHyperParamListj.BLEFactor
		GraphFactorsi, GraphFactorsj := rawKnnHyperParamListi.GraphFactors, rawKnnHyperParamListj.GraphFactors
		MaxMovementi, MaxMovementj := rawKnnHyperParamListi.MaxMovement, rawKnnHyperParamListj.MaxMovement

		if Ki < Kj {
			return (paramOrder.K == ascnd)
		}
		if Ki > Kj {
			return (paramOrder.K == dscnd)
		}

		// MinCluster
		if MinClusterRssi < MinClusterRssj {
			return (paramOrder.MinClusterRss == ascnd)
		}
		if MinClusterRssi > MinClusterRssj {
			return (paramOrder.MinClusterRss == dscnd)
		}

		// MaxEuclideanRssDist
		if MaxEuclideanRssDisti < MaxEuclideanRssDistj {
			return (paramOrder.MaxEuclideanRssDist == ascnd)
		}
		if MaxEuclideanRssDisti > MaxEuclideanRssDistj {
			return (paramOrder.MaxEuclideanRssDist == dscnd)
		}

		// BLEFactor
		BLEFactorDisti, BLEFactorDistj := math.Abs(BLEFactori-1.0), math.Abs(BLEFactorj-1.0)
		if BLEFactorDisti < BLEFactorDistj {
			return (paramOrder.BLEFactor == mddleAscnd)
		}
		if BLEFactorDisti > BLEFactorDistj {
			return (paramOrder.BLEFactor == mddleDscnd)
		}

		// GraphFactors
		GraphFactorsSumi, GraphFactorsSumj := float64(0.0), float64(0.0)
		for _, val := range GraphFactorsi {
			GraphFactorsSumi += val
		}
		for _, val := range GraphFactorsj {
			GraphFactorsSumj += val
		}

		if GraphFactorsSumi < GraphFactorsSumj {
			return (paramOrder.GraphFactor == sumAscnd)
		}
		if GraphFactorsSumi > GraphFactorsSumj {
			return (paramOrder.GraphFactor == sumDscnd)
		}

		// MaxEuclideanRssDist
		if MaxMovementi < MaxMovementj {
			return (paramOrder.MaxMovement == ascnd)
		}
		if MaxMovementi > MaxMovementj {
			return (paramOrder.MaxMovement == dscnd)
		}

		// ELSE
		return true // Completely equal!
	})

	sortedIndexes := []int{}
	sortedKnnHyperParamList := []parameters.KnnHyperParameters{}
	for _, indexedKnnHyperParam := range indexedKnnHyperParamList {
		sortedIndexes = append(sortedIndexes, indexedKnnHyperParam.Index)
		sortedKnnHyperParamList = append(sortedKnnHyperParamList, indexedKnnHyperParam.KnnHyperParameters)
	}
	return sortedIndexes, sortedKnnHyperParamList
}

// this function get SelectBestFromErrMap output if there are more than one key in bestkeys list then
// it'll find best key according to KnnHyperParameters priority
func SelectBestFromErrMapByParameterPriority(allErrDetails map[int][]int, allHyperParamDetails map[int]parameters.KnnHyperParameters) (int, []int, map[int]int) {
	bestKeys, sortedErrDetails, newErrorMap := SelectBestFromErrMap(allErrDetails)
	//glb.Error.Println(bestKeys)

	bestHyperParams := []parameters.KnnHyperParameters{}
	for _, key := range bestKeys {
		hyperParam := allHyperParamDetails[key]
		bestHyperParams = append(bestHyperParams, hyperParam)
	}
	//sortedIndexes, sortedKnnHyperParamList := SortKnnHyperParametersList(bestHyperParams)
	sortedIndexes, _ := SortKnnHyperParametersList(bestHyperParams)
	bestKey := bestKeys[sortedIndexes[0]]
	return bestKey, sortedErrDetails, newErrorMap
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
		glb.Debug.Println("Calculation for graphfactor=", graphFactor, " ended ")
	}

	/*	sort.Ints(totalErrorList)
		bestResult = totalErrorList[0]
		bestErrHyperParameters := knnErrHyperParameters[bestResult]*/

	bestKey, sortedErrDetails, newErrorMap := SelectBestFromErrMapByParameterPriority(allErrDetails, allHyperParamDetails)
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
		glb.Debug.Println("Calculation for maxMovement=", maxMovement, " ended ")
	}

	/*	sort.Ints(totalErrorList)
		bestResult = totalErrorList[0]
		bestErrHyperParameters := knnErrHyperParameters[bestResult]*/

	bestKey, sortedErrDetails, newErrorMap := SelectBestFromErrMapByParameterPriority(allErrDetails, allHyperParamDetails)
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

// Find best key(list of best key when some same error values exist)that has lowest error list:
// method : "FalseRate", "AUC" ,"LatterPercentile","Mean"
func SelectLowestError(errMap map[int][]int, method string) ([]int, []int, map[int]int, error) {
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
		bestKeys, sortedKey := glb.GetLowestValueKeys(falseCountMap)
		return bestKeys, sortedKey, falseCountMap, nil
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
		bestKeys, sortedKey := glb.GetLowestValueKeys(falseCountMap)
		return bestKeys, sortedKey, falseCountMap, nil
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
		bestKeys, sortedKey := glb.GetLowestValueKeys(percentileMap)
		return bestKeys, sortedKey, percentileMap, nil
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
		bestKeys, sortedKey := glb.GetLowestValueKeys(meanErrMap)
		return bestKeys, sortedKey, meanErrMap, nil
	}
	return []int{}, []int{}, make(map[int]int), errors.New("Invalid method parameters")
}
