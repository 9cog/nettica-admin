package dataformat

import (
	"bytes"
	"encoding/json"
	"io"
)

// JSONHandler implements FormatHandler for JSON data
type JSONHandler struct {
	// EscapeHTML specifies whether problematic HTML characters should be escaped
	EscapeHTML bool
}

// NewJSONHandler creates a new JSON format handler with default settings
func NewJSONHandler() *JSONHandler {
	return &JSONHandler{
		EscapeHTML: false,
	}
}

// Marshal converts the given value to JSON bytes
func (h *JSONHandler) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(h.EscapeHTML)

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	// Remove trailing newline that Encoder adds
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// MarshalIndent converts the given value to indented JSON bytes
func (h *JSONHandler) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal parses JSON-encoded data and stores the result in v
func (h *JSONHandler) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return ErrEmptyInput
	}
	if v == nil {
		return ErrNilValue
	}

	return json.Unmarshal(data, v)
}

// Format returns FormatJSON
func (h *JSONHandler) Format() Format {
	return FormatJSON
}

// FileExtension returns "json"
func (h *JSONHandler) FileExtension() string {
	return "json"
}

// ContentType returns the MIME type for JSON
func (h *JSONHandler) ContentType() string {
	return "application/json"
}

// Encode writes the JSON encoding of v to the writer
func (h *JSONHandler) Encode(w io.Writer, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(h.EscapeHTML)

	return encoder.Encode(v)
}

// Decode reads JSON-encoded data from the reader and stores it in v
func (h *JSONHandler) Decode(r io.Reader, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	return json.NewDecoder(r).Decode(v)
}

// IsValidJSON checks if the given bytes represent valid JSON
func IsValidJSON(data []byte) bool {
	return json.Valid(data)
}

// PrettyPrintJSON formats JSON data with indentation
func PrettyPrintJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CompactJSON removes whitespace from JSON data
func CompactJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// JSONToMap converts JSON bytes to a map
func JSONToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MapToJSON converts a map to JSON bytes
func MapToJSON(data map[string]interface{}, indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}
