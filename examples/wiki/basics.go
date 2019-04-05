package main

import (
	"image/color"

	_ "image/png"

	"github.com/bcvery1/tilepix"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var (
	winBounds = pixel.R(0, 0, 800, 600)
)

func run() {
	m, err := tilepix.ReadFile("basic.tmx")
	if err != nil {
		panic(err)
	}

	cfg := pixelgl.WindowConfig{
		Title:  "TilePix basics",
		Bounds: winBounds,
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	for !win.Closed() {
		win.Clear(color.White)

		m.DrawAll(win, color.Transparent, pixel.IM)

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
