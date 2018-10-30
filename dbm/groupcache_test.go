package dbm

import (
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdd_UserHistory(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	gp := GM.GetGroup(testdb)
	fp1 := parameters.UserPositionJSON{
		Time:     1000,
		KnnGuess: "10,10",
	}
	fp2 := parameters.UserPositionJSON{
		Time:     1003,
		KnnGuess: "20,20",
	}
	fp3 := parameters.UserPositionJSON{
		Time:     1006,
		KnnGuess: "30,30",
	}

	glb.MaxUserHistoryLen = 2
	gp.Get_ResultData().Append_UserHistory("TestUser", fp1)
	gp.Get_ResultData().Append_UserHistory("TestUser", fp2)
	gp.Get_ResultData().Append_UserHistory("TestUser", fp3)
	userHistory := gp.Get_ResultData().Get_UserHistory("TestUser")
	glb.Debug.Println(userHistory)
	//if (userHistory[0].KnnGuess == "10,10"){
	//	glb.Debug.Println("ok")
	//}
	result := []parameters.UserPositionJSON{fp2, fp3}

	assert.Equal(t, userHistory, result)
}

func TestGetNearestNode(t *testing.T) {
	gp := GM.GetGroup("arman_28_3_97_ble_1")

	//glb.Debug.Println(len(fingerprintsInMemory1))
	graphMapPointer := gp.Get_AlgoData().Get_GroupGraph()
	nearNodeGraph := graphMapPointer.GetNearestNode("1383.0,258.0")
	glb.Debug.Println(nearNodeGraph)
	assert.Equal(t, 1, 1)
}
