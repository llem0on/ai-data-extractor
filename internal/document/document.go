package document

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Text   string `json:"-"`
}

type Options struct {
	PDFPages int
	MaxChars int
}

func ReadFile(ctx context.Context, path string, opts Options) (Document, error) {
	name := filepath.Base(path)
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".pdf":
		text, err := readPDF(ctx, path, opts.PDFPages)
		if err != nil {
			return Document{}, err
		}
		return Document{Name: name, Format: "pdf", Text: limit(text, opts.MaxChars)}, nil
	case ".docx":
		data, err := os.ReadFile(path)
		if err != nil {
			return Document{}, err
		}
		text, err := readDOCX(data)
		if err != nil {
			return Document{}, err
		}
		return Document{Name: name, Format: "docx", Text: limit(text, opts.MaxChars)}, nil
	default:
		return Document{}, fmt.Errorf("unsupported file type %q (use .pdf or .docx)", ext)
	}
}

func limit(s string, n int) string {
	r := []rune(s)
	if n <= 0 || len(r) <= n {
		return s
	}
	return string(r[:n])
}
