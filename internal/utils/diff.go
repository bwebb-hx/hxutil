package utils

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type DiffParams struct {
	TrimEqual bool
}

func GetDiff(s1, s2 string, diffParams DiffParams) string {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(s1, s2, false)

	// figure out if a diff exists
	diffExists := false
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			if strings.TrimSpace(diff.Text) != "" {
				diffExists = true
				break
			}
		}
	}
	if !diffExists {
		return ""
	}

	filterDiffs := make([]diffmatchpatch.Diff, 0)
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffEqual {
			diff.Text = truncateSection(diff.Text)
		}

		filterDiffs = append(filterDiffs, diff)
	}

	return dmp.DiffPrettyText(filterDiffs)
}

func truncateSection(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= 6 {
		return s
	}
	// cut middle lines out and replace with a "..." line
	var truncated []string
	pre := lines[:3]
	post := lines[len(lines)-4:]
	truncated = append(truncated, pre...)
	truncated = append(truncated, "...")
	truncated = append(truncated, post...)

	return strings.Join(truncated, "\n")
}
