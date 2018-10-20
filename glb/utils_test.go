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
func TestSortIntKeyDictByIntVal(t *testing.T) {
	newMap := make(map[int]int)
	newMap[1] = 10
	newMap[3] = 30
	newMap[2] = 20
	sortedKey := SortIntKeyDictByIntVal(newMap)
	assert.Equal(t, []int{1, 2, 3}, sortedKey)
}

func TestGetGraphSlicesRangeRecursive(t *testing.T) {
	beginSlice := []float64{1, 1, 1, 1, 1, 1, 1}
	endSlice := []float64{10, 10, 10, 10, 3, 2, 1}
	rangeSlices := GetGraphSlicesRangeRecursive(beginSlice, endSlice)
	//Debug.Println(rangeSlices)
	Debug.Println(len(rangeSlices))

	assert.Equal(t, true, false)

}