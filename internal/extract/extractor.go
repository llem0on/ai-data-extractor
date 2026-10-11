package extract

import (
	"context"
	"strings"
	"time"

	"github.com/llem0on/ai-data-extractor/internal/document"
)

const (
	SourceRule = "rule"
	SourceLLM  = "llm"
)

type Value struct {
	Value    string `json:"value"`
	Source   string `json:"source"`
	Verified bool   `json:"verified"`
}

type Result struct {
	Document document.Document `json:"document"`
	Fields   map[string]*Value `json:"fields"`
	Missing  []string          `json:"missing"`
	LLMUsed  bool              `json:"llm_used"`
	LLMError string            `json:"llm_error,omitempty"`
	TookMS   int64             `json:"took_ms"`
	Text     string            `json:"text,omitempty"`
}

type Extractor struct {
	Schema *Schema
	LLM    *LLM
}

func (e *Extractor) Run(ctx context.Context, doc document.Document) Result {
	started := time.Now()
	res := Result{Document: doc, Fields: map[string]*Value{}}

	for name, v := range ruleExtract(e.Schema.Fields, doc.Text) {
		res.Fields[name] = &Value{Value: v, Source: SourceRule, Verified: true}
	}

	var missing []*Field
	for _, f := range e.Schema.Fields {
		if res.Fields[f.Name] == nil {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 && e.LLM.Enabled() {
		res.LLMUsed = true
		got, err := e.LLM.Extract(ctx, missing, doc.Text)
		if err != nil {
			res.LLMError = err.Error()
		}
		for _, f := range missing {
			if v := got[f.Name]; v != "" {
				res.Fields[f.Name] = &Value{Value: v, Source: SourceLLM, Verified: grounded(v, doc.Text)}
			}
		}
	}

	for _, f := range e.Schema.Fields {
		v := res.Fields[f.Name]
		if v == nil {
			res.Missing = append(res.Missing, f.Name)
			continue
		}
		switch f.Type {
		case "date":
			v.Value, _ = NormalizeDate(v.Value)
		case "enum":
			v.Value = matchEnum(v.Value, f.Values)
		}
	}
	res.TookMS = time.Since(started).Milliseconds()
	return res
}

func grounded(value, text string) bool {
	norm := func(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
	return strings.Contains(norm(text), norm(value))
}

func matchEnum(v string, options []string) string {
	lower := strings.ToLower(v)
	for _, o := range options {
		if strings.EqualFold(o, v) || strings.Contains(lower, strings.ToLower(o)) {
			return o
		}
	}
	return v
}
