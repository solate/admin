# 多租户 SaaS 数据权限实现方案

## 一、数据库隔离方案

### 1. Database-per-Tenant（独立数据库）

**实现逻辑**：每个租户拥有独立的数据库实例

**优点**：
- 数据隔离最强
- 备份恢复独立
- 性能不互相影响
- 支持跨库迁移

**缺点**：
- 成本高
- 维护复杂
- 资源利用率低

**适用场景**：大型企业租户、对数据安全要求极高的场景

---

### 2. Schema-per-Tenant（独立 Schema）

**实现逻辑**：共享数据库实例，每个租户拥有独立的 Schema

**优点**：
- 隔离性较好
- 成本适中
- 可独立备份

**缺点**：
- 跨租户查询困难
- Schema 数量有限制

**适用场景**：中型租户、需要一定隔离的场景

---

### 3. Shared Schema（共享表 + tenant_id）

**实现逻辑**：所有租户共享表，通过 `tenant_id` 字段区分

**优点**：
- 成本最低
- 资源利用率高
- 易于维护

**缺点**：
- 隔离性弱
- 存在数据泄露风险
- 性能瓶颈

**适用场景**：小型租户、SMB 市场

---

## 二、行级权限（Row-Level Security）

### 1. 应用层过滤

**实现逻辑**：
- 每个表添加 `tenant_id` 字段
- 查询时自动添加 `WHERE tenant_id = ?`
- 通过 ORM 插件或中间件实现

**实现要点**：
- 在 SQL 拦截器中自动注入租户条件
- JWT Token 中携带 tenant_id
- 上下文传递租户信息

---

### 2. 数据库层 RLS（Row Level Security）

**实现逻辑**：
- 使用 PostgreSQL RLS 策略
- 数据库自动过滤非授权行
- 策略绑定到会话的租户 ID

**示例策略**：
```sql
CREATE POLICY tenant_isolation ON users
  USING (tenant_id = current_setting('app.current_tenant')::bigint);
```

**优点**：
- 数据库层强制隔离
- 应用层绕过也无效
- 性能较好

**缺点**：
- 数据库依赖性强
- 调试复杂

---

## 三、列级权限（Column-Level Security）

**实现逻辑**：
- 根据用户角色动态返回字段
- 敏感字段脱敏（手机号、身份证等）
- DTO 层控制字段可见性

**实现方式**：
- 使用 DTO 组合（基础 DTO + 详细 DTO）
- 在序列化时动态过滤字段
- 数据库视图限制列访问

---

## 四、数据权限层级

参考企业级框架（如 GoWind Admin），数据权限通常分为五层：

### 1. 租户隔离层
- 第一层：完全隔离不同租户数据
- 所有查询必须带 tenant_id

### 2. 业务单元层
- 部门、分公司级别隔离
- 通过 `dept_id` 实现组织架构权限

### 3. 行级数据层
- 本人数据、本部门数据、全部数据
- 通过 `data_scope` 字段控制范围

### 4. 列级字段层
- 敏感字段脱敏
- 角色控制字段可见性

### 5. 操作/状态层
- 只读、编辑、删除权限
- 工作流状态权限

---

## 五、混合方案推荐

### 分级隔离策略

**小型租户（< 100 用户）**：
- 使用 Shared Schema
- 应用层 tenant_id 过滤

**中型租户（100-1000 用户）**：
- 使用 Schema-per-Tenant
- 数据库 RLS 策略

**大型租户（> 1000 用户）**：
- 使用 Database-per-Tenant
- 物理隔离 + 专属资源

---

## 六、当前项目建议

基于当前项目架构（GORM + PostgreSQL + Casbin）：

### 推荐方案：Shared Schema + 应用层过滤 + RLS 增强

**实现路径**：

1. **数据库层**：所有表添加 `tenant_id` 字段（已有）

2. **应用层**：
   - GORM Hook 自动注入 `tenant_id`
   - 中间件从 JWT 提取租户信息
   - 全局查询作用域

3. **权限层**：
   - Casbin 策略格式：`sub, dom, obj, act`
   - 支持行级策略：`sub, dom, obj, act, data_scope`

4. **增强安全**：
   - 可选启用 PostgreSQL RLS 作为兜底
   - 防止应用层 bug 导致数据泄露

---

## 七、关键安全点

1. **默认拒绝**：所有查询默认带租户过滤
2. **防止越权**：禁止手动修改 tenant_id
3. **审计日志**：记录所有跨租户访问尝试
4. **定期检查**：代码扫描确保无遗漏租户过滤的查询

---

## 参考资源

- [Azure SQL Database - Multitenant SaaS Patterns](https://learn.microsoft.com/en-us/azure/azure-sql/database/saas-tenancy-app-design-patterns?view=azuresql)
- [阿里云 - 基于 RLS 实现 SaaS 租户数据隔离](https://help.aliyun.com/zh/rds/apsaradb-rds-for-postgresql/implementation-of-saas-tenant-data-isolation-solution-based-on-row-level)
- [AWS - 行级安全建议](https://docs.aws.amazon.com/zh_cn/prescriptive-guidance/latest/saas-multitenant-managed-postgresql/rls.html)
- [Multi Tenant Data Isolation Patterns](https://securingbits.com/multi-tenant-data-isolation-patterns)
- [SpringBoot 多租户架构设计终极指南](https://modelers.csdn.net/6914586f5511483559e98c84.html)
- [Tenant Data Isolation: Patterns and Anti-Patterns](https://propelius.ai/blogs/tenant-data-isolation-patterns-and-anti-patterns)
