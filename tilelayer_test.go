package tilepix

import (
	"testing"
)

func TestTileLayer_String(t *testing.T) {
	type fields struct {
		Name string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Basic string",
			fields: fields{Name: "name tl"},
			want:   "TileLayer{Name: 'name tl', Properties: [], TileCount: 0}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &TileLayer{
				Name: tt.fields.Name,
			}
			if got := l.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
