package extract

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

//go:embed default_schema.json
var defaultSchema []byte

type Field struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Labels      []string `json:"labels"`
	Values      []string `json:"values,omitempty"`

	valueRe      *regexp.Regexp
	labelRe      *regexp.Regexp
	labelColonRe *regexp.Regexp
}

type Schema struct {
	Fields []*Field `json:"fields"`
}

func LoadSchema(path string) (*Schema, error) {
	raw := defaultSchema
	if path != "" {
		var err error
		if raw, err = os.ReadFile(path); err != nil {
			return nil, err
		}
	}
	var s Schema
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	if len(s.Fields) == 0 {
		return nil, fmt.Errorf("schema has no fields")
	}
	seen := map[string]bool{}
	for _, f := range s.Fields {
		if f.Name == "" || seen[f.Name] {
			return nil, fmt.Errorf("field names must be unique and non-empty")
		}
		seen[f.Name] = true
		if f.Type == "" {
			f.Type = "string"
		}
		if f.Type != "string" && f.Type != "date" && f.Type != "enum" {
			return nil, fmt.Errorf("field %s: type must be string, date or enum", f.Name)
		}
		if len(f.Labels) > 0 {
			labels := strings.Join(f.Labels, "|")
			var err error
			if f.valueRe, err = regexp.Compile(`(?i)^\s*(?:` + labels + `)\s*[:\-]\s*(\S.*)$`); err != nil {
				return nil, fmt.Errorf("field %s: %w", f.Name, err)
			}
			f.labelRe = regexp.MustCompile(`(?i)^\s*(?:` + labels + `)\s*[:\-]?\s*$`)
			f.labelColonRe = regexp.MustCompile(`(?i)^\s*(?:` + labels + `)\s*[:\-]\s*$`)
		}
	}
	return &s, nil
}
