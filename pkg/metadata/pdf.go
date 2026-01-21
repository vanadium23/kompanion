package metadata

import (
	"os"
	"bytes"
	"unicode/utf8"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"regexp"
	"strings"
	"io"
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
	var PDFmetadata Metadata

	const chunksize = 64 * 1024
	const tailsize = 1024

	buf := make([]byte, chunksize)
	tail := make([]byte, tailsize)

	for {
		n, err := tmpFile.Read(buf)

		block := append(tail, buf[:n]...)

		if bytes.Contains(block, []byte("/Title")) {
			PDFmetadata.Title = clean(extractValue(block, "/Title"))

		}

		if bytes.Contains(block, []byte("/Author")) {
			PDFmetadata.Author = clean(extractValue(block, "/Author"))
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return Metadata{}, err
		}

		// Break early if we've found all fields
		if PDFmetadata.Title != "" && PDFmetadata.Author != "" {
			break
		}
	}

	// use file name if no title found
	if PDFmetadata.Title == "" {
		parts := strings.Split(tmpFile.Name(), "/")
		parts = strings.Split(parts[len(parts) - 1], ".")
		PDFmetadata.Title = parts[0]
	}

	return PDFmetadata, nil
}

// extractValue extracts the value for a specific metadata field
func extractValue(block []byte, field string) string {
	fieldIndex := bytes.Index(block, []byte(field))
	start := bytes.Index(block[fieldIndex:], []byte("(")) + fieldIndex + 1
	end := findLiteralEnd(block[start:]) + start
	value := block[start:end]

	if utf8.Valid(value) {
		return string(value)
	}

	// remove escape symbol
	regex := regexp.MustCompile(`\\(.)`)
	value = regex.ReplaceAll(value, []byte("$1"))

	// UTF16BE
	if len(value) >= 2 && value[0] == 0xFE && value[1] == 0xFF {
		return decodeUTF16("be", value)
	}
	// UTF16LE
	if len(value) >= 2 && value[0] == 0xFF && value[1] == 0xFE {
		return decodeUTF16("le", value)
	}

	return ""
}

// find the literal end, excluding parentheis in the field
func findLiteralEnd(b []byte) int {
	depth := 1
	escaped := false
	for i := 0; i < len(b); i++ {
		c := b[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '(' {
			depth++
			continue
		}
		if c == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// decode UTF-16 content
func decodeUTF16(t string, data []byte) string {
	var dec transform.Transformer
	if t == "be" {
		dec = unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewDecoder()
	} else if t == "le" {
		dec = unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()
	}
	
	r := transform.NewReader(bytes.NewReader(data), dec)
	b, _ := io.ReadAll(r)
	return string(b)
}

// remove non-utf8 bytes
func clean(s string) string {
	r := strings.NewReplacer(
		"\x00", "",
		"\xFF", "",
		"\xFE", "",
	)

	return r.Replace(s)
}