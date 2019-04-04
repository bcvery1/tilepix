package tilepix

import "fmt"

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
	OffSetX    float64     `xml:"offsetx,attr"`
	OffSetY    float64     `xml:"offsety,attr"`
	Opacity    float32     `xml:"opacity,attr"`
	Visible    bool        `xml:"visible,attr"`
	Properties []*Property `xml:"properties>property"`
	Objects    []*Object   `xml:"object"`

	// parentMap is the map which contains this object
	parentMap *Map
}

func (og *ObjectGroup) String() string {
	return fmt.Sprintf("ObjectGroup{Name: %s, Properties: %v, Objects: %v}", og.Name, og.Properties, og.Objects)
}

func (og *ObjectGroup) decode() error {
	for _, o := range og.Objects {
		// Have the object decode its' type
		o.hydrateType()

		// Set the x,y offsets of the layer onto the object
		o.X += og.OffSetX
		o.Y -= og.OffSetY
	}

	return nil
}

// GetObjectByName returns the ObjectGroups' Objects by their name
func (og *ObjectGroup) GetObjectByName(name string) []*Object {
	var objs []*Object

	for _, o := range og.Objects {
		if o.Name == name {
			objs = append(objs, o)
		}
	}
	return objs
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
