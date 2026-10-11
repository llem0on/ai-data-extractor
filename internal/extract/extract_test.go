package extract

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/llem0on/ai-data-extractor/internal/document"
)

const sample = `[HEADER] PT Contoh Teknologi
[TABLE] No. Dokumen | DOC-OPS-07.002
[TABLE] Revisi | 03
[FIELD] Judul : Prosedur Backup dan Restore Data
[FIELD] Departemen : IT Operations
[FIELD] Tanggal Berlaku : 12 Agustus 2025
[TEXT] Nomor Revisi: lihat tabel di atas.
[TEXT] Disetujui oleh: Head of IT`

func mustSchema(t *testing.T) *Schema {
	t.Helper()
	s, err := LoadSchema("")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRulesReadLabelsInlineAndInTableCells(t *testing.T) {
	got := ruleExtract(mustSchema(t).Fields, sample)
	want := map[string]string{
		"document_number": "DOC-OPS-07.002",
		"revision":        "03",
		"title":           "Prosedur Backup dan Restore Data",
		"department":      "IT Operations",
		"effective_date":  "12 Agustus 2025",
		"approved_by":     "Head of IT",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: want %q, got %q", k, v, got[k])
		}
	}
	if _, ok := got["prepared_by"]; ok {
		t.Errorf("prepared_by should be missing, got %q", got["prepared_by"])
	}
}

func TestRulesReadValueOnNextLine(t *testing.T) {
	text := "[HEADING] NO. DOKUMEN:\n[HEADING] DOC 06.001.01\n[TEXT] Mulai berlaku tanggal:\n[TEXT] 01 Agustus 2026\n[TEXT] Revisi:\n[TEXT] 3.0\n[TEXT] Tanggal\n[TEXT] Revisi\n[TEXT] Perubahan"
	got := ruleExtract(mustSchema(t).Fields, text)
	if got["document_number"] != "DOC 06.001.01" || got["effective_date"] != "01 Agustus 2026" || got["revision"] != "3.0" {
		t.Fatalf("next-line values not read: %v", got)
	}
}

func TestNormalizeDate(t *testing.T) {
	cases := map[string]string{
		"12 Agustus 2025":                "2025-08-12",
		"1 Desember 2024":                "2024-12-01",
		"03/02/2025":                     "2025-02-03",
		"2025-08-12":                     "2025-08-12",
		"January 5, 2025":                "2025-01-05",
		"berlaku mulai 7 Mei 2024 pukul": "2024-05-07",
	}
	for in, want := range cases {
		if got, ok := NormalizeDate(in); !ok || got != want {
			t.Errorf("%q: want %s, got %s (%v)", in, want, got, ok)
		}
	}
}

func TestLLMFillsOnlyMissingFieldsAndIsGrounded(t *testing.T) {
	var asked []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct{ Content string } `json:"messages"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		asked = append(asked, req.Messages[1].Content)
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{
			"content": "```json\n{\"document_type\": \"Standard Operating Procedure (SOP)\", \"prepared_by\": \"Jane Doe\", \"title\": \"ignored\"}\n```",
		}}}})
	}))
	defer srv.Close()

	ex := &Extractor{Schema: mustSchema(t), LLM: &LLM{BaseURL: srv.URL, Model: "test", HTTP: srv.Client()}}
	res := ex.Run(context.Background(), document.Document{Name: "a.pdf", Text: sample})

	if !res.LLMUsed || len(asked) != 1 {
		t.Fatalf("want one LLM call, got %d", len(asked))
	}
	if res.Fields["title"].Source != SourceRule {
		t.Fatalf("rule value must not be overwritten: %+v", res.Fields["title"])
	}
	if v := res.Fields["document_type"]; v == nil || v.Value != "SOP" || v.Source != SourceLLM {
		t.Fatalf("enum not normalized: %+v", v)
	}
	if v := res.Fields["prepared_by"]; v == nil || v.Verified {
		t.Fatalf("value absent from the text must be unverified: %+v", v)
	}
	if res.Fields["effective_date"].Value != "2025-08-12" {
		t.Fatalf("date not normalized: %+v", res.Fields["effective_date"])
	}
}

func TestWithoutLLMMissingFieldsAreReported(t *testing.T) {
	ex := &Extractor{Schema: mustSchema(t), LLM: &LLM{}}
	res := ex.Run(context.Background(), document.Document{Text: sample})
	if res.LLMUsed || len(res.Missing) != 2 {
		t.Fatalf("want 2 missing fields without LLM, got %v", res.Missing)
	}
}
