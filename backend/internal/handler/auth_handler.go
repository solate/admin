package handler

// // AuthHandler JWT 认证处理器
// // 提供登录、刷新、登出等接口的处理函数
// type AuthHandler struct {
// 	config  *config.Config
// 	manager *jwt.JWTManager
// }

// // NewAuthHandler 创建认证处理器
// func NewAuthHandler(manager *jwt.JWTManager) *AuthHandler {
// 	return &AuthHandler{
// 		manager: manager,
// 	}
// }

// // LoginRequest 登录请求
// type LoginRequest struct {
// 	Username string `json:"username" binding:"required"`
// 	Password string `json:"password" binding:"required"`
// }

// // LoginResponse 登录响应
// type LoginResponse struct {
// 	AccessToken  string `json:"access_token"`
// 	RefreshToken string `json:"refresh_token"`
// 	ExpiresIn    int64  `json:"expires_in"`
// 	TokenType    string `json:"token_type"`
// }

// // Login 处理登录请求
// // 实际使用中应该先验证用户名和密码，这里仅作示例
// // 生成并返回 token pair
// func (h *AuthHandler) Login(c *gin.Context) {
// 	var req LoginRequest
// 	if err := c.BindJSON(&req); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// TODO: 验证用户名和密码
// 	// 这里应该调用你的用户服务进行身份验证
// 	// 演示代码假设验证通过

// 	// 生成 token 对（demo 数据）
// 	tokenPair, err := h.manager.GenerateTokenPair(
// 		c.Request.Context(),
// 		"tenant-001", // 租户 ID
// 		"user-123",   // 用户 ID
// 		"role-admin", // 角色 ID
// 	)
// 	if err != nil {
// 		c.JSON(500, gin.H{"error": "failed to generate token"})
// 		return
// 	}

// 	c.JSON(200, LoginResponse{
// 		AccessToken:  tokenPair.AccessToken,
// 		RefreshToken: tokenPair.RefreshToken,
// 		ExpiresIn:    h.config.JWT.AccessExpire,
// 		TokenType:    "Bearer",
// 	})
// }

// // RefreshRequest 刷新请求
// type RefreshRequest struct {
// 	RefreshToken string `json:"refresh_token" binding:"required"`
// }

// // Refresh 处理刷新 token 请求
// func (h *AuthHandler) Refresh(c *gin.Context) {
// 	var req RefreshRequest
// 	if err := c.BindJSON(&req); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	tokenPair, err := h.manager.RefreshToken(c.Request.Context(), req.RefreshToken)
// 	if err != nil {
// 		c.JSON(401, gin.H{"error": "invalid refresh token"})
// 		return
// 	}

// 	c.JSON(200, LoginResponse{
// 		AccessToken:  tokenPair.AccessToken,
// 		RefreshToken: tokenPair.RefreshToken,
// 		ExpiresIn:    h.config.JWT.AccessExpire,
// 		TokenType:    "Bearer",
// 	})
// }

// // Logout 处理登出请求
// func (h *AuthHandler) Logout(c *gin.Context) {
// 	claims := GetClaims(c)
// 	if claims == nil {
// 		c.JSON(401, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	// 撤销当前 token
// 	if err := h.manager.RevokeToken(c.Request.Context(), claims.TokenID); err != nil {
// 		c.JSON(500, gin.H{"error": "failed to logout"})
// 		return
// 	}

// 	c.JSON(200, gin.H{"message": "logged out successfully"})
// }

// // LogoutAll 处理跨设备登出请求
// func (h *AuthHandler) LogoutAll(c *gin.Context) {
// 	claims := GetClaims(c)
// 	if claims == nil {
// 		c.JSON(401, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	// 撤销用户所有会话
// 	if err := h.manager.RevokeAllUserTokens(c.Request.Context(), claims.TenantID, claims.UserID); err != nil {
// 		c.JSON(500, gin.H{"error": "failed to logout all sessions"})
// 		return
// 	}

// 	c.JSON(200, gin.H{"message": "all sessions logged out successfully"})
// }

// // GetProfile 获取当前用户信息
// func (h *AuthHandler) GetProfile(c *gin.Context) {
// 	claims := GetClaims(c)
// 	if claims == nil {
// 		c.JSON(401, gin.H{"error": "unauthorized"})
// 		return
// 	}

// 	c.JSON(200, gin.H{
// 		"user_id":    claims.UserID,
// 		"tenant_id":  claims.TenantID,
// 		"role_id":    claims.RoleID,
// 		"issued_at":  claims.IssuedAt,
// 		"expires_at": claims.ExpiresAt,
// 	})
// }

// // VerifyTokenRequest 验证 token 请求
// type VerifyTokenRequest struct {
// 	Token string `json:"token" binding:"required"`
// }

// // VerifyToken 验证 token 的有效性
// func (h *AuthHandler) VerifyToken(c *gin.Context) {
// 	var req VerifyTokenRequest
// 	if err := c.BindJSON(&req); err != nil {
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	claims, err := h.manager.VerifyAccessToken(c.Request.Context(), req.Token)
// 	if err != nil {
// 		c.JSON(401, gin.H{"error": "invalid token"})
// 		return
// 	}

// 	c.JSON(200, gin.H{
// 		"valid":     true,
// 		"user_id":   claims.UserID,
// 		"tenant_id": claims.TenantID,
// 		"role_id":   claims.RoleID,
// 	})
// }

// // SetupAuthRoutes 设置认证相关路由
// // 使用示例：
// //
// //	jwtConfig := jwt.NewConfigBuilder().
// //	    WithAccessSecret("secret").
// //	    WithAccessExpire(3600).
// //	    WithRefreshSecret("refresh-secret").
// //	    WithRefreshExpire(604800).
// //	    Build()
// //	manager := jwt.NewJWTManager(jwtConfig, store)
// //	handler := jwt.NewAuthHandler(manager)
// //	jwt.SetupAuthRoutes(router, handler)
// func SetupAuthRoutes(router *gin.Engine, handler *AuthHandler, middleware ...gin.HandlerFunc) {
// 	// 公开路由（无需认证）
// 	public := router.Group("/auth")
// 	{
// 		public.POST("/login", handler.Login)
// 		public.POST("/refresh", handler.Refresh)
// 		public.POST("/verify", handler.VerifyToken)
// 	}

// 	// 受保护路由（需要认证）
// 	protected := router.Group("/auth")
// 	protected.Use(middleware...)
// 	{
// 		protected.GET("/profile", handler.GetProfile)
// 		protected.POST("/logout", handler.Logout)
// 		protected.POST("/logout-all", handler.LogoutAll)
// 	}
// }

// // SetupAuthRoutesWithManager 快速设置认证路由（一步到位）
// // 这个函数结合了路由设置和中间件配置
// func SetupAuthRoutesWithManager(
// 	router *gin.Engine,
// 	manager *JWTManager,
// ) {
// 	handler := NewAuthHandler(manager)
// 	authMiddleware := GinAuthMiddleware(manager)
// 	SetupAuthRoutes(router, handler, authMiddleware)
// }
