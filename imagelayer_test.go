package tilepix

import "testing"

func TestImageLayer_String(t *testing.T) {
	type fields struct {
		Name  string
		Image *Image
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Basic string",
			fields: fields{
				Name: "name",
				Image: &Image{
					Source: "src",
					Width:  120,
					Height: 120,
				},
			},
			want: "ImageLayer{Name: 'name', Image: Image{Source: src, Size: 120x120}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := &ImageLayer{
				Name:  tt.fields.Name,
				Image: tt.fields.Image,
			}
			if got := im.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
