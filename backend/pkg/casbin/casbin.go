package casbin

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type Enforcer struct {
	*casbin.Enforcer
	adapter *gormadapter.Adapter
}

func NewEnforcerManager(db *gorm.DB, modelStr string) (*Enforcer, error) {
	// 创建gorm适配器
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// 从字符串创建模型
	m, err := model.NewModelFromString(modelStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create model from string: %w", err)
	}

	// 创建执行器
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 自动加载策略
	enforcer.EnableAutoSave(true)

	// 设置角色管理器（支持角色继承，限制层级）
	enforcer.EnableAutoBuildRoleLinks(true)

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	return &Enforcer{
		Enforcer: enforcer,
		adapter:  adapter,
	}, nil

}

// 按租户加载策略（性能优化）
func (e *Enforcer) LoadPolicyForTenant(tenantID string) error {
	filter := &gormadapter.Filter{
		V1: []string{tenantID}, // v1对应domain字段[citation:4]
	}
	return e.LoadFilteredPolicy(filter)
}

// 添加租户策略
func (e *Enforcer) AddPolicyForTenant(tenantID, sub, obj, act string) (bool, error) {
	return e.AddPolicy(sub, tenantID, obj, act)
}

// 检查租户权限
func (e *Enforcer) EnforceForTenant(tenantID, sub, obj, act string) (bool, error) {
	return e.Enforce(sub, tenantID, obj, act)
}
