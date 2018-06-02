package algorithms

import (
	"ParsinServer/glb"
	"strings"
	"strconv"
)

func HistoryEffect(currentUserPos glb.UserPositionJSON, userHistory []glb.UserPositionJSON) string {

	glb.Debug.Println(currentUserPos)
	glb.Debug.Println(userHistory)

	locHistory := []string{}
	tsHistory := []int64{} // timestamps
	for _, userPos := range userHistory {
		locHistory = append(locHistory, userPos.KnnGuess)
		tsHistory = append(tsHistory, userPos.Time)
	}
	locHistory = append(locHistory, currentUserPos.KnnGuess)
	tsHistory = append(tsHistory, currentUserPos.Time)

	resX := float64(0)
	resY := float64(0)
	sumFactor := float64(0)

	//for i,factor := range glb.UserHistoryEffectFactors{
	//	if i==len(locHistory){
	//		x_y := strings.Split(currentLoc, ",")
	//		if len(x_y) < 2 {
	//			//err := errors.New("Location names aren't in the format of x,y")
	//			glb.Error.Println("Location names aren't in the format of x,y")
	//		}
	//		curLocXstr := x_y[0]
	//		curLocYstr := x_y[1]
	//		curLocX, _ := strconv.ParseFloat(curLocXstr,64)
	//		curLocY, _ := strconv.ParseFloat(curLocYstr,64)
	//
	//		resX += curLocX*factor
	//		resY += curLocY*factor
	//	}else{
	//		tempLoc := locHistory[i]
	//		tempx_y := strings.Split(tempLoc, ",")
	//		if len(tempx_y) < 2 {
	//			glb.Error.Println("Location names aren't in the format of x,y")
	//		}
	//		tempLocXstr := tempx_y[0]
	//		tempLocYstr := tempx_y[1]
	//		tempLocX, _ := strconv.ParseFloat(tempLocXstr,64)
	//		tempLocY, _ := strconv.ParseFloat(tempLocYstr,64)
	//
	//		resX += tempLocX*factor
	//		resY += tempLocY*factor
	//	}
	//
	//	sumFactor += factor
	//}
	lastFPTime := tsHistory[len(locHistory)-1]

	gaussModel := NewGaussian(0, glb.UserHistoryGaussVariance)
	for i, loc := range locHistory {
		var factor float64
		if i == len(locHistory)-1 {
			glb.Debug.Println(loc)
			x_y := strings.Split(loc, ",")
			if len(x_y) < 2 {
				//err := errors.New("Location names aren't in the format of x,y")
				glb.Error.Println("Location names aren't in the format of x,y")
			}
			curLocXstr := x_y[0]
			curLocYstr := x_y[1]
			curLocX, _ := strconv.ParseFloat(curLocXstr, 64)
			curLocY, _ := strconv.ParseFloat(curLocYstr, 64)

			factor = gaussModel.Pdf(0)
			resX += curLocX * factor
			resY += curLocY * factor

		} else {
			glb.Debug.Println(loc)
			tempx_y := strings.Split(loc, ",")
			if len(tempx_y) < 2 {
				glb.Error.Println("Location names aren't in the format of x,y")
			}
			tempLocXstr := tempx_y[0]
			tempLocYstr := tempx_y[1]
			tempLocX, _ := strconv.ParseFloat(tempLocXstr, 64)
			tempLocY, _ := strconv.ParseFloat(tempLocYstr, 64)

			factor = gaussModel.Pdf(float64(lastFPTime-tsHistory[i]) / glb.UserHistoryTimeDelayFactor)

			glb.Debug.Println(factor)
			resX += tempLocX * factor
			resY += tempLocY * factor
		}

		sumFactor += factor
	}

	glb.Debug.Println(sumFactor)
	resX /= sumFactor
	resY /= sumFactor

	return glb.IntToString(int(resX)) + ".0," + glb.IntToString(int(resY)) + ".0"
}
