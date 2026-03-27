package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vanadium23/kompanion/internal/auth"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/library"
	"github.com/vanadium23/kompanion/pkg/logger"
)

type booksRoutes struct {
	shelf library.Shelf
	l     logger.Interface
}

func newBooksRoutes(handler *gin.RouterGroup, shelf library.Shelf, a auth.AuthInterface, l logger.Interface) {
	r := &booksRoutes{shelf: shelf, l: l}

	h := handler.Group("/books")
	h.Use(authDeviceMiddleware(a, l))
	{
		h.GET("", r.listBooks)
	}
}

// BookResponse represents a single book in the API response.
type BookResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Description string  `json:"description,omitempty"`
	Publisher   string  `json:"publisher,omitempty"`
	Year        int     `json:"year,omitempty"`
	Series      string  `json:"series,omitempty"`
	SeriesIndex *string `json:"series_index,omitempty"`
	ISBN        string  `json:"isbn,omitempty"`
	Format      string  `json:"format"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// BooksListResponse represents the paginated books list API response.
type BooksListResponse struct {
	Books        []BookResponse `json:"books"`
	TotalPages   int            `json:"total_pages"`
	CurrentPage  int            `json:"current_page"`
	HasNext      bool           `json:"has_next"`
	HasPrev      bool           `json:"has_prev"`
}

// listBooks handles GET /api/v1/books
// Query parameters:
// - search: search string (optional)
// - sort: field to sort by (title, author, series, created_at)
// - order: asc or desc (default: asc)
// - page: page number (default: 1)
// - limit: items per page (default: 50)
func (r *booksRoutes) listBooks(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse search query parameters
	searchQuery := c.Query("search")
	sortBy := c.Query("sort")
	sortOrder := c.Query("order")

	// Default sort order
	if sortOrder == "" {
		sortOrder = "asc"
	}

	query := entity.SearchQuery{
		Search:    searchQuery,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Page:      page,
		Limit:     limit,
	}

	books, err := r.shelf.ListBooks(c.Request.Context(), query)
	if err != nil {
		r.l.Error(err, "failed to list books")
		errorResponse(c, http.StatusInternalServerError, "failed to list books")
		return
	}

	// Convert to response format
	bookResponses := make([]BookResponse, len(books.Books))
	for i, book := range books.Books {
		var seriesIndex *string
		if book.SeriesIndex != nil && book.SeriesIndex.Valid {
			val := book.SeriesIndex.Decimal.String()
			seriesIndex = &val
		}

		bookResponses[i] = BookResponse{
			ID:          book.ID,
			Title:       book.Title,
			Author:      book.Author,
			Description: book.Description,
			Publisher:   book.Publisher,
			Year:        book.Year,
			Series:      book.Series,
			SeriesIndex: seriesIndex,
			ISBN:        book.ISBN,
			Format:      book.Format,
			CreatedAt:   book.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   book.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := BooksListResponse{
		Books:       bookResponses,
		TotalPages:  books.TotalPages(),
		CurrentPage: page,
		HasNext:     books.HasNext(),
		HasPrev:     books.HasPrev(),
	}

	c.JSON(http.StatusOK, response)
}
