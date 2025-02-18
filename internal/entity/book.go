package entity

import (
	"errors"
	"strings"
	"time"
)

var ErrBookAlreadyExists = errors.New("Book already exists")

// Book represents a book entity in the database.
type Book struct {
	ID         string    // unique identifier for the book
	Title      string    `form:"title"`     // title of the book
	Author     string    `form:"author"`    // author of the book
	Publisher  string    `form:"publisher"` // publisher of the book
	Year       int       `form:"year"`      // year of publication
	CreatedAt  time.Time // timestamp of when the book was created
	UpdatedAt  time.Time // timestamp of when the book was last updated
	ISBN       string    // ISBN of the book
	DocumentID string    // md5 hash for file content
	FilePath   string    // path to the book file
	Format     string    // format of the book file
	CoverPath  string    // path to the cover image
}

func (b Book) extension() string {
	tmp := strings.Split(b.FilePath, ".")
	return tmp[len(tmp)-1]
}

func (b Book) Filename() string {
	basename := b.ID + "." + b.extension()
	if len(b.Author) == 0 {
		return b.Title + " -- " + basename
	}
	return b.Title + " - " + b.Author + " -- " + basename
}

func (b Book) MimeType() string {
	switch b.extension() {
	case "epub":
		return "application/epub+zip"
	case "pdf":
		return "application/pdf"
	case "mobi":
		return "application/x-mobipocket-ebook"
	case "fb2":
		return "application/fb2"
	default:
		return ""
	}
}
