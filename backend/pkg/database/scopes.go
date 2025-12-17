package database

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

const (
	tenantIDKey         = "tenant_id"
	skipTenantCheckKey  = "skip_tenant_check"
)

var (
	ErrMissingTenantID = errors.New("missing tenant_id in context")
)

func SkipTenantCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipTenantCheckKey, true)
}

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

func RegisterCallbacks(db *gorm.DB) error {
	callbacks := db.Callback()
	if err := callbacks.Create().Before("gorm:create").Register("tenant:create", tenantCreateCallback); err != nil {
		return err
	}
	if err := callbacks.Query().Before("gorm:query").Register("tenant:query", tenantQueryCallback); err != nil {
		return err
	}
	if err := callbacks.Update().Before("gorm:update").Register("tenant:update", tenantQueryCallback); err != nil {
		return err
	}
	if err := callbacks.Delete().Before("gorm:delete").Register("tenant:delete", tenantQueryCallback); err != nil {
		return err
	}
	return nil
}

func tenantCreateCallback(db *gorm.DB) {
	if shouldSkipTenantCheck(db) {
		log.Warn().Str("table", db.Statement.Table).Msg("tenant check skipped for create")
		return
	}
	tenantID, ok := getTenantID(db)
	if !ok {
		db.AddError(ErrMissingTenantID)
		return
	}
	if !hasTenantColumn(db) {
		return
	}
	setTenantID(db, tenantID)
}

func tenantQueryCallback(db *gorm.DB) {
	if shouldSkipTenantCheck(db) {
		log.Warn().Str("table", db.Statement.Table).Msg("tenant check skipped for query/update/delete")
		return
	}
	tenantID, ok := getTenantID(db)
	if !ok {
		db.AddError(ErrMissingTenantID)
		return
	}
	if !hasTenantColumn(db) || hasTenantCondition(db) {
		return
	}
	db.Statement.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Eq{Column: clause.Column{Name: "tenant_id"}, Value: tenantID},
	}})
}

func getTenantID(db *gorm.DB) (string, bool) {
	if db.Statement.Context == nil {
		return "", false
	}
	id, ok := db.Statement.Context.Value(tenantIDKey).(string)
	return id, ok && id != ""
}

func shouldSkipTenantCheck(db *gorm.DB) bool {
	if db.Statement.Context == nil {
		return false
	}
	skip, ok := db.Statement.Context.Value(skipTenantCheckKey).(bool)
	return ok && skip
}

func hasTenantColumn(db *gorm.DB) bool {
	if db.Statement.Schema == nil {
		return false
	}
	for _, field := range db.Statement.Schema.Fields {
		if strings.EqualFold(field.DBName, "tenant_id") {
			return true
		}
	}
	return false
}

func hasTenantCondition(db *gorm.DB) bool {
	whereClause, ok := db.Statement.Clauses["WHERE"]
	if !ok {
		return false
	}
	where, ok := whereClause.Expression.(clause.Where)
	if !ok {
		return false
	}
	for _, expr := range where.Exprs {
		if eq, ok := expr.(clause.Eq); ok {
			if col, isCol := eq.Column.(clause.Column); isCol && strings.EqualFold(col.Name, "tenant_id") {
				return true
			}
			if str, isStr := eq.Column.(string); isStr && strings.EqualFold(str, "tenant_id") {
				return true
			}
		}
	}
	return false
}

func setTenantID(db *gorm.DB, tenantID string) {
	if db.Statement.Schema == nil {
		return
	}
	var target *schema.Field
	for _, f := range db.Statement.Schema.Fields {
		if strings.EqualFold(f.DBName, "tenant_id") {
			target = f
			break
		}
	}
	if target == nil {
		return
	}
	rv := db.Statement.ReflectValue
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return
	}
	_ = target.Set(db.Statement.Context, rv, tenantID)
}
