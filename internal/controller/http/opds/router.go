package opds

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/library"
	"github.com/vanadium23/kompanion/internal/sync"
	"github.com/vanadium23/kompanion/pkg/logger"
)

type OPDSRouter struct {
	books  library.Shelf
	logger logger.Interface
}

// opdsError sends an XML error response for OPDS clients
func opdsError(c *gin.Context, status int, code int, message string) {
	c.XML(status, gin.H{
		"XMLName": xml.Name{Local: "error"},
		"Code":    code,
		"Message": message,
	})
}

func NewRouter(
	handler *gin.Engine,
	l logger.Interface,
	a auth.AuthInterface,
	p sync.Progress,
	shelf library.Shelf) {
	sh := &OPDSRouter{shelf, l}

	h := handler.Group("/opds")
	h.Use(basicAuth(a))
	{
		h.GET("/", sh.listShelves)
		h.GET("/newest/", sh.listNewest)
		h.GET("/book/:bookID/download", sh.downloadBook)
		h.GET("/book/:bookID/cover", sh.getCover)
		h.GET("/search.xml", sh.openSearchDescription)
		h.GET("/search/:searchTerms/", sh.searchBooks)
		h.GET("/series/", sh.listSeries)
		h.GET("/series/:seriesName/", sh.listBooksBySeries)
		h.GET("/authors/", sh.listAuthors)
		h.GET("/authors/:authorName/", sh.listBooksByAuthor)
	}
}

func (r *OPDSRouter) listShelves(c *gin.Context) {
	shelves := []Entry{
		{
			ID:      "urn:kompanion:newest",
			Updated: time.Now().UTC().Format(AtomTime),
			Title:   "By Newest",
			Link: []Link{
				{
					Href: "/opds/newest/",
					Type: "application/atom+xml;type=feed;profile=opds-catalog",
				},
			},
		},
		{
			ID:      "urn:kompanion:series",
			Updated: time.Now().UTC().Format(AtomTime),
			Title:   "By Series",
			Link: []Link{
				{
					Href: "/opds/series/",
					Type: "application/atom+xml;type=feed;profile=opds-catalog",
				},
			},
		},
		{
			ID:      "urn:kompanion:authors",
			Updated: time.Now().UTC().Format(AtomTime),
			Title:   "By Author",
			Link: []Link{
				{
					Href: "/opds/authors/",
					Type: "application/atom+xml;type=feed;profile=opds-catalog",
				},
			},
		},
	}
	links := []Link{}
	feed := BuildFeed("urn:kompanion:main", "KOmpanion library", "/opds", shelves, links)
	c.XML(http.StatusOK, feed)
}

func (r *OPDSRouter) listNewest(c *gin.Context) {
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	books, err := r.books.ListBooks(c.Request.Context(), "created_at", "desc", page, 10)
	if err != nil {
		r.logger.Error("failed to list newest books", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}
	baseUrl := "/opds/newest/"
	entries := translateBooksToEntries(books.Books)
	navLinks := formNavLinks(baseUrl, books)
	feed := BuildFeed("urn:kompanion:newest", "KOmpanion library", baseUrl, entries, navLinks)
	c.XML(http.StatusOK, feed)
}

func (r *OPDSRouter) downloadBook(c *gin.Context) {
	bookID := c.Param("bookID")

	book, file, err := r.books.DownloadBook(c.Request.Context(), bookID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			r.logger.Error(err, "http - opds - downloadBook - not found")
			opdsError(c, http.StatusNotFound, 404, "Book not found")
			return
		}
		r.logger.Error(err, "http - opds - downloadBook")
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}
	defer file.Close()

	c.Header("Content-Disposition", "attachment; filename="+book.Filename())
	c.Header("Content-Type", "application/octet-stream")
	c.File(file.Name())
}

func (r *OPDSRouter) getCover(c *gin.Context) {
	bookID := c.Param("bookID")

	file, err := r.books.ViewCover(c.Request.Context(), bookID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			r.logger.Error("http - opds - getCover - not found", err)
			opdsError(c, http.StatusNotFound, 404, "Cover not found")
			return
		}
		r.logger.Error("http - opds - getCover", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}
	defer file.Close() // Ensure file is closed on all return paths

	stat, err := file.Stat()
	if err != nil {
		r.logger.Error("http - opds - getCover - file stat", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Error reading cover file")
		return
	}

	c.DataFromReader(http.StatusOK, stat.Size(), "image/jpeg", file, nil)
}

func (r *OPDSRouter) openSearchDescription(c *gin.Context) {
	c.Header("Content-Type", "application/opensearchdescription+xml")
	c.XML(http.StatusOK, BuildOpenSearchDescription())
}

func (r *OPDSRouter) searchBooks(c *gin.Context) {
	searchTerms := c.Param("searchTerms")

	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	books, err := r.books.SearchBooks(c.Request.Context(), searchTerms, page, 10)
	if err != nil {
		r.logger.Error("failed to search books", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}

	// URL encode search terms for proper pagination links
	encodedSearchTerms := url.PathEscape(searchTerms)
	baseUrl := "/opds/search/" + encodedSearchTerms + "/"
	entries := translateBooksToEntries(books.Books)
	navLinks := formNavLinks(baseUrl, books)
	feed := BuildFeed("urn:kompanion:search", "KOmpanion library - Search: "+searchTerms, baseUrl, entries, navLinks)
	c.XML(http.StatusOK, feed)
}

func (r *OPDSRouter) listSeries(c *gin.Context) {
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	series, err := r.books.ListSeries(c.Request.Context(), page, 10)
	if err != nil {
		r.logger.Error("failed to list series", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}

	baseUrl := "/opds/series/"
	entries := translateSeriesToEntries(series.Series)
	navLinks := formSeriesNavLinks(baseUrl, series)
	feed := BuildFeed("urn:kompanion:series", "KOmpanion library - Series", baseUrl, entries, navLinks)
	c.XML(http.StatusOK, feed)
}

func (r *OPDSRouter) listBooksBySeries(c *gin.Context) {
	seriesName := c.Param("seriesName")

	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	books, err := r.books.ListBooksBySeries(c.Request.Context(), seriesName, page, 10)
	if err != nil {
		r.logger.Error("failed to list books by series", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}

	// URL encode series name for proper pagination links
	encodedSeriesName := url.PathEscape(seriesName)
	baseUrl := "/opds/series/" + encodedSeriesName + "/"
	entries := translateBooksToEntries(books.Books)
	navLinks := formNavLinks(baseUrl, books)
	feed := BuildFeed("urn:kompanion:series:"+seriesName, "KOmpanion library - Series: "+seriesName, baseUrl, entries, navLinks)
	c.XML(http.StatusOK, feed)
}

func (r *OPDSRouter) listAuthors(c *gin.Context) {
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	authors, err := r.books.ListAuthors(c.Request.Context(), page, 10)
	if err != nil {
		r.logger.Error("failed to list authors", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}

	baseUrl := "/opds/authors/"
	entries := translateAuthorsToEntries(authors.Authors)
	navLinks := formAuthorsNavLinks(baseUrl, authors)
	feed := BuildFeed("urn:kompanion:authors", "KOmpanion library - Authors", baseUrl, entries, navLinks)
	c.XML(http.StatusOK, feed)
}

func (r *OPDSRouter) listBooksByAuthor(c *gin.Context) {
	authorName := c.Param("authorName")

	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	books, err := r.books.ListBooksByAuthor(c.Request.Context(), authorName, page, 10)
	if err != nil {
		r.logger.Error("failed to list books by author", err)
		opdsError(c, http.StatusServiceUnavailable, 503, "Service unavailable")
		return
	}

	// URL encode author name for proper pagination links
	encodedAuthorName := url.PathEscape(authorName)
	baseUrl := "/opds/authors/" + encodedAuthorName + "/"
	entries := translateBooksToEntries(books.Books)
	navLinks := formNavLinks(baseUrl, books)
	feed := BuildFeed("urn:kompanion:author:"+authorName, "KOmpanion library - Author: "+authorName, baseUrl, entries, navLinks)
	c.XML(http.StatusOK, feed)
}

func basicAuth(auth auth.AuthInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="KOmpanion OPDS"`)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
			c.Abort()
			return
		}
		if !auth.CheckDevicePassword(c.Request.Context(), username, password, true) {
			if !auth.CheckPassword(c.Request.Context(), username, password) {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized", "code": 2001})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
