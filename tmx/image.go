package tmx

import "github.com/faiface/pixel"

type Image struct {
	Source string `xml:"source,attr"`
	Trans  string `xml:"trans,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`

	sprite  *pixel.Sprite
	picture pixel.Picture
}
