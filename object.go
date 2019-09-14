package tilepix

import (
	"fmt"

	"github.com/faiface/pixel"

	log "github.com/sirupsen/logrus"
)

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
	GID        ID          `xml:"gid,attr"`
	ID         ID          `xml:"id,attr"`
	Visible    bool        `xml:"visible,attr"`
	Polygon    *Polygon    `xml:"polygon"`
	PolyLine   *PolyLine   `xml:"polyline"`
	Properties []*Property `xml:"properties>property"`
	Ellipse    *struct{}   `xml:"ellipse"`
	Point      *struct{}   `xml:"point"`

	objectType ObjectType
	tile       *DecodedTile

	// parentMap is the map which contains this object
	parentMap *Map
}

// GetEllipse will return a pixel.Circle representation of this object relative to the map (the co-ordinates will match
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

// GetPoint will return a pixel.Vec representation of this object relative to the map (the co-ordinates will match those
// as drawn in Tiled).  If the object type is not `PointObj` this function will return `pixel.ZV` and an error.
func (o *Object) GetPoint() (pixel.Vec, error) {
	if o.GetType() != PointObj {
		log.WithError(ErrInvalidObjectType).WithField("Object type", o.GetType()).Error("Object.GetPoint: object type mismatch")
		return pixel.ZV, ErrInvalidObjectType
	}

	return pixel.V(o.X, o.Y), nil
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

// GetPolygon will return a pixel.Vec slice representation of this object relative to the map (the co-ordinates will match
// those as drawn in Tiled).  If the object type is not `PolygonObj` this function will return `nil` and an error.
func (o *Object) GetPolygon() ([]pixel.Vec, error) {
	if o.GetType() != PolygonObj {
		log.WithError(ErrInvalidObjectType).WithField("Object type", o.GetType()).Error("Object.GetPolygon: object type mismatch")
		return nil, ErrInvalidObjectType
	}

	points, err := o.Polygon.Decode()
	if err != nil {
		log.WithError(err).Error("Object.GetPolygon: could not decode Polygon")
		return nil, err
	}

	var pixelPoints []pixel.Vec
	for _, p := range points {
		pixelPoints = append(pixelPoints, p.V())
	}

	return pixelPoints, nil
}

// GetPolyLine will return a pixel.Vec slice representation of this object relative to the map (the co-ordinates will match
// those as drawn in Tiled).  If the object type is not `PolylineObj` this function will return `nil` and an error.
func (o *Object) GetPolyLine() ([]pixel.Vec, error) {
	if o.GetType() != PolylineObj {
		log.WithError(ErrInvalidObjectType).WithField("Object type", o.GetType()).Error("Object.GetPolyLine: object type mismatch")
		return nil, ErrInvalidObjectType
	}

	points, err := o.PolyLine.Decode()
	if err != nil {
		log.WithError(err).Error("Object.GetPolyLine: could not decode Polyline")
		return nil, err
	}

	var pixelPoints []pixel.Vec
	for _, p := range points {
		pixelPoints = append(pixelPoints, p.V())
	}

	return pixelPoints, nil
}

// GetTile will return the object decoded into a DecodedTile struct.  If this
// object is not a DecodedTile, this function will return `nil` and an error.
func (o *Object) GetTile() (*DecodedTile, error) {
	if o.GetType() != TileObj {
		log.WithError(ErrInvalidObjectType).WithField("Object type", o.GetType()).Error("Object.GetTile: object type mismatch")
		return nil, ErrInvalidObjectType
	}

	if o.tile == nil {
		// Setting tileset to the first tileset in the map.  Will need updating when dealing with multiple
		// tilesets.
		ts := o.parentMap.Tilesets[0]

		o.tile = &DecodedTile{
			ID:        o.GID,
			Tileset:   ts,
			parentMap: o.parentMap,
		}

		numRows := ts.Tilecount / ts.Columns
		o.tile.setSprite(ts.Columns, numRows, ts)
	}

	return o.tile, nil
}

// GetType will return the ObjectType constant type of this object.
func (o *Object) GetType() ObjectType {
	return o.objectType
}

func (o *Object) String() string {
	return fmt.Sprintf("Object{%s, Name: '%s'}", o.objectType, o.Name)
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

	if o.GID != 0 {
		o.objectType = TileObj
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
