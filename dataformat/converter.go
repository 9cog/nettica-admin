package dataformat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Converter provides a unified interface for converting between data formats
type Converter struct {
	handlers map[Format]FormatHandler
	options  ConvertOptions
	mu       sync.RWMutex
}

// NewConverter creates a new Converter with default handlers
func NewConverter() *Converter {
	c := &Converter{
		handlers: make(map[Format]FormatHandler),
		options:  DefaultConvertOptions(),
	}

	// Register default handlers
	c.RegisterHandler(NewJSONHandler())
	c.RegisterHandler(NewCSVHandler())
	c.RegisterHandler(NewYAMLHandler())
	c.RegisterHandler(NewXMLHandler())

	return c
}

// NewConverterWithOptions creates a new Converter with custom options
func NewConverterWithOptions(opts ConvertOptions) *Converter {
	c := NewConverter()
	c.options = opts
	return c
}

// RegisterHandler registers a custom format handler
func (c *Converter) RegisterHandler(h FormatHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[h.Format()] = h
}

// GetHandler returns the handler for a specific format
func (c *Converter) GetHandler(format Format) (FormatHandler, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	h, ok := c.handlers[format]
	if !ok {
		return nil, ErrUnsupportedFormat
	}
	return h, nil
}

// SetOptions updates the converter options
func (c *Converter) SetOptions(opts ConvertOptions) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.options = opts
}

// GetOptions returns the current converter options
func (c *Converter) GetOptions() ConvertOptions {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.options
}

// Convert transforms data from one format to another
func (c *Converter) Convert(data []byte, fromFormat, toFormat Format) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyInput
	}

	// Get handlers
	fromHandler, err := c.GetHandler(fromFormat)
	if err != nil {
		return nil, fmt.Errorf("source format error: %w", err)
	}

	toHandler, err := c.GetHandler(toFormat)
	if err != nil {
		return nil, fmt.Errorf("target format error: %w", err)
	}

	// Use JSON as intermediate format for cross-format conversion
	var intermediate interface{}

	if fromFormat == FormatJSON {
		// Direct unmarshal for JSON
		if err := json.Unmarshal(data, &intermediate); err != nil {
			return nil, fmt.Errorf("failed to parse source data: %w", err)
		}
	} else {
		// Unmarshal from source format
		if err := fromHandler.Unmarshal(data, &intermediate); err != nil {
			return nil, fmt.Errorf("failed to parse source data: %w", err)
		}
	}

	// Marshal to target format
	c.mu.RLock()
	opts := c.options
	c.mu.RUnlock()

	if opts.Indent {
		return toHandler.MarshalIndent(intermediate, "", opts.IndentString)
	}
	return toHandler.Marshal(intermediate)
}

// ConvertTo converts any Go value to the specified format
func (c *Converter) ConvertTo(v interface{}, format Format) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	handler, err := c.GetHandler(format)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	opts := c.options
	c.mu.RUnlock()

	if opts.Indent {
		return handler.MarshalIndent(v, "", opts.IndentString)
	}
	return handler.Marshal(v)
}

// ConvertFrom parses data in the specified format into a Go value
func (c *Converter) ConvertFrom(data []byte, format Format, v interface{}) error {
	if len(data) == 0 {
		return ErrEmptyInput
	}
	if v == nil {
		return ErrNilValue
	}

	handler, err := c.GetHandler(format)
	if err != nil {
		return err
	}

	return handler.Unmarshal(data, v)
}

// ConvertFile reads a file, converts it to another format, and optionally writes to a new file
func (c *Converter) ConvertFile(inputPath, outputPath string, toFormat Format) error {
	// Determine source format from file extension
	fromFormat, err := FormatFromExtension(filepath.Ext(inputPath))
	if err != nil {
		return fmt.Errorf("cannot determine source format from %s: %w", inputPath, err)
	}

	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Convert
	result, err := c.Convert(data, fromFormat, toFormat)
	if err != nil {
		return err
	}

	// Write output file
	if err := os.WriteFile(outputPath, result, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// ConvertReader reads from an io.Reader, converts, and writes to an io.Writer
func (c *Converter) ConvertReader(r io.Reader, w io.Writer, fromFormat, toFormat Format) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	result, err := c.Convert(data, fromFormat, toFormat)
	if err != nil {
		return err
	}

	_, err = w.Write(result)
	return err
}

// DetectFormat attempts to detect the format of the given data
func (c *Converter) DetectFormat(data []byte) (Format, error) {
	if len(data) == 0 {
		return "", ErrEmptyInput
	}

	// Trim whitespace for detection
	data = bytes.TrimSpace(data)

	// Check for JSON
	if (data[0] == '{' || data[0] == '[') && IsValidJSON(data) {
		return FormatJSON, nil
	}

	// Check for XML
	if data[0] == '<' && IsValidXML(data) {
		return FormatXML, nil
	}

	// Check for YAML (heuristic: contains colons but not starting with { or <)
	if IsValidYAML(data) && !IsValidJSON(data) {
		return FormatYAML, nil
	}

	// Default assumption: CSV if comma-separated
	if bytes.Contains(data, []byte(",")) {
		return FormatCSV, nil
	}

	return "", ErrUnsupportedFormat
}

// AutoConvert detects the source format and converts to the target format
func (c *Converter) AutoConvert(data []byte, toFormat Format) ([]byte, error) {
	fromFormat, err := c.DetectFormat(data)
	if err != nil {
		return nil, fmt.Errorf("failed to detect source format: %w", err)
	}

	return c.Convert(data, fromFormat, toFormat)
}

// ConversionResult holds the result of a batch conversion
type ConversionResult struct {
	InputPath  string
	OutputPath string
	Format     Format
	Success    bool
	Error      error
}

// BatchConvert converts multiple files in parallel
func (c *Converter) BatchConvert(inputs []string, outputDir string, toFormat Format) []ConversionResult {
	results := make([]ConversionResult, len(inputs))
	var wg sync.WaitGroup

	for i, inputPath := range inputs {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()

			handler, _ := c.GetHandler(toFormat)
			ext := handler.FileExtension()
			base := filepath.Base(path)
			baseNoExt := base[:len(base)-len(filepath.Ext(base))]
			outputPath := filepath.Join(outputDir, baseNoExt+"."+ext)

			err := c.ConvertFile(path, outputPath, toFormat)
			results[idx] = ConversionResult{
				InputPath:  path,
				OutputPath: outputPath,
				Format:     toFormat,
				Success:    err == nil,
				Error:      err,
			}
		}(i, inputPath)
	}

	wg.Wait()
	return results
}

// Global default converter instance
var defaultConverter = NewConverter()

// Convert uses the default converter to transform data between formats
func Convert(data []byte, fromFormat, toFormat Format) ([]byte, error) {
	return defaultConverter.Convert(data, fromFormat, toFormat)
}

// ConvertTo uses the default converter to serialize a value to a format
func ConvertTo(v interface{}, format Format) ([]byte, error) {
	return defaultConverter.ConvertTo(v, format)
}

// ConvertFrom uses the default converter to deserialize data into a value
func ConvertFrom(data []byte, format Format, v interface{}) error {
	return defaultConverter.ConvertFrom(data, format, v)
}

// AutoConvert uses the default converter to auto-detect and convert
func AutoConvert(data []byte, toFormat Format) ([]byte, error) {
	return defaultConverter.AutoConvert(data, toFormat)
}

// DetectFormat uses the default converter to detect data format
func DetectFormat(data []byte) (Format, error) {
	return defaultConverter.DetectFormat(data)
}

// ToJSON converts any value to JSON bytes
func ToJSON(v interface{}, indent bool) ([]byte, error) {
	h := NewJSONHandler()
	if indent {
		return h.MarshalIndent(v, "", "  ")
	}
	return h.Marshal(v)
}

// ToCSV converts any value to CSV bytes
func ToCSV(v interface{}) ([]byte, error) {
	h := NewCSVHandler()
	return h.Marshal(v)
}

// ToYAML converts any value to YAML bytes
func ToYAML(v interface{}) ([]byte, error) {
	h := NewYAMLHandler()
	return h.Marshal(v)
}

// ToXML converts any value to XML bytes
func ToXML(v interface{}, rootName string, indent bool) ([]byte, error) {
	h := NewXMLHandler()
	h.RootElement = rootName
	if indent {
		return h.MarshalIndent(v, "", "  ")
	}
	return h.Marshal(v)
}

// JSONToYAML converts JSON data to YAML
func JSONToYAML(data []byte) ([]byte, error) {
	return Convert(data, FormatJSON, FormatYAML)
}

// YAMLToJSON converts YAML data to JSON
func YAMLToJSON(data []byte) ([]byte, error) {
	return Convert(data, FormatYAML, FormatJSON)
}

// JSONToXML converts JSON data to XML
func JSONToXML(data []byte) ([]byte, error) {
	return Convert(data, FormatJSON, FormatXML)
}

// XMLToJSON converts XML data to JSON
func XMLToJSON(data []byte) ([]byte, error) {
	return Convert(data, FormatXML, FormatJSON)
}

// JSONToCSV converts JSON data to CSV
func JSONToCSV(data []byte) ([]byte, error) {
	return Convert(data, FormatJSON, FormatCSV)
}

// CSVToJSON converts CSV data to JSON
func CSVToJSON(data []byte) ([]byte, error) {
	return Convert(data, FormatCSV, FormatJSON)
}
