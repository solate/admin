package xgin

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestContentDisposition_ChineseFilename(t *testing.T) {
	got := contentDisposition("操作日志.csv")
	// 同时含 filename 与 RFC 5987 的 filename*
	if !strings.Contains(got, `filename="操作日志.csv"`) {
		t.Errorf("missing plain filename: %q", got)
	}
	if !strings.Contains(got, "filename*=UTF-8''") {
		t.Errorf("missing RFC 5987 filename*: %q", got)
	}
	// filename* 部分必须是百分号编码，不含裸中文字节
	star := got[strings.Index(got, "filename*=UTF-8''"):]
	if strings.ContainsRune(star, '操') {
		t.Errorf("filename* should be percent-encoded, got raw CJK: %q", star)
	}
}

func TestExportCSV(t *testing.T) {
	type row struct {
		ID   string
		Name string
	}
	items := []row{{"1", "张三"}, {"2", "has,comma"}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ExportCSV(c, "用户.csv", []string{"ID", "姓名"}, items, func(r row) []string {
		return []string{r.ID, r.Name}
	})

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.Bytes()

	// 含 UTF-8 BOM
	if !bytes.HasPrefix(body, utf8BOM) {
		t.Error("output should start with UTF-8 BOM")
	}
	text := string(body[len(utf8BOM):])
	if !strings.Contains(text, "ID,姓名") {
		t.Errorf("missing header: %q", text)
	}
	if !strings.Contains(text, "1,张三") {
		t.Errorf("missing row: %q", text)
	}
	// 含逗号的字段应被引号转义
	if !strings.Contains(text, `"has,comma"`) {
		t.Errorf("comma field not escaped: %q", text)
	}
	// 中文文件名安全
	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "filename*=UTF-8''") {
		t.Errorf("missing RFC 5987 filename*: %q", cd)
	}
}

func TestParseCSVUpload(t *testing.T) {
	// 构造 multipart 上传请求
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "in.csv")
	if err != nil {
		t.Fatalf("CreateFormFile error = %v", err)
	}
	if _, err := fw.Write([]byte("ID,Name\n1,张三\n2,李四")); err != nil {
		t.Fatalf("write error = %v", err)
	}
	mw.Close()

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	p, err := ParseCSVUpload(c, "file", true)
	if err != nil {
		t.Fatalf("ParseCSVUpload() error = %v", err)
	}
	rows, err := p.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][1] != "张三" || rows[1][1] != "李四" {
		t.Errorf("unexpected rows: %v", rows)
	}
}

func TestParseCSVUpload_MissingField(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.Close()

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	if _, err := ParseCSVUpload(c, "file", true); err == nil {
		t.Error("expected error when form field is missing")
	}
}

func TestParseCSVUpload_OversizedFile(t *testing.T) {
	// 构造超限文件（大于 maxUploadSize = 10MB）
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "large.csv")
	if err != nil {
		t.Fatalf("CreateFormFile error = %v", err)
	}
	// 写一个假 11MB 内容（实际 multipart 会略大于 11MB，触发超限）
	bigContent := make([]byte, 11<<20)
	if _, err := fw.Write(bigContent); err != nil {
		t.Fatalf("write error = %v", err)
	}
	mw.Close()

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	_, err = ParseCSVUpload(c, "file", true)
	if err == nil {
		t.Error("expected error for oversized file")
	}
	if err != nil && !strings.Contains(err.Error(), "文件过大") {
		t.Errorf("expected oversized error, got: %v", err)
	}
}
