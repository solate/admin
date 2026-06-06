package department

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateDepartment 更新部门
func (s *Service) UpdateDepartment(ctx context.Context, departmentID string, req *dto.UpdateDepartmentRequest) (resp *dto.DepartmentInfo, err error) {
	var oldDept, newDept *model.Department

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleDepartment),
				audit.WithError(err),
			)
		} else if newDept != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleDepartment),
				audit.WithResource(constants.ResourceTypeDepartment, newDept.DepartmentID, newDept.DepartmentName),
				audit.WithValue(oldDept, newDept),
			)
		}
	}()

	// 获取旧部门信息
	oldDept, err = s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("department_id", departmentID).Msg("部门不存在")
			return nil, xerr.ErrDeptNotFound
		}
		log.Error().Err(err).Str("department_id", departmentID).Msg("查询部门失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	// 如果要修改父部门，验证父部门是否存在
	if req.ParentID != "" && req.ParentID != oldDept.ParentID {
		// 不能将部门设置为自己的子部门
		if req.ParentID == departmentID {
			log.Warn().Str("department_id", departmentID).Str("parent_id", req.ParentID).Msg("不能将部门设置为自己的父部门")
			return nil, xerr.ErrInvalidParentDept
		}

		// 检查是否将部门设置为自己的子孙部门
		descendantIDs, err := s.deptRepo.GetDescendantIDs(ctx, departmentID)
		if err != nil {
			log.Error().Err(err).Str("department_id", departmentID).Msg("获取子部门列表失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取子部门列表失败", err)
		}
		for _, id := range descendantIDs {
			if id == req.ParentID {
				log.Warn().Str("department_id", departmentID).Str("parent_id", req.ParentID).Msg("不能将部门设置为自己的子孙部门")
				return nil, xerr.ErrInvalidParentDept
			}
		}

		parentDept, err := s.deptRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("parent_id", req.ParentID).Msg("父部门不存在")
				return nil, xerr.ErrParentDeptNotFound
			}
			log.Error().Err(err).Str("parent_id", req.ParentID).Msg("查询父部门失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父部门失败", err)
		}
		if parentDept.TenantID != oldDept.TenantID {
			log.Warn().Str("parent_id", req.ParentID).Str("tenant_id", oldDept.TenantID).Msg("父部门不属于当前租户")
			return nil, xerr.ErrParentDeptNotFound
		}
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.DepartmentName != "" {
		updates["department_name"] = req.DepartmentName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Sort != 0 {
		updates["sort"] = req.Sort
	}
	if req.Status != constants.StatusZero {
		updates["status"] = req.Status
	}
	if req.ParentID != "" {
		updates["parent_id"] = req.ParentID
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新部门
	if err := s.deptRepo.Update(ctx, departmentID, updates); err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Interface("updates", updates).Msg("更新部门失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新部门失败", err)
	}

	// 获取更新后的部门信息
	newDept, err = s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Msg("获取更新后部门信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后部门信息失败", err)
	}

	return modelToDepartmentInfo(newDept), nil
}

// UpdateDepartmentStatus 更新部门状态
func (s *Service) UpdateDepartmentStatus(ctx context.Context, departmentID string, status int) (err error) {
	var oldDept, newDept *model.Department

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleDepartment),
				audit.WithError(err),
			)
		} else if newDept != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleDepartment),
				audit.WithResource(constants.ResourceTypeDepartment, newDept.DepartmentID, newDept.DepartmentName),
				audit.WithValue(oldDept, newDept),
			)
			log.Info().Str("department_id", departmentID).Str("department_name", newDept.DepartmentName).Int("status", status).Msg("更新部门状态成功")
		}
	}()

	// 获取旧部门信息
	oldDept, err = s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("department_id", departmentID).Msg("部门不存在")
			return xerr.ErrDeptNotFound
		}
		log.Error().Err(err).Str("department_id", departmentID).Msg("查询部门失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	// 更新部门状态
	if err := s.deptRepo.UpdateStatus(ctx, departmentID, status); err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Int("status", status).Msg("更新部门状态失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新部门状态失败", err)
	}

	// 获取更新后的部门信息
	newDept, err = s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Msg("获取更新后部门信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后部门信息失败", err)
	}

	return nil
}
