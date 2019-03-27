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

func decodePoints(s string) (points []*Point, err error) {
	pointStrings := strings.Split(s, " ")

	points = make([]*Point, len(pointStrings))
	for i, pointString := range pointStrings {
		coordStrings := strings.Split(pointString, ",")
		if len(coordStrings) != 2 {
			log.WithError(ErrInvalidPointsField).WithField("Co-ordinate strings", coordStrings).Error("decodePoints: mismatch co-ordinates string length")
			return nil, ErrInvalidPointsField
		}

		points[i].X, err = strconv.Atoi(coordStrings[0])
		if err != nil {
			log.WithError(err).WithField("Point string", coordStrings[0]).Error("decodePoints: could not parse X co-ordinate string")
			return nil, err
		}

		points[i].Y, err = strconv.Atoi(coordStrings[1])
		if err != nil {
			log.WithError(err).WithField("Point string", coordStrings[1]).Error("decodePoints: could not parse X co-ordinate string")
			return nil, err
		}
	}
	return
}
