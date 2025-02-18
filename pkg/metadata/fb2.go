package metadata

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// FictionBook struct for root element
type FictionBook struct {
	XMLName     xml.Name    `xml:"FictionBook"`
	Description Description `xml:"description"`
	Binary      []struct {
		ID      string `xml:"id,attr"`
		Content string `xml:",chardata"`
	} `xml:"binary"`
}

// Description struct contains metadata
type Description struct {
	XMLName xml.Name  `xml:"description"`
	Title   TitleInfo `xml:"title-info"`
	Publish PubInfo   `xml:"publish-info"`
	Author  []Author  `xml:"author"`
}

// TitleInfo struct holds title metadata
type TitleInfo struct {
	XMLName   xml.Name `xml:"title-info"`
	BookTitle string   `xml:"book-title"`
	Coverpage struct {
		Image struct {
			Href string `xml:"href,attr"`
		} `xml:"image"`
	} `xml:"coverpage"`
}

// PubInfo struct holds publisher information
type PubInfo struct {
	XMLName   xml.Name `xml:"publish-info"`
	Publisher string   `xml:"publisher"`
	Year      string   `xml:"year"`
}

// Author struct for author information
type Author struct {
	XMLName   xml.Name `xml:"author"`
	FirstName string   `xml:"first-name"`
	LastName  string   `xml:"last-name"`
}

func getFb2Metatada(tmpFile *os.File) (Metadata, error) {
	// Parse the XML data
	d := xml.NewDecoder(tmpFile)
	d.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch charset {
		case "windows-1251":
			return charmap.Windows1251.NewDecoder().Reader(input), nil
		default:
			return nil, fmt.Errorf("unknown charset: %s", charset)
		}
	}
	var book FictionBook
	err := d.Decode(&book)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return Metadata{}, err
	}
	cover, err := findFB2Cover(book)
	if err != nil {
		fmt.Println("Error finding cover:", err)
	}

	return Metadata{
		Title:     book.Description.Title.BookTitle,
		Publisher: book.Description.Publish.Publisher,
		Cover:     cover,
	}, nil
}

func findFB2Cover(metadata FictionBook) ([]byte, error) {
	coverHref := metadata.Description.Title.Coverpage.Image.Href
	if coverHref == "" {
		return nil, fmt.Errorf("no cover image found")
	}
	coverID := strings.TrimPrefix(coverHref, "#")
	var coverData string
	for _, binary := range metadata.Binary {
		if binary.ID == coverID {
			coverData = binary.Content
			break
		}
	}
	if coverData == "" {
		return nil, fmt.Errorf("cover image not found in binary section")
	}
	decodedImage, err := base64.StdEncoding.DecodeString(coverData)
	if err != nil {
		return nil, err
	}
	return decodedImage, nil
}
