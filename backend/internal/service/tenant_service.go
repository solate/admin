package service

// import (
// 	"admin/internal/dal/model"
// 	"admin/internal/dal/query"
// 	"admin/internal/dto"
// 	"admin/pkg/common"
// 	"context"
// 	"errors"
// 	"fmt"
// 	"time"

// 	"admin/pkg/xerr"

// 	"github.com/google/uuid"
// 	"gorm.io/gorm"
// )

// type TenantService struct {
// 	db *gorm.DB
// }

// func NewTenantService(db *gorm.DB) *TenantService {
// 	return &TenantService{
// 		db: db,
// 	}
// }

// // CreateTenant 创建租户
// func (s *TenantService) CreateTenant(ctx context.Context, req *dto.TenantCreateRequest) (*dto.TenantResponse, error) {
// 	q := query.Use(s.db)

// 	// 检查租户编码是否已存在
// 	exists, err := q.Tenant.WithContext(ctx).
// 		Unscoped().
// 		Where(q.Tenant.Code.Eq(req.Code)).
// 		Count()
// 	if err != nil {
// 		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "检查租户编码失败", err)
// 	}
// 	if exists > 0 {
// 		return nil, xerr.ErrConflict
// 	}

// 	// 创建租户
// 	now := time.Now().UnixMilli()
// 	var desc *string
// 	if req.Description != "" {
// 		desc = &req.Description
// 	}
// 	tenant := &model.Tenant{
// 		TenantID:    uuid.New().String(),
// 		Code:        req.Code,
// 		Name:        req.Name,
// 		Description: desc,
// 		Status:      int16(1),
// 		CreatedAt:   now,
// 		UpdatedAt:   now,
// 	}

// 	err = q.Tenant.WithContext(ctx).Create(tenant)
// 	if err != nil {
// 		return nil, xerr.Wrap(xerr.ErrCreateError.Code, "创建租户失败", err)
// 	}

// 	return s.toTenantResponse(tenant), nil
// }

// // GetTenantByID 根据ID获取租户
// func (s *TenantService) GetTenantByID(ctx context.Context, tenantID string) (*dto.TenantResponse, error) {
// 	q := query.Use(s.db)

// 	tenant, err := q.Tenant.WithContext(ctx).
// 		Where(q.Tenant.TenantID.Eq(tenantID)).
// 		First()
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, xerr.ErrNotFound
// 		}
// 		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询租户失败", err)
// 	}

// 	return s.toTenantResponse(tenant), nil
// }

// // UpdateTenant 更新租户
// func (s *TenantService) UpdateTenant(ctx context.Context, tenantID string, req *dto.TenantUpdateRequest) (*dto.TenantResponse, error) {
// 	q := query.Use(s.db)

// 	// 检查租户是否存在
// 	_, err := q.Tenant.WithContext(ctx).
// 		Where(q.Tenant.TenantID.Eq(tenantID)).
// 		First()
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, xerr.ErrNotFound
// 		}
// 		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询租户失败", err)
// 	}

// 	// 更新字段
// 	updates := make(map[string]any)
// 	updates["name"] = req.Name
// 	updates["updated_at"] = time.Now().UnixMilli()

// 	if req.Description != "" {
// 		updates["description"] = req.Description
// 	} else {
// 		updates["description"] = nil
// 	}

// 	if req.Status != nil {
// 		updates["status"] = int16(*req.Status)
// 	}

// 	_, err = q.Tenant.WithContext(ctx).
// 		Where(q.Tenant.TenantID.Eq(tenantID)).
// 		Updates(updates)
// 	if err != nil {
// 		return nil, xerr.Wrap(xerr.ErrUpdateError.Code, "更新租户失败", err)
// 	}

// 	// 重新查询获取最新数据
// 	return s.GetTenantByID(ctx, tenantID)
// }

// // DeleteTenant 删除租户（软删除）
// func (s *TenantService) DeleteTenant(ctx context.Context, tenantID string) error {
// 	q := query.Use(s.db)

// 	// 检查租户是否存在
// 	exists, err := q.Tenant.WithContext(ctx).
// 		Where(q.Tenant.TenantID.Eq(tenantID)).
// 		Count()
// 	if err != nil {
// 		return xerr.Wrap(xerr.ErrQueryError.Code, "检查租户失败", err)
// 	}
// 	if exists == 0 {
// 		return xerr.ErrNotFound
// 	}

// 	// 检查租户下是否还有用户
// 	userCount, err := q.User.WithContext(ctx).
// 		Where(q.User.TenantID.Eq(tenantID)).
// 		Count()
// 	if err != nil {
// 		return xerr.Wrap(xerr.ErrQueryError.Code, "检查租户用户失败", err)
// 	}
// 	if userCount > 0 {
// 		return xerr.New(xerr.ErrBadRequest.Code, fmt.Sprintf("租户下还有 %d 个用户，无法删除", userCount))
// 	}

// 	// 软删除租户
// 	_, err = q.Tenant.WithContext(ctx).
// 		Where(q.Tenant.TenantID.Eq(tenantID)).
// 		Updates(map[string]any{
// 			"deleted_at": time.Now().UnixMilli(),
// 		})
// 	if err != nil {
// 		return xerr.Wrap(xerr.ErrUpdateError.Code, "删除租户失败", err)
// 	}

// 	return nil
// }

// // ListTenants 获取租户列表
// func (s *TenantService) ListTenants(ctx context.Context, req *dto.TenantListRequest) (*dto.TenantListResponse, error) {
// 	q := query.Use(s.db)

// 	dao := q.Tenant.WithContext(ctx)

// 	// 添加筛选条件
// 	if req.Code != "" {
// 		dao = dao.Where(q.Tenant.Code.Like("%" + req.Code + "%"))
// 	}
// 	if req.Name != "" {
// 		dao = dao.Where(q.Tenant.Name.Like("%" + req.Name + "%"))
// 	}
// 	if req.Status != nil {
// 		dao = dao.Where(q.Tenant.Status.Eq(int16(*req.Status)))
// 	}

// 	// 获取总数
// 	total, err := dao.Count()
// 	if err != nil {
// 		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询租户总数失败", err)
// 	}

// 	// 分页查询
// 	page, pageSize := req.GetPageParams()
// 	offset := (page - 1) * pageSize
// 	tenants, err := dao.
// 		Order(q.Tenant.CreatedAt.Desc()).
// 		Offset(offset).
// 		Limit(pageSize).
// 		Find()
// 	if err != nil {
// 		return nil, xerr.Wrap(xerr.ErrQueryError.Code, "查询租户列表失败", err)
// 	}

// 	// 转换响应
// 	list := make([]dto.TenantResponse, len(tenants))
// 	for i, tenant := range tenants {
// 		list[i] = *s.toTenantResponse(tenant)
// 	}

// 	return &dto.TenantListResponse{
// 		List: list,
// 		PageResponse: common.PageResponse{
// 			Total:    int(total),
// 			Page:     page,
// 			PageSize: pageSize,
// 		},
// 	}, nil
// }

// // UpdateTenantStatus 更新租户状态
// func (s *TenantService) UpdateTenantStatus(ctx context.Context, tenantID string, req *dto.TenantStatusRequest) error {
// 	q := query.Use(s.db)

// 	// 检查租户是否存在
// 	exists, err := q.Tenant.WithContext(ctx).
// 		Where(q.Tenant.TenantID.Eq(tenantID)).
// 		Count()
// 	if err != nil {
// 		return xerr.Wrap(xerr.ErrQueryError.Code, "检查租户失败", err)
// 	}
// 	if exists == 0 {
// 		return xerr.ErrNotFound
// 	}

// 	// 更新状态
// 	_, err = q.Tenant.WithContext(ctx).
// 		Where(q.Tenant.TenantID.Eq(tenantID)).
// 		Updates(map[string]interface{}{
// 			"status":     int16(req.Status),
// 			"updated_at": time.Now().UnixMilli(),
// 		})
// 	if err != nil {
// 		return xerr.Wrap(xerr.ErrUpdateError.Code, "更新租户状态失败", err)
// 	}

// 	return nil
// }

// // toTenantResponse 转换为响应对象
// func (s *TenantService) toTenantResponse(tenant *model.Tenant) *dto.TenantResponse {
// 	resp := &dto.TenantResponse{
// 		TenantID:  tenant.TenantID,
// 		Code:      tenant.Code,
// 		Name:      tenant.Name,
// 		Status:    int(tenant.Status),
// 		CreatedAt: tenant.CreatedAt,
// 		UpdatedAt: tenant.UpdatedAt,
// 	}
// 	return resp
// }
