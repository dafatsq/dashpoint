package repository

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/models"
)

type auditScanner interface {
	Scan(dest ...interface{}) error
}

type auditRows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

func buildAuditWhereClause(filter AuditFilter) (string, []interface{}, int) {
	args := []interface{}{}
	argIndex := 1
	whereClause := "WHERE 1=1"

	if filter.UserID != nil {
		whereClause += fmt.Sprintf(" AND audit_logs.user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}
	if filter.Action != nil {
		whereClause += fmt.Sprintf(" AND audit_logs.action = $%d", argIndex)
		args = append(args, *filter.Action)
		argIndex++
	}
	if filter.EntityType != nil {
		whereClause += fmt.Sprintf(" AND audit_logs.entity_type = $%d", argIndex)
		args = append(args, *filter.EntityType)
		argIndex++
	}
	if filter.EntityID != nil {
		whereClause += fmt.Sprintf(" AND audit_logs.entity_id = $%d", argIndex)
		args = append(args, *filter.EntityID)
		argIndex++
	}
	if filter.Status != nil {
		whereClause += fmt.Sprintf(" AND audit_logs.status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}
	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND audit_logs.created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}
	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND audit_logs.created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}
	if filter.Search != nil && *filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (audit_logs.description ILIKE $%d OR audit_logs.user_email ILIKE $%d OR audit_logs.user_name ILIKE $%d OR u.name ILIKE $%d)", argIndex, argIndex, argIndex, argIndex)
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	return whereClause, args, argIndex
}

func collectAuditLogs(rows auditRows) ([]models.AuditLog, error) {
	logs := make([]models.AuditLog, 0)
	for rows.Next() {
		logEntry, err := scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *logEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func scanAuditLog(scanner auditScanner) (*models.AuditLog, error) {
	var logEntry models.AuditLog
	var oldValuesJSON, newValuesJSON, metadataJSON []byte

	if err := scanner.Scan(
		&logEntry.ID, &logEntry.CreatedAt, &logEntry.UserID, &logEntry.UserEmail, &logEntry.UserName, &logEntry.UserRole,
		&logEntry.Action, &logEntry.EntityType, &logEntry.EntityID, &logEntry.Description,
		&oldValuesJSON, &newValuesJSON, &metadataJSON,
		&logEntry.IPAddress, &logEntry.UserAgent, &logEntry.RequestID, &logEntry.Status,
	); err != nil {
		return nil, err
	}

	if err := decodeAuditJSONFields(&logEntry, oldValuesJSON, newValuesJSON, metadataJSON); err != nil {
		return nil, err
	}

	return &logEntry, nil
}

func decodeAuditJSONFields(logEntry *models.AuditLog, oldValuesJSON, newValuesJSON, metadataJSON []byte) error {
	if oldValuesJSON != nil {
		if err := json.Unmarshal(oldValuesJSON, &logEntry.OldValues); err != nil {
			log.Warn().Err(err).Msg("Ignoring invalid audit old_values JSON")
		}
	}
	if newValuesJSON != nil {
		if err := json.Unmarshal(newValuesJSON, &logEntry.NewValues); err != nil {
			log.Warn().Err(err).Msg("Ignoring invalid audit new_values JSON")
		}
	}
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &logEntry.Metadata); err != nil {
			log.Warn().Err(err).Msg("Ignoring invalid audit metadata JSON")
		}
	}
	return nil
}
