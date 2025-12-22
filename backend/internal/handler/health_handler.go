package handler

import (
	"admin/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 健康检查
// @Summary 健康检查
// @Description 检查服务运行状态，返回服务基本信息
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string} "健康检查成功"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "ok",
		"service": "admin-backend",
		"version": "1.0.0",
	})
}

// Ping Ping接口
// @Summary 服务连通性测试
// @Description 测试服务是否可达，用于网络连通性检查
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Ping成功响应"
// @Failure 200 {object} response.Response "服务器内部错误"
// @Router /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"service": "admin-backend",
	})
}
