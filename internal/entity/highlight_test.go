package entity

import (
	"encoding/json"
	"testing"
	"time"
)

// TestHighlightStructExists tests that Highlight struct exists
func TestHighlightStructExists(t *testing.T) {
	var h Highlight
	_ = h // Use the variable to avoid unused variable error
}

// TestHighlightTextField tests that Highlight struct has Text field (required)
func TestHighlightTextField(t *testing.T) {
	h := Highlight{Text: "sample highlighted text"}
	if h.Text != "sample highlighted text" {
		t.Error("Text field not set correctly")
	}
}

// TestHighlightNoteField tests that Highlight struct has Note field (optional)
func TestHighlightNoteField(t *testing.T) {
	h := Highlight{Note: "my note"}
	if h.Note != "my note" {
		t.Error("Note field not set correctly")
	}
}

// TestHighlightPageField tests that Highlight struct has Page field (required)
func TestHighlightPageField(t *testing.T) {
	h := Highlight{Page: "42"}
	if h.Page != "42" {
		t.Error("Page field not set correctly")
	}
}

// TestHighlightChapterField tests that Highlight struct has Chapter field (optional)
func TestHighlightChapterField(t *testing.T) {
	h := Highlight{Chapter: "Chapter 1"}
	if h.Chapter != "Chapter 1" {
		t.Error("Chapter field not set correctly")
	}
}

// TestHighlightTimestampField tests that Highlight struct has Timestamp field (int64)
func TestHighlightTimestampField(t *testing.T) {
	h := Highlight{Timestamp: 1234567890}
	if h.Timestamp != 1234567890 {
		t.Error("Timestamp field not set correctly")
	}
}

// TestHighlightDrawerAndColorFields tests that Highlight struct has Drawer and Color fields
func TestHighlightDrawerAndColorFields(t *testing.T) {
	h := Highlight{
		Drawer: "underscore",
		Color:  "yellow",
	}
	if h.Drawer != "underscore" {
		t.Error("Drawer field not set correctly")
	}
	if h.Color != "yellow" {
		t.Error("Color field not set correctly")
	}
}

// TestHighlightDeviceFields tests that Highlight struct has DocumentID, Device, DeviceID, AuthDeviceName, HighlightHash, CreatedAt fields
func TestHighlightDeviceFields(t *testing.T) {
	now := time.Now()
	h := Highlight{
		ID:             "test-id",
		DocumentID:     "abc123",
		Device:         "Kobo",
		DeviceID:       "device-uuid",
		AuthDeviceName: "my-kobo",
		HighlightHash:  "hash123",
		CreatedAt:      now,
	}
	if h.ID != "test-id" {
		t.Error("ID field not set correctly")
	}
	if h.DocumentID != "abc123" {
		t.Error("DocumentID field not set correctly")
	}
	if h.Device != "Kobo" {
		t.Error("Device field not set correctly")
	}
	if h.DeviceID != "device-uuid" {
		t.Error("DeviceID field not set correctly")
	}
	if h.AuthDeviceName != "my-kobo" {
		t.Error("AuthDeviceName field not set correctly")
	}
	if h.HighlightHash != "hash123" {
		t.Error("HighlightHash field not set correctly")
	}
	if !h.CreatedAt.Equal(now) {
		t.Error("CreatedAt field not set correctly")
	}
}

// TestHighlightJSONTagsMatchKOReader tests that JSON tags match KOReader field names
func TestHighlightJSONTagsMatchKOReader(t *testing.T) {
	h := Highlight{
		ID:         "test-id",
		DocumentID: "abc123",
		Text:       "highlighted text",
		Note:       "my note",
		Page:       "42",
		Chapter:    "Chapter 1",
		Timestamp:  1234567890,
		Drawer:     "underscore",
		Color:      "yellow",
		Device:     "Kobo",
		DeviceID:   "device-uuid",
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("Failed to marshal Highlight: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check KOReader field names
	if _, ok := result["text"]; !ok {
		t.Error("JSON should contain 'text' field")
	}
	if _, ok := result["note"]; !ok {
		t.Error("JSON should contain 'note' field")
	}
	if _, ok := result["page"]; !ok {
		t.Error("JSON should contain 'page' field")
	}
	if _, ok := result["chapter"]; !ok {
		t.Error("JSON should contain 'chapter' field")
	}
	if _, ok := result["time"]; !ok {
		t.Error("JSON should contain 'time' field (KOReader uses 'time' not 'timestamp')")
	}
	if _, ok := result["drawer"]; !ok {
		t.Error("JSON should contain 'drawer' field")
	}
	if _, ok := result["color"]; !ok {
		t.Error("JSON should contain 'color' field")
	}
	if _, ok := result["device"]; !ok {
		t.Error("JSON should contain 'device' field")
	}
	if _, ok := result["device_id"]; !ok {
		t.Error("JSON should contain 'device_id' field")
	}

	// AuthDeviceName and HighlightHash should NOT be in JSON (json:"-")
	if _, ok := result["AuthDeviceName"]; ok {
		t.Error("AuthDeviceName should not be in JSON (should use json:\"-\")")
	}
	if _, ok := result["HighlightHash"]; ok {
		t.Error("HighlightHash should not be in JSON (should use json:\"-\")")
	}
}
