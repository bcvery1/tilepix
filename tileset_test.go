package tilepix

import (
	"testing"
)

func TestTileset_String(t *testing.T) {
	type fields struct {
		Name       string
		TileWidth  int
		TileHeight int
		Spacing    int
		Tilecount  int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Basic string",
			fields: fields{
				Name:       "name ts",
				TileWidth:  10,
				TileHeight: 15,
				Spacing:    0,
				Tilecount:  150,
			},
			want: "TileSet{Name: name ts, Tile size: 10x15, Tile spacing: 0, Tilecount: 150, Properties: []}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &Tileset{
				Name:       tt.fields.Name,
				TileWidth:  tt.fields.TileWidth,
				TileHeight: tt.fields.TileHeight,
				Spacing:    tt.fields.Spacing,
				Tilecount:  tt.fields.Tilecount,
			}
			if got := ts.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
