package tmx

type Map struct {
	Version      string        `xml:"title,attr"`
	Orientation  string        `xml:"orientation,attr"`
	Width        int           `xml:"width,attr"`
	Height       int           `xml:"height,attr"`
	TileWidth    int           `xml:"tilewidth,attr"`
	TileHeight   int           `xml:"tileheight,attr"`
	Properties   []Property    `xml:"properties>property"`
	Tilesets     []Tileset     `xml:"tileset"`
	Layers       []Layer       `xml:"layer"`
	ObjectGroups []ObjectGroup `xml:"objectgroup"`
}

func (m *Map) DecodeGID(gid GID) (*DecodedTile, error) {
	if gid == 0 {
		return NilTile, nil
	}

	gidBare := gid &^ GIDFlip

	for i := len(m.Tilesets) - 1; i >= 0; i-- {
		if m.Tilesets[i].FirstGID <= gidBare {
			return &DecodedTile{
				ID:             ID(gidBare - m.Tilesets[i].FirstGID),
				Tileset:        &m.Tilesets[i],
				HorizontalFlip: gid&GIDHorizontalFlip != 0,
				VerticalFlip:   gid&GIDVerticalFlip != 0,
				DiagonalFlip:   gid&GIDDiagonalFlip != 0,
				Nil:            false,
			}, nil
		}
	}

	return nil, InvalidGIDError
}

func (m *Map) decodeLayers() error {
	for _, l := range m.Layers {
		gids, err := l.decode(m.Width, m.Height)
		if err != nil {
			return err
		}

		l.DecodedTiles = make([]*DecodedTile, len(gids))
		for j := 0; j < len(l.DecodedTiles); j++ {
			l.DecodedTiles[j], err = m.DecodeGID(gids[j])
			if err != nil {
				return err
			}
		}
	}

	return nil
}
