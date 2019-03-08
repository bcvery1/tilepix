package tmx

type Layer struct {
	Name         string         `xml:"name,attr"`
	Opacity      float32        `xml:"opacity,attr"`
	Visible      bool           `xml:"visible,attr"`
	Properties   []Property     `xml:"properties>property"`
	Data         Data           `xml:"data"`
	// DecodedTiles is the attribute you should use instead of `Data`.
	// Tile entry at (x,y) is obtained using l.DecodedTiles[y*map.Width+x].
	DecodedTiles []*DecodedTile
	// Tileset is only set when the layer uses a single tileset and NilLayer is false.
	Tileset      *Tileset
	// Empty should be set when all entries of the layer are NilTile.
	Empty        bool
}
