package tmx

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
)

const (
	GIDHorizontalFlip = 0x80000000
	GIDVerticalFlip   = 0x40000000
	GIDDiagonalFlip   = 0x20000000
	GIDFlip           = GIDHorizontalFlip | GIDVerticalFlip | GIDDiagonalFlip
)

var (
	UnknownEncodingError       = errors.New("tmx: invalid encoding scheme")
	UnknownCompressionError    = errors.New("tmx: invalid compression method")
	InvalidDecodedDataLenError = errors.New("tmx: invalid decoded data length")
	InvalidGIDError            = errors.New("tmx: invalid GID")
	InvalidPointsFieldError    = errors.New("tmx: invalid points string")
)

var (
	NilTile = &DecodedTile{Nil: true}
)

type GID uint32 // A tile ID. Could be used for GID or ID.
type ID uint32

type DataTile struct {
	GID GID `xml:"gid,attr"`
}

func Read(r io.Reader) (*Map, error) {
	d := xml.NewDecoder(r)

	m := new(Map)
	if err := d.Decode(m); err != nil {
		return nil, err
	}

	if err := m.decodeLayers(); err != nil {
		return nil, err
	}

	for i := 0; i < len(m.Layers); i++ {
		l := &m.Layers[i]
		l.mapParent = m

		tileset, isEmpty, usesMultipleTilesets := getTileset(m, l)
		if usesMultipleTilesets {
			continue
		}
		l.Empty, l.Tileset = isEmpty, tileset
	}

	return m, nil
}

func ReadFile(filePath string) (*Map, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return Read(f)
}
