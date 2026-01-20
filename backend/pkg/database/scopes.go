package database

import (
	"errors"
	"reflect"
	"strings"

	"admin/pkg/xcontext"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var (
	ErrMissingTenantID = errors.New("missing tenant_id in context")
)

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

	// 如果有 tenant_id 列，直接返回
	if hasTenantColumn(db) { // 通过表中有没有租户ID 来判断是否需要

		// 如果跳过租户检查，直接返回
		if shouldSkipTenantCheck(db) {
			return
		}

		// 获取tenantID 并设置到DB中
		tenantID, ok := getTenantID(db)
		if !ok {
			db.AddError(ErrMissingTenantID)
			return
		}
		setTenantID(db, tenantID)
	}
}

func tenantQueryCallback(db *gorm.DB) {
	if !hasTenantColumn(db) {
		return
	}

	// 检查是否跳过租户检查
	if shouldSkipTenantCheck(db) {
		return
	}

	// 默认行为：自动添加当前租户过滤
	tenantID, ok := getTenantID(db)
	if !ok {
		db.AddError(ErrMissingTenantID)
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
	id := xcontext.GetTenantID(db.Statement.Context)
	// 修改：允许空字符串作为有效的 tenant_id（用于默认租户）
	// 只要返回的值不为空（或者即使为空，但我们检查了context存在），就认为 tenant_id 有效
	return id, true
}

func shouldSkipTenantCheck(db *gorm.DB) bool {
	if db.Statement.Context == nil {
		return false
	}
	return xcontext.ShouldSkipTenantCheck(db.Statement.Context)
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
