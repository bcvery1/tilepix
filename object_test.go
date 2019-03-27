package tilepix_test

import (
	"reflect"
	"testing"

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
