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

