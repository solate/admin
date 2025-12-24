package pagination

// DefaultPageSize 默认每页大小
const DefaultPageSize = 10

// MaxPageSize 最大每页大小
const MaxPageSize = 100

// Request 分页请求
type Request struct {
	Page     int `form:"page" json:"page" binding:"omitempty,min=1"`                   // 当前页
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,min=1,max=100"` // 每页数量
}

// Response 分页响应
type Response struct {
	List      any   `json:"list"`       // 数据列表
	Page      int   `json:"page"`       // 当前页码
	PageSize  int   `json:"page_size"`  // 每页大小
	Total     int64 `json:"total"`      // 总记录数
	TotalPage int64 `json:"total_page"` // 总页数
}

// ToResponse 从请求构建分页响应
func (r *Request) ToResponse(list any, total int64) Response {
	page, pageSize := r.GetPageParams()
	return Response{
		List:      list,
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: calcTotalPage(total, int64(pageSize)),
	}
}

// GetPageParams 获取分页参数（带默认值）
func (r *Request) GetPageParams() (page, pageSize int) {
	page = r.Page
	if page <= 0 {
		page = 1
	}
	pageSize = r.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return
}

// GetOffset 获取偏移量
func (r *Request) GetOffset() int {
	page, pageSize := r.GetPageParams()
	return (page - 1) * pageSize
}

// GetLimit 获取限制数量
func (r *Request) GetLimit() int {
	_, pageSize := r.GetPageParams()
	return pageSize
}

// calcTotalPage 计算总页数
func calcTotalPage(total, pageSize int64) int64 {
	if pageSize <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}
