package database

import (
	"context"

	"gorm.io/gorm"
)

// TenantScope returns a GORM scope that filters by tenant_id from context
// Usage: db.Scopes(database.TenantScope(ctx)).Find(&users)
func TenantScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Note: The key must match what is used in Auth middleware
		// Since we can't import internal/constants here, we use the string literal "tenant_id"
		// Ensure this matches constants.CtxTenantID
		tenantID, ok := ctx.Value("tenant_id").(string)
		if ok && tenantID != "" {
			// Check if the table has tenant_id column?
			// Usually we assume the caller knows to use this scope on tenant-aware tables.
			return db.Where("tenant_id = ?", tenantID)
		}
		return db
	}
}
