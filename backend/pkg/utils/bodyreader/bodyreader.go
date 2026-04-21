package bodyreader

import (
	"bytes"
	"io"
	"strings"
)

// SanitizeParams 对敏感字段进行脱敏处理（替换为 ***）
func SanitizeParams(params string) string {
	if params == "" {
		return ""
	}

	// 敏感字段列表（小写）
	sensitiveFields := []string{
		"password", "passwd", "pwd",
		"old_password", "new_password",
		"secret", "token", "access_token", "refresh_token",
		"api_key", "apikey", "api-key",
		"phone", "mobile", "telephone",
		"id_card", "idcard",
	}

	lowerParams := strings.ToLower(params)
	for _, field := range sensitiveFields {
		// 查找 "field": 模式
		target := `"` + field + `"`
		idx := 0
		for {
			pos := strings.Index(lowerParams[idx:], target)
			if pos == -1 {
				break
			}
			actualPos := idx + pos
			// 找到冒号后的值
			colonPos := strings.Index(params[actualPos:], ":")
			if colonPos != -1 {
				valueStart := actualPos + colonPos + 1
				// 跳过空白
				for valueStart < len(params) && (params[valueStart] == ' ' || params[valueStart] == '\t') {
					valueStart++
				}
				if valueStart < len(params) && params[valueStart] == '"' {
					// 字符串值，找到结束引号
					valueEnd := valueStart + 1
					for valueEnd < len(params) && params[valueEnd] != '"' {
						if params[valueEnd] == '\\' && valueEnd+1 < len(params) {
							valueEnd += 2
						} else {
							valueEnd++
						}
					}
					if valueEnd < len(params) {
						// 替换为 "***"
						params = params[:valueStart+1] + `***` + params[valueEnd:]
						// 更新 lowerParams
						lowerParams = strings.ToLower(params)
					}
				}
			}
			idx = actualPos + len(target)
		}
	}

	return params
}

// ReadBodyString 读取 io.ReadCloser 并返回字符串
// 同时返回一个新的 io.ReadCloser 用于恢复
func ReadBodyString(body io.ReadCloser) (string, io.ReadCloser) {
	if body == nil {
		return "", nil
	}

	bodyBytes, err := io.ReadAll(body)
	body.Close()
	if err != nil {
		return "", nil
	}

	return string(bodyBytes), io.NopCloser(bytes.NewBuffer(bodyBytes))
}
