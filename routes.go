// Copyright 2015-2016 Zack Scholl. All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// routes.go contains the functions that handle the web page views

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

// slash returns the dashboard, if logged in, else it redirects to login page.
func slash(c *gin.Context) {
	var group string
	loginGroup := sessions.Default(c)
	groupCookie := loginGroup.Get("group")
	if groupCookie == nil {
		c.Redirect(302, "/change-db")
	} else {
		group = groupCookie.(string)
		c.Redirect(302, "/dashboard/"+group)
	}
}

// slashLogin shows login page
func slashLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", gin.H{})
}

// slashLoginPOST handles a POST login and returns dashboard if successful, else login.
func slashLoginPOST(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if ok := mySessionManager.IsUserValid(username, password); ok {
		session := sessions.Default(c)
		session.Set("user", mySessionManager.GetUser(username))
		session.Save()
		cookieGroup := session.Get("group")
		if cookieGroup == nil {
			c.Redirect(http.StatusFound, mySessionManager.LoginSuccessfulRedirectURL)
		} else {
			c.Redirect(http.StatusFound, "/dashboard/"+cookieGroup.(string))
		}
	} else {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{"ErrorMessage": "Invalid Username or Password",})
	}
}

// slashChangeDbPOST handles a POST login and returns dashboard if successful, else login.
func slashChangeDbPOST(c *gin.Context) {
	loginGroup := sessions.Default(c)
	group := strings.ToLower(c.PostForm("group"))
	if _, err := os.Stat(path.Join("data", group+".db")); err == nil {
		loginGroup.Set("group", group)
		loginGroup.Save()
		c.Redirect(302, "/dashboard/"+group)
	} else {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "Incorrect login.",
		})
	}
}

// slashChangeDb handles a selecting a group
func slashChangeDb(c *gin.Context) {
	var group string
	loginGroup := sessions.Default(c)
	groupCookie := loginGroup.Get("group")
	group = c.DefaultQuery("group", "noneasdf")
	if groupCookie == nil {
		if group == "noneasdf" {
			c.HTML(http.StatusOK, "changedb.tmpl", gin.H{})
		} else {
			loginGroup.Set("group", group)
			loginGroup.Save()
			c.Redirect(302, "/dashboard/"+group)
		}
	} else {
		group = groupCookie.(string)
		fmt.Println(group)
		loginGroup.Delete("group")
		loginGroup.Save()
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"Message": "Now you can change your group",
		})
	}
}

// slashDashboard displays the dashboard
func slashDashboard(c *gin.Context) {
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
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	ps, _ := openParameters(group)
	var users []string
	for user := range filterUserMap {
		users = append(users, user)
	}
	people := make(map[string]UserPositionJSON)
	if len(users) == 0 {
		people = getCurrentPositionOfAllUsers(group)
	} else {
		for _, user := range users {
			people[user] = getCurrentPositionOfUser(group, user)
		}
	}
	type DashboardData struct {
		Networks         []string
		Locations        map[string][]string
		LocationAccuracy map[string]int
		LocationCount    map[string]int
		Mixin            map[string]float64
		VarabilityCutoff map[string]float64
		Users            map[string]UserPositionJSON
	}
	var dash DashboardData
	dash.Networks = []string{}
	dash.Locations = make(map[string][]string)
	dash.LocationAccuracy = make(map[string]int)
	dash.LocationCount = make(map[string]int)
	dash.Mixin = make(map[string]float64)
	dash.VarabilityCutoff = make(map[string]float64)

	for n := range ps.NetworkLocs {
		dash.Mixin[n] = ps.Priors[n].Special["MixIn"]
		dash.VarabilityCutoff[n] = ps.Priors[n].Special["VarabilityCutoff"]
		dash.Networks = append(dash.Networks, n)
		dash.Locations[n] = []string{}
		for loc := range ps.NetworkLocs[n] {
			dash.Locations[n] = append(dash.Locations[n], loc)
			dash.LocationAccuracy[loc] = ps.Results[n].Accuracy[loc]
			dash.LocationCount[loc] = ps.Results[n].TotalLocations[loc]
		}
	}
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"Message": RuntimeArgs.Message,
		"Group":   group,
		"Dash":    dash,
		"Users":   people,
	})
}

// slashDashboard displays the Users on a map
func LiveLocationMap(c *gin.Context) {
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	c.HTML(http.StatusOK, "live_location_map.tmpl", gin.H{
		"Group": group,
	})
}

func LocationsOnMap(c *gin.Context) {
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	c.HTML(http.StatusOK, "locations_map.tmpl", gin.H{
		"Group": group,
	})
}

// slash Location returns location (to be deprecated)
func slashLocation(c *gin.Context) {
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{"success": "false", "message": "First download the app or CLI program to insert some fingerprints."})
		return
	}
	user := c.Param("user")
	userJSON := getCurrentPositionOfUser(group, user)
	c.JSON(http.StatusOK, userJSON)
}

// slashExplore returns a chart of the data
// todo: Use it
func slashExplore(c *gin.Context) {
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	network := c.Param("network")
	location := c.Param("location")
	ps, _ := openParameters(group)
	// TODO: check if network and location exists in the ps, if not return 404
	datas := []template.JS{}
	names := []template.JS{}
	indexNames := []template.JS{}
	// Sort locations
	macs := []string{}
	for m := range ps.Priors[network].P[location] {
		if float64(ps.MacVariability[m]) > ps.Priors[network].Special["VarabilityCutoff"] {
			macs = append(macs, m)
		}
	}
	sort.Strings(macs)
	it := 0
	for _, m := range macs {
		n := ps.Priors[network].P[location][m]
		names = append(names, template.JS(string(m)))
		jsonByte, _ := json.Marshal(n)
		datas = append(datas, template.JS(string(jsonByte)))
		indexNames = append(indexNames, template.JS(strconv.Itoa(it)))
		it++
	}
	rsiRange, _ := json.Marshal(RssiRange)
	c.HTML(http.StatusOK, "plot.tmpl", gin.H{
		"RssiRange":  template.JS(string(rsiRange)),
		"Datas":      datas,
		"Names":      names,
		"IndexNames": indexNames,
	})
}

// slashExplore returns a chart of the data (canvas.js)
func slashExplore2(c *gin.Context) {
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}

	network := c.Param("network")
	location := c.Param("location")
	ps, _ := openParameters(group)
	fmt.Println(ps.UniqueLocs)
	lookUpLocation := false

	for _, loc := range ps.UniqueLocs {
		if location == loc {
			lookUpLocation = true
		}
	}

	type macDatum struct {
		Name   string    `json:"name"`
		Points []float32 `json:"data"`
	}

	type macDatas struct {
		Macs []macDatum `json:"macs"`
	}

	var data macDatas
	data.Macs = []macDatum{}

	if lookUpLocation {
		// Sort locations
		macs := []string{}
		for m := range ps.Priors[network].P[location] {
			if float64(ps.MacVariability[m]) > ps.Priors[network].Special["VarabilityCutoff"] {
				macs = append(macs, m)
			}
		}
		sort.Strings(macs)

		for _, m := range macs {
			n := ps.Priors[network].P[location][m]
			data.Macs = append(data.Macs, macDatum{Name: m, Points: n})
		}
	} else {
		m := location
		for loc := range ps.Priors[network].P {
			n := ps.Priors[network].P[loc][m]
			data.Macs = append(data.Macs, macDatum{Name: strings.Replace(loc, " ", "%20", -1), Points: n})
		}
	}

	c.HTML(http.StatusOK, "plot2.tmpl", gin.H{
		"Data":    data,
		"Rssi":    RssiRange,
		"Title":   group + "/" + network + "/" + location,
		"Group":   strings.Replace(group, " ", "%20", -1),
		"Network": strings.Replace(network, " ", "%20", -1),
		"Legend":  !lookUpLocation,
	})
}

// slashPie returns a Pie chart
func slashPie(c *gin.Context) {
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}

	network := c.Param("network")
	location := c.Param("location")
	ps, _ := openParameters(group)
	vals := []int{}
	names := []string{}
	fmt.Println(ps.Results[network].Guess[location])
	for guessloc, val := range ps.Results[network].Guess[location] {
		names = append(names, guessloc)
		vals = append(vals, val)
	}
	namesJSON, _ := json.Marshal(names)
	valsJSON, _ := json.Marshal(vals)
	c.HTML(http.StatusOK, "pie.tmpl", gin.H{
		"Names": template.JS(namesJSON),
		"Vals":  template.JS(valsJSON),
	})
}

// show mac filter form
func macfilterform(c *gin.Context) {
	group := c.Param("group")
	if _, err := os.Stat(path.Join(RuntimeArgs.SourcePath, group+".db")); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "changedb.tmpl", gin.H{
			"ErrorMessage": "First download the app or CLI program to insert some fingerprints.",
		})
		return
	}
	c.HTML(http.StatusOK, "mac_filter.tmpl", gin.H{
		"Group": group,
	})
}


