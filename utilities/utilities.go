package utilities

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

func LoadSprite(img io.Reader) (*pixel.Sprite, pixel.Picture, error) {
	pic, err := loadPicture(img)
	if err != nil {
		return nil, nil, err
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())
	return sprite, pic, nil
}

func LoadSpriteFromFile(path string) (*pixel.Sprite, pixel.Picture, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0444)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	return LoadSprite(f)
}

func TileIDToCoord(tID int, numColumns int, numRows int) (x int, y int) {
	x = tID % numColumns
	y = numRows - (tID / numColumns) - 1
	return
}

func IndexToGamePos(idx int, width int, height int) pixel.Vec {
	gamePos := pixel.V(
		float64(idx%width)-1,
		float64(height)-float64(idx/width),
	)
	return gamePos
}
