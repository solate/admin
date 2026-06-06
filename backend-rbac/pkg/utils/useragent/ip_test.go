package useragent

import (
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For with public IP first",
			headers: map[string]string{
				"X-Forwarded-For": "8.8.8.8, 192.168.1.1",
			},
			expectedIP: "8.8.8.8",
		},
		{
			name: "X-Forwarded-For with private IP first, public IP second",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 8.8.8.8",
			},
			expectedIP: "8.8.8.8",
		},
		{
			name: "X-Real-IP with public IP",
			headers: map[string]string{
				"X-Real-IP": "1.1.1.1",
			},
			expectedIP: "1.1.1.1",
		},
		{
			name: "X-Real-IP with private IP, should fallback",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			remoteAddr: "8.8.8.8:12345",
			expectedIP: "8.8.8.8",
		},
		{
			name: "X-Client-IP with public IP",
			headers: map[string]string{
				"X-Client-IP": "114.114.114.114",
			},
			expectedIP: "114.114.114.114",
		},
		{
			name: "X-Client-IP with private IP, should fallback",
			headers: map[string]string{
				"X-Client-IP": "10.0.0.1",
			},
			remoteAddr: "223.5.5.5:12345",
			expectedIP: "223.5.5.5",
		},
		{
			name:       "Only RemoteAddr with public IP",
			remoteAddr: "8.8.4.4:12345",
			expectedIP: "8.8.4.4",
		},
		{
			name:       "Only RemoteAddr with private IP",
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "127.0.0.1", // fallback to default
		},
		{
			name: "All private IPs, should return default",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
				"X-Real-IP":       "172.16.0.1",
				"X-Client-IP":     "169.254.1.1",
			},
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "127.0.0.1",
		},
		{
			name: "Mixed private and public IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 8.8.8.8, 10.0.0.1",
			},
			expectedIP: "8.8.8.8",
		},
		{
			name:       "IPv6 public address",
			remoteAddr: "[2001:db8::1]:12345",
			expectedIP: "2001:db8::1",
		},
		{
			name:       "IPv6 loopback address",
			remoteAddr: "[::1]:12345",
			expectedIP: "127.0.0.1", // fallback to default
		},
		// 保留一些原来的基础测试用例
		{
			name: "X-Forwarded-For basic test",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			expectedIP: "127.0.0.1", // all private, fallback to default
		},
		{
			name: "X-Real-IP basic test",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			expectedIP: "127.0.0.1", // private IP, fallback to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			ip := GetClientIP(req)
			if ip != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestIsInternalIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Public IP - Google DNS", "8.8.8.8", false},
		{"Public IP - Cloudflare DNS", "1.1.1.1", false},
		{"Private IP - 192.168.x.x", "192.168.1.1", true},
		{"Private IP - 10.x.x.x", "10.0.0.1", true},
		{"Private IP - 172.16.x.x", "172.16.0.1", true},
		{"Loopback IP", "127.0.0.1", true},
		{"IPv6 Loopback", "::1", true},
		{"Link Local - 169.254.x.x", "169.254.1.1", true},
		{"Invalid IP", "invalid", false},
		{"Empty string", "", false},
		{"Public IPv6", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInternalIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isInternalIP(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}
