package tilepix

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/faiface/pixel"
)

const (
	GIDHorizontalFlip = 0x80000000
	GIDVerticalFlip   = 0x40000000
	GIDDiagonalFlip   = 0x20000000
	GIDFlip           = GIDHorizontalFlip | GIDVerticalFlip | GIDDiagonalFlip
)

var (
	UnknownEncodingError       = errors.New("tmx: invalid encoding scheme")
	UnknownCompressionError    = errors.New("tmx: invalid compression method")
	InvalidDecodedDataLenError = errors.New("tmx: invalid decoded data length")
	InvalidGIDError            = errors.New("tmx: invalid GID")
	InvalidPointsFieldError    = errors.New("tmx: invalid points string")
)

var (
	NilTile = &DecodedTile{Nil: true}
)

type GID uint32 // A tile ID. Could be used for GID or ID.
type ID uint32

type DataTile struct {
	GID GID `xml:"gid,attr"`
}

func Read(r io.Reader) (*Map, error) {
	d := xml.NewDecoder(r)

	m := new(Map)
	if err := d.Decode(m); err != nil {
		return nil, err
	}

	if err := m.decodeLayers(); err != nil {
		return nil, err
	}

	for i := 0; i < len(m.Layers); i++ {
		l := m.Layers[i]
		l.mapParent = m

		tileset, isEmpty, usesMultipleTilesets := getTileset(l)
		if usesMultipleTilesets {
			continue
		}
		l.Empty, l.Tileset = isEmpty, tileset
	}

	return m, nil
}

func ReadFile(filePath string) (*Map, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return Read(f)
}

/*
  ___       _
 |   \ __ _| |_ __ _
 | |) / _` |  _/ _` |
 |___/\__,_|\__\__,_|

*/

type Data struct {
	Encoding    string `xml:"encoding,attr"`
	Compression string `xml:"compression,attr"`
	RawData     []byte `xml:",innerxml"`
	// DataTiles is only used when layer encoding is XML.
	DataTiles []DataTile `xml:"tile"`
}

func (d *Data) decodeBase64() (data []byte, err error) {
	rawData := bytes.TrimSpace(d.RawData)
	r := bytes.NewReader(rawData)

	encr := base64.NewDecoder(base64.StdEncoding, r)

	var comr io.Reader
	switch d.Compression {
	case "gzip":
		comr, err = gzip.NewReader(encr)
		if err != nil {
			return
		}
	case "zlib":
		comr, err = zlib.NewReader(encr)
		if err != nil {
			return
		}
	case "":
		comr = encr
	default:
		err = UnknownCompressionError
		return
	}

	return ioutil.ReadAll(comr)
}

func (d *Data) decodeCSV() ([]GID, error) {
	cleaner := func(r rune) rune {
		if (r >= '0' && r <= '9') || r == ',' {
			return r
		}
		return -1
	}

	rawDataClean := strings.Map(cleaner, string(d.RawData))

	str := strings.Split(string(rawDataClean), ",")

	gids := make([]GID, len(str))
	for i, s := range str {
		d, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return nil, err
		}
		gids[i] = GID(d)
	}
	return gids, nil
}

/*
  ___
 |_ _|_ __  __ _ __ _ ___
  | || '  \/ _` / _` / -_)
 |___|_|_|_\__,_\__, \___|
                |___/
*/

type Image struct {
	Source string `xml:"source,attr"`
	Trans  string `xml:"trans,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`

	sprite  *pixel.Sprite
	picture pixel.Picture
}

/*
  _
 | |   __ _ _  _ ___ _ _
 | |__/ _` | || / -_) '_|
 |____\__,_|\_, \___|_|
            |__/
*/

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
		if l.Tileset == nil {
			return nil, errors.New("cannot create sprite from nil tileset")
		}

		// TODO(need to do this either by file or reader)
		sprite, pictureData, err := loadSpriteFromFile(l.Tileset.Image.Source)
		if err != nil {
			return nil, err
		}

		l.batch = pixel.NewBatch(&pixel.TrianglesData{}, pictureData)
		l.Tileset.sprite = sprite
	}

	l.batch.Clear()

	return l.batch, nil
}

func (l *Layer) Draw(target pixel.Target) error {
	// Initialise the batch
	if _, err := l.Batch(); err != nil {
		return err
	}

	ts := l.Tileset
	numRows := ts.Tilecount / ts.Columns

	// Loop through each decoded tile
	for tileIndex, tile := range l.DecodedTiles {
		tID := int(tile.ID)

		if tile.IsNil() {
			continue
		}

		// Calculate the framing for the tile within its tileset's source image
		x, y := tileIDToCoord(tID, ts.Columns, numRows)
		gamePos := indexToGamePos(tileIndex, l.mapParent.Width, l.mapParent.Height)

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

/*
  __  __
 |  \/  |__ _ _ __
 | |\/| / _` | '_ \
 |_|  |_\__,_| .__/
             |_|
*/

type Map struct {
	Version      string        `xml:"title,attr"`
	Orientation  string        `xml:"orientation,attr"`
	Width        int           `xml:"width,attr"`
	Height       int           `xml:"height,attr"`
	TileWidth    int           `xml:"tilewidth,attr"`
	TileHeight   int           `xml:"tileheight,attr"`
	Properties   []Property    `xml:"properties>property"`
	Tilesets     []Tileset     `xml:"tileset"`
	Layers       []*Layer      `xml:"layer"`
	ObjectGroups []ObjectGroup `xml:"objectgroup"`
}

// DrawAll will draw all tile layers to the target.  This will use `pixel.Batch`s for efficiency.
func (m *Map) DrawAll(target pixel.Target) error {
	for _, l := range m.Layers {
		if err := l.Draw(target); err != nil {
			return err
		}
	}

	return nil
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
		for j := 0; j < len(gids); j++ {
			decTile, err := m.DecodeGID(gids[j])
			if err != nil {
				return err
			}
			l.DecodedTiles[j] = decTile
		}
	}

	return nil
}

/*
   ___  _     _        _
  / _ \| |__ (_)___ __| |_
 | (_) | '_ \| / -_) _|  _|
  \___/|_.__// \___\__|\__|
           |__/
*/

type Object struct {
	Name       string     `xml:"name,attr"`
	Type       string     `xml:"type,attr"`
	X          float64    `xml:"x,attr"`
	Y          float64    `xml:"y,attr"`
	Width      float64    `xml:"width,attr"`
	Height     float64    `xml:"height,attr"`
	GID        int        `xml:"gid,attr"`
	Visible    bool       `xml:"visible,attr"`
	Polygons   []Polygon  `xml:"polygon"`
	PolyLines  []PolyLine `xml:"polyline"`
	Properties []Property `xml:"properties>property"`
}

/*
   ___  _     _        _    ___
  / _ \| |__ (_)___ __| |_ / __|_ _ ___ _  _ _ __
 | (_) | '_ \| / -_) _|  _| (_ | '_/ _ \ || | '_ \
  \___/|_.__// \___\__|\__|\___|_| \___/\_,_| .__/
           |__/                             |_|
*/

type ObjectGroup struct {
	Name       string     `xml:"name,attr"`
	Color      string     `xml:"color,attr"`
	Opacity    float32    `xml:"opacity,attr"`
	Visible    bool       `xml:"visible,attr"`
	Properties []Property `xml:"properties>property"`
	Objects    []Object   `xml:"object"`
}

/*
  ___     _     _
 | _ \___(_)_ _| |_
 |  _/ _ \ | ' \  _|
 |_| \___/_|_||_\__|
*/

type Point struct {
	X int
	Y int
}

func decodePoints(s string) (points []Point, err error) {
	pointStrings := strings.Split(s, " ")

	points = make([]Point, len(pointStrings))
	for i, pointString := range pointStrings {
		coordStrings := strings.Split(pointString, ",")
		if len(coordStrings) != 2 {
			return []Point{}, InvalidPointsFieldError
		}

		points[i].X, err = strconv.Atoi(coordStrings[0])
		if err != nil {
			return []Point{}, err
		}

		points[i].Y, err = strconv.Atoi(coordStrings[1])
		if err != nil {
			return []Point{}, err
		}
	}
	return
}

/*
  ___     _
 | _ \___| |_  _ __ _ ___ _ _
 |  _/ _ \ | || / _` / _ \ ' \
 |_| \___/_|\_, \__, \___/_||_|
            |__/|___/
*/

type Polygon struct {
	Points string `xml:"points,attr"`
}

func (p *Polygon) Decode() ([]Point, error) {
	return decodePoints(p.Points)
}

/*
  ___     _      _ _
 | _ \___| |_  _| (_)_ _  ___
 |  _/ _ \ | || | | | ' \/ -_)
 |_| \___/_|\_, |_|_|_||_\___|
            |__/
*/

type PolyLine struct {
	Points string `xml:"points,attr"`
}

func (p *PolyLine) Decode() ([]Point, error) {
	return decodePoints(p.Points)
}

/*
  ___                       _
 | _ \_ _ ___ _ __  ___ _ _| |_ _  _
 |  _/ '_/ _ \ '_ \/ -_) '_|  _| || |
 |_| |_| \___/ .__/\___|_|  \__|\_, |
             |_|                |__/
*/

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

/*
  _____ _ _
 |_   _(_) |___
   | | | | / -_)
   |_| |_|_\___|
*/

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

/*
  _____ _ _             _
 |_   _(_) |___ ___ ___| |_
   | | | | / -_|_-</ -_)  _|
   |_| |_|_\___/__/\___|\__|
*/

type Tileset struct {
	FirstGID   GID        `xml:"firstgid,attr"`
	Source     string     `xml:"source,attr"`
	Name       string     `xml:"name,attr"`
	TileWidth  int        `xml:"tilewidth,attr"`
	TileHeight int        `xml:"tileheight,attr"`
	Spacing    int        `xml:"spacing,attr"`
	Margin     int        `xml:"margin,attr"`
	Properties []Property `xml:"properties>property"`
	Image      Image      `xml:"image"`
	Tiles      []Tile     `xml:"tile"`
	Tilecount  int        `xml:"tilecount,attr"`
	Columns    int        `xml:"columns,attr"`

	sprite *pixel.Sprite
}

func getTileset(l *Layer) (tileset *Tileset, isEmpty, usesMultipleTilesets bool) {
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
