// Package dataformat provides integration utilities for multiple data formats
// including JSON, CSV, YAML, and XML. It offers a unified interface for
// serializing and deserializing data across different formats.
package dataformat

import (
	"errors"
	"io"
	"strings"
)

// Format represents a supported data format type
type Format string

const (
	// FormatJSON represents JSON data format
	FormatJSON Format = "json"
	// FormatCSV represents CSV data format
	FormatCSV Format = "csv"
	// FormatYAML represents YAML data format
	FormatYAML Format = "yaml"
	// FormatXML represents XML data format
	FormatXML Format = "xml"
)

// Common errors for data format operations
var (
	ErrUnsupportedFormat = errors.New("unsupported data format")
	ErrInvalidData       = errors.New("invalid data for format conversion")
	ErrEmptyInput        = errors.New("empty input data")
	ErrNilValue          = errors.New("nil value provided")
	ErrTypeMismatch      = errors.New("type mismatch during conversion")
)

// FormatHandler defines the interface for format-specific handlers
type FormatHandler interface {
	// Marshal converts the given value to the format's byte representation
	Marshal(v interface{}) ([]byte, error)

	// MarshalIndent converts the given value to an indented byte representation
	MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)

	// Unmarshal parses the format-encoded data and stores the result in v
	Unmarshal(data []byte, v interface{}) error

	// Format returns the format type this handler processes
	Format() Format

	// FileExtension returns the typical file extension for this format
	FileExtension() string

	// ContentType returns the MIME content type for this format
	ContentType() string

	// Encode writes the format encoding of v to the writer
	Encode(w io.Writer, v interface{}) error

	// Decode reads format-encoded data from the reader and stores it in v
	Decode(r io.Reader, v interface{}) error
}

// ConvertOptions specifies options for format conversion
type ConvertOptions struct {
	// Indent specifies whether output should be indented
	Indent bool

	// IndentString specifies the string to use for indentation (default: "  ")
	IndentString string

	// CSVHeader specifies whether CSV output should include headers
	CSVHeader bool

	// XMLRoot specifies the root element name for XML output
	XMLRoot string

	// PreserveOrder attempts to maintain field order during conversion
	PreserveOrder bool
}

// DefaultConvertOptions returns the default conversion options
func DefaultConvertOptions() ConvertOptions {
	return ConvertOptions{
		Indent:        true,
		IndentString:  "  ",
		CSVHeader:     true,
		XMLRoot:       "data",
		PreserveOrder: false,
	}
}

// ParseFormat parses a string into a Format type
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "xml":
		return FormatXML, nil
	default:
		return "", ErrUnsupportedFormat
	}
}

// FormatFromExtension determines the format from a file extension
func FormatFromExtension(ext string) (Format, error) {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "xml":
		return FormatXML, nil
	default:
		return "", ErrUnsupportedFormat
	}
}

// IsValidFormat checks if the given format string is valid
func IsValidFormat(s string) bool {
	_, err := ParseFormat(s)
	return err == nil
}

// SupportedFormats returns a list of all supported formats
func SupportedFormats() []Format {
	return []Format{FormatJSON, FormatCSV, FormatYAML, FormatXML}
}
