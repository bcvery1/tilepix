package tmx

type Tile struct {
	ID    ID    `xml:"id,attr"`
	Image Image `xml:"image"`
}

type DecodedTile struct {
	ID             ID
	Tileset        *Tileset
	HorizontalFlip bool
	VerticalFlip   bool
	DiagonalFlip   bool
	Nil            bool
}

func (t *DecodedTile) IsNil() bool {
	return t.Nil
}
