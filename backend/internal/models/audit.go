package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditAction represents types of auditable actions
type AuditAction string

const (
	// Auth actions
	AuditActionLogin       AuditAction = "auth.login"
	AuditActionLoginFailed AuditAction = "auth.login_failed"
	AuditActionLogout      AuditAction = "auth.logout"
	AuditActionPINLogin    AuditAction = "auth.pin_login"

	// User actions
	AuditActionUserCreate       AuditAction = "user.create"
	AuditActionUserUpdate       AuditAction = "user.update"
	AuditActionUserDelete       AuditAction = "user.delete"
	AuditActionUserArchive      AuditAction = "user.archive"
	AuditActionUserRestore      AuditAction = "user.restore"
	AuditActionPasswordChange   AuditAction = "user.password_change"
	AuditActionPINChange        AuditAction = "user.pin_change"
	AuditActionPermissionChange AuditAction = "user.permission_change"

	// Product actions
	AuditActionProductCreate  AuditAction = "product.create"
	AuditActionProductUpdate  AuditAction = "product.update"
	AuditActionProductDelete  AuditAction = "product.delete"
	AuditActionProductArchive AuditAction = "product.archive"
	AuditActionProductRestore AuditAction = "product.restore"

	// Inventory actions
	AuditActionStockAdjust AuditAction = "inventory.adjust"
	AuditActionStockCount  AuditAction = "inventory.count"

	// Category actions
	AuditActionCategoryCreate AuditAction = "category.create"
	AuditActionCategoryUpdate  AuditAction = "category.update"
	AuditActionCategoryDelete  AuditAction = "category.delete"
	AuditActionCategoryArchive AuditAction = "category.archive"
	AuditActionCategoryRestore AuditAction = "category.restore"

	// Sale actions
	AuditActionSaleCreate AuditAction = "sale.create"
	AuditActionSaleVoid   AuditAction = "sale.void"
	AuditActionRefund     AuditAction = "sale.refund"

	// Shift actions
	AuditActionShiftStart AuditAction = "shift.start"
	AuditActionShiftClose AuditAction = "shift.close"
	AuditActionCashPayIn  AuditAction = "shift.pay_in"
	AuditActionCashPayOut AuditAction = "shift.pay_out"

	// Expense actions
	AuditActionExpenseCreate AuditAction = "expense.create"
	AuditActionExpenseUpdate AuditAction = "expense.update"
	AuditActionExpenseDelete AuditAction = "expense.delete"
	AuditActionExpenseArchive AuditAction = "expense.archive"
	AuditActionExpenseRestore AuditAction = "expense.restore"

	// Report actions
	AuditActionReportExport AuditAction = "report.export"

	// System actions
	AuditActionSettingChange AuditAction = "system.setting_change"
)

// AuditEntityType represents the type of entity being audited
type AuditEntityType string

const (
	AuditEntityUser      AuditEntityType = "user"
	AuditEntityProduct   AuditEntityType = "product"
	AuditEntityCategory  AuditEntityType = "category"
	AuditEntityInventory AuditEntityType = "inventory"
	AuditEntitySale      AuditEntityType = "sale"
	AuditEntityShift     AuditEntityType = "shift"
	AuditEntityExpense   AuditEntityType = "expense"
	AuditEntityPayment   AuditEntityType = "payment"
	AuditEntityReport    AuditEntityType = "report"
	AuditEntitySystem    AuditEntityType = "system"
	AuditEntityAuth      AuditEntityType = "auth"
)

// AuditStatus represents the status of an audited action
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailure AuditStatus = "failure"
	AuditStatusWarning AuditStatus = "warning"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// Who
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	UserEmail *string    `json:"user_email,omitempty"`
	UserName  *string    `json:"user_name,omitempty"`
	UserRole  *string    `json:"user_role,omitempty"`

	// What
	Action      AuditAction     `json:"action"`
	EntityType  AuditEntityType `json:"entity_type"`
	EntityID    *string         `json:"entity_id,omitempty"`
	Description *string         `json:"description,omitempty"`

	// Changes
	OldValues map[string]interface{} `json:"old_values,omitempty"`
	NewValues map[string]interface{} `json:"new_values,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Where
	IPAddress *string `json:"ip_address,omitempty"`
	UserAgent *string `json:"user_agent,omitempty"`
	RequestID *string `json:"request_id,omitempty"`

	// Status
	Status AuditStatus `json:"status"`
}

// AuditLogEntry is a builder for creating audit log entries
type AuditLogEntry struct {
	UserID      *uuid.UUID
	UserEmail   string
	UserName    string
	UserRole    string
	Action      AuditAction
	EntityType  AuditEntityType
	EntityID    string
	Description string
	OldValues   map[string]interface{}
	NewValues   map[string]interface{}
	Metadata    map[string]interface{}
	IPAddress   string
	UserAgent   string
	RequestID   string
	Status      AuditStatus
}
