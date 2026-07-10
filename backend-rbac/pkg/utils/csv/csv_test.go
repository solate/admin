package csv

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func TestNew(t *testing.T) {
	e := New([]string{"ID", "Name", "Email"})
	if e == nil {
		t.Fatal("New() returned nil")
	}
	if len(e.headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(e.headers))
	}
	if e.GetRecordCount() != 0 {
		t.Errorf("expected 0 records, got %d", e.GetRecordCount())
	}
}

func TestExporter_AddRow(t *testing.T) {
	e := New([]string{"ID", "Name"})
	e.AddRow([]string{"1", "John"})
	e.AddRow([]string{"2", "Jane"})
	if e.GetRecordCount() != 2 {
		t.Errorf("expected 2 records, got %d", e.GetRecordCount())
	}
}

func TestExporter_AddRows(t *testing.T) {
	e := New([]string{"ID", "Name"})
	e.AddRows([][]string{{"1", "John"}, {"2", "Jane"}, {"3", "Bob"}})
	if e.GetRecordCount() != 3 {
		t.Errorf("expected 3 records, got %d", e.GetRecordCount())
	}
}

func TestRows(t *testing.T) {
	type user struct {
		ID   string
		Name string
	}
	users := []user{{"1", "John"}, {"2", "Jane"}}
	rows := Rows(users, func(u user) []string {
		return []string{u.ID, u.Name}
	})
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "1" || rows[0][1] != "John" {
		t.Errorf("row 0 incorrect: %v", rows[0])
	}
	if rows[1][0] != "2" || rows[1][1] != "Jane" {
		t.Errorf("row 1 incorrect: %v", rows[1])
	}

	// 空切片返回非 nil 空结果
	empty := Rows([]user{}, func(u user) []string { return []string{u.ID} })
	if empty == nil || len(empty) != 0 {
		t.Errorf("expected non-nil empty slice, got %v", empty)
	}
}

func TestExporter_Bytes_HasBOM(t *testing.T) {
	e := New([]string{"ID", "Name"})
	e.AddRow([]string{"1", "John"})

	data, err := e.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}
	if !bytes.HasPrefix(data, utf8BOM) {
		t.Error("output should start with UTF-8 BOM")
	}
	// 去掉 BOM 后应是合法 CSV 文本
	body := string(data[len(utf8BOM):])
	if !strings.Contains(body, "ID,Name") {
		t.Errorf("output missing header: %q", body)
	}
	if !strings.Contains(body, "1,John") {
		t.Errorf("output missing row: %q", body)
	}
}

// TestExporter_Bytes_Escaping 验证含逗号/引号/换行的字段被正确转义（旧 getBytes 的核心 bug）。
func TestExporter_Bytes_Escaping(t *testing.T) {
	e := New([]string{"A", "B"})
	e.AddRow([]string{"has,comma", "has\"quote\nnewline"})

	data, err := e.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}

	// 用解析器往返，字段应原样还原
	p, err := NewParser(bytes.NewReader(data), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	row, err := p.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[0] != "has,comma" {
		t.Errorf("field 0 not preserved: %q", row[0])
	}
	if row[1] != "has\"quote\nnewline" {
		t.Errorf("field 1 not preserved: %q", row[1])
	}
}

func TestExporter_WriteToWriter(t *testing.T) {
	e := New([]string{"ID", "Name"})
	e.AddRow([]string{"1", "John"})

	var buf bytes.Buffer
	if err := e.WriteToWriter(&buf); err != nil {
		t.Fatalf("WriteToWriter() error = %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), utf8BOM) {
		t.Error("output should start with UTF-8 BOM")
	}
	if !strings.Contains(buf.String(), "ID,Name") {
		t.Error("output should contain header")
	}
}

func TestExporter_WriteToFile(t *testing.T) {
	e := New([]string{"ID", "Name"})
	e.AddRow([]string{"1", "John"})

	path := filepath.Join(t.TempDir(), "test_export.csv")
	if err := e.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !bytes.HasPrefix(content, utf8BOM) {
		t.Error("file should start with UTF-8 BOM")
	}
	if !strings.Contains(string(content), "ID,Name") {
		t.Error("file should contain header")
	}
}

func TestExporter_Clear(t *testing.T) {
	e := New([]string{"ID", "Name"})
	e.AddRow([]string{"1", "John"})
	e.Clear()
	if e.GetRecordCount() != 0 {
		t.Errorf("expected 0 records after Clear(), got %d", e.GetRecordCount())
	}
}

func TestNewParser_UTF8(t *testing.T) {
	p, err := NewParser(strings.NewReader("ID,Name\n1,John\n2,Jane"), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	if len(p.GetHeaders()) != 2 {
		t.Errorf("expected 2 headers, got %d", len(p.GetHeaders()))
	}
}

func TestNewParser_UTF8BOM(t *testing.T) {
	raw := append(append([]byte{}, utf8BOM...), []byte("ID,Name\n1,张三")...)
	p, err := NewParser(bytes.NewReader(raw), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	// 表头不应残留 BOM
	if p.GetHeaders()[0] != "ID" {
		t.Errorf("BOM not stripped from header: %q", p.GetHeaders()[0])
	}
	row, err := p.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[1] != "张三" {
		t.Errorf("expected 张三, got %q", row[1])
	}
}

func TestNewParser_GBK(t *testing.T) {
	// 构造 GBK 编码的输入
	gbk, err := encodeWith("ID,Name\n1,张三\n2,李四", simplifiedchinese.GBK)
	if err != nil {
		t.Fatalf("failed to build GBK input: %v", err)
	}
	p, err := NewParser(bytes.NewReader(gbk), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	rows, err := p.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][1] != "张三" || rows[1][1] != "李四" {
		t.Errorf("GBK not decoded: %v", rows)
	}
}

func TestNewParser_UTF16LE(t *testing.T) {
	utf16, err := encodeWith("ID,Name\n1,张三",
		unicode.UTF16(unicode.LittleEndian, unicode.UseBOM))
	if err != nil {
		t.Fatalf("failed to build UTF-16LE input: %v", err)
	}
	p, err := NewParser(bytes.NewReader(utf16), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	row, err := p.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[1] != "张三" {
		t.Errorf("UTF-16LE not decoded: %q", row[1])
	}
}

func TestParser_Read(t *testing.T) {
	p, err := NewParser(strings.NewReader("ID,Name\n1,John\n2,Jane"), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	row, err := p.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[0] != "1" || row[1] != "John" {
		t.Errorf("expected [1 John], got %v", row)
	}
	if p.GetRow() != 1 {
		t.Errorf("expected row=1, got %d", p.GetRow())
	}
}

func TestParser_ReadAll_NoHeaders(t *testing.T) {
	p, err := NewParser(strings.NewReader("ID,Name\n1,John\n2,Jane"), false)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	rows, err := p.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(rows) != 3 { // 含被当作数据的首行
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
	// ReadAll 应把读到的行数计入 GetRow（与 Read 的逐行计数一致）
	if p.GetRow() != 3 {
		t.Errorf("expected GetRow()=3 after ReadAll, got %d", p.GetRow())
	}
}

func TestParser_ReadMap(t *testing.T) {
	p, err := NewParser(strings.NewReader("ID,Name\n1,John"), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	row, err := p.ReadMap()
	if err != nil {
		t.Fatalf("ReadMap() error = %v", err)
	}
	if row["ID"] != "1" || row["Name"] != "John" {
		t.Errorf("map incorrect: %v", row)
	}
}

func TestParser_ReadAllMap(t *testing.T) {
	p, err := NewParser(strings.NewReader("ID,Name\n1,John\n2,Jane"), true)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	rows, err := p.ReadAllMap()
	if err != nil {
		t.Fatalf("ReadAllMap() error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["ID"] != "1" || rows[1]["Name"] != "Jane" {
		t.Errorf("maps incorrect: %v", rows)
	}
}

func TestParser_ReadMap_NoHeaders(t *testing.T) {
	p, err := NewParser(strings.NewReader("1,John"), false)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}
	if _, err := p.ReadMap(); err == nil {
		t.Error("expected error when calling ReadMap() without headers")
	}
}

func TestNewParserFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "in.csv")
	if err := os.WriteFile(path, []byte("ID,Name\n1,John"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	p, err := NewParserFromFile(path, true)
	if err != nil {
		t.Fatalf("NewParserFromFile() error = %v", err)
	}
	row, err := p.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[0] != "1" {
		t.Errorf("expected 1, got %q", row[0])
	}
}

// encodeWith 把 UTF-8 字符串编码成指定编码的字节，供测试构造非 UTF-8 输入。
func encodeWith(s string, enc interface {
	NewEncoder() *encoding.Encoder
}) ([]byte, error) {
	return io.ReadAll(transform.NewReader(strings.NewReader(s), enc.NewEncoder()))
}
