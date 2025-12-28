package database

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

const (
	tenantIDKey        = "tenant_id"
	tenantModeKey      = "tenant_mode" // 租户查询模式
	skipTenantCheckKey = "skip_tenant_check"
)

// 租户查询模式常量
const (
	TenantModeAuto   = "auto"   // 自动添加当前租户过滤（默认）
	TenantModeManual = "manual" // 手动控制，不自动添加
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

// WithTenantMode 设置租户查询模式
// mode: "auto" (默认，自动添加tenant_id过滤) 或 "manual" (手动控制WHERE条件)
func WithTenantMode(ctx context.Context, mode string) context.Context {
	return context.WithValue(ctx, tenantModeKey, mode)
}

// ManualTenantMode 设置为手动模式，Repository 完全手动控制 WHERE 条件
// 适用于：查询默认租户数据、跨租户查询等需要精确控制的场景
func ManualTenantMode(ctx context.Context) context.Context {
	return context.WithValue(ctx, tenantModeKey, TenantModeManual)
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

	// 1. 优先检查是否跳过租户检查
	if shouldSkipTenantCheck(db) {
		return
	}

	// 2. 检查查询模式
	mode := getTenantMode(db)
	if mode == TenantModeManual {
		// 手动模式：不自动添加 WHERE，由 Repository 精确控制
		return
	}

	// 3. 默认行为：自动添加当前租户
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
	id, ok := db.Statement.Context.Value(tenantIDKey).(string)
	// 修改：允许空字符串作为有效的 tenant_id（用于默认租户）
	// 只要 ok 为 true，就认为 tenant_id 有效
	return id, ok
}

func shouldSkipTenantCheck(db *gorm.DB) bool {
	if db.Statement.Context == nil {
		return false
	}
	skip, ok := db.Statement.Context.Value(skipTenantCheckKey).(bool)
	return ok && skip
}

// getTenantMode 获取租户查询模式，默认为 auto
func getTenantMode(db *gorm.DB) string {
	if db.Statement.Context == nil {
		return TenantModeAuto
	}
	mode, ok := db.Statement.Context.Value(tenantModeKey).(string)
	if !ok {
		return TenantModeAuto
	}
	return mode
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
