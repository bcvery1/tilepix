package tilepix

import "testing"

func TestData_String(t *testing.T) {
	type fields struct {
		Compression string
		DataTiles   []*DataTile
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Basic string",
			fields: fields{Compression: "gzip"},
			want:   "Data{Compression: gzip, DataTiles count: 0}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Data{
				Compression: tt.fields.Compression,
				DataTiles:   tt.fields.DataTiles,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
