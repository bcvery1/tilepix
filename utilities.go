package tilepix

import (
	"image"
	"io"
	"os"

	"github.com/faiface/pixel"

	log "github.com/sirupsen/logrus"
)

// loadPicture loads picture data from a Reader and will decode based using the built in image
// package.
func loadPicture(img io.Reader) (pixel.Picture, error) {
	imgDecoded, _, err := image.Decode(img)
	if err != nil {
		log.WithError(err).Error("loadPicture: could not decode image")
		return nil, err
	}
	return pixel.PictureDataFromImage(imgDecoded), nil
}

func loadSprite(img io.Reader) (*pixel.Sprite, pixel.Picture, error) {
	pic, err := loadPicture(img)
	if err != nil {
		log.WithError(err).Error("loadSprite: could not load picture")
		return nil, nil, err
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())
	return sprite, pic, nil
}

func loadSpriteFromFile(path string) (*pixel.Sprite, pixel.Picture, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0444)
	if err != nil {
		log.WithError(err).WithField("Filepath", path).Error("loadSpriteFromFile: could not open file")
		return nil, nil, err
	}
	defer f.Close()

	return loadSprite(f)
}

func tileIDToCoord(tID ID, numColumns int, numRows int) (x int, y int) {
	tIDInt := int(tID)
	x = tIDInt % numColumns
	y = numRows - (tIDInt / numColumns) - 1
	return
}

func indexToGamePos(idx int, width int, height int) pixel.Vec {
	gamePos := pixel.V(
		float64(idx%width),
		float64(height)-float64(idx/width)-1,
	)
	return gamePos
}
