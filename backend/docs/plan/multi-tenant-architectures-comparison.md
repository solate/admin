# 多租户 SaaS 用户架构设计方案对比

> **文档目的**：记录多租户 SaaS 后台管理系统用户与租户关系的不同设计方案，便于后续设计决策参考

---

## 一、两种核心设计方案

### 方案一：用户属于单个租户（1:1 关系）

```
┌─────────────────────────────────────────────────────────────────┐
│                        租户 A                                    │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                         │
│  │ user1   │  │ user2   │  │ user3   │                         │
│  │ @a.com  │  │ @a.com  │  │ @a.com  │                         │
│  └─────────┘  └─────────┘  └─────────┘                         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        租户 B                                    │
│  ┌─────────┐  ┌─────────┐                                      │
│  │ user1   │  │ user4   │                                      │
│  │ @b.com  │  │ @b.com  │                                      │
│  └─────────┘  └─────────┘                                      │
└─────────────────────────────────────────────────────────────────┘

特点：user1@a.com 和 user1@b.com 是两个完全独立的用户
```

**数据库结构**：
```sql
-- 用户表直接绑定租户
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,  -- 直接绑定租户
    user_name VARCHAR(255) NOT NULL,  -- 租户内唯一
    password VARCHAR(255) NOT NULL,
    ...
    UNIQUE KEY uk_tenant_username (tenant_id, user_name)
);

-- 用户角色关系
CREATE TABLE user_roles (
    user_role_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    ...
);
```

---

### 方案二：用户跨多个租户（N:M 关系）

```
                        ┌─────────────────────────────────────┐
                        │           用户 (user1)               │
                        │   全局唯一：user1@global.com         │
                        └─────────┬───────────────────────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
              ▼                   ▼                   ▼
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │  租户 A         │ │  租户 B         │ │  租户 C         │
    │  角色：管理员    │ │  角色：普通用户  │ │  角色：审计员    │
    └─────────────────┘ └─────────────────┘ └─────────────────┘

特点：同一用户可在多个租户中拥有不同角色
```

**数据库结构**：
```sql
-- 用户表与租户解耦
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL UNIQUE,  -- 全局唯一
    password VARCHAR(255) NOT NULL,
    ...
    -- 注意：没有 tenant_id 字段
);

-- 用户-租户-角色关联表
CREATE TABLE user_tenant_role (
    user_tenant_role_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    ...
    UNIQUE KEY uk_user_tenant_role (user_id, tenant_id, role_id)
);
```

---

## 二、方案对比表

| 维度 | 方案一：单租户用户 | 方案二：跨租户用户 |
|------|-------------------|-------------------|
| **用户名唯一性** | 租户内唯一 | 全局唯一 |
| **数据隔离** | 完全隔离 | 共享用户数据 |
| **实现复杂度** | ⭐ 简单 | ⭐⭐⭐ 中等 |
| **多租户切换** | 需要多次登录 | 一次登录切换 |
| **用户管理** | 租户自治 | 平台统一 |
| **跨租户协作** | 不支持 | 原生支持 |
| **数据隔离安全** | ⭐⭐⭐⭐⭐ 最高 | ⭐⭐⭐ 需严格控制 |
| **用户名冲突** | 不会冲突 | 可能冲突 |
| **适用场景** | 独立部署SaaS | 平台型SaaS |

---

## 三、各方案的设计原因分析

### 方案一：单租户用户

**设计原因**：
1. **数据安全优先**：每个租户的用户数据完全物理隔离，安全风险最小
2. **简化权限模型**：不需要考虑跨租户权限问题
3. **租户自治**：租户可以完全自主管理自己的用户
4. **合规要求**：满足某些行业的数据隔离合规要求

**适用场景**：
- **垂直SaaS**：每个租户是完全独立的组织（如医院管理系统、学校管理系统）
- **数据敏感**：租户之间数据绝对不能有任何关联（如政府系统、金融系统）
- **独立部署**：有些租户可能要求独立部署

**典型产品**：
- 钉钉（企业版）
- 飞书
- 各类企业ERP系统

---

### 方案二：跨租户用户

**设计原因**：
1. **用户体验优先**：一个账号管理所有租户，无需重复登录
2. **跨租户协作**：支持顾问、管理员等角色管理多个租户
3. **平台化定位**：定位为平台而非独立SaaS，用户属于平台
4. **降低使用门槛**：减少用户记忆多个账号的负担

**适用场景**：
- **平台型SaaS**：用户是平台的用户，租户只是用户的工作空间
- **集团化管理**：集团公司需要统一管理多个子公司
- **服务商场景**：服务商需要管理多个客户（如SaaS代运维公司）
- **多组织协作**：一个人需要在多个组织中工作

**典型产品**：
- 阿里云（一个账号管理多个云账号/企业）
- 腾讯云
- Notion（个人工作区 + 团队工作区）
- Figma

---

## 四、关键问题分析

### 问题1：用户名冲突

| 方案 | 问题 | 解决方案 |
|------|------|----------|
| 方案一 | 无冲突 | 无需解决 |
| 方案二 | 全局唯一，可能被占用 | 1. 使用邮箱作为用户名<br>2. 允许用户名 + 租户前缀<br>3. 使用手机号登录 |

### 问题2：登录体验

| 方案 | 登录流程 | 体验评分 |
|------|----------|----------|
| 方案一 | 每个租户独立登录 | ⭐⭐ |
| 方案二 | 一次登录，选择租户 | ⭐⭐⭐⭐⭐ |

### 问题3：数据隔离

| 方案 | 隔离级别 | 安全性 |
|------|----------|--------|
| 方案一 | 物理隔离 | ⭐⭐⭐⭐⭐ |
| 方案二 | 逻辑隔离 | ⭐⭐⭐ |

---

## 五、选择决策树

```
是否需要跨租户用户（如顾问管理多个租户）？
    │
    ├─ 是 ──> 【方案二：跨租户用户】
    │
    └─ 否 ──> 数据安全要求是否极高（政府、金融）？
                │
                ├─ 是 ──> 【方案一：单租户用户】
                │
                └─ 否 ──> 产品定位是什么？
                            │
                            ├─ 平台型（用户属于平台）
                            │   └─> 【方案二：跨租户用户】
                            │
                            └─ 工具型（租户独立）
                                └─> 【方案一：单租户用户】
```

---

## 六、推荐方案

### 对于大多数 SaaS 后台管理系统

**推荐：方案一（单租户用户）**

**理由**：
1. ✅ 实现简单，维护成本低
2. ✅ 数据隔离安全性最高
3. ✅ 租户完全自治
4. ✅ 符合大多数SaaS场景

**何时选择方案二**：
- 产品定位是"平台"而非"SaaS"
- 需要支持跨租户角色（如超级管理员、顾问）
- 用户体验优先于数据隔离
- 类似阿里云、腾讯云的产品形态

---

## 六、超级管理员设计

### 方案一的超管模式：平台超管 + 租户超管

```
┌─────────────────────────────────────────────────────────────────┐
│                    平台超级管理员                                │
│  ┌─────────┐  ┌─────────┐                                      │
│  │ super   │  │ ops1    │                                      │
│  │ admin   │  │         │                                      │
│  │         │  │         │                                      │
│  │ (跨所有 │  │ (跨所有 │                                      │
│  │  租户)  │  │  租户)  │                                      │
│  └─────────┘  └─────────┘                                      │
└─────────────────────────────────────────────────────────────────┘
                         │
                         │ 可管理所有租户
                         ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│  租户 A     │ │  租户 B     │ │  租户 C     │
│ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │
│ │ admin   │ │ │ │ admin   │ │ │ │ admin   │ │
│ │(租户超管)│ │ │ │(租户超管)│ │ │ │(租户超管)│ │
│ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │
│             │ │             │ │             │ │
│ ┌─────────┐ │ │ ┌─────────┐ │ │ ┌─────────┐ │
│ │ user1   │ │ │ │ user3   │ │ │ │ user5   │ │
│ │ (普通)  │ │ │ │ (普通)  │ │ │ │ (普通)  │ │
│ └─────────┘ │ │ └─────────┘ │ │ └─────────┘ │
└─────────────┘ └─────────────┘ └─────────────┘

特点：
- 平台超管可以管理所有租户
- 租户超管只能管理本租户
```

#### 数据库设计

```sql
-- 用户表
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36),               -- 平台超管为 NULL，租户用户有值
    user_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    ...
    UNIQUE KEY uk_tenant_username (tenant_id, user_name)
);

-- 用户名唯一性说明：
-- - 平台超管：(NULL, 'super_admin') 全局唯一
-- - 租户用户：('tenant-a', 'admin') 租户内唯一
-- - 平台超管和租户用户可以有相同的用户名，因为 NULL != 'tenant-a'

-- 角色表
CREATE TABLE roles (
    role_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36),               -- NULL 表示平台角色，有值表示租户角色
    role_name VARCHAR(255) NOT NULL,
    ...
);

-- 用户角色关系
CREATE TABLE user_roles (
    user_role_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    ...
);
```

#### 权限判断逻辑

```go
// 权限检查
func CheckPermission(user *User, targetTenantID string, operation string) bool {
    // 平台超管：tenant_id 为空，可以管理所有租户
    if user.TenantID == nil || user.TenantID == "" {
        return true
    }

    // 租户超管：只能管理本租户
    if user.IsTenantAdmin {
        return user.TenantID == targetTenantID
    }

    // 普通用户：根据角色判断
    return checkRolePermission(user, targetTenantID, operation)
}
```

**关键点**：
- 平台超管：`tenant_id` 为 NULL，可管理所有租户
- 租户超管：`tenant_id` 有值，只能管理本租户
- 通过 `tenant_id` 是否为 NULL 区分，无需额外字段

#### 数据示例

```sql
-- 平台超管（可多个，tenant_id 为 NULL）
INSERT INTO users (user_id, tenant_id, user_name) VALUES ('u1', NULL, 'super_admin');
INSERT INTO users (user_id, tenant_id, user_name) VALUES ('u2', NULL, 'ops_user1');
INSERT INTO users (user_id, tenant_id, user_name) VALUES ('u3', NULL, 'ops_user2');

-- 租户用户（tenant_id 有值）
INSERT INTO users (user_id, tenant_id, user_name) VALUES ('u4', 'tenant-a', 'admin');
INSERT INTO users (user_id, tenant_id, user_name) VALUES ('u5', 'tenant-a', 'user1');
INSERT INTO users (user_id, tenant_id, user_name) VALUES ('u6', 'tenant-b', 'admin');
INSERT INTO users (user_id, tenant_id, user_name) VALUES ('u7', 'tenant-b', 'user1');
```

**用户名冲突示例**：

| 组合 | 是否冲突 | 说明 |
|------|----------|------|
| `(NULL, 'user1')` 和 `('tenant-a', 'user1')` | ✅ 不冲突 | NULL 和具体值不同 |
| `('tenant-a', 'admin')` 和 `('tenant-b', 'admin')` | ✅ 不冲突 | 不同租户 |
| `(NULL, 'super')` 和 `(NULL, 'super')` | ❌ 冲突 | 平台超管用户名全局唯一 |
| `('tenant-a', 'admin')` 和 `('tenant-a', 'admin')` | ❌ 冲突 | 同租户内重复 |

#### 角色初始化

每个租户创建时自动初始化：

```go
// 创建租户时
func CreateTenant(tenant *Tenant) error {
    // 1. 创建租户
    // 2. 为该租户创建默认角色（租户管理员、普通用户等）
    // 3. 创建该租户的 admin 用户（tenant_id = 新租户ID）
    // 4. 将 admin 用户与租户管理员角色关联
}
```

**初始化数据示例**：

| 租户 | 默认创建的用户 | 角色 |
|------|---------------|------|
| 租户A | admin | 租户管理员（只能管理租户A） |
| 租户B | admin | 租户管理员（只能管理租户B） |

平台超管需要单独创建，通常是系统初始化时创建。

---

## 七、实施建议

### 如果选择方案一（单租户用户）

```sql
-- 用户直接绑定租户
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    ...
    UNIQUE KEY uk_tenant_username (tenant_id, user_name),
    KEY idx_tenant_id (tenant_id)
);

-- 用户在租户内的角色
CREATE TABLE user_roles (
    user_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    ...
);
```

**优点**：
- 简单清晰
- 查询效率高
- 数据隔离彻底

**登录流程**：
```
用户输入 用户名+密码 → 验证 → 返回该租户的 token
```

### 如果选择方案二（跨租户用户）

```sql
-- 用户表独立于租户
CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL UNIQUE,  -- 全局唯一
    password VARCHAR(255) NOT NULL,
    ...
);

-- 用户-租户-角色关联
CREATE TABLE user_tenant_role (
    user_tenant_role_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    ...
    UNIQUE KEY uk_user_tenant_role (user_id, tenant_id, role_id)
);
```

**登录流程**：
```
用户输入 用户名+密码
    ↓
查询用户关联的所有租户
    ↓
┌─────────────┬─────────────────┬─────────────────┐
│ 只有一个租户 │ 多租户+上次选择  │ 多租户+无记录    │
└─────────────┴─────────────────┴─────────────────┘
     │              │               │
     ▼              ▼               ▼
  直接登录      自动进入上次      显示选择界面
```

---

## 八、总结

| 方案 | 推荐指数 | 核心特点 |
|------|----------|----------|
| 方案一 | ⭐⭐⭐⭐⭐ | 简单、安全、独立，适合大多数SaaS |
| 方案二 | ⭐⭐⭐⭐ | 用户体验好，适合平台型产品 |

**最终建议**：除非你的产品明确是"平台型"且有跨租户用户需求，否则优先选择**方案一（单租户用户）**。
