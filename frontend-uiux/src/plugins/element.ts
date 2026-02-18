import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import type { App } from 'vue'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'

// Custom theme variables - maintain existing design style
const elPrimaryColor = '#2563eb' // primary-600
const elCtaColor = '#f97316' // cta-500
const elSuccessColor = '#22c55e' // success-600
const elWarningColor = '#f59e0b' // warning-600
const elErrorColor = '#ef4444' // error-600
const elInfoColor = '#0ea5e9' // info-600

export function setupElementPlus(app: App) {

  // Set Element Plus CSS variables to match existing design system
  const style = document.createElement('style')
  style.id = 'element-plus-theme-overrides' // 添加 ID 便于调试
  style.innerHTML = `
    :root {
      --el-color-primary: ${elPrimaryColor};
      --el-color-primary-light-3: #60a5fa;
      --el-color-primary-light-5: #93c5fd;
      --el-color-primary-light-7: #bfdbfe;
      --el-color-primary-light-8: #dbeafe;
      --el-color-primary-light-9: #eff6ff;
      --el-color-primary-dark-2: #1d4ed8;

      --el-color-success: ${elSuccessColor};
      --el-color-warning: ${elWarningColor};
      --el-color-danger: ${elErrorColor};
      --el-color-error: ${elErrorColor};
      --el-color-info: ${elInfoColor};

      --el-font-size-base: 14px;
      --el-font-size-small: 12px;
      --el-font-size-large: 16px;
    }

    /* Dark mode adaptation */
    .dark {
      --el-bg-color: #1e293b;
      --el-bg-color-page: #0f172a;
      --el-text-color-primary: #f1f5f9;
      --el-text-color-regular: #cbd5e1;
      --el-border-color: #334155;
      --el-border-color-light: #475569;
      --el-border-color-lighter: #64748b;
      --el-border-color-extra-light: #94a3b8;
      --el-border-color-dark: #1e293b;
      --el-border-color-darker: #0f172a;
      --el-fill-color: #334155;
      --el-fill-color-light: #475569;
      --el-fill-color-lighter: #64748b;
      --el-fill-color-extra-light: #94a3b8;
      --el-fill-color-dark: #1e293b;
      --el-fill-color-darker: #0f172a;

      --el-box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
      --el-box-shadow-light: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
      --el-box-shadow-base: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
    }

    /* Keep consistent with existing card styles - 使用最高优先级选择器 */
    body .el-card,
    html body .el-card,
    .el-card {
      border-radius: var(--border-radius) !important;
      box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.05), 0 1px 2px -1px rgb(0 0 0 / 0.05);
      border: 1px solid rgb(226 232 240 / 1);
    }

    .dark .el-card {
      background-color: #1e293b;
      border-color: #334155;
    }

    /* Button style optimization - 使用最高优先级选择器 */
    body .el-button,
    html body .el-button,
    .el-button,
    button.el-button,
    span.el-button {
      border-radius: var(--border-radius) !important;
      font-weight: 500;
      transition: all 0.2s;
    }

    /* 各种按钮类型 */
    .el-button--default,
    .el-button--primary,
    .el-button--success,
    .el-button--warning,
    .el-button--danger,
    .el-button--info {
      border-radius: var(--border-radius) !important;
    }

    /* 小按钮 */
    body .el-button--small,
    .el-button--small {
      border-radius: var(--el-border-radius-small) !important;
    }

    /* 圆形按钮 */
    body .el-button.is-circle,
    body .el-button.is-round,
    .el-button.is-circle,
    .el-button.is-round {
      border-radius: var(--el-border-radius-round) !important;
    }

    /* Input - 单层边框，简洁样式 */
    body .el-input__wrapper,
    html body .el-input__wrapper,
    .el-input__wrapper {
      border-radius: var(--border-radius) !important;
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

    /* 暗色模式 */
    .dark .el-input__wrapper {
      background-color: rgba(15, 23, 42, 0.8);
      box-shadow: 0 0 0 1px rgba(71, 85, 105, 0.8) inset;
    }

    .dark .el-input__wrapper:hover {
      box-shadow: 0 0 0 1px rgba(96, 165, 250, 0.6) inset;
    }

    .dark .el-input__wrapper.is-focus {
      box-shadow: 0 0 0 1px #60a5fa inset;
    }

    .dark .el-input__inner {
      color: #e2e8f0;
    }

    .dark .el-input__inner::placeholder {
      color: #64748b;
    }

    /* Table style optimization */
    html .el-table,
    :root .el-table {
      border-radius: var(--border-radius) !important;
      overflow: hidden;
    }

    .el-table th.el-table__cell {
      background-color: rgb(248 250 252 / 1);
    }

    .dark .el-table th.el-table__cell {
      background-color: #1e293b;
    }

    /* Dialog style optimization */
    html .el-dialog,
    :root .el-dialog {
      border-radius: var(--border-radius) !important;
    }

    /* Select 下拉框 */
    html .el-select__wrapper,
    :root .el-select__wrapper {
      border-radius: var(--border-radius) !important;
    }

    /* Tag 标签 */
    html .el-tag,
    :root .el-tag {
      border-radius: var(--el-border-radius-small) !important;
    }

    /* Message 消息提示 */
    html .el-message,
    :root .el-message {
      border-radius: var(--border-radius) !important;
    }

    /* Notification 通知 */
    html .el-notification,
    :root .el-notification {
      border-radius: var(--border-radius) !important;
    }

    /* Popover 弹出框 */
    html .el-popover.el-popper,
    :root .el-popover.el-popper {
      border-radius: var(--border-radius) !important;
    }

    /* Dropdown 下拉菜单 */
    html .el-dropdown-menu,
    :root .el-dropdown-menu {
      border-radius: var(--border-radius) !important;
    }

    /* Drawer 抽屉 */
    html .el-drawer,
    :root .el-drawer {
      border-radius: var(--border-radius) !important;
    }

    /* Switch style optimization */
    .el-switch {
      --el-switch-on-color: ${elPrimaryColor};
      --el-switch-off-color: #cbd5e1;
    }

    .dark .el-switch {
      --el-switch-off-color: #475569;
    }

    /* Keep scrollbar styles consistent */
    .el-scrollbar__bar {
      opacity: 0.3;
    }

    .el-scrollbar__bar:hover {
      opacity: 0.5;
    }
  `
  document.head.appendChild(style)

  app.use(ElementPlus, {
    locale: zhCn,
    size: 'default',
    zIndex: 3000
  })

  return {
    zhCn,
    en
  }
}

export { zhCn, en }
