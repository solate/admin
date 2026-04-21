package middleware

import (
	"admin/internal/rbac"
	"admin/pkg/constants"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RBACMiddleware 基于 PermissionCache 的权限中间件
func RBACMiddleware(cache *rbac.PermissionCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 超管跳过权限检查
		if xcontext.HasRole(ctx, constants.SuperAdmin) {
			c.Next()
			return
		}

		roleIDs := xcontext.GetRoleIDs(ctx)
		if len(roleIDs) == 0 {
			response.ErrorWithHttpCode(c, http.StatusForbidden, xerr.ErrForbidden)
			c.Abort()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		if !cache.CheckAPI(roleIDs, path, method) {
			log.Warn().
				Strs("role_ids", roleIDs).
				Str("path", path).
				Str("method", method).
				Msg("[RBACMiddleware] 权限不足")
			response.Error(c, xerr.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
