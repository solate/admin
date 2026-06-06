package csv

import (
	"encoding/csv"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultDateTimeFormat = "2006-01-02 15:04:05"
	contentTypeHeader     = "text/csv"
	contentDisposition    = "attachment; filename=%s"
)

// Exporter CSV导出器
type Exporter struct {
	writer      *csv.Writer
	headers     []string
	records     [][]string
	file        *os.File
	filePath    string
	bufferSize  int
}

// Option 配置选项
type Option func(*Exporter)

// WithBufferSize 设置缓冲区大小
func WithBufferSize(size int) Option {
	return func(e *Exporter) {
		e.bufferSize = size
	}
}

// New 创建一个新的CSV导出器
func New(headers []string, opts ...Option) *Exporter {
	e := &Exporter{
		headers:    headers,
		records:    make([][]string, 0),
		bufferSize: 1000,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// AddRow 添加一行数据
func (e *Exporter) AddRow(row []string) {
	e.records = append(e.records, row)
}

// AddRows 批量添加多行数据
func (e *Exporter) AddRows(rows [][]string) {
	e.records = append(e.records, rows...)
}

// AddRowFromMap 从map添加一行数据，按headers顺序
func (e *Exporter) AddRowFromMap(data map[string]string) {
	row := make([]string, len(e.headers))
	for i, header := range e.headers {
		row[i] = data[header]
	}
	e.AddRow(row)
}

// AddRowFromMapAny 从map添加一行数据，支持任意类型，按headers顺序
func (e *Exporter) AddRowFromMapAny(data map[string]any) {
	row := make([]string, len(e.headers))
	for i, header := range e.headers {
		row[i] = formatValue(data[header])
	}
	e.AddRow(row)
}

// formatValue 格式化值为字符串
func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case time.Time:
		return val.Format(defaultDateTimeFormat)
	case *time.Time:
		if val != nil {
			return val.Format(defaultDateTimeFormat)
		}
		return ""
	default:
		return ""
	}
}

// WriteToFile 写入到文件
func (e *Exporter) WriteToFile(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	e.file = file
	e.filePath = filepath

	e.writer = csv.NewWriter(file)
	defer e.writer.Flush()

	// 写入表头
	if err := e.writer.Write(e.headers); err != nil {
		return err
	}

	// 写入数据
	for _, record := range e.records {
		if err := e.writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// WriteToWriter 写入到io.Writer
func (e *Exporter) WriteToWriter(w io.Writer) error {
	e.writer = csv.NewWriter(w)
	defer e.writer.Flush()

	// 写入表头
	if err := e.writer.Write(e.headers); err != nil {
		return err
	}

	// 写入数据
	for _, record := range e.records {
		if err := e.writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// Bytes 返回CSV数据的字节数组
func (e *Exporter) Bytes() ([]byte, error) {
	e.writer = csv.NewWriter(io.Discard)
	defer e.writer.Flush()

	// 写入表头
	if err := e.writer.Write(e.headers); err != nil {
		return nil, err
	}

	// 写入数据
	for _, record := range e.records {
		if err := e.writer.Write(record); err != nil {
			return nil, err
		}
	}

	// 重新获取数据
	return e.getBytes()
}

// getBytes 获取CSV字节数组
func (e *Exporter) getBytes() ([]byte, error) {
	var data [][]string
	data = append(data, e.headers)
	data = append(data, e.records...)

	var output [][]byte
	for _, row := range data {
		var csvRow []byte
		csvRow = append(csvRow, []byte(row[0])...)
		for i := 1; i < len(row); i++ {
			csvRow = append(csvRow, ',')
			csvRow = append(csvRow, []byte(row[i])...)
		}
		csvRow = append(csvRow, '\n')
		output = append(output, csvRow)
	}

	result := make([]byte, 0)
	for _, row := range output {
		result = append(result, row...)
	}
	return result, nil
}

// GetRecordCount 获取记录数
func (e *Exporter) GetRecordCount() int {
	return len(e.records)
}

// Clear 清空记录
func (e *Exporter) Clear() {
	e.records = make([][]string, 0)
}

// Close 关闭导出器并清理资源
func (e *Exporter) Close() error {
	if e.file != nil {
		return e.file.Close()
	}
	return nil
}

// GinResponse 将CSV数据写入gin.Context
func (e *Exporter) GinResponse(c *gin.Context, filename string) {
	c.Header("Content-Type", contentTypeHeader)
	c.Header("Content-Disposition", contentDisposition+filename)

	e.writer = csv.NewWriter(c.Writer)
	defer e.writer.Flush()

	// 写入表头
	e.writer.Write(e.headers)

	// 写入数据
	for _, record := range e.records {
		e.writer.Write(record)
	}
}

// GinResponseStream 流式写入gin.Context，适用于大量数据
func (e *Exporter) GinResponseStream(c *gin.Context, filename string, recordsChan <-chan []string) {
	c.Header("Content-Type", contentTypeHeader)
	c.Header("Content-Disposition", contentDisposition+filename)
	c.Stream(func(w io.Writer) bool {
		writer := csv.NewWriter(w)
		defer writer.Flush()

		// 写入表头
		writer.Write(e.headers)

		// 写入数据
		for record := range recordsChan {
			writer.Write(record)
		}
		return false
	})
}

// Parser CSV解析器
type Parser struct {
	reader     *csv.Reader
	headers    []string
	row        int
	hasHeaders bool
}

// NewParser 创建一个新的CSV解析器
func NewParser(r io.Reader, hasHeaders bool) *Parser {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // 允许不同行的字段数量不同

	p := &Parser{
		reader:     reader,
		hasHeaders: hasHeaders,
		row:        0,
	}

	if hasHeaders {
		headers, err := reader.Read()
		if err == nil {
			p.headers = headers
		}
	}

	return p
}

// NewParserFromFile 从文件创建解析器
func NewParserFromFile(filepath string, hasHeaders bool) (*Parser, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	return NewParser(file, hasHeaders), nil
}

// NewParserFromRequest 从HTTP请求创建解析器
func NewParserFromRequest(request *http.Request, fieldName string, hasHeaders bool) (*Parser, error) {
	file, _, err := request.FormFile(fieldName)
	if err != nil {
		return nil, err
	}

	return NewParser(file, hasHeaders), nil
}

// NewParserFromGin 从gin.Context创建解析器
func NewParserFromGin(c *gin.Context, fieldName string, hasHeaders bool) (*Parser, error) {
	fileHeader, err := c.FormFile(fieldName)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	return NewParser(file, hasHeaders), nil
}

// Read 读取一行数据
func (p *Parser) Read() ([]string, error) {
	record, err := p.reader.Read()
	if err != nil {
		return nil, err
	}
	p.row++
	return record, nil
}

// ReadAll 读取所有数据
func (p *Parser) ReadAll() ([][]string, error) {
	return p.reader.ReadAll()
}

// ReadMap 读取一行数据并返回map（需要hasHeaders=true）
func (p *Parser) ReadMap() (map[string]string, error) {
	if !p.hasHeaders {
		return nil, errors.New("parser must have headers to use ReadMap")
	}

	record, err := p.Read()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for i, header := range p.headers {
		if i < len(record) {
			result[header] = record[i]
		} else {
			result[header] = ""
		}
	}

	return result, nil
}

// ReadAllMap 读取所有数据并返回map数组（需要hasHeaders=true）
func (p *Parser) ReadAllMap() ([]map[string]string, error) {
	records, err := p.ReadAll()
	if err != nil {
		return nil, err
	}

	if !p.hasHeaders {
		return nil, errors.New("parser must have headers to use ReadAllMap")
	}

	result := make([]map[string]string, 0, len(records))
	for _, record := range records {
		row := make(map[string]string)
		for i, header := range p.headers {
			if i < len(record) {
				row[header] = record[i]
			} else {
				row[header] = ""
			}
		}
		result = append(result, row)
	}

	return result, nil
}

// GetHeaders 获取表头
func (p *Parser) GetHeaders() []string {
	return p.headers
}

// GetRow 获取当前行号
func (p *Parser) GetRow() int {
	return p.row
}

// ParseMultipartForm 解析multipart form中的CSV文件
func ParseMultipartForm(fileHeader *multipart.FileHeader, hasHeaders bool) (*Parser, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return NewParser(file, hasHeaders), nil
}
