package integration_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Eun/go-hit"
	. "github.com/Eun/go-hit"
	petname "github.com/dustinkirkland/golang-petname"
)

const (
	// Attempts connection
	host       = "app:8080"
	healthPath = "http://" + host + "/healthcheck"
	attempts   = 20

	// HTTP REST
	basePath = "http://" + host
)

func TestMain(m *testing.M) {
	err := healthCheck(attempts)
	if err != nil {
		log.Fatalf("Integration tests: host %s is not available: %s", host, err)
	}

	log.Printf("Integration tests: host %s is available", host)

	code := m.Run()
	os.Exit(code)
}

func healthCheck(attempts int) error {
	var err error

	for attempts > 0 {
		err = Do(Get(healthPath), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}

		log.Printf("Integration tests: url %s is not available, attempts left: %d", healthPath, attempts)

		time.Sleep(time.Second)

		attempts--
	}

	return err
}

func TestWebFooterVersion(t *testing.T) {
	Test(t,
		Description("Footer Version"),
		Get(basePath+"/"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains("github.com/vanadium23/kompanion"),
		Expect().Body().String().Contains("integration"),
	)
}

// New tests based on controller/web
// login/page
func TestWebAuthUser(t *testing.T) {
	username, password := grabTestUser()

	Test(t,
		Description("Auth Incorrect User"),
		Post(basePath+"/auth/login"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("username").Add("incorrect_username"),
		Send().Body().FormValues("password").Add(password),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains("incorrect password"),
	)

	Test(t,
		Description("Auth Incorrect Password"),
		Post(basePath+"/auth/login"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("username").Add(username),
		Send().Body().FormValues("password").Add("incorrect_password"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains("incorrect password"),
	)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	Test(t,
		HTTPClient(client),
		Description("Auth Correct"),
		Post(basePath+"/auth/login"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("username").Add(username),
		Send().Body().FormValues("password").Add(password),
		Expect().Status().Equal(http.StatusFound),
		Expect().Headers("Set-Cookie").Len().Equal(1),
		Expect().Headers("Set-Cookie").First().Contains("session"),
	)
}

// devices
func TestWebDevice(t *testing.T) {
	client, loginSteps := webAuthSteps()

	Test(t,
		Description("Login for Device"),
		loginSteps)

	// no password
	Test(t,
		Description("Device Register without Password"),
		HTTPClient(client),
		Post(basePath+"/devices/add"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("device_name").Add("custom"),
		Expect().Status().Equal(http.StatusBadRequest),
	)

	// success
	device_name := generateDeviceName()
	Test(t,
		Description("Device Register"),
		HTTPClient(client),
		Post(basePath+"/devices/add"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("device_name").Add(device_name),
		Send().Body().FormValues("password").Add("password"),
		Expect().Status().Equal(http.StatusFound),
	)

	// duplicated
	Test(t,
		Description("Device Register"),
		HTTPClient(client),
		Post(basePath+"/devices/add"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("device_name").Add(device_name),
		Send().Body().FormValues("password").Add("password"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// syncs (only with devices)
// HTTP test koreader sync progress feature for registred device
func TestHTTPKoreaderSyncProgress(t *testing.T) {
	type document struct {
		Document   string  `json:"document"`
		Percentage float64 `json:"percentage"`
		Progress   string  `json:"progress"`
		Device     string  `json:"device"`
		DeviceID   string  `json:"device_id"`
	}

	doc := document{
		Document:   "test",
		Percentage: 1.0,
		Progress:   "test",
		Device:     "test",
		DeviceID:   "test",
	}
	client, loginSteps := webAuthSteps()
	Test(t, Description("Login for Device"), loginSteps)
	deviceName := generateDeviceName()
	deviceSteps := setupDeviceSteps(client, deviceName)
	Test(t, Description("Device Register"), deviceSteps)

	// check auth
	Test(t,
		Description("Koreader Put Document Progress"),
		Put(basePath+"/syncs/progress"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(doc),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add("incorrect_password"),
		Expect().Status().Equal(http.StatusUnauthorized),
	)

	// put from device
	Test(t,
		Description("Koreader Put Document Progress"),
		Put(basePath+"/syncs/progress"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(doc),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".timestamp").NotEqual(0),
		Expect().Body().JSON().JQ(".document.document").Equal(doc.Document),
	)

	Test(t,
		Description("Koreader Get Document Progress"),
		Get(basePath+"/syncs/progress/"+doc.Document),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".percentage").Equal(doc.Percentage),
		Expect().Body().JSON().JQ(".timestamp").NotEqual(0),
		Expect().Body().JSON().JQ(".device").Equal(deviceName),
	)
}

// HTTP GET /users/auth
func TestHTTPAuth(t *testing.T) {
	username, password := grabTestUser()
	Test(t,
		Description("Auth With Incorrect Password"),
		Get(basePath+"/users/auth"),
		Send().Headers("x-auth-user").Add(username),
		Send().Headers("x-auth-key").Add("incorrect_password"),
		Expect().Status().Equal(http.StatusUnauthorized),
		Expect().Body().JSON().JQ(".code").Equal(2001),
		Expect().Body().JSON().JQ(".message").Equal("Unauthorized"),
	)

	Test(t,
		Description("Auth With User not permitted"),
		Get(basePath+"/users/auth"),
		Send().Headers("x-auth-user").Add(username),
		Send().Headers("x-auth-key").Add(hashSyncPassword(password)),
		Expect().Status().Equal(http.StatusUnauthorized),
		Expect().Body().JSON().JQ(".code").Equal(2001),
		Expect().Body().JSON().JQ(".message").Equal("Unauthorized"),
	)

	client, loginSteps := webAuthSteps()
	Test(t, Description("Login for Device"), loginSteps)
	deviceName := generateDeviceName()
	deviceSteps := setupDeviceSteps(client, deviceName)
	Test(t, Description("Device Register"), deviceSteps)

	Test(t,
		Description("Auth Success"),
		Get(basePath+"/users/auth"),
		Send().Headers("x-auth-user").Add(deviceName),
		Send().Headers("x-auth-key").Add(hashSyncPassword("password")),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".code").Equal(200),
		Expect().Body().JSON().JQ(".message").Equal("OK"),
	)
}

// library

// HTTP test kompanion shelf feature
func TestHTTPKompanionShelf(t *testing.T) {
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

	// put book
	var redirectedPath string
	Test(t,
		HTTPClient(client),
		Description("Kompanion Put Book"),
		Post(basePath+"/books/upload"),
		Send().Headers("Content-Type").Add(multipartWriter.FormDataContentType()),
		Send().Body().String(requestBody.String()),
		Expect().Status().Equal(http.StatusFound),
		Store().Response().Headers("Location").In(&redirectedPath),
	)
	bookID := strings.Split(redirectedPath, "/")[2]

	// list books
	Test(t,
		HTTPClient(client),
		Description("Kompanion List Books"),
		Get(basePath+"/books/"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains(bookID),
	)

	// get book
	Test(t,
		HTTPClient(client),
		Description("Kompanion Get Book"),
		Get(fmt.Sprintf("%s/books/%s", basePath, bookID)),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains(bookID),
	)

	// update metadata
	updatedTitle := "The Egg"
	updatedAuthor := "Andy Weier"
	Test(t,
		HTTPClient(client),
		Description("Kompanion Update Metadata"),
		Post(fmt.Sprintf("%s/books/%s", basePath, bookID)),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("title").Add(updatedTitle),
		Send().Body().FormValues("author").Add(updatedAuthor),
		Expect().Status().Equal(http.StatusOK),
	)
	// get updated book
	Test(t,
		HTTPClient(client),
		Description("Kompanion Check New Metadata on Book Page"),
		Get(fmt.Sprintf("%s/books/%s", basePath, bookID)),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains(bookID),
		Expect().Body().String().Contains(updatedTitle),
		Expect().Body().String().Contains(updatedAuthor),
	)
	// download book
	// "attachment; filename=The Egg.epub"
	filename := fmt.Sprintf("%s - %s -- %s.epub", updatedTitle, updatedAuthor, bookID)
	Test(t,
		HTTPClient(client),
		Description("Kompanion Download Book"),
		Get(fmt.Sprintf("%s/books/%s/download", basePath, bookID)),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().Bytes().Equal(bookContent),
		Expect().Headers("Content-Disposition").Equal("attachment; filename="+filename),
	)
}

// stats
func TestWebStats(t *testing.T) {
	_, password := grabTestUser()
	// arrange
	client, loginSteps := webAuthSteps()
	Test(t, Description("Login for Device"), loginSteps)
	deviceName := generateDeviceName()
	deviceSteps := setupDeviceSteps(client, deviceName)
	Test(t, Description("Device Register"), deviceSteps)

	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(deviceName+":"+password))

	statsContent, err := os.ReadFile("../test/test_data/koreader/koreader_statistics_example.sqlite3")
	if err != nil {
		t.Fatalf("Failed to read stats content: %s", err)
	}

	Test(t,
		HTTPClient(client),
		Description("Kompanion Get Stats Before"),
		Get(basePath+"/stats/?from=2025-02-01&to=2025-02-28"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().NotContains("Crime and Punishment"),
	)

	Test(t,
		Description("Kompanion Upload Stats via WebDAV"),
		Put(basePath+"/webdav/statistics.sqlite3"),
		Send().Headers("Authorization").Add(basicAuth),
		Send().Body().Bytes(statsContent),
		Expect().Status().Equal(http.StatusCreated),
	)

	// wait for stats to be processed
	// TODO: find a better way to wait for stats to be processed
	time.Sleep(2 * time.Second)

	Test(t,
		HTTPClient(client),
		Description("Kompanion Get Stats After"),
		Get(basePath+"/stats/?from=2025-02-01&to=2025-02-28"),
		Send().Headers("Authorization").Add(basicAuth),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains("Crime and Punishment"),
	)

	// regress for uploading same file
	// https://github.com/vanadium23/kompanion/issues/22
	Test(t,
		Description("Kompanion Upload Stats via WebDAV"),
		Put(basePath+"/webdav/statistics.sqlite3"),
		Send().Headers("Authorization").Add(basicAuth),
		Send().Body().Bytes(statsContent),
		Expect().Status().Equal(http.StatusCreated),
	)
}

// HTTP test kompanion shelf feature
func TestHTTPKompanionOPDS(t *testing.T) {
	username, password := grabTestUser()
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))

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

	// put book
	var redirectedPath string
	Test(t,
		HTTPClient(client),
		Description("Kompanion Put Book"),
		Post(basePath+"/books/upload"),
		Send().Headers("Content-Type").Add(multipartWriter.FormDataContentType()),
		Send().Body().String(requestBody.String()),
		Expect().Status().Equal(http.StatusFound),
		Store().Response().Headers("Location").In(&redirectedPath),
	)
	bookID := strings.Split(redirectedPath, "/")[2]

	// list books via OPDS
	Test(t,
		Description("Kompanion List Books via OPDS"),
		Get(basePath+"/opds"),
		Send().Headers("Authorization").Add(basicAuth),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains("/opds/newest"),
	)
	// list newest books
	Test(t,
		Description("Kompanion Newest Books via OPDS"),
		Get(basePath+"/opds/newest"),
		Send().Headers("Authorization").Add(basicAuth),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().String().Contains("/opds/newest"),
	)
	// download book
	Test(t,
		Description("Kompanion Get Book"),
		Send().Headers("Authorization").Add(basicAuth),
		Get(fmt.Sprintf("%s/opds/book/%s/download", basePath, bookID)),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().Bytes().Equal(bookContent),
	)
	// seach opds
}

func grabTestUser() (string, string) {
	// TODO: read from env
	return "user", "password"
}

func generateDeviceName() string {
	return petname.Generate(2, "-")
}

func hashSyncPassword(password string) string {
	// md5
	return "5f4dcc3b5aa765d61d8327deb882cf99"
}

// webAuthSteps returns a client and a step to authenticate
func webAuthSteps() (*http.Client, hit.IStep) {
	username, password := grabTestUser()

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Jar: jar,
		// do not follow redirect
		// to check Web API against status codes and location of redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	template := CombineSteps(
		HTTPClient(client),
		Description("Auth Correct"),
		Post(basePath+"/auth/login"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("username").Add(username),
		Send().Body().FormValues("password").Add(password),
		Expect().Status().Equal(http.StatusFound),
	)
	return client, template
}

func setupDeviceSteps(client *http.Client, deviceName string) hit.IStep {
	return CombineSteps(
		HTTPClient(client),
		Description("Device Register"),
		Post(basePath+"/devices/add"),
		Send().Headers("Content-Type").Add("application/x-www-form-urlencoded"),
		Send().Body().FormValues("device_name").Add(deviceName),
		Send().Body().FormValues("password").Add("password"),
		Expect().Status().Equal(http.StatusFound),
	)
}
