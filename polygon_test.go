package tilepix

import "testing"

func TestPolygon_String(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Basic string",
			want: "Polygon{Points: []}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Polygon{}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
