package tilepix

import "testing"

func TestObject_String(t *testing.T) {
	type fields struct {
		Name       string
		objectType ObjectType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Ellipse",
			fields: fields{
				Name:       "object 1",
				objectType: EllipseObj,
			},
			want: "Object{Ellipse, Name: 'object 1'}",
		},
		{
			name: "Polygon",
			fields: fields{
				Name:       "object 2",
				objectType: PolygonObj,
			},
			want: "Object{Polygon, Name: 'object 2'}",
		},
		{
			name: "Polyline",
			fields: fields{
				Name:       "object 3",
				objectType: PolylineObj,
			},
			want: "Object{Polyline, Name: 'object 3'}",
		},
		{
			name: "Rectangle",
			fields: fields{
				Name:       "object 4",
				objectType: RectangleObj,
			},
			want: "Object{Rectangle, Name: 'object 4'}",
		},
		{
			name: "Point",
			fields: fields{
				Name:       "object 5",
				objectType: PointObj,
			},
			want: "Object{Point, Name: 'object 5'}",
		},
		{
			name: "Tile",
			fields: fields{
				Name:       "object 6",
				objectType: TileObj,
			},
			want: "Object{Tile, Name: 'object 6'}",
		},
		{
			name: "Unknown",
			fields: fields{
				Name:       "object 7",
				objectType: 6,
			},
			want: "Object{Unknown, Name: 'object 7'}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Object{
				Name:       tt.fields.Name,
				objectType: tt.fields.objectType,
			}
			if got := o.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
