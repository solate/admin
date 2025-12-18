package middleware

import (
	"admin/pkg/constants"
	"admin/pkg/response"
	"admin/pkg/xerr"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuperAdmin 超级管理员权限中间件
func SuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户信息
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, xerr.ErrUserNotFound)
			c.Abort()
			return
		}

		roleType, ok := userRole.(int)
		if !ok {
			response.Error(c, http.StatusUnauthorized, xerr.ErrInternal)
			c.Abort()
			return
		}

		// 检查是否为超级管理员
		if roleType != constants.RoleTypeSuperAdmin {
			response.Error(c, http.StatusForbidden, xerr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
