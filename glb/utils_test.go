package glb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//func TestListMaps(t *testing.T) {
//	files :=[]string{"map1.png","map2.png","map3.png"}
//	resultFiles := ListMaps("C:/Users/komeil/go/src/ParsinServer/res/static/map")
//	assert.Equal(t,files , resultFiles)
//}

func TestListMaps(t *testing.T) {
	sortedList := []int64{1, 3, 4}
	newItem := int64(2)

	sortedList = SortedInsertInt64(sortedList, newItem)

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

//func TestGetGraphSlicesRangeRecursive(t *testing.T) {
//	beginSlice := []float64{1, 1, 1, 1, 1, 1, 1}
//	endSlice := []float64{10, 10, 10, 10, 3, 2, 1}
//	rangeSlices := GetGraphSlicesRangeRecursive(beginSlice, endSlice)
//	//Debug.Println(rangeSlices)
//	Debug.Println(len(rangeSlices))
//
//	assert.Equal(t, true, false)
//
//}
//
//func TestConvertStr2IntSlice(t *testing.T) {
//	sliceStr := "[1,2]"
//	slice := []int{1, 2}
//	sliceRes, err := ConvertStr2IntSlice(sliceStr)
//	Debug.Println(sliceStr)
//	if err != nil {
//		Error.Println(err)
//	}
//	Debug.Println(sliceRes)
//	assert.Equal(t, slice, sliceRes)
//}

//func TestConvertStr22DimIntSlice(t *testing.T) {
//	sliceStr := "[[1,2],[3,4],[5,6,7]]"
//	//slice := [][]int{{1,2},{3,4},{5,6,7}}
//
//	res := [][]int{}
//	if err := json.Unmarshal([]byte(sliceStr), &res); err != nil {
//		panic(err)
//	}
//	Debug.Println(res)
//
//	sliceRes, err := ConvertStr22DimIntSlice(sliceStr)
//	Debug.Println(sliceStr)
//	if err != nil {
//		Error.Println(err)
//	}
//	Debug.Println(sliceRes)
//	assert.Equal(t, true, sliceRes)
//}

func TestMakeRange(t *testing.T) {
	r1 := MakeRange(1, 13, 5)
	r2 := MakeRange(13, 1, 5)
	r3 := MakeRange(-13, -1, 5)
	r4 := MakeRange(-1, -13, 5)

	assert.Equal(t, r1, []int{1, 6, 11})
	assert.Equal(t, r2, []int{13, 8, 3})
	assert.Equal(t, r3, []int{-13, -8, -3})
	assert.Equal(t, r4, []int{-1, -6, -11})
}

func TestMakeRangeFloat(t *testing.T) {
	r1 := MakeRangeFloat(0.1, 2.1, 0.1)
	r2 := MakeRangeFloat(2.1, 0.1, 0.1)
	r3 := MakeRangeFloat(-0.1, -2.1, 0.1)
	r4 := MakeRangeFloat(-2.1, -0.1, 0.1)

	assert.Equal(t, r1, []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, 2, 2.1})
	assert.Equal(t, r2, []float64{2.1, 2, 1.9, 1.8, 1.7, 1.6, 1.5, 1.4, 1.3, 1.2, 1.1, 1, 0.9, 0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1})
	assert.Equal(t, r3, []float64{-0.1, -0.2, -0.3, -0.4, -0.5, -0.6, -0.7, -0.8, -0.9, -1, -1.1, -1.2, -1.3, -1.4, -1.5, -1.6, -1.7, -1.8, -1.9, -2, -2.1})
	assert.Equal(t, r4, []float64{-2.1, -2, -1.9, -1.8, -1.7, -1.6, -1.5, -1.4, -1.3, -1.2, -1.1, -1, -0.9, -0.8, -0.7, -0.6, -0.5, -0.4, -0.3, -0.2, -0.1})
}

func TestGetFloatPrecision(t *testing.T) {
	r1 := GetFloatPrecision(1.0)
	r2 := GetFloatPrecision(1.1)
	r3 := GetFloatPrecision(-1.13)
	assert.Equal(t, 0, r1)
	assert.Equal(t, 1, r2)
	assert.Equal(t, 2, r3)
}

func TestConvert2DimStringSliceTo3DFloat32(t *testing.T) {
	slice := [][]string{{"1,2", "3,4"}, {"5,6", "7,8", "9,10"}}
	r := Convert2DimStringSliceTo3DFloat32(slice)
	Debug.Println(r)
	assert.Equal(t, 0, r)
}
