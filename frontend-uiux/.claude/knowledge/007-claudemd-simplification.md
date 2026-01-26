# CLAUDE.md 精简优化

> 问题描述：CLAUDE.md 文件内容过多，包含大量详细实现说明，影响快速查阅

---

## 精简目标

将 CLAUDE.md 从 520+ 行精简到约 180 行，保留核心信息，将详细内容移至知识库。

---

## 精简前后对比

| 项目 | 精简前 | 精简后 |
|------|--------|--------|
| 行数 | ~520 行 | ~180 行 |
| 项目结构 | 详细展开所有目录 | 简化为主要目录 |
| 国际化 | 详细说明 | 链接到知识库 |
| 多租户架构 | 代码示例 | 链接到知识库 |
| 设计系统 | 颜色/字体详情 | 链接到知识库 |
| 核心架构 | 详细代码示例 | 表格概述 + 链接 |
| 路由结构 | 完整路由列表 | 删除 |
| 配置文件 | 详细说明 | 删除 |

---

## 新建知识库文档

| 文档 | 说明 | 内容 |
|------|------|------|
| `multi-tenant-architecture.md` | 多租户架构 | 租户隔离、上下文管理、API 拦截、状态管理 |
| `design-system.md` | 设计系统 | 颜色系统、字体、Glassmorphism、组件规范 |
| `007-claudemd-simplification.md` | 本次精简记录 | - |

---

## 保留在 CLAUDE.md 的内容

- ✅ 语言规范
- ✅ 项目概述（简短）
- ✅ 技术栈表格
- ✅ 快速开始命令
- ✅ 环境变量
- ✅ 项目结构（简化版）
- ✅ 目录设计原则
- ✅ 开发规范（命名、代码风格）
- ✅ 常见任务（表格形式）
- ✅ 核心架构概述（表格 + 链接）
- ✅ 注意事项
- ✅ 知识库导航

---

## 移至知识库的内容

| 内容 | 原章节 | 现位置 |
|------|--------|--------|
| 详细项目结构 | 项目结构 | [006-project-structure-optimization.md](006-project-structure-optimization.md) |
| 国际化详解 | 国际化 (i18n) | [i18n-implementation.md](i18n-implementation.md) |
| 多租户架构 | 多租户架构 | [multi-tenant-architecture.md](multi-tenant-architecture.md) |
| 设计系统 | 设计系统 | [design-system.md](design-system.md) |
| Element Plus 集成 | 核心架构 | [design-system.md](design-system.md) |
| API 服务层详解 | 核心架构 | [006-project-structure-optimization.md](006-project-structure-optimization.md) |
| 路由结构 | 核心架构 | 删除（可从代码查看） |
| 配置文件 | 配置文件 | 删除（可从代码查看） |

---

## 相关文件

- `CLAUDE.md` - 精简后的主文档
- `.claude/knowledge/README.md` - 更新索引
- `.claude/knowledge/multi-tenant-architecture.md` - 新建
- `.claude/knowledge/design-system.md` - 新建

---

| 状态 | ✅ 已完成 |
|------|----------|
| 日期 | 2026-01-26 |
