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
	AtomTime = "2006-01-02T15:04:05Z"
	DirMime  = "application/atom+xml;profile=opds-catalog;kind=navigation"
	DirRel   = "subsection"
	FileRel  = "http://opds-spec.org/acquisition"
	CoverRel = "http://opds-spec.org/cover"
)

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
	Rel  string `xml:"rel,attr,ommitempty"`
}

// Entry is a struct of OPDS entry properties.
type Entry struct {
	ID      string  `xml:"id"`
	Updated string  `xml:"updated"`
	Title   string  `xml:"title"`
	Author  Author  `xml:"author,ommitempty"`
	Summary Summary `xml:"summary,ommitempty"`
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

func translateBooksToEntries(books []entity.Book) []Entry {
	entries := make([]Entry, 0, len(books))
	for _, book := range books {
		entries = append(entries, Entry{
			ID:      book.ID,
			Updated: book.UpdatedAt.Format(AtomTime),
			Title:   book.Title,
			Author: Author{
				Name: book.Author,
			},
			Summary: Summary{
				Type: "text",
				Text: truncateText(book.Description, 300),
			},
			Link: []Link{
				{
					Href: fmt.Sprintf("/opds/book/%s/download", book.ID),
					Type: book.MimeType(),
					Rel:  FileRel,
					// Mtime: book.UpdatedAt.Format(AtomTime),
				},
			},
		})
	}
	return entries
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

// formNavLinks creates navigation links for OPDS pagination.
// If searchQuery is not empty, it is included in the pagination links.
func formNavLinks(baseURL string, searchQuery string, books library.PaginatedBookList) []Link {
	// Build query string prefix
	queryPrefix := "?"
	if searchQuery != "" {
		queryPrefix = fmt.Sprintf("?search=%s&", url.QueryEscape(searchQuery))
	}

	links := []Link{
		{
			Href: baseURL,
			Type: DirMime,
			Rel:  "start",
		},
		{
			Href: fmt.Sprintf("%spage=%d", queryPrefix, books.Last()),
			Type: DirMime,
			Rel:  "last",
		},
	}
	if books.HasNext() {
		links = append(links, Link{
			Href: fmt.Sprintf("%spage=%d", queryPrefix, books.Next()),
			Type: DirMime,
			Rel:  "next",
		})
	}
	if books.HasPrev() {
		links = append(links, Link{
			Href: fmt.Sprintf("%spage=%d", queryPrefix, books.Prev()),
			Type: DirMime,
			Rel:  "prev",
		})
	}
	return links
}
