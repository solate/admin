# 项目知识库 - 已知问题与解决方案

> 本目录记录项目中遇到的问题及解决方案，供后续调试参考。

---

## 问题索引

| ID | 问题 | 状态 | 日期 |
|----|------|------|------|
| [001](001-text-selection-invisible.md) | 文本选中后不可见 | ✅ 已修复 | 2025-01-23 |
| [002](002-element-plus-double-border.md) | Element Plus 输入框双层边框 | ✅ 已修复 | 2025-01-23 |
| [003](003-icon-library-migration.md) | 图标库迁移到 lucide-vue-next | ✅ 已完成 | 2025-01-23 |
| [004](004-settings-drawer-positioning.md) | 设置抽屉定位问题 | ✅ 已修复 | 2025-01-25 |
| [005](005-html-nesting-warning.md) | HTML 嵌套警告（p 元素嵌套） | ✅ 已修复 | 2025-01-25 |
| [006](006-project-structure-optimization.md) | 项目结构优化与主题切换修复 | ✅ 已完成 | 2025-01-26 |
| [007](007-claudemd-simplification.md) | CLAUDE.md 精简优化 | ✅ 已完成 | 2026-01-26 |
| [008](008-preferences-panel-enhancement.md) | 设置面板系统增强 | ✅ 已完成 | 2026-01-26 |

---

## 分类索引

### 样式问题
- [001 - 文本选中后不可见](001-text-selection-invisible.md) - `::selection` CSS 变量格式问题
- [002 - Element Plus 输入框双层边框](002-element-plus-double-border.md) - 使用 box-shadow inset

### 组件与库
- [003 - 图标库迁移到 lucide-vue-next](003-icon-library-migration.md) - 完整图标名称对照表

### Vue 问题
- [004 - 设置抽屉定位问题](004-settings-drawer-positioning.md) - 使用 Teleport 渲染到 body
- [005 - HTML 嵌套警告](005-html-nesting-warning.md) - p 元素内容模型限制

---

## 实现文档

| 文档 | 说明 |
|-----|------|
| [i18n-implementation.md](i18n-implementation.md) | 国际化实现方案 |
| [multi-tenant-architecture.md](multi-tenant-architecture.md) | 多租户架构详解 |
| [design-system.md](design-system.md) | 设计系统（颜色、字体、组件规范） |
| [006-project-structure-optimization.md](006-project-structure-optimization.md) | 项目结构优化与主题切换修复 |
| [007-claudemd-simplification.md](007-claudemd-simplification.md) | CLAUDE.md 精简记录 |
| [008-preferences-panel-enhancement.md](008-preferences-panel-enhancement.md) | 设置面板系统增强（外观/布局/通用设置） |

---

## 添加新问题

1. 在本目录创建新的 markdown 文件
2. 文件命名格式：`NNN-issue-name.md`（NNN 为递增序号）
3. 在本文件中添加索引条目
4. 更新对应的分类索引

---

## 文档模板

```markdown
# 问题标题

> 问题描述：简短描述用户看到的现象

---

## 根本原因
技术层面的根本原因分析

---

## 解决方案
具体的修复代码或配置

---

## 相关文件
- `path/to/file1`
- `path/to/file2`

---

## 调试技巧
可选的调试代码或工具使用方法

---

| 状态 | ✅ 已修复 |
|------|----------|
| 修复日期 | YYYY-MM-DD |
```
