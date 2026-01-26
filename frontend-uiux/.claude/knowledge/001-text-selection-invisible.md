# 文本选中后不可见问题

> 问题描述：当用户用鼠标滑过（选中）输入框或页面中的文字时，文字变成白色不可见。

---

## 根本原因

`src/styles/reset.css` 中的 `::selection` 规则使用了 `var(--color-primary)` CSS 变量：

```css
::selection {
  background-color: var(--color-primary);  /* 问题所在 */
  color: white;
}
```

`--color-primary` 在 `src/styles/index.css` 中定义为 `37 99 235`（Tailwind 的 RGB 通道格式），而不是标准的 `rgb(37, 99, 235)` 格式。浏览器无法正确解析，导致选中背景色无效，白色文字在白色背景上不可见。

---

## 解决方案

将 `::selection` 规则改为使用标准的 `rgb()` 函数格式：

```css
/* src/styles/reset.css */
::selection {
  background-color: rgb(37, 99, 235);  /* ✅ 使用标准格式 */
  color: white;
}

::-moz-selection {
  background-color: rgb(37, 99, 235);
  color: white;
}
```

---

## 相关文件

- `src/styles/reset.css`
- `src/styles/index.css`

---

## 调试技巧

```javascript
// 检查 ::selection 样式
const computed = window.getComputedStyle(document.body, '::selection');
return {
  backgroundColor: computed.backgroundColor,
  color: computed.color
};
```

---

| 状态 | ✅ 已修复 |
|------|----------|
| 修复日期 | 2025-01-23 |
