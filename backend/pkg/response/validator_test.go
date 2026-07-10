package response

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"admin/pkg/xerr"
)

// testReq 覆盖 required / min / email 三个典型 tag，字段名用 json 下划线风格。
type testReq struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// TestGetResponse_ValidationError 验证校验错误经 getResponse 翻成中文、
// 字段名为 json tag、业务码为 xerr.ErrInvalidParams，且走的是官方 zh 翻译器（而非手写 switch）。
func TestGetResponse_ValidationError(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"password":"123"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var req testReq
	err := c.ShouldBindJSON(&req)
	if err == nil {
		t.Fatal("期望校验失败，实际 err 为 nil")
	}

	resp := getResponse(err)

	if resp.Code != xerr.ErrInvalidParams.Code() {
		t.Errorf("期望业务码 %d，实际 %d", xerr.ErrInvalidParams.Code(), resp.Code)
	}
	// 中文提示（zh 翻译器输出「为必填字段」而非英文 "required"）
	if !strings.Contains(resp.Message, "必填") {
		t.Errorf("期望包含中文「必填」，实际消息：%s", resp.Message)
	}
	// 字段名应为 json tag（user_name / email），而非 Go 字段名（UserName / Email）
	if !strings.Contains(resp.Message, "user_name") {
		t.Errorf("期望字段名为 json tag「user_name」，实际消息：%s", resp.Message)
	}
	if strings.Contains(resp.Message, "UserName") {
		t.Errorf("不应出现 Go 字段名「UserName」，实际消息：%s", resp.Message)
	}
	// min=6 触发的长度提示应为中文
	if !strings.Contains(resp.Message, "长度") && !strings.Contains(resp.Message, "6") {
		t.Errorf("期望 password 长度中文提示，实际消息：%s", resp.Message)
	}
	t.Logf("翻译后的中文校验消息：%s", resp.Message)
}

// TestGetResponse_AppError 验证 *xerr.AppError 按其业务码与消息渲染。
func TestGetResponse_AppError(t *testing.T) {
	resp := getResponse(xerr.New(2100, "用户名或密码错误"))
	if resp.Code != 2100 || resp.Message != "用户名或密码错误" {
		t.Errorf("期望 {2100, 用户名或密码错误}，实际 {%d, %s}", resp.Code, resp.Message)
	}
}

// TestGetResponse_WrappedAppError 验证被 %w 包装的 AppError 仍能被 errors.As 穿透识别，
// 拿到内层业务码而非掉进兜底——这是用 errors.As 而非裸 type switch 的关键收益。
func TestGetResponse_WrappedAppError(t *testing.T) {
	wrapped := fmt.Errorf("service 调用失败: %w", xerr.New(2005, "保存失败"))
	resp := getResponse(wrapped)
	if resp.Code != 2005 || resp.Message != "保存失败" {
		t.Errorf("期望穿透 %%w 拿到 {2005, 保存失败}，实际 {%d, %s}", resp.Code, resp.Message)
	}
}

// TestGetResponse_Fallback 验证非 AppError、非校验错误走兜底，不泄露内部细节。
func TestGetResponse_Fallback(t *testing.T) {
	resp := getResponse(errFake("底层数据库连接串 postgres://user:pass@host"))
	if resp.Code != xerr.ErrInternal.Code() {
		t.Errorf("期望兜底码 %d，实际 %d", xerr.ErrInternal.Code(), resp.Code)
	}
	if resp.Message != "服务器内部错误" {
		t.Errorf("兜底消息应为「服务器内部错误」，实际：%s", resp.Message)
	}
	if strings.Contains(resp.Message, "postgres") {
		t.Errorf("兜底不应泄露内部错误细节，实际：%s", resp.Message)
	}
}

type errFake string

func (e errFake) Error() string { return string(e) }
