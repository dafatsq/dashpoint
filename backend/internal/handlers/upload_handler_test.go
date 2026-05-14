package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestUploadHandlerRejectsMissingFile(t *testing.T) {
	handler := NewUploadHandler(t.TempDir())
	app := fiber.New(fiber.Config{BodyLimit: int(maxUploadImageSize) + 1024*1024})
	app.Post("/upload", handler.UploadImage)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUploadHandlerRejectsInvalidContentType(t *testing.T) {
	handler := NewUploadHandler(t.TempDir())
	app := fiber.New()
	app.Post("/upload", handler.UploadImage)

	req := newMultipartUploadRequest(t, "/upload", "image", "notes.txt", "text/plain", []byte("hello"))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "INVALID_FILE_TYPE" {
		t.Fatalf("expected INVALID_FILE_TYPE, got %q", body.Code)
	}
}

func TestUploadHandlerSavesImageAndReturnsURL(t *testing.T) {
	dir := t.TempDir()
	handler := NewUploadHandler(dir)
	app := fiber.New()
	app.Post("/upload", handler.UploadImage)

	req := newMultipartUploadRequest(t, "/upload", "image", "photo.png", "image/png", []byte("png-data"))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.URL == "" || body.Filename == "" {
		t.Fatalf("expected upload url and filename, got %+v", body)
	}
	if _, err := os.Stat(filepath.Join(dir, body.Filename)); err != nil {
		t.Fatalf("expected uploaded file to exist: %v", err)
	}
}

func TestUploadHandlerDeleteReturnsNotFoundForMissingFile(t *testing.T) {
	handler := NewUploadHandler(t.TempDir())
	app := fiber.New()
	app.Delete("/upload/:filename", handler.DeleteImage)

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/upload/missing.png", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUploadHandlerDeleteRemovesExistingFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "photo.png")
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	handler := NewUploadHandler(dir)
	app := fiber.New()
	app.Delete("/upload/:filename", handler.DeleteImage)

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/upload/photo.png", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted, stat err: %v", err)
	}
}

func TestValidateUploadImageRejectsOversizedFiles(t *testing.T) {
	file := &multipart.FileHeader{
		Size: maxUploadImageSize + 1,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{"image/png"},
		},
	}

	if err := validateUploadImage(file); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateUploadImagePreservesInvalidFileTypeCodeForOversizedNonImageFiles(t *testing.T) {
	file := &multipart.FileHeader{
		Size: maxUploadImageSize + 1,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{"text/plain"},
		},
	}

	err := validateUploadImage(file)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*uploadValidationError)
	if !ok {
		t.Fatalf("expected uploadValidationError, got %T", err)
	}
	if validationErr.code != "INVALID_FILE_TYPE" {
		t.Fatalf("expected INVALID_FILE_TYPE, got %q", validationErr.code)
	}
}

func newMultipartUploadRequest(t *testing.T, target, fieldName, filename, contentType string, contents []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="`+fieldName+`"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreateFormFile returned error: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(contents)); err != nil {
		t.Fatalf("copy returned error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-File-Content-Type", contentType)
	return req
}
