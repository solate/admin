package service

import (
	"admin/pkg/casbin"
)

type CasbinService struct {
	enforcer *casbin.Enforcer
}

func NewCasbinService(enforcer *casbin.Enforcer) *CasbinService {
	return &CasbinService{
		enforcer: enforcer,
	}
}

// AddPolicyForTenant 为租户添加策略
func (s *CasbinService) AddPolicyForTenant(tenantID, sub, obj, act string) (bool, error) {
	return s.enforcer.AddPolicyForTenant(tenantID, sub, obj, act)
}

// AddRoleForUserInTenant 为用户在租户中添加角色
func (s *CasbinService) AddRoleForUserInTenant(user, role, tenantID string) (bool, error) {
	return s.enforcer.AddRoleForUserInDomain(user, role, tenantID)
}

// RemovePolicyForTenant 为租户移除策略
func (s *CasbinService) RemovePolicyForTenant(tenantID, sub, obj, act string) (bool, error) {
	return s.enforcer.RemovePolicy(sub, tenantID, obj, act)
}

// GetPoliciesForTenant 获取租户的策略列表
func (s *CasbinService) GetPoliciesForTenant(tenantID string) ([][]string, error) {
	return s.enforcer.GetFilteredPolicy(1, tenantID)
}

// HasPolicyForTenant 检查租户是否有特定策略
func (s *CasbinService) HasPolicyForTenant(tenantID, sub, obj, act string) (bool, error) {
	return s.enforcer.HasPolicy(sub, tenantID, obj, act)
}
