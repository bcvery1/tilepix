package tilepix

import "testing"

func TestPoint_String(t *testing.T) {
	type fields struct {
		X int
		Y int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Basic string",
			fields: fields{
				X: 1,
				Y: 2,
			},
			want: "Point{1, 2}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Point{
				X: tt.fields.X,
				Y: tt.fields.Y,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
