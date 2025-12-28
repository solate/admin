# 字典系统设计

## 设计原则

- **超管创建系统字典**：在默认租户下创建字典模板
- **租户覆盖显示文本**：通过相同 `type_id` 和 `value` 创建租户记录
- **查询自动合并**：租户覆盖的优先，系统默认的兜底
- **恢复系统默认**：删除租户的覆盖记录即可

---

## 数据模型

### 字典类型表 (dict_types)

```sql
CREATE TABLE dict_types (
    type_id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,       -- 默认租户为空字符串
    type_code VARCHAR(50) NOT NULL,       -- 字典编码，如: order_status
    type_name VARCHAR(100) NOT NULL,      -- 字典名称，如: 订单状态
    UNIQUE KEY uk_tenant_code(tenant_id, type_code)
);
```

### 字典项表 (dict_items)

```sql
CREATE TABLE dict_items (
    item_id VARCHAR(36) PRIMARY KEY,
    type_id VARCHAR(36) NOT NULL,         -- 关联字典类型
    tenant_id VARCHAR(36) NOT NULL,       -- 默认租户为空字符串
    label VARCHAR(100) NOT NULL,          -- 显示文本
    value VARCHAR(100) NOT NULL,          -- 实际值（不可变，用于匹配覆盖）
    sort INT DEFAULT 0,
    INDEX idx_type_value(type_id, tenant_id, value)
);
```

---

## 核心设计

### 覆盖机制

```
系统字典（tenant_id = ""）:
├── type_id: "type-001"
├── items:
│   ├── label: "待支付", value: "pending"
│   └── label: "已支付", value: "paid"

租户覆盖（tenant_id = "tenant-001"）:
├── type_id: "type-001"  -- 相同 type_id
├── items:
│   └── label: "待付款", value: "pending"  -- 相同 value，不同 label

查询结果:
├── 待付款 (pending)  -- 租户覆盖
└── 已支付 (paid)    -- 系统默认
```

**核心逻辑**：
- 租户覆盖 = 创建新记录（相同 `type_id` + `value`，不同 `tenant_id` + `label`）
- 查询时合并 = 先查系统，再查租户，相同 `value` 的用租户的 `label`

---

## 业务逻辑

### 1. 超管创建系统字典

```go
func (s *DictService) CreateSystemDict(ctx context.Context, req *CreateSystemDictRequest) error {
    // 创建字典类型
    dictType := &DictType{
        TypeID:   uuid.New(),
        TenantID: constants.DefaultTenantID,  // 空字符串
        TypeCode: req.TypeCode,  // "order_status"
        TypeName: req.TypeName,  // "订单状态"
    }
    s.dictTypeRepo.Create(ctx, dictType)

    // 创建字典项
    for _, item := range req.Items {
        s.dictItemRepo.Create(ctx, &DictItem{
            ItemID:   uuid.New(),
            TypeID:   dictType.TypeID,
            TenantID: constants.DefaultTenantID,
            Label:    item.Label,  // "待支付"
            Value:    item.Value,  // "pending"
            Sort:     item.Sort,
        })
    }
    return nil
}
```

### 2. 租户获取字典（自动合并）

```go
func (s *DictService) GetDictByCode(ctx context.Context, typeCode string) (*DictVO, error) {
    tenantID := getTenantID(ctx)

    // 1. 获取系统字典类型
    systemType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, constants.DefaultTenantID)
    if err != nil {
        return nil, errors.New("字典不存在")
    }

    // 2. 获取系统字典项
    systemItems, _ := s.dictItemRepo.GetByTypeAndTenant(ctx, systemType.TypeID, constants.DefaultTenantID)

    // 3. 获取租户覆盖的字典项
    overrideItems, _ := s.dictItemRepo.GetByTypeAndTenant(ctx, systemType.TypeID, tenantID)

    // 4. 构建覆盖映射
    overrideMap := make(map[string]*DictItem)  // value -> DictItem
    for _, item := range overrideItems {
        overrideMap[item.Value] = item
    }

    // 5. 合并（租户覆盖的优先）
    var result []*DictItemVO
    for _, item := range systemItems {
        if override, exists := overrideMap[item.Value]; exists {
            result = append(result, &DictItemVO{
                Label:  override.Label,  // 租户的 label
                Value:  item.Value,
                Sort:   override.Sort,
                Source: "custom",
            })
        } else {
            result = append(result, &DictItemVO{
                Label:  item.Label,      // 系统的 label
                Value:  item.Value,
                Sort:   item.Sort,
                Source: "system",
            })
        }
    }

    // 6. 添加租户独有的字典项（系统没有的）
    for _, item := range overrideItems {
        found := false
        for _, sysItem := range systemItems {
            if sysItem.Value == item.Value {
                found = true
                break
            }
        }
        if !found {
            result = append(result, &DictItemVO{
                Label:  item.Label,
                Value:  item.Value,
                Sort:   item.Sort,
                Source: "custom",
            })
        }
    }

    // 7. 按 sort 排序
    sort.Slice(result, func(i, j int) bool {
        return result[i].Sort < result[j].Sort
    })

    return &DictVO{
        TypeCode: systemType.TypeCode,
        TypeName: systemType.TypeName,
        Items:    result,
    }, nil
}
```

### 3. 租户覆盖字典项

```go
func (s *DictService) UpdateDictItem(ctx context.Context, req *UpdateDictItemRequest) error {
    tenantID := getTenantID(ctx)

    // 1. 获取系统字典类型
    systemType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, req.TypeCode, constants.DefaultTenantID)
    if err != nil {
        return errors.New("字典不存在")
    }

    // 2. 检查租户是否已有覆盖记录
    existing, _ := s.dictItemRepo.GetByTypeAndValue(ctx, systemType.TypeID, tenantID, req.Value)
    if existing != nil {
        // 更新现有记录
        existing.Label = req.Label
        existing.Sort = req.Sort
        s.dictItemRepo.Update(ctx, existing)
        return nil
    }

    // 3. 创建新的覆盖记录
    s.dictItemRepo.Create(ctx, &DictItem{
        ItemID:   uuid.New(),
        TypeID:   systemType.TypeID,  -- 使用系统字典的 type_id
        TenantID: tenantID,
        Label:    req.Label,  -- 租户自定义
        Value:    req.Value,  -- 保持不变（用于匹配）
        Sort:     req.Sort,
    })
    return nil
}
```

### 4. 租户恢复系统默认值

```go
func (s *DictService) ResetDictItem(ctx context.Context, typeCode, value string) error {
    tenantID := getTenantID(ctx)

    // 1. 获取系统字典类型
    systemType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, constants.DefaultTenantID)
    if err != nil {
        return errors.New("字典不存在")
    }

    // 2. 删除租户的覆盖记录（恢复系统默认）
    s.dictItemRepo.DeleteByTypeAndValue(ctx, systemType.TypeID, tenantID, value)
    return nil
}
```

---

## API 设计

### 超管接口

```
POST   /api/v1/system/dict                    创建系统字典
PUT    /api/v1/system/dict/:typeCode          更新系统字典
GET    /api/v1/system/dict                    获取系统字典列表
DELETE /api/v1/system/dict/:typeCode          删除系统字典
```

### 租户接口

```
GET    /api/v1/dict/:typeCode                 获取字典（合并系统+覆盖）
PUT    /api/v1/dict/:typeCode/items           批量更新字典项
DELETE /api/v1/dict/:typeCode/items/:value    恢复系统默认值
```

---

## 数据示例

### 系统字典（超管创建）

```sql
-- dict_types
INSERT INTO dict_types (type_id, tenant_id, type_code, type_name) VALUES
('type-001', '', 'order_status', '订单状态');

-- dict_items (系统)
INSERT INTO dict_items (item_id, type_id, tenant_id, label, value, sort) VALUES
('item-001', 'type-001', '', '待支付', 'pending', 0),
('item-002', 'type-001', '', '已支付', 'paid', 1),
('item-003', 'type-001', '', '已完成', 'finished', 2),
('item-004', 'type-001', '', '已取消', 'canceled', 3);
```

### 租户覆盖

```sql
-- dict_items (租户 tenant-001 覆盖)
INSERT INTO dict_items (item_id, type_id, tenant_id, label, value, sort) VALUES
('item-101', 'type-001', 'tenant-001', '待付款', 'pending', 0),  -- 覆盖
('item-102', 'type-001', 'tenant-001', '已付款', 'paid', 1),     -- 覆盖
('item-103', 'type-001', 'tenant-001', '已关闭', 'closed', 4);   -- 新增
```

### 查询结果

```json
{
  "typeCode": "order_status",
  "typeName": "订单状态",
  "items": [
    { "label": "待付款", "value": "pending", "sort": 0, "source": "custom" },
    { "label": "已付款", "value": "paid", "sort": 1, "source": "custom" },
    { "label": "已完成", "value": "finished", "sort": 2, "source": "system" },
    { "label": "已取消", "value": "canceled", "sort": 3, "source": "system" },
    { "label": "已关闭", "value": "closed", "sort": 4, "source": "custom" }
  ]
}
```

---

## Repository 层实现

### DictItemRepo

```go
// 获取指定类型和租户的字典项
func (r *DictItemRepo) GetByTypeAndTenant(ctx context.Context, typeID, tenantID string) ([]*DictItem, error) {
    return r.q.DictItem.WithContext(ctx).
        Where(r.q.DictItem.TypeID.Eq(typeID)).
        Where(r.q.DictItem.TenantID.Eq(tenantID)).
        Order(r.q.DictItem.Sort).
        Find()
}

// 获取指定类型、租户、值的字典项（用于检查是否已覆盖）
func (r *DictItemRepo) GetByTypeAndValue(ctx context.Context, typeID, tenantID, value string) (*DictItem, error) {
    return r.q.DictItem.WithContext(ctx).
        Where(r.q.DictItem.TypeID.Eq(typeID)).
        Where(r.q.DictItem.TenantID.Eq(tenantID)).
        Where(r.q.DictItem.Value.Eq(value)).
        First()
}

// 删除指定类型、租户、值的字典项（恢复系统默认）
func (r *DictItemRepo) DeleteByTypeAndValue(ctx context.Context, typeID, tenantID, value string) error {
    _, err := r.q.DictItem.WithContext(ctx).
        Where(r.q.DictItem.TypeID.Eq(typeID)).
        Where(r.q.DictItem.TenantID.Eq(tenantID)).
        Where(r.q.DictItem.Value.Eq(value)).
        Delete()
    return err
}
```

---

## 常量定义

```go
package constants

const (
    // 租户
    DefaultTenantID   = ""       // 默认租户ID为空字符串
    DefaultTenantCode = "default" // 默认租户code
)
```

---

## 总结

| 特性 | 实现方式 |
|------|----------|
| 系统字典模板 | `tenant_id = ""` |
| 租户覆盖 | 相同 `type_id` + `value`，不同 `tenant_id` + `label` |
| 查询合并 | 系统项 ∪ 覆盖项（相同 value 的覆盖优先） |
| 恢复默认 | 删除租户的覆盖记录 |
| 独有扩展 | 租户可以添加系统没有的字典项 |
