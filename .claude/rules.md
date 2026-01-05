# Claude Code 工作规则

## 编译和运行规则

### Go 项目（Backend）
- **位置**：`backend/` 目录
- **编译/运行**：使用 `cd backend && make run` 或在 `backend/` 目录下执行 `make run`
- **禁止**：不要直接使用 `go build ./...`，因为 go.mod 在 `backend/` 子目录中

### 前端项目（Frontend）
- **位置**：`frontend/` 目录
- **类型检查**：使用 `cd frontend && npm run type-check`
- **开发模式**：使用 `cd frontend && npm run dev`

## 项目结构

```
admin/
├── backend/          # Go 后端（go.mod 在这里）
│   └── cmd/server/   # 主程序入口
├── frontend/         # Vue 前端
└── .claude/          # Claude 配置
```

## 常用命令

```bash
# 后端
cd backend && make run          # 运行后端
cd backend && make test         # 运行测试
cd backend && make build        # 构建后端

# 前端
cd frontend && npm run dev      # 开发模式
cd frontend && npm run build    # 构建前端
cd frontend && npm run type-check  # 类型检查
```
