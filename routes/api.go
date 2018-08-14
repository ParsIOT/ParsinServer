// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// api.go handles functions that return JSON responses.

package routes

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"ParsinServer/glb"
	"ParsinServer/algorithms"
	"ParsinServer/dbm"
	"ParsinServer/dbm/parameters"
	"log"
	"io"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func PreLoadSettingsDecorator(routeFunc func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		group := c.DefaultQuery("group", "noneasdf")
		if group != "noneasdf" {
			dbm.GetSharedPrf(group)
		}
		routeFunc(c)
	}
}

func PreLoadSettings(c *gin.Context) {
	glb.Debug.Println("PreloadSettings")
	group1 := c.Param("group")
	group2 := c.DefaultQuery("group", "noneasdf")
	groupExists := false
	//glb.Debug.Println(c)
	if len(group1) != 0 {
		glb.Debug.Println(group1)
		//glb.Debug.Println(dbm.GetSharedPrf(group1))
		groupExists = dbm.GroupExists(group1)
		if groupExists {
			dbm.GetSharedPrf(group1)
		} else {
			glb.Error.Println("Group doesn't exist")
			//c.JSON(http.StatusOK, gin.H{
			//	"message":   fmt.Sprintf("There is no group with this group name: ",group1),
			//	"success":   false})
			c.Redirect(302, "/change-db?error=groupNotExists")
		}

	} else if group2 != "noneasdf" {
		glb.Debug.Println(group2)
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
	// Todo: add real "noneasdf" state
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	if groupName == "noneasdf" {
		c.JSON(http.StatusOK, gin.H{"message": "You need to specify group", "success": false})
		return
	}
	if !dbm.GroupExists(groupName) {
		c.JSON(http.StatusOK, gin.H{"message": "You should insert a fingerprint first, see documentation", "success": false})
		return
	}
	//ps, _ := dbm.OpenParameters(group)
	gp := dbm.GM.GetGroup(groupName)
	md := gp.Get_MiddleData_Val()
	algoAccuracy := gp.Get_ResultData().Get_AlgoLocAccuracy()
	locationCount := make(map[string]map[string]int)
	for n := range md.NetworkLocs {
		for loc := range md.NetworkLocs[n] {
			locationCount[loc] = make(map[string]int)
			//locationCount[loc]["count"] = gp.BayesResults[n].TotalLocations[loc]
			//locationCount[loc]["accuracy"] = gp.BayesResults[n].Accuracy[loc]
			locationCount[loc]["count"] = md.LocCount[loc]
			locationCount[loc]["accuracy"] = algoAccuracy["knn"][loc]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Found %d unique locations in group %s", len(md.UniqueLocs), groupName),
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
//	"timestamp": 1502544850139171556,
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
	groupName := c.DefaultQuery("group", "noneasdf")
	groupName = strings.ToLower(groupName)
	user := c.DefaultQuery("user", "noneasdf")
	if groupName != "noneasdf" {
		if !dbm.GroupExists(groupName) {
			c.JSON(http.StatusOK, gin.H{"message": "You should insert a fingerprint first, see documentation", "success": false})
			return
		}
		if user == "noneasdf" {
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

	//var fingerprints []parameters.Fingerprint
	//var err error
	//
	//fingerprints, err = dbm.TrackFingerprints(user, n, groupName)
	//if (err != nil) {
	//	return make([]parameters.UserPositionJSON, 0) //empty userJSONs
	//}
	//
	//glb.Debug.Printf("Got history of %d fingerprints\n", len(fingerprints))
	//userJSONs := make([]parameters.UserPositionJSON, len(fingerprints))
	//for i, fingerprint := range fingerprints {
	//	var userJSON parameters.UserPositionJSON
	//	//UTCfromUnixNano := time.Unix(0, fingerprint.Timestamp)
	//	//userJSON.Time = UTCfromUnixNano.String()
	//	userJSON.Time = fingerprint.Timestamp
	//	//bayesGuess, bayesData := bayes.CalculatePosterior(fingerprint, nil)
	//	//userJSON.BayesGuess = bayesGuess
	//	//userJSON.BayesData = bayesData
	//	//// Process SVM if needed
	//	//if glb.RuntimeArgs.Svm {
	//	//	userJSON.SvmGuess, userJSON.SvmData = algorithms.SvmClassify(fingerprint)
	//	//}
	//	//// Process RF if needed
	//	//if glb.RuntimeArgs.Scikit {
	//	//	userJSON.ScikitData = algorithms.ScikitClassify(groupName, fingerprint)
	//	//}
	//	gp := dbm.GM.GetGroup(groupName)
	//	_, userJSON.KnnGuess, userJSON.KnnData = algorithms.TrackKnn(gp, fingerprint, false)
	//	userJSONs[i] = userJSON
	//}
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

//Returns svm, rf, baysian estimations of the track fingerprints that belong to a group
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
}

// Is like getHistoricalUserPositions but only returns the last location estimation
func GetCurrentPositionOfUser(groupName string, user string) parameters.UserPositionJSON {

	val, ok := dbm.GetUserPositionCache(groupName + user)
	if ok {
		return val
	}
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
	groupName := c.DefaultQuery("group", "noneasdf")
	groupName = strings.ToLower(groupName)
	if groupName != "noneasdf" {
		if !dbm.GroupExists(groupName) {
			c.JSON(http.StatusOK, gin.H{"message": "You should insert a fingerprint first, see documentation", "success": false})
			return
		}
		algorithms.CalculateLearn(groupName)
		go dbm.ResetCache("userPositionCache")
		go dbm.SetLearningCache(groupName, false)
		c.JSON(http.StatusOK, gin.H{"message": "Parameters optimized.", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
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

	groupName := c.DefaultQuery("group", "noneasdf")
	groupName = strings.ToLower(groupName)
	userQuery := c.DefaultQuery("user", "noneasdf")
	usersQuery := c.DefaultQuery("users", "noneasdf")
	nQuery := c.DefaultQuery("n", "noneasdf")
	groupName = strings.ToLower(groupName)
	userQuery = strings.ToLower(userQuery)
	if groupName != "noneasdf" {
		if !dbm.GroupExists(groupName) {
			c.JSON(http.StatusOK, gin.H{"message": "You should insert fingerprints before tracking, see documentation", "success": false})
			return
		}
		people := make(map[string][]parameters.UserPositionJSON)
		users := strings.Split(strings.ToLower(usersQuery), ",")
		if users[0] == "noneasdf" {
			users = []string{userQuery}
		}
		if users[0] == "noneasdf" {
			//users = dbm.GetUsers(groupName)
			users = dbm.GetRecentUsers(groupName)
			glb.Debug.Println("Users:", users)
		}
		for _, user := range users {
			user = strings.ToLower(user) // todo: is it necessary? Does it conflict with learning data?
			if _, ok := people[user]; !ok {
				people[user] = []parameters.UserPositionJSON{}
			}
			if nQuery != "noneasdf" {
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

// copies a DB
// GET parameters: from, to
func MigrateDatabase(c *gin.Context) {
	fromDB := strings.ToLower(c.DefaultQuery("from", "noneasdf"))
	toDB := strings.ToLower(c.DefaultQuery("to", "noneasdf"))
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
	groupName := strings.TrimSpace(strings.ToLower(c.DefaultQuery("group", "noneasdf")))
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
//	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
//	newMixin := c.DefaultQuery("mixin", "none")
//	if group != "noneasdf" {
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
//	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
//	newCutoff := c.DefaultQuery("cutoff", "none")
//	glb.Debug.Println(group)
//	glb.Debug.Println(newCutoff)
//	if group != "noneasdf" {
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
//	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
//	newK := c.DefaultQuery("knnK", "none")
//	glb.Debug.Println(group)
//	glb.Debug.Println(newK)
//	if group != "noneasdf" {
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

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	kRangeRawStr := c.DefaultQuery("range", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(kRangeRawStr)

	if group != "noneasdf" && kRangeRawStr != "none" {
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
			err := dbm.SetSharedPrf(group, "KnnKRange", kRange)
			if err == nil {
				//optimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding KNN K range for " + group + ", now set to " + kRangeRawStr})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			}
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

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	rssRangeRawStr := c.DefaultQuery("range", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(rssRangeRawStr)

	if group != "noneasdf" && rssRangeRawStr != "none" {
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
			err := dbm.SetSharedPrf(group, "KnnMinCRssRange", minCRssRange)
			if err == nil {
				//optimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding KNN K range for " + group + ", now set to " + rssRangeRawStr})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			}
			//}else if( len(kRange) == 2){
			//	algorithms.ValidKs = glb.MakeRange(kRange[0],kRange[1])
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Knn K range length must be 2 at the maximum value "})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func PutMaxMovement(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	MaxMovementStr := c.DefaultQuery("maxMovement", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(MaxMovementStr)

	if group != "noneasdf" && MaxMovementStr != "none" {
		MaxMovement, err := strconv.ParseFloat(MaxMovementStr, 64)
		if err == nil {
			if MaxMovement == float64(-1) {
				MaxMovement = glb.MaxMovement
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
}



func ChooseMap(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	mapName := c.DefaultQuery("mapName", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(mapName)

	if group != "noneasdf" && mapName != "none" {
		//MaxMovement, err := strconv.ParseFloat(MaxMovementStr, 64)
		//if err == nil {
		//	if MaxMovement == float64(-1) {
		//		MaxMovement = glb.MaxMovement
		//	}
		mapNamesList := glb.ListMaps()
		MapWidth := mapNamesList[mapName][0]
		MapHeight := mapNamesList[mapName][1]
		MapDimensions := []int {MapWidth,MapHeight}
		glb.Debug.Println("***MapDimensions : ",MapDimensions)
		err2 := dbm.SetSharedPrf(group, "MapName", mapName)
		err3 := dbm.SetSharedPrf(group, "MapDimensions", MapDimensions)

		if err2 == nil && err3==nil {

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

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	minRss := c.DefaultQuery("minRss", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(minRss)
	if group != "noneasdf" && minRss != "none" {
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
//	group := c.DefaultQuery("group", "noneasdf")
//	oldname := c.DefaultQuery("oldname", "none")
//	newname := c.DefaultQuery("newname", "none")
//	if group != "noneasdf" {
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

	groupName := c.DefaultQuery("group", "noneasdf")
	oldloc := strings.TrimSpace(c.DefaultQuery("oldloc", "none"))
	newloc := strings.TrimSpace(c.DefaultQuery("newloc", "none"))
	if groupName != "noneasdf" && oldloc != "none" && newloc != "none" {
		numChanges := dbm.EditLocDB(oldloc, newloc, groupName)
		glb.Debug.Println("Changed location of " + strconv.Itoa(numChanges) + " fingerprints")
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))
		algorithms.CalculateLearn(groupName)
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

	groupName := c.DefaultQuery("group", "noneasdf")
	oldloc := strings.TrimSpace(c.DefaultQuery("oldloc", "none"))
	newloc := strings.TrimSpace(c.DefaultQuery("newloc", "none"))
	if groupName != "noneasdf" && oldloc != "none" && newloc != "none" {
		numChanges := dbm.EditLocBaseDB(oldloc, newloc, groupName)
		dbm.EditLocDB(oldloc, newloc, groupName)
		glb.Debug.Println("Changed location of " + strconv.Itoa(numChanges) + " fingerprints")
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))
		algorithms.CalculateLearn(groupName)
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

	groupName := c.DefaultQuery("group", "noneasdf")
	oldmac := c.DefaultQuery("oldmac", "none")
	newmac := c.DefaultQuery("newmac", "none")
	if groupName != "noneasdf" && oldmac != "none" && newmac != "none" {
		numChanges := dbm.EditMacDB(oldmac, newmac, groupName)
		glb.Debug.Println("Changed mac of " + strconv.Itoa(numChanges) + " fingerprints")
		algorithms.CalculateLearn(groupName)
		//bayes.OptimizePriorsThreaded(strings.ToLower(groupName))

		c.JSON(http.StatusOK, gin.H{"message": "Changed mac of " + strconv.Itoa(numChanges) + " fingerprints", "success": true})
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	user := strings.ToLower(c.DefaultQuery("user", "none"))
	newname := strings.ToLower(c.DefaultQuery("newname", "none"))
	if groupName != "noneasdf" && user != "none" && newname != "none" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	//gp := dbm.GM.GetGroup(groupName)
	location := strings.ToLower(c.DefaultQuery("location", "none"))
	if groupName != "noneasdf" && location != "none" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	//gp := dbm.GM.GetGroup(groupName)
	location := strings.ToLower(c.DefaultQuery("location", "none"))
	if groupName != "noneasdf" && location != "none" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	locationsQuery := strings.ToLower(c.DefaultQuery("names", "none"))
	if groupName != "noneasdf" && locationsQuery != "none" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	locationsQuery := strings.ToLower(c.DefaultQuery("names", "none"))
	if groupName != "noneasdf" && locationsQuery != "none" {
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

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	user := strings.ToLower(c.DefaultQuery("user", "noneasdf"))
	if group != "noneasdf" && user != "noneasdf" {
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
	group := c.DefaultQuery("group", "noneasdf")
	var err error
	var FilterMacs []string
	if group != "noneasdf" {
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

// Set graph
// POST parameters: graph
func Setgraph(c *gin.Context) { // not complete
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
func Getgraph(c *gin.Context) { // not complete
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	group := c.DefaultQuery("group", "noneasdf")
	var err error
	var FilterMacs []string
	if group != "noneasdf" {
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

// Get uniquemacs
// Get parameters: group
func GetUniqueMacs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	groupName := c.DefaultQuery("group", "noneasdf")
	var err error
	var uniqueMacs []string
	if groupName != "noneasdf" {
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

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))

	if group != "noneasdf" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))

	if groupName != "noneasdf" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))

	if groupName != "noneasdf" {
		cmpLevel := dbm.GetCalcCompletionLevel()
		if (cmpLevel > 0 && cmpLevel <= 1) {
			//cmpLevel = float64(int(cmpLevel*10000000))/100000
			glb.Debug.Printf("Calculation level: %f % \n", cmpLevel)
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))

	if groupName != "noneasdf" {
		dbm.ReformDBDB(groupName)
		dbm.BuildGroupDB(groupName)
		algorithms.PreProcess(groupName)
		algorithms.CalculateLearn(groupName)
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	type st struct {
		Locations []string `json:"locations"`
	}

	var tempSt st
	if groupName != "noneasdf" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	type st struct {
		Locations []string `json:"locations"`
	}
	var tempSt st

	if groupName != "noneasdf" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))

	if groupName != "noneasdf" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	user := strings.ToLower(c.DefaultQuery("user", "none"))

	if groupName != "noneasdf" && user != "none" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	location := strings.ToLower(c.DefaultQuery("location", "none"))
	maxFPDistStr := strings.ToLower(c.DefaultQuery("maxFPDist", "none"))

	if groupName != "noneasdf" && location != "none" && maxFPDistStr != "none" {
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	fpId := strings.ToLower(c.DefaultQuery("id", "none"))

	if groupName != "noneasdf" && fpId != "none" {
		fpData := dbm.GM.GetGroup(groupName).Get_RawData().Get_Fingerprints()
		if fp, ok := fpData[fpId]; ok {
			c.JSON(http.StatusOK, gin.H{"success": true, "fingerprint": fp})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Invalid fingerprint id"})
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

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	if groupName != "noneasdf" {
		mostSeenMacs := dbm.GetMostSeenMacs(groupName)
		c.JSON(http.StatusOK, gin.H{"success": true, "macs": mostSeenMacs})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or id not mentioned"})
	}

}

func UploadTrueLocationLog(c *gin.Context) {
	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	file, header, err := c.Request.FormFile("file")
	//filename := header.Filename
	fmt.Println(header.Filename)
	out, err := os.Create(path.Join(glb.RuntimeArgs.SourcePath, "TrueLocationLogs/"+groupName+".log"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func RelocateFPLocAPI(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	groupName := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	if groupName != "noneasdf" {
		err := dbm.RelocateFPLoc(groupName)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true})
		} else {
			c.JSON(http.StatusOK, gin.H{"success": true, "message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group or id not mentioned"})
	}

}

