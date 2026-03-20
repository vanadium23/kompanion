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

// epubDescription contains typographic quotes (U+201C, U+201D) and apostrophe (U+2019)
const epubDescription = "(From Wikipedia): Crime and Punishment (Russian: Преступл\u00e9ние и наказ\u00e1ние, Prestupleniye i nakazaniye) is a novel by the Russian author Fyodor Dostoyevsky. It was first published in the literary journal The Russian Messenger in twelve monthly installments during 1866. It was later published in a single volume. It is the second of Dostoyevsky\u2019s full-length novels following his return from ten years of exile in Siberia. Crime and Punishment is the first great novel of his \u201cmature\u201d period of writing. Crime and Punishment focuses on the mental anguish and moral dilemmas of Rodion Raskolnikov, an impoverished ex-student in St. Petersburg who formulates and executes a plan to kill an unscrupulous pawnbroker for her cash. Raskolnikov argues that with the pawnbroker\u2019s money he can perform good deeds to counterbalance the crime, while ridding the world of a worthless vermin. He also commits this murder to test his own hypothesis that some people are naturally capable of such things, and even have the right to do them. Several times throughout the novel, Raskolnikov justifies his actions by comparing himself with Napoleon Bonaparte, believing that murder is permissible in pursuit of a higher purpose."

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
				Title:       "Crime and Punishment",
				Author:      "Fyodor Dostoevsky",
				Language:    "en-us",
				Publisher:   "BB eBooks Co., Ltd.",
				Date:        "2016-01-03",
				Format:      "epub",
				Description: epubDescription,
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
			if tt.want.Description != "" {
				require.Equal(t, tt.want.Description, got.Description)
			}
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
