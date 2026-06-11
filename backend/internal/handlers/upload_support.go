package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const maxUploadImageSize = int64(5 * 1024 * 1024)

var allowedUploadImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var uploadImageExtensionsByType = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var allowedUploadImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

type uploadValidationError struct {
	code    string
	message string
}

func (e *uploadValidationError) Error() string {
	return e.message
}

func uploadJSONError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    code,
		"message": message,
	})
}

func validateUploadImage(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", &uploadValidationError{
			code:    "NO_FILE",
			message: "No file provided",
		}
	}
	contentType := file.Header.Get("Content-Type")
	if !allowedUploadImageTypes[contentType] {
		return "", &uploadValidationError{
			code:    "INVALID_FILE_TYPE",
			message: "Only image files (JPEG, PNG, GIF, WebP) are allowed",
		}
	}
	if file.Size > maxUploadImageSize {
		return "", &uploadValidationError{
			code:    "FILE_TOO_LARGE",
			message: "File size must be less than 5MB",
		}
	}

	openedFile, err := file.Open()
	if err != nil {
		return "", &uploadValidationError{
			code:    "INVALID_FILE",
			message: "Unable to read uploaded file",
		}
	}
	defer openedFile.Close()

	buffer := make([]byte, 512)
	n, err := openedFile.Read(buffer)
	if err != nil && n == 0 {
		return "", &uploadValidationError{
			code:    "INVALID_FILE",
			message: "Unable to read uploaded file",
		}
	}
	detectedType := http.DetectContentType(buffer[:n])
	if !allowedUploadImageTypes[detectedType] {
		return "", &uploadValidationError{
			code:    "INVALID_FILE_TYPE",
			message: "Only image files (JPEG, PNG, GIF, WebP) are allowed",
		}
	}
	return detectedType, nil
}

func buildUploadFilename(contentType string) string {
	extension := uploadImageExtensionsByType[contentType]
	if extension == "" {
		extension = ".bin"
	}
	return fmt.Sprintf("%s%s", uuid.New().String(), extension)
}

func normalizeUploadFilenameParam(filename string) (string, error) {
	if filename == "" || strings.Contains(filename, "\x00") || filename != filepath.Base(filename) {
		return "", fmt.Errorf("invalid filename")
	}
	extension := strings.ToLower(filepath.Ext(filename))
	if !allowedUploadImageExtensions[extension] {
		return "", fmt.Errorf("invalid filename")
	}
	return filename, nil
}
