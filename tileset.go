package tilepix

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	sprite  *pixel.Sprite
	picture pixel.Picture

	// parentMap is the map which contains this object
	parentMap *Map
	// dir is the directory the tsx file is located in.  This is used to access assets via a relative path.
	dir string
}

func readTileset(r io.Reader, dir string) (*Tileset, error) {
	log.Debug("readTileset: reading from io.Reader")

	d := xml.NewDecoder(r)

	var t Tileset
	if err := d.Decode(&t); err != nil {
		log.WithError(err).Error("readTileset: could not decode to Tileset")
		return nil, err
	}

	t.dir = dir

	return validate(t)
}

func readTilesetFile(filePath string) (*Tileset, error) {
	log.WithField("Filepath", filePath).Debug("readTilesetFile: reading file")

	f, err := os.Open(filePath)
	if err != nil {
		log.WithError(err).Error("ReadFile: could not open file")
		return nil, err
	}
	defer f.Close()

	dir := filepath.Dir(filePath)

	return readTileset(f, dir)
}

// GenerateTileObjectLayer will create a new ObjectGroup for the mapping of Objects to individual tiles.
func (ts Tileset) GenerateTileObjectLayer(tileLayers []*TileLayer) ObjectGroup {
	group := ObjectGroup{Name: fmt.Sprintf("%s-objectgroup", ts.Name)}
	objs := ts.TileObjects()

	// Loop all TileLayers in map.
	for _, tl := range tileLayers {
		// Loop all DecodedTiles in the TileLayer
		for ind, t := range tl.DecodedTiles {
			if t.Nil {
				// Skip blank tiles.
				continue
			}

			// Try get the Tiles' ObjectGroup.
			og, ok := objs[t.ID]
			if !ok {
				// Not object groups for this Tile.
				continue
			}

			// Loop all objects in the Tiles' ObjectGroup.
			for _, obs := range og.Objects {
				// Create a new Object based on the relative position of the Object and the DecodedTile.
				o := *obs
				tilePos := t.Position(ind, &ts)
				o.X += tilePos.X
				o.Y += tilePos.Y

				group.Objects = append(group.Objects, &o)
			}
		}
	}

	return group
}

func validate(t Tileset) (*Tileset, error) {
	if t.Columns < 1 {
		return nil, fmt.Errorf("Tileset columns value not valid")
	}
	return &t, nil
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

func (ts *Tileset) setSprite() pixel.Picture {
	if ts.sprite != nil {
		// Return if sprite already set
		return ts.picture
	}

	dir := ts.dir
	if dir == "" {
		dir = ts.parentMap.dir
	}

	sprite, pictureData, err := loadSpriteFromFile(filepath.Join(dir, ts.Image.Source))
	if err != nil {
		log.WithField("Filepath", filepath.Join(dir, ts.Image.Source)).WithError(err).Error("Tileset.setSprite: could not load sprite from file")
		return nil
	}

	ts.sprite = sprite
	ts.picture = pictureData
	return ts.picture
}

// TileObjects will return all ObjectGroups contained in Tiles.
func (ts Tileset) TileObjects() map[ID]*ObjectGroup {
	objs := make(map[ID]*ObjectGroup)
	for _, t := range ts.Tiles {
		if t.ObjectGroup != nil {
			objs[t.ID] = t.ObjectGroup
		}
	}

	return objs
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
