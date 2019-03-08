package tmx

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}
