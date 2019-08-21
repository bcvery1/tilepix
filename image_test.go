package tilepix

import (
	"testing"
)

func TestImage_String(t *testing.T) {
	type fields struct {
		Source string
		Width  int
		Height int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Basic string",
			fields: fields{Source: "src", Width: 120, Height: 120},
			want:   "Image{Source: src, Size: 120x120}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Image{
				Source: tt.fields.Source,
				Width:  tt.fields.Width,
				Height: tt.fields.Height,
			}
			if got := i.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
