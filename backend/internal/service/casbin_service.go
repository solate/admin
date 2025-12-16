package service

import (
	"admin/pkg/casbin"

	"gorm.io/gorm"
)

type CasbinService struct {
	Enforcer *casbin.Enforcer
}

func InitCasbinService(db *gorm.DB, modelStr string) (*CasbinService, error) {
	enf, err := casbin.NewEnforcerManager(db, modelStr)
	if err != nil {
		return nil, err
	}
	if err := enf.LoadPolicy(); err != nil {
		return nil, err
	}
	return &CasbinService{
		Enforcer: enf,
	}, nil
}

func (s *CasbinService) AddPolicy(sub, dom, obj, act string) (bool, error) {
	return s.Enforcer.AddPolicy(sub, dom, obj, act)
}

func (s *CasbinService) RemovePolicy(sub, dom, obj, act string) (bool, error) {
	return s.Enforcer.RemovePolicy(sub, dom, obj, act)
}

func (s *CasbinService) AddPolicyForTenant(tenantID, sub, obj, act string) (bool, error) {
	return s.Enforcer.AddPolicy(sub, tenantID, obj, act)
}

func (s *CasbinService) AddRoleForUserInTenant(userID, roleID, tenantID string) (bool, error) {
	return s.Enforcer.AddGroupingPolicy(userID, roleID, tenantID)
}

func (s *CasbinService) RemoveRoleForUserInTenant(userID, roleID, tenantID string) (bool, error) {
	return s.Enforcer.RemoveGroupingPolicy(userID, roleID, tenantID)
}

func (s *CasbinService) Enforce(userID, tenantID, obj, act string) (bool, error) {
	return s.Enforcer.Enforce(userID, tenantID, obj, act)
}
