package algorithms

import (
	"math"
	"strings"
	"strconv"
	"runtime"
	"ParsinServer/glb"
	"errors"
	"ParsinServer/algorithms/parameters"
	"ParsinServer/dbm"
	"sort"
)



var knn_regression bool

var minkowskyQ float64

var maxDist float64


var maxrssInNormal,minrssInNormal float64

//var topRssList []int

var distAlgo string

var validKs []int=[]int{}
var validMinClusterRSSs []int=[]int{}

func init() {

	knn_regression = true
	minkowskyQ = 2
	maxDist = 50
	distAlgo = "Euclidean" // Euclidean, Cosine
	//topRssList = []int{-60,-79,-90}
	maxrssInNormal = -55.0
	minrssInNormal = float64(glb.MinRssi) - 5.0

	for i:=1;i<=30;i++{
		validKs = append(validKs,i)
	}

	for i:=-60;i>=-90;i--{
		validMinClusterRSSs = append(validMinClusterRSSs,i)
	}

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



func LearnKnn(md *dbm.MiddleDataStruct,rd dbm.RawDataStruct,hyperParameters []interface{}) (parameters.KnnFingerprints,error) {
	//Debug.Println(Cosine([]float64{1,2,3},[]float64{1,2,4}))
	//jsonFingerprint = calcMacRate(jsonFingerprint,false)

	//K := hyperParameters[0].(int)
	MinClusterRSS := hyperParameters[1].(int)
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
	clusters := make(map[string][]string)
	//var err error

	fingerprints := rd.Fingerprints
	fingerprintsOrdering := rd.FingerprintsOrdering

	//fingerprintsOrdering,fingerprints,err = dbm.GetLearnFingerPrints(groupName,true)
	//if err!=nil {
	//	return err
	//}

	for fpTime,fp := range fingerprints {
		for _,rt := range fp.WifiFingerprint{
			if (rt.Rssi >= MinClusterRSS){
				clusters[rt.Mac] = append(clusters[rt.Mac],fpTime)
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

	// Add to knnData in db

	var tempKnnFingerprints parameters.KnnFingerprints
	tempKnnFingerprints.FingerprintsInMemory = fingerprints
	tempKnnFingerprints.FingerprintsOrdering = fingerprintsOrdering
	tempKnnFingerprints.Clusters = clusters

	//dbm.GM.GetGroup(groupName).Get_AlgoData().Set_KnnFPs(tempKnnFingerprints)

	//err = dbm.SetKnnFingerprints(tempKnnFingerprints, groupName)
	//if err != nil {
	//	glb.Error.Println(err)
	//	return err
	//}
	//
	//// Set in cache
	//go dbm.SetKnnFPCache(groupName,tempKnnFingerprints)
	return tempKnnFingerprints,nil
}
func TrackKnn(gp *dbm.Group, jsonFingerprint parameters.Fingerprint) (error, string) {

	//rd := gp.Get_RawData()
	//md := gp.Get_MiddleData()
	tempKnnFingerprints := gp.Get_AlgoData().Get_KnnFPs()

	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var mainFingerprintsOrdering []string
	var fingerprintsOrdering []string
	clusters := make(map[string][]string)
	//
	//tempKnnFingerprints, ok := dbm.GetKnnFPCache(jsonFingerprint.Group)
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
	//	tempKnnFingerprints,err = dbm.GetKnnFingerprints(jsonFingerprint.Group)
	//	if err!=nil{
	//		glb.Error.Println(err)
	//	}
	//	fingerprintsInMemory = tempKnnFingerprints.FingerprintsInMemory
	//	mainFingerprintsOrdering = tempKnnFingerprints.FingerprintsOrdering
	//	clusters = tempKnnFingerprints.Clusters
	//}

	//if strconv.Itoa(int(jsonFingerprint.Timestamp)) == "1516796888995082812"{
	//	glb.Error.Println("!")
	//}
	//tempKnnFingerprints := dbm.GM.GetGroup(jsonFingerprint.Group).Get_AlgoData().Get_KnnFPs()
	fingerprintsInMemory = tempKnnFingerprints.FingerprintsInMemory
	mainFingerprintsOrdering = tempKnnFingerprints.FingerprintsOrdering
	clusters = tempKnnFingerprints.Clusters

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

	//if jsonFingerprint.Location=="-165.000000,-1295.000000"{
	//	for key,val := range fingerprintsInMemory{
	//		if val.Location == jsonFingerprint.Location{
	//			glb.Warning.Println(jsonFingerprint.Timestamp)
	//			glb.Warning.Println(key)
	//		}
	//	}
	//}
	//show := false
	//if jsonFingerprint.Location == ""{
	//		for _,fp := range jsonFingerprint.WifiFingerprint{
	//			if fp.Mac == "01:17:C5:97:1B:44" && fp.Rssi == -72{
	//				//glb.Debug.Println(jsonFingerprint)
	//				show = true
	//			}
	//		}
	//}

	// fingerprintOrdering Creation according to clusters and rss rates
	highRateRssExist := false
	for _,rt := range jsonFingerprint.WifiFingerprint{
		if(rt.Rssi>=tempKnnFingerprints.MinClusterRss){
			if cluster, ok := clusters[rt.Mac]; ok {
				highRateRssExist = true
				fingerprintsOrdering = append(fingerprintsOrdering,cluster...)
			}
		}
	}
	if (!highRateRssExist){
		fingerprintsOrdering = mainFingerprintsOrdering
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
	//knnK, err := dbm.GetKnnKOverride(jsonFingerprint.Group)
	//if err != nil {
	//	knnK = glb.DefaultKnnK
	//	glb.Error.Println("Nums of AP must be greater than 3")
	//}


	knnK := tempKnnFingerprints.K

	//knnK := dbm.GetSharedPrf(jsonFingerprint.Group).KnnK

	// calculating knn
	W := make(map[string]float64)


	//var wgKnn sync.WaitGroup

	numJobs := len(fingerprintsOrdering)
	runtime.GOMAXPROCS(glb.MaxParallelism())
	chanJobs := make(chan jobW, 1+numJobs)
	chanResults := make(chan resultW, 1+numJobs)
	if(distAlgo=="Euclidean"){
		for id := 1; id <= glb.MaxParallelism(); id++ {
			//wgKnn.Add(1)
			go calcWeight(id, chanJobs, chanResults)
		}
	}else if(distAlgo=="Cosine"){
		for id := 1; id <= glb.MaxParallelism(); id++ {
			go calcWeightCosine(id, chanJobs, chanResults)
		}
	}

	for _, fpTime := range fingerprintsOrdering {
		fp := fingerprintsInMemory[fpTime]

		if (len(fp.WifiFingerprint) < glb.MinApNum) { // todo:
			numJobs -= 1
			continue
		}
		//Debug.Println(fp.WifiFingerprint)
		mac2RssFP := getMac2Rss(fp.WifiFingerprint)
		mac2RssCur := getMac2Rss(jsonFingerprint.WifiFingerprint)

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
	var currentX, currentY int
	currentX = 0
	currentY = 0


	fingerprintSorted := glb.SortDictByVal(W)
	//fmt.Println(fingerprintSorted)

	if knn_regression {
		sumW := float64(0)
		for K, fpTime := range fingerprintSorted {
			if (K < knnK) {
				x_y := strings.Split(fingerprintsInMemory[fpTime].Location, ",")
				if len(x_y) < 2 {
					err := errors.New("Location names aren't in the format of x,y")
					return err, ""
				}
				locXstr := x_y[0]
				locYstr := x_y[1]
				locX, _ := strconv.ParseFloat(locXstr,64)
				locY, _ := strconv.ParseFloat(locYstr,64)

				currentX = currentX + int(W[fpTime]*locX)
				currentY = currentY + int(W[fpTime]*locY)
				//Debug.Println(W[fpTime]*locX)
				sumW = sumW + W[fpTime]
			} else {
				break;
			}
		}

		//if show{
		//	glb.Error.Println(sumW)
		//	glb.Error.Println(jsonFingerprint)
		//}
		//glb.Debug.Println(sumW)
		currentXint := int(float64(currentX) / sumW)
		currentYint := int(float64(currentY) / sumW)
		//glb.Debug.Println(floatToString(currentX) + "," + floatToString(currentY))
		//Debug.Println(currentX)
		return nil, glb.IntToString(currentXint) + ".000000," + glb.IntToString(currentYint)+".000000"
	} else {
		KNNList := make(map[string]float64)
		for K, fpTime := range fingerprintSorted {
			if (K < knnK) {
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
		return nil, sortedKNNList[0]
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
				distance = distance + maxDist
				//distance = distance + 9
				//distance = distance + math.Pow(math.Pow(10.0,float64(-30)*0.05)-math.Pow(math.E,float64(-90)*0.05), minkowskyQ)
			}
		}
		//if(distance==float64(0)){
		//	glb.Error.Println("###@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		//}
		distance = glb.ToFixed(math.Pow(distance, float64(1)/minkowskyQ),6)+ float64(0.00000000000000001)
		weight := glb.ToFixed(float64(1) / distance * 1000,6)

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
				a = append(a,norm2zeroToOne(float64(curRssi)))
				b = append(b,norm2zeroToOne(float64(fpRss)))
				//a1 = append(a1,(float64(curRssi)))
				//b1 = append(b1,(float64(fpRss)))
			} else {
				//distance = distance + maxDist
				a = append(a,norm2zeroToOne(float64(curRssi)))
				b = append(b,norm2zeroToOne(float64(glb.MinRssi)))
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
		weight,err := Cosine(a,b)
		//Debug.Println(weight)

		//
		//weight = (weight-0.9)*10
		//weight += float64(0.0000000000001)
		//weight = norm2zeroToOneWieght(weight)
		//savedCosine = append(savedCosine,weight)
		//Debug.Println(savedCosine)
		//Debug.Println(weight)

		if err!=nil{
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
	a := 1.0 / (minrssInNormal-maxrssInNormal)
	b := -1 * maxrssInNormal * a
	x = x*a + b
	return x
}
