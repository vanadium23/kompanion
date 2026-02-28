package metadata

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// container.xml struct
type Container struct {
	Container xml.Name `xml:"container"`
	XMLNS     string   `xml:"xmlns,attr"`
	Version   string   `xml:"version,attr"`
	Rootfiles []struct {
		Rootfile  xml.Name `xml:"rootfile"`
		FullPath  string   `xml:"full-path,attr"`
		MediaType string   `xml:"media-type,attr"`
	} `xml:"rootfiles>rootfile"`
}

// content.opf struct
type EpubMetadata struct {
	Metadata struct {
		ISBN        string `xml:"identifier"`
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Creator     string `xml:"creator"`
		Date        string `xml:"date"`
		Publisher   string `xml:"publisher"`
		Language    string `xml:"language"`
		Format      string `xml:"format"`
		Meta        []struct {
			Name     string `xml:"name,attr"`
			Content  string `xml:"content,attr"`
			Property string `xml:"property,attr"`
			Refines  string `xml:"refines,attr"`
			ID       string `xml:"id,attr"`
			Value    string `xml:",chardata"`
		} `xml:"meta"`
	} `xml:"metadata"`
	Manifest struct {
		Items []struct {
			ID   string `xml:"id,attr"`
			Href string `xml:"href,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
}

func getEpubMetadata(tmpFile *os.File) (Metadata, error) {
	metadataFilepath := ""

	fileInfo, err := tmpFile.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return Metadata{}, err
	}

	reader, err := zip.NewReader(tmpFile, fileInfo.Size())
	if err != nil {
		return Metadata{}, err
	}

	for _, f := range reader.File {
		if f.Name == "META-INF/container.xml" {
			container, _ := parseContainerXML(f)
			metadataFilepath = container.Rootfiles[0].FullPath
			break
		}
	}
	var metadata EpubMetadata

	for _, f := range reader.File {
		if f.Name == metadataFilepath {
			metadata = parseMetadata(f)
			break
		}
	}

	cover := findEpubCover(reader, metadata)
	series, seriesIndex := extractEpubSeries(metadata)

	return Metadata{
		ISBN:        metadata.Metadata.ISBN,
		Title:       metadata.Metadata.Title,
		Description: metadata.Metadata.Description,
		Author:      metadata.Metadata.Creator,
		Date:        metadata.Metadata.Date,
		Publisher:   metadata.Metadata.Publisher,
		Language:    metadata.Metadata.Language,
		Series:      series,
		SeriesIndex: seriesIndex,
		Cover:       cover,
	}, nil
}

func findEpubCover(reader *zip.Reader, metadata EpubMetadata) []byte {
	var coverID string
	for _, meta := range metadata.Metadata.Meta {
		if meta.Name == "cover" {
			coverID = meta.Content
			break
		}
	}

	for _, item := range metadata.Manifest.Items {
		if item.ID == coverID {
			for _, f := range reader.File {
				if strings.Contains(f.Name, item.Href) {
					content, _ := readFileContent(f)
					return content
				}
			}
		}
	}
	return nil
}

func parseMetadata(f *zip.File) EpubMetadata {
	content, _ := readFileContent(f)

	return unmarshalMetaDataXML(content)
}

func parseContainerXML(f *zip.File) (Container, error) {
	byteValue, err := readFileContent(f)
	if err != nil {
		return Container{}, err
	}

	return unmarshalContainerXML(byteValue), nil
}

func readFileContent(f *zip.File) ([]byte, error) {
	reader, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	byteValue, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return byteValue, nil
}

func unmarshalMetaDataXML(byteValue []byte) EpubMetadata {
	var meta EpubMetadata
	xml.Unmarshal(byteValue, &meta)
	return meta
}

func unmarshalContainerXML(byteValue []byte) Container {
	var container Container
	xml.Unmarshal(byteValue, &container)
	return container
}

// extractEpubSeries extracts series name and index from EPUB metadata.
// It supports both EPUB 3.2 spec format (belongs-to-collection, group-position)
// and legacy calibre format (calibre:series, calibre:series_index).
func extractEpubSeries(metadata EpubMetadata) (string, string) {
	var series string
	var seriesIndex string
	var collectionID string

	// First pass: find series name and collect refines targets
	for _, meta := range metadata.Metadata.Meta {
		// EPUB 3.2 format: belongs-to-collection property
		if meta.Property == "belongs-to-collection" {
			// EPUB 3.2 uses element content (Value), not content attribute
			series = meta.Value
			if series == "" {
				series = meta.Content
			}
			if meta.ID != "" {
				collectionID = "#" + meta.ID
			}
		}
		// Legacy calibre format
		if meta.Name == "calibre:series" {
			series = meta.Content
		}
		if meta.Name == "calibre:series_index" {
			seriesIndex = meta.Content
		}
	}

	// Second pass: find series index via refines (EPUB 3.2 format)
	if collectionID != "" && seriesIndex == "" {
		for _, meta := range metadata.Metadata.Meta {
			if meta.Refines == collectionID && meta.Property == "group-position" {
				// EPUB 3.2 uses element content (Value), not content attribute
				seriesIndex = meta.Value
				if seriesIndex == "" {
					seriesIndex = meta.Content
				}
				break
			}
		}
	}

	return series, seriesIndex
}
