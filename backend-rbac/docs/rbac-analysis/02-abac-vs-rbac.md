# ABAC vs RBAC：多租户场景下的全面对比

> 权限控制不是选 "哪种模型更好"，而是选 "哪种模型更适合你的场景"。

---

## 1. RBAC — 基于角色的访问控制

### 核心思想

```
用户 → 角色 → 权限 → 允许/拒绝
```

一个用户拥有若干角色，每个角色拥有若干权限。权限检查就是问：**"这个用户的角色中，有哪个角色拥有这个权限吗？"**

### 数据模型

```
┌────────┐     ┌───────────┐     ┌────────┐     ┌──────────────────┐     ┌────────────┐
│  User  │────▶│ user_roles │────▶│  Role  │────▶│ role_permissions │────▶│ Permission │
└────────┘     └───────────┘     └────────┘     └──────────────────┘     └────────────┘
```

### 示例

```sql
-- 用户 Alice 在租户 A 拥有 admin 角色
INSERT INTO user_roles (user_id, role_id, tenant_id) VALUES ('alice', 'role_admin', 'tenant_a');

-- admin 角色拥有用户管理权限
INSERT INTO role_permissions (role_id, permission_id, tenant_id) VALUES ('role_admin', 'perm_user_manage', 'tenant_a');

-- 权限定义
INSERT INTO permissions (permission_id, resource, action) VALUES ('perm_user_manage', '/api/v1/users', 'POST');
```

### 优势

| 优势 | 说明 |
|------|------|
| **简单直观** | "Admin 能做什么" → 查表即知 |
| **易于管理** | 管理员在 UI 上勾选即可 |
| **行业标准** | 几乎所有框架都原生支持 |
| **审计友好** | "谁拥有什么角色" 一目了然 |
| **性能可控** | 内存缓存角色权限集，O(1) 查找 |

### 局限

| 局限 | 说明 |
|------|------|
| **粗粒度** | 只能按角色划分，不能按条件细分 |
| **角色爆炸** | 组合过多时角色数量失控 |
| **无法表达上下文** | "仅在工作时间可访问" 做不到 |
| **资源级控制弱** | "只能编辑自己创建的文档" 需要额外逻辑 |

---

## 2. ABAC — 基于属性的访问控制

### 核心思想

```
主体属性 + 资源属性 + 环境属性 + 操作 → 策略评估 → 允许/拒绝
```

不依赖固定的"角色"概念，而是基于 **属性** 动态评估。

### 属性分类

```
┌─────────────────────────────────────────────────────────┐
│                    策略评估请求                           │
├─────────────┬──────────────┬─────────────┬──────────────┤
│  主体属性    │  资源属性     │  环境属性    │  操作属性     │
│  (Subject)  │  (Resource)  │ (Environment)│  (Action)   │
├─────────────┼──────────────┼─────────────┼──────────────┤
│ 用户 ID     │ 租户 ID      │ 当前时间     │ 读/写/删除   │
│ 角色        │ 资源类型     │ IP 地址      │              │
│ 部门        │ 资源状态     │ 设备类型     │              │
│ 地区        │ 创建者 ID    │              │              │
│ 职级        │ 金额         │              │              │
└─────────────┴──────────────┴─────────────┴──────────────┘
```

### 示例策略

```yaml
# 策略：只有草稿状态的报销单，且金额 < 10000，且是本人或直属上级，才能编辑
- name: "edit-expense-draft"
  effect: ALLOW
  condition:
    - resource.status == "draft"
    - resource.amount < 10000
    - subject.id == resource.owner_id OR subject.id IN resource.owner.managers
  action: ["edit"]
```

### 优势

| 优势 | 说明 |
|------|------|
| **细粒度** | 可以表达任意复杂的条件 |
| **上下文感知** | 时间、地点、设备状态都能参与决策 |
| **减少角色数量** | 用属性替代角色爆炸 |
| **动态决策** | 每次请求实时评估，不需要预分配 |

### 局限

| 局限 | 说明 |
|------|------|
| **复杂度高** | 策略设计、测试、调试都更难 |
| **性能开销** | 每次请求都要收集属性并评估 |
| **难以审计** | "为什么 Alice 能访问这个？" 需要回放完整评估链 |
| **管理门槛高** | 非技术人员很难理解策略规则 |
| **工具链不成熟** | 不像 RBAC 有成熟的框架和 UI |

---

## 3. ReBAC — 基于关系的访问控制

### 核心思想

```
用户 → 关系 → 资源 → 允许/拒绝
```

源自 Google Zanzibar 论文，核心是通过 **关系图** 推导权限。

### 示例

```
document: quarterly-report#owner@alice          → Alice 是报告的 owner
document: quarterly-report#viewer@bob           → Bob 是报告的 viewer
folder: finance#parent@document:quarterly-report → 报告在 finance 文件夹中
folder: finance#viewer@charlie                  → Charlie 是文件夹的 viewer

→ 推导：Charlie 可以查看 quarterly-report（通过文件夹关系继承）
→ 推导：Alice 可以查看 quarterly-report（owner 隐含 viewer）
```

### 适用场景

- 文档/文件夹的层级权限
- GitHub 风格的 "仓库 → 团队 → 组织" 权限继承
- Google Drive 风格的共享权限
- 需要 **图遍历** 递归推导的权限模型

### 代表实现

- [SpiceDB](https://authzed.com/)（开源 Zanzibar 实现）
- Google Zanzibar（内部系统）
- Auth0 FGA（基于 Zanzibar）

---

## 4. 多租户场景对比

### 核心问题：每种模型怎么处理租户隔离？

| 维度 | RBAC | ABAC | ReBAC |
|------|------|------|-------|
| **租户隔离** | `WHERE tenant_id = ?` 查询过滤 | 策略中检查 `resource.tenant_id == subject.tenant_id` | 关系图中 tenant 作为 namespace |
| **跨租户超管** | `HasRole("super_admin")` 跳过检查 | 策略中加条件 `subject.is_super_admin` | 全局 namespace 中的关系 |
| **租户自定义角色** | 租户内创建角色 + 权限 | 租户自定义策略 | 租户自定义关系 |
| **角色继承** | `parent_role_id` + 递归 CTE | 策略中定义继承规则 | 关系传递性天然支持 |
| **数据量** | 中等（角色×权限矩阵） | 大（每个资源可能有不同属性） | 极大（关系图可能非常复杂） |
| **性能** | 内存缓存，O(1) | 每次评估，O(策略数) | 图遍历，O(深度) |

### 实际案例对比

#### 场景 1：SaaS 后台管理系统（本项目）

```
需求：用户→角色→菜单/按钮/API权限，多租户隔离，超管跨租户
```

| 模型 | 适配度 | 原因 |
|------|--------|------|
| **RBAC** | ⭐⭐⭐⭐⭐ | 完美匹配，权限类型固定（API/MENU/BUTTON） |
| ABAC | ⭐⭐⭐ | 能做但过度设计，增加不必要的复杂度 |
| ReBAC | ⭐⭐ | 关系推导用不上，杀鸡用牛刀 |

#### 场景 2：文档协作平台（Google Docs / Notion）

```
需求：每个文档/文件夹有独立的共享权限，支持 "任何有链接的人" 访问
```

| 模型 | 适配度 | 原因 |
|------|--------|------|
| RBAC | ⭐⭐ | 角色是全局的，无法表达 "这个文档对这几个人共享" |
| ABAC | ⭐⭐⭐⭐ | 属性可以表达共享状态 |
| **ReBAC** | ⭐⭐⭐⭐⭐ | 关系图天然适合表达 "文档→共享→用户" |

#### 场景 3：金融风控系统

```
需求：根据金额、地区、时间、审批层级等多维条件动态决策
```

| 模型 | 适配度 | 原因 |
|------|--------|------|
| RBAC | ⭐⭐ | 无法表达金额阈值、时间限制等条件 |
| **ABAC** | ⭐⭐⭐⭐⭐ | 多维属性实时评估，完美匹配 |
| ReBAC | ⭐⭐⭐ | 可以部分表达，但不是最优 |

---

## 5. 业界共识：混合模型

AWS、Auth0、WorkOS 等权威来源的共识是：

> **大多数多租户 SaaS 产品应该使用 RBAC 作为基础，在需要的地方补充少量 ABAC 属性。**

这就是所谓的 **"RBAC+ 模型"**：

```
基础层（RBAC）：用户 → 角色 → 权限
  ↓ 补充
增强层（少量 ABAC）：
  - 权限检查时附加资源属性（如 resource.owner_id == user.id）
  - 权限检查时附加环境属性（如 user.tenant_id == resource.tenant_id）
```

### 为什么 RBAC 是基础？

1. **90% 的权限需求是 "谁能做什么"** — 这是 RBAC 的天然领域
2. **管理员需要在 UI 上管理权限** — RBAC 的勾选框 UI 比 ABAC 的策略编辑器简单 100 倍
3. **审计人员需要理解权限分配** — "Alice 是 Admin" 比 "Alice 满足条件 A ∧ B ∧ ¬C" 容易理解
4. **RBAC 有 NIST 标准** — NIST SP 800-162 定义了 RBAC 的标准实现

### 什么时候补充 ABAC？

当你遇到以下需求时：
- "用户只能编辑自己创建的资源" → 加 `resource.owner_id == user.id` 检查
- "审批人不能审批自己的申请" → 加 `resource.applicant_id != user.id` 检查
- "工作日 9-18 点才能操作" → 加时间属性检查（极少见）

这些检查通常 **不在权限系统中实现**，而是在 **业务逻辑层** 实现。因为它们不是"权限"，而是"业务规则"。

---

## 6. 对本项目的建议

本项目是典型的 **后台管理系统**：
- 权限类型固定：API / MENU / BUTTON
- 用户角色清晰：super_admin / admin / auditor / user
- 多租户隔离需求明确

**结论：纯 RBAC 是最佳选择**，不需要引入 ABAC 或 ReBAC。

如果未来需要 "用户只能操作自己部门的数据" 这类需求，在 **Service 层** 加一个 `department_id` 过滤即可，不需要改权限模型。

---

## 参考资料

- [NIST SP 800-162 - Guide to Attribute Based Access Control](https://csrc.nist.gov/publications/detail/sp/800-162/final)
- [AWS - SaaS Tenant Isolation with ABAC](https://aws.amazon.com/blogs/security/how-to-implement-saas-tenant-isolation-with-abac-and-aws-iam/)
- [Auth0 - Choosing the Right Authorization Model for Multi-Tenant SaaS](https://auth0.com/blog/how-to-choose-the-right-authorization-model-for-your-multi-tenant-saas-application/)
- [Permit.io - RBAC vs ABAC vs ReBAC](https://www.permit.io/blog/rbac-vs-abac-and-rebac-choosing-the-right-authorization-model)
- [WorkOS - How to Design Multi-Tenant RBAC](https://workos.com/blog/how-to-design-multi-tenant-rbac-saas)
- [Cerbos - Scalable Multitenant Authorization](https://www.cerbos.dev/blog/how-to-implement-scalable-multitenant-authorization)

---

*最后更新：2026-06-06*
