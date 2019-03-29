package tilepix

import (
	"fmt"

	"github.com/faiface/pixel"

	log "github.com/sirupsen/logrus"
)

/*
  _____ _ _             _
 |_   _(_) |___ ___ ___| |_
   | | | | / -_|_-</ -_)  _|
   |_| |_|_\___/__/\___|\__|
*/

// Tileset is a TMX file structure which represents a Tiled Tileset
type Tileset struct {
	FirstGID   GID         `xml:"firstgid,attr"`
	Source     string      `xml:"source,attr"`
	Name       string      `xml:"name,attr"`
	TileWidth  int         `xml:"tilewidth,attr"`
	TileHeight int         `xml:"tileheight,attr"`
	Spacing    int         `xml:"spacing,attr"`
	Margin     int         `xml:"margin,attr"`
	Properties []*Property `xml:"properties>property"`
	Image      *Image      `xml:"image"`
	Tiles      []*Tile     `xml:"tile"`
	Tilecount  int         `xml:"tilecount,attr"`
	Columns    int         `xml:"columns,attr"`

	sprite *pixel.Sprite

	// parentMap is the map which contains this object
	parentMap *Map
}

func (ts *Tileset) String() string {
	return fmt.Sprintf(
		"TileSet{Name: %s, Tile size: %dx%d, Tile spacing: %d, Tilecount: %d, Properties: %v}",
		ts.Name,
		ts.TileWidth,
		ts.TileHeight,
		ts.Spacing,
		ts.Tilecount,
		ts.Properties,
	)
}

func (ts *Tileset) setParent(m *Map) {
	ts.parentMap = m

	for _, p := range ts.Properties {
		p.setParent(m)
	}
	for _, t := range ts.Tiles {
		t.setParent(m)
	}

	if ts.Image != nil {
		ts.Image.setParent(m)
	}
}

func (ts *Tileset) setSprite() {
	if ts.sprite != nil {
		// Return if sprite already set
		return
	}

	sprite, _, err := loadSpriteFromFile(ts.Image.Source)
	if err != nil {
		log.WithError(err).Error("Tileset.setSprite: could not load sprite from file")
		return
	}

	ts.sprite = sprite
}

func getTileset(l *TileLayer) (tileset *Tileset, isEmpty, usesMultipleTilesets bool) {
	for _, tile := range l.DecodedTiles {
		if !tile.Nil {
			if tileset == nil {
				tileset = tile.Tileset
			} else if tileset != tile.Tileset {
				return tileset, false, true
			}
		}
	}

	if tileset == nil {
		return nil, true, false
	}

	return tileset, false, false
}
