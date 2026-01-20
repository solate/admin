package converter

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// ModelToTenantInfo 将数据库模型转换为租户信息 DTO
func ModelToTenantInfo(tenant *model.Tenant) *dto.TenantInfo {
	if tenant == nil {
		return nil
	}

	return &dto.TenantInfo{
		TenantID:     tenant.TenantID,
		TenantCode:   tenant.TenantCode,
		Name:         tenant.Name,
		Description:  tenant.Description,
		ContactName:  tenant.ContactName,
		ContactPhone: tenant.ContactPhone,
		Status:       int(tenant.Status),
		CreatedAt:    tenant.CreatedAt,
		UpdatedAt:    tenant.UpdatedAt,
	}
}
