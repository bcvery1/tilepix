package tilepix

import "fmt"

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

func (p *Property) String() string {
	return fmt.Sprintf("Property{%s: %s}", p.Name, p.Value)
}

func (p *Property) setParent(m *Map) {
	p.parentMap = m
}
