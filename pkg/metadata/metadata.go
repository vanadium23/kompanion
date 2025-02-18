package metadata

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type Metadata struct {
	ISBN        string
	Title       string
	Description string
	Author      string
	Date        string
	Publisher   string
	Language    string
	Format      string
	Cover       []byte
}

// ExtractBookMetadata extracts metadata from a book file
func ExtractBookMetadata(tempFile *os.File) (Metadata, error) {
	extension, err := guessExtention(tempFile)
	if err != nil {
		return Metadata{}, err
	}
	var m Metadata
	switch extension {
	case "pdf":
		m, err = extractPdfMetadata(tempFile)
		if err != nil {
			return Metadata{}, err
		}
	case "epub":
		m, err = getEpubMetadata(tempFile)
		if err != nil {
			return Metadata{}, err
		}
	case "fb2":
		m, err = getFb2Metatada(tempFile)
		if err != nil {
			return Metadata{}, err
		}
	}
	m.Format = extension
	return m, nil
}

func guessExtention(file *os.File) (string, error) {
	// TODO: move extensions to enum
	data := make([]byte, 100*1024)
	_, err := file.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		return "", err
	}
	mimeType := http.DetectContentType(data)
	fmt.Println(mimeType)
	switch mimeType {
	case "application/pdf":
		return "pdf", nil
	case "application/epub+zip":
		return "epub", nil
	case "application/zip":
		return "epub", nil
	case "application/x-fictionbook+xml":
		return "fb2", nil
	case "text/xml; charset=utf-8":
		return "fb2", nil
	default:
		return "", nil
	}
}
