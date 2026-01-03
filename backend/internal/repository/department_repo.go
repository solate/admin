package repository

import (
	"admin/internal/dal/model"
	"admin/internal/dal/query"

	"context"

	"gorm.io/gorm"
)

type DepartmentRepo struct {
	db *gorm.DB
	q  *query.Query
}

func NewDepartmentRepo(db *gorm.DB) *DepartmentRepo {
	return &DepartmentRepo{
		db: db,
		q:  query.Use(db),
	}
}

// Create 创建部门
func (r *DepartmentRepo) Create(ctx context.Context, department *model.Department) error {
	return r.q.Department.WithContext(ctx).Create(department)
}

// GetByID 根据ID获取部门
func (r *DepartmentRepo) GetByID(ctx context.Context, departmentID string) (*model.Department, error) {
	return r.q.Department.WithContext(ctx).Where(r.q.Department.DepartmentID.Eq(departmentID)).First()
}

// Update 更新部门
func (r *DepartmentRepo) Update(ctx context.Context, departmentID string, updates map[string]interface{}) error {
	_, err := r.q.Department.WithContext(ctx).Where(r.q.Department.DepartmentID.Eq(departmentID)).Updates(updates)
	return err
}

// Delete 删除部门(软删除)
func (r *DepartmentRepo) Delete(ctx context.Context, departmentID string) error {
	_, err := r.q.Department.WithContext(ctx).Where(r.q.Department.DepartmentID.Eq(departmentID)).Delete()
	return err
}

// List 获取租户的所有部门（按排序权重升序）
func (r *DepartmentRepo) List(ctx context.Context) ([]*model.Department, error) {
	return r.q.Department.WithContext(ctx).
		Order(r.q.Department.Sort.Asc()).
		Find()
}

// ListWithFilters 根据筛选条件分页获取部门列表
func (r *DepartmentRepo) ListWithFilters(ctx context.Context, offset, limit int, keywordFilter string, statusFilter int, parentIDFilter string) ([]*model.Department, int64, error) {
	q := r.q.Department.WithContext(ctx)

	// 应用筛选条件
	if keywordFilter != "" {
		q = q.Where(r.q.Department.DepartmentName.Like("%" + keywordFilter + "%"))
	}
	if statusFilter != 0 {
		q = q.Where(r.q.Department.Status.Eq(int16(statusFilter)))
	}
	if parentIDFilter != "" {
		q = q.Where(r.q.Department.ParentID.Eq(parentIDFilter))
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	depts, err := q.Order(r.q.Department.Sort.Asc()).Offset(offset).Limit(limit).Find()
	return depts, total, err
}

// GetChildren 获取子部门
func (r *DepartmentRepo) GetChildren(ctx context.Context, parentID string) ([]*model.Department, error) {
	return r.q.Department.WithContext(ctx).
		Where(r.q.Department.ParentID.Eq(parentID)).
		Order(r.q.Department.Sort.Asc()).
		Find()
}

// GetDescendantIDs 获取部门及所有子部门ID（递归）
func (r *DepartmentRepo) GetDescendantIDs(ctx context.Context, departmentID string) ([]string, error) {
	var ids []string
	ids = append(ids, departmentID)

	children, err := r.GetChildren(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		childIDs, err := r.GetDescendantIDs(ctx, child.DepartmentID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, childIDs...)
	}

	return ids, nil
}

// UpdateStatus 更新部门状态
func (r *DepartmentRepo) UpdateStatus(ctx context.Context, departmentID string, status int) error {
	_, err := r.q.Department.WithContext(ctx).Where(r.q.Department.DepartmentID.Eq(departmentID)).Update(r.q.Department.Status, int16(status))
	return err
}

// CountByDepartmentID 统计部门下的用户数
func (r *DepartmentRepo) CountByDepartmentID(ctx context.Context, userRepo *UserRepo, departmentID string) (int64, error) {
	return userRepo.CountByDept(ctx, departmentID)
}

// HasChildren 检查是否有子部门
func (r *DepartmentRepo) HasChildren(ctx context.Context, departmentID string) (bool, error) {
	count, err := r.q.Department.WithContext(ctx).
		Where(r.q.Department.ParentID.Eq(departmentID)).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
