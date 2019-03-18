package tilepix_test

import (
	"image/color"

	_ "image/png"

	"github.com/bcvery1/tilepix"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func run() {
	m, err := tilepix.ReadFile("testdata/basic.tmx")
	if err != nil {
		panic(err)
	}

	cfg := pixelgl.WindowConfig{
		Title:  "Example",
		Bounds: pixel.R(0, 0, 100, 100),
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// This would normally be run within a `for !win.Closed() {` loop
	if err := m.DrawAll(win, color.White, pixel.IM); err != nil {
		panic(err)
	}
}

func Example() {
	pixelgl.Run(run)
	// Output:
}
