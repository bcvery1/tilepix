package tilepix

import (
	"image"
	"io"
	"os"

	"github.com/faiface/pixel"
)

// loadPicture loads picture data from a Reader and will decode based using the built in image
// package.
func loadPicture(img io.Reader) (pixel.Picture, error) {
	imgDecoded, _, err := image.Decode(img)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(imgDecoded), nil
}

func loadPictureFromFile(path string) (pixel.Picture, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0444)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return loadPicture(f)
}

func loadSprite(img io.Reader) (*pixel.Sprite, error) {
	pic, err := loadPicture(img)
	if err != nil {
		return nil, err
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())
	return sprite, nil
}

func loadSpriteFromFile(path string) (*pixel.Sprite, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0444)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return loadSprite(f)
}
