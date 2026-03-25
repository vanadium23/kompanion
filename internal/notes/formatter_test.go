package notes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/notes"
)

func TestFormatHighlights_SingleHighlight(t *testing.T) {
	highlights := []entity.Highlight{
		{
			Text:      "This is a highlighted text",
			Page:      "42",
			Chapter:   "Chapter One",
			Timestamp: 1672531200, // 2023-01-01 00:00:00 UTC
		},
	}

	output := notes.FormatHighlights("Test Book", "Test Author", highlights)

	// Verify title
	assert.Contains(t, output, "# Test Book")
	// Verify author with comma separator format
	assert.Contains(t, output, "##### Test Author")
	// Verify chapter heading
	assert.Contains(t, output, "## Chapter One")
	// Verify page/time format
	assert.Contains(t, output, "### Page 42 @")
	// Verify highlight text is blockquote format (per D-07)
	assert.Contains(t, output, "> This is a highlighted text")
}

func TestFormatHighlights_MultipleChapters(t *testing.T) {
	highlights := []entity.Highlight{
		{
			Text:      "First highlight",
			Page:      "10",
			Chapter:   "Chapter 1",
			Timestamp: 1672531200,
		},
		{
			Text:      "Second highlight in chapter 1",
			Page:      "15",
			Chapter:   "Chapter 1",
			Timestamp: 1672531300,
		},
		{
			Text:      "Highlight in chapter 2",
			Page:      "50",
			Chapter:   "Chapter 2",
			Timestamp: 1672531400,
		},
	}

	output := notes.FormatHighlights("Test Book", "Author", highlights)

	// Verify chapter 1 heading appears before its highlights
	assert.Contains(t, output, "## Chapter 1")
	assert.Contains(t, output, "> First highlight")
	assert.Contains(t, output, "> Second highlight in chapter 1")

	// Verify chapter 2 heading appears
	assert.Contains(t, output, "## Chapter 2")
	assert.Contains(t, output, "> Highlight in chapter 2")
}

func TestFormatHighlights_WithNote(t *testing.T) {
	highlights := []entity.Highlight{
		{
			Text:      "Highlighted text with note",
			Page:      "25",
			Chapter:   "Introduction",
			Timestamp: 1672531200,
			Note:      "This is my personal note",
		},
	}

	output := notes.FormatHighlights("Book", "Author", highlights)

	// Verify separator before note
	assert.Contains(t, output, "\n---\n")
	// Verify note text appears after separator (plain text, not blockquote per D-07)
	assert.Contains(t, output, "This is my personal note")
}

func TestFormatHighlights_EmptyChapter(t *testing.T) {
	highlights := []entity.Highlight{
		{
			Text:      "Highlight without chapter",
			Page:      "1",
			Chapter:   "",
			Timestamp: 1672531200,
		},
	}

	output := notes.FormatHighlights("Book", "Author", highlights)

	// Should not have chapter heading when chapter is empty
	// Check that "## \n" pattern (empty chapter heading) is not present
	assert.NotContains(t, output, "## \n")
	assert.Contains(t, output, "> Highlight without chapter")
}

func TestFormatTitle(t *testing.T) {
	output := notes.FormatTitle("John", "My Book")
	assert.Equal(t, "John - My Book", output)
}

func TestFormatTitle_MultiAuthor(t *testing.T) {
	// Authors separated by newlines should be comma-separated
	output := notes.FormatTitle("Author One\nAuthor Two", "Collaborative Book")
	assert.Equal(t, "Author One, Author Two - Collaborative Book", output)
}

func TestHashToInt_Stable(t *testing.T) {
	// Same input should produce same output
	result1 := notes.HashToInt("abc123")
	result2 := notes.HashToInt("abc123")
	assert.Equal(t, result1, result2)
}

func TestHashToInt_Different(t *testing.T) {
	// Different inputs should produce different outputs
	result1 := notes.HashToInt("abc")
	result2 := notes.HashToInt("def")
	assert.NotEqual(t, result1, result2)
}
