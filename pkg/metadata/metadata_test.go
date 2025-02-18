package metadata_test

import (
	"io"
	"os"
	"testing"

	"gitea.chrnv.ru/vanadium23/kompanion/pkg/metadata"
	"github.com/stretchr/testify/require"
)

const pathToTestDataFolder = "../../../test/test_data/books/"

func readAll(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	b, _ := io.ReadAll(file)
	return b
}

func TestExtractBookMetadata(t *testing.T) {

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
				Description: "(From Wikipedia): Crime and Punishment (Russian: Преступлéние и наказáние, Prestupleniye i nakazaniye) is a novel by the Russian author Fyodor Dostoyevsky. It was first published in the literary journal The Russian Messenger in twelve monthly installments during 1866. It was later published in a single volume. It is the second of Dostoyevsky’s full-length novels following his return from ten years of exile in Siberia. Crime and Punishment is the first great novel of his “mature” period of writing. Crime and Punishment focuses on the mental anguish and moral dilemmas of Rodion Raskolnikov, an impoverished ex-student in St. Petersburg who formulates and executes a plan to kill an unscrupulous pawnbroker for her cash. Raskolnikov argues that with the pawnbroker’s money he can perform good deeds to counterbalance the crime, while ridding the world of a worthless vermin. He also commits this murder to test his own hypothesis that some people are naturally capable of such things, and even have the right to do them. Several times throughout the novel, Raskolnikov justifies his actions by comparing himself with Napoleon Bonaparte, believing that murder is permissible in pursuit of a higher purpose.",
				Format:      "epub",
				Cover:       readAll(pathToTestDataFolder + "../covers/CrimePunishment-EPUB2.jpg"),
			},
		},
		{
			name:     "FB2",
			fileName: "Great Expectations -- Charles Dickens.fb2",
			want: metadata.Metadata{
				Title:  "Great Expectations",
				Format: "fb2",
				Cover:  readAll(pathToTestDataFolder + "../covers/Great Expectations -- Charles Dickens.jpg"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.Open(pathToTestDataFolder + tt.fileName)
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
