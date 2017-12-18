package geoutils

import (
	"errors"
	"math"
)

//PolygonArea get square of a polygon
func PolygonArea(locList []*Location) (square float64, err error) {
	vectorMinus := func(fromVector, toVector [2]float64) [2]float64 {
		return [2]float64{toVector[0] - fromVector[0], toVector[1] - fromVector[1]}
	}

	vectorCrossMultiple := func(firstVector, SecondVector [2]float64) float64 {
		return firstVector[0]*SecondVector[1] - firstVector[1]*SecondVector[0]
	}

	// check polygon has enough points
	if len(locList) < 4 {
		return -1, errors.New("Not enough Points")
	}

	// check polygon is closure
	if math.Abs(locList[0].EuclideanDistance(locList[len(locList)-1])-0.0) > 0.01 {
		return -1, errors.New("Polygon not closure")
	}

	// remove the closure point of polygon
	locList = locList[0 : len(locList)-1]
	var avgLat, avgLng, sumLat, sumLng float64
	for i := 0; i < len(locList); i++ {
		sumLat += locList[i].GetLatitude()
		sumLng += locList[i].GetLongitude()
	}
	avgLat, avgLng = sumLat/float64(len(locList)), sumLng/float64(len(locList))
	//1. choose first element as base point
	basePoint := NewLocation(avgLat, avgLng)
	baseVector := [2]float64{0.0, 0.0}

	//2. convert all loctList from lat/lng coordinate to x/y coordinate
	usePoint := make([]*Location, len(locList))
	useVector := make([][2]float64, len(locList))
	for i := 0; i < len(locList); i++ {
		usePoint[i] = locList[i]
		useVector[i] = [2]float64{
			basePoint.EuclideanDistance(NewLocation(usePoint[i].GetLatitude(), basePoint.GetLongitude())),
			basePoint.EuclideanDistance(NewLocation(basePoint.GetLatitude(), usePoint[i].GetLongitude()))}
		//based on the geographical latitude and longitude, the coordinate has been rotated.
		// The increasing latitude will located on the negative x-axis.
		if usePoint[i].GetLatitude() > basePoint.GetLatitude() {
			useVector[i][0] *= -1.0
		}
		if usePoint[i].GetLongitude() < basePoint.GetLongitude() {
			useVector[i][1] *= -1.0
		}
	}
	//3. cross the vector to get triangle area each by each
	var area = 0.0
	for i := 0; i < len(useVector)-1; i++ {
		firstVec := vectorMinus(useVector[i], baseVector)
		secondVec := vectorMinus(useVector[i+1], baseVector)
		tmpArea := vectorCrossMultiple(firstVec, secondVec)
		area += tmpArea
	}
	return area / 2, nil
}

//CircleArea get circle sqaure with given radius range
func CircleArea(radius float64) (square float64) {
	return math.Pi * radius * radius
}
