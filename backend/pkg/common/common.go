package common

type TimeRange struct {
	StartTime string `form:"start_time"` // 开始时间
	EndTime   string `form:"end_time"`   // 结束时间
}

// IDRequest ID请求
type IDRequest struct {
	ID uint `uri:"id" json:"id" binding:"required,min=1"`
}

// 改变状态
type StatusRequest struct {
	ID     uint `uri:"id" json:"id" binding:"required,min=1"` // ID
	Status int  `json:"status" binding:"required,oneof=1 2"`  // 状态：1-启用，2-禁用
}
