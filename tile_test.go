package tilepix

import (
	"testing"
)

func TestDecodedTile_String(t1 *testing.T) {
	type fields struct {
		ID  ID
		Nil bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Basic string - not nil tile",
			fields: fields{
				ID:  1,
				Nil: false,
			},
			want: "DecodedTile{ID: 1, Is nil: false}",
		},
		{
			name: "Basic string - nil tile",
			fields: fields{
				ID:  2,
				Nil: true,
			},
			want: "DecodedTile{ID: 2, Is nil: true}",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &DecodedTile{
				ID:  tt.fields.ID,
				Nil: tt.fields.Nil,
			}
			if got := t.String(); got != tt.want {
				t1.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
