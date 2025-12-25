package dto

// UserMenuResponse 用户菜单响应
type UserMenuResponse struct {
	List []*MenuTreeNode `json:"list"`
}

// UserButtonsRequest 按钮权限请求
type UserButtonsRequest struct {
	MenuID string `form:"menu_id" binding:"required"`
}

// ButtonInfo 按钮信息
type ButtonInfo struct {
	PermissionID string  `json:"permission_id"`
	Name         string  `json:"name"`
	Action       *string `json:"action"`
	Resource     *string `json:"resource"`
}

// UserButtonsResponse 按钮权限响应
type UserButtonsResponse struct {
	Buttons []*ButtonInfo `json:"buttons"`
}
