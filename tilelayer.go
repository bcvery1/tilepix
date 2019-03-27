package tilepix

import (
	"errors"
	"fmt"

	"github.com/faiface/pixel"

	log "github.com/sirupsen/logrus"
)

/*
  _____ _ _     _
 |_   _(_) |___| |   __ _ _  _ ___ _ _
   | | | | / -_) |__/ _` | || / -_) '_|
   |_| |_|_\___|____\__,_|\_, \___|_|
                          |__/
*/

// TileLayer is a TMX file structure which can hold any type of Tiled layer.
type TileLayer struct {
	Name       string      `xml:"name,attr"`
	Opacity    float32     `xml:"opacity,attr"`
	Visible    bool        `xml:"visible,attr"`
	Properties []*Property `xml:"properties>property"`
	Data       Data        `xml:"data"`
	// DecodedTiles is the attribute you should use instead of `Data`.
	// Tile entry at (x,y) is obtained using l.DecodedTiles[y*map.Width+x].
	DecodedTiles []*DecodedTile
	// Tileset is only set when the layer uses a single tileset and NilLayer is false.
	Tileset *Tileset
	// Empty should be set when all entries of the layer are NilTile.
	Empty bool

	batch *pixel.Batch

	// parentMap is the map which contains this object
	parentMap *Map
}

// Batch returns the batch with the picture data from the tileset associated with this layer.
func (l *TileLayer) Batch() (*pixel.Batch, error) {
	if l.batch == nil {
		log.Debug("TileLayer.Batch: batch not initialised, creating")

		if l.Tileset == nil {
			err := errors.New("cannot create sprite from nil tileset")
			log.WithError(err).Error("TileLayer.Batch: layers' tileset is nil")
			return nil, err
		}

		sprite, pictureData, err := loadSpriteFromFile(l.Tileset.Image.Source)
		if err != nil {
			log.WithError(err).Error("TileLayer.Batch: could not load sprite from file")
			return nil, err
		}

		l.batch = pixel.NewBatch(&pixel.TrianglesData{}, pictureData)
		l.Tileset.sprite = sprite
	}

	l.batch.Clear()

	return l.batch, nil
}

// Draw will use the TileLayers' batch to draw all tiles within the TileLayer to the target.
func (l *TileLayer) Draw(target pixel.Target) error {
	// Initialise the batch
	if _, err := l.Batch(); err != nil {
		log.WithError(err).Error("TileLayer.Draw: could not get batch")
		return err
	}

	ts := l.Tileset
	numRows := ts.Tilecount / ts.Columns

	// Loop through each decoded tile
	for tileIndex, tile := range l.DecodedTiles {
		tile.Draw(tileIndex, ts.Columns, numRows, ts, l.batch)
	}

	l.batch.Draw(target)
	return nil
}

func (l *TileLayer) String() string {
	return fmt.Sprintf("TileLayer{Name: '%s', Properties: %v, TileCount: %d}", l.Name, l.Properties, len(l.DecodedTiles))
}

func (l *TileLayer) decode(width, height int) ([]GID, error) {
	log.WithField("Encoding", l.Data.Encoding).Debug("TileLayer.decode: determining encoding")

	switch l.Data.Encoding {
	case "csv":
		return l.decodeLayerCSV(width, height)
	case "base64":
		return l.decodeLayerBase64(width, height)
	case "":
		// XML "encoding"
		return l.decodeLayerXML(width, height)
	}

	log.WithError(ErrUnknownEncoding).Error("TileLayer.decode: unrecognised encoding")
	return nil, ErrUnknownEncoding
}

func (l *TileLayer) decodeLayerXML(width, height int) ([]GID, error) {
	if len(l.Data.DataTiles) != width*height {
		log.WithError(ErrInvalidDecodedDataLen).WithFields(log.Fields{"Length datatiles": len(l.Data.DataTiles), "W*H": width * height}).Error("TileLayer.decodeLayerXML: data length mismatch")
		return nil, ErrInvalidDecodedDataLen
	}

	gids := make([]GID, len(l.Data.DataTiles))
	for i := 0; i < len(gids); i++ {
		gids[i] = l.Data.DataTiles[i].GID
	}

	return gids, nil
}

func (l *TileLayer) decodeLayerCSV(width, height int) ([]GID, error) {
	gids, err := l.Data.decodeCSV()
	if err != nil {
		log.WithError(err).Error("TileLayer.decodeLayerCSV: could not decode CSV")
		return nil, err
	}

	if len(gids) != width*height {
		log.WithError(ErrInvalidDecodedDataLen).WithFields(log.Fields{"Length GIDSs": len(gids), "W*H": width * height}).Error("TileLayer.decodeLayerCSV: data length mismatch")
		return nil, ErrInvalidDecodedDataLen
	}

	return gids, nil
}

func (l *TileLayer) decodeLayerBase64(width, height int) ([]GID, error) {
	dataBytes, err := l.Data.decodeBase64()
	if err != nil {
		log.WithError(err).Error("TileLayer.decodeLayerBase64: could not decode base64")
		return nil, err
	}

	if len(dataBytes) != width*height*4 {
		log.WithError(ErrInvalidDecodedDataLen).WithFields(log.Fields{"Length databytes": len(dataBytes), "W*H": width * height}).Error("TileLayer.decodeLayerBase64: data length mismatch")
		return nil, ErrInvalidDecodedDataLen
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

func (l *TileLayer) setParent(m *Map) {
	l.parentMap = m

	for _, p := range l.Properties {
		p.setParent(m)
	}

	for _, dt := range l.DecodedTiles {
		dt.setParent(m)
	}

	if l.Tileset != nil {
		l.Tileset.setParent(m)
	}
}
