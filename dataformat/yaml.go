package dataformat

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLHandler implements FormatHandler for YAML data
type YAMLHandler struct {
	// IndentSpaces specifies the number of spaces for indentation (default: 2)
	IndentSpaces int
}

// NewYAMLHandler creates a new YAML format handler with default settings
func NewYAMLHandler() *YAMLHandler {
	return &YAMLHandler{
		IndentSpaces: 2,
	}
}

// Marshal converts the given value to YAML bytes
func (h *YAMLHandler) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	return yaml.Marshal(v)
}

// MarshalIndent converts the given value to indented YAML bytes
// Note: YAML is inherently indented, prefix is ignored
func (h *YAMLHandler) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	buf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(len(indent))

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal parses YAML-encoded data and stores the result in v
func (h *YAMLHandler) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return ErrEmptyInput
	}
	if v == nil {
		return ErrNilValue
	}

	return yaml.Unmarshal(data, v)
}

// Format returns FormatYAML
func (h *YAMLHandler) Format() Format {
	return FormatYAML
}

// FileExtension returns "yaml"
func (h *YAMLHandler) FileExtension() string {
	return "yaml"
}

// ContentType returns the MIME type for YAML
func (h *YAMLHandler) ContentType() string {
	return "application/x-yaml"
}

// Encode writes the YAML encoding of v to the writer
func (h *YAMLHandler) Encode(w io.Writer, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(h.IndentSpaces)

	if err := encoder.Encode(v); err != nil {
		return err
	}

	return encoder.Close()
}

// Decode reads YAML-encoded data from the reader and stores it in v
func (h *YAMLHandler) Decode(r io.Reader, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	return yaml.NewDecoder(r).Decode(v)
}

// IsValidYAML checks if the given bytes represent valid YAML
func IsValidYAML(data []byte) bool {
	var result interface{}
	return yaml.Unmarshal(data, &result) == nil
}

// YAMLToMap converts YAML bytes to a map
func YAMLToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MapToYAML converts a map to YAML bytes
func MapToYAML(data map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

// YAMLNode represents a YAML node for advanced manipulation
type YAMLNode struct {
	*yaml.Node
}

// ParseYAMLNode parses YAML data into a node tree for manipulation
func ParseYAMLNode(data []byte) (*YAMLNode, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return &YAMLNode{&node}, nil
}

// Encode writes the node to YAML bytes
func (n *YAMLNode) Encode() ([]byte, error) {
	return yaml.Marshal(n.Node)
}

// MultiDocumentYAML represents multiple YAML documents
type MultiDocumentYAML struct {
	Documents []interface{}
}

// ParseMultiDocYAML parses YAML data containing multiple documents
func ParseMultiDocYAML(data []byte) (*MultiDocumentYAML, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	result := &MultiDocumentYAML{
		Documents: make([]interface{}, 0),
	}

	for {
		var doc interface{}
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		result.Documents = append(result.Documents, doc)
	}

	return result, nil
}

// Encode writes all documents to YAML bytes
func (m *MultiDocumentYAML) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)

	for _, doc := range m.Documents {
		if err := encoder.Encode(doc); err != nil {
			return nil, err
		}
	}

	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
