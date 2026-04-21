package rbac

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// APIPermission API 权限
type APIPermission struct {
	Path   string
	Method string
}

// PermissionCache 权限缓存
// 基于 role_ancestors 递归 CTE 实现角色继承的权限查询
type PermissionCache struct {
	mu          sync.RWMutex
	apiPerms    map[string][]APIPermission // roleID → API 权限列表
	menuPerms   map[string][]string        // roleID → menuID 列表
	buttonPerms map[string][]string        // roleID → button permission ID 列表
	db          *gorm.DB
	ttl         time.Duration
	lastLoad    time.Time
	refreshCh   chan struct{}
	stopCh      chan struct{}
}

// NewPermissionCache 创建权限缓存并启动后台刷新
func NewPermissionCache(db *gorm.DB, ttl time.Duration) *PermissionCache {
	c := &PermissionCache{
		apiPerms:    make(map[string][]APIPermission),
		menuPerms:   make(map[string][]string),
		buttonPerms: make(map[string][]string),
		db:          db,
		ttl:         ttl,
		refreshCh:   make(chan struct{}, 1),
		stopCh:      make(chan struct{}),
	}

	// 启动后台刷新协程
	go c.watchRefresh()

	return c
}

// NotifyRefresh 通知缓存需要刷新（非阻塞）
// 在角色/权限变更时调用
func (c *PermissionCache) NotifyRefresh() {
	select {
	case c.refreshCh <- struct{}{}:
	default:
		// 已经有待处理的刷新请求，跳过
	}
}

// Stop 停止后台刷新协程
func (c *PermissionCache) Stop() {
	close(c.stopCh)
}

// watchRefresh 后台监听刷新请求和 TTL 过期
func (c *PermissionCache) watchRefresh() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	// 启动时立即加载一次
	ctx := context.Background()
	if err := c.Refresh(ctx); err != nil {
		log.Error().Err(err).Msg("权限缓存初始加载失败")
	}

	for {
		select {
		case <-c.stopCh:
			return
		case <-c.refreshCh:
			if err := c.Refresh(ctx); err != nil {
				log.Error().Err(err).Msg("权限缓存刷新失败")
			}
		case <-ticker.C:
			if err := c.Refresh(ctx); err != nil {
				log.Error().Err(err).Msg("权限缓存定时刷新失败")
			}
		}
	}
}

// roleAncestorsCTE 返回 descendant→ancestor 映射的递归 CTE
// 对于每个 role_id，找到它和所有祖先角色的 ancestor_role_id
const roleAncestorsCTE = `
WITH RECURSIVE role_ancestors AS (
    -- 基础：每个角色是自己的祖先
    SELECT role_id, role_id AS ancestor_role_id, 0 AS depth
    FROM roles
    WHERE deleted_at = 0
    UNION
    -- 递归：祖先的父角色也是祖先（depth < 10 防止循环引用）
    SELECT ra.role_id, r.parent_role_id AS ancestor_role_id, ra.depth + 1
    FROM role_ancestors ra
    JOIN roles r ON r.role_id = ra.ancestor_role_id
    WHERE r.parent_role_id != '' AND r.deleted_at = 0 AND ra.depth < 10
)
`

// Refresh 刷新权限缓存
func (c *PermissionCache) Refresh(ctx context.Context) error {
	newAPIPerms := make(map[string][]APIPermission)
	newMenuPerms := make(map[string][]string)
	newButtonPerms := make(map[string][]string)

	// 查询所有角色的 API 权限（含继承角色的权限）
	var apiResults []struct {
		RoleID   string
		Resource string
		Action   string
	}
	err := c.db.WithContext(ctx).Raw(roleAncestorsCTE + `
		SELECT DISTINCT ra.role_id, p.resource, p.action
		FROM role_ancestors ra
		JOIN role_permissions rp ON rp.role_id = ra.ancestor_role_id
		JOIN permissions p ON p.permission_id = rp.permission_id
		WHERE p.deleted_at = 0 AND p.status = 1 AND p.type = 'API'
	`).Scan(&apiResults).Error
	if err != nil {
		return err
	}

	for _, r := range apiResults {
		newAPIPerms[r.RoleID] = append(newAPIPerms[r.RoleID], APIPermission{
			Path:   r.Resource,
			Method: r.Action,
		})
	}

	// 查询所有角色的菜单权限（含继承）
	var menuResults []struct {
		RoleID   string
		Resource string
	}
	err = c.db.WithContext(ctx).Raw(roleAncestorsCTE + `
		SELECT DISTINCT ra.role_id, p.resource
		FROM role_ancestors ra
		JOIN role_permissions rp ON rp.role_id = ra.ancestor_role_id
		JOIN permissions p ON p.permission_id = rp.permission_id
		WHERE p.deleted_at = 0 AND p.status = 1 AND p.type = 'MENU'
	`).Scan(&menuResults).Error
	if err != nil {
		return err
	}

	for _, r := range menuResults {
		menuID := strings.TrimPrefix(r.Resource, "menu:")
		newMenuPerms[r.RoleID] = append(newMenuPerms[r.RoleID], menuID)
	}

	// 查询所有角色的按钮权限（含继承）
	var buttonResults []struct {
		RoleID       string
		PermissionID string
	}
	err = c.db.WithContext(ctx).Raw(roleAncestorsCTE + `
		SELECT DISTINCT ra.role_id, p.permission_id
		FROM role_ancestors ra
		JOIN role_permissions rp ON rp.role_id = ra.ancestor_role_id
		JOIN permissions p ON p.permission_id = rp.permission_id
		WHERE p.deleted_at = 0 AND p.status = 1 AND p.type = 'BUTTON'
	`).Scan(&buttonResults).Error
	if err != nil {
		return err
	}

	for _, r := range buttonResults {
		newButtonPerms[r.RoleID] = append(newButtonPerms[r.RoleID], r.PermissionID)
	}

	c.mu.Lock()
	c.apiPerms = newAPIPerms
	c.menuPerms = newMenuPerms
	c.buttonPerms = newButtonPerms
	c.lastLoad = time.Now()
	c.mu.Unlock()

	log.Info().Int("api_rules", len(apiResults)).Int("menu_rules", len(menuResults)).
		Msg("权限缓存刷新完成")

	return nil
}

// CheckAPI 检查角色是否有指定 API 权限
func (c *PermissionCache) CheckAPI(roleIDs []string, path, method string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, roleID := range roleIDs {
		for _, perm := range c.apiPerms[roleID] {
			if matchPath(perm.Path, path) && matchMethod(perm.Method, method) {
				return true
			}
		}
	}
	return false
}

// GetMenuIDs 获取角色的菜单 ID 列表
func (c *PermissionCache) GetMenuIDs(roleIDs []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	seen := make(map[string]bool)
	var result []string
	for _, roleID := range roleIDs {
		for _, menuID := range c.menuPerms[roleID] {
			if !seen[menuID] {
				seen[menuID] = true
				result = append(result, menuID)
			}
		}
	}
	return result
}

// GetButtonPermissionIDs 获取角色的按钮权限 ID 列表
func (c *PermissionCache) GetButtonPermissionIDs(roleIDs []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	seen := make(map[string]bool)
	var result []string
	for _, roleID := range roleIDs {
		for _, permID := range c.buttonPerms[roleID] {
			if !seen[permID] {
				seen[permID] = true
				result = append(result, permID)
			}
		}
	}
	return result
}

// matchPath 路径匹配
// 支持：
//   - 精确匹配: /api/v1/users
//   - 单段通配: /api/v1/users/:id 或 /api/v1/users/*
//   - 多段通配: /api/v1/**
func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	if pattern == "/**" {
		return true
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	// ** 匹配剩余所有段
	for i := 0; i < len(patternParts); i++ {
		if patternParts[i] == "**" {
			// ** 必须是最后一段
			return true
		}
		if i >= len(pathParts) {
			return false
		}
		if strings.HasPrefix(patternParts[i], ":") || patternParts[i] == "*" {
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}

	return len(patternParts) == len(pathParts)
}

// matchMethod HTTP 方法匹配
func matchMethod(pattern, method string) bool {
	return pattern == "*" || strings.EqualFold(pattern, method)
}
