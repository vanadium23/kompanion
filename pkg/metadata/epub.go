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
			Name    string `xml:"name,attr"`
			Content string `xml:"content,attr"`
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

	return Metadata{
		ISBN:        metadata.Metadata.ISBN,
		Title:       metadata.Metadata.Title,
		Description: metadata.Metadata.Description,
		Author:      metadata.Metadata.Creator,
		Date:        metadata.Metadata.Date,
		Publisher:   metadata.Metadata.Publisher,
		Language:    metadata.Metadata.Language,
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
