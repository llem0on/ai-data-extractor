package document

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	reFieldLine = regexp.MustCompile(`^(.{3,40}?)\s*:\s*(.+)$`)
	reUpper     = regexp.MustCompile(`[A-Z]`)
)

func readPDF(ctx context.Context, path string, pages int) (string, error) {
	args := []string{}
	if pages > 0 {
		args = append(args, "-f", "1", "-l", strconv.Itoa(pages))
	}
	args = append(args, path, "-")
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "pdftotext", args...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext: %w", err)
	}
	return markLines(out.String()), nil
}

func markLines(text string) string {
	var b strings.Builder
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case t == "":
			continue
		case reFieldLine.MatchString(t):
			b.WriteString("[FIELD] ")
		case len(t) <= 60 && t == strings.ToUpper(t) && reUpper.MatchString(t):
			b.WriteString("[HEADING] ")
		default:
			b.WriteString("[TEXT] ")
		}
		b.WriteString(t)
		b.WriteString("\n")
	}
	return b.String()
}
