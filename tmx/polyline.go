package tmx

type PolyLine struct {
	Points string `xml:"points,attr"`
}
func (p *PolyLine) Decode() ([]Point, error) {
	return decodePoints(p.Points)
}
