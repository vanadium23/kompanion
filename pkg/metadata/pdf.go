package metadata

import (
	"bufio"
	"os"
	"strings"
)

// PDFMetadata holds the extracted PDFmetadata information
type PDFMetadata struct {
	Title    string
	Author   string
	Subject  string
	Keywords string
}

// extractPDFMetadataFromHeader scans the first part of the PDF file for PDFmetadata information
func extractPdfMetadata(tmpFile *os.File) (Metadata, error) {
	scanner := bufio.NewScanner(tmpFile)
	var PDFmetadata Metadata
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "/Title") {
			PDFmetadata.Title = extractValue(line, "/Title")
		}
		if strings.Contains(line, "/Author") {
			PDFmetadata.Author = extractValue(line, "/Author")
		}
		// if strings.Contains(line, "/Subject") {
		// 	PDFmetadata.Subject = extractValue(line, "/Subject")
		// }
		// if strings.Contains(line, "/Keywords") {
		// 	PDFmetadata.Keywords = extractValue(line, "/Keywords")
		// }

		// Break early if we've found all fields
		if PDFmetadata.Title != "" && PDFmetadata.Author != "" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return Metadata{}, err
	}

	return PDFmetadata, nil
}

// extractValue extracts the value for a specific metadata field
func extractValue(line string, field string) string {
	start := strings.Index(line, field+"(")
	if start == -1 {
		return ""
	}
	start += len(field) + 1 // Skip past the field and the opening parenthesis
	end := strings.Index(line[start:], ")")
	if end == -1 {
		return ""
	}
	return line[start : start+end]
}
