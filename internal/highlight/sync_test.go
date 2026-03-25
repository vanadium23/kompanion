package highlight_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/highlight"
)

func TestHighlightSync_NewHighlightSync(t *testing.T) {
	t.Parallel()

	mockCtl := gomock.NewController(t)
	repo := NewMockHighlightRepo(mockCtl)

	uc := highlight.NewHighlightSync(repo)
	require.NotNil(t, uc)
}

func TestHighlightSync_Sync_EmptyHighlights(t *testing.T) {
	t.Parallel()

	uc, _ := mockedHighlightSync(t)

	synced, err := uc.Sync(context.Background(), "doc123", []entity.Highlight{}, "device1")

	require.NoError(t, err)
	require.Equal(t, 0, synced)
}

func TestHighlightSync_Sync_SetsFields(t *testing.T) {
	t.Parallel()

	documentID := "doc123"
	deviceName := "device1"
	highlights := []entity.Highlight{
		{
			Text:      "highlight text",
			Page:      "42",
			Timestamp: 1700000000,
		},
	}

	uc, repo := mockedHighlightSync(t)

	// Expect Store to be called
	repo.EXPECT().Store(context.Background(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, h entity.Highlight) error {
			require.Equal(t, documentID, h.DocumentID)
			require.Equal(t, deviceName, h.AuthDeviceName)
			require.NotEmpty(t, h.HighlightHash)
			require.False(t, h.CreatedAt.IsZero())
			return nil
		},
	)

	synced, err := uc.Sync(context.Background(), documentID, highlights, deviceName)

	require.NoError(t, err)
	require.Equal(t, 1, synced)
}

func TestHighlightSync_Sync_ContinuesOnError(t *testing.T) {
	t.Parallel()

	documentID := "doc123"
	deviceName := "device1"
	errInternal := errors.New("internal error")
	highlights := []entity.Highlight{
		{Text: "first", Page: "1", Timestamp: 1},
		{Text: "second", Page: "2", Timestamp: 2},
		{Text: "third", Page: "3", Timestamp: 3},
	}

	uc, repo := mockedHighlightSync(t)

	// First fails, second succeeds, third fails
	repo.EXPECT().Store(context.Background(), gomock.Any()).Return(errInternal)
	repo.EXPECT().Store(context.Background(), gomock.Any()).Return(nil)
	repo.EXPECT().Store(context.Background(), gomock.Any()).Return(errInternal)

	synced, err := uc.Sync(context.Background(), documentID, highlights, deviceName)

	require.NoError(t, err)
	require.Equal(t, 1, synced) // Only one succeeded
}

func TestHighlightSync_Sync_GeneratesHash(t *testing.T) {
	t.Parallel()

	highlights := []entity.Highlight{
		{Text: "test text", Page: "42", Timestamp: 1700000000},
	}

	uc, repo := mockedHighlightSync(t)

	var capturedHighlight entity.Highlight
	repo.EXPECT().Store(context.Background(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, h entity.Highlight) error {
			capturedHighlight = h
			return nil
		},
	)

	_, err := uc.Sync(context.Background(), "doc", highlights, "device")
	require.NoError(t, err)

	// Hash should be consistent MD5 of "text:page:timestamp"
	require.NotEmpty(t, capturedHighlight.HighlightHash)
	// Verify hash is consistent by generating it the same way
	expectedHash := generateExpectedHash("test text", "42", 1700000000)
	require.Equal(t, expectedHash, capturedHighlight.HighlightHash)
}

func TestHighlightSync_Fetch(t *testing.T) {
	t.Parallel()

	documentID := "doc123"
	expectedHighlights := []entity.Highlight{
		{DocumentID: documentID, Text: "highlight 1"},
		{DocumentID: documentID, Text: "highlight 2"},
	}

	uc, repo := mockedHighlightSync(t)
	repo.EXPECT().GetByDocumentID(context.Background(), documentID).Return(expectedHighlights, nil)

	result, err := uc.Fetch(context.Background(), documentID)

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, expectedHighlights, result)
}

func TestHighlightSync_Fetch_Error(t *testing.T) {
	t.Parallel()

	errInternal := errors.New("database error")

	uc, repo := mockedHighlightSync(t)
	repo.EXPECT().GetByDocumentID(context.Background(), "doc123").Return(nil, errInternal)

	result, err := uc.Fetch(context.Background(), "doc123")

	require.Error(t, err)
	require.Nil(t, result)
}

func mockedHighlightSync(t *testing.T) (*highlight.HighlightSyncUseCase, *MockHighlightRepo) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	repo := NewMockHighlightRepo(mockCtl)
	uc := highlight.NewHighlightSync(repo)

	return uc, repo
}

// generateExpectedHash mirrors the internal hash generation for testing
func generateExpectedHash(text, page string, timestamp int64) string {
	data := fmt.Sprintf("%s:%s:%d", text, page, timestamp)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
