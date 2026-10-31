// Package dataset loads the original question/key files without changing their text.
package dataset

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Source struct {
	ID        string `json:"id"`
	Questions string `json:"questions"`
	Answers   string `json:"answers"`
}

type Config struct {
	Datasets []Source `json:"datasets"`
	Models   []string `json:"models"`
}

type Asset struct {
	MIME  string `json:"mime"`
	Bytes []byte `json:"bytes"`
}

type Option struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
	Image  string `json:"image,omitempty"`
	Asset  *Asset `json:"asset,omitempty"`
}

type Question struct {
	Number   int      `json:"number"`
	Category string   `json:"category"`
	Section  string   `json:"section,omitempty"` // Optional subdivision of a category, such as an exam section.
	Text     string   `json:"question"`
	Options  []Option `json:"options"`
	Image    string   `json:"image,omitempty"`
	Asset    *Asset   `json:"asset,omitempty"`
	Dataset  string   `json:"dataset"`
	Correct  int      `json:"correct_option"`
}

func (q Question) ID() string { return fmt.Sprintf("%s:%d", q.Dataset, q.Number) }
func (q Question) HasImage() bool {
	if q.Asset != nil {
		return true
	}
	for _, o := range q.Options {
		if o.Asset != nil {
			return true
		}
	}
	return false
}

type Snapshot struct {
	Config    Config                     `json:"config"`
	Questions []Question                 `json:"questions"`
	Metadata  map[string]json.RawMessage `json:"metadata"`
	Warnings  []string                   `json:"warnings,omitempty"`
}

func Hash(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("hash non-JSON value: %v", err))
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func ReadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func Load(path string) (Snapshot, error) {
	s := Snapshot{Metadata: map[string]json.RawMessage{}}
	if err := ReadJSON(path, &s.Config); err != nil {
		return s, err
	}
	if len(s.Config.Datasets) == 0 || len(s.Config.Models) == 0 {
		return s, fmt.Errorf("configuration must contain datasets and models")
	}
	seenModels := map[string]bool{}
	for _, m := range s.Config.Models {
		base := strings.TrimSuffix(m, ":exacto")
		if !strings.Contains(m, "/") || strings.ContainsAny(m, " \t\n") || seenModels[base] {
			return s, fmt.Errorf("invalid or duplicate model %q", m)
		}
		seenModels[base] = true
	}
	seen := map[string]bool{}
	for _, src := range s.Config.Datasets {
		if src.ID == "" || strings.Contains(src.ID, ":") || seen[src.ID] {
			return s, fmt.Errorf("invalid or duplicate dataset ID %q", src.ID)
		}
		seen[src.ID] = true
		qpath, apath := resolve(filepath.Dir(path), src.Questions), resolve(filepath.Dir(path), src.Answers)
		var qf struct {
			Metadata  json.RawMessage `json:"metadata"`
			Questions []Question      `json:"questions"`
		}
		var af struct {
			Answers []struct {
				Number  int `json:"number"`
				Correct int `json:"correct_option"`
			} `json:"answers"`
		}
		if err := ReadJSON(qpath, &qf); err != nil {
			return s, err
		}
		if err := ReadJSON(apath, &af); err != nil {
			return s, err
		}
		if len(qf.Questions) == 0 {
			return s, fmt.Errorf("dataset %s has no questions", src.ID)
		}
		s.Metadata[src.ID] = qf.Metadata
		keys := map[int]int{}
		for _, a := range af.Answers {
			if a.Number < 1 || a.Correct < 1 || a.Correct > 4 || keys[a.Number] != 0 {
				return s, fmt.Errorf("%s: invalid or duplicate answer key %d", apath, a.Number)
			}
			keys[a.Number] = a.Correct
		}
		questions := map[int]bool{}
		for _, q := range qf.Questions {
			q.Dataset, q.Correct = src.ID, keys[q.Number]
			if strings.TrimSpace(q.Category) == "" {
				q.Category = "Uncategorized"
			}
			if q.Number < 1 || questions[q.Number] {
				return s, fmt.Errorf("%s: invalid or duplicate question %d", qpath, q.Number)
			}
			questions[q.Number] = true
			if q.Correct == 0 {
				return s, fmt.Errorf("%s: missing answer key", q.ID())
			}
			if len(q.Options) != 4 {
				return s, fmt.Errorf("%s: expected four options", q.ID())
			}
			if strings.TrimSpace(q.Text) == "" && q.Image == "" {
				return s, fmt.Errorf("%s: question has no content", q.ID())
			}
			var err error
			q.Asset, err = readAsset(filepath.Dir(qpath), q.Image)
			if err != nil {
				return s, fmt.Errorf("%s: %w", q.ID(), err)
			}
			opts := map[int]bool{}
			for i := range q.Options {
				o := &q.Options[i]
				if o.Number < 1 || o.Number > 4 || opts[o.Number] {
					return s, fmt.Errorf("%s: invalid or duplicate option %d", q.ID(), o.Number)
				}
				opts[o.Number] = true
				if strings.TrimSpace(o.Text) == "" && o.Image == "" {
					return s, fmt.Errorf("%s: option %d has no content", q.ID(), o.Number)
				}
				o.Asset, err = readAsset(filepath.Dir(qpath), o.Image)
				if err != nil {
					return s, fmt.Errorf("%s option %d: %w", q.ID(), o.Number, err)
				}
			}
			sort.Slice(q.Options, func(i, j int) bool { return q.Options[i].Number < q.Options[j].Number })
			s.Questions = append(s.Questions, q)
		}
		extra := 0
		for n := range keys {
			if !questions[n] {
				extra++
			}
		}
		if extra > 0 {
			s.Warnings = append(s.Warnings, fmt.Sprintf("%s: %d extra answer keys ignored", src.ID, extra))
		}
	}
	sort.Slice(s.Questions, func(i, j int) bool { return s.Questions[i].ID() < s.Questions[j].ID() })
	return s, nil
}

func resolve(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(base, p)
}

func readAsset(base, path string) (*Asset, error) {
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(resolve(base, path))
	if err != nil {
		return nil, fmt.Errorf("image %s: %w", path, err)
	}
	mime := http.DetectContentType(b)
	switch mime {
	case "image/png", "image/jpeg", "image/gif":
	default:
		return nil, fmt.Errorf("image %s: unsupported content type %s (use PNG, JPEG, or GIF)", path, mime)
	}
	if _, _, err := image.Decode(bytes.NewReader(b)); err != nil {
		return nil, fmt.Errorf("image %s: invalid image: %w", path, err)
	}
	return &Asset{MIME: mime, Bytes: b}, nil
}
