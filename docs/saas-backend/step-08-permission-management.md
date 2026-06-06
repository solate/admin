# Step 08: 权限管理接口 — 权限 CRUD / 角色分配权限

## 目标

实现权限资源的 CRUD 管理接口，以及角色分配/取消权限的功能。这些接口本身需要超管权限。

## 前置条件

- Step 07 完成，RBAC 中间件和 PermissionCache 可用

## 文件清单

```
internal/
├── handler/permission/
│   ├── handler.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go               # List + GetTree（树形）
│   └── dto.go
├── service/permission/
│   ├── service.go
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   └── converter.go
├── repository/
│   ├── permission_repo.go     # 权限 CRUD
│   └── role_permission_repo.go # 角色-权限关联
```

## 实现规范

### 1. API 设计

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/v1/permissions | 权限列表（树形） | super_admin |
| POST | /api/v1/permissions | 创建权限 | super_admin |
| PUT | /api/v1/permissions/:id | 更新权限 | super_admin |
| DELETE | /api/v1/permissions/:id | 删除权限 | super_admin |
| POST | /api/v1/roles/:id/permissions | 分配权限给角色 | admin |
| GET | /api/v1/roles/:id/permissions | 获取角色权限列表 | admin |

### 2. 权限初始化策略

权限表是**全局表**（无 tenant_id），由系统管理员（super_admin）统一管理。

初始化方式：
- **SQL 种子数据**：在 `scripts/seed/` 中维护基础权限定义
- **每增加新接口**：在 migrations 中追加 INSERT 语句
- **不支持动态创建 API 类型权限**：API 权限与代码路由一一对应，必须预定义

### 3. 角色分配权限

```go
// POST /api/v1/roles/:id/permissions
// Body: { "permission_ids": ["p1", "p2", "p3"] }

// 实现策略：先删后插（全量替换）
// 1. DELETE FROM role_permissions WHERE role_id = ? AND tenant_id = ?
// 2. INSERT INTO role_permissions (role_id, permission_id, tenant_id, created_at) VALUES ...
// 3. rbacCache.NotifyRefresh()
```

**全量替换 vs 增量**：
- 选择全量替换：前端传当前角色应该拥有的所有权限 ID 列表
- 简单可靠，不需要前端计算差集
- 事务内执行（先删后插）

### 4. 权限树形返回

```go
// 权限按 type 分组，每组内按 parent_id 构建树
// 返回格式：
// [
//   { type: "API", children: [...] },
//   { type: "MENU", children: [...] },
//   { type: "BUTTON", children: [...] }
// ]
```

### 5. DTO

```go
type CreatePermissionRequest struct {
    Name        string `json:"name" binding:"required"`
    Type        string `json:"type" binding:"required,oneof=API MENU BUTTON"`
    Resource    string `json:"resource" binding:"required"`
    Action      string `json:"action"`     // API 类型必填
    ParentID    string `json:"parent_id"`
    Description string `json:"description"`
    Sort        int    `json:"sort"`
}

type AssignPermissionsRequest struct {
    PermissionIDs []string `json:"permission_ids" binding:"required"`
}

type PermissionInfo struct {
    PermissionID string            `json:"permission_id"`
    Name         string            `json:"name"`
    Type         string            `json:"type"`
    Resource     string            `json:"resource"`
    Action       string            `json:"action"`
    ParentID     string            `json:"parent_id"`
    Sort         int               `json:"sort"`
    Children     []*PermissionInfo `json:"children,omitempty"`
}
```

## 验收标准

```bash
# 1. 获取权限树
curl -s http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer $TOKEN_ADMIN"
# 期望：返回权限树形结构

# 2. 创建权限
curl -s -X POST http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"name":"用户列表","type":"API","resource":"/api/v1/users","action":"GET"}'
# 期望：{"code":0,"data":{"permission_id":"..."}}

# 3. 分配权限给角色
curl -s -X POST http://localhost:8080/api/v1/roles/ROLE_ID/permissions \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"permission_ids":["p1","p2"]}'
# 期望：{"code":0,"message":"success"}

# 4. 分配后权限生效
# 给角色分配 GET:/api/v1/users 权限后，该角色用户可以访问该接口

# 5. 非超管无法管理权限
curl -s -X POST http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer $TOKEN_USER"
# 期望：{"code":40300,"message":"无操作权限"}
```

## AI 协作提示

```
请按 step-08-permission-management.md 实现权限管理接口。

要点：
1. handler/permission/ — CRUD + 树形查询
2. service/permission/ — 创建/更新/删除 + 角色分配权限
3. 权限是全局表（无 tenant_id），只有 super_admin 能管理
4. 角色分配权限：全量替换（事务内先删后插）
5. 分配后调用 rbacCache.NotifyRefresh()
6. 权限树按 parent_id 递归构建
7. 路由注册到 auth 组，RBAC 中间件会检查权限
```

---

*上一步：[Step 07 - RBAC 核心](step-07-rbac-permission-cache.md) | 下一步：[Step 09 - 租户管理](step-09-domain-tenant.md)*
