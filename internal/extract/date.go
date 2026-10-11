package extract

import (
	"regexp"
	"strings"
	"time"
)

var (
	dateLayouts = []string{
		"2 January 2006", "2 Jan 2006", "January 2, 2006", "Jan 2, 2006",
		"02-01-2006", "02/01/2006", "2-1-2006", "2/1/2006", "2006-01-02", "02.01.2006",
	}
	indonesianMonths = strings.NewReplacer(
		"januari", "january", "februari", "february", "maret", "march", "mei", "may",
		"juni", "june", "juli", "july", "agustus", "august", "oktober", "october", "desember", "december",
		"agu", "aug", "okt", "oct", "des", "dec",
	)
	reDateLike = regexp.MustCompile(`(?i)\d{1,4}[\s./-]+(?:\d{1,2}|[a-z]+)[\s./,-]+\d{2,4}`)
	reWord     = regexp.MustCompile(`[a-z]+`)
)

func NormalizeDate(s string) (string, bool) {
	candidate := strings.TrimSpace(s)
	if m := reDateLike.FindString(candidate); m != "" {
		candidate = m
	}
	lower := indonesianMonths.Replace(strings.ToLower(candidate))
	title := reWord.ReplaceAllStringFunc(lower, func(w string) string {
		return strings.ToUpper(w[:1]) + w[1:]
	})
	title = strings.Join(strings.Fields(title), " ")
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, title); err == nil {
			return t.Format("2006-01-02"), true
		}
	}
	return strings.TrimSpace(s), false
}
