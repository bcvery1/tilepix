package tilepix

import "testing"

func TestObjectGroup_String(t *testing.T) {
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
			fields: fields{Name: "name_og"},
			want:   "ObjectGroup{Name: name_og, Properties: [], Objects: []}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			og := &ObjectGroup{
				Name: tt.fields.Name,
			}
			if got := og.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
