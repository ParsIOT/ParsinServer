// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// fingerprint.go contains structures and functions for handling fingerprints.

package parameters

import (
	"strings"
	"ParsinServer/glb"
)

//Fingerprint is the prototypical glb.Information from the fingerprinting device
//IF you change Fingerprint, follow these steps to re-generate fingerprint_ffjson.go
//find ./ -name "*.go" -type f | xargs sed -i  's/package main/package main/g'
//Uncomment json.Marshal/Unmarshal functions
//$GOPATH/bin/ffjson fingerprint.go
//find ./ -name "*.go" -type f | xargs sed -i  's/package main/package main/g'
//Comment json.Marshal/Unmarshal functions
type Fingerprint struct {
	Group           string   `json:"group"`
	Username        string   `json:"username"`
	Location        string   `json:"location"`
	Timestamp       int64    `json:"timestamp"`
	WifiFingerprint []Router `json:"wifi-fingerprint"`
}

type BulkFingerprint struct {
	Fingerprints []Fingerprint `json:"fingerprints"`
}

// Router is the router glb.Information for each individual mac address
type Router struct {
	Mac  string `json:"mac"`
	Rssi int    `json:"rssi"`
}

var jsonExample = `{
	"group": "whatevergroup",
	"username": "iamauser",
	"location": null,
	"wififingerprint": [{
		"mac": "AA:AA:AA:AA:AA:AA",
		"rssi": -45
	}, {
		"mac": "BB:BB:BB:BB:BB:BB",
		"rssi": -55
	}]
}`





// compression 9 us -> 900 us
// Marsahal and compress a fingerprint
func DumpFingerprint(res Fingerprint) []byte {
	dumped, _ := res.MarshalJSON()
	//dumped, _ := json.Marshal(res)
	return glb.CompressByte(dumped)
}

// compression 30 us -> 600 us
//loadFingerprint returns a fingerprint from given jsonByte input
func LoadRawFingerprint(jsonByte []byte) Fingerprint {
	res := Fingerprint{}
	//json.Unmarshal(decompressByte(jsonByte), res)
	res.UnmarshalJSON(glb.DecompressByte(jsonByte))
	return res
}

//returns the filtered macs from macs.json file and remove the other macs from fingerprint
func FilterRawFingerprint(res *Fingerprint,filterMacs []string) {

	//glb.Warning.Println(res.Group)
	// end function if there is no macfilter set
	//glb.Debug.Println(res)
	//glb.Debug.Println(glb.RuntimeArgs.NeedToFilter[res.Group])

	ok2, ok1 := glb.RuntimeArgs.NeedToFilter[res.Group] //check need for filtering
	ok3, ok4 := glb.RuntimeArgs.NotNullFilterMap[res.Group] //check that filterMap is null

	if ok2 && ok1 && ok3 && ok4{
		//glb.Debug.Println("1")
		if _, ok := glb.RuntimeArgs.FilterMacsMap[res.Group]; !ok {
			//err, filterMacs := dbm.GetFilterMacDB(res.Group)
			//glb.Warning.Println(filterMacs)
			//if err != nil {
			//	return
			//}
			glb.RuntimeArgs.FilterMacsMap[res.Group] = filterMacs
			//Rglb.RuntimeArgs.NeedToFilter[res.Group] = false //ToDo: filtering in loadfingerprint that was called by scikit.go not working! So i comment this line !
		}

		filterMacs := glb.RuntimeArgs.FilterMacsMap[res.Group]
		//glb.Debug.Println(filterMacs)
		newFingerprint := make([]Router, len(res.WifiFingerprint))
		curNum := 0

		for i := range res.WifiFingerprint {
			for _, mac := range filterMacs {
				if res.WifiFingerprint[i].Mac == mac {
					//glb.Debug.Println("4")
					//Error.Println("filtered mac : ",res.WifiFingerprint[i].Mac)
					newFingerprint[curNum] = res.WifiFingerprint[i]

					//newFingerprint[curNum].Mac = newFingerprint[curNum].Mac[0:len(newFingerprint[curNum].Mac)-1] + "0"
					curNum++
				}
			}
		}
		//glb.Debug.Println(newFingerprint[0:curNum])
		res.WifiFingerprint = newFingerprint[0:curNum]
	}
}

// convert quality (0 to 100) to rss(-100 to -50) and delete the records that their mac are "00:00:00:00:00"
func CleanFingerprint(res *Fingerprint) {
	res.Group = strings.TrimSpace(strings.ToLower(res.Group))
	res.Location = strings.TrimSpace(strings.ToLower(res.Location))
	res.Username = strings.TrimSpace(strings.ToLower(res.Username))
	deleteIndex := -1
	for r := range res.WifiFingerprint {
		if res.WifiFingerprint[r].Rssi >= 0 { // https://stackoverflow.com/questions/15797920/how-to-convert-wifi-signal-strength-from-quality-percent-to-rssi-dbm
			res.WifiFingerprint[r].Rssi = int(res.WifiFingerprint[r].Rssi/2) - 100
		}
		if res.WifiFingerprint[r].Mac == "00:00:00:00:00:00" {
			deleteIndex = r
		}
	}
	// delete res.WifiFingerprint[deleteIndex]
	if deleteIndex > -1 {
		res.WifiFingerprint[deleteIndex] = res.WifiFingerprint[len(res.WifiFingerprint)-1]
		res.WifiFingerprint = res.WifiFingerprint[:len(res.WifiFingerprint)-1]
	}
}

