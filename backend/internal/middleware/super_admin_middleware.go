package middleware

import (
	"admin/pkg/constants"
	"admin/pkg/response"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
)

// SuperAdminMiddleware 超级管理员权限中间件
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户信息
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, xerr.ErrUserNotFound)
			c.Abort()
			return
		}

		roleType, ok := userRole.(int)
		if !ok {
			response.Error(c, xerr.ErrInternal)
			c.Abort()
			return
		}

		// 检查是否为超级管理员
		if roleType != constants.RoleTypeSuperAdmin {
			response.Error(c, xerr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
