package tmx

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
}

type Image struct {
	Source string `xml:"source,attr"`
	Trans  string `xml:"trans,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

type Tile struct {
	ID    ID    `xml:"id,attr"`
	Image Image `xml:"image"`
}

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

type Data struct {
	Encoding    string     `xml:"encoding,attr"`
	Compression string     `xml:"compression,attr"`
	RawData     []byte     `xml:",innerxml"`
	// DataTiles is only used when layer encoding is XML.
	DataTiles   []DataTile `xml:"tile"`
}

type ObjectGroup struct {
	Name       string     `xml:"name,attr"`
	Color      string     `xml:"color,attr"`
	Opacity    float32    `xml:"opacity,attr"`
	Visible    bool       `xml:"visible,attr"`
	Properties []Property `xml:"properties>property"`
	Objects    []Object   `xml:"object"`
}

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

type Polygon struct {
	Points string `xml:"points,attr"`
}

type PolyLine struct {
	Points string `xml:"points,attr"`
}

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
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
			return []GID{}, err
		}
		gids[i] = GID(d)
	}
	return gids, nil
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

type Point struct {
	X int
	Y int
}

type DataTile struct {
	GID GID `xml:"gid,attr"`
}

func (p *Polygon) Decode() ([]Point, error) {
	return decodePoints(p.Points)
}
func (p *PolyLine) Decode() ([]Point, error) {
	return decodePoints(p.Points)
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

func getTileset(_ *Map, l *Layer) (tileset *Tileset, isEmpty, usesMultipleTilesets bool) {
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
		l := &m.Layers[i]

		tileset, isEmpty, usesMultipleTilesets := getTileset(m, l)
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
