package handler

import (
	"admin/internal/service"
	"admin/pkg/constants"
	"admin/pkg/response"
	"admin/pkg/xerr"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PolicyHandler struct {
	service *service.CasbinService
}

func NewPolicyHandler(s *service.CasbinService) *PolicyHandler {
	return &PolicyHandler{service: s}
}

type addPolicyReq struct {
	Sub string `json:"sub" binding:"required"`
	Obj string `json:"obj" binding:"required"`
	Act string `json:"act" binding:"required"`
}

func (h *PolicyHandler) AddPolicy(c *gin.Context) {
	tenantID := c.GetString(constants.CtxTenantID)
	if tenantID == "" {
		response.Error(c, http.StatusForbidden, xerr.ErrForbidden)
		return
	}
	var req addPolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, xerr.ErrBadRequest)
		return
	}
	ok, err := h.service.AddPolicyForTenant(tenantID, req.Sub, req.Obj, req.Act)
	if err != nil || !ok {
		response.Error(c, http.StatusInternalServerError, xerr.ErrInternal)
		return
	}
	response.Success(c, gin.H{"added": ok})
}

type addRoleReq struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

func (h *PolicyHandler) AddRole(c *gin.Context) {
	tenantID := c.GetString(constants.CtxTenantID)
	if tenantID == "" {
		response.Error(c, http.StatusForbidden, xerr.ErrForbidden)
		return
	}
	var req addRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, xerr.ErrBadRequest)
		return
	}
	ok, err := h.service.AddRoleForUserInTenant(req.UserID, req.RoleID, tenantID)
	if err != nil || !ok {
		response.Error(c, http.StatusInternalServerError, xerr.ErrInternal)
		return
	}
	response.Success(c, gin.H{"added": ok})
}
