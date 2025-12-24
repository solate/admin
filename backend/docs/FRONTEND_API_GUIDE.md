# 租户成员管理 - 前端 API 开发指南

## 概述

租户成员管理功能允许租户管理员添加、管理租户成员，并分配角色。管理员添加成员后会获得初始密码，需要通过安全方式传递给用户。

---

## 基础信息

- **Base URL**: `/api/v1/tenant-members`
- **认证方式**: Bearer Token (JWT)
- **权限要求**: 需要租户管理员权限
- **租户隔离**: 自动从当前登录用户的上下文中获取租户 ID

---

## API 接口列表

### 1. 添加租户成员

创建新用户并添加到当前租户，自动生成初始密码。

**请求**
```http
POST /api/v1/tenant-members
Authorization: Bearer {token}
Content-Type: application/json

{
  "username": "zhangsan",           // 必填，用户名（全局唯一）
  "name": "张三",                   // 必填，姓名/昵称
  "phone": "13800138000",           // 可选，手机号
  "email": "zhangsan@example.com",  // 可选，邮箱
  "role_ids": [                    // 必填，角色ID列表（至少一个）
    "role-id-1",
    "role-id-2"
  ]
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "zhangsan",
    "name": "张三",
    "initial_password": "aB3xY7kP9mQ2wL4n",  // ⚠️ 重要：初始密码只返回一次
    "tenant_id": "tenant-id-123",
    "role_ids": ["role-id-1", "role-id-2"]
  }
}
```

**重要提示**:
- `initial_password` 字段**只会在创建时返回一次**，后续无法再获取
- 前端需要以**显眼的方式**展示此密码，提醒管理员复制或记录
- 建议使用弹窗或专门的确认页面展示密码

---

### 2. 获取租户成员列表

分页获取当前租户的所有成员，支持关键词搜索和状态筛选。

**请求**
```http
GET /api/v1/tenant-members?page=1&page_size=10&keyword=张&status=1
Authorization: Bearer {token}
```

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 10，最大 100 |
| keyword | string | 否 | 关键词搜索（匹配用户名或姓名） |
| status | int | 否 | 状态筛选：1-正常，2-禁用 |

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "page": 1,
    "page_size": 10,
    "total": 25,
    "total_page": 3,
    "list": [
      {
        "user_id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "zhangsan",
        "name": "张三",
        "phone": "13800138000",
        "email": "zhangsan@example.com",
        "status": 1,
        "role_ids": ["role-id-1", "role-id-2"],
        "first_login": true,           // 是否首次登录
        "last_login_time": 0,          // 最后登录时间（0表示未登录）
        "created_at": 1703123456789
      }
    ]
  }
}
```

---

### 3. 移除租户成员

从当前租户中移除成员（仅移除租户关联，不删除用户）。

**请求**
```http
POST /api/v1/tenant-members/remove
Authorization: Bearer {token}
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

或使用路径参数方式：
```http
DELETE /api/v1/tenant-members/{user_id}
Authorization: Bearer {token}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 4. 更新成员角色

更新租户成员的角色列表。

**请求**
```http
PUT /api/v1/tenant-members/roles
Authorization: Bearer {token}
Content-Type: application/json

{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "role_ids": [                    // 新的角色列表（至少一个）
    "role-id-1",
    "role-id-3"
  ]
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "role_ids": ["role-id-1", "role-id-3"]
  }
}
```

---

## 状态码说明

| Code | Message | 说明 |
|------|---------|------|
| 0 | success | 成功 |
| 1001 | 用户名已存在 | 创建成员时用户名冲突 |
| 1002 | 角色不存在 | 指定的角色ID无效 |
| 1003 | 成员不存在 | 用户不在当前租户中 |
| 1004 | 角色ID列表不能为空 | 至少需要分配一个角色 |
| 401 | 未认证 | Token 无效或过期 |
| 403 | 无权限 | 非租户管理员 |

---

## 前端实现建议

### 1. 成员列表页面

**推荐布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  租户成员管理                                [+ 添加成员]    │
├─────────────────────────────────────────────────────────────┤
│  搜索: [____________] 状态: [全部 ▼]                         │
├─────────────────────────────────────────────────────────────┤
│  姓名       用户名      手机号     邮箱        状态    操作  │
│  ───────────────────────────────────────────────────────────│
│  张三       zhangsan    138...    zhang...    [正常]  编辑  │
│                                           [首次]  删除      │
│  李四       lisi        139...    lisi@...    [正常]  编辑  │
│                                                      删除      │
└─────────────────────────────────────────────────────────────┘
```

**字段显示建议**:
- `first_login` 为 `true` 时，显示"首次登录"标识，提醒管理员通知用户修改密码
- `last_login_time` 为 `0` 或 `null` 时，显示"未登录"
- 状态字段：1 显示"正常"（绿色），2 显示"禁用"（灰色）

### 2. 添加成员对话框

**实现要点**:

1. **表单验证**:
   - 用户名：必填，全局唯一（提交后由后端验证）
   - 姓名：必填
   - 手机号：可选，格式验证
   - 邮箱：可选，格式验证
   - 角色：必选，至少选择一个

2. **成功后处理**:
   - 弹出专门的"初始密码展示"对话框
   - 密码以**大字体、高亮**显示
   - 提供"复制密码"按钮
   - 提供"我已记下密码"确认按钮
   - 确认后关闭对话框并刷新列表

**初始密码展示对话框示例**:
```
┌─────────────────────────────────────────────┐
│        成员添加成功！                        │
│                                             │
│  请将以下凭据安全地传递给用户：              │
│                                             │
│  用户名: zhangsan                           │
│  初始密码: [aB3xY7kP9mQ2wL4n]  [复制]       │
│                                             │
│  ⚠️ 此密码只显示一次，请立即记录！           │
│                                             │
│  [我已记下密码]                              │
└─────────────────────────────────────────────┘
```

### 3. 编辑角色对话框

**实现要点**:
- 加载当前用户的所有角色
- 多选框形式展示可用角色
- 预选中用户当前拥有的角色
- 至少选择一个角色的验证

### 4. 删除确认

**实现要点**:
- 显示要删除成员的姓名和用户名
- 明确提示"此操作将移除该成员的租户访问权限"
- 提供确认和取消按钮

---

## 错误处理建议

### 用户名已存在
```typescript
// 提示用户换一个用户名
message.error('该用户名已被使用，请使用其他用户名')
```

### 角色不存在
```typescript
// 可能是角色已被删除，刷新角色列表
message.error('所选角色不存在，请刷新页面后重试')
// 自动刷新角色列表并重新打开对话框
```

### 成员不存在
```typescript
// 成员可能已被删除，刷新列表
message.error('该成员不存在，可能已被移除')
// 刷新成员列表
```

---

## 完整示例（TypeScript + React）

```typescript
// API 定义
interface TenantMember {
  user_id: string;
  username: string;
  name: string;
  phone: string;
  email: string;
  status: number;
  role_ids: string[];
  first_login: boolean;
  last_login_time: number;
  created_at: number;
}

interface AddMemberRequest {
  username: string;
  name: string;
  phone?: string;
  email?: string;
  role_ids: string[];
}

interface AddMemberResponse {
  user_id: string;
  username: string;
  name: string;
  initial_password: string;  // ⚠️ 只返回一次
  tenant_id: string;
  role_ids: string[];
}

// API 调用
const addTenantMember = async (data: AddMemberRequest): Promise<AddMemberResponse> => {
  const response = await fetch('/api/v1/tenant-members', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${getToken()}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  const result = await response.json();
  if (result.code !== 0) {
    throw new Error(result.message);
  }
  return result.data;
};

// 组件使用
const handleAddMember = async (values: AddMemberRequest) => {
  try {
    const result = await addTenantMember(values);

    // 显示初始密码对话框
    setPasswordDialog({
      visible: true,
      username: result.username,
      password: result.initial_password,
    });

    // 刷新列表
    fetchMemberList();
  } catch (error) {
    message.error(error.message);
  }
};

// 初始密码对话框组件
const PasswordDialog: React.FC<Props> = ({ username, password }) => {
  const copyPassword = () => {
    navigator.clipboard.writeText(password);
    message.success('密码已复制到剪贴板');
  };

  return (
    <Modal open={visible} title="成员添加成功" footer={null}>
      <p>请将以下凭据安全地传递给用户：</p>
      <div style={{ margin: '20px 0' }}>
        <div>用户名: <strong>{username}</strong></div>
        <div style={{ marginTop: '10px' }}>
          初始密码:
          <Input
            value={password}
            readOnly
            style={{ width: '200px', margin: '0 10px', fontSize: '18px', fontWeight: 'bold' }}
          />
          <Button onClick={copyPassword}>复制</Button>
        </div>
      </div>
      <Alert type="warning" message="此密码只显示一次，请立即记录！" />
      <Button type="primary" onClick={onClose} style={{ marginTop: '20px' }}>
        我已记下密码
      </Button>
    </Modal>
  );
};
```

---

## 给前端的提示词（AI 辅助开发）

你可以使用以下提示词让 AI 帮你生成前端代码：

```
请帮我实现一个租户成员管理页面，包含以下功能：

1. 成员列表展示
   - 表格展示：姓名、用户名、手机号、邮箱、状态、操作
   - 支持关键词搜索（姓名/用户名）
   - 支持状态筛选（全部/正常/禁用）
   - 分页功能

2. 添加成员功能
   - 表单字段：用户名（必填）、姓名（必填）、手机号（可选）、邮箱（可选）、角色多选（必选）
   - 提交成功后弹出专门的对话框展示初始密码
   - 初始密码对话框包含：用户名、初始密码（大字体高亮显示）、复制密码按钮、确认按钮
   - 重要提示：初始密码只显示一次

3. 编辑角色功能
   - 弹出对话框，多选框形式展示可用角色
   - 预选中用户当前角色
   - 至少选择一个角色

4. 删除成员功能
   - 确认对话框，显示成员姓名
   - 提示"此操作将移除该成员的租户访问权限"

API 接口：
- GET /api/v1/tenant-members - 获取成员列表
- POST /api/v1/tenant-members - 添加成员
- PUT /api/v1/tenant-members/roles - 更新成员角色
- DELETE /api/v1/tenant-members/{user_id} - 删除成员

响应数据结构参考文档中的示例。

使用 React + TypeScript + Ant Design 实现。
```

---

## 补充说明

### 角色列表获取

角色管理接口参考：
- `GET /api/v1/roles` - 获取当前租户的所有角色

### 首次登录检测

通过 `first_login` 字段判断：
- `true` - 用户从未登录过，首次登录时需强制修改密码
- `false` - 用户已登录过

### 时间戳处理

所有时间字段均为 Unix 毫秒时间戳，前端需要转换：
```typescript
const date = new Date(last_login_time);
```

---

## 联系方式

如有疑问，请联系后端开发团队。
