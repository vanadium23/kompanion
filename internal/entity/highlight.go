package entity

import "time"

// Highlight represents a book highlight synced from KOReader.
type Highlight struct {
	ID             string    `json:"id"`
	DocumentID     string    `json:"document"`          // MD5 hash (koreader_partial_md5)
	Text           string    `json:"text" binding:"required"`
	Note           string    `json:"note"`
	Page           string    `json:"page"`
	Chapter        string    `json:"chapter"`
	Timestamp      int64     `json:"time"`
	Drawer         string    `json:"drawer"`            // highlight style
	Color          string    `json:"color"`             // highlight color
	Device         string    `json:"device"`
	DeviceID       string    `json:"device_id"`
	AuthDeviceName string    `json:"-"`                 // set from middleware, not from KOReader
	HighlightHash  string    `json:"-"`                 // generated for deduplication
	CreatedAt      time.Time `json:"created_at"`
}
