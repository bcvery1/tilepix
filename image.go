package tilepix

import (
	"fmt"
	"path/filepath"

	"github.com/faiface/pixel"

	log "github.com/sirupsen/logrus"
)

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

func (i *Image) String() string {
	return fmt.Sprintf("Image{Source: %s, Size: %dx%d}", i.Source, i.Width, i.Height)
}

func (i *Image) initSprite() error {
	if i.sprite != nil {
		return nil
	}

	log.WithFields(log.Fields{"Path": i.Source, "Width": i.Width, "Height": i.Height}).Debug("Image.initSprite: loading sprite")

	sprite, pictureData, err := loadSpriteFromFile(filepath.Join(i.parentMap.dir, i.Source))
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
