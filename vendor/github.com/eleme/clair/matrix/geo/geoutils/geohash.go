package geoutils

import (
	"errors"
	"fmt"
	structure "github.com/eleme/clair/matrix/structure"
	"math"
	"strconv"
)

const (
	base32 = "0123456789bcdefghjkmnpqrstuvwxyz"
)

var (
	encodeParamMap = map[int]encodePrecison{
		1:  encodePrecison{precision: 1, latBinBits: 2, lngBinBits: 3, vagueRange: 2500000.0},
		2:  encodePrecison{precision: 2, latBinBits: 5, lngBinBits: 5, vagueRange: 630000.0},
		3:  encodePrecison{precision: 3, latBinBits: 7, lngBinBits: 8, vagueRange: 78000.0},
		4:  encodePrecison{precision: 4, latBinBits: 10, lngBinBits: 10, vagueRange: 20000.0},
		5:  encodePrecison{precision: 5, latBinBits: 12, lngBinBits: 13, vagueRange: 2400.0},
		6:  encodePrecison{precision: 6, latBinBits: 15, lngBinBits: 15, vagueRange: 610.0},
		7:  encodePrecison{precision: 7, latBinBits: 17, lngBinBits: 18, vagueRange: 76.0},
		8:  encodePrecison{precision: 8, latBinBits: 20, lngBinBits: 20, vagueRange: 19.11},
		9:  encodePrecison{precision: 9, latBinBits: 22, lngBinBits: 23, vagueRange: 4.78},
		10: encodePrecison{precision: 10, latBinBits: 25, lngBinBits: 25, vagueRange: 0.59},
		11: encodePrecison{precision: 11, latBinBits: 27, lngBinBits: 28, vagueRange: 0.15},
		12: encodePrecison{precision: 12, latBinBits: 30, lngBinBits: 30, vagueRange: 0.01},
	}
	globalLatitudeRange  = Range{minVal: -90, maxVal: 90}
	globalLongitudeRange = Range{minVal: -180, maxVal: 180}
	base32Dict           = func() []int {
		baseDictResult := make([]int, 128)
		for i, v := range base32 {
			baseDictResult[v] = i
		}
		return baseDictResult
	}()
)

func checkValidBase32(base32Bytes []byte) bool {
	for _, base32Byte := range base32Bytes {
		if base32Byte != '0' && base32Dict[base32Byte] == 0 {
			return false
		}
	}
	return true
}

func binaryToBase32(binaryBytes []byte) ([]byte, error) {
	base32Len := (len(binaryBytes) + 4) / 5
	resultBase32Bytes := make([]byte, base32Len)
	tmpIntVal, err := strconv.ParseInt(string(binaryBytes), 2, 64)
	if err != nil {
		return nil, err
	}
	for i := 0; i < base32Len; i++ {
		resultBase32Bytes[base32Len-i-1] = base32[tmpIntVal%32]
		tmpIntVal /= 32
	}
	return resultBase32Bytes, nil
}

func base32ToBinary(base32Bytes []byte) ([]byte, error) {
	if !checkValidBase32(base32Bytes) {
		return []byte{}, errors.New("wrong format base32 bytes")
	}
	base32Len := len(base32Bytes)
	var tmpIntVal int64
	for i, v := range base32Bytes {
		tmpIntVal += int64(base32Dict[v]) * int64(math.Pow(32, float64(base32Len-i-1)))
	}

	binaryBytes := []byte(fmt.Sprintf("%b", tmpIntVal))
	autoPadding := make([]byte, base32Len*5-len(binaryBytes))
	for i := range autoPadding {
		autoPadding[i] = '0'
	}
	result := append(autoPadding, binaryBytes...)

	return result, nil
}

//Encode transform the latitude and longtitude to a geohash string
func Encode(latitude, longitude float64, precison int) (string, error) {

	findBinaryCode := func(r Range, val float64) (byte, Range) {
		midVal := r.GetMidVal()
		var binaryCode byte
		var nextRange Range
		if r.minVal <= val && val <= midVal {
			binaryCode = '0'
			nextRange = Range{minVal: r.minVal, maxVal: midVal}
		} else {
			binaryCode = '1'
			nextRange = Range{minVal: midVal, maxVal: r.maxVal}
		}
		return binaryCode, nextRange
	}
	if !globalLatitudeRange.CheckInRange(latitude) || !globalLongitudeRange.CheckInRange(longitude) {
		return "", errors.New("wrong params for latitude and longtitude")
	}

	precisonParam, ok := encodeParamMap[precison]
	if !ok {
		return "", errors.New("error precision param")
	}
	// encode latitude
	latitudeBinaryEncode := make([]byte, precisonParam.latBinBits)
	for i, latRange := 0, globalLatitudeRange; i < precisonParam.latBinBits; i++ {
		latitudeCode, nextLatRange := findBinaryCode(latRange, latitude)
		latRange = nextLatRange
		latitudeBinaryEncode[i] = latitudeCode
	}
	// encode longtitude
	longitudeBinaryEncode := make([]byte, precisonParam.lngBinBits)
	for i, lngRange := 0, globalLongitudeRange; i < precisonParam.lngBinBits; i++ {
		longitudeCode, nextLngRange := findBinaryCode(lngRange, longitude)
		lngRange = nextLngRange
		longitudeBinaryEncode[i] = longitudeCode
	}
	binaryEncode := make([]byte, precisonParam.latBinBits+precisonParam.lngBinBits)
	// merge lat encode and lng encode
	for i := 0; i < precisonParam.latBinBits+precisonParam.lngBinBits; i++ {
		if i%2 == 0 {
			binaryEncode[i] = longitudeBinaryEncode[int(i/2)]
		} else {
			binaryEncode[i] = latitudeBinaryEncode[int(i/2)]
		}
	}
	base32Bytes, err := binaryToBase32(binaryEncode)
	if err != nil {
		return "", err
	}
	return string(base32Bytes), nil
}

//Decode transform the geohash string to a latitude && longitude location
func Decode(geohashStr string) (float64, float64, error) {
	latitudeRange, longtitudeRange, err := decodeToRange(geohashStr)
	if err != nil {
		return 0, 0, err
	}
	return latitudeRange.GetMidVal(), longtitudeRange.GetMidVal(), nil
}

func decodeToRange(geohashStr string) (*Range, *Range, error) {
	getRange := func(r Range, binaryCodes []byte) *Range {
		for _, binaryByte := range binaryCodes {
			midVal := r.GetMidVal()
			if binaryByte == '0' {
				r.maxVal = midVal
			} else {
				r.minVal = midVal
			}
		}
		return &Range{minVal: r.minVal, maxVal: r.maxVal}
	}
	binaryEncodes, err := base32ToBinary([]byte(geohashStr))
	if err != nil {
		return nil, nil, err
	}
	// longtitude encode index: even; latitude encode index: odd
	var latitudeBinaryEncode []byte
	var longitudeBinaryEncode []byte
	for i, val := range binaryEncodes {
		if i%2 == 0 {
			longitudeBinaryEncode = append(longitudeBinaryEncode, val)
		} else {
			latitudeBinaryEncode = append(latitudeBinaryEncode, val)
		}
	}
	return getRange(globalLatitudeRange, latitudeBinaryEncode), getRange(globalLongitudeRange, longitudeBinaryEncode), nil
}

//Neighbours find the neighbours geohash strings based on a center geohash.
//Specifically, the neighbourPos list indicates the precise selected postions.
//The input param geohashStr is the center postion.
//The neighbourhood postions is stipultaed as follows:
// --------------------------------------------
// |              |              |            |
// |  northwest   |     north    | northeast  |
// |              |              |            |
// |--------------|--------------|------------|
// |              |              |            |
// |    west      |    center    |   east     |
// |              |              |            |
// |--------------|--------------|------------|
// |              |              |            |
// |  southwest   |    south     | southeast  |
// |              |              |            |
// |--------------|--------------|------------|
func Neighbours(geohashStr string, neighbourPos []string) (map[string]string, error) {

	findNeighbourShiftIndex := func(pos string) (int16, int16) {
		switch pos {
		case "northwest":
			return 1, -1
		case "north":
			return 1, 0
		case "northeast":
			return 1, 1
		case "west":
			return 0, -1
		case "center":
			return 0, 0
		case "east":
			return 0, 1
		case "southwest":
			return -1, -1
		case "south":
			return -1, 0
		case "southeast":
			return -1, 1
		default:
			return 0, 0
		}
	}
	latitudeRange, longtitudeRange, err := decodeToRange(geohashStr)
	if err != nil {
		return make(map[string]string), err
	}
	lattitudeRangeDiff, longtitudeRangeDiff := latitudeRange.GetRangeDiff(), longtitudeRange.GetRangeDiff()
	neighbours := make(map[string]string)
	precision := len(geohashStr)
	for _, pos := range neighbourPos {
		latitudeShiftIndex, longitudeShiftIndex := findNeighbourShiftIndex(pos)
		tmpEncode, tmpErr := Encode(
			latitudeRange.GetMidVal()+float64(latitudeShiftIndex)*lattitudeRangeDiff, longtitudeRange.GetMidVal()+float64(longitudeShiftIndex)*longtitudeRangeDiff, precision)
		if tmpErr == nil {
			neighbours[pos] = tmpEncode
		}
	}
	return neighbours, nil

}

// Nearby returns the nearby geohash string list within the certain area
func Nearby(geohashStr string, radius int32) ([]string, error) {
	if len(geohashStr) > 7 {
		return []string{}, errors.New("geohash length cannot larger than 7")
	}
	precisonParam, ok := encodeParamMap[len(geohashStr)]
	if !ok {
		return []string{}, errors.New("error precision param")
	}
	extendLayer := int32(math.Ceil(float64(radius) / precisonParam.vagueRange))
	var curLayer int32
	s := newSquare(geohashStr)
	for {
		if curLayer >= extendLayer {
			break
		}
		s = s.extendSqaure()
		curLayer++
	}
	var result []string
	allElementSet := s.listAllElements()
	for _, v := range allElementSet.All() {
		nearbyElement := v.(string)
		if dis, err := geohashDistance(nearbyElement, geohashStr); err == nil && dis <= float64(radius) {
			result = append(result, nearbyElement)
		}
	}
	return result, nil
}

func geohashDistance(fromGeohash, toGeohash string) (float64, error) {
	if fromLat, fromLng, err := Decode(fromGeohash); err == nil {
		if toLat, toLng, err := Decode(toGeohash); err == nil {
			fromPoint := NewLocation(fromLat, fromLng)
			toPoint := NewLocation(toLat, toLng)
			return fromPoint.EuclideanDistance(toPoint), nil
		}
	}
	return -1, errors.New("Invalid geohash")
}

type square struct {
	innerSqaure                                                            *square
	eastSide, southSide, westSide, northSide                               *structure.Set
	northeastElement, northwestElement, southeastElement, southwestElement string
}

func (s *square) extendSqaure() *square {

	extendDirectionSide := func(extendSet *structure.Set, sqaureSideSet *structure.Set, direction string) {
		sideSet := sqaureSideSet.All()
		for _, side := range sideSet {
			neighbour, tmpErr := Neighbours(side.(string), []string{direction})
			if tmpErr == nil {
				tmpNeighbour, tmpOK := neighbour[direction]
				if tmpOK {
					extendSet.Add(tmpNeighbour)
				}
			}
		}
	}

	extendEastSide := structure.NewSet()
	extendSouthSide := structure.NewSet()
	extendWestSide := structure.NewSet()
	extendNorthSide := structure.NewSet()
	var northeastElement, northwestElement, southeastElement, southwestElement string
	if s.eastSide != nil {
		extendDirectionSide(extendEastSide, s.eastSide, "east")
	}
	if s.southSide != nil {
		extendDirectionSide(extendSouthSide, s.southSide, "south")
	}
	if s.westSide != nil {
		extendDirectionSide(extendWestSide, s.westSide, "west")
	}
	if s.northSide != nil {
		extendDirectionSide(extendNorthSide, s.northSide, "north")
	}
	northeastNeighbours, northeastNeighboursErr := Neighbours(s.northeastElement, []string{"northeast", "north", "east"})
	if northeastNeighboursErr == nil {
		if northeastNeighbour, northeastNeighbourOK := northeastNeighbours["northeast"]; northeastNeighbourOK {
			northeastElement = northeastNeighbour
		}
		if northNeighbour, northNeighbourOK := northeastNeighbours["north"]; northNeighbourOK {
			extendNorthSide.Add(northNeighbour)
		}
		if eastNeighbour, eastNeighbourOK := northeastNeighbours["east"]; eastNeighbourOK {
			extendEastSide.Add(eastNeighbour)
		}
	}
	northwestNeighbours, northwestNeighboursErr := Neighbours(s.northwestElement, []string{"northwest", "north", "west"})
	if northwestNeighboursErr == nil {
		if northwestNeighbour, northwestNeighbourOK := northwestNeighbours["northwest"]; northwestNeighbourOK {
			northwestElement = northwestNeighbour
		}
		if northNeighbour, northNeighbourOK := northwestNeighbours["north"]; northNeighbourOK {
			extendNorthSide.Add(northNeighbour)
		}
		if westNeighbour, westNeighbourOK := northwestNeighbours["west"]; westNeighbourOK {
			extendWestSide.Add(westNeighbour)
		}
	}
	southeastNeighbours, southeastNeighboursErr := Neighbours(s.southeastElement, []string{"southeast", "south", "east"})
	if southeastNeighboursErr == nil {
		if southeastNeighbour, southeastNeighbourOK := southeastNeighbours["southeast"]; southeastNeighbourOK {
			southeastElement = southeastNeighbour
		}
		if southNeighbour, southNeighbourOK := southeastNeighbours["south"]; southNeighbourOK {
			extendSouthSide.Add(southNeighbour)
		}
		if eastNeighbour, eastNeighbourOK := southeastNeighbours["east"]; eastNeighbourOK {
			extendEastSide.Add(eastNeighbour)
		}
	}
	southwestNeighbours, southwestNeighboursErr := Neighbours(s.southwestElement, []string{"southwest", "south", "west"})
	if southwestNeighboursErr == nil {
		if southwestNeighbour, southwestNeighbourOK := southwestNeighbours["southwest"]; southwestNeighbourOK {
			southwestElement = southwestNeighbour
		}
		if southNeighbour, southNeighbourOK := southwestNeighbours["south"]; southNeighbourOK {
			extendSouthSide.Add(southNeighbour)
		}
		if westNeighbour, westNeighbourOK := southwestNeighbours["west"]; westNeighbourOK {
			extendWestSide.Add(westNeighbour)
		}
	}
	return &square{
		innerSqaure:      s,
		eastSide:         extendEastSide,
		southSide:        extendSouthSide,
		westSide:         extendWestSide,
		northSide:        extendNorthSide,
		northeastElement: northeastElement,
		northwestElement: northwestElement,
		southeastElement: southeastElement,
		southwestElement: southwestElement,
	}
}

func (s *square) listAllElements() *structure.Set {
	set := structure.NewSet()
	set.Add(s.northeastElement)
	set.Add(s.northwestElement)
	set.Add(s.southeastElement)
	set.Add(s.southwestElement)
	set = structure.Union(set, s.southSide, s.northSide, s.westSide, s.eastSide)
	if s.innerSqaure == nil {
		return set
	}
	return structure.Union(set, s.innerSqaure.listAllElements())
}

func newSquare(element string) *square {
	return &square{
		innerSqaure:      nil,
		eastSide:         nil,
		southSide:        nil,
		westSide:         nil,
		northSide:        nil,
		northeastElement: element,
		northwestElement: element,
		southeastElement: element,
		southwestElement: element,
	}
}

// Range indicates a float64 range marked by min val and max val
type Range struct {
	minVal, maxVal float64
}

// GetMidVal return the average val of min and max
func (r *Range) GetMidVal() float64 {
	return 0.5 * (r.minVal + r.maxVal)
}

// GetRangeDiff return the substract val of max and min
func (r *Range) GetRangeDiff() float64 {
	return r.maxVal - r.minVal
}

// CheckInRange check the input param is in the range of min and max
func (r *Range) CheckInRange(val float64) bool {
	if val >= r.minVal && val <= r.maxVal {
		return true
	}
	return false
}

//encodePrecison indicates the encode and decode's precison degree.
// precision: defined as base32 encode string length,
// normnally precison increases in accordance with the encode string length
// latBinBits: defined as the binary encoding bits for latitude value
// lngBinBits: defined as the binary encoding bits for longtitude value
type encodePrecison struct {
	precision, latBinBits, lngBinBits int
	vagueRange                        float64
}
