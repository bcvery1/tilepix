package tilepix_test

import (
	"bytes"
	"encoding/xml"
	"io"
	"io/ioutil"
	"testing"

	"github.com/bcvery1/tilepix"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

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
		{
			name:     "map is infinite",
			filepath: "testdata/infinite.tmx",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "external tileset",
			filepath: "testdata/external_tileset.tmx",
			want:     nil,
			wantErr:  false,
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

func TestRead(t *testing.T) {
	tests := []struct {
		name        string
		input       io.Reader
		dir         string
		want        *tilepix.Map
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "Missing columns parameter in tileset",
			input:       getInput(),
			dir:         "testdata",
			want:        nil,
			wantErr:     true,
			expectedErr: "Tileset columns value not valid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tilepix.Read(tt.input, tt.dir, nil)

			if tt.wantErr && err == nil {
				t.Errorf("tsx.Read: expected error but not nil")
			}

			if tt.wantErr && err != nil && err.Error() != tt.expectedErr {
				t.Errorf("tsx.Read: expected error '%s' but not '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func getInput() io.Reader {
	m := tilepix.Map{
		Version:     "1.2",
		Orientation: "orthogonal",
		Width:       640,
		Height:      480,
		TileWidth:   32,
		TileHeight:  32,
		Tilesets: []*tilepix.Tileset{
			&tilepix.Tileset{
				Source: "tileset_no_columns.tsx",
			},
		},
	}
	byteMap, _ := xml.Marshal(m)
	return bytes.NewReader(byteMap)
}
