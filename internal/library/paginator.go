package library

import "github.com/vanadium23/kompanion/internal/entity"

type PaginatedBookList struct {
	Books []entity.Book
	// for pagination
	totalCount  int
	perPage     int
	currentPage int
}

func NewPaginatedBookList(books []entity.Book, perPage, currentPage, totalCount int) PaginatedBookList {
	return PaginatedBookList{
		Books:       books,
		perPage:     perPage,
		currentPage: currentPage,
		totalCount:  totalCount,
	}
}

func (p PaginatedBookList) TotalPages() int {
	if p.totalCount == 0 {
		return 0
	}
	return (p.totalCount + p.perPage - 1) / p.perPage // Ceiling division
}

func (p PaginatedBookList) HasNext() bool {
	return p.currentPage < p.TotalPages()
}

func (p PaginatedBookList) HasPrev() bool {
	return p.currentPage > 1
}

func (p PaginatedBookList) First() int {
	return 1
}

func (p PaginatedBookList) Last() int {
	return p.TotalPages()
}

func (p PaginatedBookList) Next() int {
	if p.HasNext() {
		return p.currentPage + 1
	}
	return p.currentPage
}

func (p PaginatedBookList) Prev() int {
	if p.HasPrev() {
		return p.currentPage - 1
	}
	return p.currentPage
}

type PaginatedSeriesList struct {
	Series []string
	// for pagination
	totalCount  int
	perPage     int
	currentPage int
}

func NewPaginatedSeriesList(series []string, perPage, currentPage, totalCount int) PaginatedSeriesList {
	return PaginatedSeriesList{
		Series:      series,
		perPage:     perPage,
		currentPage: currentPage,
		totalCount:  totalCount,
	}
}

func (p PaginatedSeriesList) TotalPages() int {
	if p.totalCount == 0 {
		return 0
	}
	return (p.totalCount + p.perPage - 1) / p.perPage // Ceiling division
}

func (p PaginatedSeriesList) HasNext() bool {
	return p.currentPage < p.TotalPages()
}

func (p PaginatedSeriesList) HasPrev() bool {
	return p.currentPage > 1
}

func (p PaginatedSeriesList) First() int {
	return 1
}

func (p PaginatedSeriesList) Last() int {
	return p.TotalPages()
}

func (p PaginatedSeriesList) Next() int {
	if p.HasNext() {
		return p.currentPage + 1
	}
	return p.currentPage
}

func (p PaginatedSeriesList) Prev() int {
	if p.HasPrev() {
		return p.currentPage - 1
	}
	return p.currentPage
}

type PaginatedAuthorList struct {
	Authors []string
	// for pagination
	totalCount  int
	perPage     int
	currentPage int
}

func NewPaginatedAuthorList(authors []string, perPage, currentPage, totalCount int) PaginatedAuthorList {
	return PaginatedAuthorList{
		Authors:     authors,
		perPage:     perPage,
		currentPage: currentPage,
		totalCount:  totalCount,
	}
}

func (p PaginatedAuthorList) TotalPages() int {
	if p.totalCount == 0 {
		return 0
	}
	return (p.totalCount + p.perPage - 1) / p.perPage // Ceiling division
}

func (p PaginatedAuthorList) HasNext() bool {
	return p.currentPage < p.TotalPages()
}

func (p PaginatedAuthorList) HasPrev() bool {
	return p.currentPage > 1
}

func (p PaginatedAuthorList) First() int {
	return 1
}

func (p PaginatedAuthorList) Last() int {
	return p.TotalPages()
}

func (p PaginatedAuthorList) Next() int {
	if p.HasNext() {
		return p.currentPage + 1
	}
	return p.currentPage
}

func (p PaginatedAuthorList) Prev() int {
	if p.HasPrev() {
		return p.currentPage - 1
	}
	return p.currentPage
}
