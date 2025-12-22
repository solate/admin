package common

// PageRequest 分页请求
type PageRequest struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1"`                   // 当前页
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"` // 每页数量
}

// PageResponse 分页响应
type PageResponse struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// GetPageParams 获取分页参数（带默认值）
func (p *PageRequest) GetPageParams() (page, pageSize int) {
	page = p.Page
	if page <= 0 {
		page = 1
	}
	pageSize = p.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 { // 最大分页大小为100
		pageSize = 100
	}
	return
}

// GetOffset 获取偏移量
func GetOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}
