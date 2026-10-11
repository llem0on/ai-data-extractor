package extract

import (
	"strings"
)

var lineMarkers = []string{"[TABLE] ", "[HEADER] ", "[FIELD] ", "[HEADING] ", "[TEXT] "}

func ruleExtract(fields []*Field, text string) map[string]string {
	found := map[string]string{}
	var pending []*Field
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		for _, m := range lineMarkers {
			line = strings.TrimPrefix(line, m)
		}
		if line == "" {
			continue
		}
		cells := splitCells(line)
		for _, f := range pending {
			if found[f.Name] == "" && !isLabel(fields, cells[0]) {
				found[f.Name] = cleanValue(cells[0])
			}
		}
		pending = pending[:0]
		last := cells[len(cells)-1]
		for _, f := range fields {
			if f.labelColonRe != nil && found[f.Name] == "" && f.labelColonRe.MatchString(last) {
				pending = append(pending, f)
			}
		}
		for _, f := range fields {
			if f.valueRe == nil || found[f.Name] != "" {
				continue
			}
			if v := matchCells(f, cells); v != "" {
				found[f.Name] = v
			}
		}
	}
	return found
}

func isLabel(fields []*Field, cell string) bool {
	for _, f := range fields {
		if f.labelColonRe != nil && f.labelColonRe.MatchString(cell) {
			return true
		}
	}
	return false
}

func splitCells(line string) []string {
	var cells []string
	for _, part := range strings.FieldsFunc(line, func(r rune) bool { return r == '|' || r == '\t' }) {
		if p := strings.TrimSpace(part); p != "" {
			cells = append(cells, p)
		}
	}
	return cells
}

func matchCells(f *Field, cells []string) string {
	for i, c := range cells {
		if m := f.valueRe.FindStringSubmatch(c); m != nil {
			return cleanValue(m[1])
		}
		if f.labelRe.MatchString(c) && i+1 < len(cells) {
			return cleanValue(strings.TrimLeft(cells[i+1], ":- "))
		}
	}
	return ""
}

func cleanValue(s string) string {
	return strings.Trim(strings.Join(strings.Fields(s), " "), " :-")
}
