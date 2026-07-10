// Package csv 基于标准库 encoding/csv 的轻量封装，聚焦两件标准库不做、
// 而每个业务都要重复踩坑的事：
//  1. 导出：写 UTF-8 BOM，保证 Excel 双击打开中文不乱码。
//  2. 导入：自动识别 UTF-8(BOM) / UTF-16(LE/BE) / GBK 编码并转成 UTF-8。
//
// 行数据一律用 []string（标准库原生形态），类型安全的 struct→行映射由泛型 Rows 提供，
// 不提供 map[string]any 弱类型入口（实践中无人使用且绕过类型检查）。
//
// 本包只碰 stdlib + golang.org/x/text，不依赖任何 web 框架，可整目录 copy。
// gin 相关的下载/上传便利方法见 pkg/xgin。
package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// utf8BOM UTF-8 字节序标记。写在导出内容最前，Excel 见到它才用 UTF-8 解码（否则按 GBK 解，中文乱码）。
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

var (
	utf16LEBOM = []byte{0xFF, 0xFE}
	utf16BEBOM = []byte{0xFE, 0xFF}
)

// ContentTypeCSV CSV 下载响应的 Content-Type。供 web 层（如 pkg/xgin）设置响应头用。
const ContentTypeCSV = "text/csv; charset=utf-8"

// Exporter CSV 导出器。持有表头与行数据，各输出方法均自动写 UTF-8 BOM。
type Exporter struct {
	headers []string
	records [][]string
}

// New 创建导出器。headers 为表头，后续 AddRow 的每行长度应与之一致。
func New(headers []string) *Exporter {
	return &Exporter{
		headers: headers,
		records: make([][]string, 0),
	}
}

// AddRow 追加一行。
func (e *Exporter) AddRow(row []string) {
	e.records = append(e.records, row)
}

// AddRows 批量追加多行。
func (e *Exporter) AddRows(rows [][]string) {
	e.records = append(e.records, rows...)
}

// Rows 泛型行映射：把 []T 按 mapper 转成 [][]string。
// 类型安全、无反射、无依赖——替代弱类型的 map[string]any 入口。
//
//	exporter.AddRows(csv.Rows(logs, func(l OperationLog) []string {
//	    return []string{l.LogID, l.UserName, timeFmt(l.CreatedAt)}
//	}))
func Rows[T any](items []T, mapper func(T) []string) [][]string {
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, mapper(item))
	}
	return rows
}

// write 把 BOM + 表头 + 数据写入 w。所有输出方法的公共实现。
func (e *Exporter) write(w io.Writer) error {
	if _, err := w.Write(utf8BOM); err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	if err := writer.Write(e.headers); err != nil {
		return err
	}
	if err := writer.WriteAll(e.records); err != nil { // WriteAll 内部已 Flush
		return err
	}
	return writer.Error()
}

// Bytes 返回完整 CSV 字节（含 BOM）。
func (e *Exporter) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := e.write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteToWriter 写入任意 io.Writer（含 BOM）。
func (e *Exporter) WriteToWriter(w io.Writer) error {
	return e.write(w)
}

// WriteToFile 写入文件（含 BOM）。
func (e *Exporter) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return e.write(file)
}

// GetRecordCount 返回已添加的数据行数（不含表头）。
func (e *Exporter) GetRecordCount() int {
	return len(e.records)
}

// Clear 清空数据行（保留表头）。
func (e *Exporter) Clear() {
	e.records = make([][]string, 0)
}

// Parser CSV 解析器。构造时已把输入统一转成 UTF-8。
type Parser struct {
	reader     *csv.Reader
	headers    []string
	row        int
	hasHeaders bool
}

// NewParser 从 io.Reader 创建解析器，自动识别编码并转 UTF-8。
// hasHeaders=true 时构造中即读掉首行作为表头。
func NewParser(r io.Reader, hasHeaders bool) (*Parser, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	data, err := convertToUTF8(raw)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1 // 允许各行字段数不同

	p := &Parser{reader: reader, hasHeaders: hasHeaders}
	if hasHeaders {
		headers, err := reader.Read()
		if err == nil {
			p.headers = headers
		}
	}
	return p, nil
}

// convertToUTF8 按 BOM / 内容特征识别编码并转 UTF-8。
// 支持：UTF-8(含BOM去头) / UTF-16 LE/BE / GBK(中文 Windows Excel 最常见)。
func convertToUTF8(data []byte) ([]byte, error) {
	switch {
	case bytes.HasPrefix(data, utf8BOM):
		return data[len(utf8BOM):], nil
	case bytes.HasPrefix(data, utf16LEBOM):
		return decodeWith(data[2:], unicode.UTF16(unicode.LittleEndian, unicode.UseBOM))
	case bytes.HasPrefix(data, utf16BEBOM):
		return decodeWith(data[2:], unicode.UTF16(unicode.BigEndian, unicode.UseBOM))
	case utf8.Valid(data):
		return data, nil
	default:
		// 非法 UTF-8：按 GBK/GB18030 兜底
		converted, err := decodeWith(data, simplifiedchinese.GB18030)
		if err != nil {
			return nil, errors.New("无法识别文件编码，请使用 UTF-8 或 GBK 编码的 CSV 文件")
		}
		return converted, nil
	}
}

// decodeWith 用指定编码把字节转 UTF-8。
func decodeWith(data []byte, enc encoding.Encoding) ([]byte, error) {
	return io.ReadAll(transform.NewReader(bytes.NewReader(data), enc.NewDecoder()))
}

// Read 读取一行。
func (p *Parser) Read() ([]string, error) {
	record, err := p.reader.Read()
	if err != nil {
		return nil, err
	}
	p.row++
	return record, nil
}

// ReadAll 读取剩余全部行。
func (p *Parser) ReadAll() ([][]string, error) {
	records, err := p.reader.ReadAll()
	p.row += len(records)
	return records, err
}

// ReadMap 读一行并按表头组成 map（需 hasHeaders）。
func (p *Parser) ReadMap() (map[string]string, error) {
	if !p.hasHeaders {
		return nil, errors.New("parser must have headers to use ReadMap")
	}
	record, err := p.Read()
	if err != nil {
		return nil, err
	}
	return p.toMap(record), nil
}

// ReadAllMap 读全部行并按表头组成 []map（需 hasHeaders）。
func (p *Parser) ReadAllMap() ([]map[string]string, error) {
	if !p.hasHeaders {
		return nil, errors.New("parser must have headers to use ReadAllMap")
	}
	records, err := p.ReadAll()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]string, 0, len(records))
	for _, record := range records {
		result = append(result, p.toMap(record))
	}
	return result, nil
}

// toMap 按表头把一行组装成 map，缺失列补空串。
func (p *Parser) toMap(record []string) map[string]string {
	row := make(map[string]string, len(p.headers))
	for i, header := range p.headers {
		if i < len(record) {
			row[header] = record[i]
		} else {
			row[header] = ""
		}
	}
	return row
}

// GetHeaders 返回表头。
func (p *Parser) GetHeaders() []string {
	return p.headers
}

// GetRow 返回已读行号。
func (p *Parser) GetRow() int {
	return p.row
}

// NewParserFromFile 从文件路径创建解析器。
func NewParserFromFile(path string, hasHeaders bool) (*Parser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return NewParser(file, hasHeaders)
}

// ParseMultipartForm 从 multipart 文件头创建解析器。
func ParseMultipartForm(fileHeader *multipart.FileHeader, hasHeaders bool) (*Parser, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return NewParser(file, hasHeaders)
}
