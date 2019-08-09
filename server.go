// server.go handles Flag parsing and starts the Gin-Tonic webserver.

package main

import (
	"ParsinServer/algorithms"
	"ParsinServer/algorithms/particlefilter"
	"ParsinServer/dbm"
	"ParsinServer/glb"
	"ParsinServer/routes"
	"flag"
	"fmt"
	"github.com/MA-Heshmatkhah/SimpleAuth"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

// VersionNum keeps track of the version
var VersionNum string
var BuildTime string
var Build string

// init initiates the paths in gvar.RuntimeArgs
func init() {
	cwd, _ := os.Getwd()
	glb.RuntimeArgs.Cwd = cwd
	glb.RuntimeArgs.SourcePath = path.Join(glb.RuntimeArgs.Cwd, "data")
	glb.RuntimeArgs.Message = ""
}

func main() {
	fmt.Println("-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----")

	glb.SessionManager.Initialize(path.Join(glb.RuntimeArgs.SourcePath, "Settings.db"), &SimpleAuth.Options{
		LoginURL:                   "/login",
		LogoutURL:                  "/logout",
		UnauthorizedURL:            "/change-db",
		LoginSuccessfulRedirectURL: "/change-db",
	})

	dbm.Wg.Add(1)
	defer dbm.Wg.Wait()
	go dbm.GM.Flusher()

	// _, executableFile, _, _ := runtime.Caller(0) // get full path of this file
	if len(Build) == 0 {
		Build = "devdevdevdevdevdevdev"
	}
	// Bing flags for changing parameters of FIND
	flag.StringVar(&glb.RuntimeArgs.Port, "p", ":8003", "port to bind")
	flag.StringVar(&glb.RuntimeArgs.Socket, "s", "", "unix socket")
	flag.StringVar(&glb.RuntimeArgs.ServerCRT, "crt", "", "location of ssl crt")
	flag.StringVar(&glb.RuntimeArgs.ServerKey, "key", "", "location of ssl key")
	flag.StringVar(&glb.RuntimeArgs.MqttServer, "mqtt", "", "ADDRESS:PORT of mosquitto server")
	flag.StringVar(&glb.RuntimeArgs.MqttAdmin, "mqttadmin", "", "admin to read all messages")
	flag.StringVar(&glb.RuntimeArgs.MqttAdminPassword, "mqttadminpass", "", "admin to read all messages")
	flag.StringVar(&glb.RuntimeArgs.MosquittoPID, "mosquitto", "", "mosquitto PID (`pgrep mosquitto`)")
	flag.StringVar(&glb.RuntimeArgs.Dump, "dump", "", "db with json format dump to data folder")
	flag.StringVar(&glb.RuntimeArgs.DumpCalc, "dumpcalc", "", "calculated db data with json format dump to data folder")
	flag.StringVar(&glb.RuntimeArgs.DumpRaw, "dumpraw", "", "raw db data with csv format dump to data folder")
	flag.StringVar(&glb.RuntimeArgs.Message, "message", "", "message to display to all users")
	flag.StringVar(&glb.RuntimeArgs.SourcePath, "data", "", "path to data folder")
	flag.StringVar(&glb.RuntimeArgs.ScikitPort, "scikit", "", "port for scikit-learn calculations")
	flag.StringVar(&glb.RuntimeArgs.ParticleFilterServer, "particlefilterServer", glb.RuntimeArgs.ParticleFilterServer, "ip:port of particleFilter grpc server ")
	flag.StringVar(&glb.RuntimeArgs.AdminAdd, "adminadd", "", "Add an admin user or change his password, foramt:<username>:<password>, e.g.:admin:admin")
	flag.BoolVar(&glb.RuntimeArgs.GaussianDist, "gaussian", false, "Use gaussian distribution instead of historgram")
	flag.BoolVar(&glb.RuntimeArgs.Debug, "debug", false, "run in debug mode")
	flag.IntVar(&glb.RuntimeArgs.MinRssOpt, "minrss", -100, "Select minimum rss; Any Rss lower than minRss will be ignored.")

	flag.CommandLine.Usage = func() {
		fmt.Println(`find (version ` + VersionNum + ` (` + Build[0:8] + `), built ` + BuildTime + `)
				Example: 'ParsinServer yourserver.com'
				Example: 'ParsinServer -p :8080 localhost:8080'
				Example (mosquitto): 'ParsinServer -mqtt 127.0.0.1:1883 -mqttadmin admin -mqttadminpass somepass -mosquitto ` + "`pgrep mosquitto`" + `
				Options:`)
		flag.CommandLine.PrintDefaults()
	}
	flag.Parse()
	glb.RuntimeArgs.ExternalIP = flag.Arg(0)

	// Handle RuntimeArgs
	if glb.RuntimeArgs.ExternalIP == "" {
		glb.RuntimeArgs.ExternalIP = glb.GetLocalIP() + glb.RuntimeArgs.Port
	}

	if glb.RuntimeArgs.SourcePath == "" {
		glb.RuntimeArgs.SourcePath = path.Join(glb.RuntimeArgs.Cwd, "data")
	}
	fmt.Println(glb.RuntimeArgs.SourcePath)

	// Check whether all the MQTT variables are passed to initiate the MQTT routines
	if len(glb.RuntimeArgs.MqttServer) > 0 && len(glb.RuntimeArgs.MqttAdmin) > 0 && len(glb.RuntimeArgs.MosquittoPID) > 0 {
		glb.RuntimeArgs.Mqtt = true
		//routes.SetupMqtt()
	} else {
		if len(glb.RuntimeArgs.MqttServer) > 0 {
			glb.RuntimeArgs.Mqtt = true
			glb.RuntimeArgs.MqttExisting = true
			//routes.SetupMqtt()
		} else {
			glb.RuntimeArgs.Mqtt = false
		}
	}

	// Check whether random forests are used
	if len(glb.RuntimeArgs.ScikitPort) > 0 {
		glb.RuntimeArgs.Scikit = true
	}

	if glb.ParticleFilterEnabled {
		particlefilter.Connect2Server()
	}

	if glb.RuntimeArgs.Debug {
		fmt.Println("Running in debug mode")
	}

	// Check whether we are just dumping the database
	if len(glb.RuntimeArgs.Dump) > 0 {
		err := dbm.DumpFingerprints(strings.ToLower(glb.RuntimeArgs.Dump))
		if err == nil {
			fmt.Println("Successfully dumped.")
		} else {
			log.Fatal(err)
		}
		os.Exit(1)
	}
	if len(glb.RuntimeArgs.DumpCalc) > 0 {
		err := dbm.DumpCalculatedFingerprints(strings.ToLower(glb.RuntimeArgs.DumpCalc))
		if err == nil {
			fmt.Println("Successfully dumped.")
		} else {
			log.Fatal(err)
		}
		os.Exit(1)
	}
	if len(glb.RuntimeArgs.DumpRaw) > 0 {
		err := dbm.DumpRawFingerprints(strings.ToLower(glb.RuntimeArgs.DumpRaw))
		if err == nil {
			fmt.Println("Successfully dumped.")
		} else {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	// Useradded command
	// Check whether we are just dumping the database
	if len(glb.RuntimeArgs.AdminAdd) > 0 {
		addRequestSlice := strings.Split(strings.ToLower(glb.RuntimeArgs.AdminAdd), ":")
		//group := addRequestSlice[0]
		username := addRequestSlice[0]
		password := addRequestSlice[1]
		_, err := glb.SessionManager.RegisterNewUser(username, password, []string{"Admins"})
		if err == nil {
			fmt.Printf("Successfully new admin(username:%+v, password:%+v) was added Or new password was set.", username, password)
			adminList, _ := glb.SessionManager.ListAllUsers()
			fmt.Println("Current admin list:", *adminList)
		} else {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	//Set minRssOpt
	glb.MinRssiOpt = glb.RuntimeArgs.MinRssOpt

	// Check if there is a message from the admin
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.Cwd, "message.txt")); err == nil {
		messageByte, _ := ioutil.ReadFile(path.Join(glb.RuntimeArgs.Cwd, "message.txt"))
		glb.RuntimeArgs.Message = string(messageByte)
	}

	// Check whether SVM libraries are available
	cmdOut, _ := exec.Command("svm-scale", "").CombinedOutput()
	if len(cmdOut) == 0 {
		glb.RuntimeArgs.Svm = false
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
		glb.RuntimeArgs.Svm = true
	}

	// Setup Gin-Gonic
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	noNeedLogRoutes := []string{"data", "calcLevel"}
	logger := glb.Logger(noNeedLogRoutes...)
	engine.Use(logger, gin.Recovery())
	r := engine

	// Load templates
	r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*"))

	// Load static files (if they are not hosted by external service)
	r.Static("static/", path.Join(glb.RuntimeArgs.Cwd, "res/static/"))

	// Create cookie store to keep track of logged in user
	store := sessions.NewCookieStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// 404-page redirects to login
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "Please Choose your group first.",
		})
	})

	privateRoutes := r.Group("/", glb.SessionManager.AuthenticatedOnly())
	{
		privateRoutes.GET("/logout", glb.SessionManager.Logout)
		// Routes for logging in and viewing dashboards (pages.go)
		privateRoutes.GET("/", routes.Slash)
		privateRoutes.GET("/change-db", routes.SlashChangeDb)
		privateRoutes.POST("/change-db", routes.SlashChangeDbPOST)
		//privateRoutes.PUT("/mqtt", routes.PutMQTT) // Routes for MQTT (mqtt.go)
		privateRoutes.Static("data/", path.Join(glb.RuntimeArgs.Cwd, "data/")) // Load db files
		privateRoutes.GET("/status", routes.GetStatus)

		needToLoadSettings := privateRoutes.Group("/", routes.PreLoadSettings)
		{
			//Todo: Url must be same format to mention group name (now, group can be url param or be GET param)
			// Pages :
			needToLoadSettings.GET("/dashboard/:group", func(context *gin.Context) {
				r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*"))
				routes.SlashDashboard(context)
			})
			needToLoadSettings.GET("/explore/:group/:network/:location", routes.GetLocationMacs)

			needToLoadSettings.GET("/locationsmap/:group", routes.LocationsOnMap)
			needToLoadSettings.GET("/livemap/:group", routes.LiveLocationMap)
			//needToLoadSettings.GET("/userhistory/:group", routes.UserHistoryMap)
			needToLoadSettings.GET("/userhistory/:group", routes.UserHistoryMap)
			needToLoadSettings.GET("/testValidTracksMap/:group", func(context *gin.Context) {
				r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*")) // TODO: remove this for performance
				routes.TestValidTracksMap(context)
			})
			needToLoadSettings.GET("/fingerprintAmbiguity/:group", routes.FingerprintAmbiguity)
			needToLoadSettings.GET("/heatmap/:group", routes.Heatmap)
			needToLoadSettings.GET("/errorHeatmap/:group", routes.ErrorHeatMap)
			needToLoadSettings.GET("/uwbUserMap/:group", func(context *gin.Context) {
				r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*")) // TODO: remove this for performance
				routes.UWBUserMap(context)
			})
			//needToLoadSettings.GET("/graphForm/:group", routes.Graphform)
			needToLoadSettings.GET("/graphForm/:group", func(context *gin.Context) {
				r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*"))
				routes.Graphform(context)
			})
			needToLoadSettings.GET("/algorithmsCDF/:group", routes.AlgorithmsCDF)
			//needToLoadSettings.GET("/algorithmsCDF/:group", func(context *gin.Context) {
			//	r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*"))
			//	routes.AlgorithmsCDF(context)
			//})
			needToLoadSettings.GET("/testValidTracksDetails/:group", func(context *gin.Context) { //komeil: graph map
				r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*"))
				routes.TestValidTracksDetails(context)
			})
			needToLoadSettings.GET("/arbitraryLocations/:group", routes.ArbitraryLocations)
			//needToLoadSettings.GET("/InfrastructureDetails/:group", routes.InfrastructureDetails)
			needToLoadSettings.GET("/infrastructureDetails/:group", func(context *gin.Context) { //komeil: graph map
				r.LoadHTMLGlob(path.Join(glb.RuntimeArgs.Cwd, "res/templates/*"))
				routes.InfrastructureDetails(context)
			})

			// APIs:

			//needToLoadSettings.GET("/getFingerprint/", routes.GetFingerprint)
			needToLoadSettings.GET("/locations", routes.GetLocationList)
			needToLoadSettings.GET("/editloc", routes.EditLoc)
			needToLoadSettings.GET("/editlocBaseDB", routes.EditLocBaseDB)
			needToLoadSettings.GET("/editusername", routes.EditUserName)
			needToLoadSettings.DELETE("/location", routes.DeleteLocation)
			needToLoadSettings.DELETE("/locationBaseDB", routes.DeleteLocationBaseDB)
			needToLoadSettings.DELETE("/locations", routes.DeleteLocations)
			needToLoadSettings.DELETE("/locationsBaseDB", routes.DeleteLocationsBaseDB)

			needToLoadSettings.DELETE("/user", routes.DeleteUser)
			needToLoadSettings.DELETE("/database", routes.DeleteDatabase)
			needToLoadSettings.GET("/fingerprintLikeness", routes.FingerprintLikeness)

			needToLoadSettings.GET("/calculate", routes.Calculate)
			needToLoadSettings.GET("/cvresults", routes.CVResults)
			needToLoadSettings.GET("/calcLevel", routes.CalcCompletionLevel)

			needToLoadSettings.GET("/buildGroup", routes.BuildGroup)
			needToLoadSettings.PUT("/database", routes.MigrateDatabase)

			needToLoadSettings.GET("/lastFingerprint", routes.GetLastFingerprint)
			needToLoadSettings.GET("/reformDB", routes.ReformDB)
			needToLoadSettings.GET("/macfilterForm/:group", routes.Macfilterform)

			needToLoadSettings.GET("/getMostSeenMacs", routes.GetMostSeenMacsAPI)
			needToLoadSettings.POST("/setFilterMacs", routes.SetFilterMacs)
			needToLoadSettings.GET("/getFilterMacs", routes.GetFilterMacs)
			needToLoadSettings.GET("/getUniqueMacs", routes.GetUniqueMacs)

			needToLoadSettings.POST("/addNodeToGraph", routes.AddNodeToGraph) // komeil: set and get for graph
			needToLoadSettings.GET("/getGraph", routes.GetGraph)
			needToLoadSettings.POST("/addEdgeToGraph", routes.AddEdgeToGraph)
			needToLoadSettings.GET("/saveEdgesToDB", routes.SaveEdgesToDB)
			needToLoadSettings.POST("/removeEdgesOrVertices", routes.RemoveEdgesOrVertices)
			needToLoadSettings.GET("/deleteWholeGraph", routes.DeleteWholeGraph)
			needToLoadSettings.GET("/getGraphNodeAdjacentFPs", routes.GetGraphNodeAdjacentFPs)

			needToLoadSettings.PUT("/chooseMap", routes.ChooseMap) // komeil: choose a map for group
			//Arbitrary locations
			needToLoadSettings.POST("/addArbitLocations", routes.AddArbitLocations)
			needToLoadSettings.POST("/delArbitLocations", routes.DelArbitLocations)

			//Infrastructure :
			needToLoadSettings.POST("/addChangeTransmitter", routes.AddChangeTransmitter)
			needToLoadSettings.POST("/delTransmitter", routes.DelTransmitter)

			//needToLoadSettings.GET("/getArbitLocations", routes.GetArbitLocations)
			needToLoadSettings.DELETE("/clearConfigData", routes.ClearConfigData)
			needToLoadSettings.POST("/setKnnConfig", routes.SetKnnConfig)
			needToLoadSettings.POST("/setGroupOtherConfig", routes.SetGroupOtherConfig)
		}
	}
	r.GET("/location", routes.GetUserLocations)
	r.GET("/getFingerprint/", routes.GetFingerprint)
	r.GET("/getRPFDetails/", routes.GetRPFDetails)
	r.GET("/getRPFDetailsMapDots/", routes.GetRPFDetailsMapDots)

	r.GET("/editMac", routes.EditMac)
	r.GET("/reloadDB", routes.ReloadDB)
	r.GET("/setRelocateFPLocState", routes.SetRelocateFPLocStateAPI)
	r.GET("/getRelocateFPLocState", routes.GetRelocateFPLocStateAPI)
	r.GET("/getRSSData", routes.GetRSSDataAPI)
	r.GET("/getMapDetails", routes.GetMapDetails)
	r.GET("/getArbitLocations", routes.GetArbitLocations)
	r.GET("/getInfrastructureDetails", routes.GetInfrastructureDetails)

	r.GET("/getTestValidTracks", routes.GetTestValidTracks) // deprecated
	r.DELETE("/delTestValidTracks", routes.DelTestValidTracks)
	r.GET("/getTestValidTracksDetails", routes.GetTestValidTracksDetails)
	r.GET("/getTestErrorAlgoAccuracy", routes.GetTestErrorAlgoAccuracy)
	r.DELETE("/clearTestValidTrueLocation", routes.ClearTestValidTrueLocation)

	r.POST("/uploadTrueLocationLog", routes.UploadTrueLocationLog)
	r.POST("/learn", algorithms.LearnFingerprintPOST)
	r.POST("/bulklearn", algorithms.BulkLearnFingerprintPOST)
	r.POST("/track", algorithms.TrackFingerprintPOST)

	r.DELETE("/delResults", routes.DelResults)

	// Authentication
	auth := r.Group("/")
	{
		auth.GET("/login", routes.SlashLogin)
		auth.POST("/login", routes.SlashLoginPOST)
	}

	// Load and display the logo
	dat, _ := ioutil.ReadFile("./res/static/logo.txt")
	fmt.Println(string(dat))

	// Check whether user is providing certificates
	if glb.RuntimeArgs.Socket != "" {
		r.RunUnix(glb.RuntimeArgs.Socket)
	} else if glb.RuntimeArgs.ServerCRT != "" && glb.RuntimeArgs.ServerKey != "" {
		fmt.Println(`(version ` + VersionNum + ` build ` + Build[0:8] + `) is up and running on https://` + glb.RuntimeArgs.ExternalIP)
		fmt.Println("-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----")
		r.RunTLS(glb.RuntimeArgs.Port, glb.RuntimeArgs.ServerCRT, glb.RuntimeArgs.ServerKey)
	} else {
		fmt.Println(`(version ` + VersionNum + ` build ` + Build[0:8] + `) is up and running on http://` + glb.RuntimeArgs.ExternalIP)
		fmt.Println("-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----")
		r.Run(glb.RuntimeArgs.Port)
	}

}
