# 数据权限实现方案

## 一、表设计

### 1. 角色表扩展（roles）

在角色表中添加数据权限字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| tenant_id | bigint | 租户ID |
| name | string | 角色名称 |
| data_scope | smallint | 数据权限范围（1:全部 2:自定义 3:本部门 4:本部门及以下 5:仅本人） |
| data_scope_dept_ids | jsonb | 自定义部门列表（data_scope=2时使用，JSON 数组） |
| status | smallint | 状态（1:启用 2:禁用） |
| sort_order | int | 排序 |
| remark | string | 备注 |
| ... | ... | 其他字段 |

**索引**：
```sql
CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE INDEX idx_roles_tenant_data_scope ON roles(tenant_id, data_scope);
-- 可选：支持反向查询（查询哪些角色包含某个部门）
CREATE INDEX idx_roles_dept_ids ON roles USING GIN (data_scope_dept_ids);
```

---

### 2. 代码常量定义

数据权限范围使用代码常量定义，不依赖数据库字典表：

```go
// 数据权限范围常量
const (
    DataScopeAll         = 1  // 全部数据
    DataScopeCustom      = 2  // 自定义部门
    DataScopeDept        = 3  // 本部门数据
    DataScopeDeptAndSub  = 4  // 本部门及下级
    DataScopeSelf        = 5  // 仅本人数据
)

// 数据权限范围映射
var DataScopeMap = map[int32]string{
    1: "全部数据",
    2: "自定义部门",
    3: "本部门数据",
    4: "本部门及下级",
    5: "仅本人数据",
}
```

---

### 3. 业务表设计原则

所有需要数据权限控制的业务表必须包含：

| 字段 | 类型 | 说明 |
|------|------|------|
| tenant_id | bigint | 租户ID（必须） |
| dept_id | bigint | 部门ID（可选，用于部门级过滤） |
| user_id | bigint | 创建人ID（可选，用于本人数据过滤） |

**示例表（订单 orders）**：
- id
- tenant_id（租户隔离）
- dept_id（所属部门）
- created_by（创建人ID）
- order_no
- amount
- ...

---

## 二、业务流程

### 完整链路

```
1. 用户登录 → JWT Token 生成
   ↓
2. 每次请求 → JWT 中间件解析
   ↓
3. 提取 user_id、tenant_id 和 role_ids
   ↓
4. 查询角色的 data_scope 和 data_scope_dept_ids
   ↓
5. 根据 data_scope 类型，构建数据过滤条件
   ↓
6. GORM 查询时自动注入 WHERE 条件
   ↓
7. 返回符合权限的数据
```

---

## 三、核心实现逻辑

### 步骤 1：JWT Token 结构

登录时在 JWT Payload 中存储：

```json
{
  "user_id": "123456789012345678",
  "tenant_id": "987654321098765432",
  "username": "admin",
  "role_ids": ["1", "2"],
  "dept_id": "111"
}
```

---

### 步骤 2：中间件提取权限上下文

**JWT 中间件职责**：
1. 解析 Token，提取 user_id、tenant_id、role_ids、dept_id
2. 查询数据库，获取角色的 data_scope 和 data_scope_dept_ids（JSONB）
3. 解析 data_scope_dept_ids，获取自定义部门 ID 列表
4. 将权限信息存入上下文（Context）

**上下文结构**：
```go
type UserDataScope struct {
    UserID           string   // 当前用户ID
    TenantID         string   // 租户ID
    DeptID           string   // 用户所属部门ID
    DataScope        int      // 数据权限范围（1/2/3/4/5）
    CustomDeptIDs    []int64  // 自定义部门ID列表（仅 data_scope=2 时有值）
}
```

---

### 步骤 3：五种权限范围的 SQL 过滤逻辑

#### 1. 全部数据（all）

**WHERE 条件**：
```sql
WHERE tenant_id = ?
```

**说明**：只能看到本租户数据，无部门限制

---

#### 2. 自定义部门（custom）

**WHERE 条件**：
```sql
WHERE tenant_id = ?
  AND dept_id IN (?, ?, ?)  -- 从 roles.data_scope_dept_ids 解析
```

**说明**：
- 只能看到指定部门的数据
- data_scope_dept_ids 字段存储 JSON 数组：`[1, 2, 3]`
- 在中间件阶段解析 JSONB，构建部门 ID 列表
- 可选：使用 PostgreSQL GIN 索引优化反向查询

---

#### 3. 本部门数据（dept）

**WHERE 条件**：
```sql
WHERE tenant_id = ?
  AND dept_id = ?  -- 当前用户所在部门
```

**说明**：只能看到本部门数据

---

#### 4. 本部门及下级（dept_and_sub）

**WHERE 条件**：
```sql
WHERE tenant_id = ?
  AND dept_id IN (
    ?,  -- 当前部门
    ?,  -- 子部门1
    ?   -- 子部门2
  )
```

**说明**：
- 需要先查询部门树，获取所有子部门 ID
- 递归查询或维护 path 字段（如：1/2/3）

**部门表设计支持**：
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 部门ID |
| parent_id | bigint | 父部门ID |
| tree_path | string | 树路径（如：1/2/3） |
| level | int | 层级 |

**查询子部门逻辑**：
```
1. 查询当前用户的 dept_id
2. 查询部门树，获取所有子部门
3. 构建部门 ID 列表
4. SQL 使用 IN (dept_id_list)
```

---

#### 5. 仅本人数据（self）

**WHERE 条件**：
```sql
WHERE tenant_id = ?
  AND created_by = ?  -- 当前用户ID
```

**说明**：只能看到自己创建的数据

---

## 四、GORM 实现方案

### 方案 1：查询作用域（Query Scope）

**实现逻辑**：
1. 定义全局 Scope 函数
2. 从上下文读取权限信息
3. 动态构建 WHERE 条件
4. 每次查询自动应用 Scope

**使用方式**：
```go
// Repository 层查询
query.Users.WithContext(ctx).
    Scopes(DataScope()).
    Find(&users)
```

---

### 方案 2：GORM Hook（推荐）

**实现逻辑**：
1. 注册全局 Query Hook
2. 拦截所有查询语句
3. 自动注入 tenant_id 和数据权限条件
4. 对业务透明，无需手动调用

**优点**：
- 自动化，防止遗漏
- 统一管理，易于维护
- 安全性高，无法绕过

---

### 方案 3：中间件注入（结合方案 2）

**实现逻辑**：
1. 在中间件阶段预计算权限条件
2. 将过滤条件存入 Context
3. GORM Hook 从 Context 读取并注入

**优点**：
- 查询时无需重复计算
- 性能更好
- 权限逻辑集中管理

---

## 五、多角色权限合并

### 问题：用户有多个角色，权限如何处理？

**策略**：取最大权限（宽松策略）

**示例**：
- 角色 A：仅本人（self）
- 角色 B：本部门（dept）
- **最终权限**：本部门（dept）—— 范围更大

**合并规则**：
```
all > custom > dept_and_sub > dept > self
```

**实现逻辑**：
1. 查询用户所有角色的 data_scope 和 data_scope_dept_ids（JSONB）
2. 按优先级排序，取最高级别（值越小权限越大）
3. 如果多个 custom 角色，合并部门列表（JSON 数组合并去重）

**示例代码逻辑**：
```go
// 假设用户有两个角色
role1: data_scope=3, data_scope_dept_ids=[1,2]
role2: data_scope=2, data_scope_dept_ids=[3,4]

// 最终权限：取最大权限
data_scope = 2  (custom)
custom_dept_ids = [1,2,3,4]  (合并去重)
```

---

## 六、缓存策略

### 缓存角色权限信息

**为什么需要缓存**：
- 每次请求都查询数据库影响性能
- 角色权限变更频率低

**缓存内容**：
```go
Key: role:permission:{role_id}
Value: {
  data_scope: 3,              // int 类型
  custom_dept_ids: [1,2,3]   // JSON 数组
}
TTL: 1 小时
```

**JSONB 序列化示例**：
```go
// 查询时：数据库 → 缓存
role := &Role{
    DataScope: 2,
    DataScopeDeptIDs: datatypes.JSON([]byte("[1,2,3]")),
}
cache.Set("role:permission:1", role, 1*time.Hour)

// 读取时：缓存 → 上下文
customDeptIDs := []int64{}
json.Unmarshal(role.DataScopeDeptIDs, &customDeptIDs)
```

**缓存失效**：
- 修改角色权限时删除缓存
- 用户登录时预热缓存

---

## 七、特殊场景处理

### 1. 超级管理员

**处理方式**：
- 设置特殊角色标识（is_superuser: true）
- 跳过数据权限检查
- 可看到所有租户数据（如需要）

---

### 2. 系统管理员

**处理方式**：
- data_scope = "all"
- 只能看本租户所有数据
- 不受部门限制

---

### 3. 跨部门协作

**场景**：需要临时访问其他部门数据

**解决方案**：
- 使用 custom（自定义部门）
- 为角色配置可访问的多个部门
- 支持临时授权机制

---

### 4. 委派/代理

**场景**：请假期间，工作由他人代管

**解决方案**：
- 添加数据权限委托表
- 记录委托关系和有效期
- 查询时合并委托人和代理人的权限

---

## 八、性能优化

### 1. 部门树查询优化

**问题**：每次查询子部门都很慢

**方案 1：维护 tree_path 字段**
```
dept_id=1, tree_path="1"
dept_id=2, tree_path="1/2"
dept_id=3, tree_path="1/2/3"

-- 查询子部门
WHERE tree_path LIKE '1/%'
```

**方案 2：使用 Closure Table**
- 维护部门关系表
- 存储所有祖先-后代关系
- 查询速度快，更新略慢

**方案 3：物化路径缓存**
- Redis 缓存部门树
- 定时刷新
- 查询优先读缓存

---

### 2. 索引优化

**必建索引**：
- tenant_id（所有表）
- dept_id（业务表）
- created_by（业务表）
- (tenant_id, dept_id) 联合索引
- (tenant_id, created_by) 联合索引

---

### 3. 查询优化

**避免全表扫描**：
- 必须带 tenant_id 条件
- data_scope 避免使用子查询
- custom 部门列表限制数量（建议最多 50 个）
- 使用 PostgreSQL JSONB 而非 JSON（性能更好）

**JSONB 查询性能**：
```sql
-- 好的：中间件阶段解析后传参
WHERE dept_id IN (1, 2, 3)  -- ✅ 使用索引

-- 避免：直接在 SQL 中解析 JSONB
WHERE dept_id IN (SELECT jsonb_array_elements_text(data_scope_dept_ids)::int)  -- ❌ 慢
```

---

## 九、安全性考虑

### 1. 防止绕过

**风险**：直接写 SQL 绕过 ORM

**措施**：
- 代码审查：禁止原生 SQL
- 数据库审计：记录异常查询
- RLS 兜底：启用 PostgreSQL Row Level Security

---

### 2. 防止越权

**风险**：手动修改请求参数伪造权限

**措施**：
- 权限信息从 JWT 读取，不从请求参数获取
- 服务端验证，禁用客户端传 data_scope
- 审计日志：记录权限变更和访问

---

### 3. 防止数据泄露

**风险**：错误配置导致跨租户访问

**措施**：
- 强制 tenant_id 过滤（所有查询）
- 单元测试：覆盖所有数据权限场景
- 集成测试：模拟多租户隔离

---

## 十、实现检查清单

### 数据库层
- [ ] 所有表添加 tenant_id 字段
- [ ] 业务表添加 dept_id、created_by 字段
- [ ] roles 表添加 data_scope 字段（SMALLINT）
- [ ] roles 表添加 data_scope_dept_ids 字段（JSONB）
- [ ] 建立索引：idx_roles_tenant、idx_roles_tenant_data_scope
- [ ] 可选：建立 GIN 索引支持反向查询（idx_roles_dept_ids）
- [ ] 业务表建立联合索引：(tenant_id, dept_id)、(tenant_id, created_by)

### 业务层
- [ ] 定义数据权限常量（DataScopeAll = 1 等）
- [ ] JWT 存储 role_ids
- [ ] 中间件提取权限上下文（解析 JSONB）
- [ ] 实现五级权限过滤逻辑（1/2/3/4/5）
- [ ] JSONB 序列化/反序列化工具函数
- [ ] GORM Hook 自动注入 WHERE 条件
- [ ] 多角色权限合并逻辑（取最大权限 + 合并部门列表）
- [ ] 部门树查询工具（支持 dept_and_sub 场景）

### 缓存层
- [ ] 角色权限缓存
- [ ] 部门树缓存
- [ ] 缓存失效策略

### 安全层
- [ ] 审计日志
- [ ] 异常查询告警
- [ ] 单元测试覆盖
- [ ] 可选 RLS 兜底

---

## 十一、示例流程

### 查询订单列表

```
1. 用户请求 GET /api/orders
   Header: Authorization: Bearer <token>

2. JWT 中间件解析
   user_id: "123"
   role_ids: ["5"]
   tenant_id: "1"
   dept_id: "10"

3. 查询角色权限（缓存/数据库）
   role_id=5 → data_scope=4 (dept_and_sub)

4. 查询部门树（如果需要）
   dept_id=10 的子部门：[10, 11, 12, 13]

5. 构建查询条件
   WHERE tenant_id = 1
     AND dept_id IN (10, 11, 12, 13)

6. GORM Hook 自动注入条件
   db.Where("tenant_id = ? AND dept_id IN (?)", 1, []int{10,11,12,13})

7. 执行查询返回结果
```

---

### 自定义部门权限示例（custom）

```
1. 用户有角色：role_id=3
   data_scope=2 (custom)
   data_scope_dept_ids=[1,2,5]  (JSONB)

2. 中间件解析 JSONB
   custom_dept_ids = []int64{1, 2, 5}

3. 构建查询条件
   WHERE tenant_id = 1
     AND dept_id IN (1, 2, 5)

4. 执行查询返回结果
```

---

## 总结

**核心要点**：
1. **表设计**：roles.data_scope（SMALLINT）+ roles.data_scope_dept_ids（JSONB）
2. **代码常量**：定义 DataScopeAll=1, DataScopeCustom=2 等常量
3. **中间件**：JWT → 角色 → 解析 JSONB → 权限范围 → 存入 Context
4. **查询注入**：GORM Hook 自动添加 WHERE 条件
5. **五种范围**：1(全部) 2(自定义) 3(本部门) 4(本部门及下级) 5(仅本人)
6. **JSONB 优势**：简单高效，支持 GIN 索引，PostgreSQL 原生支持

**推荐实现顺序**：
1. ✅ 先实现 tenant_id 隔离（必须，已完成）
2. 🔨 实现代码常量和表结构变更（data_scope 字段）
3. 🔨 实现 dept 和 self 级别（最常用，80% 场景）
4. 🔨 实现 custom 级别（JSONB 解析）
5. 🔨 最后实现 dept_and_sub（需要部门树支持）

**关键技术点**：
- JSONB 字段存储自定义部门列表
- 中间件阶段解析 JSONB，避免 SQL 解析
- 多角色权限取最大范围
- 可选 GIN 索引支持反向查询
