package tilepix

import "fmt"

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
	return decodePoints(p.Points)
}

func (p *PolyLine) String() string {
	return fmt.Sprintf("Polyline{Points: %v}", p.decodedPoints)
}

func (p *PolyLine) setParent(m *Map) {
	p.parentMap = m

	for _, dp := range p.decodedPoints {
		dp.setParent(m)
	}
}
