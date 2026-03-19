package sync_test

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/vanadium23/kompanion/internal/entity"
	"github.com/vanadium23/kompanion/internal/sync"
)

func TestProgressFetch(t *testing.T) {
	t.Parallel()

	bookID := "bookID"
	errInternalServErr := errors.New("internal server error")

	tests := []struct {
		name string
		mock func(*MockProgressRepo)
		res  entity.Progress
		err  error
	}{
		{
			name: "empty result",
			mock: func(repo *MockProgressRepo) {
				repo.EXPECT().GetBookHistory(context.Background(), bookID, 1).Return(nil, nil)
			},
			res: entity.Progress{},
			err: nil,
		},
		{
			name: "first result",
			mock: func(repo *MockProgressRepo) {
				repo.EXPECT().GetBookHistory(context.Background(), bookID, 1).Return(
					[]entity.Progress{{
						Document: bookID,
					}, {
						Document: "anotherBookID",
					}}, nil)
			},
			res: entity.Progress{
				Document: bookID,
			},
			err: nil,
		},
		{
			name: "result with error",
			mock: func(repo *MockProgressRepo) {
				repo.EXPECT().GetBookHistory(context.Background(), bookID, 1).Return(nil, errInternalServErr)
			},
			res: entity.Progress{},
			err: errInternalServErr,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			progressSync, repo := mockedProgress(t)
			tc.mock(repo)

			res, err := progressSync.Fetch(context.Background(), bookID)

			require.Equal(t, tc.res, res)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestProgressSync(t *testing.T) {
	t.Parallel()

	progressDoc := entity.Progress{
		Document:  "bookID",
		Timestamp: 1,
	}
	errInternalServErr := errors.New("internal server error")

	tests := []struct {
		name string
		mock func(*MockProgressRepo)
		err  error
	}{
		{
			name: "empty result",
			mock: func(repo *MockProgressRepo) {
				repo.EXPECT().Store(context.Background(), progressDoc).Return(nil)
			},
			err: nil,
		},
		{
			name: "result with error",
			mock: func(repo *MockProgressRepo) {
				repo.EXPECT().Store(context.Background(), progressDoc).Return(errInternalServErr)
			},
			err: errInternalServErr,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			progressSync, repo := mockedProgress(t)
			tc.mock(repo)

			_, err := progressSync.Sync(context.Background(), progressDoc)

			require.ErrorIs(t, err, tc.err)
		})
	}
}

func mockedProgress(t *testing.T) (*sync.ProgressSyncUseCase, *MockProgressRepo) {
	t.Helper()

	mockCtl := gomock.NewController(t)

	repo := NewMockProgressRepo(mockCtl)
	progress := sync.NewProgressSync(repo)

	return progress, repo
}
