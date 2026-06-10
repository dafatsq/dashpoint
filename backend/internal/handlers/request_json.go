package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/gofiber/fiber/v2"
)

var (
	errEmptyJSONBody      = errors.New("empty request body")
	errJSONBodyTooLarge   = errors.New("request body is too large")
	errMultipleJSONValues = errors.New("request body must contain a single JSON object")
)

func parseStrictJSONBody(c *fiber.Ctx, dest interface{}, maxBytes int) error {
	body := bytes.TrimSpace(c.Body())
	if len(body) == 0 {
		return errEmptyJSONBody
	}
	if len(body) > maxBytes {
		return errJSONBodyTooLarge
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errMultipleJSONValues
	}
	return nil
}
