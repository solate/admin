# ABAC / ReBAC / 跨域策略：真实产品场景对照

> 文档里总说 "ABAC 和 ReBAC 适合复杂场景"，但到底什么才算复杂？这里用真实产品来说明。

---

## 1. ABAC 的真实产品场景

### 场景 A：AWS IAM ABAC（标签策略）

**产品**：Amazon Web Services

**权限逻辑**：用户能访问哪些 AWS 资源，取决于 **用户身上的标签** 和 **资源身上的标签** 是否匹配。

```
用户标签：  { department: "engineering", project: "alpha", level: "senior" }
资源标签：  { department: "engineering", project: "alpha", classification: "internal" }

策略规则：
  允许访问 IF  user.department == resource.department
          AND user.project == resource.project
          AND resource.classification IN ["internal", "public"]
          AND current_time BETWEEN 09:00 AND 18:00
```

**为什么 RBAC 做不了？**

RBAC 只能回答 "用户是什么角色"，无法表达：
- "用户只能访问 **自己部门** 的资源"（需要匹配 user.dept 和 resource.dept）
- "只有 **工作时间段** 才能操作"（环境属性）
- "机密级资源只有 senior 以上能看"（资源属性 + 用户属性组合）

**如果用 RBAC 硬做**：需要创建 `engineering_alpha_senior` 角色，然后 AWS 有 100 个部门 × 50 个项目 × 3 个级别 = **15,000 个角色**。这就是 "角色爆炸"。

---

### 场景 B：Azure Conditional Access（条件访问）

**产品**：Microsoft Entra ID（原 Azure AD）

**权限逻辑**：登录是否允许，取决于设备、位置、风险等级等 **实时属性**。

```
条件访问策略：
  IF   用户在 "财务部" 组
  AND  登录位置 NOT IN "中国"
  AND  设备合规状态 == "不合规"
  AND  登录风险等级 == "高"
  THEN 要求 MFA + 即时阻断
```

**为什么 RBAC 做不了？**

RBAC 的决策是 **静态的**（角色分配后不会变），但条件访问需要 **每次登录实时评估**：
- 用户在办公室用公司电脑 → 允许
- 同一个用户在境外用陌生电脑 → 拒绝
- 同一个人、同一个角色，不同条件 → 不同结果

---

### 场景 C：Google Cloud IAM Conditions

**产品**：Google Cloud Platform

**权限逻辑**：权限绑定上加 CEL 表达式，运行时动态评估。

```yaml
# 只允许在特定时间段、特定资源上、特定条件下操作
role: roles/storage.objectAdmin
condition:
  expression: |
    request.time.getHours(request.origin) >= 9
    && request.time.getHours(request.origin) <= 17
    && resource.matchTag("env", "production")
    && request.auth.claims.department == "devops"
```

---

### 场景 D：医疗系统（HIPAA 合规）

**产品**：Epic Systems、Cerner 等电子病历系统

**权限逻辑**：医生能看到患者的哪些数据，取决于 **医患关系、科室、数据敏感度**。

```
规则：
  - 主治医生 → 看完整病历
  - 急诊医生 → 看急诊相关记录（不看精神科记录）
  - 心理科医生 → 只看心理评估（不看身体检查结果）
  - HIV 阳性标记 → 只有指定医生能看到
  - 未成年人 → 父母能看部分，但隐私话题（如心理咨询）父母不能看
  - 已出院 → 医生只能看摘要，不能看详细记录
```

**为什么 RBAC 做不了？**

这不是 "角色有什么权限" 的问题，而是 **每条数据都有不同的访问规则**：
- 同一个医生，对 A 患者能看完整病历，对 B 患者只能看摘要
- 同一条记录，对主治医生可见，对其他科室不可见
- 规则依赖 **资源内容**（HIV 标记、年龄、科室分类）

---

### 场景 E：金融风控系统

**产品**：蚂蚁集团风控引擎、Stripe Radar

**权限逻辑**：交易是否允许，取决于金额、时间、设备指纹、历史行为等数十个维度。

```
规则（简化）：
  IF   交易金额 > 用户.单日限额
  OR   交易地点距上次交易地点 > 500km 且 时间差 < 2h
  OR   设备指纹不在用户常用设备列表中
  OR   用户.风险评分 > 80
  THEN 拦截 + 短信验证
```

**为什么 RBAC 做不了？**

决策依赖 **交易属性**（金额、地点、时间）和 **用户行为画像**（常用设备、风险评分），这些都不是"角色"能表达的。

---

## 2. ReBAC 的真实产品场景

### 场景 F：Google Drive 文档共享

**产品**：Google Drive / Google Docs

**权限逻辑**：文档权限通过 **关系链** 传递推导。

```
用户 Alice 创建了文档 D
  → Alice 是 D 的 owner
  → Alice 共享 D 给 Team-A (editor)
  → Bob 是 Team-A 的 member
  → 推导：Bob 可以编辑 D

  → D 在 Folder-F 中
  → Folder-F 共享给 Team-B (viewer)
  → Charlie 是 Team-B 的 member
  → 推导：Charlie 可以查看 D（通过文件夹关系继承）

  → D 的链接分享设为 "任何有链接的人可评论"
  → 陌生用户访客 有链接
  → 推导：访客 可以评论 D
```

**为什么 RBAC 做不了？**

RBAC 的粒度是 "操作类型"（读/写/删），而 ReBAC 需要在 **每个资源实例** 上定义关系：
- 文档 A → Alice 是 owner → Alice 可编辑
- 文档 B → Alice 是 viewer → Alice 只读
- 同一个人、同一个 "角色"，对 **不同文档** 有不同权限
- 权限通过 **文件夹 → 团队 → 用户** 的关系链 **递归推导**

如果用 RBAC 做：每个文档都要建一套角色（`doc_A_owner`, `doc_A_editor`, `doc_A_viewer`...），Google Drive 有几十亿文档，角色数量会爆炸。

---

### 场景 G：GitHub 仓库权限

**产品**：GitHub

**权限逻辑**：通过 **组织 → 团队 → 仓库** 的多层关系推导。

```
GitHub 的关系图：
  Organization: AcmeCorp
    ├── Team: Backend
    │     ├── Member: Alice
    │     ├── Member: Bob
    │     └── Repo Access: api-server (write), infra (read)
    ├── Team: Frontend
    │     ├── Member: Charlie
    │     └── Repo Access: web-app (write)
    └── Team: Leads
          ├── Member: Alice          ← Alice 同时在两个团队
          └── Repo Access: * (admin) ← 所有仓库的管理权限

推导：
  Alice 对 api-server → write (通过 Backend 团队) + admin (通过 Leads 团队) → 最终 admin
  Alice 对 web-app    → admin (通过 Leads 团队)
  Bob 对 web-app      → 无权限
  Charlie 对 api-server → 无权限
```

**为什么 RBAC 做不了？**

- 权限不是 "Alice 有 write 角色"，而是 "Alice 通过 Backend 团队对 api-server 有 write 权限"
- 同一个人通过不同的 **关系路径** 到达同一个资源，权限可以叠加
- 需要沿着 **组织 → 团队 → 仓库** 三层关系做 **图遍历**

---

### 场景 H：Notion 页面权限树

**产品**：Notion

**权限逻辑**：页面权限沿着树形结构 **向下继承**，可以被子页面 **覆盖**。

```
Workspace
├── Page A (共享给 Team-1: editor)
│   ├── Sub-page A1 (继承 → Team-1: editor)
│   ├── Sub-page A2 (覆盖 → Team-1: viewer)  ← 特殊：降低权限
│   └── Sub-page A3 (共享给 user-X: editor)   ← 额外授权
├── Page B (私有)
└── Page C (公开链接: viewer)
    └── Sub-page C1 (继承 → 公开链接: viewer)

推导：
  Team-1 成员 → 可编辑 A, A1 / 只读 A2 / 不可见 A3
  user-X      → 可编辑 A3（即使不在 Team-1 中）
  任何人      → 可查看 C, C1（通过公开链接）
```

**为什么 RBAC 做不了？**

权限是 **页面实例级** 的，而且需要：
- 向下继承（父页面权限传播到子页面）
- 覆盖（子页面可以降低或提升权限）
- 合并（多个来源的权限取并集）

---

## 3. 跨域策略组合

### 场景 I：跨组织协作（B2B SaaS）

**产品**：Slack Connect、Microsoft Teams 外部共享

**权限逻辑**：两个组织的策略 **叠加**，取 **交集**（更严格的那个）。

```
Acme 公司的策略：
  - 员工可以共享内部文档
  - 但不能共享标记为 "机密" 的文档

合作伙伴公司的策略：
  - 外部人员只能通过 MFA 访问共享内容
  - 共享内容 30 天后自动撤销

叠加结果：
  Acme 员工共享文档给合作伙伴 →
    1. 不能是机密文档（Acme 策略）
    2. 合作伙伴必须 MFA 验证（合作伙伴策略）
    3. 30 天后自动撤销（合作伙伴策略）
```

**为什么 RBAC 做不了？**

两套独立的角色体系需要 **合并评估**：
- Acme 的 "editor" 角色在合作伙伴的上下文中可能映射为 "viewer"
- 一个操作要同时满足两边的策略
- 策略可能冲突（一边允许、一边禁止），需要冲突解决规则

---

### 场景 J：联邦身份（SSO 跨系统）

**产品**：Okta → Slack + Jira + Confluence 联合登录

**权限逻辑**：IdP（身份提供方）的策略 + SP（服务提供方）的策略 **同时生效**。

```
IdP (Okta) 策略：
  - 开发人员组 → 可以登录开发工具类应用
  - 但不允许在工作时间之外登录
  - 设备必须通过合规检查

SP (Jira) 策略：
  - 项目成员 → 可以查看和编辑 issue
  - 非 project lead → 不能删除 issue

SP (Confluence) 策略：
  - 空间成员 → 可以查看页面
  - 管理员 → 可以管理空间设置

登录 Jira 的完整评估：
  1. Okta 检查：是开发人员？在工作时间？设备合规？→ 通过
  2. Jira 检查：是项目成员？→ 通过
  3. 最终：允许登录并授予项目权限
```

---

### 场景 K：供应链系统

**产品**：SAP Ariba、Oracle SCM

**权限逻辑**：供应商、制造商、分销商各自有独立的权限体系，需要在交叉点上协调。

```
供应链多层权限：
  制造商（Manufacturer）
    → 能看自己的订单和供应商报价
    → 不能看其他制造商的数据

  供应商 A（Supplier A）
    → 只能看到制造商发给自己的订单
    → 不能看到供应商 B 的报价

  审计公司（Auditor）
    → 能看所有交易记录（只读）
    → 但不能看具体价格（合规要求）

  物流公司（Logistics）
    → 只能看到物流相关信息（地址、重量、时间）
    → 不能看价格、供应商信息
```

---

## 4. 对比总结：后台管理系统会不会遇到？

| 场景 | 真实产品 | 后台管理系统会碰到吗 | 需要什么模型 |
|------|----------|---------------------|-------------|
| A. AWS 标签策略 | AWS IAM | ❌ 不会 | ABAC |
| B. 条件访问 | Azure AD | ❌ 不会 | ABAC |
| C. GCP Conditions | Google Cloud | ❌ 不会 | ABAC |
| D. 医疗病历 | Epic/Cerner | ❌ 不会（除非做医疗系统） | ABAC |
| E. 金融风控 | 蚂蚁/Stripe | ❌ 不会（除非做支付系统） | ABAC |
| F. 文档共享 | Google Drive | ❌ 不会 | ReBAC |
| G. 仓库权限 | GitHub | ❌ 不会 | ReBAC |
| H. 页面权限树 | Notion | ❌ 不会 | ReBAC |
| I. 跨组织协作 | Slack Connect | ⚠️ 可能（如果做 B2B 协作） | 跨域策略 |
| J. 联邦身份 SSO | Okta | ⚠️ 可能（如果接 SSO） | 跨域策略 |
| K. 供应链 | SAP Ariba | ❌ 不会 | 跨域策略 |

### 后台管理系统的权限需求

```
典型的后台管理系统（如本项目的 Go Admin）：

✅ 功能权限：用户能调哪些 API → RBAC（已实现）
✅ 数据权限：用户看哪些数据行 → RBAC + data_scope（规划中）
✅ 租户隔离：用户只看自己租户的数据 → WHERE tenant_id = ?（已实现）
✅ 超管绕行：超管跳过所有检查 → HasRole("super_admin")（已实现）

❌ 不需要：按金额/时间/地点动态决策（ABAC）
❌ 不需要：每个资源实例独立权限（ReBAC）
❌ 不需要：跨组织策略叠加（跨域策略）
```

### 判断标准

问自己三个问题：

```
1. 权限是跟着"角色"走的，还是跟着"每个资源实例"走的？
   → 跟角色 = RBAC
   → 跟资源实例 = ReBAC

2. 权限决策需要看"资源的内容"吗？（金额、状态、标签...）
   → 不需要 = RBAC
   → 需要 = ABAC

3. 权限需要跨越两个独立组织的策略体系吗？
   → 不需要 = RBAC
   → 需要 = 跨域策略
```

**本项目的答案**：1. 跟角色 2. 不需要 3. 不需要 → **纯 RBAC 完全够用**。

---

*最后更新：2026-06-06*
