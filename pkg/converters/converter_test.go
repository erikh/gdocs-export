package converters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/erikh/gdocs-export/pkg/downloader"
	"google.golang.org/api/docs/v1"
)

const testdataDir = "testdata"

func TestFixtures(t *testing.T) {
	fis, err := ioutil.ReadDir(testdataDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		name := fi.Name()

		dir := filepath.Join(testdataDir, name)
		doc := &docs.Document{}
		manifest := downloader.Manifest{}

		f, err := os.Open(filepath.Join(dir, name+".json"))
		if err != nil {
			t.Fatalf("While testing %q: %v", name, err)
		}

		if err := json.NewDecoder(f).Decode(doc); err != nil {
			t.Fatalf("%q: could not decode document: %v", name, err)
		}
		f.Close()

		if fi, err := os.Stat(filepath.Join(dir, "assets")); err == nil && fi.IsDir() {
			f, err := os.Open(filepath.Join(dir, "assets", "manifest.json"))
			if err != nil {
				t.Fatalf("%q could not find manifest.json, but found assets dir", name)
			}

			if err := json.NewDecoder(f).Decode(&manifest); err != nil {
				t.Fatalf("%q: could not decode manifest: %v", name, err)
			}

			f.Close()
		}

		for _, typ := range []string{"md", "html"} {
			out, err := Convert(typ, doc, manifest)
			if err != nil {
				t.Fatalf("while converting %q to %q: %v", name, typ, err)
			}

			fixture, err := ioutil.ReadFile(filepath.Join(dir, name+"."+typ))
			if err != nil {
				t.Fatalf("while reading fixture for %q, type %q: %v", name, typ, err)
			}

			fixture = bytes.TrimSpace(fixture)
			out = strings.TrimSpace(out)

			if string(fixture) != out {
				fmt.Println(diff.LineDiff(string(fixture), out))
				t.Fatalf("%q: content does not match for type %q", name, typ)
			}
		}
	}
}
