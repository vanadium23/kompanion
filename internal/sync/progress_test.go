package sync_test

import (
	"context"
	"errors"
	"testing"

	"gitea.chrnv.ru/vanadium23/kompanion/internal/entity"
	"gitea.chrnv.ru/vanadium23/kompanion/internal/sync"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type test struct {
	name string
	mock func()
	res  interface{}
	err  error
}

func TestProgressFetch(t *testing.T) {
	t.Parallel()

	progressSync, repo := mockedProgress(t)

	bookID := "bookID"
	errInternalServErr := errors.New("internal server error")

	tests := []test{
		{
			name: "empty result",
			mock: func() {
				repo.EXPECT().GetBookHistory(context.Background(), bookID, 1).Return(nil, nil)
			},
			res: entity.Progress{},
			err: nil,
		},
		{
			name: "first result",
			mock: func() {
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
			mock: func() {
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

			tc.mock()

			res, err := progressSync.Fetch(context.Background(), "bookID")

			require.Equal(t, res, tc.res)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestProgressSync(t *testing.T) {
	t.Parallel()

	progressSync, repo := mockedProgress(t)

	progressDoc := entity.Progress{
		Document: "bookID",
	}
	errInternalServErr := errors.New("internal server error")

	tests := []test{
		{
			name: "empty result",
			mock: func() {
				repo.EXPECT().Store(context.Background(), progressDoc).Return(nil)
			},
			res: nil,
			err: nil,
		},
		{
			name: "result with error",
			mock: func() {
				repo.EXPECT().Store(context.Background(), progressDoc).Return(errInternalServErr)
			},
			res: nil,
			err: errInternalServErr,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.mock()

			_, err := progressSync.Sync(context.Background(), progressDoc)

			require.ErrorIs(t, err, tc.err)
		})
	}
}

func mockedProgress(t *testing.T) (*sync.ProgressSyncUseCase, *MockProgressRepo) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := NewMockProgressRepo(mockCtl)

	progress := sync.NewProgressSync(repo)

	return progress, repo
}
