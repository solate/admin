package useragent

import (
	"net/http"

	"github.com/mssola/user_agent"
)

// ClientInfo 包含客户端的详细信息
type ClientInfo struct {
	IP           string // 客户端 IP 地址
	UserAgent    string // 原始 User-Agent 字符串
	Browser      string // 浏览器名称
	BrowserVer   string // 浏览器版本
	OS           string // 操作系统
	Device       string // 设备类型
	Proxy        string // 是否使用代理（通过 X-Forwarded-For 判断）
	Platform     string // 平台
	Localization string // 语言
}

// GetClientInfo 从 HTTP 请求中提取客户端信息
func GetClientInfo(r *http.Request) *ClientInfo {
	clientInfo := &ClientInfo{}

	// 获取客户端 IP
	clientInfo.IP = GetClientIP(r)

	// 获取 User-Agent 并解析
	clientInfo.UserAgent = r.UserAgent()
	ua := user_agent.New(clientInfo.UserAgent)
	clientInfo.Browser, clientInfo.BrowserVer = ua.Browser()
	clientInfo.OS = ua.OS()
	if ua.Mobile() {
		clientInfo.Device = "Mobile"
	} else {
		clientInfo.Device = "Desktop"
	}

	clientInfo.Platform = ua.Platform()
	clientInfo.Localization = ua.Localization()

	// 判断是否使用代理
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientInfo.Proxy = xff
	} else {
		clientInfo.Proxy = ""
	}

	return clientInfo
}
