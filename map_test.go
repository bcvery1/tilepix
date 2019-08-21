package tilepix_test

import (
	"image/color"
	"os"
	"testing"

	_ "image/png"

	"github.com/bcvery1/tilepix"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func TestMain(m *testing.M) {
	pixelgl.Run(func() {
		os.Exit(m.Run())
	})
}

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
	m, err := tilepix.ReadFile("examples/t1.tmx")
	if err != nil {
		t.Fatalf("Could not create TilePix map: %v", err)
	}

	target, err := pixelgl.NewWindow(pixelgl.WindowConfig{Bounds: pixel.R(0, 0, 100, 100)})
	if err != nil {
		t.Fatal(err)
	}

	if err := m.DrawAll(target, color.Transparent, pixel.IM); err != nil {
		t.Fatalf("Could not draw map: %v", err)
	}
}

func BenchmarkMap_DrawAll(b *testing.B) {
	b.SetParallelism(1)

	m, err := tilepix.ReadFile("examples/t1.tmx")
	if err != nil {
		b.Fatalf("Could not create TilePix map: %v", err)
	}

	target, err := pixelgl.NewWindow(pixelgl.WindowConfig{Bounds: pixel.R(0, 0, 100, 100)})
	if err != nil {
		b.Fatal(err)
	}

	// Run as sub benchmark to prevent multiple windows being created
	b.Run("Drawing", func(bb *testing.B) {
		for i := 0; i < bb.N; i++ {
			_ = m.DrawAll(target, color.Transparent, pixel.IM)
		}
	})
}
