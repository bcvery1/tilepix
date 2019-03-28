package tilepix

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/faiface/pixel"

	log "github.com/sirupsen/logrus"
)

/*
  ___     _     _
 | _ \___(_)_ _| |_
 |  _/ _ \ | ' \  _|
 |_| \___/_|_||_\__|
*/

// Point is a TMX file structure holding a Tiled Point object.
type Point struct {
	X int
	Y int

	// parentMap is the map which contains this object
	parentMap *Map
}

func (p *Point) String() string {
	return fmt.Sprintf("Point{%d, %d}", p.X, p.Y)
}

// V converts the Tiled Point to a Pixel Vector.
func (p *Point) V() pixel.Vec {
	return pixel.V(float64(p.X), float64(p.Y))
}

func (p *Point) setParent(m *Map) {
	p.parentMap = m
}

func decodePoints(s string) ([]*Point, error) {
	pointStrings := strings.Split(s, " ")

	var points []*Point
	var err error
	for _, pointString := range pointStrings {
		coordStrings := strings.Split(pointString, ",")
		if len(coordStrings) != 2 {
			log.WithError(ErrInvalidPointsField).WithField("Co-ordinate strings", coordStrings).Error("decodePoints: mismatch co-ordinates string length")
			return nil, ErrInvalidPointsField
		}

		point := &Point{}

		point.X, err = strconv.Atoi(coordStrings[0])
		if err != nil {
			log.WithError(err).WithField("Point string", coordStrings[0]).Error("decodePoints: could not parse X co-ordinate string")
			return nil, err
		}

		point.Y, err = strconv.Atoi(coordStrings[1])
		if err != nil {
			log.WithError(err).WithField("Point string", coordStrings[1]).Error("decodePoints: could not parse X co-ordinate string")
			return nil, err
		}

		points = append(points, point)
	}

	return points, nil
}

// flipY will get the inverse Y co-ordinate based on the parent maps' size.  This is because Tiled draws from the
// top-right instead of the bottom-left.
func (p *Point) flipY() {
	p.Y = int(p.parentMap.pixelHeight()) - p.Y
}
