package algorithms

import (
	"math"
	"strings"
	"strconv"
	"runtime"
	"ParsinServer/glb"
	"errors"
	"ParsinServer/dbm/parameters"
	"ParsinServer/dbm"
	"sort"
)

var knn_regression bool
var minkowskyQ float64
var maxrssInNormal, minrssInNormal float64
//var topRssList []int
var distAlgo string

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
	distAlgo = "Euclidean" // Euclidean, Cosine
	//topRssList = []int{-60,-79,-90}
	maxrssInNormal = -55.0
	minrssInNormal = float64(glb.MinRssi) - 5.0
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

	/*	node2FPs := make(map[string][]string)
		for fpTime, fp := range fingerprints{
			gp := dbm.GM.GetGroup("arman_28_3_97_ble_1") // Note:
			graphMapPointer := gp.Get_AlgoData().Get_GroupGraph()
			nearNodeGraph := graphMapPointer.GetNearestNode(fp.Location)

			if tempNode2FPs, ok :=node2FPs[fpTime]; ok {
				node2FPs[nearNodeGraph.Label] = append(tempNode2FPs,fpTime)
			}else{
				node2FPs[nearNodeGraph.Label] = []string{fpTime}
			}
		}*/

	//// Cluster print
	//for key,val := range clusters{
	//	fmt.Println("mac: "+key+" ")
	//	for _,fp:=range val{
	//		fmt.Println(fingerprints[fp])
	//	}
	//	fmt.Println("---------------------------------")
	//}

	// Add to knnData in db

	var tempKnnFingerprints parameters.KnnFingerprints
	tempKnnFingerprints.FingerprintsInMemory = fingerprints
	tempKnnFingerprints.FingerprintsOrdering = fingerprintsOrdering
	tempKnnFingerprints.Clusters = clusters
	//tempKnnFingerprints.Node2FPs = node2FPs
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
func TrackKnn(gp *dbm.Group, curFingerprint parameters.Fingerprint, historyConsidered bool) (error, string, map[string]float64) {

	//rd := gp.Get_RawData()
	//md := gp.Get_MiddleData()

	tempKnnFingerprints := gp.Get_AlgoData().Get_KnnFPs()

	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var mainFingerprintsOrdering []string
	var fingerprintsOrdering []string
	clusters := make(map[string][]string)
	//node2FPs := make(map[string][]string)
	//
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

	// fingerprintOrdering Creation according to clusters and rss rates
	highRateRssExist := false // komeil: a variable to decide if it is needed to search all fingerprints instead of one or some clusters
	for _, rt := range curFingerprint.WifiFingerprint {
		if (rt.Rssi >= hyperParams.MinClusterRss) {
			if cluster, ok := clusters[rt.Mac]; ok {
				highRateRssExist = true
				fingerprintsOrdering = append(fingerprintsOrdering, cluster...)
			}
		}
	}
	if (!highRateRssExist) {
		fingerprintsOrdering = mainFingerprintsOrdering
	}

	// History effect:
	// Idea from: Dynamic Subarea Method in "A Hybrid Indoor Positioning Algorithm based on WiFi Fingerprinting and Pedestrian Dead Reckoning""
	var tempFingerprintOrdering []string
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
				//glb.Debug.Println(lastUserPos.KnnGuess)
				//glb.Debug.Println(lastUserPos.Location)
			}
		}
		//glb.Error.Println(baseLoc)


		//FP2A := make(map[string]float64)
		if baseLoc != "" { // ignore when baseLoc is empty (for example there is no userhistory!)
			if glb.GraphEnabled {
				//graphMapPointer := gp.Get_AlgoData().Get_GroupGraph()
				//baseNodeGraph := graphMapPointer.GetNearestNode(baseLoc)
				//sliceOfHops := graphMapPointer.BFSTraverse(baseNodeGraph) // edit this function to return a nested slice with
				//// nodes with correspondign hops
				//maxLevel := 3
				//As := []float64{10,5,2};
				//minA := float64(1);
				//for i,levelSliceOfHops := range sliceOfHops{
				//		for _,node := range levelSliceOfHops{
				//			//hopFPs := append(hopFPs,node2FPs[node]...)
				//			hopFPs := node2FPs[node.Label]
				//			for _,fp := range hopFPs{
				//				if (i > maxLevel){
				//					FP2A[fp] = minA
				//				}else{
				//					FP2A[fp] = As[i]
				//				}
				//			}
				//		}
				//
				//
				//}



			} else {
				baseLocX, baseLocY := glb.GetDotFromString(baseLoc)
				maxMovement := dbm.GetSharedPrf(gp.Get_Name()).MaxMovement
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
					//glb.Error.Println(glb.CalcDist(fpLocX,fpLocY,baseLocX,baseLocY))
					if glb.CalcDist(fpLocX, fpLocY, baseLocX, baseLocY) < maxMovement {
						//glb.Error.Println("OK addded")
						tempFingerprintOrdering = append(tempFingerprintOrdering, fpTime)
					}
				}
				if len(tempFingerprintOrdering) != 0 {
					glb.Error.Println(len(fingerprintsOrdering))
					glb.Error.Println(len(tempFingerprintOrdering))
					fingerprintsOrdering = tempFingerprintOrdering
				} else {
					glb.Error.Println("There is long distance between base location(last location or PDR current location) and current location")
				}
			}

		}
	}

	//tempList := []string{}
	//tempList = append(tempList,fingerprintsOrdering...)
	//sort.Sort(sort.StringSlice(fingerprintsOrdering))
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
			//wgKnn.Add(1)
			go calcWeight(id, chanJobs, chanResults)
		}
	} else if (distAlgo == "Cosine") {
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcWeightCosine(id, chanJobs, chanResults)
		}
	}

	NumofMinAPNum := 0
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
			mac2RssFP: mac2RssFP}

	}

	close(chanJobs)

	for i := 1; i <= numJobs; i++ {
		res := <-chanResults
		W[res.fpTime] = res.weight
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

	fingerprintSorted := glb.SortDictByVal(W)

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
				//curW := W[fpTime]*FP2A[fpTime]
				curW := W[fpTime]
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
		sortedKNNList := glb.SortDictByVal(KNNList)
		//glb.Debug.Println(sortedKNNList[0])
		return nil, sortedKNNList[0], KNNList
	}
}

func calcWeight(id int, jobs <-chan jobW, results chan<- resultW) {

	for job := range jobs {
		distance := float64(0)
		for curMac, curRssi := range job.mac2RssCur {
			if fpRss, ok := job.mac2RssFP[curMac]; ok {
				distance = distance + math.Pow(float64(curRssi-fpRss), minkowskyQ)
				//curDist := math.Pow(10.0,float64(curRssi)*0.05)
				//fpDist := math.Pow(10.0,float64(fpRss)*0.05)
				//distance = distance + math.Pow(curDist-fpDist, minkowskyQ)
			} else {
				distance = distance + math.Pow(float64(glb.MaxEuclideanRssVectorDist), minkowskyQ)
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
		weight := glb.Round(float64(1.0)/(float64(1.0)+distance), 5)

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
