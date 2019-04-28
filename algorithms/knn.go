package algorithms

import (
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"errors"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var knn_regression bool
var minkowskyQ float64
var maxrssInNormal, minrssInNormal float64
//var topRssList []int
var distAlgo string
var MaxEuclideanRssDist float64 // used in thread so must be declared as global variable(just reading so there's not any race condition)
var BLEFactor float64
var uniqueMacs []string
//var ValidKs []int = defaultValidKs()
//var ValidMinClusterRSSs []int = defaultValidMinClusterRSSs()
//
//
//
//func defaultValidKs() []int {
//	validKs := []int{}
//	for i:=glb.DefaultKnnKRange[0];i<=glb.DefaultKnnKRange[1];i++{
//		validKs = append(validKs,i)
//	}
//	return validKs
//}
//
//func defaultValidMinClusterRSSs() []int {
//	validMinClusterRSSs := []int{}
//	for i:=-60;i>=-90;i--{
//		validMinClusterRSSs = append(validMinClusterRSSs,i)
//	}
//	return validMinClusterRSSs
//}

func init() {
	knn_regression = true
	minkowskyQ = 2
	distAlgo = "Euclidean" // Euclidean, Cosine, NewEuclidean, RedpinEuclidean, CombinedCosEuclid
	//topRssList = []int{-60,-79,-90}
	maxrssInNormal = -30.0
	minrssInNormal = float64(glb.MinRssi)
}

type resultW struct {
	fpTime string
	weight float64
}

type jobW struct {
	fpTime     string
	mac2RssCur map[string]int
	mac2RssFP  map[string]int
}

func LearnKnn(gp *dbm.Group, hyperParameters parameters.KnnHyperParameters) (parameters.KnnFingerprints, error) {
	//Debug.Println(Cosine([]float64{1,2,3},[]float64{1,2,4}))
	//jsonFingerprint = calcMacRate(jsonFingerprint,false)
	//K := hyperParameters[0].(int)
	rd := gp.Get_RawData()

	MinClusterRSS := hyperParameters.MinClusterRss //komeil: min threshold for determining whether ...
	// a fingerprint is in the cluster of a beacon or not
	//glb.Debug.Printf("Knn is running (K:%d, MinClusterRss:%d)\n",K,MinClusterRSS)
	//jsonFingerprint = calcMacJustRate(jsonFingerprint,false)

	//Debug.Println(jsonFingerprint)
	//if (len(jsonFingerprint.WifiFingerprint) < minApNum) {
	//	err := glb.Errors.New("Location names aren't in the format of x,y")
	//	return err,"NaN,Nan"
	//}
	//Debug.Println(jsonFingerprint)
	//glb.RuntimeArgs.NeedToFilter[jsonFingerprint.Group] = true

	//fingerprints := make(map[string]parameters.Fingerprint)
	//var fingerprintsOrdering []string
	clusters := make(map[string][]string) // komeil: key of map: Mac - value: fpTime
	//var err error

	fingerprints := rd.Fingerprints
	fingerprintsOrdering := rd.FingerprintsOrdering // komeil: timestamps of fingerprints as id

	//fingerprintsOrdering,fingerprints,err = dbm.GetLearnFingerPrints(groupName,true)
	//if err!=nil {
	//	return err
	//}

	for fpTime, fp := range fingerprints {
		for _, rt := range fp.WifiFingerprint { //rt ==> Router = mac + RSS of an Access Point
			if (rt.Rssi >= MinClusterRSS) {
				clusters[rt.Mac] = append(clusters[rt.Mac], fpTime)
			}
		}
	}

	node2FPs := make(map[string][]string)
	graphMapPointer := gp.Get_ConfigData().Get_GroupGraph()
	if !graphMapPointer.IsEmpty() {
		for fpTime, fp := range fingerprints {
			nearNodeGraph := graphMapPointer.GetNearestNode(fp.Location)
			//glb.Debug.Println("near node Graph: ",nearNodeGraph.Label)
			if nearNodeGraph == nil {
				glb.Error.Println("Nearest node is empty!")
				continue
			}
			nodeLabel := nearNodeGraph.Label
			if tempFPList, ok := node2FPs[nodeLabel]; ok {
				node2FPs[nodeLabel] = append(tempFPList, fpTime)
			} else {
				if nearNodeGraph == nil {
					glb.Error.Println("*** near node was nil for ", fp.Location)
				} else {
					node2FPs[nodeLabel] = []string{fpTime}
				}
			}
		}
	}
	//// Cluster print
	//for key,val := range clusters{
	//	fmt.Println("mac: "+key+" ")
	//	for _,fp:=range val{
	//		fmt.Println(fingerprints[fp])
	//	}
	//	fmt.Println("---------------------------------")
	//}

	// RBF calculation
	// check graph is not empty
	RPFs := make(map[string]float64)
	if !graphMapPointer.IsEmpty() {
		for fpTime, fp := range fingerprints {
			RPFs[fpTime] = CalculateDotRPF(fp.Location, graphMapPointer)
		}
	}

	// Add to knnData in db

	var tempKnnFingerprints parameters.KnnFingerprints
	tempKnnFingerprints.FingerprintsInMemory = fingerprints
	tempKnnFingerprints.FingerprintsOrdering = fingerprintsOrdering
	tempKnnFingerprints.Clusters = clusters
	tempKnnFingerprints.Node2FPs = node2FPs
	tempKnnFingerprints.RPFs = RPFs
	tempKnnFingerprints.HyperParameters = hyperParameters
	//dbm.GM.GetGroup(groupName).Get_AlgoData().Set_KnnFPs(tempKnnFingerprints)

	//err = dbm.SetKnnFingerprints(tempKnnFingerprints, groupName)
	//if err != nil {
	//	glb.Error.Println(err)
	//	return err
	//}
	//
	//// Set in cache
	//go dbm.SetKnnFPCache(groupName,tempKnnFingerprints)
	return tempKnnFingerprints, nil
}

func ConvertRSS2Dist(rss int) int {
	return int((float64(rss)+60)*(float64(-300)/float64(7)) + 100)
}

func GetMiddleOfLine(dotVals []float64) (float64, float64) {
	x1 := dotVals[0]
	y1 := dotVals[1]
	w1 := dotVals[2]

	x2 := dotVals[3]
	y2 := dotVals[4]
	w2 := dotVals[5]

	return (x1*w1 + x2*w2) / (w1 + w2), (y1*w1 + y2*w2) / (w1 + w2)
}

func TriangulateWith3Point(triangulationVals []float64) (float64, float64) {

	x1 := triangulationVals[0]
	y1 := triangulationVals[1]
	r1 := triangulationVals[2]

	x2 := triangulationVals[3]
	y2 := triangulationVals[4]
	r2 := triangulationVals[5]

	x3 := triangulationVals[6]
	y3 := triangulationVals[7]
	r3 := triangulationVals[8]

	A := 2 * (x2 - x1)
	B := 2 * (y2 - y1)
	C := math.Pow(r1, 2) - math.Pow(r2, 2) - math.Pow(x1, 2) + math.Pow(x2, 2) - math.Pow(y1, 2) + math.Pow(y2, 2)
	D := 2 * (x3 - x2)
	E := 2 * (y3 - y2)
	F := math.Pow(r2, 2) - math.Pow(r3, 2) - math.Pow(x2, 2) + math.Pow(x3, 2) - math.Pow(y2, 2) + math.Pow(y3, 2)

	resX := (C*E - F*B) / (E*A - B*D)
	resY := (C*D - A*F) / (B*D - A*E)

	distFrom1 := glb.CalcDist(resX, resY, x1, y1)
	if distFrom1 > 3*r1 {
		//return x1,y1
		glb.Error.Println("out of band!")
		glb.Error.Println([]float64{x1, y1, 1 / r1, x2, y2, 1 / r2})
		return GetMiddleOfLine([]float64{x1, y1, 1 / r1, x2, y2, 1 / r2})
	}
	//glb.Debug.Println(math.Pow(x1-resX,2)+math.Pow(y1-resY,2)-math.Pow(r1,2))
	//glb.Debug.Println(math.Pow(x2-resX,2)+math.Pow(y2-resY,2)-math.Pow(r2,2))
	//glb.Debug.Println(math.Pow(x3-resX,2)+math.Pow(y3-resY,2)-math.Pow(r3,2))

	return resX, resY
}

func TrackKnn(gp *dbm.Group, curFingerprint parameters.Fingerprint, historyConsidered bool) (error, string, map[string]float64) {

	//rd := gp.Get_RawData()
	//md := gp.Get_MiddleData()

	tempKnnFingerprints := gp.Get_AlgoData().Get_KnnFPs()
	knnConfig := gp.Get_ConfigData().Get_KnnConfig()

	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var mainFingerprintsOrdering []string
	var fingerprintsOrdering []string
	clusters := make(map[string][]string)
	node2FPs := make(map[string][]string)
	RPFs := make(map[string]float64)
	/*
		// Proximity solution:
		apsLocation := map[string][]float64{
			"oof":[]float64{-1000.0,-920.0},
			"sham":[]float64{-840.0,-40.0},
			"komeil":[]float64{-940.0,920.0},
			"hadi4":[]float64{-275.0,-375.0},
			"hadi5":[]float64{-250.0,450.0},
		}

		mac2RssCurTemp := getMac2Rss(curFingerprint.WifiFingerprint)
		mac2RssCurTempfloat64 := make(map[string]float64)
		for mac,rss := range mac2RssCurTemp {
			mac2RssCurTempfloat64[mac] = float64(rss)
		}
		sortedMacByRSS := glb.SortReverseDictByVal(mac2RssCurTempfloat64)

		proximityRss := -75

		resultX := float64(0)
		resultY := float64(0)

		if len(sortedMacByRSS) >= 3{ //triangulation
			glb.Error.Println("########### \n Traingulation: ")
			glb.Error.Println(curFingerprint.Username+":")
			traingulationVals := []float64{}
			for _,mac := range sortedMacByRSS[:3]{
				apXY := apsLocation[mac]
				x := apXY[0]
				y := apXY[1]
				r := float64(ConvertRSS2Dist(mac2RssCurTemp[mac]))
				if (r < 0){
					r = 1
				}
				glb.Error.Println(mac)
				glb.Error.Println(r)
				traingulationVals = append(traingulationVals, x)
				traingulationVals = append(traingulationVals, y)
				traingulationVals = append(traingulationVals, r)
			}
			resultX, resultY := TriangulateWith3Point(traingulationVals)
			glb.Error.Println(glb.IntToString(int(resultX)) + ".0," + glb.IntToString(int(resultY)) + ".0")
			return nil, glb.IntToString(int(resultX)) + ".0," + glb.IntToString(int(resultY)) + ".0", nil
		}else { //proximity

			nearMacs := []string{}
			for _, mac := range sortedMacByRSS {
				rss := mac2RssCurTemp[mac]
				if rss > proximityRss {
					if _, ok := apsLocation[mac]; ok {
						//glb.Error.Println("###########")
						//glb.Error.Println(curFingerprint.Username+":")
						//glb.Error.Println(mac)
						//glb.Error.Println(mac2RssCurTemp)
						nearMacs = append(nearMacs, mac)
					} else {
						glb.Error.Println("this mac doesn't exists in apsLocation map, ", mac)
					}
				} else {
					if len(nearMacs) != 0 {
						glb.Error.Println(curFingerprint.Username + ":")
						glb.Error.Println(nearMacs)
						glb.Error.Println(mac2RssCurTemp)
						sumW := float64(0)
						for _, mac := range nearMacs {
							w := float64(mac2RssCurTemp[mac] - proximityRss)
							sumW += w
							apXY := apsLocation[mac]
							resultX += apXY[0] * w
							resultY += apXY[1] * w
						}
						resultX /= sumW
						resultY /= sumW
						return nil, glb.IntToString(int(resultX)) + ".0," + glb.IntToString(int(resultY)) + ".0", nil
					}
				}
			}

		}*/
	//tempKnnFingerprints, ok := dbm.GetKnnFPCache(curFingerprint.Group)
	//if ok {
	//	//Debug.Println(tempKnnFingerprints)
	//	fingerprintsInMemory = tempKnnFingerprints.FingerprintsInMemory
	//	mainFingerprintsOrdering = tempKnnFingerprints.FingerprintsOrdering
	//	clusters = tempKnnFingerprints.Clusters
	//
	//}else{
	//	// get knnFp from db
	//	var tempKnnFingerprints parameters.KnnFingerprints
	//	var err error
	//
	//	tempKnnFingerprints,err = dbm.GetKnnFingerprints(curFingerprint.Group)
	//	if err!=nil{
	//		glb.Error.Println(err)
	//	}
	//	fingerprintsInMemory = tempKnnFingerprints.FingerprintsInMemory
	//	mainFingerprintsOrdering = tempKnnFingerprints.FingerprintsOrdering
	//	clusters = tempKnnFingerprints.Clusters
	//}

	//if strconv.Itoa(int(curFingerprint.Timestamp)) == "1516796888995082812"{
	//	glb.Error.Println("!")
	//}
	//tempKnnFingerprints := dbm.GM.GetGroup(curFingerprint.Group).Get_AlgoData().Get_KnnFPs()
	fingerprintsInMemory = tempKnnFingerprints.FingerprintsInMemory
	mainFingerprintsOrdering = tempKnnFingerprints.FingerprintsOrdering
	clusters = tempKnnFingerprints.Clusters
	hyperParams := tempKnnFingerprints.HyperParameters
	//node2FPs = tempKnnFingerprints.Node2FPs
	node2FPs = tempKnnFingerprints.Node2FPs
	RPFs = tempKnnFingerprints.RPFs

	uniqueMacs = gp.Get_MiddleData().Get_UniqueMacs()

	MaxEuclideanRssDist = float64(hyperParams.MaxEuclideanRssDist)
	BLEFactor = float64(hyperParams.BLEFactor)
	//BLEFactor = float64(1.5)
	MaxMovement := float64(hyperParams.MaxMovement)
	//tempList := []string{}
	//tempList = append(tempList,mainFingerprintsOrdering...)
	//sort.Sort(sort.StringSlice(tempList))
	//
	//sum := int64(0)
	//for _,i := range tempList{
	//	num,_:=strconv.ParseInt(i, 10, 64)
	//	num = num % 100000
	//	sum += num
	//
	//}
	//glb.Debug.Println(sum/int64(len(tempList)))

	//if curFingerprint.Location=="-165.000000,-1295.000000"{
	//	for key,val := range fingerprintsInMemory{
	//		if val.Location == curFingerprint.Location{
	//			glb.Warning.Println(curFingerprint.Timestamp)
	//			glb.Warning.Println(key)
	//		}
	//	}
	//}
	//show := false
	//if curFingerprint.Location == ""{
	//		for _,fp := range curFingerprint.WifiFingerprint{
	//			if fp.Mac == "01:17:C5:97:1B:44" && fp.Rssi == -72{
	//				//glb.Debug.Println(curFingerprint)
	//				show = true
	//			}
	//		}
	//}

	/*	tempKnnFpOrdering := []string{}
		for _,fpTime := range mainFingerprintsOrdering{
			if len(fingerprintsInMemory[fpTime].WifiFingerprint)>= glb.MinApNum{
				tempKnnFpOrdering = append(tempKnnFpOrdering,fpTime)
			}
		}
		mainFingerprintsOrdering = tempKnnFpOrdering*/

	// fingerprintOrdering Creation according to clusters and rss rates
	repeatFP := make(map[string]int)
	for _, fpTime := range mainFingerprintsOrdering {
		repeatFP[fpTime] = 1
	}

	if glb.MinRssClustringEnabled || hyperParams.MinClusterRss == 0 {
		AtleastInOneCluster := false // komeil: a variable to decide if it is needed to search all fingerprints instead of one or some clusters
		for _, rt := range curFingerprint.WifiFingerprint {
			if (rt.Rssi >= hyperParams.MinClusterRss) {
				if cluster, ok := clusters[rt.Mac]; ok {
					//glb.Error.Println(rt.Mac,":",rt.Rssi)
					AtleastInOneCluster = true
					for _, fpTimeMem := range cluster {
						//if !glb.StringInSlice(fpTimeMem,fingerprintsOrdering){
						fingerprintsOrdering = append(fingerprintsOrdering, fpTimeMem)
						//}else{
						//	//repeatFP[fpTimeMem] *=10
						//	repeatFP[fpTimeMem] +=1
						//
						//}
					}
					//fingerprintsOrdering = append(fingerprintsOrdering, cluster...)

				}
			}
		}
		if (!AtleastInOneCluster) {
			//glb.Error.Println("Not in cluster")
			fingerprintsOrdering = mainFingerprintsOrdering
		}
	} else {
		fingerprintsOrdering = mainFingerprintsOrdering
	}

	/*	if (curFingerprint.Timestamp == int64(1538064063095)){
			glb.Error.Println(fingerprintsOrdering)
		}
	*/
	FP2AFactor := make(map[string]float64)
	//hyperParams.GraphFactors = []float64{10,10,3,2,1}
	maxHopLevel := len(hyperParams.GraphFactors) - 2 // last item is minAdjacencyFactor

	adjacencyFactors := hyperParams.GraphFactors
	minAdjacencyFactor := hyperParams.GraphFactors[maxHopLevel+1] // assigning zero make errors in other functions
	//if minAdjacencyFactor != 0 {
	//	for _,fpTime := range fingerprintsOrdering{
	//		FP2AFactor[fpTime] = minAdjacencyFactor
	//	}
	//}

	// History effect:
	//historyConsidered = false // Note:deleteit .

	// Idea from: Dynamic Subarea Method in "A Hybrid Indoor Positioning Algorithm based on WiFi Fingerprinting and Pedestrian Dead Reckoning""
	if historyConsidered {
		baseLoc := "" // according to last location or pdr location, filter far fingerprints
		if (curFingerprint.Location != "" && glb.PDREnabledForDynamicSubareaMethod) { // Current PDRLocation is available
			baseLoc = curFingerprint.Location
			// todo : we can use lastUserPos.Location even when PDRLocation is available too.
		} else {
			userPosHistory := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)
			if len(userPosHistory) != 0 {
				lastUserPos := userPosHistory[len(userPosHistory)-1]
				//glb.Error.Println(lastUserPos)
				// todo:use lastUserPos.loction instead of knnguess

				baseLoc = lastUserPos.Location // Current PDRLocation isn't  available, use last location estimated
				/*	coGp, coGpExistErr := gp.Get_CoGroup()
					if coGpExistErr == nil {
						coGpUsrHisotry := coGp.Get_ResultData().Get_UserHistory(glb.TesterUsername)
						if len(coGpUsrHisotry) >0{
							glb.Error.Println(baseLoc)
							baseLoc = Fusion(lastUserPos, userPosHistory, coGpUsrHisotry)
							glb.Error.Println(baseLoc)

						}
					}*/

				/*				if strconv.FormatInt(curFingerprint.Timestamp, 10) == "1538071196747"{ //1538071209118
									glb.Error.Println(len(userPosHistory))
									glb.Error.Println(baseLoc)
								}*/
				//glb.Debug.Println(lastUserPos.KnnGuess)
				//glb.Debug.Println(lastUserPos.Location)
			}
		}
		//glb.Error.Println(baseLoc)

		if baseLoc != "" { // ignore when baseLoc is empty (for example there is no userhistory!)
			if knnConfig.GraphEnabled {
				var tempFingerprintOrdering []string
				graphMapPointer := gp.Get_ConfigData().Get_GroupGraph()

				if !graphMapPointer.IsEmpty() {
					baseNodeGraph := graphMapPointer.GetNearestNode(baseLoc)
					sliceOfHops := graphMapPointer.BFSTraverse(baseNodeGraph) // edit this function to return a nested slice with
					// nodes with corresponding hops

					for i, levelSliceOfHops := range sliceOfHops {
						var factor float64
						if (i <= maxHopLevel && adjacencyFactors[i] != 0) {
							factor = adjacencyFactors[i]
						} else if (minAdjacencyFactor != 0) { // last member of adjacencyFactors is minAdjacencyFactor
							factor = minAdjacencyFactor
						} else { // if minAdjacencyFactor is zero ignore remaining fingerprints(related to father graph nodes)
							break
						}
						//for i, levelSliceOfHops := range sliceOfHops {
						//var factor float64
						//if (i <= maxHopLevel) {
						//	factor = adjacencyFactors[i]
						//} else if (minAdjacencyFactor != 0) { // last member of adjacencyFactors is minAdjacencyFactor
						//	factor = minAdjacencyFactor
						//} else { // if minAdjacencyFactor is zero ignore remaining fingerprints(related to father graph nodes)
						//	break
						//}

						for _, node := range levelSliceOfHops {
							//hopFPs := append(hopFPs,node2FPs[node]...)
							hopFPs := node2FPs[node.Label]
							for _, fpTime := range hopFPs {
								FP2AFactor[fpTime] = factor
								tempFingerprintOrdering = append(tempFingerprintOrdering, fpTime)
							}
						}
					}

					fingerprintsOrdering = tempFingerprintOrdering
					//glb.Error.Println(FP2AFactor)
				}

			} else if knnConfig.DSAEnabled {
				var tempFingerprintOrdering []string
				baseLocX, baseLocY := glb.GetDotFromString(baseLoc)
				//glb.Error.Println()
				//maxMovement = float64(1)
				//hist := gp.Get_ResultData().Get_UserHistory(curFingerprint.Username)

				for _, fpTime := range fingerprintsOrdering {
					fp := fingerprintsInMemory[fpTime]
					fpLocX, fpLocY := glb.GetDotFromString(fp.Location)
					//glb.Error.Println()
					//glb.Error.Println(baseLoc)
					//glb.Error.Println(fp.Location)
					//glb.Error.Println(fp)
					//glb.Error.Println(fpLocX,",",fpLocY," - ",baseLocX,",",baseLocY)
					//glb.Error.Println(glb.CalcDist(fpLocX,fpLocY,baseLocX,baseLocY))
					if glb.CalcDist(fpLocX, fpLocY, baseLocX, baseLocY) < MaxMovement {
						//glb.Error.Println("OK addded")
						tempFingerprintOrdering = append(tempFingerprintOrdering, fpTime)
					}
				}
				if len(tempFingerprintOrdering) != 0 {
					//glb.Error.Println(len(fingerprintsOrdering))
					//glb.Error.Println(len(tempFingerprintOrdering))
					fingerprintsOrdering = tempFingerprintOrdering
				} else {
					glb.Error.Println("There is long distance between base location(last location or PDR current location) and current location")
				}
			} else if knnConfig.RPFEnabled {
				FP2AFactor = RPFs
			}

		}
	}

	//tempList := []string{}
	//tempList = append(tempList,fingerprintsOrdering...)
	sort.Sort(sort.StringSlice(fingerprintsOrdering))
	//
	//sum := int64(0)
	//for _,i := range tempList{
	//	num,_:=strconv.ParseInt(i, 10, 64)
	//	num = num % 100000
	//	sum += num
	//
	//}
	//glb.Debug.Println(sum/int64(len(tempList)))

	//glb.Debug.Println(len(fingerprintsOrdering))
	//glb.Debug.Println(len(mainFingerprintsOrdering))

	// Get k from db
	//knnK, err := dbm.GetKnnKOverride(curFingerprint.Group)
	//if err != nil {
	//	knnK = glb.DefaultKnnK
	//	glb.Error.Println("Nums of AP must be greater than 3")
	//}

	knnK := hyperParams.K

	//knnK := dbm.GetSharedPrf(curFingerprint.Group).KnnK

	// calculating knn
	W := make(map[string]float64)

	//var wgKnn sync.WaitGroup

	numJobs := len(fingerprintsOrdering)
	runtime.GOMAXPROCS(glb.MaxParallelism())
	chanJobs := make(chan jobW, 1+numJobs)
	chanResults := make(chan resultW, 1+numJobs)
	if (distAlgo == "Euclidean") {
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcWeight(id, chanJobs, chanResults)
		}
	} else if (distAlgo == "Cosine") {
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcWeightCosine(id, chanJobs, chanResults)
		}
	} else if (distAlgo == "NewEuclidean") {
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcWeight1(id, chanJobs, chanResults)
		}
	} else if (distAlgo == "RedpinEuclidean") {
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcWeightRedpin(id, chanJobs, chanResults)
		}
	} else if (distAlgo == "CombinedCosEuclid") {
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcCombinedCosEuclidWeight(id, chanJobs, chanResults)
		}
	}

	NumofMinAPNum := 0
	//glb.Debug.Println(fingerprintsInMemory)
	for _, fpTime := range fingerprintsOrdering {
		fp := fingerprintsInMemory[fpTime]

		if (len(fp.WifiFingerprint) < glb.MinApNum) { // todo:
			numJobs -= 1
			continue
		} else {
			NumofMinAPNum++
		}
		//Debug.Println(fp.WifiFingerprint)
		mac2RssFP := getMac2Rss(fp.WifiFingerprint)
		mac2RssCur := getMac2Rss(curFingerprint.WifiFingerprint)

		chanJobs <- jobW{fpTime: fpTime,
			mac2RssCur: mac2RssCur,
			mac2RssFP:  mac2RssFP}

	}

	close(chanJobs)

	if (knnConfig.GraphEnabled || knnConfig.RPFEnabled) && len(FP2AFactor) != 0 { //FP2AFactor length is zero when user history is empty
		for i := 1; i <= numJobs; i++ {
			res := <-chanResults
			/*			if (res.weight*FP2AFactor[res.fpTime]*float64(repeatFP[res.fpTime]) == 0) {
							glb.Error.Println(FP2AFactor[res.fpTime])
							glb.Error.Println(FP2AFactor)
						}*/
			W[res.fpTime] = res.weight * FP2AFactor[res.fpTime] * float64(repeatFP[res.fpTime])
		}
	} else {
		for i := 1; i <= numJobs; i++ {
			res := <-chanResults
			W[res.fpTime] = res.weight * float64(repeatFP[res.fpTime])
		}
	}

	if distAlgo == "NewEuclidean" {
		W = ConvertDist2Wigth(W)
	}

	close(chanResults)

	//if curFingerprint.Timestamp==int64(1516794991872647445){
	//	var keys []string
	//	for k := range W {
	//		keys = append(keys, k)
	//	}
	//	sort.Sort(sort.StringSlice(keys))
	//	var vals []float64
	//	for _,key := range keys{
	//		vals = append(vals, W[key])
	//	}
	//	glb.Error.Println(keys)
	//	glb.Error.Println(vals)
	//}

	var currentX, currentY int64
	currentX = 0
	currentY = 0

	if NumofMinAPNum == 0 {
		glb.Error.Println("There is no fingerprint that its number of APs be more than ", glb.MinApNum, "MinApNum")
		return errors.New("NumofAP_lowerThan_MinApNum"), ",", nil
	}

	fingerprintSorted := glb.SortReverseDictByVal(W)

	ws := []float64{}
	for _, w := range W {
		//if(w>float64(0.2)){
		//glb.Error.Println(w)
		ws = append(ws, w)
		//}

	}

	stopNum := 0 //used instead of knnK
	countWs := glb.DuplicateCountFloat64(ws)
	uniqueWs := glb.UniqueListFloat64(ws)

	sort.Sort(sort.Reverse(sort.Float64Slice(uniqueWs)))

	// instead of using knnK to stop knn algorithm, because there are some dots with same weight ,
	//		stopNum is set to minimum number of weight (from high to low).
	for _, w := range uniqueWs {
		//if curFingerprint.Timestamp==int64(1516794991872647445) {
		//	glb.Debug.Println(w)
		//}
		stopNum += countWs[w]
		if stopNum >= knnK {
			break
		}
	}

	////fmt.Println(fingerprintSorted)
	//if curFingerprint.Timestamp==int64(1516794991872647445) {
	//	glb.Debug.Println(countWs)
	//	glb.Debug.Println(uniqueWs)
	//	glb.Debug.Println(stopNum)
	//	glb.Error.Println()
	//	glb.Error.Println(len(W))
	//	glb.Error.Println(len(fingerprintSorted))
	//	glb.Error.Println(fingerprintSorted)
	//}
	if knn_regression {
		sumW := float64(0)
		KNNList := make(map[string]float64)
		//var xHist []int64
		//var xHistequ []string
		//var xHistMap []string
		for K, fpTime := range fingerprintSorted {
			if (K < stopNum) {
				KNNList[fpTime] = W[fpTime]
				x_y := strings.Split(fingerprintsInMemory[fpTime].Location, ",")
				if !(len(x_y) == 2) {
					err := errors.New("Location names aren't in the format of x,y")
					return err, "", nil
				}
				locXstr := x_y[0]
				locYstr := x_y[1]
				locX, _ := strconv.ParseFloat(locXstr, 64)
				locY, _ := strconv.ParseFloat(locYstr, 64)
				locX = glb.Round(locX, 5)
				locY = glb.Round(locY, 5)
				//currentX = currentX + int(W[fpTime]*locX)
				//currentY = currentY + int(W[fpTime]*locY)

				//if curFingerprint.Timestamp==int64(1516794991872647445) {
				//	xHist = append(xHist,int64(W[fpTime]*locX))
				//	xHistequ = append(xHistequ,fmt.Sprint(W[fpTime],"*",locX))
				//	xHistMap = append(xHistMap,fmt.Sprint(W[fpTime],"*",locX,":",int64(W[fpTime]*locX),":",fpTime))

				//glb.Error.Println()
				//glb.Error.Println(W[fpTime], "*",locX)
				//glb.Error.Println(currentX)
				//glb.Error.Println(W[fpTime], " * ",locY)
				//glb.Error.Println(currentY)
				//glb.Error.Println(currentX,"::",currentY)
				//}
				curW := W[fpTime]
				//glb.Debug.Println(curW)
				//curW := W[fpTime]
				currentX = currentX + int64(curW*locX)
				currentY = currentY + int64(curW*locY)
				//Debug.Println(W[fpTime]*locX)
				sumW = sumW + curW
			} else {
				break;
			}
		}
		//if curFingerprint.Timestamp==int64(1516794991872647445) {
		//	glb.Error.Println(xHist)
		//	sort.Sort(sort.StringSlice(xHistequ))
		//	glb.Error.Println(xHistequ)
		//	glb.Error.Println(xHistMap)
		//}
		//sumW = glb.Round(sumW,5)
		//if curFingerprint.Timestamp==int64(1516794991872647445) {
		//	glb.Error.Println(sumW)
		//	glb.Error.Println(currentX,"::",currentY)
		//}
		//if show{
		//	glb.Error.Println(sumW)
		//	glb.Error.Println(curFingerprint)
		//}
		if sumW == float64(0) {
			return errors.New("NoValidFingerprints"), "", nil
		}

		//glb.Error.Println(float64(currentX) / sumW , ",",float64(currentY) / sumW)
		currentXint := int(float64(currentX) / sumW)
		currentYint := int(float64(currentY) / sumW)
		//glb.Debug.Println(curFingerprint.Location)
		//glb.Debug.Println(glb.IntToString(currentXint) + ".0," + glb.IntToString(currentYint)+".0")
		//Debug.Println(currentX)
		return nil, glb.IntToString(currentXint) + ".0," + glb.IntToString(currentYint) + ".0", KNNList
	} else {
		KNNList := make(map[string]float64)
		for K, fpTime := range fingerprintSorted {
			if (K < stopNum) {
				fpLoc := fingerprintsInMemory[fpTime].Location
				if _, ok := KNNList[fpLoc]; ok {
					KNNList[fpLoc] += W[fpTime]
				} else {
					KNNList[fpLoc] = W[fpTime]
				}
			} else {
				break;
			}
		}
		sortedKNNList := glb.SortReverseDictByVal(KNNList)
		//glb.Debug.Println(sortedKNNList[0])
		return nil, sortedKNNList[0], KNNList
	}
}

func IsBLEMac(mac string) bool {
	if mac[:3] == "BLE" {
		return true
	} else {
		return false
	}
}

func calcWeight(id int, jobs <-chan jobW, results chan<- resultW) {

	for job := range jobs {
		distance := float64(0)
		techFactor := float64(1)
		length := float64(0.000001)
		for curMac, curRssi := range job.mac2RssCur {
			/*if curRssi < -77 {
				continue
			}*/
			if IsBLEMac(curMac) { //BLE
				techFactor = BLEFactor
			}
			if fpRss, ok := job.mac2RssFP[curMac]; ok {
				distance = distance + math.Pow(float64(curRssi-fpRss)*techFactor, minkowskyQ)
				length++
				//curDist := math.Pow(10.0,float64(curRssi)*0.05)
				//fpDist := math.Pow(10.0,float64(fpRss)*0.05)
				//distance = distance + math.Pow(curDist-fpDist, minkowskyQ)
			} else if glb.StringInSlice(curMac, uniqueMacs) {
				distance = distance + math.Pow(MaxEuclideanRssDist*techFactor, minkowskyQ)
				length++
				//distance = distance + 9
				//distance = distance + math.Pow(math.Pow(10.0,float64(-30)*0.05)-math.Pow(math.E,float64(-90)*0.05), minkowskyQ)
			}
		}
		//distance = distance / float64(len(job.mac2RssCur))
		distance = distance / float64(length)
		//if(distance==float64(0)){
		//	glb.Error.Println("###@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		//}
		precision := 10
		distance = glb.Round(math.Pow(distance, float64(1.0)/minkowskyQ), precision)
		if distance == float64(0) {
			//glb.Error.Println("Distance zero")
			//glb.Error.Println(job.mac2RssCur)
			//glb.Error.Println(job.mac2RssFP)
			distance = math.Pow(10, -1*float64(precision))
			//distance = maxDist
		}
		weight := glb.Round(float64(1.0)/(float64(1.0)+distance), 5)

		//glb.Debug.Println("distance: ",distance)
		//glb.Debug.Println("weight: ",weight)
		results <- resultW{fpTime: job.fpTime,
			weight: weight * 100}
	}

}

func calcWeightRedpin(id int, jobs <-chan jobW, results chan<- resultW) {
	alpha := float64(1)
	beta := float64(0.4)
	gama := float64(0.2)

	for job := range jobs {
		length := float64(0.000001)

		existanceSum := float64(0)
		xorExistanceSum := float64(0)
		curMacList := []string{}
		euclidDist := float64(0)

		for curMac, curRssi := range job.mac2RssCur {
			if fpRss, ok := job.mac2RssFP[curMac]; ok {
				existanceSum += 1
				euclidDist = euclidDist + math.Pow(float64(curRssi-fpRss), minkowskyQ)
				length++
			} else if glb.StringInSlice(curMac, uniqueMacs) {
				xorExistanceSum++;
			}
			curMacList = append(curMacList, curMac)
		}

		for fpMac, _ := range job.mac2RssFP { // no need to check in uniqueMacs beacuse uniqueMacs created from these fps
			if !glb.StringInSlice(fpMac, curMacList) {
				xorExistanceSum++;
			}
		}

		euclidDist = euclidDist / length

		weight := alpha*existanceSum - beta*xorExistanceSum - gama*euclidDist

		results <- resultW{fpTime: job.fpTime,
			weight: weight * 100}
	}

}

func calcWeight1(id int, jobs <-chan jobW, results chan<- resultW) {

	for job := range jobs {
		distance := float64(0)
		length := float64(0.000001)

		for curMac, curRssi := range job.mac2RssCur {
			if fpRss, ok := job.mac2RssFP[curMac]; ok {
				distance = distance + math.Pow(float64(curRssi-fpRss), minkowskyQ)
				length++
				//curDist := math.Pow(10.0,float64(curRssi)*0.05)
				//fpDist := math.Pow(10.0,float64(fpRss)*0.05)
				//distance = distance + math.Pow(curDist-fpDist, minkowskyQ)
			} else if glb.StringInSlice(curMac, uniqueMacs) {
				distance = distance + math.Pow(MaxEuclideanRssDist, minkowskyQ)
				length++
				//distance = distance + 9
				//distance = distance + math.Pow(math.Pow(10.0,float64(-30)*0.05)-math.Pow(math.E,float64(-90)*0.05), minkowskyQ)
			}
		}
		distance = distance / float64(len(job.mac2RssCur))
		//if(distance==float64(0)){
		//	glb.Error.Println("###@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		//}
		precision := 10
		distance = glb.Round(math.Pow(distance, float64(1.0)/minkowskyQ), precision)
		if distance == float64(0) {
			//glb.Error.Println("Distance zero")
			//glb.Error.Println(job.mac2RssCur)
			//glb.Error.Println(job.mac2RssFP)
			distance = math.Pow(10, -1*float64(precision))
			//distance = maxDist
		}
		//weight := glb.Round(float64(1.0)/(float64(1.0)+distance), 5)
		weight := distance

		//glb.Debug.Println("distance: ",distance)
		//glb.Debug.Println("weight: ",weight)
		results <- resultW{fpTime: job.fpTime,
			weight: weight}
	}

}

func calcWeightCosine(id int, jobs <-chan jobW, results chan<- resultW) {
	for job := range jobs {
		//distance := float64(0)
		var a []float64
		var b []float64

		//var a1 []float64
		//var b1 []float64

		for curMac, curRssi := range job.mac2RssCur {
			if fpRss, ok := job.mac2RssFP[curMac]; ok {
				//distance = distance + math.Pow(float64(curRssi-fpRss), minkowskyQ)
				a = append(a, norm2zeroToOne(float64(curRssi)))
				b = append(b, norm2zeroToOne(float64(fpRss)))
				//a1 = append(a1,(float64(curRssi)))
				//b1 = append(b1,(float64(fpRss)))
			} else {
				//distance = distance + maxDist
				a = append(a, norm2zeroToOne(float64(curRssi)))
				b = append(b, norm2zeroToOne(float64(glb.MinRssi)))
				//a1 = append(a1,(float64(curRssi)))
				//b1 = append(b1,(float64(MinRssi)))
			}
		}

		//distance = math.Pow(distance, float64(1)/minkowskyQ)+ float64(0.0000001)
		//weight := float64(1) / distance

		//Debug.Println(a)
		//Debug.Println(a1)
		//Debug.Println(b)
		//Debug.Println(b1)
		weight, err := Cosine(a, b)
		//Debug.Println(weight)

		//
		//weight = (weight-0.9)*10
		//weight += float64(0.0000000000001)
		//weight = norm2zeroToOneWieght(weight)
		//savedCosine = append(savedCosine,weight)
		//Debug.Println(savedCosine)
		//Debug.Println(weight)

		if err != nil {
			glb.Error.Println(err)
		}
		//Debug.Println("weight: ",weight)
		results <- resultW{fpTime: job.fpTime,
			weight: weight}
	}
}

func calcCombinedCosEuclidWeight(id int, jobs <-chan jobW, results chan<- resultW) {

	for job := range jobs {
		distance := float64(0)
		techFactor := float64(1)
		length := float64(0.000001)

		var a []float64
		var b []float64

		for curMac, curRssi := range job.mac2RssCur {
			/*if curRssi < -77 {
				continue
			}*/
			if IsBLEMac(curMac) { //BLE
				techFactor = BLEFactor
			}
			if fpRss, ok := job.mac2RssFP[curMac]; ok {
				distance = distance + math.Pow(float64(curRssi-fpRss)*techFactor, minkowskyQ)
				a = append(a, norm2zeroToOne(float64(curRssi)))
				b = append(b, norm2zeroToOne(float64(fpRss)))
				length++
				//curDist := math.Pow(10.0,float64(curRssi)*0.05)
				//fpDist := math.Pow(10.0,float64(fpRss)*0.05)
				//distance = distance + math.Pow(curDist-fpDist, minkowskyQ)
			} else if glb.StringInSlice(curMac, uniqueMacs) {
				distance = distance + math.Pow(MaxEuclideanRssDist*techFactor, minkowskyQ)
				a = append(a, norm2zeroToOne(float64(curRssi)))
				b = append(b, norm2zeroToOne(float64(glb.MinRssi)))
				length++
				//distance = distance + 9
				//distance = distance + math.Pow(math.Pow(10.0,float64(-30)*0.05)-math.Pow(math.E,float64(-90)*0.05), minkowskyQ)
			}
		}
		//distance = distance / float64(len(job.mac2RssCur))
		distance = distance / float64(length)
		//if(distance==float64(0)){
		//	glb.Error.Println("###@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		//}
		precision := 10
		distance = glb.Round(math.Pow(distance, float64(1.0)/minkowskyQ), precision)
		if distance == float64(0) {
			//glb.Error.Println("Distance zero")
			//glb.Error.Println(job.mac2RssCur)
			//glb.Error.Println(job.mac2RssFP)
			distance = math.Pow(10, -1*float64(precision))
			//distance = maxDist
		}
		weightEuclid := glb.Round(float64(1.0)/(float64(1.0)+distance), 5)
		weightCosine, err := Cosine(a, b)
		if err != nil {
			glb.Error.Println(err)
		}
		weight := weightCosine / weightEuclid
		//glb.Debug.Println("distance: ",distance)
		//glb.Debug.Println("weight: ",weight)
		results <- resultW{fpTime: job.fpTime,
			weight: weight * 100}
	}

}


func getMac2Rss(routeList []parameters.Router) map[string]int {
	mac2Rss := make(map[string]int)
	for _, rt := range routeList {
		mac2Rss[rt.Mac] = rt.Rssi
	}
	return mac2Rss
}

// Add mac:rate to wifiFingerprints , top n's rss of WifiFingerprint add according to topRssList values (n = len(topRssList))
// if calLearnFp is true  mac:rate is created for other than tops and their rssi set to MinRssi
//func calcMacRate(fp Fingerprint, calLearnFp bool) Fingerprint{
//	//routes := fp.WifiFingerprint
//	macDict :=  make(map[string]float64)
//	routes := fp.WifiFingerprint
//
//	for _,rt := range fp.WifiFingerprint{
//		macDict[rt.Mac] = float64(rt.Rssi)
//	}
//	sortedMac := sortDictByVal(macDict)
//
//
//	for i,mac := range sortedMac{
//		if(i<len(topRssList)){ // add 3
//			routes = append(routes,Router{Mac:mac+":Rate",Rssi:topRssList[i]})
//		}else if(calLearnFp){
//			routes = append(routes,Router{Mac:mac+":Rate",Rssi:MinRssi})
//		}
//	}
//	fp.WifiFingerprint = routes
//	//Debug.Print(fp.WifiFingerprint)
//
//
//	return fp
//}

//func calcMacJustRate(fp Fingerprint, calLearnFp bool) Fingerprint{
//	//routes := fp.WifiFingerprint
//	macDict :=  make(map[string]float64)
//	var routes []Router
//
//	for _,rt := range fp.WifiFingerprint{
//		macDict[rt.Mac] = float64(rt.Rssi)
//	}
//	sortedMac := sortDictByVal(macDict)
//
//
//	for i,mac := range sortedMac{
//		//if(i<len(topRssList)){ // add 3
//			routes = append(routes,Router{Mac:mac+":Rate",Rssi:i})
//		//}else if(calLearnFp){
//		//	routes = append(routes,Router{Mac:mac+":Rate",Rssi:len(sortedMac)+2})
//		//}
//	}
//	fp.WifiFingerprint = routes
//	//Debug.Print(fp.WifiFingerprint)
//
//
//	return fp
//}

func Cosine(a []float64, b []float64) (cosine float64, err error) {
	count := 0
	length_a := len(a)
	length_b := len(b)
	if length_a > length_b {
		count = length_a
	} else {
		count = length_b
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= length_a {
			s2 += math.Pow(b[k], 2)
			continue
		}
		if k >= length_b {
			s1 += math.Pow(a[k], 2)
			continue
		}
		sumA += a[k] * b[k]
		s1 += math.Pow(a[k], 2)
		s2 += math.Pow(b[k], 2)
	}
	if s1 == 0 || s2 == 0 {
		return 0.0, errors.New("Vectors should not be null (all zeros)")
	}
	return sumA / (math.Sqrt(s1) * math.Sqrt(s2)), nil
}

func norm2zeroToOne(x float64) float64 {
	a := 1.0 / (minrssInNormal - maxrssInNormal)
	b := -1 * maxrssInNormal * a
	x = x*a + b
	return x
}

func ConvertDist2Wigth(distMap map[string]float64) map[string]float64 {
	maxDist := float64(-1)
	minDist := math.MaxFloat64
	newDistMap := make(map[string]float64)

	for _, dist := range distMap {
		if dist > maxDist {
			maxDist = dist
		}
		if dist < minDist {
			minDist = dist
		}
	}

	for fpTime, dist := range distMap {
		newDistMap[fpTime] = dist*-1 + maxDist
	}

	return newDistMap

}
