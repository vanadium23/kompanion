package library

import "gitea.chrnv.ru/vanadium23/kompanion/internal/entity"

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
