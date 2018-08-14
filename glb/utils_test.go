package glb

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

//func TestListMaps(t *testing.T) {
//	files :=[]string{"map1.png","map2.png","map3.png"}
//	resultFiles := ListMaps("C:/Users/komeil/go/src/ParsinServer/res/static/map")
//	assert.Equal(t,files , resultFiles)
//}

func TestListMaps(t *testing.T) {
	sortedList := []int64{1, 3, 4}
	newItem := int64(2)

	sortedList = SortedInsert(sortedList, newItem)

	assert.Equal(t, sortedList, []int64{1, 2, 3, 4})
}
