package audit

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// Service provides audit logging functionality
type Service struct {
	repo *repository.AuditRepository
}

// NewService creates a new audit service
func NewService(repo *repository.AuditRepository) *Service {
	return &Service{repo: repo}
}

// Global instance for easy access
var globalService *Service

// Init initializes the global audit service
func Init(repo *repository.AuditRepository) {
	globalService = NewService(repo)
}

// Log creates an audit log entry using the global service
func Log(ctx context.Context, entry *models.AuditLogEntry) {
	if globalService == nil {
		log.Warn().Msg("Audit service not initialized")
		return
	}

	go func() {
		if err := globalService.repo.Create(context.Background(), entry); err != nil {
			log.Error().Err(err).
				Str("action", string(entry.Action)).
				Str("entity_type", string(entry.EntityType)).
				Msg("Failed to create audit log")
		}
	}()
}

// LogSync creates an audit log entry synchronously
func LogSync(ctx context.Context, entry *models.AuditLogEntry) error {
	if globalService == nil {
		log.Warn().Msg("Audit service not initialized")
		return nil
	}
	return globalService.repo.Create(ctx, entry)
}

// LogFromFiber creates an audit log entry with context from Fiber
func LogFromFiber(c *fiber.Ctx, action models.AuditAction, entityType models.AuditEntityType, entityID string, description string) {
	entry := &models.AuditLogEntry{
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		Status:      models.AuditStatusSuccess,
	}

	// Extract user info from context
	if claims := middleware.GetClaims(c); claims != nil {
		entry.UserID = &claims.UserID
		entry.UserEmail = claims.Email
		entry.UserName = claims.Name
		entry.UserRole = claims.RoleName
	}

	// Extract request info
	entry.IPAddress = c.IP()
	entry.UserAgent = c.Get("User-Agent")
	if requestID := c.Locals("requestid"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			entry.RequestID = rid
		}
	}

	Log(c.Context(), entry)
}

// LogWithValues creates an audit log entry with old/new values
func LogWithValues(c *fiber.Ctx, action models.AuditAction, entityType models.AuditEntityType, entityID string, description string, oldValues, newValues map[string]interface{}) {
	entry := &models.AuditLogEntry{
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		OldValues:   oldValues,
		NewValues:   newValues,
		Status:      models.AuditStatusSuccess,
	}

	// Extract user info from context
	if claims := middleware.GetClaims(c); claims != nil {
		entry.UserID = &claims.UserID
		entry.UserEmail = claims.Email
		entry.UserName = claims.Name
		entry.UserRole = claims.RoleName
	}

	// Extract request info
	entry.IPAddress = c.IP()
	entry.UserAgent = c.Get("User-Agent")
	if requestID := c.Locals("requestid"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			entry.RequestID = rid
		}
	}

	Log(c.Context(), entry)
}

// LogFailure creates a failure audit log entry
func LogFailure(c *fiber.Ctx, action models.AuditAction, entityType models.AuditEntityType, entityID string, description string, metadata map[string]interface{}) {
	entry := &models.AuditLogEntry{
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		Metadata:    metadata,
		Status:      models.AuditStatusFailure,
	}

	// Extract user info from context
	if claims := middleware.GetClaims(c); claims != nil {
		entry.UserID = &claims.UserID
		entry.UserEmail = claims.Email
		entry.UserRole = claims.RoleName
	}

	// Extract request info
	entry.IPAddress = c.IP()
	entry.UserAgent = c.Get("User-Agent")
	if requestID := c.Locals("requestid"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			entry.RequestID = rid
		}
	}

	Log(c.Context(), entry)
}

// LogAuth creates an auth-related audit log
func LogAuth(c *fiber.Ctx, action models.AuditAction, userID *uuid.UUID, email string, success bool, metadata map[string]interface{}) {
	status := models.AuditStatusSuccess
	if !success {
		status = models.AuditStatusFailure
	}

	entityID := ""
	if userID != nil {
		entityID = userID.String()
	}

	entry := &models.AuditLogEntry{
		UserID:     userID,
		UserEmail:  email,
		Action:     action,
		EntityType: models.AuditEntityAuth,
		EntityID:   entityID,
		Metadata:   metadata,
		Status:     status,
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
	}

	if requestID := c.Locals("requestid"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			entry.RequestID = rid
		}
	}

	Log(c.Context(), entry)
}

// GetRepo returns the audit repository for direct access
func GetRepo() *repository.AuditRepository {
	if globalService == nil {
		return nil
	}
	return globalService.repo
}
