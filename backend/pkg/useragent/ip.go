package useragent

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP 获取客户端公网IP地址，过滤内网IP
// 优先级：X-Forwarded-For > X-Real-IP > X-Client-IP > RemoteAddr
// 专门用于需要真实公网IP的场景，如第三方支付接口
func GetClientIP(r *http.Request) string {
	// 优先从X-Forwarded-For获取（适用于反向代理）
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For可能包含多个IP，取第一个公网IP
		ips := strings.Split(xForwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && !isInternalIP(ip) {
				return ip
			}
		}
	}

	// 从X-Real-IP获取（nginx等反向代理）
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" && !isInternalIP(xRealIP) {
		return xRealIP
	}

	// 从X-Client-IP获取
	xClientIP := r.Header.Get("X-Client-IP")
	if xClientIP != "" && !isInternalIP(xClientIP) {
		return xClientIP
	}

	// 最后从RemoteAddr获取
	remoteAddr := r.RemoteAddr
	if remoteAddr != "" {
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			// 如果没有端口，直接返回IP（如果是公网IP）
			if !isInternalIP(remoteAddr) {
				return remoteAddr
			}
		} else if !isInternalIP(ip) {
			return ip
		}
	}

	// 如果都获取不到公网IP，返回默认值
	return "127.0.0.1"
}

// isInternalIP 判断是否为内网IP
func isInternalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查是否为回环地址
	if ip.IsLoopback() {
		return true
	}

	// 检查是否为内网地址
	if ip.IsPrivate() {
		return true
	}

	// 检查是否为链路本地地址
	if ip.IsLinkLocalUnicast() {
		return true
	}

	return false
}
