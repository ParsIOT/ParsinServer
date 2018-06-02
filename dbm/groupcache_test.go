package dbm

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"ParsinServer/glb"
)

func TestAdd_UserHistory(t *testing.T) {
	testdb := gettestdbName()
	defer freedb(testdb)

	gp := GM.GetGroup(testdb)
	fp1 := glb.UserPositionJSON{
		Time:     1000,
		KnnGuess: "10,10",
	}
	fp2 := glb.UserPositionJSON{
		Time:     1003,
		KnnGuess: "20,20",
	}
	fp3 := glb.UserPositionJSON{
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
	result := []glb.UserPositionJSON{fp2, fp3}

	assert.Equal(t, userHistory, result)
}
