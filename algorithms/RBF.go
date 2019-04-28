package algorithms

import (
	"ParsinServer/dbm/parameters"
	"ParsinServer/glb"
	"math"
)

func CalculateDotRPF(dot string, graphMapPointer parameters.Graph) float64 {

	radiusThreshold := float64(200)
	routeVariance := radiusThreshold / 3

	rpfVal := float64(0)

	if graphMapPointer.IsEmpty() {
		glb.Error.Println("Empty graph: RBF can't be calculated ")
		return float64(0)
	}

	dotFloat := []float64{}
	dotFloatX, dotFloatY := glb.GetDotFromString(dot)
	dotFloat = append(dotFloat, dotFloatX)
	dotFloat = append(dotFloat, dotFloatY)

	ConnectedComponents := graphMapPointer.GetConnectedTreeComponents()
	for _, connectedComponent := range ConnectedComponents {
		connectedComponentDotDists := []float64{}
		for _, lineSegment := range connectedComponent {
			x1, y1 := glb.GetDotFromString(lineSegment[0])
			x2, y2 := glb.GetDotFromString(lineSegment[1])
			line := [][]float64{{x1, y1}, {x2, y2}}
			dist := glb.DistLineSegmentAndPoint(line, dotFloat)
			connectedComponentDotDists = append(connectedComponentDotDists, dist)
		}
		//glb.Debug.Println(connectedComponentDotDists)

		// find min distance to the route
		minDist := glb.MinFloat64Slice(connectedComponentDotDists)
		rpfVal += gaussianProbability(minDist, routeVariance)
		//glb.Debug.Println(rpfVal)

	}
	return rpfVal
}

func gaussianProbability(dist float64, routeVariance float64) float64 {
	return float64(1.0) / (routeVariance * math.Sqrt(2.0*math.Pi)) * math.Exp(-1.0/2.0*math.Pow(dist/routeVariance, 2.0))
}
