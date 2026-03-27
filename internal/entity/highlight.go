// Package entity provides domain entities for the application.
package entity

import "time"

// HighlightSyncRequest binds KOReader JSON payload for highlight sync.
type HighlightSyncRequest struct {
	Document string     `json:"document"` // maps to koreader_partial_md5
	Title    string     `json:"title"`
	Author   string     `json:"author"`
	Entries  []SyncEntry `json:"highlights"`
}

// SyncEntry represents an individual highlight from KOReader.
type SyncEntry struct {
	Text    string `json:"text"`
	Note    string `json:"note"`
	Page    string `json:"page"`
	Chapter string `json:"chapter"`
	Time    int64  `json:"time"` // Unix timestamp
	Drawer  string `json:"drawer"` // "highlight" or "note"
	Color   string `json:"color"`
}

// Highlight represents a stored highlight in the database.
type Highlight struct {
	KoreaderPartialMD5 string    // maps to document from KOReader
	TextHash           string    // SHA-256 hash of Text field for deduplication
	Text               string
	Note               string
	Page               string
	Chapter            string
	Time               int64
	Drawer             string
	Color              string
	DeviceName         string
	CreatedAt          time.Time
}
