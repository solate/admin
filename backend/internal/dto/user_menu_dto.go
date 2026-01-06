package dto

// UserMenuResponse 用户菜单响应
type UserMenuResponse struct {
	List []*MenuTreeNode `json:"list"`
}

// UserButtonsRequest 按钮权限请求
type UserButtonsRequest struct {
	MenuID string `form:"menu_id" binding:"required" example:"123456789012345678"`
}

// ButtonInfo 按钮信息
type ButtonInfo struct {
	PermissionID string  `json:"permission_id" example:"123456789012345678"`
	Name         string  `json:"name" example:"创建用户"`
	Action       *string `json:"action" example:"POST"`
	Resource     *string `json:"resource" example:"button:123456789012345678"`
}

// UserButtonsResponse 按钮权限响应
type UserButtonsResponse struct {
	Buttons []*ButtonInfo `json:"buttons"`
}
