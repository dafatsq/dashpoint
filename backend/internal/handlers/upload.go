package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// UploadHandler handles file upload endpoints
type UploadHandler struct {
	uploadDir string
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadDir string) *UploadHandler {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("Failed to create upload directory")
	}

	return &UploadHandler{
		uploadDir: uploadDir,
	}
}

// UploadImage handles POST /api/v1/upload/image
func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	// Get the file from the request
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "NO_FILE",
			"message": "No file provided",
		})
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !allowedTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_FILE_TYPE",
			"message": "Only image files (JPEG, PNG, GIF, WebP) are allowed",
		})
	}

	// Validate file size (max 5MB)
	maxSize := int64(5 * 1024 * 1024)
	if file.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "FILE_TOO_LARGE",
			"message": "File size must be less than 5MB",
		})
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(h.uploadDir, filename)

	// Save the file
	if err := c.SaveFile(file, filePath); err != nil {
		log.Error().Err(err).Msg("Failed to save file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "SAVE_FAILED",
			"message": "Failed to save file",
		})
	}

	// Return the URL
	imageURL := fmt.Sprintf("/uploads/%s", filename)

	return c.JSON(fiber.Map{
		"url":      imageURL,
		"filename": filename,
	})
}

// DeleteImage handles DELETE /api/v1/upload/image/:filename
func (h *UploadHandler) DeleteImage(c *fiber.Ctx) error {
	filename := c.Params("filename")

	// Sanitize filename to prevent directory traversal
	filename = filepath.Base(filename)

	filePath := filepath.Join(h.uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "FILE_NOT_FOUND",
			"message": "File not found",
		})
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		log.Error().Err(err).Msg("Failed to delete file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "DELETE_FAILED",
			"message": "Failed to delete file",
		})
	}

	return c.JSON(fiber.Map{
		"message": "File deleted successfully",
	})
}
