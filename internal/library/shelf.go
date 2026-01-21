package library

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/moroz/uuidv7-go"

	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/storage"
	"github.com/vanadium23/kompanion/pkg/logger"
	"github.com/vanadium23/kompanion/pkg/metadata"
	"github.com/vanadium23/kompanion/pkg/utils"
)

type BookShelf struct {
	storage storage.Storage
	repo    BookRepo
	logger  logger.Interface
}

func NewBookShelf(storage storage.Storage, repo BookRepo, l logger.Interface) *BookShelf {
	return &BookShelf{
		storage: storage,
		repo:    repo,
		logger:  l,
	}
}

func (uc *BookShelf) StoreBook(ctx context.Context, tempFile *os.File, uploadedFilename string) (entity.Book, error) {
	koreaderPartialMD5, err := utils.PartialMD5(tempFile.Name())
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookShelf - StoreBook - PartialMD5: %w", err)
	}
	foundBook, err := uc.repo.GetByFileHash(ctx, koreaderPartialMD5)
	if err == nil {
		return foundBook, entity.ErrBookAlreadyExists
	}

	m, err := metadata.ExtractBookMetadata(tempFile)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookShelf - StoreBook - exractMetadata: %w", err)
	}
	if m.Format == "" {
		return entity.Book{}, errors.New("BookShelf - StoreBook - unknown file format")
	}

	bookID := uuidv7.Generate()
	createDate := time.Now()
	storagepath := fmt.Sprintf("%s/%s.%s", createDate.Format("2006/01/02"), bookID, m.Format)

	err = uc.storage.Write(ctx, tempFile.Name(), storagepath)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookShelf - StoreBook - s.storage.Write: %w", err)
	}
	uc.logger.Info("BookShelf - StoreBook - documentID: %s", koreaderPartialMD5)

	coverPath, err := writeCover(ctx, uc.storage, m.Cover, bookID.String())
	if err != nil {
		uc.logger.Error("BookShelf - StoreBook - writeCover: %s", err)
	}

	book := entity.Book{
		ID:         bookID.String(),
		Title:      m.Title,
		Author:     m.Author,
		Publisher:  m.Publisher,
		Year:       0,
		CreatedAt:  createDate,
		UpdatedAt:  createDate,
		ISBN:       m.ISBN,
		DocumentID: koreaderPartialMD5,
		FilePath:   storagepath,
		Format:     m.Format,
		CoverPath:  coverPath,
	}

	// place in database
	err = uc.repo.Store(
		ctx,
		book,
	)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookShelf - StoreBook - s.repo.Store: %w", err)
	}
	return book, nil
}

func (uc *BookShelf) ListBooks(ctx context.Context,
	sortBy, sortOrder string,
	page, perPage int) (PaginatedBookList, error) {
	books, err := uc.repo.List(ctx, sortBy, sortOrder, page, perPage)
	if err != nil {
		return PaginatedBookList{}, fmt.Errorf("BookShelf - ListBooks - s.repo.List: %w", err)
	}

	totalCount, err := uc.repo.Count(ctx)
	if err != nil {
		return PaginatedBookList{}, fmt.Errorf("BookShelf - ListBooks - s.repo.Count: %w", err)
	}

	pbl := NewPaginatedBookList(
		books,
		perPage,
		page,
		totalCount,
	)

	return pbl, nil
}

func (uc *BookShelf) ViewBook(ctx context.Context, bookID string) (entity.Book, error) {
	book, err := uc.repo.GetById(ctx, bookID)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookShelf - GetBook - s.repo.Get: %w", err)
	}

	return book, nil
}

func (uc *BookShelf) UpdateBookMetadata(ctx context.Context, bookID string, metadata entity.Book) (entity.Book, error) {
	book, err := uc.repo.GetById(ctx, bookID)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookShelf - UpdateBookMetadata - s.repo.Get: %w", err)
	}

	updatedBook := entity.Book{
		ID:        book.ID,
		Title:     utils.If(metadata.Title == "", book.Title, metadata.Title),
		Author:    utils.If(metadata.Author == "", book.Author, metadata.Author),
		Publisher: utils.If(metadata.Publisher == "", book.Publisher, metadata.Publisher),
		Year:      utils.If(metadata.Year == 0, book.Year, metadata.Year),
		ISBN:      utils.If(metadata.ISBN == "", book.ISBN, metadata.ISBN),
		UpdatedAt: time.Now(),
	}

	err = uc.repo.Update(ctx, updatedBook)
	if err != nil {
		return entity.Book{}, fmt.Errorf("BookShelf - UpdateBookMetadata - s.repo.Update: %w", err)
	}

	return updatedBook, nil
}

func (uc *BookShelf) DeleteBook(ctx context.Context, bookID string) error {
	book, err := uc.repo.GetById(ctx, bookID)
	if err != nil {
		return fmt.Errorf("BookShelf - DeleteBook - s.repo.Get: %w", err)
	}
	
	err = uc.deleteCover(ctx, bookID)
	if err != nil {
		return fmt.Errorf("BookShelf - deleteCover - s.repo.Delete: %w", err)
	}

	err = uc.repo.Delete(ctx, book)
	if err != nil {
		return fmt.Errorf("BookShelf - DeleteBook - s.repo.Delete: %w", err)
	}

	err = uc.storage.Delete(ctx, book.FilePath)
	if err != nil {
		return fmt.Errorf("BookShelf - DeleteBook - s.storage.Delete: %w", err)
	}

	return nil
}

func (uc *BookShelf) DownloadBook(ctx context.Context, bookID string) (entity.Book, *os.File, error) {
	book, err := uc.repo.GetById(ctx, bookID)
	if err != nil {
		return book, nil, fmt.Errorf("BookShelf - DownloadBook - s.repo.Get: %s", err)
	}
	file, err := uc.storage.Read(ctx, book.FilePath)
	if err != nil {
		return book, nil, fmt.Errorf("BookShelf - DownloadBook - s.storage.Read: %s", err)
	}
	return book, file, nil
}

func (uc *BookShelf) ViewCover(ctx context.Context, bookID string) (*os.File, error) {
	book, err := uc.repo.GetById(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("BookShelf - ViewCover - s.repo.Get: %s", err)
	}
	if book.CoverPath == "" {
		return nil, fmt.Errorf("BookShelf - ViewCover - no cover")
	}
	file, err := uc.storage.Read(ctx, book.CoverPath)
	if err != nil {
		return nil, fmt.Errorf("BookShelf - ViewCover - s.storage.Read: %s", err)
	}
	return file, nil
}

func writeCover(
	ctx context.Context,
	storage storage.Storage,
	cover []byte,
	bookID string,
) (string, error) {
	if len(cover) == 0 {
		return "", nil
	}
	coverTempFile, err := os.CreateTemp("", "cover")
	if err != nil {
		return "", fmt.Errorf("BookShelf - writeCover - os.CreateTemp: %w", err)
	}
	defer coverTempFile.Close()
	_, err = coverTempFile.Write(cover)
	if err != nil {
		return "", fmt.Errorf("BookShelf - writeCover - coverTempFile.Write: %w", err)
	}

	coverpath := fmt.Sprintf("covers/%s.jpg", bookID)
	err = storage.Write(ctx, coverTempFile.Name(), coverpath)
	if err != nil {
		return "", fmt.Errorf("BookShelf - writeCover - s.storage.Write: %w", err)
	}
	return coverpath, nil
}

func (uc *BookShelf) deleteCover(ctx context.Context, bookID string) error {
	book, err := uc.repo.GetById(ctx, bookID)
	if err != nil {
		return fmt.Errorf("BookShelf - deleteCover - s.repo.Get: %w", err)
	}
	if book.CoverPath == "" {
		return nil
	}
	err = uc.storage.Delete(ctx, book.CoverPath)
	if err != nil {
		return fmt.Errorf("BookShelf - deleteCover - s.Storage.Delete: %w", err)
	}
	return nil
}