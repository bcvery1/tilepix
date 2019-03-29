package tilepix

import (
	"fmt"

	"github.com/faiface/pixel"

	log "github.com/sirupsen/logrus"
)

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
	// Shift image by layer offset.
	mat = mat.Moved(pixel.V(float64(im.Image.Width/2), float64(im.Image.Height/-2))).Moved(pixel.V(im.OffSetX, -im.OffSetY))

	im.Image.sprite.Draw(target, mat)
	return nil
}

func (im *ImageLayer) String() string {
	return fmt.Sprintf("ImageLayer{Name: '%s', Image: %s}", im.Name, im.Image)
}

func (im *ImageLayer) setParent(m *Map) {
	im.parentMap = m

	if im.Image != nil {
		im.Image.setParent(m)
	}
}
