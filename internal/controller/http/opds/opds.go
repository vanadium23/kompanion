package opds

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/library"
)

const (
	AtomTime    = "2006-01-02T15:04:05Z"
	DirMime     = "application/atom+xml;profile=opds-catalog;kind=navigation"
	DirRel      = "subsection"
	FileRel     = "http://opds-spec.org/acquisition"
	CoverRel    = "http://opds-spec.org/cover"
	ThumbnailRel = "http://opds-spec.org/thumbnail"
)

// OpenSearchDescription is the OpenSearch description document structure
type OpenSearchDescription struct {
	XMLName     xml.Name `xml:"OpenSearchDescription"`
	Xmlns       string   `xml:"xmlns,attr"`
	ShortName   string   `xml:"ShortName"`
	Description string   `xml:"Description"`
	URL         OpenSearchURL `xml:"Url"`
}

// OpenSearchURL is the URL element in OpenSearch description
type OpenSearchURL struct {
	Type     string `xml:"type,attr"`
	Template string `xml:"template,attr"`
}

// Feed is a main frame of OPDS.
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Xmlns   string   `xml:"xmlns,attr"`
	Updated string   `xml:"updated"`
	Link    []Link   `xml:"link"`
	Entry   []Entry  `xml:"entry"`
}

// Link is link properties.
type Link struct {
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
}

// Entry is a struct of OPDS entry properties.
type Entry struct {
	ID      string  `xml:"id"`
	Updated string  `xml:"updated"`
	Title   string  `xml:"title"`
	Author  Author  `xml:"author,omitempty"`
	Summary Summary `xml:"summary,omitempty"`
	Link    []Link  `xml:"link"`
}

type Author struct {
	Name string `xml:"name"`
}

type Summary struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

func BuildFeed(id, title, href string, entries []Entry, additionalLinks []Link) *Feed {
	finalLinks := []Link{
		{
			Href: "/opds/",
			Type: DirMime,
			Rel:  "start",
		},
		{
			Href: href,
			Type: DirMime,
			Rel:  "self",
		},
		{
			Href: "/opds/search/{searchTerms}/",
			Type: "application/atom+xml",
			Rel:  "search",
		},
	}
	finalLinks = append(finalLinks, additionalLinks...)
	return &Feed{
		ID:      id,
		Title:   title,
		Xmlns:   "http://www.w3.org/2005/Atom",
		Updated: time.Now().UTC().Format(AtomTime),
		Link:    finalLinks,
		Entry:   entries,
	}
}

func BuildOpenSearchDescription() *OpenSearchDescription {
	return &OpenSearchDescription{
		Xmlns:       "http://a9.com/-/spec/opensearch/1.1/",
		ShortName:   "KOmpanion",
		Description: "Search KOmpanion library",
		URL: OpenSearchURL{
			Type:     "application/atom+xml",
			Template: "/opds/search/{searchTerms}/",
		},
	}
}

func translateBooksToEntries(books []entity.Book) []Entry {
	entries := make([]Entry, 0, len(books))
	for _, book := range books {
		links := []Link{
			{
				Href: fmt.Sprintf("/opds/book/%s/download", book.ID),
				Type: book.MimeType(),
				Rel:  FileRel,
				// Mtime: book.UpdatedAt.Format(AtomTime),
			},
		}
		// Add cover and thumbnail links if book has a cover
		if book.CoverPath != "" {
			links = append(links,
				Link{
					Href: fmt.Sprintf("/opds/book/%s/cover", book.ID),
					Type: "image/jpeg",
					Rel:  ThumbnailRel,
				},
				Link{
					Href: fmt.Sprintf("/opds/book/%s/cover", book.ID),
					Type: "image/jpeg",
					Rel:  CoverRel,
				},
			)
		}
		entry := Entry{
			ID:      book.ID,
			Updated: book.UpdatedAt.Format(AtomTime),
			Title:   book.Title,
			Author: Author{
				Name: book.Author,
			},
			Link: links,
		}
		// Only include summary if it's not empty
		if book.Summary != "" {
			entry.Summary = Summary{
				Type: "text",
				Text: book.Summary,
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func formNavLinks(baseURL string, books library.PaginatedBookList) []Link {
	links := []Link{
		{
			Href: baseURL,
			Type: DirMime,
			Rel:  "start",
		},
		{
			Href: fmt.Sprintf("%s?page=%d", baseURL, books.Last()),
			Type: DirMime,
			Rel:  "last",
		},
	}
	if books.HasNext() {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?page=%d", baseURL, books.Next()),
			Type: DirMime,
			Rel:  "next",
		})
	}
	if books.HasPrev() {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?page=%d", baseURL, books.Prev()),
			Type: DirMime,
			Rel:  "prev",
		})
	}
	return links
}

func translateSeriesToEntries(series []string) []Entry {
	entries := make([]Entry, 0, len(series))
	for _, s := range series {
		encodedName := url.PathEscape(s)
		entry := Entry{
			ID:      "urn:kompanion:series:" + s,
			Updated: time.Now().UTC().Format(AtomTime),
			Title:   s,
			Link: []Link{
				{
					Href: "/opds/series/" + encodedName + "/",
					Type: "application/atom+xml;type=feed;profile=opds-catalog",
				},
			},
		}
		entries = append(entries, entry)
	}
	return entries
}

func formSeriesNavLinks(baseURL string, series library.PaginatedSeriesList) []Link {
	links := []Link{
		{
			Href: baseURL,
			Type: DirMime,
			Rel:  "start",
		},
		{
			Href: fmt.Sprintf("%s?page=%d", baseURL, series.Last()),
			Type: DirMime,
			Rel:  "last",
		},
	}
	if series.HasNext() {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?page=%d", baseURL, series.Next()),
			Type: DirMime,
			Rel:  "next",
		})
	}
	if series.HasPrev() {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?page=%d", baseURL, series.Prev()),
			Type: DirMime,
			Rel:  "prev",
		})
	}
	return links
}

func translateAuthorsToEntries(authors []string) []Entry {
	entries := make([]Entry, 0, len(authors))
	for _, a := range authors {
		encodedName := url.PathEscape(a)
		entry := Entry{
			ID:      "urn:kompanion:author:" + a,
			Updated: time.Now().UTC().Format(AtomTime),
			Title:   a,
			Link: []Link{
				{
					Href: "/opds/authors/" + encodedName + "/",
					Type: "application/atom+xml;type=feed;profile=opds-catalog",
				},
			},
		}
		entries = append(entries, entry)
	}
	return entries
}

func formAuthorsNavLinks(baseURL string, authors library.PaginatedAuthorList) []Link {
	links := []Link{
		{
			Href: baseURL,
			Type: DirMime,
			Rel:  "start",
		},
		{
			Href: fmt.Sprintf("%s?page=%d", baseURL, authors.Last()),
			Type: DirMime,
			Rel:  "last",
		},
	}
	if authors.HasNext() {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?page=%d", baseURL, authors.Next()),
			Type: DirMime,
			Rel:  "next",
		})
	}
	if authors.HasPrev() {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?page=%d", baseURL, authors.Prev()),
			Type: DirMime,
			Rel:  "prev",
		})
	}
	return links
}
