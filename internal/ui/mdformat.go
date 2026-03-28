package ui

import (
	"regexp"
	"strings"
)

var calloutMeta = map[string]struct{ icon, label string }{
	"note":     {"ℹ️", "Note"},
	"advanced": {"ℹ️", "Advanced"},
	"caution":  {"⚠️", "Caution"},
}

var exercismBlockRe = regexp.MustCompile(`(?m)^~~~+exercism/(\w+)\s*$`)
var htmlCommentRe = regexp.MustCompile(`<!--[\s\S]*?-->`)

// preprocessMarkdown converts exercism-specific markdown into standard markdown
// that Glamour can render well.
func preprocessMarkdown(md string) string {
	md = convertCalloutBlocks(md)
	md = htmlCommentRe.ReplaceAllString(md, "")
	return md
}

// convertCalloutBlocks turns ~~~exercism/TYPE ... ~~~ blocks into blockquotes.
// Joins continuation lines into paragraphs so Glamour can re-wrap them at
// the correct width (raw exercism markdown is pre-wrapped at ~80 cols).
func convertCalloutBlocks(md string) string {
	lines := strings.Split(md, "\n")
	var out []string
	var inCallout bool
	var fenceLen int
	var paraLines []string

	flushPara := func() {
		if len(paraLines) > 0 {
			out = append(out, "> "+strings.Join(paraLines, " "))
			paraLines = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inCallout {
			if m := exercismBlockRe.FindStringSubmatch(trimmed); m != nil {
				inCallout = true
				fenceLen = strings.Count(strings.Split(trimmed, "exercism")[0], "~")
				meta, ok := calloutMeta[m[1]]
				if !ok {
					meta = calloutMeta["note"]
				}
				out = append(out, "")
				out = append(out, "> **"+meta.icon+" "+meta.label+"**")
				out = append(out, ">")
				continue
			}
			out = append(out, line)
			continue
		}

		// Inside callout
		closeFence := strings.Repeat("~", fenceLen)
		if strings.HasPrefix(trimmed, closeFence) && !strings.Contains(trimmed, "exercism") {
			flushPara()
			inCallout = false
			out = append(out, "")
			continue
		}

		if trimmed == "" {
			flushPara()
			out = append(out, ">")
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "```") {
			// List items and code fences stay on their own line
			flushPara()
			out = append(out, "> "+trimmed)
		} else if len(paraLines) > 0 && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
			// Indented continuation of a list item — keep joined
			paraLines = append(paraLines, trimmed)
		} else {
			paraLines = append(paraLines, trimmed)
		}
	}
	flushPara()

	return strings.Join(out, "\n")
}
