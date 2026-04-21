package department

import (
	"admin/internal/dal/model"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/convert"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeleteDepartment 删除部门
func (s *Service) DeleteDepartment(ctx context.Context, departmentID string) (err error) {
	var dept *model.Department

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleDepartment),
				audit.WithError(err),
			)
		} else if dept != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleDepartment),
				audit.WithResource(constants.ResourceTypeDepartment, dept.DepartmentID, dept.DepartmentName),
				audit.WithValue(dept, nil),
			)
			log.Info().Str("department_id", departmentID).Str("department_name", dept.DepartmentName).Msg("删除部门成功")
		}
	}()

	// 检查部门是否存在
	dept, err = s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("department_id", departmentID).Msg("部门不存在")
			return xerr.ErrDeptNotFound
		}
		log.Error().Err(err).Str("department_id", departmentID).Msg("查询部门失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	// 检查是否有子部门
	hasChildren, err := s.deptRepo.HasChildren(ctx, departmentID)
	if err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Msg("检查子部门失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "检查子部门失败", err)
	}
	if hasChildren {
		log.Warn().Str("department_id", departmentID).Msg("部门存在子部门，无法删除")
		return xerr.ErrDeptHasChildren
	}

	// 检查是否有关联用户
	count, err := s.userRepo.CountByDept(ctx, departmentID)
	if err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Msg("检查部门用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "检查部门用户失败", err)
	}
	if count > 0 {
		log.Warn().Str("department_id", departmentID).Int("user_count", int(count)).Msg("部门存在关联用户，无法删除")
		return xerr.ErrDeptHasUsers
	}

	// 删除部门
	if err := s.deptRepo.Delete(ctx, departmentID); err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Str("department_name", dept.DepartmentName).Msg("删除部门失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除部门失败", err)
	}

	return nil
}

// BatchDeleteDepartments 批量删除部门
func (s *Service) BatchDeleteDepartments(ctx context.Context, departmentIDs []string) (err error) {
	var deptMap map[string]*model.Department

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleDepartment),
				audit.WithError(err),
			)
		} else if len(deptMap) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(deptMap))
			names := make([]string, 0, len(deptMap))
			for _, dept := range deptMap {
				ids = append(ids, dept.DepartmentID)
				names = append(names, dept.DepartmentName)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleDepartment),
				audit.WithBatchResource(constants.ResourceTypeDepartment, ids, names),
				audit.WithValue(deptMap, nil),
			)
			log.Info().Strs("department_ids", departmentIDs).Int("count", len(departmentIDs)).Msg("批量删除部门成功")
		}
	}()

	// 获取所有部门信息
	departments, err := s.deptRepo.GetByIDs(ctx, departmentIDs)
	if err != nil {
		log.Error().Err(err).Strs("department_ids", departmentIDs).Msg("查询部门信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询部门信息失败", err)
	}
	deptMap = convert.ToMap(departments, func(d *model.Department) string { return d.DepartmentID })

	// 验证所有部门都存在
	if len(deptMap) != len(departmentIDs) {
		var missingIDs []string
		for _, id := range departmentIDs {
			if _, exists := deptMap[id]; !exists {
				missingIDs = append(missingIDs, id)
			}
		}
		log.Warn().Strs("missing_ids", missingIDs).Msg("部分部门不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分部门不存在")
	}

	// 检查是否有部门存在子部门
	for _, departmentID := range departmentIDs {
		hasChildren, err := s.deptRepo.HasChildren(ctx, departmentID)
		if err != nil {
			log.Error().Err(err).Str("department_id", departmentID).Msg("检查子部门失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "检查子部门失败", err)
		}
		if hasChildren {
			log.Warn().Str("department_id", departmentID).Msg("部门存在子部门，无法删除")
			return xerr.ErrDeptHasChildren
		}
	}

	// 检查是否有关联用户
	for _, departmentID := range departmentIDs {
		count, err := s.userRepo.CountByDept(ctx, departmentID)
		if err != nil {
			log.Error().Err(err).Str("department_id", departmentID).Msg("检查部门用户失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "检查部门用户失败", err)
		}
		if count > 0 {
			log.Warn().Str("department_id", departmentID).Int("user_count", int(count)).Msg("部门存在关联用户，无法删除")
			return xerr.ErrDeptHasUsers
		}
	}

	// 批量删除部门
	if err := s.deptRepo.BatchDelete(ctx, departmentIDs); err != nil {
		log.Error().Err(err).Strs("department_ids", departmentIDs).Msg("批量删除部门失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除部门失败", err)
	}

	return nil
}
