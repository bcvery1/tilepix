package tilepix_test

import (
	"os"
	"testing"

	"github.com/bcvery1/tilepix"
)

func TestReadFile(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		want     *tilepix.Map
		wantErr  bool
	}{
		{
			name:     "base64",
			filepath: "testdata/base64.tmx",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "base64-zlib",
			filepath: "testdata/base64-zlib.tmx",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "base64-gzip",
			filepath: "testdata/base64-gzip.tmx",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "csv",
			filepath: "testdata/csv.tmx",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "xml",
			filepath: "testdata/xml.tmx",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "missing file",
			filepath: "testdata/foo.tmx",
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tilepix.ReadFile(tt.filepath)

			if !tt.wantErr && err != nil {
				t.Errorf("tmx.ReadFile(): got unexpected error: %v", err)
			}
			if tt.wantErr && err == nil {
				t.Errorf("tmx.ReadFile(): expected error but not nil")
			}
		})
	}
}

func TestProperties(t *testing.T) {
	t.Log("Reading", "testdata/poly.tmx")

	r, err := os.Open("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}

	m, err := tilepix.Read(r)
	if err != nil {
		t.Fatal(err)
	}

	for _, group := range m.ObjectGroups {
		for _, object := range group.Objects {
			if object.Properties[0].Name != "foo" {
				t.Error("No properties")
			}
			return
		}
	}

	t.Fatal("No property found")
}
