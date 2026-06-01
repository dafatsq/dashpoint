package models

// AuditAction represents types of auditable actions.
type AuditAction string

const (
	AuditActionLogin       AuditAction = "auth.login"
	AuditActionLoginFailed AuditAction = "auth.login_failed"
	AuditActionLogout      AuditAction = "auth.logout"
	AuditActionPINLogin    AuditAction = "auth.pin_login"

	AuditActionUserCreate       AuditAction = "user.create"
	AuditActionUserUpdate       AuditAction = "user.update"
	AuditActionUserDelete       AuditAction = "user.delete"
	AuditActionUserArchive      AuditAction = "user.archive"
	AuditActionUserRestore      AuditAction = "user.restore"
	AuditActionPasswordChange   AuditAction = "user.password_change"
	AuditActionPINChange        AuditAction = "user.pin_change"
	AuditActionPermissionChange AuditAction = "user.permission_change"

	AuditActionProductCreate  AuditAction = "product.create"
	AuditActionProductUpdate  AuditAction = "product.update"
	AuditActionProductDelete  AuditAction = "product.delete"
	AuditActionProductArchive AuditAction = "product.archive"
	AuditActionProductRestore AuditAction = "product.restore"

	AuditActionStockAdjust     AuditAction = "inventory.adjust"
	AuditActionStockCount      AuditAction = "inventory.count"
	AuditActionThresholdUpdate AuditAction = "inventory.threshold_update"

	AuditActionCategoryCreate  AuditAction = "category.create"
	AuditActionCategoryUpdate  AuditAction = "category.update"
	AuditActionCategoryDelete  AuditAction = "category.delete"
	AuditActionCategoryArchive AuditAction = "category.archive"
	AuditActionCategoryRestore AuditAction = "category.restore"

	AuditActionSaleCreate AuditAction = "sale.create"
	AuditActionSaleVoid   AuditAction = "sale.void"

	AuditActionShiftStart AuditAction = "shift.start"
	AuditActionShiftClose AuditAction = "shift.close"
	AuditActionCashPayIn  AuditAction = "shift.pay_in"
	AuditActionCashPayOut AuditAction = "shift.pay_out"

	AuditActionExpenseCreate  AuditAction = "expense.create"
	AuditActionExpenseUpdate  AuditAction = "expense.update"
	AuditActionExpenseDelete  AuditAction = "expense.delete"
	AuditActionExpenseArchive AuditAction = "expense.archive"
	AuditActionExpenseRestore AuditAction = "expense.restore"

	AuditActionReportExport  AuditAction = "report.export"
	AuditActionSettingChange AuditAction = "system.setting_change"
)

// AuditEntityType represents the type of entity being audited.
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

// AuditStatus represents the status of an audited action.
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailure AuditStatus = "failure"
	AuditStatusWarning AuditStatus = "warning"
)
