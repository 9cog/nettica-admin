package dataformat

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

// CSVHandler implements FormatHandler for CSV data
type CSVHandler struct {
	// Delimiter specifies the field delimiter (default: comma)
	Delimiter rune

	// UseCRLF specifies whether to use \r\n as the line terminator
	UseCRLF bool

	// IncludeHeader specifies whether to include headers in output
	IncludeHeader bool

	// FlattenNested specifies whether to flatten nested structures
	FlattenNested bool

	// NullValue specifies the string to use for null/nil values
	NullValue string
}

// NewCSVHandler creates a new CSV format handler with default settings
func NewCSVHandler() *CSVHandler {
	return &CSVHandler{
		Delimiter:     ',',
		UseCRLF:       false,
		IncludeHeader: true,
		FlattenNested: true,
		NullValue:     "",
	}
}

// Marshal converts the given value to CSV bytes
func (h *CSVHandler) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	buf := &bytes.Buffer{}
	if err := h.Encode(buf, v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MarshalIndent is not applicable for CSV, calls Marshal
func (h *CSVHandler) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return h.Marshal(v)
}

// Unmarshal parses CSV-encoded data and stores the result in v
// v must be a pointer to a slice
func (h *CSVHandler) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return ErrEmptyInput
	}
	if v == nil {
		return ErrNilValue
	}

	return h.Decode(bytes.NewReader(data), v)
}

// Format returns FormatCSV
func (h *CSVHandler) Format() Format {
	return FormatCSV
}

// FileExtension returns "csv"
func (h *CSVHandler) FileExtension() string {
	return "csv"
}

// ContentType returns the MIME type for CSV
func (h *CSVHandler) ContentType() string {
	return "text/csv"
}

// Encode writes the CSV encoding of v to the writer
func (h *CSVHandler) Encode(w io.Writer, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	// Convert to slice of maps for uniform handling
	records, headers, err := h.toRecords(v)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	writer.Comma = h.Delimiter
	writer.UseCRLF = h.UseCRLF

	// Write header if enabled
	if h.IncludeHeader && len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	// Write records
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			if val, ok := record[header]; ok {
				row[i] = h.formatValue(val)
			} else {
				row[i] = h.NullValue
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// Decode reads CSV-encoded data from the reader and stores it in v
func (h *CSVHandler) Decode(r io.Reader, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	reader := csv.NewReader(r)
	reader.Comma = h.Delimiter
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return ErrEmptyInput
	}

	// First row is headers
	headers := records[0]
	dataRecords := records[1:]

	// Convert to slice of maps
	var result []map[string]interface{}
	for _, record := range dataRecords {
		m := make(map[string]interface{})
		for i, header := range headers {
			if i < len(record) {
				m[header] = record[i]
			}
		}
		result = append(result, m)
	}

	// Marshal to JSON then unmarshal to target type
	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

// toRecords converts any value to a slice of maps and extracts headers
func (h *CSVHandler) toRecords(v interface{}) ([]map[string]interface{}, []string, error) {
	// First convert to JSON, then to []map[string]interface{}
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}

	var records []map[string]interface{}

	// Try to unmarshal as array first
	if err := json.Unmarshal(jsonData, &records); err != nil {
		// Try as single object
		var single map[string]interface{}
		if err := json.Unmarshal(jsonData, &single); err != nil {
			return nil, nil, ErrInvalidData
		}
		records = []map[string]interface{}{single}
	}

	// Flatten nested structures if enabled
	if h.FlattenNested {
		for i, record := range records {
			records[i] = h.flattenMap(record, "")
		}
	}

	// Extract unique headers from all records
	headerSet := make(map[string]bool)
	for _, record := range records {
		for key := range record {
			headerSet[key] = true
		}
	}

	// Sort headers for consistent output
	headers := make([]string, 0, len(headerSet))
	for header := range headerSet {
		headers = append(headers, header)
	}
	sort.Strings(headers)

	return records, headers, nil
}

// flattenMap flattens a nested map structure
func (h *CSVHandler) flattenMap(data map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively flatten nested maps
			for k, val := range h.flattenMap(v, fullKey) {
				result[k] = val
			}
		case []interface{}:
			// Convert arrays to JSON string
			if jsonBytes, err := json.Marshal(v); err == nil {
				result[fullKey] = string(jsonBytes)
			} else {
				result[fullKey] = fmt.Sprintf("%v", v)
			}
		default:
			result[fullKey] = value
		}
	}

	return result
}

// formatValue converts a value to its string representation
func (h *CSVHandler) formatValue(v interface{}) string {
	if v == nil {
		return h.NullValue
	}

	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		// Handle integer-like floats
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case float32:
		if val == float32(int32(val)) {
			return fmt.Sprintf("%d", int32(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// CSVToMaps converts CSV bytes to a slice of maps
func CSVToMaps(data []byte) ([]map[string]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, ErrEmptyInput
	}

	headers := records[0]
	result := make([]map[string]string, 0, len(records)-1)

	for _, record := range records[1:] {
		m := make(map[string]string)
		for i, header := range headers {
			if i < len(record) {
				m[header] = record[i]
			}
		}
		result = append(result, m)
	}

	return result, nil
}

// MapsToCSV converts a slice of maps to CSV bytes
func MapsToCSV(data []map[string]string, headers []string) ([]byte, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	// If no headers provided, extract from first map
	if len(headers) == 0 && len(data) > 0 {
		for key := range data[0] {
			headers = append(headers, key)
		}
		sort.Strings(headers)
	}

	// Write header
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Write records
	for _, record := range data {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = record[header]
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetCSVHeaders extracts headers from CSV data
func GetCSVHeaders(data []byte) ([]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	return headers, nil
}

// StructToCSVHeaders extracts field names from a struct for use as CSV headers
func StructToCSVHeaders(v interface{}) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Check for json tag first
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			headers = append(headers, parts[0])
		} else if tag := field.Tag.Get("csv"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			headers = append(headers, parts[0])
		} else {
			headers = append(headers, field.Name)
		}
	}

	return headers
}
