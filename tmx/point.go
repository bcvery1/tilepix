package tmx

import (
	"strconv"
	"strings"
)

type Point struct {
	X int
	Y int
}

func decodePoints(s string) (points []Point, err error) {
	pointStrings := strings.Split(s, " ")

	points = make([]Point, len(pointStrings))
	for i, pointString := range pointStrings {
		coordStrings := strings.Split(pointString, ",")
		if len(coordStrings) != 2 {
			return []Point{}, InvalidPointsFieldError
		}

		points[i].X, err = strconv.Atoi(coordStrings[0])
		if err != nil {
			return []Point{}, err
		}

		points[i].Y, err = strconv.Atoi(coordStrings[1])
		if err != nil {
			return []Point{}, err
		}
	}
	return
}
