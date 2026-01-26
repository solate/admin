# Element Plus 输入框双层边框问题

> 问题描述：使用 `<el-input>` 组件时，出现"双层边框"效果——外层 wrapper 和内部 input 各自都有边框，视觉上不协调。

---

## 根本原因

在 `src/plugins/element.ts` 中，`.el-input__wrapper` 和 `.el-input__inner` 都设置了边框相关样式，导致双层效果：

```css
/* 问题代码 */
.el-input__wrapper {
  border-radius: 8px;
}
.el-input__inner {
  border-radius: 8px;  /* 冗余的圆角 */
}
```

---

## 解决方案

使用 `box-shadow` inset 实现单层边框，移除内部 input 的边框样式：

```css
/* src/plugins/element.ts */
.el-input__wrapper {
  border-radius: 8px;
  box-shadow: 0 0 0 1px var(--el-border-color) inset;
  transition: box-shadow 0.2s;
}

.el-input__wrapper:hover {
  box-shadow: 0 0 0 1px var(--el-border-color-hover) inset;
}

.el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
}

.el-input__inner {
  box-shadow: none;
}
```

---

## 关键点

1. 使用 `box-shadow: 0 0 0 1px ... inset` 替代 `border`
2. 只在 wrapper 上设置边框效果
3. transition 只针对 `box-shadow`，避免 `all` 影响性能

---

## 相关文件

- `src/plugins/element.ts`
- `src/views/settings/SettingsView.vue`

---

## 调试技巧

```javascript
// 使用 Chrome DevTools MCP 调试输入框问题
const input = document.querySelector('input[type="email"]');
const styles = window.getComputedStyle(input);

return {
  color: styles.color,
  backgroundColor: styles.backgroundColor,
  webkitTextFillColor: styles.webkitTextFillColor,
  opacity: styles.opacity,
  border: styles.border,
  boxShadow: styles.boxShadow
};
```

---

| 状态 | ✅ 已修复 |
|------|----------|
| 修复日期 | 2025-01-23 |
