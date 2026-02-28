package metadata

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractEpubSeries_CalibreFormat(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <dc:title>Test Book</dc:title>
    <meta name="calibre:series" content="The Expanse"/>
    <meta name="calibre:series_index" content="1"/>
  </metadata>
</package>`

	var metadata EpubMetadata
	err := xml.Unmarshal([]byte(xmlContent), &metadata)
	require.NoError(t, err)

	series, seriesIndex := extractEpubSeries(metadata)
	require.Equal(t, "The Expanse", series)
	require.Equal(t, "1", seriesIndex)
}

func TestExtractEpubSeries_CalibreFormat_FractionalIndex(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <dc:title>Test Book</dc:title>
    <meta name="calibre:series" content="The Expanse"/>
    <meta name="calibre:series_index" content="2.5"/>
  </metadata>
</package>`

	var metadata EpubMetadata
	err := xml.Unmarshal([]byte(xmlContent), &metadata)
	require.NoError(t, err)

	series, seriesIndex := extractEpubSeries(metadata)
	require.Equal(t, "The Expanse", series)
	require.Equal(t, "2.5", seriesIndex)
}

func TestExtractEpubSeries_Epub32Format(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<package version="3.0" xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <dc:title>Leviathan Wakes</dc:title>
    <meta id="collection-id" property="belongs-to-collection">The Expanse</meta>
    <meta refines="#collection-id" property="group-position">1</meta>
  </metadata>
</package>`

	var metadata EpubMetadata
	err := xml.Unmarshal([]byte(xmlContent), &metadata)
	require.NoError(t, err)

	series, seriesIndex := extractEpubSeries(metadata)
	require.Equal(t, "The Expanse", series)
	require.Equal(t, "1", seriesIndex)
}

func TestExtractEpubSeries_Epub32Format_FractionalIndex(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<package version="3.0" xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <dc:title>Strange Dogs</dc:title>
    <meta id="collection-id" property="belongs-to-collection">The Expanse</meta>
    <meta refines="#collection-id" property="group-position">6.5</meta>
  </metadata>
</package>`

	var metadata EpubMetadata
	err := xml.Unmarshal([]byte(xmlContent), &metadata)
	require.NoError(t, err)

	series, seriesIndex := extractEpubSeries(metadata)
	require.Equal(t, "The Expanse", series)
	require.Equal(t, "6.5", seriesIndex)
}

func TestExtractEpubSeries_NoSeries(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <dc:title>Standalone Book</dc:title>
  </metadata>
</package>`

	var metadata EpubMetadata
	err := xml.Unmarshal([]byte(xmlContent), &metadata)
	require.NoError(t, err)

	series, seriesIndex := extractEpubSeries(metadata)
	require.Equal(t, "", series)
	require.Equal(t, "", seriesIndex)
}

func TestExtractEpubSeries_CalibreSeriesOnly(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf">
  <metadata>
    <dc:title>Test Book</dc:title>
    <meta name="calibre:series" content="Some Series"/>
  </metadata>
</package>`

	var metadata EpubMetadata
	err := xml.Unmarshal([]byte(xmlContent), &metadata)
	require.NoError(t, err)

	series, seriesIndex := extractEpubSeries(metadata)
	require.Equal(t, "Some Series", series)
	require.Equal(t, "", seriesIndex)
}

func TestFb2SequenceParsing(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Leviathan Wakes</book-title>
      <sequence name="The Expanse" number="1"/>
    </title-info>
  </description>
</FictionBook>`

	var book FictionBook
	err := xml.Unmarshal([]byte(xmlContent), &book)
	require.NoError(t, err)

	require.NotNil(t, book.Description.Title.Sequence)
	require.Equal(t, "The Expanse", book.Description.Title.Sequence.Name)
	require.Equal(t, "1", book.Description.Title.Sequence.Number)
}

func TestFb2SequenceParsing_FractionalNumber(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Strange Dogs</book-title>
      <sequence name="The Expanse" number="6.5"/>
    </title-info>
  </description>
</FictionBook>`

	var book FictionBook
	err := xml.Unmarshal([]byte(xmlContent), &book)
	require.NoError(t, err)

	require.NotNil(t, book.Description.Title.Sequence)
	require.Equal(t, "The Expanse", book.Description.Title.Sequence.Name)
	require.Equal(t, "6.5", book.Description.Title.Sequence.Number)
}

func TestFb2SequenceParsing_NoSequence(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Standalone Book</book-title>
    </title-info>
  </description>
</FictionBook>`

	var book FictionBook
	err := xml.Unmarshal([]byte(xmlContent), &book)
	require.NoError(t, err)

	require.Nil(t, book.Description.Title.Sequence)
}
