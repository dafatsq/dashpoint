package handlers

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// UploadHandler handles file upload endpoints
type UploadHandler struct {
	uploadDir string
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadDir string) *UploadHandler {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatal().Err(err).Msg("Failed to create upload directory")
	}

	return &UploadHandler{
		uploadDir: uploadDir,
	}
}

// UploadImage handles POST /api/v1/upload/image
func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return uploadJSONError(c, fiber.StatusBadRequest, "NO_FILE", "No file provided")
	}

	detectedType, err := validateUploadImage(file)
	if err != nil {
		var validationErr *uploadValidationError
		if errors.As(err, &validationErr) {
			return uploadJSONError(c, fiber.StatusBadRequest, validationErr.code, validationErr.message)
		}
		return uploadJSONError(c, fiber.StatusBadRequest, "INVALID_FILE_TYPE", err.Error())
	}

	filename := buildUploadFilename(detectedType)
	filePath := filepath.Join(h.uploadDir, filename)

	if err := c.SaveFile(file, filePath); err != nil {
		log.Error().Err(err).Msg("Failed to save file")
		return uploadJSONError(c, fiber.StatusInternalServerError, "SAVE_FAILED", "Failed to save file")
	}

	imageURL := "/uploads/" + filename

	return c.JSON(fiber.Map{
		"url":      imageURL,
		"filename": filename,
	})
}

// DeleteImage handles DELETE /api/v1/upload/image/:filename
func (h *UploadHandler) DeleteImage(c *fiber.Ctx) error {
	filename, err := normalizeUploadFilenameParam(c.Params("filename"))
	if err != nil {
		return uploadJSONError(c, fiber.StatusBadRequest, "INVALID_FILENAME", "Invalid filename")
	}

	filePath := filepath.Join(h.uploadDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return uploadJSONError(c, fiber.StatusNotFound, "FILE_NOT_FOUND", "File not found")
	}

	if err := os.Remove(filePath); err != nil {
		log.Error().Err(err).Msg("Failed to delete file")
		return uploadJSONError(c, fiber.StatusInternalServerError, "DELETE_FAILED", "Failed to delete file")
	}

	return c.JSON(fiber.Map{
		"message": "File deleted successfully",
	})
}
