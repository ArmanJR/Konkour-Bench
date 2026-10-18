package supervisor

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareDoesNotStartOrEmbedSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "job.plist")
	args := []string{"run", "--env-file", filepath.Join(dir, "key & config.env"), "--timeout", "1h", "--budget-usd", "50"}
	if err := Prepare(path, "/does-not-exist/bench", dir, args); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := xml.NewDecoder(strings.NewReader(string(raw)))
	var values []string
	var lastKey string
	flags := map[string]bool{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if e, ok := token.(xml.StartElement); ok {
			switch e.Name.Local {
			case "key":
				if err := decoder.DecodeElement(&lastKey, &e); err != nil {
					t.Fatal(err)
				}
			case "string":
				var value string
				if err := decoder.DecodeElement(&value, &e); err != nil {
					t.Fatal(err)
				}
				values = append(values, value)
			case "true", "false":
				flags[lastKey] = e.Name.Local == "true"
			}
		}
	}
	if !flags["RunAtLoad"] || flags["KeepAlive"] {
		t.Fatal("unsafe launch semantics", flags)
	}
	if !strings.Contains(strings.Join(values, "\n"), args[2]) || strings.Contains(string(raw), "OPENROUTER_API_KEY") {
		t.Fatal("env path lost or secret embedded")
	}
	if err := Prepare(path, "/other", dir, nil); err == nil {
		t.Fatal("silently overwrote existing service")
	}
}
