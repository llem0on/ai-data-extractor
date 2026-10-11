package document

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

const docxBody = `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>
<w:tbl><w:tr><w:tc><w:p><w:r><w:t>No. Dokumen</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>DOC-</w:t></w:r><w:r><w:t>01</w:t></w:r></w:p></w:tc></w:tr></w:tbl>
<w:p><w:r><w:t>1. Tujuan</w:t></w:r></w:p>
</w:body></w:document>`

const docxHeader = `<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006">
<w:p><w:r><w:t>Judul: Kebijakan Akses</w:t></w:r></w:p>
<mc:AlternateContent><mc:Fallback><w:p><w:r><w:t>DUPLICATE</w:t></w:r></w:p></mc:Fallback></mc:AlternateContent>
</w:hdr>`

func buildDOCX(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range map[string]string{"word/document.xml": docxBody, "word/header1.xml": docxHeader} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(content))
	}
	zw.Close()
	return buf.Bytes()
}

func TestReadDOCXKeepsHeaderTablesAndRuns(t *testing.T) {
	text, err := readDOCX(buildDOCX(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[HEADER] Judul: Kebijakan Akses", "[TABLE] No. Dokumen | DOC-01", "1. Tujuan"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "DUPLICATE") {
		t.Errorf("fallback content must be skipped:\n%s", text)
	}
}

func TestMarkLines(t *testing.T) {
	got := markLines("STANDARD OPERATING PROCEDURE\nNo. Dokumen : DOC-01\nThis procedure applies to all servers.\n")
	want := "[HEADING] STANDARD OPERATING PROCEDURE\n[FIELD] No. Dokumen : DOC-01\n[TEXT] This procedure applies to all servers.\n"
	if got != want {
		t.Fatalf("got:\n%s", got)
	}
}
