package csv

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewExporter(t *testing.T) {
	headers := []string{"ID", "Name", "Email"}
	exporter := New(headers)

	if exporter == nil {
		t.Fatal("New() returned nil")
	}

	if len(exporter.headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(exporter.headers))
	}
}

func TestExporter_AddRow(t *testing.T) {
	headers := []string{"ID", "Name"}
	exporter := New(headers)

	exporter.AddRow([]string{"1", "John"})
	exporter.AddRow([]string{"2", "Jane"})

	if exporter.GetRecordCount() != 2 {
		t.Errorf("Expected 2 records, got %d", exporter.GetRecordCount())
	}
}

func TestExporter_AddRows(t *testing.T) {
	headers := []string{"ID", "Name"}
	exporter := New(headers)

	rows := [][]string{
		{"1", "John"},
		{"2", "Jane"},
		{"3", "Bob"},
	}
	exporter.AddRows(rows)

	if exporter.GetRecordCount() != 3 {
		t.Errorf("Expected 3 records, got %d", exporter.GetRecordCount())
	}
}

func TestExporter_AddRowFromMap(t *testing.T) {
	headers := []string{"ID", "Name", "Email"}
	exporter := New(headers)

	data := map[string]string{
		"ID":    "1",
		"Name":  "John",
		"Email": "john@example.com",
	}
	exporter.AddRowFromMap(data)

	if exporter.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", exporter.GetRecordCount())
	}
}

func TestExporter_AddRowFromMapAny(t *testing.T) {
	headers := []string{"ID", "Name", "Age", "Active", "CreatedAt"}
	exporter := New(headers)

	now := time.Now()
	data := map[string]any{
		"ID":        1,
		"Name":      "John",
		"Age":       30,
		"Active":    true,
		"CreatedAt": now,
	}
	exporter.AddRowFromMapAny(data)

	if exporter.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", exporter.GetRecordCount())
	}
}

func TestExporter_WriteToWriter(t *testing.T) {
	headers := []string{"ID", "Name"}
	exporter := New(headers)

	exporter.AddRow([]string{"1", "John"})
	exporter.AddRow([]string{"2", "Jane"})

	var buf bytes.Buffer
	err := exporter.WriteToWriter(&buf)
	if err != nil {
		t.Fatalf("WriteToWriter() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ID,Name") {
		t.Error("Output should contain headers")
	}
	if !strings.Contains(output, "1,John") {
		t.Error("Output should contain first row")
	}
	if !strings.Contains(output, "2,Jane") {
		t.Error("Output should contain second row")
	}
}

func TestExporter_WriteToFile(t *testing.T) {
	headers := []string{"ID", "Name"}
	exporter := New(headers)

	exporter.AddRow([]string{"1", "John"})
	exporter.AddRow([]string{"2", "Jane"})

	tmpFile := "/tmp/test_export.csv"
	defer os.Remove(tmpFile)

	err := exporter.WriteToFile(tmpFile)
	if err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("File was not created")
	}

	// Verify file contents
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "ID,Name") {
		t.Error("File should contain headers")
	}
}

func TestExporter_Clear(t *testing.T) {
	headers := []string{"ID", "Name"}
	exporter := New(headers)

	exporter.AddRow([]string{"1", "John"})
	exporter.Clear()

	if exporter.GetRecordCount() != 0 {
		t.Errorf("Expected 0 records after Clear(), got %d", exporter.GetRecordCount())
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "test", "test"},
		{"int", 123, "123"},
		{"int64", int64(456), "456"},
		{"uint", uint(789), "789"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"nil", nil, ""},
		{"time", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), "2024-01-01 12:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewParser(t *testing.T) {
	csvData := "ID,Name\n1,John\n2,Jane"
	reader := strings.NewReader(csvData)

	parser := NewParser(reader, true)

	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}

	headers := parser.GetHeaders()
	if len(headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(headers))
	}
}

func TestParser_Read(t *testing.T) {
	csvData := "ID,Name\n1,John\n2,Jane"
	reader := strings.NewReader(csvData)

	parser := NewParser(reader, true)

	// First data row
	row, err := parser.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[0] != "1" || row[1] != "John" {
		t.Errorf("Expected [1 John], got %v", row)
	}

	// Second data row
	row, err = parser.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if row[0] != "2" || row[1] != "Jane" {
		t.Errorf("Expected [2 Jane], got %v", row)
	}
}

func TestParser_ReadAll(t *testing.T) {
	csvData := "ID,Name\n1,John\n2,Jane"
	reader := strings.NewReader(csvData)

	parser := NewParser(reader, false)

	rows, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows (including header), got %d", len(rows))
	}
}

func TestParser_ReadMap(t *testing.T) {
	csvData := "ID,Name\n1,John\n2,Jane"
	reader := strings.NewReader(csvData)

	parser := NewParser(reader, true)

	row, err := parser.ReadMap()
	if err != nil {
		t.Fatalf("ReadMap() error = %v", err)
	}

	if row["ID"] != "1" {
		t.Errorf("Expected ID=1, got %s", row["ID"])
	}
	if row["Name"] != "John" {
		t.Errorf("Expected Name=John, got %s", row["Name"])
	}
}

func TestParser_ReadAllMap(t *testing.T) {
	csvData := "ID,Name\n1,John\n2,Jane"
	reader := strings.NewReader(csvData)

	parser := NewParser(reader, true)

	rows, err := parser.ReadAllMap()
	if err != nil {
		t.Fatalf("ReadAllMap() error = %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}

	if rows[0]["ID"] != "1" || rows[0]["Name"] != "John" {
		t.Errorf("First row incorrect: %v", rows[0])
	}

	if rows[1]["ID"] != "2" || rows[1]["Name"] != "Jane" {
		t.Errorf("Second row incorrect: %v", rows[1])
	}
}

func TestParser_ReadMap_NoHeaders(t *testing.T) {
	csvData := "1,John\n2,Jane"
	reader := strings.NewReader(csvData)

	parser := NewParser(reader, false)

	_, err := parser.ReadMap()
	if err == nil {
		t.Error("Expected error when calling ReadMap() without headers")
	}
}
