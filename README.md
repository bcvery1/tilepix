[![Build Status](https://travis-ci.org/bcvery1/tilepix.svg?branch=master)](https://travis-ci.org/bcvery1/tilepix)
[![Go Report Card](https://goreportcard.com/badge/github.com/bcvery1/tilepix)](https://goreportcard.com/report/github.com/bcvery1/tilepix)
[![GitHub](https://img.shields.io/github/license/bcvery1/tilepix.svg)](https://github.com/bcvery1/tilepix/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/bcvery1/tilepix?status.svg)](https://godoc.org/github.com/bcvery1/tilepix)

# ![TilePixLogo](https://github.com/bcvery1/tilepix/blob/master/.github/assets/logo_small.png) TilePix
TilePix is a complementary library, designed to be used with the [Pixel](https://github.com/faiface/pixel) library (see
below for more details on Pixel).  TilePix was born out of [Ludum Dare](https://ldjam.com/); having found that a vast
amount of very limited time during the Ludum Dare weekends was used planing out map layouts, and defining collision
or activation areas.  TilePix should make those activities a trivially short amount of time.

## Pixel
This library is a complement to the [Pixel](https://github.com/faiface/pixel) 2D games library, and is largely inspired
by it.  A huge thanks to [faiface](https://github.com/faiface) for one, giving us access to such a fantastic library;
and two, providing the inspiration for this library!

TilePix would not have been possible without the great amount of care and effort that has been put into
[Pixel](https://github.com/faiface/pixel).

### Legal
Pixel is subject to the [MIT](https://github.com/faiface/pixel/blob/master/LICENSE) licence.

## Stability
TilePix is a work-in-progress project; as such, expect bugs and missing features.  If you notice a bug or a feature you
feel is missing, please consider [contributing](https://github.com/bcvery1/tilepix/blob/master/CONTRIBUTING.md) - simply
(and correctly) raising issues is just as valuable as writing code!

### Releases
The aim is that releases on this library will fairly regular, and well planned.  You can use
[Go modules](https://github.com/golang/go/wiki/Modules) with TilePix if you want version security.

## Example
Here is a very basic example of using the library.  It is advisable to view the excellent
[Pixel tutorials](https://github.com/faiface/pixel/wiki) before trying to understand this package, as TilePix is very
Pixel centric.

```go
package main

import (
	"image/color"

	// We must use blank imports for any image formats in the tileset image sources.
	// You will get an error if a blank import is not made; TilePix does not import
	// specific image formats, that is the responsibility of the calling code.
	_ "image/png"

	"github.com/bcvery1/tilepix"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title: "TilePix",
		Bounds: pixel.R(0, 0, 640, 320),
		VSync: true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Load and initialise the map.
	m, err := tilepix.ReadFile("myMap.tmx")
	if err != nil {
		panic(err)
	}

	for !win.Closed() {
		win.Clear(color.White)

		// Draw all layers to the window.
		if err := m.DrawAll(win, color.White, pixel.IM); err != nil {
			panic(err)
		}

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
```

Futher examples can be found in the [examples directory](https://github.com/bcvery1/tilepix/tree/master/examples).

## Contributing
Thanks for considering contributing to TilePix; for details on how you can contribute, please consult the
[contribution guide](https://github.com/bcvery1/tilepix/blob/master/CONTRIBUTING.md).
