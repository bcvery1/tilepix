package tilepix_test

import (
	"reflect"
	"testing"

	// Required to decode PNG for object tile testing
	_ "image/png"

	"github.com/bcvery1/tilepix"
	"github.com/faiface/pixel"
)

func TestObject_GetEllipse(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/ellipse.tmx")
	if err != nil {
		t.Fatal(err)
	}

	o := m.GetObjectLayerByName("Object Layer 1").Objects[0]

	tests := []struct {
		name    string
		object  *tilepix.Object
		want    pixel.Circle
		wantErr bool
	}{
		{
			name:    "getting ellipse",
			object:  o,
			want:    pixel.C(pixel.V(50, 150), 100),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := o.GetEllipse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Object.GetEllipse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Object.GetEllipse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObject_GetPoint(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/point.tmx")
	if err != nil {
		t.Fatal(err)
	}

	o := m.GetObjectLayerByName("Object Layer 1").Objects[0]

	tests := []struct {
		name    string
		object  *tilepix.Object
		want    pixel.Vec
		wantErr bool
	}{
		{
			name:    "getting point",
			object:  o,
			want:    pixel.V(160, 160),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := o.GetPoint()
			if (err != nil) != tt.wantErr {
				t.Errorf("Object.GetPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Object.GetPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObject_GetRect(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/rectangle.tmx")
	if err != nil {
		t.Fatal(err)
	}

	o := m.GetObjectLayerByName("Object Layer 1").Objects[0]

	tests := []struct {
		name    string
		object  *tilepix.Object
		want    pixel.Rect
		wantErr bool
	}{
		{
			name:    "getting rectangle",
			object:  o,
			want:    pixel.R(0, 0, 100, 100),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := o.GetRect()
			if (err != nil) != tt.wantErr {
				t.Errorf("Object.GetRect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Object.GetRect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObject_GetPolygon(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}

	o := m.GetObjectLayerByName("Object Layer 1").Objects[0]

	tests := []struct {
		name    string
		object  *tilepix.Object
		want    []pixel.Vec
		wantErr bool
	}{
		{
			name:   "getting polygon",
			object: o,
			want: []pixel.Vec{
				pixel.V(0, 256),
				pixel.V(2, 165),
				pixel.V(100, 202),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := o.GetPolygon()
			if (err != nil) != tt.wantErr {
				t.Errorf("Object.GetPolygon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Object.GetPolygon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObject_GetPolyLine(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}

	o := m.GetObjectLayerByName("Object Layer 1").Objects[1]

	tests := []struct {
		name    string
		object  *tilepix.Object
		want    []pixel.Vec
		wantErr bool
	}{
		{
			name:   "getting polyline",
			object: o,
			want: []pixel.Vec{
				pixel.V(0, 256),
				pixel.V(-46, 202),
				pixel.V(-1, 179),
				pixel.V(-43, 142),
				pixel.V(5, 102),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := o.GetPolyLine()
			if (err != nil) != tt.wantErr {
				t.Errorf("Object.GetPolyLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Object.GetPolyLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObjectProperties(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range m.ObjectGroups {
		for _, object := range group.Objects {
			if object.Properties[0].Name != "foo" {
				t.Error("No properties")
			}
			return
		}
	}
	t.Fatal("No property found")
}

func TestObject_GetTile(t *testing.T) {
	m, err := tilepix.ReadFile("testdata/tileobject.tmx")
	if err != nil {
		t.Fatal(err)
	}

	o := m.GetObjectLayerByName("Object Layer 1").Objects[0]

	tests := []struct {
		name    string
		object  *tilepix.Object
		want    *tilepix.DecodedTile
		wantErr bool
	}{
		{
			name:   "getting tile",
			object: o,
			want: &tilepix.DecodedTile{
				ID: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := o.GetTile()
			if (err != nil) != tt.wantErr {
				t.Errorf("Object.GetTile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got.ID != tt.want.ID ||
				got.HorizontalFlip != tt.want.HorizontalFlip ||
				got.VerticalFlip != tt.want.VerticalFlip ||
				got.DiagonalFlip != tt.want.DiagonalFlip ||
				got.Nil != tt.want.Nil {
				t.Errorf("Object.GetTile() = %v, want %v", got, tt.want)
			}
		})
	}
}
