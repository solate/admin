package response

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhtrans "github.com/go-playground/validator/v10/translations/zh"
)

var trans ut.Translator // 包内全局翻译器

func init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)
	trans, _ = uni.GetTranslator("zh")
	_ = zhtrans.RegisterDefaultTranslations(v, trans)
	// 报错用 json 字段名（user_name）而非 Go 字段名（UserName）——防止暴露内部结构
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// translateValidationError 检测并翻译参数校验错误。
// 单一职责：只判断 err 是否为 validator.ValidationErrors 并翻成中文字符串，
// 不组装 Response、不关心业务码——码值与 Response 组装统一交给 getResponse。
// 返回 (中文消息, true) 如果是校验错误；("", false) 如果不是。
func translateValidationError(err error) (string, bool) {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return "", false // 不是校验错误
	}
	if trans == nil {
		return "参数校验失败", true // 翻译器未就绪，退化为通用提示
	}
	translated := verrs.Translate(trans)
	msgs := make([]string, 0, len(translated))
	for _, msg := range translated {
		msgs = append(msgs, msg)
	}
	return strings.Join(msgs, "; "), true
}
