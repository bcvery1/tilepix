package tmx

type Layer struct {
	Name       string     `xml:"name,attr"`
	Opacity    float32    `xml:"opacity,attr"`
	Visible    bool       `xml:"visible,attr"`
	Properties []Property `xml:"properties>property"`
	Data       Data       `xml:"data"`
	// DecodedTiles is the attribute you should use instead of `Data`.
	// Tile entry at (x,y) is obtained using l.DecodedTiles[y*map.Width+x].
	DecodedTiles []*DecodedTile
	// Tileset is only set when the layer uses a single tileset and NilLayer is false.
	Tileset *Tileset
	// Empty should be set when all entries of the layer are NilTile.
	Empty bool
}

func (l *Layer) decode(width, height int) ([]GID, error) {
	switch l.Data.Encoding {
	case "csv":
		return l.decodeLayerCSV(width, height)
	case "base64":
		return l.decodeLayerBase64(width, height)
	case "":
		// XML "encoding"
		return l.decodeLayerXML(width, height)
	}
	return nil, UnknownEncodingError
}

func (l *Layer) decodeLayerXML(width, height int) ([]GID, error) {
	if len(l.Data.DataTiles) != width*height {
		return nil, InvalidDecodedDataLenError
	}

	gids := make([]GID, len(l.Data.DataTiles))
	for i := 0; i < len(gids); i++ {
		gids[i] = l.Data.DataTiles[i].GID
	}

	return gids, nil
}

func (l *Layer) decodeLayerCSV(width, height int) ([]GID, error) {
	gids, err := l.Data.decodeCSV()
	if err != nil {
		return nil, err
	}

	if len(gids) != width*height {
		return nil, InvalidDecodedDataLenError
	}

	return gids, nil
}

func (l *Layer) decodeLayerBase64(width, height int) ([]GID, error) {
	dataBytes, err := l.Data.decodeBase64()
	if err != nil {
		return nil, err
	}

	if len(dataBytes) != width*height*4 {
		return nil, InvalidDecodedDataLenError
	}

	gids := make([]GID, width*height)

	j := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gid := GID(dataBytes[j]) +
				GID(dataBytes[j+1])<<8 +
				GID(dataBytes[j+2])<<16 +
				GID(dataBytes[j+3])<<24
			j += 4

			gids[y*width+x] = gid
		}
	}

	return gids, nil
}
