// Package notes provides markdown formatting for highlights
// compatible with Nextcloud Notes API and KOReader's exporter format.
package notes

import (
	"hash/crc32"
	"strings"
	"time"

	"github.com/vanadium23/kompanion/internal/entity"
)

// FormatHighlights converts a slice of highlights to markdown format
// matching KOReader's expected structure.
// Title header: "# {title}\n"
// Author line: "##### {author}\n\n"
// For each highlight grouped by chapter:
//   - Chapter heading if changed: "## {chapter}\n"
//   - Page/time line: "### Page {page} @ {timestamp}\n"
//   - Highlight text: "> {text}\n" (blockquote per D-07)
//   - Note if present: "\n---\n{note}\n"
func FormatHighlights(title, author string, highlights []entity.Highlight) string {
	var sb strings.Builder

	// Title header
	sb.WriteString("# ")
	sb.WriteString(title)
	sb.WriteString("\n")

	// Author line with newline replacement
	cleanAuthor := strings.ReplaceAll(author, "\n", ", ")
	sb.WriteString("##### ")
	sb.WriteString(cleanAuthor)
	sb.WriteString("\n\n")

	// Track current chapter for grouping
	currentChapter := ""

	for _, h := range highlights {
		// Add chapter heading if chapter changed and is non-empty
		if h.Chapter != currentChapter {
			if h.Chapter != "" {
				sb.WriteString("## ")
				sb.WriteString(h.Chapter)
				sb.WriteString("\n")
			}
			currentChapter = h.Chapter
		}

		// Page/time line: "### Page {page} @ {timestamp}"
		sb.WriteString("### Page ")
		sb.WriteString(h.Page)
		sb.WriteString(" @ ")
		sb.WriteString(formatTimestamp(h.Timestamp))
		sb.WriteString("\n")

		// Highlight text as blockquote (per D-07)
		sb.WriteString("> ")
		sb.WriteString(h.Text)
		sb.WriteString("\n")

		// Note if present (plain text per D-07)
		if h.Note != "" {
			sb.WriteString("\n---\n")
			sb.WriteString(h.Note)
			sb.WriteString("\n")
		}

		// Blank line between highlights
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatTitle produces the note title format "{author} - {title}"
// with newlines in author replaced by ", " for multi-author handling.
// This is the title KOReader uses for update detection.
func FormatTitle(author, title string) string {
	cleanAuthor := strings.ReplaceAll(author, "\n", ", ")
	return cleanAuthor + " - " + title
}

// HashToInt converts a string to a stable integer using CRC32 IEEE checksum.
// Required because Nextcloud Notes API expects integer IDs.
func HashToInt(hash string) int {
	return int(crc32.ChecksumIEEE([]byte(hash)))
}

// formatTimestamp converts Unix timestamp to KOReader's expected format:
// "02 January 2006 03:04:05 PM" (matches os.date format in Lua)
func formatTimestamp(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	// KOReader uses: os.date("%d %B %Y %I:%M:%S %p", entry.time)
	// %d = day with leading zero (02)
	// %B = full month name (January)
	// %Y = 4-digit year (2006)
	// %I = hour 12-hour format with leading zero (03)
	// %M = minute with leading zero (04)
	// %S = second with leading zero (05)
	// %p = AM/PM (PM)
	return t.Format("02 January 2006 03:04:05 PM")
}
