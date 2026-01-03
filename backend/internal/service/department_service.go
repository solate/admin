package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/auditlog"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"gorm.io/gorm"
)

// DepartmentService 部门服务
type DepartmentService struct {
	deptRepo *repository.DepartmentRepo
	userRepo *repository.UserRepo
}

// NewDepartmentService 创建部门服务
func NewDepartmentService(deptRepo *repository.DepartmentRepo, userRepo *repository.UserRepo) *DepartmentService {
	return &DepartmentService{
		deptRepo: deptRepo,
		userRepo: userRepo,
	}
}

// CreateDepartment 创建部门
func (s *DepartmentService) CreateDepartment(ctx context.Context, req *dto.CreateDepartmentRequest) (*dto.DepartmentResponse, error) {
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 如果指定了父部门，验证父部门是否存在
	if req.ParentID != "" {
		parentDept, err := s.deptRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, xerr.ErrParentDeptNotFound
			}
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父部门失败", err)
		}
		if parentDept.TenantID != tenantID {
			return nil, xerr.ErrParentDeptNotFound
		}
	}

	// 生成部门ID
	deptID, err := idgen.GenerateUUID()
	if err != nil {
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
	dept := &model.Department{
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
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建部门失败", err)
	}

	// 记录操作日志
	ctx = auditlog.RecordCreate(ctx, constants.ModuleDept, constants.ResourceTypeDept, dept.DepartmentID, dept.DepartmentName, dept)

	return s.toDepartmentResponse(dept), nil
}

// GetDepartmentByID 获取部门详情
func (s *DepartmentService) GetDepartmentByID(ctx context.Context, departmentID string) (*dto.DepartmentResponse, error) {
	dept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrDeptNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	return s.toDepartmentResponse(dept), nil
}

// UpdateDepartment 更新部门
func (s *DepartmentService) UpdateDepartment(ctx context.Context, departmentID string, req *dto.UpdateDepartmentRequest) (*dto.DepartmentResponse, error) {
	// 检查部门是否存在，获取旧值用于日志
	oldDept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrDeptNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	// 如果要修改父部门，验证父部门是否存在
	if req.ParentID != "" && req.ParentID != oldDept.ParentID {
		// 不能将部门设置为自己的子部门
		if req.ParentID == departmentID {
			return nil, xerr.ErrInvalidParentDept
		}

		// 检查是否将部门设置为自己的子孙部门
		descendantIDs, err := s.deptRepo.GetDescendantIDs(ctx, departmentID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取子部门列表失败", err)
		}
		for _, id := range descendantIDs {
			if id == req.ParentID {
				return nil, xerr.ErrInvalidParentDept
			}
		}

		parentDept, err := s.deptRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, xerr.ErrParentDeptNotFound
			}
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父部门失败", err)
		}
		if parentDept.TenantID != oldDept.TenantID {
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
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	if req.ParentID != "" {
		updates["parent_id"] = req.ParentID
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新部门
	if err := s.deptRepo.Update(ctx, departmentID, updates); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新部门失败", err)
	}

	// 获取更新后的部门信息
	updatedDept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后部门信息失败", err)
	}

	// 记录操作日志
	ctx = auditlog.RecordUpdate(ctx, constants.ModuleDept, constants.ResourceTypeDept, updatedDept.DepartmentID, updatedDept.DepartmentName, oldDept, updatedDept)

	return s.toDepartmentResponse(updatedDept), nil
}

// DeleteDepartment 删除部门
func (s *DepartmentService) DeleteDepartment(ctx context.Context, departmentID string) error {
	// 检查部门是否存在，获取部门信息用于日志
	dept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrDeptNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	// 检查是否有子部门
	hasChildren, err := s.deptRepo.HasChildren(ctx, departmentID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "检查子部门失败", err)
	}
	if hasChildren {
		return xerr.ErrDeptHasChildren
	}

	// 检查是否有关联用户
	count, err := s.userRepo.CountByDept(ctx, departmentID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "检查部门用户失败", err)
	}
	if count > 0 {
		return xerr.ErrDeptHasUsers
	}

	// 删除部门
	if err := s.deptRepo.Delete(ctx, departmentID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除部门失败", err)
	}

	// 记录操作日志
	auditlog.RecordDelete(ctx, constants.ModuleDept, constants.ResourceTypeDept, dept.DepartmentID, dept.DepartmentName, dept)

	return nil
}

// ListDepartments 获取部门列表
func (s *DepartmentService) ListDepartments(ctx context.Context, req *dto.ListDepartmentsRequest) (*dto.ListDepartmentsResponse, error) {
	depts, total, err := s.deptRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status, req.ParentID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门列表失败", err)
	}

	// 转换为响应格式
	deptResponses := make([]*dto.DepartmentResponse, len(depts))
	for i, dept := range depts {
		deptResponses[i] = s.toDepartmentResponse(dept)
	}

	return &dto.ListDepartmentsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     deptResponses,
	}, nil
}

// GetDepartmentTree 获取部门树
func (s *DepartmentService) GetDepartmentTree(ctx context.Context) (*dto.DepartmentTreeResponse, error) {
	// 获取所有部门
	allDepts, err := s.deptRepo.List(ctx)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门列表失败", err)
	}

	// 构建部门树
	tree := s.buildDepartmentTree(allDepts, "")

	return &dto.DepartmentTreeResponse{
		Tree: tree,
	}, nil
}

// GetChildren 获取子部门
func (s *DepartmentService) GetChildren(ctx context.Context, departmentID string) ([]*dto.DepartmentResponse, error) {
	children, err := s.deptRepo.GetChildren(ctx, departmentID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询子部门失败", err)
	}

	responses := make([]*dto.DepartmentResponse, len(children))
	for i, child := range children {
		responses[i] = s.toDepartmentResponse(child)
	}

	return responses, nil
}

// UpdateDepartmentStatus 更新部门状态
func (s *DepartmentService) UpdateDepartmentStatus(ctx context.Context, departmentID string, status int) error {
	// 检查部门是否存在，获取旧值用于日志
	oldDept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrDeptNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询部门失败", err)
	}

	// 更新部门状态
	if err := s.deptRepo.UpdateStatus(ctx, departmentID, status); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新部门状态失败", err)
	}

	// 获取更新后的部门信息
	updatedDept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后部门信息失败", err)
	}

	// 记录操作日志
	auditlog.RecordUpdate(ctx, constants.ModuleDept, constants.ResourceTypeDept, updatedDept.DepartmentID, updatedDept.DepartmentName, oldDept, updatedDept)

	return nil
}

// buildDepartmentTree 构建部门树
func (s *DepartmentService) buildDepartmentTree(depts []*model.Department, parentID string) []*dto.DepartmentTreeNode {
	var tree []*dto.DepartmentTreeNode

	// 找出所有子节点
	for _, dept := range depts {
		if dept.ParentID == parentID {
			node := &dto.DepartmentTreeNode{
				DepartmentResponse: s.toDepartmentResponse(dept),
				Children:           s.buildDepartmentTree(depts, dept.DepartmentID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// toDepartmentResponse 转换为部门响应格式
func (s *DepartmentService) toDepartmentResponse(dept *model.Department) *dto.DepartmentResponse {
	return &dto.DepartmentResponse{
		DepartmentID:   dept.DepartmentID,
		ParentID:       dept.ParentID,
		DepartmentName: dept.DepartmentName,
		Description:    dept.Description,
		Sort:           int(dept.Sort),
		Status:         int(dept.Status),
		CreatedAt:      dept.CreatedAt,
		UpdatedAt:      dept.UpdatedAt,
	}
}
