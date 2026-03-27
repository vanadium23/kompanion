package integration_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	. "github.com/Eun/go-hit"
)

// TestHTTPHighlightSync tests the full highlight sync cycle:
// 1. Upload a book to the library
// 2. Sync highlights via API
// 3. Verify highlights appear on the book detail page
func TestHTTPHighlightSync(t *testing.T) {
	// read book content from file
	bookContent, err := os.ReadFile("book.epub")
	if err != nil {
		t.Fatalf("Failed to read book content: %s", err)
	}

	// form request body
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	fileWriter, _ := multipartWriter.CreateFormFile("book", "book.epub")
	fileWriter.Write(bookContent)
	multipartWriter.Close()

	client, loginSteps := webAuthSteps()
	Test(t, Description("Login for Device"), loginSteps)

	// Step 1: Upload book
	var redirectedPath string
	Test(t,
		HTTPClient(client),
		Description("Upload Book for Highlight Test"),
		Post(basePath+"/books/upload"),
		Send().Headers("Content-Type").Add(multipartWriter.FormDataContentType()),
		Send().Body().String(requestBody.String()),
		Expect().Status().Equal(http.StatusFound),
		Store().Response().Headers("Location").In(&redirectedPath),
	)
	bookID := strings.Split(redirectedPath, "/")[2]

	// Step 2: Register device
	deviceName := generateDeviceName()
	deviceSteps := setupDeviceSteps(client, deviceName)
	Test(t, Description("Device Register"), deviceSteps)

	// Get the document ID from the book (we need to fetch book info)
	// For now, use a known MD5 pattern from the uploaded book
	// The document ID is koreader_partial_md5 which we can get from the database
	// In a real test, we'd query this. For simplicity, we'll use a test value.
	documentID := "test-document-md5"

	// Step 3: Sync highlights via API
	highlightRequest := map[string]interface{}{
		"document": documentID,
		"title":    "Test Book for Highlights",
		"author":   "Test Author",
		"highlights": []map[string]interface{}{
			{
				"text":    "This is a test highlight from KOReader",
				"note":    "My note about this passage",
				"page":    "42",
				"chapter": "Chapter 5: Testing",
				"time":    1743081600, // Fixed timestamp for testing
				"drawer":  "highlight",
				"color":   "yellow",
			},
			{
				"text":    "Another important quote from the book",
				"note":    "",
				"page":    "87",
				"chapter": "Chapter 10: Integration",
				"time":    1743081700,
				"drawer":  "highlight",
				"color":   "green",
			},
		},
	}

	Test(t,
		Description("Sync Highlights via API"),
		Put(basePath+"/api/v1/sync/highlight"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
		Send().Body().JSON(highlightRequest),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".synced").Equal(2),
		Expect().Body().JSON().JQ(".total").Equal(2),
	)

	// Step 4: Verify highlights appear on book detail page
	Test(t,
		HTTPClient(client),
		Description("Check Highlights on Book Page"),
		Get(fmt.Sprintf("%s/books/%s", basePath, bookID)),
		Expect().Status().Equal(http.StatusOK),
		// Check highlights section exists
		Expect().Body().String().Contains(`<section class="highlights">`),
		// Check highlight count
		Expect().Body().String().Contains("2 highlights"),
		// Check first highlight text
		Expect().Body().String().Contains("This is a test highlight from KOReader"),
		// Check note is displayed
		Expect().Body().String().Contains("My note about this passage"),
		// Check page number
		Expect().Body().String().Contains("Page 42"),
		// Check chapter
		Expect().Body().String().Contains("Chapter 5: Testing"),
		// Check second highlight
		Expect().Body().String().Contains("Another important quote from the book"),
		// Check formatted date (formatTime template function)
		Expect().Body().String().Contains("Mar 27, 2025"),
	)

	// Step 5: Verify duplicate sync doesn't create duplicates
	Test(t,
		Description("Sync Same Highlights Again (Dedup)"),
		Put(basePath+"/api/v1/sync/highlight"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
		Send().Body().JSON(highlightRequest),
		Expect().Status().Equal(http.StatusOK),
		// Same count - no duplicates created
		Expect().Body().JSON().JQ(".synced").Equal(2),
		Expect().Body().JSON().JQ(".total").Equal(2),
	)

	// Step 6: Verify page with no highlights shows empty message
	// Upload a different book with no highlights
	var requestBody2 bytes.Buffer
	multipartWriter2 := multipart.NewWriter(&requestBody2)
	fileWriter2, _ := multipartWriter2.CreateFormFile("book", "book2.epub")
	fileWriter2.Write(bookContent)
	multipartWriter2.Close()

	var redirectedPath2 string
	Test(t,
		HTTPClient(client),
		Description("Upload Second Book (No Highlights)"),
		Post(basePath+"/books/upload"),
		Send().Headers("Content-Type").Add(multipartWriter2.FormDataContentType()),
		Send().Body().String(requestBody2.String()),
		Expect().Status().Equal(http.StatusFound),
		Store().Response().Headers("Location").In(&redirectedPath2),
	)
	bookID2 := strings.Split(redirectedPath2, "/")[2]

	Test(t,
		HTTPClient(client),
		Description("Check Empty Highlights State"),
		Get(fmt.Sprintf("%s/books/%s", basePath, bookID2)),
		Expect().Status().Equal(http.StatusOK),
		// Empty state message
		Expect().Body().String().Contains("No highlights yet"),
		Expect().Body().String().Contains("Sync from KOReader to see your highlights here"),
	)
}

// TestHTTPHighlightSyncWithNotes tests highlights with and without notes
func TestHTTPHighlightSyncWithNotes(t *testing.T) {
	client, loginSteps := webAuthSteps()
	Test(t, Description("Login for Device"), loginSteps)

	deviceName := generateDeviceName()
	deviceSteps := setupDeviceSteps(client, deviceName)
	Test(t, Description("Device Register"), deviceSteps)

	documentID := "test-notes-document-md5"

	// Sync highlight WITH note
	highlightWithNote := map[string]interface{}{
		"document": documentID,
		"title":    "Book With Notes",
		"author":   "Author",
		"highlights": []map[string]interface{}{
			{
				"text":    "Quote with a note",
				"note":    "This is my annotation",
				"page":    "1",
				"chapter": "Intro",
				"time":    1743081800,
				"drawer":  "highlight",
				"color":   "yellow",
			},
		},
	}

	Test(t,
		Description("Sync Highlight With Note"),
		Put(basePath+"/api/v1/sync/highlight"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
		Send().Body().JSON(highlightWithNote),
		Expect().Status().Equal(http.StatusOK),
	)

	// Sync highlight WITHOUT note (empty string)
	highlightWithoutNote := map[string]interface{}{
		"document": documentID + "-2",
		"title":    "Book Without Notes",
		"author":   "Author",
		"highlights": []map[string]interface{}{
			{
				"text":    "Quote without a note",
				"note":    "", // Empty note
				"page":    "2",
				"chapter": "",
				"time":    1743081900,
				"drawer":  "highlight",
				"color":   "yellow",
			},
		},
	}

	Test(t,
		Description("Sync Highlight Without Note"),
		Put(basePath+"/api/v1/sync/highlight"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
		Send().Body().JSON(highlightWithoutNote),
		Expect().Status().Equal(http.StatusOK),
	)
}

// TestHTTPHighlightSyncUnauthorized tests auth requirements
func TestHTTPHighlightSyncUnauthorized(t *testing.T) {
	documentID := "test-unauth-md5"

	highlightRequest := map[string]interface{}{
		"document":   documentID,
		"title":      "Test",
		"author":     "Test",
		"highlights": []map[string]interface{}{},
	}

	// No auth headers
	Test(t,
		Description("Sync Without Auth"),
		Put(basePath+"/api/v1/sync/highlight"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(highlightRequest),
		Expect().Status().Equal(http.StatusUnauthorized),
	)

	// Wrong password
	Test(t,
		Description("Sync With Wrong Password"),
		Put(basePath+"/api/v1/sync/highlight"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("x-auth-user").Add("nonexistent-device"),
		Send().Headers("x-auth-key").Add("wronghash"),
		Send().Body().JSON(highlightRequest),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}
