package tilepix

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"image/color"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	log "github.com/sirupsen/logrus"
)

const (
	gidHorizontalFlip = 0x80000000
	gidVerticalFlip   = 0x40000000
	gidDiagonalFlip   = 0x20000000
	gidFlip           = gidHorizontalFlip | gidVerticalFlip | gidDiagonalFlip
)

// ObjectType is used to represent the types an object can be.
type ObjectType int

// These are the currently supported object types.
const (
	EllipseObj ObjectType = iota
	PolygonObj
	PolylineObj
	RectangleObj
	PointObj
)

// Errors which are returned from various places in the package.
var (
	ErrUnknownEncoding       = errors.New("tmx: invalid encoding scheme")
	ErrUnknownCompression    = errors.New("tmx: invalid compression method")
	ErrInvalidDecodedDataLen = errors.New("tmx: invalid decoded data length")
	ErrInvalidGID            = errors.New("tmx: invalid GID")
	ErrInvalidObjectType     = errors.New("tmx: the object type requested does not match this object")
	ErrInvalidPointsField    = errors.New("tmx: invalid points string")
	ErrInfiniteMap           = errors.New("tmx: infinite maps are not currently supported")
)

var (
	// NilTile is a tile with no tile set.  Will be skipped over when drawing.
	NilTile = &DecodedTile{Nil: true}
)

// GID is a global tile ID. Tiles can use GID or ID.
type GID uint32

// ID is a tile ID. Tiles can use GID or ID.
type ID uint32

// DataTile is a tile from a data object.
type DataTile struct {
	GID GID `xml:"gid,attr"`
}

// Read will read, decode and initialise a Tiled Map from a data reader.
func Read(r io.Reader) (*Map, error) {
	log.Debug("Read: reading from io.Reader")

	d := xml.NewDecoder(r)

	m := new(Map)
	if err := d.Decode(m); err != nil {
		log.WithError(err).Error("Read: could not decode to Map")
		return nil, err
	}

	if m.Infinite {
		log.WithError(ErrInfiniteMap).Error("Read: map has attribute 'infinite=true', not supported")
		return nil, ErrInfiniteMap
	}

	if err := m.decodeLayers(); err != nil {
		log.WithError(err).Error("Read: could not decode layers")
		return nil, err
	}

	m.setParents()

	log.WithField("Layer count", len(m.Layers)).Debug("Read: processing layer tilesets")
	for _, l := range m.Layers {
		tileset, isEmpty, usesMultipleTilesets := getTileset(l)
		if usesMultipleTilesets {
			log.Debug("Read: multiple tilesets in use")
			continue
		}
		l.Empty, l.Tileset = isEmpty, tileset
	}

	// Tiled calculates co-ordinates from the top-left, flipping the y co-ordinate means we match the standard
	// bottom-left calculation.
	log.WithField("Object layer count", len(m.ObjectGroups)).Debug("Read: processing object layers")
	for _, og := range m.ObjectGroups {
		og.flipY()
	}

	return m, nil
}

// ReadFile will read, decode and initialise a Tiled Map from a file path.
func ReadFile(filePath string) (*Map, error) {
	log.WithField("Filepath", filePath).Debug("ReadFile: reading file")

	f, err := os.Open(filePath)
	if err != nil {
		log.WithError(err).Error("ReadFile: could not open file")
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

// Data is a TMX file structure holding data.
type Data struct {
	Encoding    string `xml:"encoding,attr"`
	Compression string `xml:"compression,attr"`
	RawData     []byte `xml:",innerxml"`
	// DataTiles is only used when layer encoding is XML.
	DataTiles []*DataTile `xml:"tile"`

	// parentMap is the map which contains this object
	parentMap *Map
}

func (d *Data) decodeBase64() (data []byte, err error) {
	rawData := bytes.TrimSpace(d.RawData)
	r := bytes.NewReader(rawData)

	encr := base64.NewDecoder(base64.StdEncoding, r)

	var comr io.Reader
	switch d.Compression {
	case "gzip":
		log.Debug("decodeBase64: compression is gzip")

		comr, err = gzip.NewReader(encr)
		if err != nil {
			return
		}
	case "zlib":
		log.Debug("decodeBase64: compression is zlib")

		comr, err = zlib.NewReader(encr)
		if err != nil {
			return
		}
	case "":
		log.Debug("decodeBase64: no compression")

		comr = encr
	default:
		err = ErrUnknownCompression
		log.WithError(ErrUnknownCompression).WithField("Compression", d.Compression).Error("decodeBase64: unable to handle this compression type")
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
			log.WithError(err).WithField("String to convert", s).Error("decodeCSV: could not parse UInt")
			return nil, err
		}
		gids[i] = GID(d)
	}
	return gids, nil
}

func (d *Data) setParent(m *Map) {
	d.parentMap = m
}

/*
  ___
 |_ _|_ __  __ _ __ _ ___
  | || '  \/ _` / _` / -_)
 |___|_|_|_\__,_\__, \___|
                |___/
*/

// Image is a TMX file structure which referencing an image file, with associated properies.
type Image struct {
	Source string `xml:"source,attr"`
	Trans  string `xml:"trans,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`

	sprite  *pixel.Sprite
	picture pixel.Picture

	// parentMap is the map which contains this object
	parentMap *Map
}

func (i *Image) initSprite() error {
	if i.sprite != nil {
		return nil
	}

	log.WithFields(log.Fields{"Path": i.Source, "Width": i.Width, "Height": i.Height}).Debug("Image.initSprite: loading sprite")

	// TODO(need to do this either by file or reader)
	sprite, pictureData, err := loadSpriteFromFile(i.Source)
	if err != nil {
		log.WithError(err).Error("Image.initSprite: could not load sprite from file")
		return err
	}

	i.sprite = sprite
	i.picture = pictureData

	return nil
}

func (i *Image) setParent(m *Map) {
	i.parentMap = m
}

/*
  ___                     _
 |_ _|_ __  __ _ __ _ ___| |   __ _ _  _ ___ _ _
  | || '  \/ _` / _` / -_) |__/ _` | || / -_) '_|
 |___|_|_|_\__,_\__, \___|____\__,_|\_, \___|_|
                |___/               |__/
*/

// ImageLayer is a TMX file structure which references an image layer, with associated properties.
type ImageLayer struct {
	Locked  bool    `xml:"locked,attr"`
	Name    string  `xml:"name,attr"`
	OffSetX float64 `xml:"offsetx,attr"`
	OffSetY float64 `xml:"offsety,attr"`
	Opacity float64 `xml:"opacity,attr"`
	Image   *Image  `xml:"image"`

	// parentMap is the map which contains this object
	parentMap *Map
}

// Draw will draw the image layer to the target provided, shifted with the provided matrix.
func (im *ImageLayer) Draw(target pixel.Target, mat pixel.Matrix) error {
	if err := im.Image.initSprite(); err != nil {
		log.WithError(err).Error("ImageLayer.Draw: could not initialise image sprite")
		return err
	}

	// Shift image right-down by half its' dimensions.
	mat = mat.Moved(pixel.V(float64(im.Image.Width/2), float64(im.Image.Height/-2)))

	im.Image.sprite.Draw(target, mat)
	return nil
}

func (im *ImageLayer) setParent(m *Map) {
	im.parentMap = m

	if im.Image != nil {
		im.Image.setParent(m)
	}
}

/*
  _
 | |   __ _ _  _ ___ _ _
 | |__/ _` | || / -_) '_|
 |____\__,_|\_, \___|_|
            |__/
*/

// Layer is a TMX file structure which can hold any type of Tiled layer.
type Layer struct {
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
func (l *Layer) Batch() (*pixel.Batch, error) {
	if l.batch == nil {
		log.Debug("Layer.Batch: batch not initialised, creating")

		if l.Tileset == nil {
			err := errors.New("cannot create sprite from nil tileset")
			log.WithError(err).Error("Layer.Batch: layers' tileset is nil")
			return nil, err
		}

		// TODO(need to do this either by file or reader)
		sprite, pictureData, err := loadSpriteFromFile(l.Tileset.Image.Source)
		if err != nil {
			log.WithError(err).Error("Layer.Batch: could not load sprite from file")
			return nil, err
		}

		l.batch = pixel.NewBatch(&pixel.TrianglesData{}, pictureData)
		l.Tileset.sprite = sprite
	}

	l.batch.Clear()

	return l.batch, nil
}

// Draw will use the Layers' batch to draw all tiles within the Layer to the target.
func (l *Layer) Draw(target pixel.Target) error {
	// Initialise the batch
	if _, err := l.Batch(); err != nil {
		log.WithError(err).Error("Layer.Draw: could not get batch")
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

func (l *Layer) decode(width, height int) ([]GID, error) {
	log.WithField("Encoding", l.Data.Encoding).Debug("Layer.decode: determining encoding")

	switch l.Data.Encoding {
	case "csv":
		return l.decodeLayerCSV(width, height)
	case "base64":
		return l.decodeLayerBase64(width, height)
	case "":
		// XML "encoding"
		return l.decodeLayerXML(width, height)
	}

	log.WithError(ErrUnknownEncoding).Error("Layer.decode: unrecognised encoding")
	return nil, ErrUnknownEncoding
}

func (l *Layer) decodeLayerXML(width, height int) ([]GID, error) {
	if len(l.Data.DataTiles) != width*height {
		log.WithError(ErrInvalidDecodedDataLen).WithFields(log.Fields{"Length datatiles": len(l.Data.DataTiles), "W*H": width * height}).Error("Layer.decodeLayerXML: data length mismatch")
		return nil, ErrInvalidDecodedDataLen
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
		log.WithError(err).Error("Layer.decodeLayerCSV: could not decode CSV")
		return nil, err
	}

	if len(gids) != width*height {
		log.WithError(ErrInvalidDecodedDataLen).WithFields(log.Fields{"Length GIDSs": len(gids), "W*H": width * height}).Error("Layer.decodeLayerCSV: data length mismatch")
		return nil, ErrInvalidDecodedDataLen
	}

	return gids, nil
}

func (l *Layer) decodeLayerBase64(width, height int) ([]GID, error) {
	dataBytes, err := l.Data.decodeBase64()
	if err != nil {
		log.WithError(err).Error("Layer.decodeLayerBase64: could not decode base64")
		return nil, err
	}

	if len(dataBytes) != width*height*4 {
		log.WithError(ErrInvalidDecodedDataLen).WithFields(log.Fields{"Length databytes": len(dataBytes), "W*H": width * height}).Error("Layer.decodeLayerBase64: data length mismatch")
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

func (l *Layer) setParent(m *Map) {
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

/*
  __  __
 |  \/  |__ _ _ __
 | |\/| / _` | '_ \
 |_|  |_\__,_| .__/
             |_|
*/

// Map is a TMX file structure representing the map as a whole.
type Map struct {
	Version     string `xml:"title,attr"`
	Orientation string `xml:"orientation,attr"`
	// Width is the number of tiles - not the width in pixels
	Width int `xml:"width,attr"`
	// Height is the number of tiles - not the height in pixels
	Height       int            `xml:"height,attr"`
	TileWidth    int            `xml:"tilewidth,attr"`
	TileHeight   int            `xml:"tileheight,attr"`
	Properties   []*Property    `xml:"properties>property"`
	Tilesets     []*Tileset     `xml:"tileset"`
	Layers       []*Layer       `xml:"layer"`
	ObjectGroups []*ObjectGroup `xml:"objectgroup"`
	Infinite     bool           `xml:"infinite,attr"`
	ImageLayers  []*ImageLayer  `xml:"imagelayer"`

	canvas *pixelgl.Canvas
}

// DrawAll will draw all tile layers and image layers to the target.
// Tile layers are first draw to their own `pixel.Batch`s for efficiency.
// All layers are drawn to a `pixel.Canvas` before being drawn to the target.
//
// - target - The target to draw layers to.
// - clearColour - The colour to clear the maps' canvas before drawing.
// - mat - The matrix to draw the canvas to the target with.
func (m *Map) DrawAll(target pixel.Target, clearColour color.Color, mat pixel.Matrix) error {
	if m.canvas == nil {
		m.canvas = pixelgl.NewCanvas(m.bounds())
	}
	m.canvas.Clear(clearColour)

	for _, l := range m.Layers {
		if err := l.Draw(m.canvas); err != nil {
			log.WithError(err).Error("Map.DrawAll: could not draw layer")
			return err
		}
	}

	for _, il := range m.ImageLayers {
		// The matrix shift is because images are drawn from the top-left in Tiled.
		if err := il.Draw(m.canvas, pixel.IM.Moved(pixel.V(0, m.pixelHeight()))); err != nil {
			log.WithError(err).Error("Map.DrawAll: could not draw image layer")
			return err
		}
	}

	m.canvas.Draw(target, mat.Moved(m.bounds().Center()))

	return nil
}

// bounds will return a pixel rectangle representing the width-height in pixels.
func (m *Map) bounds() pixel.Rect {
	return pixel.R(0, 0, m.pixelWidth(), m.pixelHeight())
}

func (m *Map) pixelWidth() float64 {
	return float64(m.Width * m.TileWidth)
}
func (m *Map) pixelHeight() float64 {
	return float64(m.Height * m.TileHeight)
}

func (m *Map) decodeGID(gid GID) (*DecodedTile, error) {
	if gid == 0 {
		return NilTile, nil
	}

	gidBare := gid &^ gidFlip

	for i := len(m.Tilesets) - 1; i >= 0; i-- {
		if m.Tilesets[i].FirstGID <= gidBare {
			return &DecodedTile{
				ID:             ID(gidBare - m.Tilesets[i].FirstGID),
				Tileset:        m.Tilesets[i],
				HorizontalFlip: gid&gidHorizontalFlip != 0,
				VerticalFlip:   gid&gidVerticalFlip != 0,
				DiagonalFlip:   gid&gidDiagonalFlip != 0,
				Nil:            false,
			}, nil
		}
	}

	log.WithError(ErrInvalidGID).Error("Map.decodeGID: GID is invalid")
	return nil, ErrInvalidGID
}

func (m *Map) decodeLayers() error {
	// Decode tile layers
	for _, l := range m.Layers {
		gids, err := l.decode(m.Width, m.Height)
		if err != nil {
			log.WithError(err).Error("Map.decodeLayers: could not decode layer")
			return err
		}

		l.DecodedTiles = make([]*DecodedTile, len(gids))
		for j := 0; j < len(gids); j++ {
			decTile, err := m.decodeGID(gids[j])
			if err != nil {
				log.WithError(err).Error("Map.decodeLayers: could not GID")
				return err
			}
			l.DecodedTiles[j] = decTile
		}
	}

	// Decode object layers
	for _, og := range m.ObjectGroups {
		if err := og.decode(); err != nil {
			log.WithError(err).Error("Map.decodeLayers: could not decode Object Group")
			return err
		}
	}

	return nil
}

func (m *Map) setParents() {
	for _, p := range m.Properties {
		p.setParent(m)
	}
	for _, t := range m.Tilesets {
		t.setParent(m)
	}
	for _, og := range m.ObjectGroups {
		og.setParent(m)
	}
	for _, im := range m.ImageLayers {
		im.setParent(m)
	}
	for _, l := range m.Layers {
		l.setParent(m)
	}
}

/*
   ___  _     _        _
  / _ \| |__ (_)___ __| |_
 | (_) | '_ \| / -_) _|  _|
  \___/|_.__// \___\__|\__|
           |__/
*/

// Object is a TMX file struture holding a specific Tiled object.
type Object struct {
	Name       string      `xml:"name,attr"`
	Type       string      `xml:"type,attr"`
	X          float64     `xml:"x,attr"`
	Y          float64     `xml:"y,attr"`
	Width      float64     `xml:"width,attr"`
	Height     float64     `xml:"height,attr"`
	GID        int         `xml:"id,attr"`
	Visible    bool        `xml:"visible,attr"`
	Polygon    *Polygon    `xml:"polygon"`
	PolyLine   *PolyLine   `xml:"polyline"`
	Properties []*Property `xml:"properties>property"`
	Ellipse    *struct{}   `xml:"ellipse"`
	Point      *struct{}   `xml:"point"`

	objectType ObjectType

	// parentMap is the map which contains this object
	parentMap *Map
}

// GetRect will return a pixel.Rect representation of this object relative to the map (the co-ordinates will match those
// as drawn in Tiled).  If the object type is not `RectangleObj` this function will return `pixel.R(0, 0, 0, 0)` and an
// error.
func (o *Object) GetRect() (pixel.Rect, error) {
	if o.GetType() != RectangleObj {
		log.WithError(ErrInvalidObjectType).WithField("Object type", o.GetType()).Error("Object.GetRect: object type mismatch")
		return pixel.R(0, 0, 0, 0), ErrInvalidObjectType
	}

	return pixel.R(o.X, o.Y, o.X+o.Width, o.Y+o.Height), nil
}

// GetRect will return a pixel.Circle representation of this object relative to the map (the co-ordinates will match
// those as drawn in Tiled).  If the object type is not `EllipseObj` this function will return `pixel.C(pixel.ZV, 0)`
// and an error.
//
// Because there is no pixel geometry code for irregular ellipses, this function will average the width and height of
// the ellipse object from the TMX file, and return a regular circle about the centre of the ellipse.
func (o *Object) GetEllipse() (pixel.Circle, error) {
	if o.GetType() != EllipseObj {
		log.WithError(ErrInvalidObjectType).WithField("Object type", o.GetType()).Error("Object.GetEllipse: object type mismatch")
		return pixel.C(pixel.ZV, 0), ErrInvalidObjectType
	}

	// In TMX files, ellipses are defined by the containing rectangle.  The X, Y positions are the bottom-left (after we
	// have flipped them).
	// Because Pixel does not support irregular ellipses, we take the average of width and height.
	radius := (o.Width + o.Height) / 4
	// The centre should be the same as the ellipses drawn in Tiled, this will make outputs more intuitive.
	centre := pixel.V(o.X+(o.Width/2), o.Y+(o.Height/2))

	return pixel.C(centre, radius), nil
}

// GetType will return the ObjectType constant type of this object.
func (o *Object) GetType() ObjectType {
	return o.objectType
}

func (o *Object) flipY() {
	o.Y = o.parentMap.pixelHeight() - o.Y - o.Height
}

// hydrateType will work out what type this object is.
func (o *Object) hydrateType() {
	if o.Polygon != nil {
		o.objectType = PolygonObj
		return
	}

	if o.PolyLine != nil {
		o.objectType = PolylineObj
		return
	}

	if o.Ellipse != nil {
		o.objectType = EllipseObj
		return
	}

	if o.Point != nil {
		o.objectType = PointObj
		return
	}

	o.objectType = RectangleObj
}

func (o *Object) setParent(m *Map) {
	o.parentMap = m

	if o.Polygon != nil {
		o.Polygon.setParent(m)
	}
	if o.PolyLine != nil {
		o.PolyLine.setParent(m)
	}
	for _, p := range o.Properties {
		p.setParent(m)
	}
}

/*
   ___  _     _        _    ___
  / _ \| |__ (_)___ __| |_ / __|_ _ ___ _  _ _ __
 | (_) | '_ \| / -_) _|  _| (_ | '_/ _ \ || | '_ \
  \___/|_.__// \___\__|\__|\___|_| \___/\_,_| .__/
           |__/                             |_|
*/

// ObjectGroup is a TMX file structure holding a Tiled ObjectGroup.
type ObjectGroup struct {
	Name       string      `xml:"name,attr"`
	Color      string      `xml:"color,attr"`
	Opacity    float32     `xml:"opacity,attr"`
	Visible    bool        `xml:"visible,attr"`
	Properties []*Property `xml:"properties>property"`
	Objects    []*Object   `xml:"object"`

	// parentMap is the map which contains this object
	parentMap *Map
}

func (og *ObjectGroup) decode() error {
	for _, o := range og.Objects {
		o.hydrateType()
	}

	return nil
}

func (og *ObjectGroup) flipY() {
	for _, o := range og.Objects {
		o.flipY()
	}
}

func (og *ObjectGroup) setParent(m *Map) {
	og.parentMap = m

	for _, p := range og.Properties {
		p.setParent(m)
	}
	for _, o := range og.Objects {
		o.setParent(m)
	}
}

/*
  ___     _     _
 | _ \___(_)_ _| |_
 |  _/ _ \ | ' \  _|
 |_| \___/_|_||_\__|
*/

// Point is a TMX file structure holding a Tiled Point object.
type Point struct {
	X int
	Y int

	// parentMap is the map which contains this object
	parentMap *Map
}

// V converts the Tiled Point to a Pixel Vector.
func (p *Point) V() pixel.Vec {
	return pixel.V(float64(p.X), float64(p.Y))
}

func (p *Point) setParent(m *Map) {
	p.parentMap = m
}

func decodePoints(s string) (points []*Point, err error) {
	pointStrings := strings.Split(s, " ")

	points = make([]*Point, len(pointStrings))
	for i, pointString := range pointStrings {
		coordStrings := strings.Split(pointString, ",")
		if len(coordStrings) != 2 {
			log.WithError(ErrInvalidPointsField).WithField("Co-ordinate strings", coordStrings).Error("decodePoints: mismatch co-ordinates string length")
			return nil, ErrInvalidPointsField
		}

		points[i].X, err = strconv.Atoi(coordStrings[0])
		if err != nil {
			log.WithError(err).WithField("Point string", coordStrings[0]).Error("decodePoints: could not parse X co-ordinate string")
			return nil, err
		}

		points[i].Y, err = strconv.Atoi(coordStrings[1])
		if err != nil {
			log.WithError(err).WithField("Point string", coordStrings[1]).Error("decodePoints: could not parse X co-ordinate string")
			return nil, err
		}
	}
	return
}

// GetLayerByName returns a Map's Layer by its name
func (m *Map) GetLayerByName(name string) *Layer {
	for _, l := range m.Layers {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// GetImageLayerByName returns a Map's ImageLayer by its name
func (m *Map) GetImageLayerByName(name string) *ImageLayer {
	for _, l := range m.ImageLayers {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// GetObjectLayerByName returns a Map's ObjectGroup by its name
func (m *Map) GetObjectLayerByName(name string) *ObjectGroup {
	for _, l := range m.ObjectGroups {
		if l.Name == name {
			return l
		}
	}
	return nil
}

/*
  ___     _
 | _ \___| |_  _ __ _ ___ _ _
 |  _/ _ \ | || / _` / _ \ ' \
 |_| \___/_|\_, \__, \___/_||_|
            |__/|___/
*/

// Polygon is a TMX file structure representing a Tiled Polygon.
type Polygon struct {
	Points string `xml:"points,attr"`

	decodedPoints []*Point

	// parentMap is the map which contains this object
	parentMap *Map
}

// Decode will return a slice of points which make up this polygon.
func (p *Polygon) Decode() ([]*Point, error) {
	return decodePoints(p.Points)
}

func (p *Polygon) setParent(m *Map) {
	p.parentMap = m

	for _, dp := range p.decodedPoints {
		dp.setParent(m)
	}
}

/*
  ___     _      _ _
 | _ \___| |_  _| (_)_ _  ___
 |  _/ _ \ | || | | | ' \/ -_)
 |_| \___/_|\_, |_|_|_||_\___|
            |__/
*/

// PolyLine is a TMX file structure representing a Tiled Polyline.
type PolyLine struct {
	Points string `xml:"points,attr"`

	decodedPoints []*Point

	// parentMap is the map which contains this object
	parentMap *Map
}

// Decode will return a slice of points which make up this polyline.
func (p *PolyLine) Decode() ([]*Point, error) {
	return decodePoints(p.Points)
}

func (p *PolyLine) setParent(m *Map) {
	p.parentMap = m

	for _, dp := range p.decodedPoints {
		dp.setParent(m)
	}
}

/*
  ___                       _
 | _ \_ _ ___ _ __  ___ _ _| |_ _  _
 |  _/ '_/ _ \ '_ \/ -_) '_|  _| || |
 |_| |_| \___/ .__/\___|_|  \__|\_, |
             |_|                |__/
*/

// Property is a TMX file structure which holds a Tiled property.
type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`

	// parentMap is the map which contains this object
	parentMap *Map
}

func (p *Property) setParent(m *Map) {
	p.parentMap = m
}

/*
  _____ _ _
 |_   _(_) |___
   | | | | / -_)
   |_| |_|_\___|
*/

// Tile is a TMX file structure which holds a Tiled tile.
type Tile struct {
	ID    ID     `xml:"id,attr"`
	Image *Image `xml:"image"`

	// parentMap is the map which contains this object
	parentMap *Map
}

func (t *Tile) setParent(m *Map) {
	t.parentMap = m

	if t.Image != nil {
		t.Image.setParent(m)
	}
}

// DecodedTile is a convenience struct, which stores the decoded data from a Tile.
type DecodedTile struct {
	ID             ID
	Tileset        *Tileset
	HorizontalFlip bool
	VerticalFlip   bool
	DiagonalFlip   bool
	Nil            bool

	sprite *pixel.Sprite
	pos    pixel.Vec

	// parentMap is the map which contains this object
	parentMap *Map
}

// Draw will draw the tile to the target provided.  This will calculate the sprite from the provided tileset and set the
// DecodedTiles' internal `sprite` property; this is so it is only calculated the first time.
func (t *DecodedTile) Draw(ind, columns, numRows int, ts *Tileset, target pixel.Target) {
	if t.IsNil() {
		return
	}

	if t.sprite == nil {
		// Calculate the framing for the tile within its tileset's source image
		x, y := tileIDToCoord(t.ID, columns, numRows)
		gamePos := indexToGamePos(ind, t.parentMap.Width, t.parentMap.Height)

		iX := float64(x) * float64(ts.TileWidth)
		fX := iX + float64(ts.TileWidth)
		iY := float64(y) * float64(ts.TileHeight)
		fY := iY + float64(ts.TileHeight)

		t.sprite = pixel.NewSprite(ts.sprite.Picture(), pixel.R(iX, iY, fX, fY))
		t.pos = gamePos.ScaledXY(pixel.V(float64(ts.TileWidth), float64(ts.TileHeight))).Add(pixel.V(float64(ts.TileWidth), float64(ts.TileHeight)).Scaled(0.5))
	}
	t.sprite.Draw(target, pixel.IM.Moved(t.pos))
}

// IsNil returns whether this tile is nil.  If so, it means there is nothing set for the tile, and should be skipped in
// drawing.
func (t *DecodedTile) IsNil() bool {
	return t.Nil
}

func (t *DecodedTile) setParent(m *Map) {
	t.parentMap = m
}

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
