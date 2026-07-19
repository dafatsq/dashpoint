package repository

import (
	"context"
	"time"
)

// GetActionSummary gets a summary of actions in a time period.
func (r *AuditRepository) GetActionSummary(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	query := `
		SELECT
			action,
			COUNT(*) as count,
			COUNT(DISTINCT user_id) as unique_users
		FROM audit_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY action
		ORDER BY count DESC
	`

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var action string
		var count, uniqueUsers int
		if err := rows.Scan(&action, &count, &uniqueUsers); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"action":       action,
			"count":        count,
			"unique_users": uniqueUsers,
		})
	}

	return results, nil
}

// Cleanup deletes audit logs older than specified days.
func (r *AuditRepository) Cleanup(ctx context.Context, olderThanDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)
	result, err := r.pool.Exec(ctx, `DELETE FROM audit_logs WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
