export default {
  // 通用
  common: {
    appName: 'MultiTenant',
    search: '搜索',
    searchPlaceholder: '搜索...',
    addNew: '新建',
    help: '帮助',
    loading: '加载中...',
    submit: '提交',
    cancel: '取消',
    save: '保存',
    delete: '删除',
    edit: '编辑',
    view: '查看',
    back: '返回',
    next: '下一步',
    prev: '上一步',
    close: '关闭',
    confirm: '确认',
    backToHome: '返回首页',
    rememberMe: '记住我',
    forgotPassword: '忘记密码？'
  },

  // 导航/路由
  nav: {
    dashboard: '仪表盘',
    overview: '概览',
    tenants: '租户管理',
    services: '服务管理',
    users: '用户管理',
    business: '业务管理',
    analytics: '数据分析',
    settings: '系统设置',
    profile: '个人资料'
  },

  // 语言切换
  language: {
    title: '语言',
    zhCN: '简体中文',
    enUS: 'English'
  },

  // 登录页
  auth: {
    login: '登录',
    register: '注册',
    loggingIn: '登录中...',
    registering: '注册中...',
    loginToAccount: '登录到您的账户',
    noAccount: '还没有账户？',
    createAccount: '创建新账户',
    emailAddress: '邮箱地址',
    password: '密码',
    loginFailed: '登录失败，请检查邮箱和密码'
  },

  // 首页
  landing: {
    // 导航
    features: '功能特性',
    pricing: '价格方案',
    about: '关于我们',
    login: '登录',
    freeTrial: '免费试用',

    // Hero 区域
    heroTitle: '企业级多租户',
    heroTitleHighlight: 'SaaS 管理平台',
    heroSubtitle: '一站式多租户管理解决方案，帮助企业快速构建和部署 SaaS 应用。包含基础服务管理和业务模块，让您专注于核心业务创新。',
    startFreeTrial: '免费开始 14 天试用',
    viewDemo: '查看演示',
    productDemo: '产品演示视频',
    clickToView: '点击查看完整功能演示',

    // 统计数据
    stats: {
      customers: '企业客户',
      users: '活跃用户',
      availability: '服务可用性',
      support: '技术支持'
    },

    // 功能特性
    featuresTitle: '强大的功能特性',
    featuresSubtitle: '完整的多租户管理解决方案，从基础服务到业务模块，全方位满足您的需求',

    featureNames: {
      multiTenant: '多租户管理',
      basicServices: '基础服务集成',
      security: '数据安全',
      monitoring: '实时监控',
      highAvailability: '高可用架构',
      quickDeploy: '快速部署'
    },

    featureDescs: {
      multiTenant: '灵活的多租户架构，支持数据隔离和个性化配置，满足不同企业客户的需求。',
      basicServices: '云存储、消息队列、身份认证等开箱即用的基础服务，快速构建业务能力。',
      security: '企业级安全防护，包括数据加密、访问控制、审计日志等全方位安全保障。',
      monitoring: '全方位系统监控和数据分析，实时掌握服务状态和业务指标。',
      highAvailability: '分布式集群部署，支持弹性扩展，确保服务99.9%可用性保障。',
      quickDeploy: '一键部署和自动化运维，大幅降低运维成本，让您专注业务创新。'
    },

    // 价格方案
    pricingTitle: '灵活的价格方案',
    pricingSubtitle: '选择适合您企业的方案，随时可以升级',
    mostPopular: '最受欢迎',

    plans: {
      basic: '基础版',
      professional: '专业版',
      enterprise: '企业版'
    },

    planDescs: {
      basic: '适合小型团队和初创企业',
      professional: '适合成长型企业和中等规模团队',
      enterprise: '适合大型企业和特殊需求'
    },

    planPeriods: {
      perMonth: '/月'
    },

    pricing: {
      basic: {
        price: '299',
        features: [
          '最多 5 个租户',
          '基础数据隔离',
          '核心服务访问',
          '邮件支持',
          '99% SLA 保障'
        ]
      },
      professional: {
        price: '899',
        features: [
          '最多 20 个租户',
          '高级数据隔离',
          '全部服务访问',
          '优先技术支持',
          '99.9% SLA 保障',
          '自定义域名',
          'API 访问权限'
        ]
      },
      enterprise: {
        price: '定制',
        features: [
          '无限租户数',
          '私有化部署',
          '专属客户经理',
          '7x24 技术支持',
          '99.95% SLA 保障',
          '定制功能开发',
          '源码授权',
          '培训服务'
        ]
      }
    },

    planActions: {
      basic: '开始试用',
      professional: '立即购买',
      enterprise: '联系销售'
    },

    // CTA 区域
    ctaTitle: '准备好开始了吗？',
    ctaSubtitle: '立即注册，获得 14 天免费试用，无需信用卡',
    ctaButton: '免费开始使用',

    // 页脚
    footer: {
      product: '产品',
      resources: '资源',
      company: '公司',

      links: {
        features: '功能特性',
        pricing: '价格方案',
        changelog: '更新日志',
        docs: '文档中心',
        api: 'API 参考',
        help: '帮助中心',
        about: '关于我们',
        contact: '联系我们',
        privacy: '隐私政策'
      },

      copyright: '© 2024 MultiTenant. All rights reserved.',
      description: '企业级多租户 SaaS 管理平台，帮助企业快速构建和部署应用。'
    }
  }
}
