package document

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	wordNS   = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	markupNS = "http://schemas.openxmlformats.org/markup-compatibility/2006"
)

func readDOCX(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("not a valid DOCX file: %w", err)
	}

	var headers []string
	var body string
	files := append([]*zip.File(nil), zr.File...)
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	for _, f := range files {
		isHeader := strings.HasPrefix(f.Name, "word/header") && strings.HasSuffix(f.Name, ".xml")
		if !isHeader && f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		raw, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		text := wordText(raw)
		if isHeader {
			if text != "" {
				headers = append(headers, prefixLines("[HEADER] ", text))
			}
			continue
		}
		body = text
	}
	if body == "" && len(headers) == 0 {
		return "", fmt.Errorf("DOCX contains no text")
	}
	return strings.TrimSpace(strings.Join(headers, "\n") + "\n" + body), nil
}

func wordText(data []byte) string {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var out, para strings.Builder
	var row []string
	inText, inCell := false, false
	fallback := 0

	flushPara := func() {
		t := strings.TrimSpace(para.String())
		para.Reset()
		if t == "" {
			return
		}
		if inCell {
			row = append(row, t)
			return
		}
		out.WriteString(t)
		out.WriteString("\n")
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Space == markupNS && t.Name.Local == "Fallback" {
				fallback++
			}
			if fallback > 0 || t.Name.Space != wordNS {
				continue
			}
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab":
				para.WriteString(" ")
			case "tc":
				inCell = true
			case "tr":
				row = row[:0]
			}
		case xml.EndElement:
			if t.Name.Space == markupNS && t.Name.Local == "Fallback" {
				fallback--
				continue
			}
			if fallback > 0 || t.Name.Space != wordNS {
				continue
			}
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				flushPara()
			case "tc":
				inCell = false
			case "tr":
				if len(row) > 0 {
					out.WriteString("[TABLE] " + strings.Join(row, " | ") + "\n")
				}
			}
		case xml.CharData:
			if inText && fallback == 0 {
				para.Write(t)
			}
		}
	}
	return strings.TrimSpace(out.String())
}

func prefixLines(prefix, text string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if !strings.HasPrefix(l, "[TABLE] ") {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "\n")
}
