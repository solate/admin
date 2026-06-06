# 权限管理系统 API 接口设计

## 接口规范

### 基础路径
- 生产环境：`/api/v1`
- 测试环境：`/api/v1`

### URL 设计规范
本项目采用统一的 URL 设计规范：

1. **列表查询**：`GET /api/v1/resources`
   - 使用 query 参数传递过滤条件、分页参数
   - 示例：`GET /api/v1/menus?page=1&page_size=20`

2. **详情查询**：`GET /api/v1/resources/detail`
   - 使用 query 参数传递 id
   - 示例：`GET /api/v1/menus/detail?menu_id=2`

3. **创建资源**：`POST /api/v1/resources`
   - 请求 body 中包含完整资源数据
   - 示例：`POST /api/v1/menus`

4. **更新资源**：`PUT /api/v1/resources`
   - 请求 body 中包含 id 和需要更新的字段
   - 示例：`PUT /api/v1/menus`（body 中包含 menu_id）

5. **删除资源**：`DELETE /api/v1/resources`
   - 使用 query 参数传递 id
   - 示例：`DELETE /api/v1/menus?menu_id=2`

6. **批量删除**：`DELETE /api/v1/resources/batch-delete`
   - 请求 body 中包含 id 数组
   - 示例：`DELETE /api/v1/menus/batch-delete`

7. **状态更新**：`PUT /api/v1/resources/status`
   - 用于启用/禁用等状态切换
   - 示例：`PUT /api/v1/menus/status`

8. **获取全部（不分页）**：`GET /api/v1/resources/all`
   - 用于下拉选择等场景
   - 示例：`GET /api/v1/menus/all`

9. **子资源操作**：`/api/v1/{parent_id}/{sub-resource}`
   - 嵌套资源使用路径参数
   - 示例：`GET /api/v1/roles/1/menu-permissions`

### 请求头
```http
Content-Type: application/json
Authorization: Bearer {access_token}
X-Tenant-ID: {tenant_id}  // 多租户场景
```

### 响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": 1706745600000
}
```

### 错误响应
```json
{
  "code": 10001,
  "message": "参数验证失败",
  "data": null,
  "timestamp": 1706745600000
}
```

---


## 三、数据权限管理

### 3.1 数据权限规则管理

#### 3.1.1 创建数据权限规则
**接口**：`POST /api/v1/data-permission-rules`

**描述**：创建新的数据权限规则模板。

**权限**：`system:data-permission:create`

**请求参数**：
```json
{
  "name": "本部门及下级部门",
  "code": "dept_and_sub",
  "scope_type": "custom",
  "description": "可以查看本部门及所有下级部门的数据"
}
```

**参数说明**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | 规则名称 |
| code | string | 是 | 规则代码（唯一） |
| scope_type | string | 是 | 范围类型（all/dept/self/custom） |
| description | string | 否 | 规则描述 |

**响应示例**：
```json
{
  "code": 0,
  "message": "创建成功",
  "data": {
    "id": 4,
    "name": "本部门及下级部门",
    "code": "dept_and_sub",
    "scope_type": "custom",
    "description": "可以查看本部门及所有下级部门的数据",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### 3.1.2 更新数据权限规则
**接口**：`PUT /api/v1/data-permission-rules`

**描述**：更新数据权限规则信息。

**权限**：`system:data-permission:update`

**请求参数**：
```json
{
  "id": 4,
  "name": "本部门及下级部门",
  "code": "dept_and_sub",
  "scope_type": "custom",
  "description": "可以查看本部门及所有下级部门的数据"
}
```

**参数说明**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int64 | 是 | 规则 ID |
| name | string | 是 | 规则名称 |
| code | string | 是 | 规则代码（唯一） |
| scope_type | string | 是 | 范围类型（all/dept/self/custom） |
| description | string | 否 | 规则描述 |

**响应示例**：
```json
{
  "code": 0,
  "message": "更新成功",
  "data": {
    "id": 4,
    "updated_at": "2024-01-01T01:00:00Z"
  }
}
```

---

#### 3.1.3 删除数据权限规则
**接口**：`DELETE /api/v1/data-permission-rules`

**描述**：删除数据权限规则（同时删除相关的绑定关系）。

**权限**：`system:data-permission:delete`

**查询参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int64 | 是 | 规则 ID |

**请求示例**：
```http
DELETE /api/v1/data-permission-rules?id=4
```

**响应示例**：
```json
{
  "code": 0,
  "message": "删除成功",
  "data": null
}
```

---

#### 3.1.4 分页查询数据权限规则
**接口**：`GET /api/v1/data-permission-rules`

**描述**：分页查询数据权限规则列表。

**权限**：`system:data-permission:query`

**查询参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |
| scope_type | string | 否 | 按范围类型过滤 |

**响应示例**：
```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "total": 4,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": 1,
        "name": "全部数据",
        "code": "all",
        "scope_type": "all",
        "description": "可以查看所有数据",
        "created_at": "2024-01-01T00:00:00Z"
      },
      {
        "id": 2,
        "name": "本部门数据",
        "code": "dept",
        "scope_type": "dept",
        "description": "只能查看本部门的数据",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### 3.2 数据权限绑定管理

#### 3.2.1 为角色绑定数据权限
**接口**：`PUT /api/v1/roles/data-permissions`

**描述**：为角色绑定特定资源的数据权限规则。

**权限**：`system:role:assign-data-permission`

**请求参数**：
```json
{
  "role_id": 1,
  "bindings": [
    {
      "resource_type": "user",
      "rule_id": 2
    },
    {
      "resource_type": "order",
      "rule_id": 3
    }
  ]
}
```

**参数说明**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| role_id | int64 | 是 | 角色 ID |
| bindings | array | 是 | 绑定规则数组 |
| bindings[].resource_type | string | 是 | 资源类型/业务实体（如：user, order） |
| bindings[].rule_id | int | 是 | 规则 ID |

**业务逻辑**：
1. 删除该角色原有的所有数据权限绑定（针对指定的 resource_type）
2. 批量插入新的数据权限绑定
3. 事务处理，保证数据一致性

**响应示例**：
```json
{
  "code": 0,
  "message": "绑定成功",
  "data": {
    "role_id": 1,
    "binding_count": 2
  }
}
```

---

#### 3.2.2 查询角色的数据权限绑定
**接口**：`GET /api/v1/roles/data-permissions`

**描述**：查询角色的数据权限绑定情况。

**权限**：`system:role:query-data-permission`

**查询参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| role_id | int64 | 是 | 角色 ID |

**请求示例**：
```http
GET /api/v1/roles/data-permissions?role_id=1
```

**响应示例**：
```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "role_id": 1,
    "bindings": [
      {
        "id": 1,
        "resource_type": "user",
        "rule_id": 2,
        "rule_name": "本部门数据",
        "scope_type": "dept",
        "created_at": "2024-01-01T00:00:00Z"
      },
      {
        "id": 2,
        "resource_type": "order",
        "rule_id": 3,
        "rule_name": "仅本人数据",
        "scope_type": "self",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

#### 3.2.3 为角色配置自定义数据权限
**接口**：`POST /api/v1/roles/data-permissions/custom`

**描述**：为角色配置自定义数据权限规则（当 scope_type 为 custom 时使用）。

**权限**：`system:role:assign-data-permission`

**请求参数**：
```json
{
  "role_id": 1,
  "resource_type": "order",
  "custom_rules": [
    {
      "field_name": "dept_id",
      "operator": "custom",
      "custom_expression": "dept_id IN (SELECT id FROM dept WHERE path LIKE CONCAT((SELECT path FROM dept WHERE id = ?), '%'))",
      "sort": 1
    },
    {
      "field_name": "region_id",
      "operator": "in",
      "field_value": "[\"region_001\", \"region_002\"]",
      "sort": 2
    }
  ]
}
```

**参数说明**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| role_id | int64 | 是 | 角色 ID |
| resource_type | string | 是 | 资源类型 |
| custom_rules | array | 是 | 自定义规则数组 |
| custom_rules[].field_name | string | 是 | 字段名称 |
| custom_rules[].operator | string | 是 | 操作符（eq/in/custom） |
| custom_rules[].field_value | string | 否 | 字段值（JSON 格式） |
| custom_rules[].custom_expression | string | 否 | 自定义 SQL 表达式 |
| custom_rules[].sort | int | 是 | 排序序号 |

**响应示例**：
```json
{
  "code": 0,
  "message": "配置成功",
  "data": {
    "binding_id": 10,
    "custom_rule_count": 2
  }
}
```

---

#### 3.2.4 获取用户的数据权限 SQL
**接口**：`POST /api/v1/user/data-permission-sql`

**描述**：根据用户角色生成数据权限过滤的 SQL WHERE 条件。

**权限**：需要登录认证

**请求参数**：
```json
{
  "resource_type": "order"
}
```

**响应示例**：
```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "resource_type": "order",
    "scope_type": "dept",
    "sql": "dept_id = ?",
    "params": ["dept_001"]
  }
}
```

---

## 四、统一权限查询接口

### 4.1 获取角色的所有权限
**接口**：`GET /api/v1/roles/all-permissions`

**描述**：查询角色的所有权限（API 权限、菜单权限、数据权限），用于角色管理界面展示。

**权限**：`system:role:query-permission`

**查询参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| role_id | int64 | 是 | 角色 ID |

**请求示例**：
```http
GET /api/v1/roles/all-permissions?role_id=1
```

**响应示例**：
```json
{
  "code": 0,
  "message": "查询成功",
  "data": {
    "role_id": 1,
    "role_name": "管理员",
    "api_permissions": {
      "total": 45,
      "modules": [
        {
          "module": "用户管理",
          "count": 15,
          "resources": [
            {
              "id": 1,
              "name": "用户列表",
              "path": "/api/v1/users",
              "method": "GET"
            }
          ]
        }
      ]
    },
    "menu_permissions": {
      "total": 23,
      "menu_ids": ["2", "3", "4", "5"],
      "tree": [
        {
          "menu_id": "2",
          "name": "用户管理",
          "type": "menu",
          "children": [
            {
              "menu_id": "3",
              "name": "新增",
              "type": "button"
            }
          ]
        }
      ]
    },
    "data_permissions": {
      "total": 2,
      "bindings": [
        {
          "resource_type": "user",
          "rule_id": 2,
          "rule_name": "本部门数据"
        }
      ]
    }
  }
}
```

---

### 4.2 复制角色权限
**接口**：`POST /api/v1/roles/copy-permissions`

**描述**：将源角色的所有权限复制到目标角色。

**权限**：`system:role:copy-permission`

**请求参数**：
```json
{
  "source_role_id": 1,
  "target_role_id": 5,
  "copy_api_permissions": true,
  "copy_menu_permissions": true,
  "copy_data_permissions": true
}
```

**参数说明**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| source_role_id | int64 | 是 | 源角色 ID |
| target_role_id | int64 | 是 | 目标角色 ID |
| copy_api_permissions | bool | 否 | 是否复制 API 权限 |
| copy_menu_permissions | bool | 否 | 是否复制菜单权限 |
| copy_data_permissions | bool | 否 | 是否复制数据权限 |

**响应示例**：
```json
{
  "code": 0,
  "message": "复制成功",
  "data": {
    "source_role_id": 1,
    "target_role_id": 5,
    "copied": {
      "api_permissions": 45,
      "menu_permissions": 23,
      "data_permissions": 2
    }
  }
}
```

---

## 五、权限校验接口（前端调用）

### 5.1 ~~批量检查前端按钮权限~~
**状态**：✅ 已废弃

**废弃原因**：
- 前端页面加载时重复调用API，性能差
- 增加不必要的网络开销
- 造成按钮显示闪烁

**替代方案**：
- 使用 `GET /api/v1/user/menus` 接口返回的 `button_permissions` 字段
- 前端在登录时一次性获取并缓存所有按钮权限
- 通过自定义指令（如 `v-permission`）控制按钮显示

**参考实现**：
```javascript
// Vue 权限指令示例
Vue.directive('permission', {
  inserted(el, binding) {
    const { value } = binding
    const permissions = store.getters.buttonPermissions

    if (value && !permissions.includes(value)) {
      el.parentNode && el.parentNode.removeChild(el)
    }
  }
})

// 使用
<el-button v-permission="'system:user:add'">新增</el-button>
```

---

## 六、接口汇总表

### API ACL 权限管理接口

| 接口路径 | 方法 | 功能 | 权限标识 |
|---------|------|------|---------|
| /api/v1/api-resources | POST | 创建 API 资源 | system:api-resource:create |
| /api/v1/api-resources | PUT | 更新 API 资源 | system:api-resource:update |
| /api/v1/api-resources | DELETE | 删除 API 资源 | system:api-resource:delete |
| /api/v1/api-resources/detail | GET | 查询 API 资源详情 | system:api-resource:query |
| /api/v1/api-resources | GET | 分页查询 API 资源列表 | system:api-resource:query |
| /api/v1/api-resources/modules | GET | 获取所有模块列表 | system:api-resource:query |
| /api/v1/roles/api-permissions | PUT | 为角色分配 API 权限 | system:role:assign-permission |
| /api/v1/roles/api-permissions | GET | 查询角色的 API 权限 | system:role:query-permission |
| /api/v1/api-resources/batch-import | POST | 批量导入 API 资源 | system:api-resource:create |

### 菜单和按钮权限管理接口

| 接口路径 | 方法 | 功能 | 权限标识 | 状态 |
|---------|------|------|---------|------|
| /api/v1/menus | POST | 创建菜单资源 | system:menu:create | ✅ 有效 |
| /api/v1/menus | PUT | 更新菜单资源 | system:menu:update | ✅ 有效 |
| /api/v1/menus | DELETE | 删除菜单资源 | system:menu:delete | ✅ 有效 |
| /api/v1/menus/detail | GET | 查询菜单资源详情 | system:menu:query | ✅ 有效 |
| /api/v1/menus/tree | GET | 获取菜单树 | system:menu:query | ✅ 有效 |
| /api/v1/user/menus | GET | 获取用户的菜单树和按钮权限 | 需要登录 | ✅ 有效 |
| /api/v1/roles/menu-permissions | PUT | 为角色分配菜单权限 | system:role:assign-permission | ✅ 有效 |
| /api/v1/roles/menu-permissions | GET | 查询角色的菜单权限 | system:role:query-permission | ✅ 有效 |
| /api/v1/user/check-button-permission | POST | 检查用户按钮权限 | 需要登录 | ❌ 已废弃 |
| /api/v1/user/buttons/check | POST | 批量检查按钮权限 | 需要登录 | ❌ 已废弃 |

### 数据权限管理接口

| 接口路径 | 方法 | 功能 | 权限标识 |
|---------|------|------|---------|
| /api/v1/data-permission-rules | POST | 创建数据权限规则 | system:data-permission:create |
| /api/v1/data-permission-rules | PUT | 更新数据权限规则 | system:data-permission:update |
| /api/v1/data-permission-rules | DELETE | 删除数据权限规则 | system:data-permission:delete |
| /api/v1/data-permission-rules | GET | 分页查询数据权限规则 | system:data-permission:query |
| /api/v1/roles/data-permissions | PUT | 为角色绑定数据权限 | system:role:assign-data-permission |
| /api/v1/roles/data-permissions | GET | 查询角色的数据权限绑定 | system:role:query-data-permission |
| /api/v1/roles/data-permissions/custom | POST | 配置自定义数据权限 | system:role:assign-data-permission |
| /api/v1/user/data-permission-sql | POST | 获取用户的数据权限 SQL | 需要登录 |

### 统一权限查询接口

| 接口路径 | 方法 | 功能 | 权限标识 |
|---------|------|------|---------|
| /api/v1/roles/all-permissions | GET | 获取角色的所有权限 | system:role:query-permission |
| /api/v1/roles/copy-permissions | POST | 复制角色权限 | system:role:copy-permission |
| /api/v1/user/buttons/check | POST | 批量检查按钮权限 | 需要登录 |

---

## 七、错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 10001 | 参数验证失败 |
| 10002 | 资源不存在 |
| 10003 | 资源已存在 |
| 10004 | 无权限访问 |
| 10005 | 角色不存在 |
| 10006 | 菜单不存在 |
| 10007 | 父级菜单不存在 |
| 10008 | 菜单类型错误 |
| 10009 | 存在子菜单，禁止删除 |
| 10010 | API 资源的 path 和 method 重复 |
| 20001 | 数据库操作失败 |
| 20002 | Casbin 策略同步失败 |
| 30001 | 未登录或令牌过期 |
| 30002 | 租户 ID 无效 |

---

## 八、注意事项

### 8.1 权限标识命名规范
- API 权限：`{module}:{resource}:{action}`，如 `system:user:create`
- 菜单权限：`{module}:{resource}:{action}`，如 `system:user:add`
- 按钮权限：`{module}:{resource}:{action}`，如 `system:user:delete`

### 8.2 数据权限使用场景
1. 在业务接口中，根据用户角色自动拼接数据权限的 SQL WHERE 条件
2. 将生成的 SQL 条件应用到业务查询中
3. 支持缓存用户的数据权限规则，避免每次查询都计算

### 8.3 前端集成建议
1. **登录流程**：
   - 登录成功后调用 `/api/v1/user/menus` 获取用户菜单树和按钮权限
   - 将 `menus` 存储到路由管理器，动态生成路由
   - 将 `button_permissions` 存储到 Vuex/Pinia，供权限指令使用

2. **按钮权限控制**：
   - 使用自定义权限指令（如 `v-permission`）控制按钮显示
   - 指令根据缓存的 `button_permissions` 判断是否显示按钮
   - **无需**调用额外的API检查权限

3. **后端安全保障**：
   - 后端接口权限由 Casbin 中间件自动校验
   - 前端权限控制仅用于优化UI显示，不影响安全性
   - 即使前端被绕过，后端也会拦截未授权请求

### 8.4 性能优化建议
1. **菜单权限和按钮权限缓存**：
   - 使用 Redis 缓存用户的菜单树和按钮权限列表
   - 缓存键格式：`user:menus:{user_id}`、`user:buttons:{user_id}`
   - 过期时间：30分钟（权限变更时可主动清除）

2. **API权限缓存**：
   - API 权限建议使用 Redis 缓存
   - 缓存键格式：`user:api_permissions:{user_id}`

3. **数据权限SQL缓存**：
   - 数据权限 SQL 建议缓存到 Redis
   - 缓存键格式：`user:data_permission:{resource_type}:{user_id}`
   - 过期时间根据业务需求设置

4. **权限变更处理**：
   - 当角色权限被修改时，清除该角色下所有用户的权限缓存
   - 支持用户手动刷新权限（可选实现刷新接口）

### 8.5 多租户注意事项
1. 所有接口都需要在请求头中携带 `X-Tenant-ID`
2. 所有数据查询都需要自动添加租户 ID 过滤条件
3. 权限资源（API 资源、菜单、数据权限规则）都按租户隔离
4. Casbin 策略的 `dom` 字段存储租户 ID
