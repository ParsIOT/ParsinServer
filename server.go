// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// server.go handles Flag parsing and starts the Gin-Tonic webserver.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	//"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/appleboy/gin-jwt" // Authentication middleware lib
	// installation : 1.go get github.com/gin-gonic/gin 2.Solve the miscellaneous lib problem 3. git get github.com/appleboy/gin-jwt
	"time"
	"github.com/gin-gonic/contrib/sessions"
	//"github.com/gin-gonic/contrib/jwt"
)

// RuntimeArgs contains all runtime
// arguments available
var RuntimeArgs struct {
	RFPort            string
	FilterMacFile     string
	ExternalIP        string
	Port              string
	ServerCRT         string
	ServerKey         string
	SourcePath        string
	Socket            string
	Cwd               string
	MqttServer        string
	MqttAdmin         string
	MosquittoPID      string
	MqttAdminPassword string
	Dump              string
	Message           string
	Mqtt              bool
	MqttExisting      bool
	Svm               bool
	RandomForests     bool
	Filtering         bool
	FilterMacs        map[string]bool
	AdminAdd          string
	GaussianDist      bool
	MinRssOpt         int
	KNN               bool
}

// VersionNum keeps track of the version
var VersionNum string
var BuildTime string
var Build string

// init initiates the paths in RuntimeArgs
func init() {
	cwd, _ := os.Getwd()
	RuntimeArgs.Cwd = cwd
	RuntimeArgs.SourcePath = path.Join(RuntimeArgs.Cwd, "data")
	RuntimeArgs.Message = ""
}

func main() {
	// _, executableFile, _, _ := runtime.Caller(0) // get full path of this file
	if len(Build) == 0 {
		Build = "devdevdevdevdevdevdev"
	}
	// Bing flags for changing parameters of FIND
	flag.StringVar(&RuntimeArgs.Port, "p", ":8003", "port to bind")
	flag.StringVar(&RuntimeArgs.Socket, "s", "", "unix socket")
	flag.StringVar(&RuntimeArgs.ServerCRT, "crt", "", "location of ssl crt")
	flag.StringVar(&RuntimeArgs.ServerKey, "key", "", "location of ssl key")
	flag.StringVar(&RuntimeArgs.MqttServer, "mqtt", "", "ADDRESS:PORT of mosquitto server")
	flag.StringVar(&RuntimeArgs.MqttAdmin, "mqttadmin", "", "admin to read all messages")
	flag.StringVar(&RuntimeArgs.MqttAdminPassword, "mqttadminpass", "", "admin to read all messages")
	flag.StringVar(&RuntimeArgs.MosquittoPID, "mosquitto", "", "mosquitto PID (`pgrep mosquitto`)")
	flag.StringVar(&RuntimeArgs.Dump, "dump", "", "group to dump to folder")
	flag.StringVar(&RuntimeArgs.Message, "message", "", "message to display to all users")
	flag.StringVar(&RuntimeArgs.SourcePath, "data", "", "path to data folder")
	flag.StringVar(&RuntimeArgs.RFPort, "rf", "", "port for random forests calculations")
	flag.StringVar(&RuntimeArgs.FilterMacFile, "filter", "", "JSON file for macs to filter")
	flag.StringVar(&RuntimeArgs.AdminAdd, "adminadd", "", "Add an admin user or change his password, foramt:<username>:<password>, e.g.:admin:admin")
	flag.BoolVar(&RuntimeArgs.GaussianDist, "gaussian", false, "Use gaussian distribution instead of historgram")
	flag.IntVar(&RuntimeArgs.MinRssOpt, "minrss", -100, "Select minimum rss; Any Rss lower than minRss will be ignored.")

	flag.CommandLine.Usage = func() {
		fmt.Println(`find (version ` + VersionNum + ` (` + Build[0:8] + `), built ` + BuildTime + `)
Example: 'findserver yourserver.com'
Example: 'findserver -p :8080 localhost:8080'
Example (mosquitto): 'findserver -mqtt 127.0.0.1:1883 -mqttadmin admin -mqttadminpass somepass -mosquitto ` + "`pgrep mosquitto`" + `
Options:`)
		flag.CommandLine.PrintDefaults()
	}
	flag.Parse()
	RuntimeArgs.ExternalIP = flag.Arg(0)
	if RuntimeArgs.ExternalIP == "" {
		RuntimeArgs.ExternalIP = GetLocalIP() + RuntimeArgs.Port
	}

	if RuntimeArgs.SourcePath == "" {
		RuntimeArgs.SourcePath = path.Join(RuntimeArgs.Cwd, "data")
	}
	fmt.Println(RuntimeArgs.SourcePath)

	// Check whether all the MQTT variables are passed to initiate the MQTT routines
	if len(RuntimeArgs.MqttServer) > 0 && len(RuntimeArgs.MqttAdmin) > 0 && len(RuntimeArgs.MosquittoPID) > 0 {
		RuntimeArgs.Mqtt = true
		setupMqtt()
	} else {
		if len(RuntimeArgs.MqttServer) > 0 {
			RuntimeArgs.Mqtt = true
			RuntimeArgs.MqttExisting = true
			setupMqtt()
		} else {
			RuntimeArgs.Mqtt = false
		}
	}

	// Check whether random forests are used
	if len(RuntimeArgs.RFPort) > 0 {
		RuntimeArgs.RandomForests = true
	}

	// Check whether macs should be filtered
	if len(RuntimeArgs.FilterMacFile) > 0 {
		b, err := ioutil.ReadFile(RuntimeArgs.FilterMacFile)
		if err != nil {
			panic(err)
		}
		RuntimeArgs.FilterMacs = make(map[string]bool)
		json.Unmarshal(b, &RuntimeArgs.FilterMacs)
		fmt.Printf("Filtering %+v", RuntimeArgs.FilterMacs)
		RuntimeArgs.Filtering = true
	}
	// Check whether we are just dumping the database
	if len(RuntimeArgs.Dump) > 0 {
		err := dumpFingerprints(strings.ToLower(RuntimeArgs.Dump))
		if err == nil {
			fmt.Println("Successfully dumped.")
		} else {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	// Useradded command
	// Check whether we are just dumping the database
	if len(RuntimeArgs.AdminAdd) > 0 {
		addRequestSlice := strings.Split(strings.ToLower(RuntimeArgs.AdminAdd), ":")
		//group := addRequestSlice[0]
		username := addRequestSlice[0]
		password := addRequestSlice[1]
		err := addAdminUser(username, password)
		if err == nil {
			fmt.Printf("Successfully new admin(username:%+v, password:%+v) was added Or new password was set.", username, password)
			adminList, _ := getAdminUsers()
			fmt.Println("Current admin list: \n", adminList)
		} else {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	//Set minRssOpt

	MinRssiOpt = RuntimeArgs.MinRssOpt

	// Check if there is a message from the admin
	if _, err := os.Stat(path.Join(RuntimeArgs.Cwd, "message.txt")); err == nil {
		messageByte, _ := ioutil.ReadFile(path.Join(RuntimeArgs.Cwd, "message.txt"))
		RuntimeArgs.Message = string(messageByte)
	}

	// Check whether SVM libraries are available
	cmdOut, _ := exec.Command("svm-scale", "").CombinedOutput()
	if len(cmdOut) == 0 {
		RuntimeArgs.Svm = false
		fmt.Println("SVM is not detected.")
		fmt.Println(`To install:
sudo apt-get install g++
wget http://www.csie.ntu.edu.tw/~cjlin/cgi-bin/libsvm.cgi?+http://www.csie.ntu.edu.tw/~cjlin/libsvm+tar.gz
tar -xvf libsvm-*.tar.gz
cd libsvm-*
make
cp svm-scale /usr/local/bin/
cp svm-predict /usr/local/bin/
cp svm-train /usr/local/bin/`)
	} else {
		RuntimeArgs.Svm = true
	}

	// Setup Gin-Gonic
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob(path.Join(RuntimeArgs.Cwd, "templates/*"))

	// Load static files (if they are not hosted by external service)
	r.Static("static/", path.Join(RuntimeArgs.Cwd, "static/"))

	// the jwt middleware
	adminUsers, err := getAdminUsers()
	if err != nil {
		Error.Println("Add an admin user first")
		os.Exit(0)
	}

	authMiddleware := &jwt.GinJWTMiddleware{//todo: Problem with token saving after authentication(no way to save!)
		Realm: "test zone",
		Key: []byte("secret key"),
		Timeout: time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
			//c.HTML(http.StatusOK, "test.tmpl", gin.H{
			//	"ErrorMessage": "Please login first.",
			//})
			for user := range adminUsers {
				if (userId == user && password == adminUsers[userId]) {
					return userId, true
				}
			}
			return userId, false
		},
		Authorizator: func(userId string, c *gin.Context) bool {
			for user := range adminUsers {
				if userId == user {
					return true
				}
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			//c.JSON(code, gin.H{
			//	"code":    code,
			//	"message": message,
			//})
			c.HTML(http.StatusOK, "test.tmpl", gin.H{
				"ErrorMessage": "Please login first.",
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup: "header:Authorization",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	}

	// Create cookie store to keep track of logged in user
	store := sessions.NewCookieStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// 404-page redirects to login
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"ErrorMessage": "Please login first.",
		})
	})

	// r.PUT("/message", putMessage)

	// Routes for logging in and viewing dashboards (routes.go)
	r.GET("/", slash)
	r.GET("/login", slashLogin)
	r.POST("/login", slashLoginPOST)
	r.GET("/logout", slashLogout)
	r.GET("/dashboard/:group", slashDashboard)
	r.GET("/explore/:group/:network/:location", slashExplore2)
	r.GET("/pie/:group/:network/:location", slashPie)
	r.GET("/livemap/:group", LiveLocationMap)
	/*
	r.GET("/livemap/:group", func(context *gin.Context) {
		r.LoadHTMLGlob(path.Join(RuntimeArgs.Cwd, "templates/*"))
		LiveLocationMap(context)
	})
	*/
	r.GET("/locationsmap/:group", LocationsOnMap)

	// Routes for performing fingerprinting (fingerprint.go)
	r.POST("/learn", learnFingerprintPOST)
	r.POST("/bulklearn", bulkLearnFingerprintPOST)
	r.POST("/track", trackFingerprintPOST)

	// Routes for MQTT (mqtt.go)
	r.PUT("/mqtt", putMQTT)

	// Routes for API access (api.go)
	r.GET("/location", getUserLocations)
	r.GET("/locations", getLocationList)
	r.GET("/editname", editName)
	r.GET("/editusername", editUserName)
	r.GET("/editnetworkname", editNetworkName)
	r.DELETE("/location", deleteLocation)
	r.DELETE("/locations", deleteLocations)
	r.DELETE("/user", deleteUser)
	r.DELETE("/database", deleteDatabase)
	r.GET("/calculate", calculate)
	r.GET("/status", getStatus)
	// Done: delete these deprecated routes
	r.GET("/userlocs", userLocations) // to be deprecated
	//r.GET("/whereami", whereAmI)      // to be deprecated
	r.PUT("/mixin", putMixinOverride)
	r.PUT("/cutoff", putCutoffOverride)
	r.PUT("/database", migrateDatabase)
	r.PUT("/k_knn", putKnnK)
	r.PUT("/minrss", putMinRss)
	r.GET("/lastfingerprint", apiGetLastFingerprint)
	//r.Static("data/", path.Join(RuntimeArgs.Cwd, "data/"))

	// Authentication
	r.POST("/authenticate", authMiddleware.LoginHandler)
	auth := r.Group("/")
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)
		auth.Static("data/", path.Join(RuntimeArgs.Cwd, "data/")) // Load db files
	}

	// Authenticate test:
	// http -v --json POST http://192.168.0.45:8003/login1 username=t2 password=t2
	// http -f GET http://192.168.0.45:8003/auth/data/iust_13_5_96_1.rf.json "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDM4NjIzMzEsImlkIjoiYWRtaW4xIiwib3JpZ19pYXQiOjE1MDM4NTg3MzF9.2ln7ccDV3mwrdXBLK06jNaBTJHxL0oR319Ile4ZNnzc"  "Content-Type: application/json"

	// Load and display the logo
	dat, _ := ioutil.ReadFile("./static/logo.txt")
	fmt.Println(string(dat))

	// Check whether user is providing certificates
	if RuntimeArgs.Socket != "" {
		r.RunUnix(RuntimeArgs.Socket)
	} else if RuntimeArgs.ServerCRT != "" && RuntimeArgs.ServerKey != "" {
		fmt.Println(`(version ` + VersionNum + ` build ` + Build[0:8] + `) is up and running on https://` + RuntimeArgs.ExternalIP)
		fmt.Println("-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----")
		r.RunTLS(RuntimeArgs.Port, RuntimeArgs.ServerCRT, RuntimeArgs.ServerKey)
	} else {
		fmt.Println(`(version ` + VersionNum + ` build ` + Build[0:8] + `) is up and running on http://` + RuntimeArgs.ExternalIP)
		fmt.Println("-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----")
		r.Run(RuntimeArgs.Port)
	}
}

// // putMessage usage: curl -G -X PUT "http://localhost:8003/message" --data-urlencode "text=hello world"
// func putMessage(c *gin.Context) {
// 	newText := c.DefaultQuery("text", "none")
// 	if newText != "none" {
// 		RuntimeArgs.Message = newText
// 		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Message set as '" + newText + "'"})
// 	} else {
// 		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
// 	}
// }
