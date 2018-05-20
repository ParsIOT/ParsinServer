package dbm

import (
	"ParsinServer1/dbm"
	"ParsinServer/glb"
	"errors"
)

func AddArbitLocations(groupName string, addLocList []string) error {
	shrPrf := dbm.GetSharedPrf(groupName)
	allLocations := shrPrf.ArbitLocations
	tempLocationList := []string{}

	tempLocationList = append(tempLocationList,allLocations...)

	for _,loc := range addLocList{
		if !glb.StringInSlice(loc,tempLocationList){
			tempLocationList = append(tempLocationList, loc)
		}
	}

	err := dbm.SetSharedPrf(groupName,"ArbitLocations",tempLocationList)

	if err != nil{
		glb.Error.Println("Can't add Arbitrary locations")
		return errors.New("Can't add Arbitrary locations")
	}else {
		return nil
	}
}

func DelArbitLocations(groupName string, delLocList []string) error {
	shrPrf := dbm.GetSharedPrf(groupName)
	allLocations := shrPrf.ArbitLocations

	tempLocationList := []string{}

	for _,loc := range allLocations{
		if !glb.StringInSlice(loc,delLocList){
			tempLocationList = append(tempLocationList, loc)
		}
	}

	err := dbm.SetSharedPrf(groupName,"ArbitLocations",tempLocationList)

	if err != nil{
		glb.Error.Println("Can't add Arbitrary locations")
		return errors.New("Can't add Arbitrary locations")
	}else {
		return nil
	}
}

func GetArbitLocations(groupName string) []string{
	return dbm.GetSharedPrf(groupName).ArbitLocations
}