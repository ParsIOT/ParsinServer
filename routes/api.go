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
	"ParsinServer/algorithms/parameters"
	"ParsinServer/algorithms/bayes"
	"ParsinServer/algorithms"
	"ParsinServer/dbm"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

//returns uptime, starttime, number of cpu cores
func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"uptime": time.Since(startTime).Seconds(), "registered": startTime.String(), "status": "standard", "num_cores": runtime.NumCPU(), "success": true})
}

// glb.UserPositionJSON stores the a users time, location and bayes after calculatePosterior()


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

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	if group == "noneasdf" {
		c.JSON(http.StatusOK, gin.H{"message": "You need to specify group", "success": false})
		return
	}
	if !dbm.GroupExists(group) {
		c.JSON(http.StatusOK, gin.H{"message": "You should insert a fingerprint first, see documentation", "success": false})
		return
	}
	ps, _ := dbm.OpenParameters(group)
	locationCount := make(map[string]map[string]int)
	for n := range ps.NetworkLocs {
		for loc := range ps.NetworkLocs[n] {
			locationCount[loc] = make(map[string]int)
			locationCount[loc]["count"] = ps.Results[n].TotalLocations[loc]
			locationCount[loc]["accuracy"] = ps.Results[n].Accuracy[loc]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Found %d unique locations in group %s", len(ps.UniqueLocs), group),
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
	group := c.DefaultQuery("group", "noneasdf")
	user := c.DefaultQuery("user", "noneasdf")
	if group != "noneasdf" {
		if !dbm.GroupExists(group) {
			c.JSON(http.StatusOK, gin.H{"message": "You should insert a fingerprint first, see documentation", "success": false})
			return
		}
		if user == "noneasdf" {
			c.JSON(http.StatusOK, gin.H{"message": "You need to specify user", "success": false})
			return
		}
		c.String(http.StatusOK, dbm.LastFingerprint(group, user))
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "You need to specify group", "success": false})
	}
}


//Returns n of the last location estimations that were stored in fingerprints-track bucket in db
func GetHistoricalUserPositions(group string, user string, n int) []glb.UserPositionJSON {
	group = strings.ToLower(group)
	user = strings.ToLower(user)
	var fingerprints []parameters.Fingerprint
	var err error

	fingerprints,err = dbm.TrackFingerprints(user,n,group)
	if(err!=nil){
		return make([]glb.UserPositionJSON, 0) //empty userJSONs
	}

	glb.Debug.Printf("Got history of %d fingerprints\n", len(fingerprints))
	userJSONs := make([]glb.UserPositionJSON, len(fingerprints))
	for i, fingerprint := range fingerprints {
		var userJSON glb.UserPositionJSON
		UTCfromUnixNano := time.Unix(0, fingerprint.Timestamp)
		userJSON.Time = UTCfromUnixNano.String()
		bayesGuess, bayesData := bayes.CalculatePosterior(fingerprint, *parameters.NewFullParameters())
		userJSON.BayesGuess = bayesGuess
		userJSON.BayesData = bayesData
		// Process SVM if needed
		if glb.RuntimeArgs.Svm {
			userJSON.SvmGuess, userJSON.SvmData = algorithms.SvmClassify(fingerprint)
		}
		// Process RF if needed
		if glb.RuntimeArgs.Scikit {
			userJSON.ScikitData = algorithms.ScikitClassify(group, fingerprint)
		}
		//_, userJSON.KnnGuess = calculateKnn(fingerprint)
		userJSONs[i] = userJSON
	}
	return userJSONs
}

//Returns svm, rf, baysian estimations of the track fingerprints that belong to a group
func GetCurrentPositionOfAllUsers(group string) map[string]glb.UserPositionJSON {
	group = strings.ToLower(group)
	userPositions := make(map[string]glb.UserPositionJSON)
	userFingerprints := make(map[string]parameters.Fingerprint)
	var err error
	userPositions,userFingerprints,err = dbm.TrackFingerprintsEmptyPosition(group)
	if (err!=nil ){
		return userPositions
	}

	for user := range userPositions {
		bayesGuess, bayesData := bayes.CalculatePosterior(userFingerprints[user], *parameters.NewFullParameters())
		foo := userPositions[user]
		foo.BayesGuess = bayesGuess
		foo.BayesData = bayesData
		// Process SVM if needed
		if glb.RuntimeArgs.Svm {
			foo.SvmGuess, foo.SvmData = algorithms.SvmClassify(userFingerprints[user])
		}
		if glb.RuntimeArgs.Scikit {
			foo.ScikitData = algorithms.ScikitClassify(group, userFingerprints[user])
		}
		//_, foo.KnnGuess = calculateKnn(userFingerprints[user])
		go dbm.SetUserPositionCache(group+user, foo)
		userPositions[user] = foo
	}

	return userPositions
}

// Is like getHistoricalUserPositions but only returns the last location estimation
func GetCurrentPositionOfUser(group string, user string) glb.UserPositionJSON {
	group = strings.ToLower(group)
	user = strings.ToLower(user)
	val, ok := dbm.GetUserPositionCache(group + user)
	if ok {
		return val
	}
	var userJSON glb.UserPositionJSON
	var userFingerprint parameters.Fingerprint
	var err error
	userJSON,userFingerprint,err = dbm.TrackFingeprintEmptyPosition(user,group)
	if (err!=nil){
		return userJSON
	}

	bayesGuess, bayesData := bayes.CalculatePosterior(userFingerprint, *parameters.NewFullParameters())
	userJSON.BayesGuess = bayesGuess
	userJSON.BayesData = bayesData
	// Process SVM if needed
	if glb.RuntimeArgs.Svm {
		userJSON.SvmGuess, userJSON.SvmData = algorithms.SvmClassify(userFingerprint)
	}
	if glb.RuntimeArgs.Scikit {
		userJSON.ScikitData = algorithms.ScikitClassify(group, userFingerprint)
	}
	//_, userJSON.KnnGuess = calculateKnn(userFingerprint)
	go dbm.SetUserPositionCache(group+user, userJSON)
	return userJSON
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
	group := c.DefaultQuery("group", "noneasdf")
	if group != "noneasdf" {
		if !dbm.GroupExists(group) {
			c.JSON(http.StatusOK, gin.H{"message": "You should insert a fingerprint first, see documentation", "success": false})
			return
		}
		group = strings.ToLower(group)
		bayes.OptimizePriorsThreaded(group)
		if glb.RuntimeArgs.Svm {
			algorithms.DumpFingerprintsSVM(group)
			err := algorithms.CalculateSVM(group)
			if err != nil {
				glb.Warning.Println("Encountered error when calculating SVM")
				glb.Warning.Println(err)
			}
		}
		if glb.RuntimeArgs.Scikit {
			algorithms.ScikitLearn(group)
		}
		algorithms.LearnKnn(group)
		go dbm.ResetCache("userPositionCache")
		go dbm.SetLearningCache(group, false)
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

	group := c.DefaultQuery("group", "noneasdf")
	userQuery := c.DefaultQuery("user", "noneasdf")
	usersQuery := c.DefaultQuery("users", "noneasdf")
	nQuery := c.DefaultQuery("n", "noneasdf")
	group = strings.ToLower(group)
	if group != "noneasdf" {
		if !dbm.GroupExists(group) {
			c.JSON(http.StatusOK, gin.H{"message": "You should insert fingerprints before tracking, see documentation", "success": false})
			return
		}
		people := make(map[string][]glb.UserPositionJSON)
		users := strings.Split(strings.ToLower(usersQuery), ",")
		if users[0] == "noneasdf" {
			users = []string{userQuery}
		}
		if users[0] == "noneasdf" {
			users = dbm.GetUsers(group)
		}
		for _, user := range users {
			if _, ok := people[user]; !ok {
				people[user] = []glb.UserPositionJSON{}
			}
			if nQuery != "noneasdf" {
				number, _ := strconv.ParseInt(nQuery, 10, 0)
				glb.Debug.Println("Getting history for " + user)
				people[user] = append(people[user], GetHistoricalUserPositions(group, user, int(number))...)
			} else {
				people[user] = append(people[user], GetCurrentPositionOfUser(group, user))
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
		dbm.MigrateDatabaseDB(fromDB,toDB)
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

	group := strings.TrimSpace(strings.ToLower(c.DefaultQuery("group", "noneasdf")))
	if glb.Exists(path.Join(glb.RuntimeArgs.SourcePath, group+".db")) {
		os.Remove(path.Join(glb.RuntimeArgs.SourcePath, group+".db"))
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Successfully deleted " + group})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Group does not exist"})
	}
}

// Calls setMixinOverride() and then calls optimizePriorsThreaded()
// GET parameters: group, mixin
func PutMixinOverride(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	newMixin := c.DefaultQuery("mixin", "none")
	if group != "noneasdf" {
		newMixinFloat, err := strconv.ParseFloat(newMixin, 64)
		if err == nil {
			err2 := dbm.SetMixinOverride(group, newMixinFloat)
			if err2 == nil {
				bayes.OptimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding mixin for " + group + ", now set to " + newMixin})
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

// Calls setCutoffOverride() and then calls optimizePriorsThreaded()
// GET parameters: group, cutoff
func PutCutoffOverride(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	newCutoff := c.DefaultQuery("cutoff", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(newCutoff)
	if group != "noneasdf" {
		newCutoffFloat, err := strconv.ParseFloat(newCutoff, 64)
		if err == nil {
			err2 := dbm.SetCutoffOverride(group, newCutoffFloat)
			if err2 == nil {
				bayes.OptimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding cutoff for " + group + ", now set to " + newCutoff})
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

// Calls setCutoffOverride() and then calls optimizePriorsThreaded()
// GET parameters: group, cutoff
func PutKnnK(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	newK := c.DefaultQuery("knnK", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(newK)
	if group != "noneasdf" {
		newKnnK, err := strconv.Atoi(newK)
		if err == nil {
			err2 := dbm.SetKnnK(group, newKnnK)
			if err2 == nil {
				//optimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding KNN K for " + group + ", now set to " + newK})
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
	newK := c.DefaultQuery("minRss", "none")
	glb.Debug.Println(group)
	glb.Debug.Println(newK)
	if group != "noneasdf" {
		newMinRss, err := strconv.Atoi(newK)
		if err == nil {
			err2 := dbm.SetMinRSS(group, newMinRss)
			if err2 == nil {
				//optimizePriorsThreaded(strings.ToLower(group))
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "Overriding Minimum RSS for " + group + ", now set to " + newK})
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
func EditNetworkName(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := c.DefaultQuery("group", "noneasdf")
	oldname := c.DefaultQuery("oldname", "none")
	newname := c.DefaultQuery("newname", "none")
	if group != "noneasdf" {
		glb.Debug.Println("Attempting renaming ", group, oldname, newname)
		dbm.RenameNetwork(group, oldname, newname)
		bayes.OptimizePriorsThreaded(group)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Finished"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Changes a location name in db(fingerprints and fingerprints-track buckets)
// GET parameters: group, location (the old name), newname
func EditName(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := c.DefaultQuery("group", "noneasdf")
	location := c.DefaultQuery("location", "none")
	newname := c.DefaultQuery("newname", "none")
	if group != "noneasdf" {
		numChanges := dbm.EditNameDB(location,newname,group)
		bayes.OptimizePriorsThreaded(strings.ToLower(group))

		c.JSON(http.StatusOK, gin.H{"message": "Changed name of " + strconv.Itoa(numChanges) + " things", "success": true})
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

	group := c.DefaultQuery("group", "noneasdf")
	oldmac := c.DefaultQuery("oldmac", "none")
	newmac := c.DefaultQuery("newmac", "none")
	if group != "noneasdf" {
		numChanges := dbm.EditMacDB(oldmac,newmac,group)
		bayes.OptimizePriorsThreaded(strings.ToLower(group))

		c.JSON(http.StatusOK, gin.H{"message": "Changed name of " + strconv.Itoa(numChanges) + " things", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Same to editName() but edits username instead of the location name
// GET paramets: group, user(the old username), newname
func EditUserName(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	user := strings.ToLower(c.DefaultQuery("user", "none"))
	newname := strings.ToLower(c.DefaultQuery("newname", "none"))
	if group != "noneasdf" {
		numChanges := dbm.EditUserNameDB(user,newname,group)

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

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	location := strings.ToLower(c.DefaultQuery("location", "none"))
	if group != "noneasdf" {
		numChanges := dbm.DeleteLocationDB(location,group)
		bayes.OptimizePriorsThreaded(strings.ToLower(group))

		c.JSON(http.StatusOK, gin.H{"message": "Deleted " + strconv.Itoa(numChanges) + " locations", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

// Is like deleteLocation(), deletes a list of locations instead.
// GET parameters: group, names
func DeleteLocations(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	group := strings.ToLower(c.DefaultQuery("group", "noneasdf"))
	locationsQuery := strings.ToLower(c.DefaultQuery("names", "none"))
	if group != "noneasdf" && locationsQuery != "none" {
		locations := strings.Split(strings.ToLower(locationsQuery), ",")
		numChanges := dbm.DeleteLocationsDB(locations,group)
		bayes.OptimizePriorsThreaded(strings.ToLower(group))
		c.JSON(http.StatusOK, gin.H{"message": "Deleted " + strconv.Itoa(numChanges) + " locations", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Need to provide group and location list. DELETE /locations?group=X&names=Y,Z,W"})
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
		numChanges := dbm.DeleteUser(user,group)
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

	var filterMacs glb.FilterMacs

	//x, _ := ioutil.ReadAll(c.Request.Body)
	//Warning.Println("%s", string(x))

	if glb.BindJSON(&filterMacs, c) == nil {

		if len(filterMacs.Macs) == 0 {
			glb.RuntimeArgs.NeedToFilter[filterMacs.Group] = false
			glb.RuntimeArgs.NotNullFilterMap[filterMacs.Group] = false
		} else {
			glb.RuntimeArgs.NeedToFilter[filterMacs.Group] = true
			glb.RuntimeArgs.NotNullFilterMap[filterMacs.Group] = true
		}

		err := dbm.SetFilterMacDB(filterMacs.Group, filterMacs.Macs)
		if err == nil {
			glb.RuntimeArgs.FilterMacsMap[filterMacs.Group] = filterMacs.Macs
			glb.Warning.Println("MacFilter set successfully ")
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
		err, FilterMacs = dbm.GetFilterMacDB(group)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "group field is null", "success": false})
	}

	if err == nil {
		glb.Warning.Println("FilterMacs: ", FilterMacs)
		c.JSON(http.StatusOK, gin.H{"message": FilterMacs, "success": true})
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
		glb.Warning.Println("DB reformed successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Changed name of " + strconv.Itoa(numChanges) + " things", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

