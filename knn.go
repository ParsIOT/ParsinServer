package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"path"
	"math"
	"sort"
	"strings"
	"strconv"
	"errors"
)

// Default K in KNN algorithm
var defaultKnnK int

var knn_regression bool
func init() {
	defaultKnnK = 60
	knn_regression = false
}

func calculateKnn(jsonFingerprint Fingerprint) (error, string) {
	knnK, err := getKnnKOverride(jsonFingerprint.Group)
	if err != nil {
		knnK = defaultKnnK
		Error.Println("KNN K Override is not set!")
	}

	fingerprintsInMemory := make(map[string]Fingerprint)
	var fingerprintsOrdering []string

	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, jsonFingerprint.Group+".db"), 0600, nil)
	if err != nil {
		return err, ""
	}

	err = db.View(func(tx *bolt.Tx) error {
		//gets the fingerprint bucket
		b := tx.Bucket([]byte("fingerprints"))
		if b == nil {
			return fmt.Errorf("No fingerprint bucket")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fingerprintsInMemory[string(k)] = loadFingerprint(v)
			//fingerprintsOrdering is an array of fingerprintsInMemory keys
			fingerprintsOrdering = append(fingerprintsOrdering, string(k))
		}
		return nil
	})
	db.Close()

	if err != nil {
		return err, ""
	}

	// calculating knn
	W := make(map[string]float64)

	for _, fpTime := range fingerprintsOrdering {
		fp := fingerprintsInMemory[fpTime]
		distance := float64(0)
		mac2RssFP := getMac2Rss(fp.WifiFingerprint)
		mac2RssCur := getMac2Rss(jsonFingerprint.WifiFingerprint)

		for curMac, curRssi := range mac2RssCur {
			if fpRss, ok := mac2RssFP[curMac]; ok {
				distance = distance + math.Pow(float64(curRssi-fpRss), 2)
			}
		}
		distance = math.Sqrt(distance)
		W[fpTime] = float64(1) / distance
	}
	var currentX, currentY float64
	currentX = 0
	currentY = 0

	fingerprintSorted := sortedW(W)
	fmt.Println(fingerprintSorted)

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
				locX, _ := strconv.ParseFloat(locXstr, 64)
				locY, _ := strconv.ParseFloat(locYstr, 64)
				currentX = currentX + W[fpTime]*locX
				currentY = currentY + W[fpTime]*locY
				sumW = sumW + W[fpTime]
			} else {
				break;
			}
		}

		currentX = currentX / sumW
		currentY = currentY / sumW
		return nil, FloatToString(currentX) + "," + FloatToString(currentY)
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
		sortedKNNList := sortedW(KNNList)
		return nil, sortedKNNList[0]
	}
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func getMac2Rss(routeList []Router) map[string]int {
	mac2Rss := make(map[string]int)
	for _, rt := range routeList {
		mac2Rss[rt.Mac] = rt.Rssi
	}
	return mac2Rss
}

func sortedW(W map[string]float64) []string {
	var fingerprintSorted []string
	reverseMap := map[float64][]string{}
	var valueList sort.Float64Slice
	for k, v := range W {
		reverseMap[v] = append(reverseMap[v], k)
	}
	for k := range reverseMap {
		valueList = append(valueList, k)
	}
	valueList.Sort()
	sort.Sort(sort.Reverse(valueList))

	for _, k := range valueList {
		for _, s := range reverseMap[k] {
			fingerprintSorted = append(fingerprintSorted, s)
		}
	}
	return fingerprintSorted
}
