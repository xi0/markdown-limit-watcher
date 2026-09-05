package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	clearScreen = "\033[H\033[2J"
)

var (
	headingRe = regexp.MustCompile(`^#{1,6}\s+(.+)$`)
	limitRe   = regexp.MustCompile(`^-\s+Limit:\s+(\d+)\s+characters\s*$`)
)

type section struct {
	title   string
	limit   int
	content string
}

// parseSections finds all sections that contain a limit declaration anywhere
// within the section body (not necessarily immediately after the heading).
// Only the content from the limit declaration to the next heading is counted.
func parseSections(data string) []section {
	lines := strings.Split(data, "\n")
	var sections []section

	// First, identify where each heading starts and ends.
	type headingSpan struct {
		title   string
		start   int // index of the heading line
		bodyEnd int // index of the line where the next heading starts (exclusive)
	}

	var spans []headingSpan
	for i, line := range lines {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			if len(spans) > 0 {
				spans[len(spans)-1].bodyEnd = i
			}
			spans = append(spans, headingSpan{
				title:   strings.TrimSpace(m[1]),
				start:   i,
				bodyEnd: len(lines),
			})
		}
	}

	for _, span := range spans {
		// Search for a limit declaration anywhere in this section's body.
		limitLineIdx := -1
		var limit int
		for j := span.start + 1; j < span.bodyEnd; j++ {
			if lm := limitRe.FindStringSubmatch(lines[j]); lm != nil {
				fmt.Sscanf(lm[1], "%d", &limit)
				limitLineIdx = j
				break
			}
		}

		if limitLineIdx == -1 {
			continue // no limit declaration in this section
		}

		// Collect content lines after the limit declaration until the next heading.
		var contentLines []string
		for j := limitLineIdx + 1; j < span.bodyEnd; j++ {
			contentLines = append(contentLines, lines[j])
		}

		content := strings.TrimSpace(strings.Join(contentLines, "\n"))

		sections = append(sections, section{
			title:   span.title,
			limit:   limit,
			content: content,
		})
	}

	return sections
}

// renderBar renders a coloured progress bar of width `width`.
func renderBar(count, limit, width int) string {
	pct := 0.0
	if limit > 0 {
		pct = float64(count) / float64(limit)
	}

	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}

	var color string
	switch {
	case pct > 1.0:
		color = colorRed
	case pct >= 0.9:
		color = colorYellow
	default:
		color = colorGreen
	}

	bar := color + strings.Repeat("█", filled) + colorReset + strings.Repeat("░", width-filled)
	return "[" + bar + "]"
}

func display(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		return
	}

	sections := parseSections(string(data))

	fmt.Print(clearScreen)
	fmt.Printf("📄  Watching: %s\n\n", filename)

	if len(sections) == 0 {
		fmt.Println("No sections with limits found.")
		return
	}

	barWidth := 40

	for _, s := range sections {
		count := len([]rune(s.content))
		pct := 0.0
		if s.limit > 0 {
			pct = float64(count) / float64(s.limit)
		}

		var color string
		switch {
		case pct > 1.0:
			color = colorRed
		case pct >= 0.9:
			color = colorYellow
		default:
			color = colorGreen
		}

		bar := renderBar(count, s.limit, barWidth)

		fmt.Printf("  %s%s%s\n", color, s.title, colorReset)
		fmt.Printf("  %s%d / %d characters (%.1f%%)%s\n", color, count, s.limit, pct*100, colorReset)
		fmt.Printf("  %s\n\n", bar)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: limit_watcher <markdown-file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Get initial mtime
	info, err := os.Stat(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error stating file: %v\n", err)
		os.Exit(1)
	}
	lastMod := info.ModTime()

	// Initial render
	display(filename)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		info, err := os.Stat(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error stating file: %v\n", err)
			continue
		}

		if info.ModTime().After(lastMod) {
			lastMod = info.ModTime()
			// Wait one additional second
			time.Sleep(1 * time.Second)
			display(filename)
		}
	}
}
