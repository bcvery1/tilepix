package tmx

type Polygon struct {
	Points string `xml:"points,attr"`
}

func (p *Polygon) Decode() ([]Point, error) {
	return decodePoints(p.Points)
}
