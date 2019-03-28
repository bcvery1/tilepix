package tilepix

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

/*
  ___     _      _ _
 | _ \___| |_  _| (_)_ _  ___
 |  _/ _ \ | || | | | ' \/ -_)
 |_| \___/_|\_, |_|_|_||_\___|
            |__/
*/

// PolyLine is a TMX file structure representing a Tiled Polyline.
type PolyLine struct {
	Points string `xml:"points,attr"`

	decodedPoints []*Point

	// parentMap is the map which contains this object
	parentMap *Map
}

// Decode will return a slice of points which make up this polyline.
func (p *PolyLine) Decode() ([]*Point, error) {
	if p.decodedPoints == nil {
		dp, err := decodePoints(p.Points)
		if err != nil {
			log.WithError(err).Error("PolyLine.Decode: could not decode points")
			return nil, err
		}

		p.decodedPoints = dp
	}
	return p.decodedPoints, nil
}

func (p *PolyLine) String() string {
	return fmt.Sprintf("Polyline{Points: %v}", p.decodedPoints)
}

func (p *PolyLine) setParent(m *Map) {
	p.parentMap = m

	// Must decode points before they can be set
	_, err := p.Decode()
	if err != nil {
		log.WithError(err).Error("PolyLine.setParent: could not decode points")
		return
	}

	for _, dp := range p.decodedPoints {
		dp.setParent(m)

		// We have to flip the Y co-ordinate here because the `tilepix.Point` is only used to provide `pixel.Vec`s
		dp.flipY()
	}
}
