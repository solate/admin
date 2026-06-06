package department

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateDepartment 创建部门
func (s *Service) CreateDepartment(ctx context.Context, req *dto.CreateDepartmentRequest) (resp *dto.DepartmentInfo, err error) {
	var dept *model.Department

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleDepartment),
				audit.WithError(err),
			)
		} else if dept != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleDepartment),
				audit.WithResource(constants.ResourceTypeDepartment, dept.DepartmentID, dept.DepartmentName),
				audit.WithValue(nil, dept),
			)
		}
	}()

	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 如果指定了父部门，验证父部门是否存在
	if req.ParentID != "" {
		parentDept, err := s.deptRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("parent_id", req.ParentID).Msg("父部门不存在")
				return nil, xerr.ErrParentDeptNotFound
			}
			log.Error().Err(err).Str("parent_id", req.ParentID).Msg("查询父部门失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父部门失败", err)
		}
		if parentDept.TenantID != tenantID {
			log.Warn().Str("parent_id", req.ParentID).Str("tenant_id", tenantID).Msg("父部门不属于当前租户")
			return nil, xerr.ErrParentDeptNotFound
		}
	}

	// 生成部门ID
	var deptID string
	deptID, err = idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成部门ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成部门ID失败", err)
	}

	// 设置默认值
	sort := req.Sort
	if sort == 0 {
		sort = 0 // 默认排序
	}
	status := req.Status
	if status == 0 {
		status = 1 // 默认启用
	}

	// 构建部门模型
	dept = &model.Department{
		DepartmentID:   deptID,
		TenantID:       tenantID,
		ParentID:       req.ParentID,
		DepartmentName: req.DepartmentName,
		Description:    req.Description,
		Sort:           int32(sort),
		Status:         int16(status),
	}

	// 创建部门
	if err := s.deptRepo.Create(ctx, dept); err != nil {
		log.Error().Err(err).Str("department_id", deptID).Str("department_name", req.DepartmentName).Str("parent_id", req.ParentID).Msg("创建部门失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建部门失败", err)
	}

	return modelToDepartmentInfo(dept), nil
}
