package department

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
)

// GetDepartmentTree 获取部门树
func (s *Service) GetDepartmentTree(ctx context.Context) (*dto.DepartmentTreeResponse, error) {
	// 获取所有部门
	allDepts, err := s.deptRepo.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("查询部门列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询部门列表失败", err)
	}

	// 构建部门树
	tree := s.buildDepartmentTree(allDepts, "")

	return &dto.DepartmentTreeResponse{
		Tree: tree,
	}, nil
}

// buildDepartmentTree 构建部门树
func (s *Service) buildDepartmentTree(depts []*model.Department, parentID string) []*dto.DepartmentTreeNode {
	var tree []*dto.DepartmentTreeNode

	// 找出所有子节点
	for _, dept := range depts {
		if dept.ParentID == parentID {
			node := &dto.DepartmentTreeNode{
				DepartmentInfo: modelToDepartmentInfo(dept),
				Children:       s.buildDepartmentTree(depts, dept.DepartmentID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}
