package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// deleteImageFile removes an uploaded image file from disk.
// imageURL is expected to be in the form "/uploads/<filename>".
func (h *ProductHandler) deleteImageFile(imageURL string) {
	if imageURL == "" || h.uploadDir == "" {
		return
	}
	filename := strings.TrimPrefix(imageURL, "/uploads/")
	filename = filepath.Base(filename)
	if filename == "." || filename == "" {
		return
	}
	filePath := filepath.Join(h.uploadDir, filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		log.Warn().Err(err).Str("file", filePath).Msg("Failed to delete orphaned image file")
	}
}
