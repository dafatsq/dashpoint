package handlers

import (
	"fmt"
	"mime/multipart"
	"path/filepath"

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

func validateUploadImage(file *multipart.FileHeader) error {
	if file == nil {
		return &uploadValidationError{
			code:    "NO_FILE",
			message: "No file provided",
		}
	}
	contentType := file.Header.Get("Content-Type")
	if !allowedUploadImageTypes[contentType] {
		return &uploadValidationError{
			code:    "INVALID_FILE_TYPE",
			message: "Only image files (JPEG, PNG, GIF, WebP) are allowed",
		}
	}
	if file.Size > maxUploadImageSize {
		return &uploadValidationError{
			code:    "FILE_TOO_LARGE",
			message: "File size must be less than 5MB",
		}
	}
	return nil
}

func buildUploadFilename(originalName string) string {
	return fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(originalName))
}
