// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// api.go handles functions that return JSON responses.

package routes

import (
	"ParsinServer/algorithms"
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func PreLoadSettingsDecorator(routeFunc func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		group := c.DefaultQuery("group", "none")
		if group != "none" {
			dbm.GetSharedPrf(group)
		}
		routeFunc(c)
	}
}

func PreLoadSettings(c *gin.Context) {
	//glb.Debug.Println("PreloadSettings")
	group1 := c.Param("group")
	group2 := c.DefaultQuery("group", "none")
	groupExists := false
	//glb.Debug.Println(c)
	if len(group1) != 0 {
		glb.Debug.Println(group1)
		//glb.Debug.Println(dbm.GetSharedPrf(group1))
		groupExists = dbm.GroupExists(group1)
		if groupExists {
			dbm.GetSharedPrf(group1)
		} else {
			glb.Error.Println("Group haven't loaded")
			//c.JSON(http.StatusOK, gin.H{
			//	"message":   fmt.Sprintf("There is no group with this group name: ",group1),
			//	"success":   false})
			c.Redirect(302, "/change-db?error=groupNotExists")
		}

	} else if group2 != "none" {
		//glb.Debug.Println(group2)
		//glb.Debug.Println(dbm.GetSharedPrf(group2))
		groupExists = dbm.GroupExists(group2)
		if groupExists {
			dbm.GetSharedPrf(group2)
		} else {
			glb.Error.Println("Group isn't Exists")
			//c.JSON(http.StatusOK, gin.H{
			//	"message":   fmt.Sprintf("There is no group with this group name: ",group2),
			//	"success":   false})
			c.Redirect(302, "/change-db?error=groupNotExists")
		}

	} else {
		glb.Error.Println("Group name not mentioned in url")
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("group name must be mentioned in url(e.g.: /groupName (url param) or ?group=groupName (GET param))"),
			"success": false})
	}
	// Todo: add real "none" state
}

//returns uptime, starttime, number of cpu cores
func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"uptime": time.Since(startTime).Seconds(), "registered": startTime.String(), "status": "standard", "num_cores": runtime.NumCPU(), "success": true})
}

// parameters.UserPositionJSON stores the a users time, location and bayes after calculatePosterior()

// Gets location list:
// Example:
// {"locations":{
//		"p1":{"accuracy":76,"count":13},
//		"p2":{"accuracy":33,"count":12}
// },
// "message":"Found 2 unique locations in group arman4",
// "success":true}
// GET parameters: group
func GetLocationList(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	if groupName == "none" {
		c.JSON(http.StatusOK, gin.H{"message": "You need to specify group", "success": false})
		return
	}
	if !dbm.GroupExists(groupName) {
		glb.Error.Println("Group doesn't exist")
		c.JSON(http.StatusOK, gin.H{"message": "Group doesn't exist", "success": false})
		return
	}
	//ps, _ := dbm.OpenParameters(group)
	gp := dbm.GM.GetGroup(groupName)
	//md := gp.Get_MiddleData_Val()
	md := gp.Get_MiddleData()
	algoAccuracy := gp.Get_ResultData().Get_AlgoLocAccuracy()
	locationCount := make(map[string]map[string]int)

	fpData := gp.Get_RawData().Get_Fingerprints()
	uniqueLocations := []string{}
	//glb.Debug.Println(fpData)
	for _, fp := range fpData {
		if len(fp.WifiFingerprint) < glb.MinApNum {
			continue
		}
		if !glb.StringInSlice(fp.Location, uniqueLocations) {
			uniqueLocations = append(uniqueLocations, fp.Location)
		}
	}

	glb.Debug.Println(uniqueLocations)
	//md.Set_UniqueLocs(uniqueLocations)

	//for n := range uniqueLocations {
	//	for loc := range md.NetworkLocs[n] {
	for _, loc := range uniqueLocations {
		locationCount[loc] = make(map[string]int)
		//locationCount[loc]["count"] = gp.BayesResults[n].TotalLocations[loc]
		//locationCount[loc]["accuracy"] = gp.BayesResults[n].Accuracy[loc]
		locationCount[loc]["count"] = md.LocCount[loc]
		locationCount[loc]["accuracy"] = algoAccuracy["knn"][loc]
	}
	//}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Found %d unique locations in group %s", len(uniqueLocations), groupName),
		"locations": locationCount,
		"success":   true})
}

// An api that call getLastFingerprint()
// Example:
//sent as /track
//{
//	"group": "test_1",
//	"username": "hadi",
//	"location": "-10,-46",
//	"time": 1502544850139171556,
//	"wifi-fingerprint": [
//	{
//		"mac": "FA:CF:CB:5D:0E:B0",
//		"rssi": -82
//	},{
//		"mac": "F0:AB:CE:31:10:B0",
//		"rssi": -83
//	}]
//}
// GET parameters: group, user
func GetLastFingerprint(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	groupName := c.DefaultQuery("group", "none")
	groupName = strings.ToLower(groupName)
	user := c.DefaultQuery("user", "none")
	if groupName != "none" {
		if !dbm.GroupExists(groupName) {
			glb.Error.Println("Group doesn't exist")
			c.JSON(http.StatusOK, gin.H{"message": "Group doesn't exist", "success": false})
			return
		}
		if user == "none" {
			c.JSON(http.StatusOK, gin.H{"message": "You need to specify user", "success": false})
			return
		}
		glb.Debug.Println(groupName)
		c.String(http.StatusOK, dbm.LastFingerprint(groupName, user))
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "You need to specify groupName", "success": false})
	}
}

//Returns n of the last location estimations that were stored in fingerprints-track bucket in db
func GetHistoricalUserPositions(groupName string, user string, n int) []parameters.UserPositionJSON {

	//return userJSONs
	gp := dbm.GM.GetGroup(groupName)
	tempUserPositions := gp.Get_ResultData().Get_UserResults(user)
	var userPositions []parameters.UserPositionJSON

	// Get n last userPositions
	for i := len(tempUserPositions) - 1; len(tempUserPositions)-n <= i && i >= 0; i-- {
		//glb.Debug.Println(tempUserPositions[i].Fingerprint)
		userPositions = append(userPositions, tempUserPositions[i])
	}

	return userPositions
}

/*//Returns svm, rf, baysian estimations of the track fingerprints that belong to a group
func GetCurrentPositionOfAllUsers(groupName string) map[string]parameters.UserPositionJSON {
	//groupName = strings.ToLower(groupName)
	userPositions := make(map[string]parameters.UserPositionJSON)
	userFingerprints := make(map[string]parameters.Fingerprint)
	var err error
	userPositions, userFingerprints, err = dbm.TrackFingerprintsEmptyPosition(groupName)
	if (err != nil) {
		return userPositions
	}

	for user := range userPositions {
		//bayesGuess, bayesData := bayes.CalculatePosterior(userFingerprints[user], nil)
		foo := userPositions[user]
		//foo.BayesGuess = bayesGuess
		//foo.BayesData = bayesData
		// Process SVM if needed
		//if glb.RuntimeArgs.Svm {
		//	foo.SvmGuess, foo.SvmData = algorithms.SvmClassify(userFingerprints[user])
		//}
		//if glb.RuntimeArgs.Scikit {
		//	foo.ScikitData = algorithms.ScikitClassify(groupName, userFingerprints[user])
		//}
		gp := dbm.GM.GetGroup(groupName)
		_, foo.KnnGuess, foo.KnnData = algorithms.TrackKnn(gp, userFingerprints[user], false)
		go dbm.SetUserPositionCache(groupName+user, foo)
		userPositions[user] = foo
	}

	return userPositions
}*/

// Is like getHistoricalUserPositions but only returns the last location estimation
func GetCurrentPositionOfUser(groupName string, user string) parameters.UserPositionJSON {

	//val, ok := dbm.GetUserPositionCache(groupName + user)
	//if ok {
	//	return val
	//}
	//var userJSON parameters.UserPositionJSON
	//var userFingerprint parameters.Fingerprint
	//var err error
	//userJSON, userFingerprint, err = dbm.TrackFingeprintEmptyPosition(user, groupName)
	//if (err != nil) {
	//	return userJSON
	//}

	//bayesGuess, bayesData := bayes.CalculatePosterior(userFingerprint,nil)
	//userJSON.BayesGuess = bayesGuess
	//userJSON.BayesData = bayesData
	//// Process SVM if needed
	//if glb.RuntimeArgs.Svm {
	//	userJSON.SvmGuess, userJSON.SvmData = algorithms.SvmClassify(userFingerprint)
	//}
	//if glb.RuntimeArgs.Scikit {
	//	userJSON.ScikitData = algorithms.ScikitClassify(groupName, userFingerprint)
	//}
	gp := dbm.GM.GetGroup(groupName)
	var lastUserPos parameters.UserPositionJSON
	userPositions := gp.Get_ResultData().Get_UserResults(user)
	if len(userPositions) > 0 {

		lastUserPos = userPositions[len(userPositions)-1]

		//if userPositions[0].Fingerprint.Username != "04a31697d9c8" {
		//	lastUserPos = userPositions[len(userPositions)-1]
		//} else {
		//	glb.Error.Println("04a31697d9c8")
		//	for i := range userPositions {
		//		usrPos := userPositions[len(userPositions)-1-i]
		//		if len(usrPos.Fingerprint.WifiFingerprint) >= glb.MinApNum {
		//			return usrPos
		//		}
		//	}
		//}

	}

	//_, userJSON.KnnGuess, userJSON.KnnData = algorithms.TrackKnn(gp, userFingerprint, false)

	//_, userJSON.KnnGuess = calculateKnn(userFingerprint)
	//go dbm.SetUserPositionCache(groupName+user, userJSON)
	return lastUserPos
}

// calls optimizePriorsThreaded(),calculateSVM() and rfLearn()
// GET parameters: group
func Calculate(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	justCalcTestValidTracks := strings.ToLower(c.DefaultQuery("justCalcTestValidTracks", "none"))

	if groupName != "none" {
		if !dbm.GroupExists(groupName) {
			glb.Error.Println("Group doesn't exist")
			c.JSON(http.StatusOK, gin.H{"message": "Group doesn't exist", "success": false})
			return
		}

		calculateLearnData(groupName, justCalcTestValidTracks)

		go dbm.ResetCache("userPositionCache")
		go dbm.SetLearningCache(groupName, false)
		c.JSON(http.StatusOK, gin.H{"message": "Parameters optimized.", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func calculateLearnData(groupName string, justCalcGraphFactors string) {

	if justCalcGraphFactors != "true" {
		algorithms.CalculateLearn(groupName)
	} else {
		glb.Debug.Println("Just Calculating Graph Factors")
	}

	// test-valid hyper parameters selecting
	gp := dbm.GM.GetGroup(groupName)
	rsd := gp.Get_ResultData()
	knnConfig := gp.Get_ConfigData().Get_KnnConfig()
	testValidTracks := rsd.Get_TestValidTracks()
	if len(testValidTracks) != 0 {

		// Find true locations
		gp := dbm.GM.GetGroup(groupName)
		rsd := gp.Get_ResultData()
		_, _, testValidTracksRes := dbm.CalculateTestErrorAndRelocateTestValid(groupName, testValidTracks)
		if len(testValidTracksRes) != 0 { // if truelocations doesn't match avoid reseting testvalidtracks(testValidTracksRes is empty!)
			rsd.Set_TestValidTracks(testValidTracksRes)
		} else if len(testValidTracksRes) != len(testValidTracks) {
			glb.Error.Println("testValidTracksRes length and testValidTracks length doesn't equal")
		} else {
			glb.Error.Println("Empty testValidTracksRes(truelocations and testvalids timestamp don't match.)")
		}

		// calculate beset graphfactors
		algorithms.CalculateByTestValidTracks(groupName)
	} else if !knnConfig.GraphEnabled {
		rsd.Set_AlgoAccuracy("knn_testvalid_graph", 0)
	} else if !knnConfig.DSAEnabled {
		rsd.Set_AlgoAccuracy("knn_testvalid_dsa", 0)
	}
}

// An api that calls getHistoricalUserPositions() & getCurrentPositionOfUser()
// Returns location of a user, user list or users of a group
// GET parameters: group, user, users, n
func GetUserLocations(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := c.DefaultQuery("group", "none")
	groupName = strings.ToLower(groupName)
	userQuery := c.DefaultQuery("user", "none")
	usersQuery := c.DefaultQuery("users", "none")
	nQuery := c.DefaultQuery("n", "none")
	groupName = strings.ToLower(groupName)
	userQuery = strings.ToLower(userQuery)
	if groupName != "none" {
		if !dbm.GroupExists(groupName) {
			glb.Error.Println("Group doesn't exist")
			c.JSON(http.StatusOK, gin.H{"message": "Group doesn't exist", "success": false})
			return
		}
		people := make(map[string][]parameters.UserPositionJSON)
		users := strings.Split(strings.ToLower(usersQuery), ",")
		if users[0] == "none" {
			users = []string{userQuery}
		}
		if users[0] == "none" {
			//users = dbm.GetUsers(groupName)
			users = dbm.GetRecentUsers(groupName)
			glb.Debug.Println("Users:", users)
		}
		for _, user := range users {
			user = strings.ToLower(user) // todo: is it necessary? Does it conflict with learning data?
			if _, ok := people[user]; !ok {
				people[user] = []parameters.UserPositionJSON{}
			}
			if nQuery != "none" {
				number, _ := strconv.ParseInt(nQuery, 10, 0)
				glb.Debug.Println("Getting history for " + user)
				people[user] = append(people[user], GetHistoricalUserPositions(groupName, user, int(number))...)
			} else {
				people[user] = append(people[user], GetCurrentPositionOfUser(groupName, user))
			}
		}

		// Add fp wifi data to results
		//if glb.RuntimeArgs.Debug {
		//	fpData := dbm.GM.GetGroup(groupName).Get_RawData().Get_Fingerprints()
		//	tempKnnData := make(map[string]float64)
		//	for user,userposs := range people{
		//		for i,userpos := range userposs{
		//			for fpTime,val := range userpos.KnnData{
		//				out, err := json.Marshal(fpData[fpTime].WifiFingerprint)
		//				if err != nil {
		//					panic (err)
		//				}
		//				//tempKnnData[fpTime]=val
		//				tempKnnData[fpTime+" : "+string(out)]=val
		//			}
		//			//userpos.KnnData = tempKnnData
		//			//glb.Debug.Println(tempKnnData)
		//			people[user][i].KnnData = tempKnnData
		//		}
		//	}
		//}

		if glb.RuntimeArgs.Debug {
			fpData := dbm.GM.GetGroup(groupName).Get_RawData().Get_Fingerprints()
			tempKnnData := make(map[string]float64)
			for user, userposs := range people {
				for i, userpos := range userposs {
					for fpTime, val := range userpos.KnnData {
						tempKnnData[fpTime+" - "+fpData[fpTime].Location] = val
					}
					//userpos.KnnData = tempKnnData
					//glb.Debug.Println(tempKnnData)
					people[user][i].KnnData = tempKnnData
				}
			}
		}

		message := "Correctly found locations."
		if len(people) == 0 {
			message = "No users found for username " + strings.Join(users, " or ")
			people = nil
		}
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true, "users": people})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Get test-valid tracks
// GET parameters: group
func GetTestValidTracks(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		testValidTracks := dbm.GM.GetGroup(groupName).Get_ResultData().Get_TestValidTracks()
		glb.Debug.Println(testValidTracks)
		c.JSON(http.StatusOK, gin.H{"success": true, "testvalidtracks": testValidTracks})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "group or user isn't given"})
	}
}

func RecalculateTestvalidTrackFingerprint(mainGroupName string, repredict bool) []parameters.TestValidTrack {
	gp := dbm.GM.GetGroup(mainGroupName)
	coGp, _, coGpExistErr := gp.Get_CoGroup()

	testValidTracks := gp.Get_ResultData().Get_TestValidTracks()
	tempTestValidTracks := []parameters.TestValidTrack{}
	if len(testValidTracks) != 0 {
		if repredict {
			glb.Debug.Println("Repredicting test-valid tracks")
			allTestValidTracks := []parameters.TestValidTrack{}

			gp.Get_ResultData().Set_UserHistory(glb.TesterUsername, []parameters.UserPositionJSON{}) // clear last history
			if coGpExistErr == nil {
				coGp.Get_ResultData().Set_UserHistory(glb.TesterUsername, []parameters.UserPositionJSON{}) // clear last history
				coGpTestValidTracks := coGp.Get_ResultData().Get_TestValidTracks()

				gpC := 0
				coGpC := 0
				for gpC < len(testValidTracks) && coGpC < len(coGpTestValidTracks) {

					testValidTrack := testValidTracks[gpC]
					coGpTestValidTrack := coGpTestValidTracks[coGpC]
					if (testValidTrack.UserPosition.Time < coGpTestValidTrack.UserPosition.Time) {
						allTestValidTracks = append(allTestValidTracks, testValidTrack)
						gpC++
					} else {
						allTestValidTracks = append(allTestValidTracks, coGpTestValidTrack)
						coGpC++
					}
				}
				glb.Debug.Println(gpC)
				glb.Debug.Println(coGpC)
				glb.Debug.Println(len(testValidTracks))
				glb.Debug.Println(len(coGpTestValidTracks))
				if gpC == len(testValidTracks) {
					for ; coGpC < len(coGpTestValidTracks); coGpC++ {
						allTestValidTracks = append(allTestValidTracks, coGpTestValidTracks[coGpC])
					}
				} else if coGpC == len(coGpTestValidTracks) {
					for ; gpC < len(coGpTestValidTracks); gpC++ {
						allTestValidTracks = append(allTestValidTracks, testValidTracks[gpC])
					}
				}
			} else {
				copier.Copy(&allTestValidTracks, &testValidTracks)
			}

			// Repredict test-valid FPs
			for i, testValidTrack := range allTestValidTracks {
				//if i>10 {
				//	break
				//}
				fp := testValidTrack.UserPosition.Fingerprint
				//if (len(fp.WifiFingerprint) >= glb.MinApNum) {
				// Set the timestamp difference between testvalid fingerprints
				timeDiffWithLastFP := int64(0)
				if i != 0 {
					timeDiffWithLastFP = fp.Timestamp - allTestValidTracks[i-1].UserPosition.Time
				}
				newUserPositiong, err := algorithms.RecalculateTrackFingerprint(fp, timeDiffWithLastFP)
				if err == nil && fp.Group == mainGroupName { // ignore error-full ones and CoGroup testvalidtracks
					testValidTrack.UserPosition = newUserPositiong
					tempTestValidTracks = append(tempTestValidTracks, testValidTrack)
				}
			}

		} else {
			//gp := dbm.GM.GetGroup(mainGroupName)
			//testValidTracks := gp.Get_ResultData().Get_TestValidTracks()
			//fpData := gp.Get_RawData().Get_Fingerprints()
			//c.JSON(http.StatusOK, gin.H{"success": true, "testvalidtracks": testValidTracks, "fpdata": fpData})
			copier.Copy(&tempTestValidTracks, &testValidTracks)
		}
	}
	return tempTestValidTracks
}

func DelTestValidTracks(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		if !dbm.GroupExists(groupName) {
			glb.Error.Println("Group doesn't exist")
			c.JSON(http.StatusOK, gin.H{"message": "Group doesn't exist", "success": false})
			return
		}

		gp := dbm.GM.GetGroup(groupName)
		// delete testvalid tracks
		gp.Get_ResultData().Set_TestValidTracks([]parameters.TestValidTrack{})

		// delete truelocation(uwb) of testvalidtracks
		emptyTestValidTrueLocations := make(map[int64]string)
		gp.Get_RawData().Set_TestValidTrueLocations(emptyTestValidTrueLocations)

		c.JSON(http.StatusOK, gin.H{"success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "group or user isn't given"})
	}
}

// Get test-valid tracks details or recalculate track (repredict) location of these FPs
// GET parameters: group, repredict, calculate_err
// if repredict is true, all of testvalid fingerprints will be repredicted with the proposed algorithm(e.g. knn)
// If calculate_err is true, predicted locations and true locations will be compared and errDetails will be return
//	false calculate_err is used in test_valid_tracks_details and true one is used in test_valid_tracks_map
func GetTestValidTracksDetails(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-ContrgetTestValidTracksDetailsol-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	repredictStr := strings.ToLower(c.DefaultQuery("repredict", "none"))
	calculateErrStr := strings.ToLower(c.DefaultQuery("calculate_err", "none"))

	if groupName != "none" && repredictStr != "none" {
		if !dbm.GroupExists(groupName) {
			glb.Error.Println("Group doesn't exist")
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group doesn't exist"})
			return
		}

		fpData := dbm.GM.GetGroup(groupName).Get_RawData().Get_Fingerprints()
		repredict := false
		if repredictStr == "true" {
			repredict = true
		} else if repredictStr == "false" {
			repredict = false
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "repredict must be true or false"})
			return
		}

		tempTestValidTracks := RecalculateTestvalidTrackFingerprint(groupName, repredict)

		if len(tempTestValidTracks) != 0 {

			if (calculateErrStr != "none" && calculateErrStr == "true") { //Recalculate Error by true locations
				// testValidTracksRes is a temporary variable, don't save it in db
				err, errDetails, testValidTracksRes := dbm.CalculateTestErrorAndRelocateTestValid(groupName, tempTestValidTracks)
				if len(testValidTracksRes) != len(tempTestValidTracks) {
					glb.Error.Println("testValidTracksRes length and testValidTracks aren't equal")
				}
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				} else {
					/*				rsd.Set_TestValidTracks(testValidTracks)
									glb.Debug.Println(details)*/
					c.JSON(http.StatusOK, gin.H{"success": true, "errDetails": errDetails, "testvalidtracks": testValidTracksRes, "fpdata": fpData})
				}
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true, "testvalidtracks": tempTestValidTracks, "fpdata": fpData})
			}
		} else {
			glb.Error.Println("empty TestValidTracks")
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Empty TestValidTracks"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "group isn't given"})
	}
}

func GetTestErrorAlgoAccuracy(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		if !dbm.GroupExists(groupName) {
			glb.Error.Println("Group doesn't exist")
			c.JSON(http.StatusOK, gin.H{"message": "Group doesn't exist", "success": false})
			return
		}

		algosAccuracy := dbm.GM.GetGroup(groupName).Get_ResultData().Get_AlgoTestErrorAccuracy()
		c.JSON(http.StatusOK, gin.H{"success": true, "algosAccuracy": algosAccuracy})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "group isn't given"})
	}
}

// copies a DB
// GET parameters: from, to
func MigrateDatabase(c *gin.Context) {
	fromDB := strings.ToLower(c.DefaultQuery("from", "none"))
	toDB := strings.ToLower(c.DefaultQuery("to", "none"))
	glb.Debug.Printf("Migrating %s to %s.\n", fromDB, toDB)
	if !glb.Exists(path.Join(glb.RuntimeArgs.SourcePath, fromDB+".db")) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Can't migrate from " + fromDB + ", it does not exist."})
		return
	}
	if !glb.Exists(path.Join(glb.RuntimeArgs.SourcePath, toDB)) {
		glb.CopyFile(path.Join(glb.RuntimeArgs.SourcePath, fromDB+".db"), path.Join(glb.RuntimeArgs.SourcePath, toDB+".db"))
	} else {
		dbm.MigrateDatabaseDB(fromDB, toDB)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Successfully migrated " + fromDB + " to " + toDB})
}

// Deletes a db
// GET parameters: group
func DeleteDatabase(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	groupName := strings.TrimSpace(strings.ToLower(c.DefaultQuery("group", "none")))
	groupName = strings.ToLower(groupName)
	if glb.Exists(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")) {

		os.Remove(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db"))
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Successfully deleted " + groupName})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group does not exist"})
	}
}

// Calls setMixinOverride() and then calls optimizePriorsThreaded()
//// GET parameters: group, mixin
//func PutMixinOverride(c *gin.Context) {
//	c.Writer.Header().Set("Content-Type", "application/json")
//	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
//	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
//	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
//	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
//	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
//
//	group := strings.ToLower(c.DefaultQuery("group", "none"))
//	newMixin := c.DefaultQuery("mixin", "none")
//	if group != "none" {
//		newMixinFloat, err := strconv.ParseFloat(newMixin, 64)
//		if err == nil {
//			//err2 := dbm.SetMixinOverride(group, newMixinFloat)
//			err2 := dbm.SetSharedPrf(group,"Mixin", newMixinFloat)
//			if err2 == nil {
//				bayes.OptimizePriorsThreaded(strings.ToLower(group))
//				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding mixin for " + group + ", now set to " + newMixin})
//			} else {
//				c.JSON(http.StatusOK, gin.H{"success": false, "message": err2.Error()})
//			}
//		} else {
//			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
//		}
//	} else {
//		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
//	}
//}
//
//// Calls setCutoffOverride() and then calls optimizePriorsThreaded()
//// GET parameters: group, cutoff
//func PutCutoffOverride(c *gin.Context) {
//	c.Writer.Header().Set("Content-Type", "application/json")
//	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
//	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
//	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
//	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
//	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
//
//	group := strings.ToLower(c.DefaultQuery("group", "none"))
//	newCutoff := c.DefaultQuery("cutoff", "none")
//	glb.Debug.Println(group)
//	glb.Debug.Println(newCutoff)
//	if group != "none" {
//		newCutoffFloat, err := strconv.ParseFloat(newCutoff, 64)
//		if err == nil {
//			err2 := dbm.SetSharedPrf(group, "Cutoff", newCutoffFloat)
//			if err2 == nil {
//				bayes.OptimizePriorsThreaded(strings.ToLower(group))
//				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding cutoff for " + group + ", now set to " + newCutoff})
//			} else {
//				c.JSON(http.StatusOK, gin.H{"success": false, "message": err2.Error()})
//			}
//		} else {
//			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
//		}
//	} else {
//		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
//	}
//}
//
//// Calls setCutoffOverride() and then calls optimizePriorsThreaded()
//// GET parameters: group, cutoff
//func PutKnnK(c *gin.Context) {
//	c.Writer.Header().Set("Content-Type", "application/json")
//	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
//	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
//	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
//	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
//	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
//
//	group := strings.ToLower(c.DefaultQuery("group", "none"))
//	newK := c.DefaultQuery("knnK", "none")
//	glb.Debug.Println(group)
//	glb.Debug.Println(newK)
//	if group != "none" {
//		newKnnK, err := strconv.Atoi(newK)
//		if err == nil {
//			err2 := dbm.SetSharedPrf(group,"KnnK", newKnnK)
//			if err2 == nil {
//				//optimizePriorsThreaded(strings.ToLower(group))
//				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding KNN K for " + group + ", now set to " + newK})
//			} else {
//				c.JSON(http.StatusOK, gin.H{"success": false, "message": err2.Error()})
//			}
//		} else {
//			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
//		}
//	} else {
//		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
//	}
//}

func PutKnnKRange(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "none"))
	kRangeRawStr := c.DefaultQuery("range", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(kRangeRawStr)

	if group != "none" && kRangeRawStr != "none" {
		// convert string to int slice
		kRangeRawStr = strings.TrimSpace(kRangeRawStr)
		kRangeRawStr = kRangeRawStr[1:][:len(kRangeRawStr)-2]
		kRangeListStr := strings.Split(kRangeRawStr, ",")
		kRange := []int{}

		for _, numStr := range kRangeListStr {
			num, _ := strconv.Atoi(numStr)
			kRange = append(kRange, num)
		}

		// check kRange length
		if len(kRange) == 1 || len(kRange) == 2 {
			//validKs := glb.MakeRange(kRange[0],kRange[0])
			cd := dbm.GM.GetGroup(group).Get_ConfigData()
			knnConfig := cd.Get_KnnConfig()
			//err := dbm.SetSharedPrf(group, "KRange", kRange)
			knnConfig.KRange = kRange
			cd.Set_KnnConfig(knnConfig)
			//if err == nil {
			//	//optimizePriorsThreaded(strings.ToLower(group))
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding KNN K range for " + group + ", now set to " + kRangeRawStr})
			//} else {
			//	c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			//}
			//}else if( len(kRange) == 2){
			//	algorithms.ValidKs = glb.MakeRange(kRange[0],kRange[1])
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Knn K range length must be 2 at the maximum value "})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func PutKnnMinClusterRSSRange(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "none"))
	rssRangeRawStr := c.DefaultQuery("range", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(rssRangeRawStr)

	if group != "none" && rssRangeRawStr != "none" {
		// convert string to int slice
		rssRangeRawStr = strings.TrimSpace(rssRangeRawStr)
		rssRangeRawStr = rssRangeRawStr[1:][:len(rssRangeRawStr)-2]
		rssRangeListStr := strings.Split(rssRangeRawStr, ",")
		minCRssRange := []int{}

		for _, numStr := range rssRangeListStr {
			num, err := strconv.Atoi(numStr)
			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
			}
			minCRssRange = append(minCRssRange, num)
		}

		// check kRange length
		if len(minCRssRange) == 1 || len(minCRssRange) == 2 {
			//validKs := glb.MakeRange(kRange[0],kRange[0])
			cd := dbm.GM.GetGroup(group).Get_ConfigData()
			knnConfig := cd.Get_KnnConfig()
			//err := dbm.SetSharedPrf(group, "KnnMinCRssRange", minCRssRange)
			knnConfig.MinClusterRssRange = minCRssRange
			cd.Set_KnnConfig(knnConfig)

			//if err == nil {
			//optimizePriorsThreaded(strings.ToLower(group))
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding KNN K range for " + group + ", now set to " + rssRangeRawStr})
			//} else {
			//	c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			//}
			//}else if( len(kRange) == 2){
			//	algorithms.ValidKs = glb.MakeRange(kRange[0],kRange[1])
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Knn K range length must be 2 at the maximum value "})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

/*func PutMaxMovement(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "none"))
	MaxMovementStr := c.DefaultQuery("maxMovement", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(MaxMovementStr)

	if group != "none" && MaxMovementStr != "none" {
		MaxMovement, err := strconv.ParseFloat(MaxMovementStr, 64)
		if err == nil {
			if MaxMovement == float64(-1) {
				MaxMovement = glb.DefaultMaxMovement
			}
			err2 := dbm.SetSharedPrf(group, "MaxMovement", MaxMovement)
			if err2 == nil {
				//optimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding MaxMovement for " + group + ", now set to " + MaxMovementStr})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err2.Error()})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}*/

/*func PutMaxEuclideanRssDist(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	MaxEuclideanRssDistStr := c.DefaultQuery("maxEuclideanRssDist", "none")

	glb.Debug.Println(groupName)
	glb.Debug.Println(MaxEuclideanRssDistStr)

	if groupName != "none" && MaxEuclideanRssDistStr != "none" {
		MaxEuclideanRssDist, err := strconv.Atoi(MaxEuclideanRssDistStr)
		if err == nil {
			if MaxEuclideanRssDist == -1 {
				MaxEuclideanRssDist = glb.DefaultMaxEuclideanRssDistRange[0]
			}

			cd := dbm.GM.GetGroup(groupName).Get_ConfigData()
			knnParams := cd.Get_KnnConfig()
			knnParams.MaxEuclideanRssDist = MaxEuclideanRssDist
			cd.Set_KnnConfig(knnParams)

			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding MaxEuclideanRssDist for " + groupName + ", now set to " + MaxEuclideanRssDistStr})

		} else {
			glb.Error.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "group or maxEuclideanRssDist isn't given"})
	}
}
*/
func ChooseMap(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "none"))
	mapName := c.DefaultQuery("mapName", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(mapName)

	if group != "none" && mapName != "none" {
		//MaxMovement, err := strconv.ParseFloat(MaxMovementStr, 64)
		//if err == nil {
		//	if MaxMovement == float64(-1) {
		//		MaxMovement = glb.DefaultMaxMovement
		//	}
		mapNamesList := glb.ListMaps()
		MapWidth := mapNamesList[mapName][0]
		MapHeight := mapNamesList[mapName][1]
		MapDimensions := []int{MapWidth, MapHeight}
		glb.Debug.Println("***MapDimensions : ", MapDimensions)
		err2 := dbm.SetSharedPrf(group, "MapName", mapName)
		err3 := dbm.SetSharedPrf(group, "MapDimensions", MapDimensions)

		if err2 == nil && err3 == nil {

			//optimizePriorsThreaded(strings.ToLower(group))
			glb.Debug.Println(dbm.GetSharedPrf(group).MapName)
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding mapName for " + group + ", now set to " + mapName})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err2.Error()})
		}
	}
}

// Calls setCutoffOverride() and then calls optimizePriorsThreaded()
// GET parameters: group, cutoff
func PutMinRss(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "none"))
	minRss := c.DefaultQuery("minRss", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(minRss)
	if group != "none" && minRss != "none" {
		newMinRss, err := strconv.Atoi(minRss)
		if err == nil {
			err2 := dbm.SetSharedPrf(group, "MinRss", newMinRss)
			if err2 == nil {
				//optimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding Minimum RSS for " + group + ", now set to " + minRss})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err2.Error()})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Calls renameNetwork() and then calls optimizePriors()
// Done: replace optimizePriors() with optimizePriorsThreaded()
// GET parameters: group, oldname, newname
//func EditNetworkName(c *gin.Context) {
//	c.Writer.Header().Set("Content-Type", "application/json")
//	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
//	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
//	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
//	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
//	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
//
//	group := c.DefaultQuery("group", "none")
//	oldname := c.DefaultQuery("oldname", "none")
//	newname := c.DefaultQuery("newname", "none")
//	if group != "none" {
//		//glb.Debug.Println("Attempting renaming ", group, oldname, newname)
//		dbm.RenameNetwork(group, oldname, newname)
//		CalculateLearn(group)
//		//bayes.OptimizePriorsThreaded(group)
//		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Finished"})
//	} else {
//		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
//	}
//}

// Changes a location name in db(fingerprints and fingerprints-track buckets)
// GET parameters: group, location (the old name), newname
func EditLoc(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := c.DefaultQuery("group", "none")
	oldloc := strings.TrimSpace(c.DefaultQuery("oldloc", "none"))
	newloc := strings.TrimSpace(c.DefaultQuery("newloc", "none"))
	glb.Debug.Println(oldloc)
	glb.Debug.Println(newloc)

	if groupName != "none" && oldloc != "none" && newloc != "none" {
		numChanges := dbm.EditLocDB(oldloc, newloc, groupName)
		glb.Debug.Println("Changed location of " + strconv.Itoa(numChanges) + " fingerprints")
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))
		//algorithms.CalculateLearn(groupName)
		c.JSON(http.StatusOK, gin.H{"message": "Changed location of " + strconv.Itoa(numChanges) + " fingerprints", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func EditLocBaseDB(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := c.DefaultQuery("group", "none")
	oldloc := strings.TrimSpace(c.DefaultQuery("oldloc", "none"))
	newloc := strings.TrimSpace(c.DefaultQuery("newloc", "none"))
	if groupName != "none" && oldloc != "none" && newloc != "none" {
		numChanges := dbm.EditLocBaseDB(oldloc, newloc, groupName)
		dbm.EditLocDB(oldloc, newloc, groupName)
		glb.Debug.Println("Changed location of " + strconv.Itoa(numChanges) + " fingerprints")
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))
		//algorithms.CalculateLearn(groupName)
		c.JSON(http.StatusOK, gin.H{"message": "Changed location of " + strconv.Itoa(numChanges) + " fingerprints", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Changes a mac name in db(fingerprints and fingerprints-track buckets)
// GET parameters: group, oldmac, newmac
func EditMac(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := c.DefaultQuery("group", "none")
	oldmac := c.DefaultQuery("oldmac", "none")
	newmac := c.DefaultQuery("newmac", "none")
	forceEdit := c.DefaultQuery("forceEdit", "none")

	if groupName != "none" && oldmac != "none" && newmac != "none" {
		gp := dbm.GM.GetGroup(groupName)
		fpData := gp.Get_RawData().Get_Fingerprints()

		dbMacs := []string{}
		for _, fp := range fpData {
			for _, rt := range fp.WifiFingerprint {
				if !glb.StringInSlice(rt.Mac, dbMacs) {
					dbMacs = append(dbMacs, rt.Mac)
				}
			}
		}

		if glb.StringInSlice(newmac, dbMacs) && forceEdit != "true" {
			glb.Error.Println("There are some fingerprints that conatins the newmac already(edit newmac first)")
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "There are some fingerprints that conatins the newmac already(edit newmac first)"})
		} else {
			numChanges := dbm.EditMacDB(oldmac, newmac, groupName)
			glb.Debug.Println("Changed mac of " + strconv.Itoa(numChanges) + " fingerprints")
			//algorithms.CalculateLearn(groupName)
			//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))
			c.JSON(http.StatusOK, gin.H{"message": "Changed mac of " + strconv.Itoa(numChanges) + " fingerprints", "success": true})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Same to editLoc() but edits username instead of the location name
// GET paramets: group, user(the old username), newname
func EditUserName(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	user := strings.ToLower(c.DefaultQuery("user", "none"))
	newname := strings.ToLower(c.DefaultQuery("newname", "none"))
	if groupName != "none" && user != "none" && newname != "none" {
		numChanges := dbm.EditUserNameDB(user, newname, groupName)

		// reset the cache (cache.go)
		go dbm.ResetCache("usersCache")
		go dbm.ResetCache("userPositionCache")

		c.JSON(http.StatusOK, gin.H{"message": "Changed name of " + strconv.Itoa(numChanges) + " things", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Deletes the fingerprints associated to the location and then calls optimizePriorsThreaded()
// GET parameters: group, location
func DeleteLocation(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	//gp := dbm.GM.GetGroup(groupName)
	location := strings.ToLower(c.DefaultQuery("location", "none"))
	if groupName != "none" && location != "none" {
		numChanges := dbm.DeleteLocationDB(location, groupName)

		// todo: can't calculateLearn( there is problem with goroutine)
		//algorithms.CalculateLearn(groupName)
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))

		c.JSON(http.StatusOK, gin.H{"message": "Deleted " + strconv.Itoa(numChanges) + " locations", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func DeleteLocationBaseDB(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	//gp := dbm.GM.GetGroup(groupName)
	location := strings.ToLower(c.DefaultQuery("location", "none"))
	if groupName != "none" && location != "none" {
		numChangesBaseDB := dbm.DeleteLocationBaseDB(location, groupName)
		numChangesGpCache := dbm.DeleteLocationDB(location, groupName)
		if (numChangesBaseDB != numChangesGpCache) {
			glb.Error.Printf("number of deletation from (baseDB,groupCache) are not equal: (%d,%d)\n", numChangesBaseDB, numChangesGpCache)
		}
		// todo: can't calculateLearn( there is problem with goroutine)
		//algorithms.CalculateLearn(groupName)
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))

		c.JSON(http.StatusOK, gin.H{"message": "Deleted " + strconv.Itoa(numChangesBaseDB) + " locations", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func DeleteLocations(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	locationsQuery := strings.ToLower(c.DefaultQuery("names", "none"))
	if groupName != "none" && locationsQuery != "none" {
		locations := strings.Split(strings.ToLower(locationsQuery), ",")
		numChangesBaseDB := dbm.DeleteLocationsBaseDB(locations, groupName)
		numChangesGpCache := dbm.DeleteLocationsDB(locations, groupName)
		if (numChangesBaseDB != numChangesGpCache) {
			glb.Error.Printf("number of deletation from (baseDB,groupCache) are not equal: (%d,%d)\n", numChangesBaseDB, numChangesGpCache)
		}
		// todo: can't calculateLearn( there is problem with goroutine)
		algorithms.CalculateLearn(groupName)
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))
		c.JSON(http.StatusOK, gin.H{"message": "Deleted " + strconv.Itoa(numChangesBaseDB) + " locations", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Need to provide groupName and location list. DELETE /locations?groupName=X&names=Y,Z,W"})
	}
}

// Is like deleteLocation(), deletes a list of locations instead.
// GET parameters: group, names
func DeleteLocationsBaseDB(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	locationsQuery := strings.ToLower(c.DefaultQuery("names", "none"))
	if groupName != "none" && locationsQuery != "none" {
		locations := strings.Split(strings.ToLower(locationsQuery), ",")
		numChanges := dbm.DeleteLocationsDB(locations, groupName)
		// todo: can't calculateLearn( there is problem with goroutine)
		algorithms.CalculateLearn(groupName)
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))
		c.JSON(http.StatusOK, gin.H{"message": "Deleted " + strconv.Itoa(numChanges) + " locations", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Need to provide groupName and location list. DELETE /locations?groupName=X&names=Y,Z,W"})
	}
}

// Deletes a user from fingerprint-track(not fingerprints) then calls resetCache()
// GET parameters: group, user
func DeleteUser(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "none"))
	user := strings.ToLower(c.DefaultQuery("user", "none"))
	if group != "none" && user != "none" {
		numChanges := dbm.DeleteUser(user, group)
		// reset the cache (cache.go)
		go dbm.ResetCache("usersCache")
		go dbm.ResetCache("userPositionCache")

		c.JSON(http.StatusOK, gin.H{"message": "Deletes " + strconv.Itoa(numChanges) + " things " + " with user " + user, "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Set filterMacs
// POST parameters: filterMacs
func Setfiltermacs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	var filterMacs parameters.FilterMacs

	//x, _ := ioutil.ReadAll(c.Request.Body)
	//Warning.Println("%s", string(x))

	if glb.BindJSON(&filterMacs, c) == nil {
		if len(filterMacs.Macs) == 0 {
			//glb.RuntimeArgs.NeedToFilter[filterMacs.Group] = false
			//glb.RuntimeArgs.NotNullFilterList[filterMacs.Group] = false
			dbm.SetRuntimePrf(filterMacs.Group, "NeedToFilter", false)
			dbm.SetRuntimePrf(filterMacs.Group, "NotNullFilterList", false)
		} else {
			//glb.RuntimeArgs.NeedToFilter[filterMacs.Group] = true
			//glb.RuntimeArgs.NotNullFilterList[filterMacs.Group] = true
			dbm.SetRuntimePrf(filterMacs.Group, "NeedToFilter", true)
			dbm.SetRuntimePrf(filterMacs.Group, "NotNullFilterList", true)
		}

		//err := dbm.SetFilterMacDB(filterMacs.Group, filterMacs.Macs)
		err := dbm.SetSharedPrf(filterMacs.Group, "FilterMacsMap", filterMacs.Macs)
		if err == nil {
			//glb.RuntimeArgs.FilterMacsMap[filterMacs.Group] = filterMacs.Macs
			glb.Debug.Println("MacFilter set successfully ")
			if len(filterMacs.Macs) == 0 {
				c.JSON(http.StatusOK, gin.H{"message": "MacFilter Cleared.", "success": true})
			} else {
				c.JSON(http.StatusOK, gin.H{"message": "MacFilter set successfully", "success": true})
			}
		} else {
			glb.Warning.Println(err)
			c.JSON(http.StatusOK, gin.H{"message": "setFilterMacDB problem", "success": false})
		}
	} else {
		glb.Warning.Println("Can't bind json")
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Can't bind json"})
		//c.JSON(http.StatusOK, gin.H{"message": "Nums of the FilterMacs are zero", "success": false})
	}

}

// Get filterMacs
// Get parameters: group
func Getfiltermacs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	group := c.DefaultQuery("group", "none")
	var err error
	var FilterMacs []string
	if group != "none" {
		//err, FilterMacs = dbm.GetFilterMacDB(group)
		//glb.Debug.Println("filterMacs")
		FilterMacs = dbm.GetSharedPrf(group).FilterMacsMap
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "group field is null", "success": false})
	}

	if err == nil {
		glb.Debug.Println("FilterMacs: ", FilterMacs)
		c.JSON(http.StatusOK, gin.H{"message": FilterMacs, "success": true})
	} else {
		glb.Warning.Println(err)
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	}

}

// POST parameters: graph
func AddNodeToGraph(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	type st struct {
		NewVertexKey string `json:newVertexKey`
	}
	gp := dbm.GM.GetGroup(groupName)
	curGroupGraph := gp.Get_ConfigData().Get_GroupGraph()

	//glb.Debug.Println(curGroupGraph.GetNearestNode("998.0,1040.0").Label)
	//glb.Debug.Println(curGroupGraph.GetNearestNode("-252,-1223").Label)
	//glb.Debug.Println(curGroupGraph.GetNearestNode("-246,1534").Label)

	var tempSt st
	if groupName != "none" {
		//glb.Debug.Println(c.Request.GetBody())
		if err := c.ShouldBindJSON(&tempSt); err == nil {
			//glb.Debug.Println("newVertexLabel : %%%---------> ",tempSt.NewVertexKey)
			newVertexLabel := tempSt.NewVertexKey
			//glb.Debug.Println("newVertexLabel : ---------> ",newVertexLabel)
			curGroupGraph.AddNodeByLabel(newVertexLabel)
			//curGroupGraph.DeleteGraph()
			//glb.Debug.Println("graph after adding : ---------> ",curGroupGraph.GetGraphMap())
			cd := gp.Get_ConfigData()
			cd.Set_GroupGraph(curGroupGraph)
			//glb.Debug.Println("####### exited AddNodeByLabel in the bindJson ########")
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true})
			}
		} else {
			glb.Warning.Println("Can't bind json")
			glb.Error.Println(err)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Can't bind json, Error:" + err.Error()})
			//c.JSON(http.StatusOK, gin.H{"message": "Nums of the FilterMacs are zero", "success": false})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group not mentioned"})
	}
}

func SaveEdgesToDB(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	group := c.DefaultQuery("group", "none")
	var err error

	if group != "none" {
		gp := dbm.GM.GetGroup(group)
		curGroupGraph := gp.Get_ConfigData().Get_GroupGraph()
		cd := gp.Get_ConfigData()
		cd.Set_GroupGraph(curGroupGraph)
		//glb.Debug.Println("$$$$ the graph is saved to DB")
		//graphMap = curGraphMap.GetGraphMap()
		//glb.Debug.Println(graphMap)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "group field is null", "success": false})
	}

	if err == nil {
		//glb.Debug.Println("graphMap:",graphMap)
		c.JSON(http.StatusOK, gin.H{"message": "edges saved to DB", "success": true})
	} else {
		glb.Warning.Println(err)
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	}

}

func DeleteWholeGraph(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	group := c.DefaultQuery("group", "none")
	var err error
	//glb.Debug.Println("### entered delete whole graph api ###")
	if group != "none" {
		gp := dbm.GM.GetGroup(group)
		curGroupGraph := gp.Get_ConfigData().Get_GroupGraph()
		curGroupGraph.DeleteGraph()
		cd := gp.Get_ConfigData()
		cd.Set_GroupGraph(curGroupGraph)
		//glb.Debug.Println("$$$$ the graph is removed from DB")
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "group field is null", "success": false})
	}

	if err == nil {
		//glb.Debug.Println("graphMap:",graphMap)
		c.JSON(http.StatusOK, gin.H{"message": "the graph has been removed", "success": true})
	} else {
		glb.Warning.Println(err)
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	}

}

func AddEdgeToGraph(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	type st struct {
		NewEdge []string `json:NewEdge`
	}
	gp := dbm.GM.GetGroup(groupName)
	curGroupGraph := gp.Get_ConfigData().Get_GroupGraph()
	var tempSt st
	if groupName != "none" {
		if err := c.ShouldBindJSON(&tempSt); err == nil {
			newEdgeLabel := tempSt.NewEdge
			glb.Debug.Println("newVertexLabel : ---------> ", newEdgeLabel)
			curGroupGraph.AddEdgeByLabel(newEdgeLabel[0], newEdgeLabel[1])
			//glb.Debug.Println("graph after adding : ---------> ",curGroupGraph.GetGraphMap())
			//ad := gp.Get_AlgoData() // saving to DB has been moved to another function
			//ad.Set_GroupGraph(curGroupGraph)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true})
			}
		} else {
			glb.Warning.Println("Can't bind json")
			glb.Error.Println(err)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Can't bind json, Error:" + err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group not mentioned"})
	}
}

func RemoveEdgesOrVertices(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	type st struct {
		RemovedVertices []string `json:"removed_vertices"`
		RemovedEdges    []string `json:"removed_edges"`
	}
	gp := dbm.GM.GetGroup(groupName)
	curGroupGraph := gp.Get_ConfigData().Get_GroupGraph()
	var tempSt st
	if groupName != "none" {
		if err := c.ShouldBindJSON(&tempSt); err == nil {
			ToBeRemovedVertices := tempSt.RemovedVertices
			ToBeRemovedEdges := tempSt.RemovedEdges
			glb.Debug.Println(ToBeRemovedEdges, ToBeRemovedVertices, curGroupGraph)
			for k, _ := range ToBeRemovedEdges {
				//glb.Debug.Println("%%% ToBeRemovedEdges[k] %%%",ToBeRemovedEdges[k])
				curGroupGraph.RemoveEdgeByLabel(ToBeRemovedEdges[k])
			}
			for k, _ := range ToBeRemovedVertices {
				//glb.Debug.Println("%%% ToBeRemovedVertices[k] %%%",ToBeRemovedVertices[k])
				curGroupGraph.RemoveNodeByLabel(ToBeRemovedVertices[k])
			}

			cd := gp.Get_ConfigData()
			cd.Set_GroupGraph(curGroupGraph)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true})
			}
		} else {
			glb.Warning.Println("Can't bind json")
			glb.Error.Println(err)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Can't bind json, Error:" + err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group not mentioned"})
	}
}

// Get Graph of Group in the format of a map that its key is like "10#10" and values are a slice of strings
// Get parameters: group
func Getgraph(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	group := c.DefaultQuery("group", "none")
	var err error
	graphMap := make(map[string][]string)
	//graphMap["10,10"] = append(graphMap["10,10"], "20,20")
	//graphMap["10,10"] = append(graphMap["10,10"], "20,30")
	//graphMap["20,20"] = append(graphMap["20,20"], "10,10")
	//graphMap["20,20"] = append(graphMap["20,20"], "20,30")
	//graphMap["20,30"] = append(graphMap["20,30"], "10,10")
	//graphMap["20,30"] = append(graphMap["20,30"], "20,20")
	//graphMap["20,30"] = append(graphMap["20,30"], "50,50")
	//graphMap["40,40"] = make([]string, 0)
	//graphMap["50,50"] = append(graphMap["50,50"], "20,30")

	if group != "none" {
		gp := dbm.GM.GetGroup(group)
		graphMapPointer := gp.Get_ConfigData().Get_GroupGraph()
		graphMap = graphMapPointer.GetGraphMap()
		glb.Debug.Println("graphmap", graphMap)
		//glb.Debug.Println(graphMapPointer.GetUndirectionalGraphMap())
		allLines := graphMapPointer.AllLines()
		glb.Debug.Println(allLines)
		glb.Debug.Println(len(allLines))
		//glb.Debug.Println(graphMap)
		//root,_ := graphMapPointer.GetNodeByLabel("-1152,1334")
		//glb.Debug.Println("returned value from BFSTraverse",graphMapPointer.BFSTraverse(root))
		//	{glb.Debug.Printf("%v\n", n)
		//})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "group field is null", "success": false})
	}

	if err == nil {
		//glb.Debug.Println("graphMap:",graphMap)
		c.JSON(http.StatusOK, gin.H{"message": graphMap, "success": true})
	} else {
		glb.Warning.Println(err)
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	}

}

// Get Adjacent fingerprint locations to a specific node of graph
// Get parameters: group, node
func GetGraphNodeAdjacentFPs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	group := strings.ToLower(c.DefaultQuery("group", "none"))
	node := strings.ToLower(c.DefaultQuery("node", "none"))
	if group != "none" && node != "none" {
		gp := dbm.GM.GetGroup(group)
		knnAlgoData := gp.Get_AlgoData().Get_KnnFPs()
		node2FPs := knnAlgoData.Node2FPs
		fpData := gp.Get_RawData().Get_Fingerprints()
		glb.Debug.Println("node:", node)
		glb.Debug.Println("Graph node2fps details:")
		node2FPSLen := 0
		for node, fps := range node2FPs {
			glb.Debug.Println(node, ": number of adjacent fps = ", len(fps))
			node2FPSLen += len(fps)
		}

		if node2FPSLen != 0 {
			fpIndexes := node2FPs[node]
			glb.Debug.Println("Num of adjacent Fingerprints: ", len(fpIndexes))
			fpLocations := []string{}
			otherFpLocations := []string{}
			for fpTime, fp := range fpData {
				if glb.StringInSlice(fpTime, fpIndexes) {
					fpLocations = append(fpLocations, fp.Location)
				} else {
					otherFpLocations = append(otherFpLocations, fp.Location)
				}
			}
			c.JSON(http.StatusOK, gin.H{"fpLocations": fpLocations, "otherFpLocations": otherFpLocations, "success": true})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "can get adjacent fps, calculate first ", "success": false})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "group or node field is null", "success": false})
	}
}

// Get uniquemacs
// Get parameters: group
func GetUniqueMacs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	groupName := c.DefaultQuery("group", "none")
	var err error
	var uniqueMacs []string
	if groupName != "none" {
		//err, FilterMacs = dbm.GetFilterMacDB(groupName)
		//glb.Debug.Println("filterMacs")
		uniqueMacs = dbm.GM.GetGroup(groupName).Get_MiddleData().Get_UniqueMacs()
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "groupName field is null", "success": false})
	}

	if err == nil {
		glb.Debug.Println("UniqueMacs: ", uniqueMacs)
		c.JSON(http.StatusOK, gin.H{"message": uniqueMacs, "success": true})
	} else {
		glb.Warning.Println(err)
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	}

}

// reset the db name as group name to fingerprints(fingerprint & fingerprint-track buckets)
// Get parameters: group
func ReformDB(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "none"))

	if group != "none" {
		numChanges := dbm.ReformDBDB(group)
		glb.Debug.Println("DB reformed successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Changed name of " + strconv.Itoa(numChanges) + " things", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func CVResults(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		algoAccuracy := dbm.GetCVResults(groupName)

		glb.Debug.Println("Got algorithms accuracy")
		c.JSON(http.StatusOK, gin.H{"Algorithms Accuracy": algoAccuracy,})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func CalcCompletionLevel(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		cmpLevel := dbm.GetCalcCompletionLevel()
		if (cmpLevel > 0 && cmpLevel <= 1) {
			//cmpLevel = float64(int(cmpLevel*10000000))/100000
			//glb.Debug.Printf("Calculation level: %f % \n", cmpLevel)
			c.JSON(http.StatusOK, gin.H{"success": true, "message": cmpLevel,})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "No calculation is running"})
		}

	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func BuildGroup(c *gin.Context) {
	//glb.Debug.Println("############# enetered BuildGroup #############")
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {

		//1: reform legacy db
		dbm.ReformDBDB(groupName)

		err := dbm.CorrectLearnFPsTimestamp(groupName)
		if err != nil {
			glb.Error.Println(err.Error())
		}
		//2: form raw groupcache and rawdata initialization with raw data in db
		dbm.BuildGroupDB(groupName)
		//3.relocate fp with true location that was uploaded
		if dbm.GetSharedPrf(groupName).NeedToRelocateFP {
			err := dbm.RelocateFPLoc(groupName)
			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				return
			}
		}
		//4. Preprocess on rss raw data
		//rd := dbm.GM.GetGroup(groupName).Get_RawData()
		//needToRelocateFP := dbm.GetSharedPrf(groupName).NeedToRelocateFP
		//algorithms.PreProcess(rd, needToRelocateFP) // use preprocess just once in buildgroup(not use in calculatelearn)
		//5. Run learning for raw data
		calculateLearnData(groupName, "false")

		glb.Debug.Println("Struct reformed successfully")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "struct renewed"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func AddArbitLocations(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	type st struct {
		Locations []string `json:"locations"`
	}

	var tempSt st
	if groupName != "none" {
		//glb.Warning.Println(c.Request.GetBody())
		if err := c.ShouldBindJSON(&tempSt); err == nil {
			locations := tempSt.Locations
			glb.Debug.Println(locations)
			err := dbm.AddArbitLocations(groupName, locations)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true})
			}
		} else {
			glb.Warning.Println("Can't bind json")
			glb.Error.Println(err)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Can't bind json, Error:" + err.Error()})
			//c.JSON(http.StatusOK, gin.H{"message": "Nums of the FilterMacs are zero", "success": false})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group not mentioned"})
	}
}

func DelArbitLocations(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	type st struct {
		Locations []string `json:"locations"`
	}
	var tempSt st

	if groupName != "none" {
		if err := c.ShouldBindJSON(&tempSt); err == nil {
			locations := tempSt.Locations
			err := dbm.DelArbitLocations(groupName, locations)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				glb.Debug.Println("Arbit locations are deleted: ", locations)
				c.JSON(http.StatusOK, gin.H{"success": true})
			}
		} else {
			glb.Warning.Println("Can't bind json")
			glb.Error.Println(err)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Can't bind json, Error:" + err.Error()})
			//c.JSON(http.StatusOK, gin.H{"message": "Nums of the FilterMacs are zero", "success": false})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group not mentioned"})
	}
}

func GetArbitLocations(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		locations := dbm.GetArbitLocations(groupName)
		c.JSON(http.StatusOK, gin.H{"success": true, "locations": locations})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group not mentioned"})
	}
}

func GetLocationMacs(c *gin.Context) {
	groupName := c.Param("group")
	location := c.Param("location")

	if len(groupName) != 0 && len(location) != 0 {
		gp := dbm.GM.GetGroup(groupName)
		fpInMemory := gp.Get_RawData().Get_Fingerprints()

		macs := []string{}
		for _, fp := range fpInMemory {
			if fp.Location == location {
				for _, rt := range fp.WifiFingerprint {
					if !glb.StringInSlice(rt.Mac, macs) {
						macs = append(macs, rt.Mac)
					}
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "macs": macs})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or location not mentioned"})
	}
}

func DelResults(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	user := strings.ToLower(c.DefaultQuery("user", "none"))

	if groupName != "none" && user != "none" {
		err := dbm.GM.GetGroup(groupName).Get_ResultData().Clear_UserResults(user)
		//locations := dbm.DelResults(groupName)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		}

	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or user not mentioned"})
	}
}

func FingerprintLikeness(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	location := strings.ToLower(c.DefaultQuery("location", "none"))
	maxFPDistStr := strings.ToLower(c.DefaultQuery("maxFPDist", "none"))

	if groupName != "none" && location != "none" && maxFPDistStr != "none" {
		maxFPDist, err := strconv.ParseFloat(maxFPDistStr, 64)
		if err == nil {
			resultMap, fingerprintRssDetails := dbm.FingerprintLikeness(groupName, location, maxFPDist)
			rssDetailsStr := ""
			for _, fpRSSs := range fingerprintRssDetails {
				line := ""
				for _, rss := range fpRSSs {
					line += rss + ","
				}
				rssDetailsStr += line + "\n"
			}

			c.JSON(http.StatusOK, gin.H{"success": true, "resultMap": resultMap, "fingerprintDetails": fingerprintRssDetails, "rssDetailsStr": rssDetailsStr})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or user not mentioned"})
	}
}

func GetFingerprint(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	fpId := strings.ToLower(c.DefaultQuery("id", "none"))

	if groupName != "none" {
		fpData := dbm.GM.GetGroup(groupName).Get_RawData().Get_Fingerprints()
		if fpId != "none" {
			if fp, ok := fpData[fpId]; ok {
				c.JSON(http.StatusOK, gin.H{"success": true, "fingerprints": fp})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": "Invalid fingerprint id"})
			}
		} else {
			fps := []parameters.Fingerprint{}
			for _, fp := range fpData {
				fps = append(fps, fp)
			}
			c.JSON(http.StatusOK, gin.H{"success": false, "fingerprints": fps})
		}

	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or id not mentioned"})
	}
}

func GetMostSeenMacsAPI(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	if groupName != "none" {
		mostSeenMacs := dbm.GetMostSeenMacs(groupName)
		c.JSON(http.StatusOK, gin.H{"success": true, "macs": mostSeenMacs})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or id not mentioned"})
	}

}

func UploadTrueLocationLog(c *gin.Context) {
	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	method := c.DefaultQuery("method", "none") // must be learn,learnAppend,test,testAppend

	if groupName != "none" && method != "none" {

		//1:File upload:

		// Set file path according to method
		filePath := "/" + groupName + ".log"
		mainMethod := ""
		if method == "learn" || method == "learnAppend" {
			filePath = "/TrueLocationLogs/learn" + filePath
			mainMethod = "learn"
		} else if method == "test" || method == "testAppend" {
			filePath = "/TrueLocationLogs/test" + filePath
			mainMethod = "test"
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "method must be learn,learnAppend,test or testAppend"})
			return
		}

		// Get uploaded file
		file, _, err := c.Request.FormFile("file")
		defer file.Close()
		//filename := header.Filename
		if err != nil {
			glb.Error.Println(err)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}

		// Append to last file or replace it
		if method == "learnAppend" || method == "testAppend" { // Append to existent file if exists
			//Read old file
			out, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, filePath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			defer out.Close()
			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				return
			}

			//Read uploaded file
			newDataByte, err := ioutil.ReadAll(file)
			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				return
			}
			//glb.Debug.Println(string(newDataByte))

			if _, err = out.Write(newDataByte); err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				return
			} else {
				glb.Debug.Println("New log file was created or appended to " + filePath)

				//2(A): insert logs from uploaded file to groupcache
				err := dbm.SetTrueLocationFromLog(groupName, mainMethod)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				} else {
					c.JSON(http.StatusOK, gin.H{"success": true})
				}
			}

		} else if method == "learn" || method == "test" { //replace new file instead of last file

			//todo: why os.openFile can't replace file!?
			//out, err := os.OpenFile(path.Join(glb.RuntimeArgs.SourcePath, filePath), os.O_CREATE|os.O_RDWR, 0666)
			//defer out.Close()
			//if err != nil {
			//	glb.Error.Println(err)
			//	c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			//	return
			//}
			out, err := os.Create(path.Join(glb.RuntimeArgs.SourcePath, filePath))
			defer out.Close()

			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				return
			}

			_, err = io.Copy(out, file) // copy file
			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				glb.Debug.Println("New log file was created or replaced with " + filePath)

				//2(B): insert logs from uploaded file to groupcache
				err := dbm.SetTrueLocationFromLog(groupName, mainMethod)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
				} else {
					c.JSON(http.StatusOK, gin.H{"success": true})
				}
			}
		}

		// 3:After file learning file upload , set NeedToRelocateFP variable to true
		if method == "learn" || method == "learnAppend" {
			err := dbm.SetSharedPrf(groupName, "NeedToRelocateFP", true)
			if err != nil {
				glb.Error.Println(err.Error())
			}
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or method isn't mentioned"})
	}

}

func SetRelocateFPLocStateAPI(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	relocateFP := strings.ToLower(c.DefaultQuery("relocateFP", "none"))

	if groupName != "none" && relocateFP != "none" {
		//err := dbm.RelocateFPLoc(groupName)
		if (relocateFP == "true") {
			err := dbm.SetSharedPrf(groupName, "NeedToRelocateFP", true)
			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true})
			}
		} else if (relocateFP == "false") {
			err := dbm.SetSharedPrf(groupName, "NeedToRelocateFP", false)
			if err != nil {
				glb.Error.Println(err)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "relocateFP must be true or false"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or relocateFP not mentioned"})
	}

}

func GetRelocateFPLocStateAPI(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		needToRelocateFP := dbm.GetSharedPrf(groupName).NeedToRelocateFP
		c.JSON(http.StatusOK, gin.H{"success": true, "needToRelocateFP": needToRelocateFP})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group is not mentioned"})
	}

}

func ClearTestValidTrueLocation(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		rd := dbm.GM.GetGroup(groupName).Get_RawData()
		emptyTestValidTrueLocations := make(map[int64]string)
		rd.Set_TestValidTrueLocations(emptyTestValidTrueLocations)
		c.JSON(http.StatusOK, gin.H{"success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group is not mentioned"})
	}

}

func GetRSSDataAPI(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))
	mac := c.DefaultQuery("mac", "none")

	if groupName != "none" && mac != "none" {
		uniqueMacs := dbm.GM.GetGroup(groupName).Get_MiddleData().Get_UniqueMacs()
		if !glb.StringInSlice(mac, uniqueMacs) {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "mac doesn't exist"})
			glb.Debug.Println("this mac doesn't exist")
		} else {
			LatLngRSS := dbm.GetRSSData(groupName, mac)
			c.JSON(http.StatusOK, gin.H{"success": true, "LatLngRSS": LatLngRSS}) // list  of (x,y,rss)
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or id not mentioned"})
	}

}

func GetMapDetails(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		if dbm.GroupExists(groupName) {
			MapName := dbm.GetSharedPrf(groupName).MapName
			MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
			MapPath := path.Join(glb.RuntimeArgs.MapPath, MapName)
			c.JSON(http.StatusOK, gin.H{"success": true, "MapName": MapName, "MapPath": MapPath, "MapDimensions": MapDimensions})
		} else {
			MapName := glb.DefaultMapName
			MapDimensions := glb.DefaultMapDimensions
			MapPath := path.Join(glb.RuntimeArgs.MapPath, MapName)
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group doesn't exist", "MapName": MapName, "MapPath": MapPath, "MapDimensions": MapDimensions})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group is not mentioned"})
	}

}

func ClearConfigData(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {
		if dbm.GroupExists(groupName) {
			gp := dbm.GM.GetGroup(groupName)
			gp.Set_ConfigData(gp.NewConfigDataStruct())
			c.JSON(http.StatusOK, gin.H{"success": true})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group doesn't exist"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group is not mentioned"})
	}

}

// Reload groupcache from db, Note: if group cache flushes before this function call reloadDB doesn't have any influence!
func ReloadDB(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	if groupName != "none" {

		dbm.GM.ReloadGroup(groupName)
		c.JSON(http.StatusOK, gin.H{"success": true})

	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group is not mentioned"})
	}

}

func SetKnnConfig(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	kRangeStr := strings.TrimSpace(c.PostForm("kRange"))
	minClusterRssRangeStr := strings.TrimSpace(c.PostForm("minClusterRssRange"))
	maxEuclideanRssDistRangeStr := strings.TrimSpace(c.PostForm("maxEuclideanRssDistRange"))
	bleFactorRangeStr := strings.TrimSpace(c.PostForm("bleFactorRange"))
	graphEnabledStr := strings.TrimSpace(c.PostForm("graphEnabled"))
	graphFactorRangeStr := strings.TrimSpace(c.PostForm("graphFactorRange"))
	dsaEnabledStr := strings.TrimSpace(c.PostForm("dsaEnabled"))
	maxMovementRangeStr := strings.TrimSpace(c.PostForm("maxMovementRange"))

	if groupName != "none" {

		cd := dbm.GM.GetGroup(groupName).Get_ConfigData()
		knnConfig := cd.Get_KnnConfig()
		glb.Debug.Println(knnConfig)

		// Parsing KRangeStr
		if kRangeStr != "" {
			var kRange []int
			if err := json.Unmarshal([]byte(kRangeStr), &kRange); err != nil {
				glb.Error.Println(err)
			} else {
				glb.Debug.Println("kRange: ", kRange)
				knnConfig.KRange = kRange
			}
		}

		// Parsing minClusterRssRange
		if minClusterRssRangeStr != "" {
			var minClusterRssRange []int
			if err := json.Unmarshal([]byte(minClusterRssRangeStr), &minClusterRssRange); err != nil {
				glb.Error.Println(err)
			} else {
				glb.Debug.Println("minClusterRssRange: ", minClusterRssRange)
				knnConfig.MinClusterRssRange = minClusterRssRange
			}
		}

		// Parsing maxEuclideanRssDistRange
		if maxEuclideanRssDistRangeStr != "" {
			var maxEuclideanRssDistRange []int
			if err := json.Unmarshal([]byte(maxEuclideanRssDistRangeStr), &maxEuclideanRssDistRange); err != nil {
				glb.Error.Println(err)
			} else {
				glb.Debug.Println("maxEuclideanRssDistRange: ", maxEuclideanRssDistRange)
				knnConfig.MaxEuclideanRssDistRange = maxEuclideanRssDistRange
			}
		}

		// Parsing bleFactorRange
		if bleFactorRangeStr != "" {
			var bleFactorRange []float64
			if err := json.Unmarshal([]byte(bleFactorRangeStr), &bleFactorRange); err != nil {
				glb.Error.Println(err)
			} else {
				glb.Debug.Println("bleFactorRange: ", bleFactorRange)
				knnConfig.BLEFactorRange = bleFactorRange
			}
		}

		// Parsing graphState
		if graphEnabledStr != "" {
			graphEnabled, err := strconv.ParseBool(graphEnabledStr)
			if err != nil {
				glb.Error.Println(err)
				glb.Error.Println("Can't parse graphEnabled")
			} else {
				glb.Debug.Println("graphEnabled: ", graphEnabled)
				knnConfig.GraphEnabled = graphEnabled
			}
		}

		// Parsing graphState
		if graphFactorRangeStr != "" {
			var graphFactorRange [][]float64
			if err := json.Unmarshal([]byte(graphFactorRangeStr), &graphFactorRange); err != nil {
				glb.Error.Println(err)
			} else {
				glb.Debug.Println("graphFactorRange: ", graphFactorRange)
				knnConfig.GraphFactorRange = graphFactorRange
			}
		}

		// Parsing graphState
		if dsaEnabledStr != "" {
			dsaEnabled, err := strconv.ParseBool(dsaEnabledStr)
			if err != nil {
				glb.Error.Println(err)
				glb.Error.Println("Can't parse dsaEnabled")
			} else {
				glb.Debug.Println("dsaEnabled: ", dsaEnabled)
				knnConfig.DSAEnabled = dsaEnabled
			}
		}

		// Parsing maxMovementRange
		if maxMovementRangeStr != "" {
			var maxMovementRange []int
			if err := json.Unmarshal([]byte(maxMovementRangeStr), &maxMovementRange); err != nil {
				glb.Error.Println(err)
			} else {
				glb.Debug.Println("maxMovementRange: ", maxMovementRange)
				knnConfig.MaxMovementRange = maxMovementRange
			}
		}

		cd.Set_KnnConfig(knnConfig)

	}
	c.Redirect(http.StatusFound, "/dashboard/"+groupName)
}

func SetGroupOtherConfig(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "none"))

	coGroupStr := strings.TrimSpace(c.PostForm("coGroup"))
	simpleHistoryEnabledStr := strings.TrimSpace(c.PostForm("simpleHistoryEnabled"))

	if groupName != "none" {

		cd := dbm.GM.GetGroup(groupName).Get_ConfigData()
		otherGpConfig := cd.Get_OtherGroupConfig()
		glb.Debug.Println(otherGpConfig)

		// Parsing coGroupStr
		if coGroupStr != "" {
			glb.Debug.Println("CoGroup: ", coGroupStr)
			if (coGroupStr == "No CoGroup") {
				if otherGpConfig.CoGroup != "" {
					coGpCd := dbm.GM.GetGroup(otherGpConfig.CoGroup).Get_ConfigData()
					coGpOtherGpConfig := coGpCd.Get_OtherGroupConfig()
					// Change co-group names
					otherGpConfig.CoGroup = ""
					coGpOtherGpConfig.CoGroup = ""

					// Change co-group modes
					otherGpConfig.CoGroupMode = parameters.CoGroupState_None
					coGpOtherGpConfig.CoGroupMode = parameters.CoGroupState_None
					coGpCd.Set_OtherGroupConfig(coGpOtherGpConfig)
				}
			} else {

				// change cp-group names
				coGpCd := dbm.GM.GetGroup(coGroupStr).Get_ConfigData()
				coGpOtherGpConfig := coGpCd.Get_OtherGroupConfig()

				otherGpConfig.CoGroup = coGroupStr
				coGpOtherGpConfig.CoGroup = groupName

				// Change co-group modes
				otherGpConfig.CoGroupMode = parameters.CoGroupState_Master    // set current group mode to master
				coGpOtherGpConfig.CoGroupMode = parameters.CoGroupState_Slave // set other one (co-group of current group) mode to slave
				glb.Debug.Println(coGpOtherGpConfig)
				coGpCd.Set_OtherGroupConfig(coGpOtherGpConfig)
			}
		}

		// Parsing simpleHistoryEnabledStr
		if simpleHistoryEnabledStr != "" {
			simpleHistoryEnabled, err := strconv.ParseBool(simpleHistoryEnabledStr)
			if err != nil {
				glb.Error.Println(err)
				glb.Error.Println("Can't parse simpleHistoryEnabled")
			} else {
				glb.Debug.Println("simpleHistoryEnabled: ", simpleHistoryEnabled)
				otherGpConfig.SimpleHistoryEnabled = simpleHistoryEnabled
			}
		}

		cd.Set_OtherGroupConfig(otherGpConfig)
	}
	c.Redirect(http.StatusFound, "/dashboard/"+groupName)
}
