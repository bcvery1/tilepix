package main

import (
	"image/color"
	"math"
	"time"

	"github.com/bcvery1/tilepix"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var (
	winBounds    = pixel.R(0, 0, 800, 600)
	camPos       = pixel.ZV
	camSpeed     = 64.0
	camZoom      = 1.0
	camZoomSpeed = 1.2
)

func run() {
	m, err := tilepix.ReadFile("map.tmx")
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

	last := time.Now()
	for !win.Closed() {
		win.Clear(color.White)

		dt := time.Since(last).Seconds()
		last = time.Now()

		cam := pixel.IM.Scaled(camPos.Add(winBounds.Center()), camZoom).Moved(pixel.ZV.Sub(camPos))
		win.SetMatrix(cam)

		if win.Pressed(pixelgl.KeyA) || win.Pressed(pixelgl.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyD) || win.Pressed(pixelgl.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyS) || win.Pressed(pixelgl.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyW) || win.Pressed(pixelgl.KeyUp) {
			camPos.Y += camSpeed * dt
		}
		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		m.DrawAll(win, color.Transparent, pixel.IM) // nolint: errcheck

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
