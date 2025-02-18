package web

import (
	"fmt"
	"os"
	"strconv"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/entity"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/library"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/stats"
	"gitea.chrnv.ru/vanadium23/kompanion/pkg/logger"
	"github.com/gin-gonic/gin"
)

type booksRoutes struct {
	shelf  library.Shelf
	stats  stats.ReadingStats
	logger logger.Interface
}

func newBooksRoutes(handler *gin.RouterGroup, shelf library.Shelf, stats stats.ReadingStats, l logger.Interface) {
	r := &booksRoutes{shelf: shelf, stats: stats, logger: l}

	handler.GET("/", r.listBooks)
	handler.POST("/upload", r.uploadBook)
	handler.GET("/:bookID", r.viewBook)
	handler.POST("/:bookID", r.updateBookMetadata)
	handler.GET("/:bookID/download", r.downloadBook)
	handler.GET("/:bookID/cover", r.viewBookCover)
}

func (r *booksRoutes) listBooks(c *gin.Context) {
	page := 1
	perPage := 12 // Show 12 books per page for grid layout
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	books, err := r.shelf.ListBooks(c.Request.Context(), "created_at", "desc", page, perPage)
	if err != nil {
		c.HTML(500, "error", passStandartContext(c, gin.H{"error": err.Error()}))
		return
	}

	c.HTML(200, "books", passStandartContext(c, gin.H{
		"books": books.Books,
		"pagination": gin.H{
			"currentPage": page,
			"perPage":     perPage,
			"totalPages":  books.TotalPages(),
			"hasNext":     books.HasNext(),
			"hasPrev":     books.HasPrev(),
			"nextPage":    books.Next(),
			"prevPage":    books.Prev(),
			"firstPage":   books.First(),
			"lastPage":    books.Last(),
		},
	}))
}

func (r *booksRoutes) uploadBook(c *gin.Context) {
	// single uploadedBookFile
	uploadedBookFile, err := c.FormFile("book")
	if err != nil {
		r.logger.Error(err, "http - v1 - shelf - uploadBook")
		c.JSON(400, passStandartContext(c, gin.H{"message": "book file is required"}))
		return
	}

	// make by temp files
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		r.logger.Error(err, "http - v1 - shelf - putBook")
		c.JSON(500, passStandartContext(c, gin.H{"message": "bad request"}))
		return
	}
	filepath := tempFile.Name()
	defer os.Remove(filepath)
	defer tempFile.Close()
	c.SaveUploadedFile(uploadedBookFile, filepath)

	book, err := r.shelf.StoreBook(c.Request.Context(), tempFile, uploadedBookFile.Filename)
	if err != nil && err != entity.ErrBookAlreadyExists {
		r.logger.Error(err, "http - v1 - shelf - putBook")
		c.JSON(500, passStandartContext(c, gin.H{"message": "internal server error"}))
		return
	}
	c.Redirect(302, "/books/"+book.ID)
}

func (r *booksRoutes) downloadBook(c *gin.Context) {
	bookID := c.Param("bookID")

	book, file, err := r.shelf.DownloadBook(c.Request.Context(), bookID)
	if err != nil {
		c.JSON(500, passStandartContext(c, gin.H{"message": "internal server error"}))
		return
	}
	defer file.Close()

	c.Header("Content-Disposition", "attachment; filename="+book.Filename())
	c.Header("Content-Type", "application/octet-stream")
	c.File(file.Name())
}

func (r *booksRoutes) viewBook(c *gin.Context) {
	bookID := c.Param("bookID")

	book, err := r.shelf.ViewBook(c.Request.Context(), bookID)
	if err != nil {
		c.HTML(500, "error", passStandartContext(c, gin.H{"error": err.Error()}))
		return
	}

	bookStats, err := r.stats.GetBookStats(c.Request.Context(), book.DocumentID)
	if err != nil {
		r.logger.Error(err, "failed to get book stats")
		bookStats = &stats.BookStats{} // Use empty stats in case of error
	}

	c.HTML(200, "book", passStandartContext(c, gin.H{
		"book":  book,
		"stats": bookStats,
	}))
}

func (r *booksRoutes) updateBookMetadata(c *gin.Context) {
	bookID := c.Param("bookID")

	var metadata entity.Book
	if err := c.ShouldBind(&metadata); err != nil {
		r.logger.Error(err, "http - v1 - shelf - updateBookMetadata")
		// TODO: move to template
		c.JSON(400, passStandartContext(c, gin.H{"message": "invalid request"}))
		return
	}

	book, err := r.shelf.UpdateBookMetadata(c.Request.Context(), bookID, metadata)
	if err != nil {
		r.logger.Error(err, "http - v1 - shelf - updateBookMetadata")
		// TODO: move to template
		c.JSON(500, passStandartContext(c, gin.H{"message": "internal server error"}))
		return
	}

	// TODO: why not redirect?
	c.HTML(200, "book", passStandartContext(c, gin.H{"book": book}))
}

func (r *booksRoutes) viewBookCover(c *gin.Context) {
	bookID := c.Param("bookID")

	book, err := r.shelf.ViewBook(c.Request.Context(), bookID)
	if err != nil {
		c.JSON(500, passStandartContext(c, gin.H{"message": "internal server error"}))
		return
	}

	cover, err := r.shelf.ViewCover(c.Request.Context(), bookID)

	if err != nil {
		width := 600
		height := 800
		backgroundColor := "#6496FA" // Цвет фона (голубой)
		textColor := "white"         // Цвет текста
		title := book.Title
		subtitle := book.Author
		fontSizeTitle := 48
		fontSizeSubtitle := 24

		svgContent := fmt.Sprintf(`
		<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">
			<rect width="100%%" height="100%%" fill="%s" />
			<text x="50%%" y="40%%" font-family="Arial" font-size="%d" fill="%s" text-anchor="middle">%s</text>
			<text x="50%%" y="55%%" font-family="Arial" font-size="%d" fill="%s" text-anchor="middle">%s</text>
		</svg>
		`, width, height, backgroundColor, fontSizeTitle, textColor, title, fontSizeSubtitle, textColor, subtitle)

		c.Data(200, "image/svg+xml", []byte(svgContent))
		return
	}
	c.File(cover.Name())
}
