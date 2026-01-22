import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'

// 自定义主题变量 - 保持现有设计风格
const elPrimaryColor = '#2563eb' // primary-600
const elCtaColor = '#f97316' // cta-500
const elSuccessColor = '#22c55e' // success-600
const elWarningColor = '#f59e0b' // warning-600
const elErrorColor = '#ef4444' // error-600
const elInfoColor = '#0ea5e9' // info-600

export function setupElementPlus(app) {
  // 注册所有 Element Plus 图标
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }

  // 设置 Element Plus CSS 变量以匹配现有设计系统
  const style = document.createElement('style')
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

      --el-border-radius-base: 8px;
      --el-border-radius-small: 4px;
      --el-border-radius-round: 20px;

      --el-font-size-base: 14px;
      --el-font-size-small: 12px;
      --el-font-size-large: 16px;
    }

    /* 暗黑模式适配 */
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

    /* 保持与现有卡片样式一致 */
    .el-card {
      border-radius: 12px;
      box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.05), 0 1px 2px -1px rgb(0 0 0 / 0.05);
      border: 1px solid rgb(226 232 240 / 1);
    }

    .dark .el-card {
      background-color: #1e293b;
      border-color: #334155;
    }

    /* 按钮样式优化 */
    .el-button {
      border-radius: 8px;
      font-weight: 500;
      transition: all 0.2s;
    }

    /* 输入框样式优化 */
    .el-input__wrapper {
      border-radius: 8px;
      transition: all 0.2s;
    }

    /* 表格样式优化 */
    .el-table {
      border-radius: 8px;
      overflow: hidden;
    }

    .el-table th.el-table__cell {
      background-color: rgb(248 250 252 / 1);
    }

    .dark .el-table th.el-table__cell {
      background-color: #1e293b;
    }

    /* 对话框样式优化 */
    .el-dialog {
      border-radius: 12px;
    }

    /* Switch 样式优化 */
    .el-switch {
      --el-switch-on-color: ${elPrimaryColor};
      --el-switch-off-color: #cbd5e1;
    }

    .dark .el-switch {
      --el-switch-off-color: #475569;
    }

    /* 保持滚动条样式一致 */
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
