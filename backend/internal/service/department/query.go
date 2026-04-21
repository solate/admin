package department

import (
	"admin/internal/dto"
	"admin/pkg/utils/pagination"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetDepartmentByID 获取部门详情
func (s *Service) GetDepartmentByID(ctx context.Context, departmentID string) (*dto.DepartmentInfo, error) {
	dept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("department_id", departmentID).Msg("部门不存在")
			return nil, xerr.ErrDeptNotFound
		}
		log.Error().Err(err).Str("department_id", departmentID).Msg("查询部门失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	return modelToDepartmentInfo(dept), nil
}

// ListDepartments 获取部门列表
func (s *Service) ListDepartments(ctx context.Context, req *dto.ListDepartmentsRequest) (*dto.ListDepartmentsResponse, error) {
	depts, total, err := s.deptRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.DepartmentName, req.Status, req.ParentID)
	if err != nil {
		log.Error().Err(err).
			Str("department_name", req.DepartmentName).
			Int("status", req.Status).
			Str("parent_id", req.ParentID).
			Msg("查询部门列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门列表失败", err)
	}

	return &dto.ListDepartmentsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     modelListToDepartmentInfoList(depts),
	}, nil
}

// GetChildren 获取子部门
func (s *Service) GetChildren(ctx context.Context, departmentID string) ([]*dto.DepartmentInfo, error) {
	children, err := s.deptRepo.GetChildren(ctx, departmentID)
	if err != nil {
		log.Error().Err(err).Str("department_id", departmentID).Msg("查询子部门失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询子部门失败", err)
	}

	return modelListToDepartmentInfoList(children), nil
}
