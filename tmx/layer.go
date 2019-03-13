package tmx

import (
	"github.com/bcvery1/tilepix/utilities"
	"github.com/faiface/pixel"
)

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

	batch     *pixel.Batch
	mapParent *Map
}

// Batch returns the batch with the picture data from the tileset associated with this layer.
func (l *Layer) Batch() (*pixel.Batch, error) {
	if l.batch == nil {
		// TODO(need to do this either by file or reader)
		sprite, pictureData, err := utilities.LoadSpriteFromFile(l.Tileset.Image.Source)
		if err != nil {
			return nil, err
		}

		l.batch = pixel.NewBatch(&pixel.TrianglesData{}, pictureData)
		l.Tileset.sprite = sprite
	}

	return l.batch, nil
}

func (l *Layer) Draw(target pixel.Target) error {
	// Initialise the batch
	if _, err := l.Batch(); err != nil {
		return err
	}

	// Loop through each decoded tile
	for tileIndex, tile := range l.DecodedTiles {
		ts := l.Tileset
		tID := int(tile.ID)

		if tID == 0 {
			// Tile ID 0 means blank, skip it.
			continue
		}

		// Calculate the framing for the tile within its tileset's source image
		numRows := ts.Tilecount / ts.Columns
		x, y := utilities.TileIDToCoord(tID, ts.Columns, numRows)
		gamePos := utilities.IndexToGamePos(tileIndex, l.mapParent.Width, l.mapParent.Height)

		iX := float64(x) * float64(ts.TileWidth)
		fX := iX + float64(ts.TileWidth)
		iY := float64(y) * float64(ts.TileHeight)
		fY := iY + float64(ts.TileHeight)

		l.Tileset.sprite.Set(l.Tileset.sprite.Picture(), pixel.R(iX, iY, fX, fY))
		pos := gamePos.ScaledXY(pixel.V(float64(ts.TileWidth), float64(ts.TileHeight)))
		l.Tileset.sprite.Draw(l.batch, pixel.IM.Moved(pos))
	}

	l.batch.Draw(target)
	return nil
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
