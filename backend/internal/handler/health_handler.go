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
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "ok",
		"service": "admin-backend",
	})
}

// Ping Ping接口
// @Summary Ping
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response
// @Router /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
