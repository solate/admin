# 组织结构系统设计

## 设计原则

- **租户独立管理**：每个租户拥有独立的组织架构树，互不干扰
- **部门岗位分离**：部门按职能划分，岗位按职责定义，职责清晰
- **标准化岗位编码**：岗位编码（如 `DEPT_LEADER`）与 Casbin 角色对应，简化权限配置
- **支持一人多岗**：通过中间表支持用户兼任多个岗位

---

## 数据模型

### 部门表 (departments)

```sql
CREATE TABLE departments (
    department_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    parent_id VARCHAR(20),
    department_name VARCHAR(100) NOT NULL,
    description TEXT,
    sort INT DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    INDEX idx_tenant_parent(tenant_id, parent_id, deleted_at)
);
```

### 岗位表 (positions)

```sql
CREATE TABLE positions (
    position_id VARCHAR(20) PRIMARY KEY,
    tenant_id VARCHAR(20) NOT NULL,
    position_code VARCHAR(50) NOT NULL,  -- DEPT_LEADER, EMPLOYEE, HR 等
    position_name VARCHAR(100) NOT NULL,
    level INT,                           -- 职级
    description TEXT,                     -- 岗位描述
    sort INT DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at BIGINT DEFAULT 0,
    UNIQUE KEY uk_tenant_code(tenant_id, position_code, deleted_at)
);
```

### 用户表扩展 (users)

```sql
ALTER TABLE users ADD COLUMN department_id VARCHAR(20);
ALTER TABLE users ADD COLUMN position_id VARCHAR(20);
```

### 用户多岗关联表 (user_positions) - 可选

**说明：支持一人多岗（兼职、代理、临时授权场景）**

- `users.position_id` = 主岗位（日常主要岗位）
- `user_positions` = 额外岗位（兼职岗位、临时授权）

**示例：** 张三主岗位是"高级工程师"，同时兼任"测试主管"

```sql
CREATE TABLE user_positions (
    user_id VARCHAR(20) NOT NULL,
    position_id VARCHAR(20) NOT NULL,
    is_primary SMALLINT DEFAULT 2,  -- 1=主岗位, 2=额外岗位
    PRIMARY KEY (user_id, position_id)
);
```

---

## Casbin 集成

### 策略设计

```conf
# 岗位权限定义（租户级别共享）
p, SUPER_ADMIN, default, *, *
p, DEPT_LEADER, default, org:mine, *
p, DEPT_LEADER, default, org:mine:children, *
p, HR, default, user:*, view
p, EMPLOYEE, default, user:mine, view

# 岗位继承（g2）
g2, SENIOR_LEADER, DEPT_LEADER
g2, MANAGER, DEPT_LEADER

# 用户绑定岗位（g）
g, user-001, DEPT_LEADER, tenant-a
g, user-002, EMPLOYEE, tenant-a
```

### 权限说明

| 岗位 | 数据权限范围 |
|-----|-------------|
| SUPER_ADMIN | 全部数据 |
| DEPT_LEADER | 本部门 + 子部门 |
| HR | 全部用户（只读） |
| EMPLOYEE | 自己的数据 |

---

## 业务逻辑

### 1. 创建部门

```go
func (s *DeptService) Create(ctx context.Context, req *CreateDeptRequest) error {
    tenantID := getTenantID(ctx)

    // 验证父部门
    if req.ParentID != nil {
        parent, _ := s.deptRepo.GetByID(ctx, *req.ParentID)
        if parent == nil || parent.TenantID != tenantID {
            return errors.New("父部门不存在")
        }
    }

    dept := &Department{
        DepartmentID:   uuid.New().String(),
        TenantID:       tenantID,
        ParentID:       req.ParentID,
        DepartmentName: req.DepartmentName,
        Status:         1,
    }
    return s.deptRepo.Create(ctx, dept)
}
```

### 2. 创建岗位

```go
func (s *PositionService) Create(ctx context.Context, req *CreatePositionRequest) error {
    tenantID := getTenantID(ctx)

    // 检查编码唯一性
    existing, _ := s.positionRepo.GetByCode(ctx, tenantID, req.PositionCode)
    if existing != nil {
        return errors.New("岗位编码已存在")
    }

    position := &Position{
        PositionID:   uuid.New().String(),
        TenantID:     tenantID,
        PositionCode: req.PositionCode,  // DEPT_LEADER
        PositionName: req.PositionName,  // 部门组长
        Level:        req.Level,
        Status:       1,
    }

    // 创建岗位时同步创建 Casbin 角色权限（如果不存在）
    s.syncCasbinRole(position.PositionCode)

    return s.positionRepo.Create(ctx, position)
}
```

### 3. 用户数据权限查询（核心）

```go
func (s *UserService) ListUsers(ctx context.Context) ([]*User, error) {
    user := getUserFromContext(ctx)

    // 获取用户的所有岗位（包括继承的）
    positions := s.getPositionCodesForUser(ctx, user.ID)

    // 按优先级检查权限
    if contains(positions, "SUPER_ADMIN") {
        return s.userRepo.ListAll(ctx)
    }

    if contains(positions, "DEPT_LEADER") || contains(positions, "MANAGER") {
        return s.userRepo.ListByDeptWithChildren(ctx, user.DepartmentID)
    }

    if contains(positions, "HR") {
        return s.userRepo.ListAll(ctx)
    }

    // 默认：只查自己
    return s.userRepo.ListByIDs(ctx, []string{user.ID})
}

// 获取用户的所有岗位代码（包括 g2 继承）
func (s *UserService) getPositionCodesForUser(ctx context.Context, userID string) []string {
    tenantCode := getTenantCode(ctx)

    // 从 g 策略获取直接绑定的岗位
    directRoles := s.enforcer.GetRolesForUserInDomain(userID, tenantCode)

    // 递归获取 g2 继承的岗位
    var allRoles []string
    visited := make(map[string]bool)

    var dfs func(role string)
    dfs = func(role string) {
        if visited[role] {
            return
        }
        visited[role] = true
        allRoles = append(allRoles, role)

        // g2 继承
        parents := s.enforcer.GetRolesForUser(role)
        for _, parent := range parents {
            dfs(parent)
        }
    }

    for _, role := range directRoles {
        dfs(role)
    }

    return allRoles
}
```

### 4. 用户调岗

```go
func (s *UserService) UpdatePosition(ctx context.Context, userID, positionID string) error {
    tenantID := getTenantID(ctx)

    // 验证岗位存在
    position, _ := s.positionRepo.GetByID(ctx, positionID)
    if position == nil || position.TenantID != tenantID {
        return errors.New("岗位不存在")
    }

    user, _ := s.userRepo.GetByID(ctx, userID)
    if user == nil || user.TenantID != tenantID {
        return errors.New("用户不存在")
    }

    // 更新用户岗位
    user.PositionID = positionID
    s.userRepo.Update(ctx, user)

    // 同步更新 Casbin g 策略
    tenantCode := getTenantCode(ctx)
    s.enforcer.DeleteRoleForUserInDomain(userID, tenantCode)
    s.enforcer.AddRoleForUserInDomain(userID, position.PositionCode, tenantCode)

    return nil
}
```

### 5. 删除部门

```go
func (s *DeptService) Delete(ctx context.Context, deptID string) error {
    tenantID := getTenantID(ctx)

    dept, _ := s.deptRepo.GetByID(ctx, deptID)
    if dept == nil || dept.TenantID != tenantID {
        return errors.New("部门不存在")
    }

    // 检查子部门
    children, _ := s.deptRepo.GetChildren(ctx, deptID)
    if len(children) > 0 {
        return errors.New("存在子部门，无法删除")
    }

    // 检查关联用户
    count, _ := s.userRepo.CountByDept(ctx, deptID)
    if count > 0 {
        return errors.New("部门下存在用户，无法删除")
    }

    return s.deptRepo.Delete(ctx, deptID)
}
```

---

## API 设计

### 部门接口

```
GET    /api/v1/depts                   获取部门树
POST   /api/v1/depts                   创建部门
PUT    /api/v1/depts/:id               更新部门
DELETE /api/v1/depts/:id               删除部门
GET    /api/v1/depts/:id/children      获取子部门
```

### 岗位接口

```
GET    /api/v1/positions               获取岗位列表
POST   /api/v1/positions               创建岗位
PUT    /api/v1/positions/:id           更新岗位
DELETE /api/v1/positions/:id           删除岗位
```

### 用户接口（扩展）

```
PUT    /api/v1/users/:id/position      调整用户岗位
GET    /api/v1/users/:id/positions     获取用户多岗位
POST   /api/v1/users/:id/positions     分配额外岗位
```

---

## 初始化数据结构

### 部门树（19个部门）

```
总公司
├── 总经办
├── 技术中心
│   ├── 研发部
│   ├── 测试部
│   └── 运维部
├── 产品中心
│   ├── 产品部
│   └── 设计部
├── 运营中心
│   ├── 用户运营部
│   └── 内容运营部
├── 市场中心
├── 销售中心
│   ├── 直销部
│   └── 渠道部
├── 人力资源部
├── 财务部
└── 行政部
```

### 岗位层级（37个岗位）

| 职级 | 岗位 |
|------|------|
| L100 | CEO |
| L90 | CTO |
| L88 | COO |
| L87 | CFO |
| L85 | CPO |
| L60 | 部门经理 |
| L55 | 架构师、销售总监 |
| L52 | 运营经理、市场经理、销售经理 |
| L50 | 部门主管、产品经理、人事经理、财务经理 |
| L48 | 测试主管、行政经理 |
| L45 | 高级工程师、产品负责人、品牌经理 |
| L40 | 团队组长 |
| L38 | 工程师、测试工程师、运维工程师、UI/UX设计师、内容运营、用户运营、人事专员、招聘专员、会计 |
| L35 | 初级工程师、内容运营、用户运营、人事专员、招聘专员 |
| L32 | 行政专员 |
| L30 | 销售代表、出纳 |
| L25 | 初级工程师 |
| L10 | 员工 |
| L5 | 实习生 |

---

## Repository 层实现

### DeptRepo

```go
// 查询租户的所有部门
func (r *DeptRepo) ListByTenant(ctx context.Context, tenantID string) ([]*Department, error) {
    ctx = database.WithTenantID(ctx, tenantID)
    return r.q.Department.WithContext(ctx).
        Order(r.q.Department.Sort).
        Find()
}

// 获取子部门
func (r *DeptRepo) GetChildren(ctx context.Context, parentID string) ([]*Department, error) {
    return r.q.Department.WithContext(ctx).
        Where(r.q.Department.ParentID.Eq(parentID)).
        Find()
}

// 获取部门及所有子部门ID（递归）
func (r *DeptRepo) GetDescendantIDs(ctx context.Context, deptID string) ([]string, error) {
    var ids []string
    ids = append(ids, deptID)

    children, _ := r.GetChildren(ctx, deptID)
    for _, child := range children {
        childIDs, _ := r.GetDescendantIDs(ctx, child.DepartmentID)
        ids = append(ids, childIDs...)
    }

    return ids, nil
}
```

### UserRepo

```go
// 按部门及子部门查询用户
func (r *UserRepo) ListByDeptWithChildren(ctx context.Context, deptID string) ([]*User, error) {
    // 获取所有子部门ID
    deptIDs, _ := r.deptRepo.GetDescendantIDs(ctx, deptID)

    return r.q.User.WithContext(ctx).
        Where(r.q.User.DepartmentID.In(deptIDs...)).
        Find()
}

// 统计部门下用户数
func (r *UserRepo) CountByDept(ctx context.Context, deptID string) (int64, error) {
    return r.q.User.WithContext(ctx).
        Where(r.q.User.DepartmentID.Eq(deptID)).
        Count()
}
```

---

## 常量定义

```go
package constants

const (
    // 部门状态
    DeptStatusEnabled  = 1
    DeptStatusDisabled = 0

    // 标准岗位编码（与 Casbin 角色对应）
    PositionCodeSuperAdmin  = "SUPER_ADMIN"
    PositionCodeDeptLeader  = "DEPT_LEADER"
    PositionCodeManager     = "MANAGER"
    PositionCodeHR          = "HR"
    PositionCodeEmployee    = "EMPLOYEE"
)
```

---

## 总结

| 特性 | 实现方式 |
|------|----------|
| 租户隔离 | `tenant_id` 字段 + 自动过滤 |
| 部门树形结构 | `parent_id` 构建父子关系 |
| 标准化岗位 | `position_code` 与 Casbin 角色对应 |
| 数据权限 | 岗位编码 + Casbin 策略控制 |
| 用户调岗 | 更新 `position_id` + 同步 Casbin g 策略 |
| 一人多岗 | `user_positions` 中间表 |
