package extract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const systemPrompt = `You extract structured fields from business documents written in Indonesian or English.
Return only one JSON object. Use exactly the keys you are given. Use null when the document does not state a value.
Copy values as they appear in the document; do not translate, summarize or guess.
Lines may start with markers: [HEADER] page header, [TABLE] table row with cells separated by |, [FIELD] label: value, [HEADING] heading, [TEXT] body text.`

type LLM struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

func (l *LLM) Enabled() bool {
	return l != nil && l.BaseURL != "" && l.Model != ""
}

func (l *LLM) Extract(ctx context.Context, fields []*Field, text string) (map[string]string, error) {
	var spec strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&spec, "- %s: %s", f.Name, f.Description)
		switch f.Type {
		case "date":
			spec.WriteString(" (date, keep the original wording)")
		case "enum":
			fmt.Fprintf(&spec, " (one of: %s)", strings.Join(f.Values, ", "))
		}
		spec.WriteString("\n")
	}
	user := "Fields:\n" + spec.String() + "\nDocument:\n" + text

	body := map[string]any{
		"model":           l.Model,
		"temperature":     0,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": user},
		},
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := l.post(ctx, body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("llm: empty response")
	}
	return parseJSONObject(resp.Choices[0].Message.Content)
}

func parseJSONObject(content string) (map[string]string, error) {
	start, end := strings.Index(content, "{"), strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("llm: no JSON object in response")
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(content[start:end+1]), &raw); err != nil {
		return nil, fmt.Errorf("llm: invalid JSON: %w", err)
	}
	out := map[string]string{}
	for k, v := range raw {
		switch t := v.(type) {
		case string:
			if s := strings.TrimSpace(t); s != "" && !strings.EqualFold(s, "null") {
				out[k] = s
			}
		case float64:
			out[k] = strings.TrimSuffix(fmt.Sprintf("%v", t), ".0")
		}
	}
	return out, nil
}

func (l *LLM) post(ctx context.Context, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(l.BaseURL, "/")+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if l.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+l.APIKey)
	}
	resp, err := l.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("llm: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("llm: status %d: %.200s", resp.StatusCode, raw)
	}
	return json.Unmarshal(raw, out)
}
