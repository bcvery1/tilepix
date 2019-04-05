package tilepix_test

import (
	"image/color"
	"testing"

	_ "image/png"

	"github.com/bcvery1/tilepix"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func TestGetLayerByName(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}
	layer := m.GetTileLayerByName("Tile Layer 1")
	if layer.Name != "Tile Layer 1" {
		t.Error("error get layer")
	}
}

func TestGetObjectLayerByName(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}
	layer := m.GetObjectLayerByName("Object Layer 1")
	if layer.Name != "Object Layer 1" {
		t.Error("error get object layer")
	}
}

func TestGetObjectByName(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}
	objs := m.GetObjectByName("Polygon")
	if len(objs) != 1 {
		t.Error("error invalid objects found")
	}
	for _, obj := range objs {
		if obj.Name != "Polygon" {
			t.Error("error get object by name")
		}
	}
}

func TestBounds(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}
	rect := m.Bounds()
	if rect.H() != 256 {
		t.Error("error height bound invalid")
	}
	if rect.W() != 256 {
		t.Error("error width bound invalid")
	}
}

func TestCentre(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}
	centre := m.Centre()
	if centre.X != 128 {
		t.Error("error centre X invalid")
	}
	if centre.Y != 128 {
		t.Error("error centre Y invalid")
	}
}

func TestMap_DrawAll(t *testing.T) {
	pixelgl.Run(func() {
		m, err := tilepix.ReadFile("testdata/tileobject.tmx")
		if err != nil {
			t.Fatalf("Could not create TilePix map: %v", err)
		}

		target := pixelgl.NewCanvas(pixel.R(0, 0, 320, 320))

		if err := m.DrawAll(target, color.Transparent, pixel.IM); err != nil {
			t.Fatalf("Could not draw map: %v", err)
		}
	})
}
