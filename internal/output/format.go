package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Format represents the output format type
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// Writer handles output formatting
type Writer struct {
	format Format
	writer io.Writer
}

// NewWriter creates a new output writer
func NewWriter(format Format) *Writer {
	return &Writer{
		format: format,
		writer: os.Stdout,
	}
}

// WriteJSON writes data as JSON
func (w *Writer) WriteJSON(data interface{}) error {
	encoder := json.NewEncoder(w.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// WriteCSV writes data as CSV
// data should be a slice of structs or maps
func (w *Writer) WriteCSV(data interface{}, headers []string) error {
	csvWriter := csv.NewWriter(w.writer)
	defer csvWriter.Flush()

	// Write headers
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	// Convert data to rows
	rows, err := convertToCSVRows(data, headers)
	if err != nil {
		return err
	}

	// Write rows
	for _, row := range rows {
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// convertToCSVRows converts interface{} to CSV rows based on headers
func convertToCSVRows(data interface{}, headers []string) ([][]string, error) {
	var rows [][]string

	// Handle slice of data
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			row, err := itemToRow(item, headers)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
	} else {
		// Single item
		row, err := itemToRow(data, headers)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// itemToRow converts a single item to a CSV row
func itemToRow(item interface{}, headers []string) ([]string, error) {
	row := make([]string, len(headers))

	// Handle map[string]interface{}
	if m, ok := item.(map[string]interface{}); ok {
		for i, header := range headers {
			if val, exists := m[header]; exists {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		return row, nil
	}

	// Handle struct via reflection
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type for CSV conversion: %s", val.Kind())
	}

	typ := val.Type()
	fieldMap := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Get json tag name or use field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			jsonTag = field.Name
		} else {
			// Remove options like omitempty
			parts := strings.Split(jsonTag, ",")
			jsonTag = parts[0]
		}

		fieldMap[jsonTag] = fieldValue.Interface()
	}

	for i, header := range headers {
		if val, exists := fieldMap[header]; exists {
			row[i] = formatValue(val)
		}
	}

	return row, nil
}

// formatValue formats a value for CSV output
func formatValue(val interface{}) string {
	if val == nil {
		return ""
	}

	// Handle pointer types
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		val = v.Elem().Interface()
	}

	// Handle slices (like tags)
	if v.Kind() == reflect.Slice {
		var items []string
		for i := 0; i < v.Len(); i++ {
			items = append(items, fmt.Sprintf("%v", v.Index(i).Interface()))
		}
		return strings.Join(items, ";")
	}

	return fmt.Sprintf("%v", val)
}

// GetFormat returns the current format
func (w *Writer) GetFormat() Format {
	return w.format
}

// IsJSON returns true if format is JSON
func (w *Writer) IsJSON() bool {
	return w.format == FormatJSON
}

// IsCSV returns true if format is CSV
func (w *Writer) IsCSV() bool {
	return w.format == FormatCSV
}

// IsTable returns true if format is table (default)
func (w *Writer) IsTable() bool {
	return w.format == FormatTable
}
