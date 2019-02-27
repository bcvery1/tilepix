package tmx_test

import (
	"os"
	"testing"

	"github.com/bcvery1/tilepix/tmx"
)

func TestProperties(t *testing.T) {
	t.Log("Reading", "testdata/poly.tmx")

	r, err := os.Open("testdata/poly.tmx")
	if err != nil {
		t.Fatal(err)
	}

	m, err := tmx.Read(r)
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
