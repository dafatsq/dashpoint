package handlers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
)

const reportDateLayout = "2006-01-02"

var reportBusinessLocation = loadReportBusinessLocation()

type reportDateRange struct {
	start    time.Time
	end      time.Time
	startStr string
	endStr   string
}

func reportError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    code,
		"message": message,
	})
}

func parseReportDay(value, field string) (time.Time, error) {
	parsed, err := time.ParseInLocation(reportDateLayout, value, reportBusinessLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s", field)
	}
	return parsed, nil
}

func loadReportBusinessLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return location
}

func reportDayStart(date time.Time) time.Time {
	localDate := date.In(reportBusinessLocation)
	return time.Date(localDate.Year(), localDate.Month(), localDate.Day(), 0, 0, 0, 0, reportBusinessLocation)
}

func parseReportRange(c *fiber.Ctx, defaultDays int, requireProvided bool, maxDays int) (*reportDateRange, error) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		if requireProvided {
			return nil, errors.New("missing")
		}
		endDate = reportDayStart(time.Now().In(reportBusinessLocation))
		startDate = endDate.AddDate(0, 0, -defaultDays)
		startStr = startDate.Format(reportDateLayout)
		endStr = endDate.Format(reportDateLayout)
	} else {
		startDate, err = parseReportDay(startStr, "start_date")
		if err != nil {
			return nil, err
		}
		endDate, err = parseReportDay(endStr, "end_date")
		if err != nil {
			return nil, err
		}
	}

	if endDate.Before(startDate) {
		return nil, errors.New("range")
	}

	if maxDays > 0 && endDate.Sub(startDate) > time.Duration(maxDays)*24*time.Hour {
		return nil, errors.New("too_large")
	}

	return &reportDateRange{
		start:    startDate,
		end:      endDate,
		startStr: startStr,
		endStr:   endStr,
	}, nil
}

func parseReportRangeResponse(c *fiber.Ctx, defaultDays int, requireProvided bool, maxDays int) (*reportDateRange, error) {
	dateRange, err := parseReportRange(c, defaultDays, requireProvided, maxDays)
	if err == nil {
		return dateRange, nil
	}

	switch err.Error() {
	case "missing":
		return nil, reportError(c, fiber.StatusBadRequest, "INVALID_DATE", "start_date and end_date are required")
	case "start_date":
		return nil, reportError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid start_date format. Use YYYY-MM-DD")
	case "end_date":
		return nil, reportError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid end_date format. Use YYYY-MM-DD")
	case "range":
		return nil, reportError(c, fiber.StatusBadRequest, "INVALID_RANGE", "end_date must be after start_date")
	case "too_large":
		return nil, reportError(c, fiber.StatusBadRequest, "RANGE_TOO_LARGE", "Date range cannot exceed 90 days")
	default:
		return nil, reportError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid date format. Use YYYY-MM-DD")
	}
}

func parseReportCategoryID(c *fiber.Ctx) (*uuid.UUID, error) {
	categoryIDStr := c.Query("category_id")
	if categoryIDStr == "" {
		return nil, nil
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return nil, reportError(c, fiber.StatusBadRequest, "INVALID_CATEGORY_ID", "Invalid category_id format")
	}

	return &categoryID, nil
}

func parseBoundedIntQuery(c *fiber.Ctx, key string, defaultValue, maxValue int) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 || (maxValue > 0 && value > maxValue) {
		return 0, fmt.Errorf("%s", key)
	}

	return value, nil
}

func parseNonNegativeIntQuery(c *fiber.Ctx, key string, defaultValue, maxValue int) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || (maxValue > 0 && value > maxValue) {
		return 0, fmt.Errorf("%s", key)
	}

	return value, nil
}

func parseReportLimitQuery(c *fiber.Ctx, defaultValue, maxValue int) (int, error) {
	limit, err := parseBoundedIntQuery(c, "limit", defaultValue, maxValue)
	if err != nil {
		return 0, reportError(c, fiber.StatusBadRequest, "INVALID_LIMIT", "Invalid limit")
	}
	return limit, nil
}

func reportInternalError(c *fiber.Ctx, err error, message string) error {
	log.Error().Err(err).Msg(message)
	return reportError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate report")
}

func reportExportInternalError(c *fiber.Ctx, err error, message string) error {
	log.Error().Err(err).Msg(message)
	return reportError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to export data")
}

func writeCSV(rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = ';'
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sendCSV(c *fiber.Ctx, filename string, data []byte) error {
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(data)
}

func csvText(value string) string {
	sanitized := strings.NewReplacer("\r", " ", "\n", " ", "\t", " ", "\x00", "").Replace(value)
	trimmed := strings.TrimLeftFunc(sanitized, unicode.IsSpace)
	if trimmed == "" {
		return sanitized
	}

	switch trimmed[0] {
	case '=', '+', '-', '@':
		return "'" + sanitized
	default:
		return sanitized
	}
}

func logReportExport(c *fiber.Ctx, exportType, filename string, dateRange *reportDateRange, extra map[string]interface{}) {
	values := map[string]interface{}{
		"export_type": exportType,
		"filename":    filename,
	}
	if dateRange != nil {
		values["start_date"] = dateRange.startStr
		values["end_date"] = dateRange.endStr
	}
	for key, value := range extra {
		values[key] = value
	}

	audit.LogWithValues(
		c,
		models.AuditActionReportExport,
		models.AuditEntityReport,
		exportType,
		"Exported report: "+exportType,
		nil,
		values,
	)
}
