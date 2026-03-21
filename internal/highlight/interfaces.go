package highlight

import (
	"context"

	"github.com/vanadium23/kompanion/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=highlight_test

// DocumentInfo contains metadata for a document with highlights.
type DocumentInfo struct {
	PartialMD5 string
	Title      string
	Author     string
}

// HighlightRepo defines the repository interface for highlight persistence.
type HighlightRepo interface {
	Store(ctx context.Context, h entity.Highlight) error
	GetByDocumentID(ctx context.Context, documentID string) ([]entity.Highlight, error)
	GetDocumentsByDevice(ctx context.Context, deviceName string) ([]DocumentInfo, error)
}

// Highlight defines the use case interface for highlight synchronization.
type Highlight interface {
	Sync(ctx context.Context, documentID string, highlights []entity.Highlight, deviceName string) (int, error)
	Fetch(ctx context.Context, documentID string) ([]entity.Highlight, error)
	GetDocumentsByDevice(ctx context.Context, deviceName string) ([]DocumentInfo, error)
}
