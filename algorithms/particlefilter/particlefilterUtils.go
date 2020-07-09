package particlefilter

import (
	pb "ParsinServer/algorithms/particlefilter/particlefilterclasses"
)

func GetMapGraph(floatGraphMap [][][]float32) pb.Graph {
	newGraphLines := []*pb.Line{}
	for _, floatLine := range floatGraphMap {
		newLineDots := []*pb.Dot{}
		// Create new line
		for _, floatDot := range floatLine {
			// Swap X and Y

			swapFloatDot := []float32{floatDot[1], floatDot[0]}
			// Create new Dot
			newDot := pb.Dot{XY: swapFloatDot}
			newLineDots = append(newLineDots, &newDot)
		}
		newLine := pb.Line{Dots: newLineDots}

		// Append line to lines
		newGraphLines = append(newGraphLines, &newLine)
	}
	return pb.Graph{Lines: newGraphLines}
}
