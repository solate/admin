# 关联表设计规范

## 规则 1：独立主键，禁止复合主键

```sql
-- ✅ 独立主键 + 唯一约束
CREATE TABLE role_permissions (
    id            BIGSERIAL    PRIMARY KEY,
    role_id       VARCHAR(20)  NOT NULL,
    permission_id VARCHAR(20)  NOT NULL,
    tenant_id     VARCHAR(20)  NOT NULL,
    created_at    BIGINT       NOT NULL DEFAULT 0,
    UNIQUE(role_id, permission_id, tenant_id)
);
```

## 规则 2：必须有时间戳

所有关联表包含 `created_at BIGINT`。如果记录会被更新，也加 `updated_at`。

## 规则 3：必须有唯一约束

在外键组合上创建唯一约束防重复：`UNIQUE(role_id, menu_id)`。

## 规则 4：Repository 创建时生成主键

使用 `idgen.GenerateUUID()` 或 `idgen.GenerateUUIDs(n)` 生成主键，不要依赖数据库自增（BIGSERIAL 表除外）。

## 规则 5：双向查询索引

```sql
CREATE INDEX idx_user_roles_user_tenant ON user_roles(user_id, tenant_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
```
