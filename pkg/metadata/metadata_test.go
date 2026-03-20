package metadata_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vanadium23/kompanion/pkg/metadata"
)

func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "..", "..")
}

func TestExtractBookMetadata(t *testing.T) {
	projectRoot := getProjectRoot()
	booksPath := filepath.Join(projectRoot, "test", "test_data", "books")

	tests := []struct {
		name     string
		fileName string
		want     metadata.Metadata
	}{
		{
			name:     "PDF",
			fileName: "PrincessOfMars-PDF.pdf",
			want: metadata.Metadata{
				Title:  "A Princess of Mars",
				Author: "Edgar Rice Burroughs",
				Format: "pdf",
			},
		},
		{
			name:     "EPUB",
			fileName: "CrimePunishment-EPUB2.epub",
			want: metadata.Metadata{
				Title:     "Crime and Punishment",
				Author:    "Fyodor Dostoevsky",
				Language:  "en-us",
				Publisher: "BB eBooks Co., Ltd.",
				Date:      "2016-01-03",
				Format:    "epub",
			},
		},
		{
			name:     "FB2",
			fileName: "Great Expectations -- Charles Dickens.fb2",
			want: metadata.Metadata{
				Title:  "Great Expectations",
				Format: "fb2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open(filepath.Join(booksPath, tt.fileName))
			if err != nil {
				t.Fatalf("failed to open file: %s", err)
			}
			defer file.Close()

			got, err := metadata.ExtractBookMetadata(file)
			if err != nil {
				t.Fatalf("failed to get metadata: %s", err)
			}
			require.Equal(t, tt.want.Title, got.Title)
			require.Equal(t, tt.want.Author, got.Author)
			require.Equal(t, tt.want.Format, got.Format)
			if tt.want.Language != "" {
				require.Equal(t, tt.want.Language, got.Language)
			}
			if tt.want.Publisher != "" {
				require.Equal(t, tt.want.Publisher, got.Publisher)
			}
			if tt.want.Date != "" {
				require.Equal(t, tt.want.Date, got.Date)
			}
		})
	}
}
