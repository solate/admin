package dto

import "admin/pkg/pagination"

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	Name      string `json:"name" binding:"required"`         // 菜单名称
	ParentID  string `json:"parent_id" binding:"omitempty"`    // 父菜单ID
	Path      string `json:"path" binding:"omitempty"`         // 前端路由路径
	Component string `json:"component" binding:"omitempty"`    // 前端组件路径
	Redirect  string `json:"redirect" binding:"omitempty"`     // 重定向路径
	Icon      string `json:"icon" binding:"omitempty"`         // 图标
	Sort      *int16 `json:"sort" binding:"omitempty"`         // 排序
	Status    int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:显示 2:隐藏
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	Name      string `json:"name" binding:"omitempty"`          // 菜单名称
	ParentID  string `json:"parent_id" binding:"omitempty"`     // 父菜单ID
	Path      string `json:"path" binding:"omitempty"`          // 前端路由路径
	Component string `json:"component" binding:"omitempty"`     // 前端组件路径
	Redirect  string `json:"redirect" binding:"omitempty"`      // 重定向路径
	Icon      string `json:"icon" binding:"omitempty"`          // 图标
	Sort      *int16 `json:"sort" binding:"omitempty"`          // 排序
	Status    int    `json:"status" binding:"omitempty,oneof=1 2"` // 状态 1:显示 2:隐藏
}

// MenuInfo 菜单信息
type MenuInfo struct {
	MenuID      string  `json:"menu_id"`      // 菜单ID（修正：PermissionID -> MenuID）
	Name        string  `json:"name"`         // 菜单名称
	Type        string  `json:"type"`         // 类型（固定为 "MENU"）
	ParentID    *string `json:"parent_id"`    // 父菜单ID
	Resource    *string `json:"resource"`     // 资源路径（menu:menu_id）
	Action      *string `json:"action"`       // 请求方法（固定为 "*"）
	Path        *string `json:"path"`         // 前端路由路径
	Component   *string `json:"component"`    // 前端组件路径
	Redirect    *string `json:"redirect"`     // 重定向路径
	Icon        *string `json:"icon"`         // 图标
	Sort        *int16  `json:"sort"`         // 排序
	Status      int16   `json:"status"`       // 状态
	Description *string `json:"description"`  // 描述
	CreatedAt   int64   `json:"created_at"`   // 创建时间
	UpdatedAt   int64   `json:"updated_at"`   // 更新时间
}

// MenuTreeNode 菜单树节点
type MenuTreeNode struct {
	*MenuInfo
	Children []*MenuTreeNode `json:"children"` // 子菜单
}

// ListMenusRequest 菜单列表请求
type ListMenusRequest struct {
	pagination.Request `json:",inline"`
	Name                string `form:"name" binding:"omitempty"`          // 菜单名称搜索
	Status              *int16 `form:"status" binding:"omitempty,oneof=1 2"` // 状态筛选
}

// ListMenusResponse 菜单列表响应
type ListMenusResponse struct {
	pagination.Response `json:",inline"`
	List                []*MenuInfo `json:"list"` // 列表数据
}

// MenuTreeResponse 菜单树响应
type MenuTreeResponse struct {
	List []*MenuTreeNode `json:"list"` // 菜单树
}

// AllMenusResponse 所有菜单响应（平铺）
type AllMenusResponse struct {
	List []*MenuInfo `json:"list"` // 菜单列表
}

// MenuDetailResponse 菜单详情响应
type MenuDetailResponse struct {
	*MenuInfo
}

// UpdateMenuStatusRequest 更新菜单状态请求
type UpdateMenuStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=1 2"` // 状态 1:显示 2:隐藏
}
