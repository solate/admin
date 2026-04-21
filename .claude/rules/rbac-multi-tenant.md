# RBAC 与多租户架构规则

> Casbin 已完全移除（2026-04），权限系统改为纯数据库 RBAC。

## 权限系统

- **PermissionCache**（内存）：从 role_permissions + permissions 表加载，三张 map 分别缓存 API/MENU/BUTTON 权限
- **角色继承**：roles.parent_role_id + 递归 CTE（`role_ancestors`，depth < 10 防循环）
- **缓存刷新**：`NotifyRefresh()`（事件驱动）+ TTL 定时器 + `watchRefresh` 协程
- **JWT Claims**：Roles []string（角色编码）+ RoleIDs []string（角色ID，用于 PermissionCache 查询）

## 多租户

- **租户隔离查询**：`.Where(r.q.XXX.TenantID.Eq(xcontext.GetTenantID(ctx)))`
- **创建赋值**：`model.TenantID = xcontext.GetTenantID(ctx)`
- **跨租户查询**：不加 TenantID 条件（登录、超管统计、PermissionCache 加载等）
- **全局表**（无 tenant_id）：tenants, menus, permissions, user_positions

## 禁止

- 禁止引入 Casbin 或任何策略引擎
- 禁止使用 SkipTenantCheck（已删除）
- 禁止使用 database.TenantScope()（已移除，用 Gen Where 替代）
- 禁止在 Repository 外部写 SQL 查权限 — 统一走 PermissionCache

## CopyContext

异步场景必须用 `xcontext.CopyContext(ctx)` 拷贝全部字段：
TenantID, TenantCode, UserID, UserName, TokenID, Roles, RoleIDs
