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

func (m *Map) decodeLayerXML(l *Layer) ([]GID, error) {
	if len(l.Data.DataTiles) != m.Width*m.Height {
		return []GID{}, InvalidDecodedDataLenError
	}

	gids := make([]GID, len(l.Data.DataTiles))
	for i := 0; i < len(gids); i++ {
		gids[i] = l.Data.DataTiles[i].GID
	}

	return gids, nil
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

func (m *Map) decodeLayerCSV(l *Layer) ([]GID, error) {
	gids, err := l.Data.decodeCSV()
	if err != nil {
		return []GID{}, err
	}

	if len(gids) != m.Width*m.Height {
		return []GID{}, InvalidDecodedDataLenError
	}

	return gids, nil
}

func (m *Map) decodeLayerBase64(l *Layer) ([]GID, error) {
	dataBytes, err := l.Data.decodeBase64()
	if err != nil {
		return []GID{}, err
	}

	if len(dataBytes) != m.Width*m.Height*4 {
		return []GID{}, InvalidDecodedDataLenError
	}

	gids := make([]GID, m.Width*m.Height)

	j := 0
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			gid := GID(dataBytes[j]) +
				GID(dataBytes[j+1])<<8 +
				GID(dataBytes[j+2])<<16 +
				GID(dataBytes[j+3])<<24
			j += 4

			gids[y*m.Width+x] = gid
		}
	}

	return gids, nil
}

func (m *Map) decodeLayer(l *Layer) ([]GID, error) {
	switch l.Data.Encoding {
	case "csv":
		return m.decodeLayerCSV(l)
	case "base64":
		return m.decodeLayerBase64(l)
	case "": // XML "encoding"
		return m.decodeLayerXML(l)
	}
	return []GID{}, UnknownEncodingError
}

func (m *Map) decodeLayers() error {
	for i := 0; i < len(m.Layers); i++ {
		l := &m.Layers[i]
		gids, err := m.decodeLayer(l)
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
