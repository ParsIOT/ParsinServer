// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// pages.go contains the functions that handle the web page views

package routes

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"ParsinServer/glb"
	"ParsinServer/dbm"
	"sort"
	"ParsinServer/dbm/parameters"

)



// slash returns the dashboard, if logged in, else it redirects to login page.
func Slash(c *gin.Context) {
	var groupName string
	loginGroup := sessions.Default(c)
	groupCookie := loginGroup.Get("group")
	if groupCookie == nil {
		c.Redirect(302, "/change-db")
	} else {
		groupName = groupCookie.(string)
		c.Redirect(302, "/dashboard/"+groupName)
	}
}

// slashLogin shows login page
func SlashLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", gin.H{})
}

// slashLoginPOST handles a POST login and returns dashboard if successful, else login.
func SlashLoginPOST(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if ok := glb.SessionManager.IsUserValid(username, password); ok {
		session := sessions.Default(c)
		session.Set("user", glb.SessionManager.GetUser(username))
		session.Save()
		cookieGroup := session.Get("group")
		if cookieGroup == nil {
			c.Redirect(http.StatusFound, glb.SessionManager.LoginSuccessfulRedirectURL)
		} else {
			c.Redirect(http.StatusFound, "/dashboard/"+cookieGroup.(string))
		}
	} else {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{"ErrorMessage": "Invalid Username or Password",})
	}
}

// slashChangeDbPOST handles a POST login and returns dashboard if successful, else login.
func SlashChangeDbPOST(c *gin.Context) {
	loginGroup := sessions.Default(c)
	groupName := strings.ToLower(c.PostForm("group"))
	if _, err := os.Stat(path.Join("data", groupName+".db")); err == nil {
		loginGroup.Set("group", groupName)
		loginGroup.Save()
		c.Redirect(302, "/dashboard/"+groupName)
	} else {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "Incorrect login.",
		})
	}
}

// slashChangeDb handles a selecting a group
func SlashChangeDb(c *gin.Context) {
	var groupName string
	loginGroup := sessions.Default(c)
	groupCookie := loginGroup.Get("group")
	groupName = c.DefaultQuery("group", "none")
	errorMessage := c.DefaultQuery("error", "none")

	if errorMessage != "none" {
		if errorMessage == "groupNotExists"{
			c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
				"ErrorMessage": "There is no group with this name.",
			})
			return
		}
	}

	if groupCookie == nil {
		if groupName == "none" {
			c.HTML(http.StatusOK, "changedb.tmpl", gin.H{})
		} else {
			loginGroup.Set("group", groupName)
			loginGroup.Save()
			c.Redirect(302, "/dashboard/"+groupName)
		}
	} else {
		groupName = groupCookie.(string)
		fmt.Println(groupName)
		loginGroup.Delete("group")
		loginGroup.Save()
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"Message": "Now you can change your group",
		})
	}
}

// slashDashboard displays the dashboard
func SlashDashboard(c *gin.Context) {
	// skipUsers := c.DefaultQuery("skip", "")
	// skipAllUsers := false
	// if len(skipUsers) > 0 {
	// 	skipAllUsers = true
	// }

	filterUser := c.DefaultQuery("user", "")
	filterUsers := c.DefaultQuery("users", "")
	filterUserMap := make(map[string]bool)
	if len(filterUser) > 0 {
		u := strings.Replace(strings.TrimSpace(filterUser), ":", "", -1)
		filterUserMap[u] = true
	}
	if len(filterUsers) > 0 {
		for _, user := range strings.Split(filterUsers, ",") {
			u := strings.Replace(strings.TrimSpace(user), ":", "", -1)
			filterUserMap[u] = true
		}
	}
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}

	//ps, _ := dbm.OpenParameters(group)
	//gp := dbm.GM.GetGroup(groupName)

	var users []string
	for user := range filterUserMap {
		users = append(users, user)
	}
	people := make(map[string]parameters.UserPositionJSON)
	//if len(users) == 0 {
	//	people = GetCurrentPositionOfAllUsers(groupName)
	//} else {
	//	for _, user := range users {
	//		people[user] = GetCurrentPositionOfUser(groupName, user)
	//	}
	//}
	//glb.Debug.Println("3333333333")


	type DashboardData struct {
		Networks         []string
		Locations        map[string][]string
		LocationAccuracy map[string]int
		LocationCount    map[string]int
		Mixin            map[string]float64
		VarabilityCutoff map[string]float64
		Users            map[string]parameters.UserPositionJSON
	}
	var dash DashboardData
	dash.Networks = []string{}
	dash.Locations = make(map[string][]string)
	dash.LocationAccuracy = make(map[string]int)
	dash.LocationCount = make(map[string]int)
	dash.Mixin = make(map[string]float64)
	dash.VarabilityCutoff = make(map[string]float64)


	kRange := dbm.GetSharedPrf(groupName).KnnKRange
	knnMinCRssRange := dbm.GetSharedPrf(groupName).KnnMinCRssRange

	gp := dbm.GM.GetGroup(groupName)
	md := gp.Get_MiddleData_Val()

	knnAlgo := gp.Get_AlgoData().Get_KnnFPs()
	bestK := knnAlgo.K
	bestMinClusterRss := knnAlgo.MinClusterRss
	maxMovement := dbm.GetSharedPrf(groupName).MaxMovement

	for n := range md.NetworkLocs {
		//dash.Mixin[n] = gp.Get_Priors()[n].Special["MixIn"]
		//dash.VarabilityCutoff[n] = gp.Get_Priors()[n].Special["VarabilityCutoff"]
		dash.Networks = append(dash.Networks, n)
		dash.Locations[n] = []string{}
		uniqueLocs := md.UniqueLocs
		sort.Sort(sort.StringSlice(uniqueLocs))

		for _,loc := range uniqueLocs {
			dash.Locations[n] = append(dash.Locations[n], loc)

			//dash.LocationAccuracy[loc] = gp.Get_Results()[n].Accuracy[loc]
			//glb.Debug.Println(ps.BayesResults[n].TotalLocations[loc])
			dash.LocationCount = md.LocCount
		}
		totalError := 0
		algoAccuracy := gp.Get_ResultData().Get_AlgoLocAccuracy()
		for loc, accuracy := range algoAccuracy["knn"] {
			dash.LocationAccuracy[loc] = accuracy
			totalError += accuracy
		}
		dash.LocationAccuracy["all"] = totalError
	}
	//glb.Debug.Println(dash)
	mapNamesList := glb.ListMaps()
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"Message":           glb.RuntimeArgs.Message,
		"Group":             groupName,
		"Dash":              dash,
		"Users":             people,
		"kRange":            kRange,
		"knnMinCRssRange":   knnMinCRssRange,
		"bestK":             bestK,
		"bestMinClusterRss": bestMinClusterRss,
		"maxMovement":       maxMovement,
		"mapNamesList":		 mapNamesList,
	})
}

// slashDashboard displays the Users on a map
func LiveLocationMap(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	MapName := dbm.GetSharedPrf(groupName).MapName
	MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
	//MapWidth := dbm.GetSharedPrf(groupName).MapWidth
	//MapHeight := dbm.GetSharedPrf(groupName).MapHeight
	MapPath := path.Join(glb.RuntimeArgs.MapPath,MapName)
	//MapPathCorrected := filepath.FromSlash(MapPath)
	glb.Debug.Println("final MapPath: ", MapPath)
	glb.Debug.Println("final MapWidth: ", MapDimensions[0])
	glb.Debug.Println("final MapHeight: ", MapDimensions[1])
	c.HTML(http.StatusOK, "live_location_map.tmpl", gin.H{
		"Group": groupName,
		"MapPath": MapPath,
		"MapWidth":MapDimensions[0],
		"MapHeight":MapDimensions[1],
	})
}

func LocationsOnMap(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	MapName := dbm.GetSharedPrf(groupName).MapName
	MapPath := path.Join(glb.RuntimeArgs.MapPath,MapName)
	MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
	c.HTML(http.StatusOK, "locations_map.tmpl", gin.H{
		"Group": groupName,
		"MapPath": MapPath,
		"MapWidth":MapDimensions[0],
		"MapHeight":MapDimensions[1],
	})
}

func ArbitraryLocations(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	MapName := dbm.GetSharedPrf(groupName).MapName
	MapPath := path.Join(glb.RuntimeArgs.MapPath,MapName)
	MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
	c.HTML(http.StatusOK, "arbitrary_locations.tmpl", gin.H{
		"Group": groupName,
		"MapPath": MapPath,
		"MapWidth":MapDimensions[0],
		"MapHeight":MapDimensions[1],
	})
}

// slashDashboard displays the Users on a map
func UserHistoryMap(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	MapName := dbm.GetSharedPrf(groupName).MapName
	MapPath := path.Join(glb.RuntimeArgs.MapPath,MapName)
	MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
	c.HTML(http.StatusOK, "trace_history_map.tmpl", gin.H{
		"Group": groupName,
		"MapPath": MapPath,
		"MapWidth":MapDimensions[0],
		"MapHeight":MapDimensions[1],
	})
}

func FingerprintAmbiguity(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	MapName := dbm.GetSharedPrf(groupName).MapName
	MapPath := path.Join(glb.RuntimeArgs.MapPath,MapName)
	MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
	c.HTML(http.StatusOK, "fingerprint_ambiguity_map.tmpl", gin.H{
		"Group": groupName,
		"MapPath": MapPath,
		"MapWidth":MapDimensions[0],
		"MapHeight":MapDimensions[1],
	})
}

func Heatmap(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	MapName := dbm.GetSharedPrf(groupName).MapName
	MapPath := path.Join(glb.RuntimeArgs.MapPath,MapName)
	MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
	c.HTML(http.StatusOK, "heatmap.tmpl", gin.H{
		"Group": groupName,
		"MapPath": MapPath,
		"MapWidth":MapDimensions[0],
		"MapHeight":MapDimensions[1],
	})
}

// slash Location returns location (to be deprecated)
//func SlashLocation(c *gin.Context) {
//	groupName := c.Param("group")
//	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
//		c.JSON(http.StatusOK, gin.H{"success": "false", "message": "First download the app or CLI program to insert some fingerprints."})
//		return
//	}
//	user := c.Param("user")
//	userJSON := GetCurrentPositionOfUser(groupName, user)
//	c.JSON(http.StatusOK, userJSON)
//}

// slashExplore returns a chart of the data
//// todo: Use it
//func SlashExplore(c *gin.Context) {
//	groupName := c.Param("group")
//	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
//		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
//			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
//		})
//		return
//	}
//	network := c.Param("network")
//	location := c.Param("location")
//	//ps, _ := dbm.OpenParameters(group)
//	gp := dbm.GM.GetGroup(groupName)
//
//	// TODO: check if network and location exists in the ps, if not return 404
//	datas := []template.JS{}
//	names := []template.JS{}
//	indexNames := []template.JS{}
//	// Sort locations
//	macs := []string{}
//	for m := range gp.Get_Priors()[network].P[location] {
//		if float64(gp.Get_MacVariability()[m]) > gp.Get_Priors()[network].Special["VarabilityCutoff"] {
//			macs = append(macs, m)
//		}
//	}
//	sort.Strings(macs)
//	it := 0
//	for _, m := range macs {
//		n := gp.Get_Priors()[network].P[location][m]
//		names = append(names, template.JS(string(m)))
//		jsonByte, _ := json.Marshal(n)
//		datas = append(datas, template.JS(string(jsonByte)))
//		indexNames = append(indexNames, template.JS(strconv.Itoa(it)))
//		it++
//	}
//	rsiRange, _ := json.Marshal(glb.RssiRange)
//	c.HTML(http.StatusOK, "plot.tmpl", gin.H{
//		"RssiRange":  template.JS(string(rsiRange)),
//		"Datas":      datas,
//		"Names":      names,
//		"IndexNames": indexNames,
//	})
//}

//// slashExplore returns a chart of the data (canvas.js)
//func SlashExplore2(c *gin.Context) {
//	groupName := c.Param("group")
//	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
//		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
//			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
//		})
//		return
//	}
//
//	network := c.Param("network")
//	location := c.Param("location")
//	//ps, _ := dbm.OpenParameters(group)
//	gp := dbm.GM.GetGroup(groupName)
//
//	fmt.Println(gp.Get_UniqueLocs())
//	lookUpLocation := false
//
//	for _, loc := range gp.Get_UniqueLocs() {
//		if location == loc {
//			lookUpLocation = true
//		}
//	}
//
//	type macDatum struct {
//		Name   string    `json:"name"`
//		Points []float32 `json:"data"`
//	}
//
//	type macDatas struct {
//		Macs []macDatum `json:"macs"`
//	}
//
//	var data macDatas
//	data.Macs = []macDatum{}
//
//	if lookUpLocation {
//		// Sort locations
//		macs := []string{}
//		for m := range gp.Get_Priors()[network].P[location] {
//			if float64(gp.Get_MacVariability()[m]) > gp.Get_Priors()[network].Special["VarabilityCutoff"] {
//				macs = append(macs, m)
//			}
//		}
//		sort.Strings(macs)
//
//		for _, m := range macs {
//			n := gp.Get_Priors()[network].P[location][m]
//			data.Macs = append(data.Macs, macDatum{Name: m, Points: n})
//		}
//	} else {
//		m := location
//		for loc := range gp.Get_Priors()[network].P {
//			n := gp.Get_Priors()[network].P[loc][m]
//			data.Macs = append(data.Macs, macDatum{Name: strings.Replace(loc, " ", "%20", -1), Points: n})
//		}
//	}
//
//	c.HTML(http.StatusOK, "plot2.tmpl", gin.H{
//		"Data":    data,
//		"Rssi":    glb.RssiRange,
//		"Title":   groupName + "/" + network + "/" + location,
//		"Group":   strings.Replace(groupName, " ", "%20", -1),
//		"Network": strings.Replace(network, " ", "%20", -1),
//		"Legend":  !lookUpLocation,
//	})
//}

//// slashPie returns a Pie chart
//func SlashPie(c *gin.Context) {
//	groupName := c.Param("group")
//	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
//		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
//			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
//		})
//		return
//	}
//
//	network := c.Param("network")
//	location := c.Param("location")
//	gp := dbm.GM.GetGroup(groupName)
//
//	//ps, _ := dbm.OpenParameters(group)
//	vals := []int{}
//	names := []string{}
//	fmt.Println(gp.Get_Results()[network].Guess[location])
//	for guessloc, val := range gp.Get_Results()[network].Guess[location] {
//		names = append(names, guessloc)
//		vals = append(vals, val)
//	}
//	namesJSON, _ := json.Marshal(names)
//	valsJSON, _ := json.Marshal(vals)
//	c.HTML(http.StatusOK, "pie.tmpl", gin.H{
//		"Names": template.JS(namesJSON),
//		"Vals":  template.JS(valsJSON),
//	})
//}

// show mac filter form
func Macfilterform(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	c.HTML(http.StatusOK, "mac_filter.tmpl", gin.H{
		"Group": groupName,
	})
}


// show graph form
func Graphform(c *gin.Context) {
	groupName := c.Param("group")
	if _, err := os.Stat(path.Join(glb.RuntimeArgs.SourcePath, groupName+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	MapName := dbm.GetSharedPrf(groupName).MapName
	MapPath := path.Join(glb.RuntimeArgs.MapPath,MapName)
	MapDimensions := dbm.GetSharedPrf(groupName).MapDimensions
	c.HTML(http.StatusOK, "graph.tmpl", gin.H{
		"Group": groupName,
		"MapPath": MapPath,
		"MapWidth":MapDimensions[0],
		"MapHeight":MapDimensions[1],
	})
}


