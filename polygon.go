package tilepix

import "fmt"

/*
  ___     _
 | _ \___| |_  _ __ _ ___ _ _
 |  _/ _ \ | || / _` / _ \ ' \
 |_| \___/_|\_, \__, \___/_||_|
            |__/|___/
*/

// Polygon is a TMX file structure representing a Tiled Polygon.
type Polygon struct {
	Points string `xml:"points,attr"`

	decodedPoints []*Point

	// parentMap is the map which contains this object
	parentMap *Map
}

// Decode will return a slice of points which make up this polygon.
func (p *Polygon) Decode() ([]*Point, error) {
	return decodePoints(p.Points)
}

func (p *Polygon) setParent(m *Map) {
	p.parentMap = m

	for _, dp := range p.decodedPoints {
		dp.setParent(m)
	}
}

func (p *Polygon) String() string {
	return fmt.Sprintf("Polygon{Points: %v}", p.decodedPoints)
}
