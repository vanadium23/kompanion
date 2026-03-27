package metadata_test

import (
	"io"
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

func readAll(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	b, _ := io.ReadAll(file)
	return b
}

// epubDescription contains typographic quotes and apostrophes (U+201C, U+201D, U+2019)
const epubDescription = "(From Wikipedia): Crime and Punishment (Russian: Преступл\u00e9ние и наказ\u00e1ние, Prestupleniye i nakazaniye) is a novel by the Russian author Fyodor Dostoyevsky. It was first published in the literary journal The Russian Messenger in twelve monthly installments during 1866. It was later published in a single volume. It is the second of Dostoyevsky\u2019s full-length novels following his return from ten years of exile in Siberia. Crime and Punishment is the first great novel of his \u201cmature\u201d period of writing. Crime and Punishment focuses on the mental anguish and moral dilemmas of Rodion Raskolnikov, an impoverished ex-student in St. Petersburg who formulates and executes a plan to kill an unscrupulous pawnbroker for her cash. Raskolnikov argues that with the pawnbroker\u2019s money he can perform good deeds to counterbalance the crime, while ridding the world of a worthless vermin. He also commits this murder to test his own hypothesis that some people are naturally capable of such things, and even have the right to do them. Several times throughout the novel, Raskolnikov justifies his actions by comparing himself with Napoleon Bonaparte, believing that murder is permissible in pursuit of a higher purpose."

func TestExtractBookMetadata(t *testing.T) {
	projectRoot := getProjectRoot()
	booksPath := filepath.Join(projectRoot, "test", "test_data", "books")
	coversPath := filepath.Join(projectRoot, "test", "test_data", "covers")

	tests := []struct {
		name     string
		fileName string
		want     metadata.Metadata
		err      error
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
				Language:    "en-us",
				Publisher:   "BB eBooks Co., Ltd.",
				Date:        "2016-01-03",
				Author:      "Fyodor Dostoevsky",
				ISBN:        "urn:uuid:12c6fed8-ec29-4343-ab36-9a48312ee01d",
				Title:       "Crime and Punishment",
				Description: epubDescription,
				Format:      "epub",
				Cover:       readAll(filepath.Join(coversPath, "CrimePunishment-EPUB2.jpg")),
			},
		},
		{
			name:     "FB2",
			fileName: "Great Expectations -- Charles Dickens.fb2",
			want: metadata.Metadata{
				Title:       "Great Expectations",
				Description: "Great Expectations chronicles the progress of Pip from childhood through adulthood. As he moves from the marshes of Kent to London society, he encounters a variety of extraordinary characters: from Magwitch, the escaped convict, to Miss Havisham and her ward, the arrogant and beautiful Estella. In this fascinating story, Dickens shows the dangers of being driven by a desire for wealth and social status. Pip must establish a sense of self against the plans which others seem to have for him \u043f\u0457\u0405 and somehow discover a firm set of values and priorities.",
				Format:      "fb2",
				Cover:       readAll(filepath.Join(coversPath, "Great Expectations -- Charles Dickens.jpg")),
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
			require.Equal(t, tt.want, got)
			require.ErrorIs(t, tt.err, err)
		})
	}
}
