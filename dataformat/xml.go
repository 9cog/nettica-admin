package dataformat

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// XMLHandler implements FormatHandler for XML data
type XMLHandler struct {
	// RootElement specifies the root element name (default: "root")
	RootElement string

	// ItemElement specifies the element name for array items (default: "item")
	ItemElement string

	// IncludeHeader specifies whether to include XML declaration
	IncludeHeader bool

	// IndentString specifies the indentation string (default: "  ")
	IndentString string
}

// NewXMLHandler creates a new XML format handler with default settings
func NewXMLHandler() *XMLHandler {
	return &XMLHandler{
		RootElement:   "root",
		ItemElement:   "item",
		IncludeHeader: true,
		IndentString:  "  ",
	}
}

// Marshal converts the given value to XML bytes
func (h *XMLHandler) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	buf := &bytes.Buffer{}
	if h.IncludeHeader {
		buf.WriteString(xml.Header)
	}

	// Try standard XML marshal first
	data, err := xml.Marshal(v)
	if err == nil {
		buf.Write(data)
		return buf.Bytes(), nil
	}

	// Fall back to generic conversion for maps/slices
	return h.marshalGeneric(v)
}

// MarshalIndent converts the given value to indented XML bytes
func (h *XMLHandler) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	if v == nil {
		return nil, ErrNilValue
	}

	buf := &bytes.Buffer{}
	if h.IncludeHeader {
		buf.WriteString(xml.Header)
	}

	// Try standard XML marshal first
	data, err := xml.MarshalIndent(v, prefix, indent)
	if err == nil {
		buf.Write(data)
		return buf.Bytes(), nil
	}

	// Fall back to generic conversion
	return h.marshalGenericIndent(v, prefix, indent)
}

// marshalGeneric handles marshaling for non-struct types (maps, slices)
func (h *XMLHandler) marshalGeneric(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	if h.IncludeHeader {
		buf.WriteString(xml.Header)
	}

	// Convert to intermediate representation via JSON
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	buf.WriteString("<" + h.RootElement + ">")
	h.writeValue(buf, data, "")
	buf.WriteString("</" + h.RootElement + ">")

	return buf.Bytes(), nil
}

// marshalGenericIndent handles indented marshaling for non-struct types
func (h *XMLHandler) marshalGenericIndent(v interface{}, prefix, indent string) ([]byte, error) {
	buf := &bytes.Buffer{}
	if h.IncludeHeader {
		buf.WriteString(xml.Header)
	}

	// Convert to intermediate representation via JSON
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	buf.WriteString("<" + h.RootElement + ">\n")
	h.writeValueIndent(buf, data, prefix+indent, indent)
	buf.WriteString(prefix + "</" + h.RootElement + ">")

	return buf.Bytes(), nil
}

// writeValue recursively writes XML for generic values
func (h *XMLHandler) writeValue(buf *bytes.Buffer, v interface{}, elementName string) {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			buf.WriteString("<" + sanitizeXMLName(k) + ">")
			h.writeValue(buf, val[k], k)
			buf.WriteString("</" + sanitizeXMLName(k) + ">")
		}
	case []interface{}:
		for _, item := range val {
			buf.WriteString("<" + h.ItemElement + ">")
			h.writeValue(buf, item, h.ItemElement)
			buf.WriteString("</" + h.ItemElement + ">")
		}
	case string:
		xml.EscapeText(buf, []byte(val))
	case nil:
		// Empty element
	default:
		buf.WriteString(fmt.Sprintf("%v", val))
	}
}

// writeValueIndent recursively writes indented XML for generic values
func (h *XMLHandler) writeValueIndent(buf *bytes.Buffer, v interface{}, currentIndent, indent string) {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			buf.WriteString(currentIndent + "<" + sanitizeXMLName(k) + ">")
			if isSimpleValue(val[k]) {
				h.writeValue(buf, val[k], k)
				buf.WriteString("</" + sanitizeXMLName(k) + ">\n")
			} else {
				buf.WriteString("\n")
				h.writeValueIndent(buf, val[k], currentIndent+indent, indent)
				buf.WriteString(currentIndent + "</" + sanitizeXMLName(k) + ">\n")
			}
		}
	case []interface{}:
		for _, item := range val {
			buf.WriteString(currentIndent + "<" + h.ItemElement + ">")
			if isSimpleValue(item) {
				h.writeValue(buf, item, h.ItemElement)
				buf.WriteString("</" + h.ItemElement + ">\n")
			} else {
				buf.WriteString("\n")
				h.writeValueIndent(buf, item, currentIndent+indent, indent)
				buf.WriteString(currentIndent + "</" + h.ItemElement + ">\n")
			}
		}
	case string:
		xml.EscapeText(buf, []byte(val))
	case nil:
		// Empty element
	default:
		buf.WriteString(fmt.Sprintf("%v", val))
	}
}

// Unmarshal parses XML-encoded data and stores the result in v
func (h *XMLHandler) Unmarshal(data []byte, v interface{}) error {
	if len(data) == 0 {
		return ErrEmptyInput
	}
	if v == nil {
		return ErrNilValue
	}

	return xml.Unmarshal(data, v)
}

// Format returns FormatXML
func (h *XMLHandler) Format() Format {
	return FormatXML
}

// FileExtension returns "xml"
func (h *XMLHandler) FileExtension() string {
	return "xml"
}

// ContentType returns the MIME type for XML
func (h *XMLHandler) ContentType() string {
	return "application/xml"
}

// Encode writes the XML encoding of v to the writer
func (h *XMLHandler) Encode(w io.Writer, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	data, err := h.MarshalIndent(v, "", h.IndentString)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// Decode reads XML-encoded data from the reader and stores it in v
func (h *XMLHandler) Decode(r io.Reader, v interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	return xml.NewDecoder(r).Decode(v)
}

// sanitizeXMLName ensures a string is valid as an XML element name
func sanitizeXMLName(s string) string {
	if s == "" {
		return "element"
	}

	// Replace invalid characters with underscores
	result := strings.Builder{}
	for i, r := range s {
		if i == 0 {
			// First character must be a letter or underscore
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' {
				result.WriteRune(r)
			} else {
				result.WriteRune('_')
			}
		} else {
			// Subsequent characters can also include digits, hyphens, and periods
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
				r == '_' || r == '-' || r == '.' {
				result.WriteRune(r)
			} else {
				result.WriteRune('_')
			}
		}
	}

	return result.String()
}

// isSimpleValue checks if a value is a simple (non-nested) type
func isSimpleValue(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}, []interface{}:
		return false
	default:
		return true
	}
}

// XMLElement represents a generic XML element for flexible parsing
type XMLElement struct {
	XMLName  xml.Name
	Attrs    []xml.Attr     `xml:",any,attr"`
	Content  string         `xml:",chardata"`
	Children []*XMLElement  `xml:",any"`
}

// IsValidXML checks if the given bytes represent valid XML
func IsValidXML(data []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
	}
}

// XMLToMap converts XML bytes to a map (best effort)
func XMLToMap(data []byte) (map[string]interface{}, error) {
	var root XMLElement
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	return elementToMap(&root), nil
}

// elementToMap recursively converts XMLElement to map
func elementToMap(elem *XMLElement) map[string]interface{} {
	result := make(map[string]interface{})

	if len(elem.Children) == 0 {
		// Leaf node with text content
		content := strings.TrimSpace(elem.Content)
		if content != "" {
			return map[string]interface{}{elem.XMLName.Local: content}
		}
		return result
	}

	// Build children map
	childMap := make(map[string]interface{})
	for _, child := range elem.Children {
		childData := elementToMap(child)
		for k, v := range childData {
			if existing, ok := childMap[k]; ok {
				// Convert to slice if multiple elements with same name
				switch e := existing.(type) {
				case []interface{}:
					childMap[k] = append(e, v)
				default:
					childMap[k] = []interface{}{e, v}
				}
			} else {
				childMap[k] = v
			}
		}
	}

	if elem.XMLName.Local != "" {
		result[elem.XMLName.Local] = childMap
	} else {
		return childMap
	}

	return result
}

// MapToXML converts a map to XML bytes with the given root element name
func MapToXML(data map[string]interface{}, rootName string, indent bool) ([]byte, error) {
	handler := NewXMLHandler()
	handler.RootElement = rootName

	if indent {
		return handler.MarshalIndent(data, "", "  ")
	}
	return handler.Marshal(data)
}
