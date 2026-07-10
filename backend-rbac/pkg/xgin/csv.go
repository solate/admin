package xgin

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"admin/pkg/utils/csv"
)

// maxUploadSize 上传 CSV 的大小上限。编码识别（BOM 嗅探 + utf8.Valid）天然要把整份内容读进内存，
// 无法干净流式，故在入口卡大小上限，防未授信上传打爆内存。按业务需要调整。
const maxUploadSize = 10 << 20 // 10MB

// ExportCSV 一行导出：headers + items + 行映射 → 带 BOM 的 CSV 下载，中文文件名安全。
// 屏蔽 csv 构造 / BOM / Content-Disposition 编码 / 流式写 的全部细节。
//
//	xgin.ExportCSV(c, "登录日志.csv", headers, resp.List, func(l *dto.LoginLogInfo) []string {
//	    return []string{l.UserName, l.IP, constants.LoginStatusText[l.Status],
//	        time.UnixMilli(l.CreatedAt).Format("2006-01-02 15:04:05")}
//	})
func ExportCSV[T any](c *gin.Context, filename string, headers []string, items []T, mapper func(T) []string) {
	exp := csv.New(headers)
	exp.AddRows(csv.Rows(items, mapper))

	c.Header("Content-Type", csv.ContentTypeCSV)
	c.Header("Content-Disposition", contentDisposition(filename))
	c.Status(http.StatusOK)
	if err := exp.WriteToWriter(c.Writer); err != nil {
		c.Error(err) // 流已开始，无法转 JSON，仅记录供 OperationLogMiddleware
	}
}

// ParseCSVUpload 一行解析上传：从 form 字段取文件 → 自动编码识别（UTF-8/UTF-16/GBK）→ 返回 Parser。
func ParseCSVUpload(c *gin.Context, field string, hasHeaders bool) (*csv.Parser, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		return nil, err
	}
	if fh.Size > maxUploadSize {
		return nil, fmt.Errorf("文件过大（%d 字节），上限 %d 字节", fh.Size, maxUploadSize)
	}
	return csv.ParseMultipartForm(fh, hasHeaders)
}

// contentDisposition 构造下载响应头。同时给 filename 与 RFC 5987 的 filename*，
// 兼容中文文件名（老浏览器读 filename，新浏览器优先 filename*）。
func contentDisposition(filename string) string {
	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s",
		filename, url.PathEscape(filename))
}
