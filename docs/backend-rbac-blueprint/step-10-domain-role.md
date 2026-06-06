# Step 10: 域 — 角色管理 + 权限分配

## 目标

实现角色管理 CRUD + 权限分配 + 角色继承。权限变更后触发 PermissionCache 刷新。

## 前置条件

- Step 07 RBAC 核心完成
- Step 08, 09 的域模式已建立

## 文件清单

```
internal/
├── handler/role/
│   ├── role.go                # Handler + RegisterRoutes
│   ├── create.go              # POST /roles
│   ├── update.go              # PUT /roles/:id
│   ├── delete.go              # DELETE /roles/:id
│   ├── query.go               # GET /roles + GET /roles/:id
│   └── permission.go          # POST /roles/:id/permissions + GET /roles/:id/permissions
├── service/role/
│   ├── role.go                # Service struct（持有 rbacCache）
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── query.go
│   ├── permission.go          # 权限分配逻辑 + NotifyRefresh
│   └── converter.go           # ModelToRoleInfo（exported，跨域引用）
├── repository/
│   ├── role_repo.go           # 角色查询
│   ├── role_permission_repo.go # 角色-权限关联
│   └── permission_repo.go     # 权限查询
└── dto/
    └── role_dto.go
```

## 实现细节

### 1. 角色创建（含继承）

```
POST /roles
Body: { "role_code": "dept_manager", "name": "部门经理", "parent_role_code": "admin" }

Service 逻辑：
1. 从 xcontext 获取 tenant_id
2. 校验 role_code 在租户内唯一
3. 如果有 parent_role_code：
   a. 查找父角色（只在 default 租户中查找 —— 模板角色）
   b. 设置 parent_role_id
4. 生成 role_id
5. 插入 roles 表
```

**设计决策**：`parent_role_code` 只允许指向 default 租户中的模板角色。这保证了继承链的可控性。

### 2. 权限分配（核心！）

```
POST /roles/:id/permissions
Body: {
  "permissions": [
    { "permission_id": "perm_001", "type": "MENU" },
    { "permission_id": "perm_002", "type": "BUTTON" }
  ]
}

Service 逻辑（必须在事务中）：
1. 验证角色属于当前租户
2. 删除旧的 role_permissions（WHERE role_id = ? AND tenant_id = ?）
3. 批量插入新的 role_permissions
4. 调用 rbacCache.NotifyRefresh()  ← 关键：即时刷新缓存
5. 事务提交
```

### 3. 权限查询

```
GET /roles/:id/permissions?type=MENU

Service 逻辑：
1. 查询 role_permissions WHERE role_id = ? AND tenant_id = ?
2. JOIN permissions 获取详情
3. 如果角色有 parent_role_id，递归获取继承的权限
4. 合并自有权限和继承权限，去重
5. 按类型过滤返回
```

### 4. 角色删除保护

```
DELETE /roles/:id

Service 逻辑：
1. 检查是否有用户使用该角色（user_roles）
2. 检查是否有子角色引用该角色（roles WHERE parent_role_id = ?）
3. 如果有 → 返回 409
4. 软删除角色
5. 调用 rbacCache.NotifyRefresh()
```

### 5. 角色列表

```
GET /roles?page=1&page_size=10&keyword=xxx

返回：
  - 角色基本信息
  - 角色权限数量（MENU/BUTTON/API 各多少）
  - 角色用户数量
  - 父角色信息（如果有）
```

### 6. Converter（跨域 exported）

```go
// service/role/converter.go

// ModelToRoleInfo — exported，其他域（如 user）可以引用
func ModelToRoleInfo(role *model.Role) *dto.RoleInfo {
    return &dto.RoleInfo{
        RoleID:   role.RoleID,
        RoleCode: role.RoleCode,
        Name:     role.Name,
        // ...
    }
}

// ModelsToRoleInfos — 批量转换
func ModelsToRoleInfos(roles []*model.Role) []*dto.RoleInfo { ... }
```

**其他域引用方式**：

```go
import roleconv "admin/backend-rbac/internal/service/role"
```

### 7. 权限管理 API（超管专用）

```
GET /permissions?type=API    → 获取所有 API 权限（全局表）
GET /permissions?type=MENU   → 获取所有菜单权限
GET /permissions?type=BUTTON → 获取所有按钮权限
```

> permissions 是全局表，不需要租户隔离。所有租户共享同一套权限定义。

## 验收标准

- [ ] `POST /roles` 创建角色成功
- [ ] `POST /roles` role_code 重复 → 返回 409
- [ ] `POST /roles` 带 parent_role_code → 角色继承设置成功
- [ ] `POST /roles/:id/permissions` 权限分配成功
- [ ] 权限分配后 `GET /roles/:id/permissions` 返回新权限
- [ ] 权限分配后 RBAC 缓存立即刷新（新权限即时生效）
- [ ] `DELETE /roles/:id` 角色有用户 → 返回 409
- [ ] `DELETE /roles/:id` 角色有子角色 → 返回 409
- [ ] `DELETE /roles/:id` 成功删除 + 缓存刷新
- [ ] 角色列表包含权限数量统计

## AI 协作提示

```
请按 step-10-domain-role.md 实现角色管理。
关键点：
1. 权限分配后必须调用 rbacCache.NotifyRefresh()
2. 角色删除前检查用户引用和子角色引用
3. parent_role_code 只查找 default 租户的模板角色
4. Converter 必须是 exported（ModelToRoleInfo），其他域要引用
5. 权限分配在事务中执行
```
