package dataformat

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

// TestData represents sample data for testing
type TestData struct {
	ID      string   `json:"id" xml:"id" yaml:"id"`
	Name    string   `json:"name" xml:"name" yaml:"name"`
	Value   int      `json:"value" xml:"value" yaml:"value"`
	Active  bool     `json:"active" xml:"active" yaml:"active"`
	Tags    []string `json:"tags" xml:"tags" yaml:"tags"`
}

var sampleData = TestData{
	ID:     "test-123",
	Name:   "Test Item",
	Value:  42,
	Active: true,
	Tags:   []string{"tag1", "tag2"},
}

var sampleSlice = []TestData{
	{ID: "1", Name: "First", Value: 10, Active: true, Tags: []string{"a"}},
	{ID: "2", Name: "Second", Value: 20, Active: false, Tags: []string{"b", "c"}},
}

// ============= Interface Tests =============

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
		hasError bool
	}{
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"csv", FormatCSV, false},
		{"yaml", FormatYAML, false},
		{"yml", FormatYAML, false},
		{"xml", FormatXML, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			format, err := ParseFormat(tt.input)
			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if format != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, format)
			}
		})
	}
}

func TestFormatFromExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected Format
		hasError bool
	}{
		{".json", FormatJSON, false},
		{"json", FormatJSON, false},
		{".csv", FormatCSV, false},
		{".yaml", FormatYAML, false},
		{".yml", FormatYAML, false},
		{".xml", FormatXML, false},
		{".txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			format, err := FormatFromExtension(tt.ext)
			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if format != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, format)
			}
		})
	}
}

func TestIsValidFormat(t *testing.T) {
	if !IsValidFormat("json") {
		t.Error("json should be valid")
	}
	if IsValidFormat("invalid") {
		t.Error("invalid should not be valid")
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) != 4 {
		t.Errorf("expected 4 formats, got %d", len(formats))
	}
}

// ============= JSON Handler Tests =============

func TestJSONHandler_Marshal(t *testing.T) {
	h := NewJSONHandler()

	data, err := h.Marshal(sampleData)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !strings.Contains(string(data), `"id":"test-123"`) {
		t.Error("expected id field in output")
	}

	// Test nil value
	_, err = h.Marshal(nil)
	if err != ErrNilValue {
		t.Error("expected ErrNilValue for nil input")
	}
}

func TestJSONHandler_MarshalIndent(t *testing.T) {
	h := NewJSONHandler()

	data, err := h.MarshalIndent(sampleData, "", "  ")
	if err != nil {
		t.Fatalf("marshal indent error: %v", err)
	}

	if !strings.Contains(string(data), "\n") {
		t.Error("expected indented output with newlines")
	}
}

func TestJSONHandler_Unmarshal(t *testing.T) {
	h := NewJSONHandler()
	jsonData := []byte(`{"id":"test-123","name":"Test Item","value":42,"active":true}`)

	var result TestData
	if err := h.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.ID != "test-123" {
		t.Errorf("expected id test-123, got %s", result.ID)
	}

	// Test empty input
	if err := h.Unmarshal([]byte{}, &result); err != ErrEmptyInput {
		t.Error("expected ErrEmptyInput for empty input")
	}

	// Test nil value
	if err := h.Unmarshal(jsonData, nil); err != ErrNilValue {
		t.Error("expected ErrNilValue for nil target")
	}
}

func TestJSONHandler_EncodeDecode(t *testing.T) {
	h := NewJSONHandler()

	// Encode
	buf := &bytes.Buffer{}
	if err := h.Encode(buf, sampleData); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Decode
	var result TestData
	if err := h.Decode(buf, &result); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if result.ID != sampleData.ID {
		t.Errorf("expected %s, got %s", sampleData.ID, result.ID)
	}
}

func TestJSONHandler_Metadata(t *testing.T) {
	h := NewJSONHandler()

	if h.Format() != FormatJSON {
		t.Error("wrong format")
	}
	if h.FileExtension() != "json" {
		t.Error("wrong extension")
	}
	if h.ContentType() != "application/json" {
		t.Error("wrong content type")
	}
}

func TestIsValidJSON(t *testing.T) {
	if !IsValidJSON([]byte(`{"key": "value"}`)) {
		t.Error("should be valid JSON")
	}
	if IsValidJSON([]byte(`{invalid}`)) {
		t.Error("should be invalid JSON")
	}
}

func TestPrettyPrintJSON(t *testing.T) {
	compact := []byte(`{"a":"b","c":"d"}`)
	pretty, err := PrettyPrintJSON(compact)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(string(pretty), "\n") {
		t.Error("expected formatted output")
	}
}

func TestCompactJSON(t *testing.T) {
	pretty := []byte(`{
		"a": "b",
		"c": "d"
	}`)
	compact, err := CompactJSON(pretty)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if strings.Contains(string(compact), "\n") {
		t.Error("expected compact output without newlines")
	}
}

// ============= CSV Handler Tests =============

func TestCSVHandler_Marshal(t *testing.T) {
	h := NewCSVHandler()

	data, err := h.Marshal(sampleSlice)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		t.Error("expected at least header and one data row")
	}
}

func TestCSVHandler_Unmarshal(t *testing.T) {
	h := NewCSVHandler()
	csvData := []byte("id,name,value\n1,First,10\n2,Second,20")

	var result []map[string]interface{}
	if err := h.Unmarshal(csvData, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 records, got %d", len(result))
	}
}

func TestCSVHandler_Metadata(t *testing.T) {
	h := NewCSVHandler()

	if h.Format() != FormatCSV {
		t.Error("wrong format")
	}
	if h.FileExtension() != "csv" {
		t.Error("wrong extension")
	}
	if h.ContentType() != "text/csv" {
		t.Error("wrong content type")
	}
}

func TestCSVToMaps(t *testing.T) {
	csvData := []byte("a,b,c\n1,2,3\n4,5,6")
	maps, err := CSVToMaps(csvData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(maps) != 2 {
		t.Errorf("expected 2 maps, got %d", len(maps))
	}

	if maps[0]["a"] != "1" {
		t.Errorf("expected 1, got %s", maps[0]["a"])
	}
}

func TestMapsToCSV(t *testing.T) {
	data := []map[string]string{
		{"a": "1", "b": "2"},
		{"a": "3", "b": "4"},
	}

	csvBytes, err := MapsToCSV(data, []string{"a", "b"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !strings.Contains(string(csvBytes), "a,b") {
		t.Error("expected header in output")
	}
}

func TestGetCSVHeaders(t *testing.T) {
	csvData := []byte("x,y,z\n1,2,3")
	headers, err := GetCSVHeaders(csvData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !reflect.DeepEqual(headers, []string{"x", "y", "z"}) {
		t.Errorf("unexpected headers: %v", headers)
	}
}

// ============= YAML Handler Tests =============

func TestYAMLHandler_Marshal(t *testing.T) {
	h := NewYAMLHandler()

	data, err := h.Marshal(sampleData)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !strings.Contains(string(data), "id: test-123") {
		t.Error("expected id field in YAML output")
	}
}

func TestYAMLHandler_Unmarshal(t *testing.T) {
	h := NewYAMLHandler()
	yamlData := []byte("id: test-123\nname: Test Item\nvalue: 42\nactive: true")

	var result TestData
	if err := h.Unmarshal(yamlData, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.ID != "test-123" {
		t.Errorf("expected id test-123, got %s", result.ID)
	}
}

func TestYAMLHandler_EncodeDecode(t *testing.T) {
	h := NewYAMLHandler()

	// Encode
	buf := &bytes.Buffer{}
	if err := h.Encode(buf, sampleData); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Decode
	var result TestData
	if err := h.Decode(bytes.NewReader(buf.Bytes()), &result); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if result.Name != sampleData.Name {
		t.Errorf("expected %s, got %s", sampleData.Name, result.Name)
	}
}

func TestYAMLHandler_Metadata(t *testing.T) {
	h := NewYAMLHandler()

	if h.Format() != FormatYAML {
		t.Error("wrong format")
	}
	if h.FileExtension() != "yaml" {
		t.Error("wrong extension")
	}
	if h.ContentType() != "application/x-yaml" {
		t.Error("wrong content type")
	}
}

func TestIsValidYAML(t *testing.T) {
	if !IsValidYAML([]byte("key: value")) {
		t.Error("should be valid YAML")
	}
}

func TestYAMLToMap(t *testing.T) {
	yamlData := []byte("a: 1\nb: 2")
	m, err := YAMLToMap(yamlData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if m["a"] != 1 {
		t.Errorf("expected a=1, got %v", m["a"])
	}
}

func TestMultiDocYAML(t *testing.T) {
	yamlData := []byte("---\na: 1\n---\nb: 2")
	docs, err := ParseMultiDocYAML(yamlData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(docs.Documents) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs.Documents))
	}
}

// ============= XML Handler Tests =============

func TestXMLHandler_Marshal(t *testing.T) {
	h := NewXMLHandler()

	data, err := h.Marshal(sampleData)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !strings.Contains(string(data), "<?xml") {
		t.Error("expected XML declaration")
	}
}

func TestXMLHandler_MarshalMap(t *testing.T) {
	h := NewXMLHandler()
	h.RootElement = "data"

	mapData := map[string]interface{}{
		"id":   "123",
		"name": "Test",
	}

	data, err := h.Marshal(mapData)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if !strings.Contains(string(data), "<data>") {
		t.Error("expected root element")
	}
}

func TestXMLHandler_Unmarshal(t *testing.T) {
	h := NewXMLHandler()
	xmlData := []byte(`<TestData><ID>test-123</ID><Name>Test</Name></TestData>`)

	type SimpleXML struct {
		ID   string `xml:"ID"`
		Name string `xml:"Name"`
	}

	var result SimpleXML
	if err := h.Unmarshal(xmlData, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.ID != "test-123" {
		t.Errorf("expected id test-123, got %s", result.ID)
	}
}

func TestXMLHandler_Metadata(t *testing.T) {
	h := NewXMLHandler()

	if h.Format() != FormatXML {
		t.Error("wrong format")
	}
	if h.FileExtension() != "xml" {
		t.Error("wrong extension")
	}
	if h.ContentType() != "application/xml" {
		t.Error("wrong content type")
	}
}

func TestIsValidXML(t *testing.T) {
	if !IsValidXML([]byte(`<root><child/></root>`)) {
		t.Error("should be valid XML")
	}
	if IsValidXML([]byte(`<unclosed>`)) {
		t.Error("should be invalid XML")
	}
}

func TestSanitizeXMLName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"valid", "valid"},
		{"with space", "with_space"},
		{"123start", "_23start"},
		{"", "element"},
		{"with-dash", "with-dash"},
		{"with.dot", "with.dot"},
	}

	for _, tt := range tests {
		result := sanitizeXMLName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeXMLName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// ============= Converter Tests =============

func TestNewConverter(t *testing.T) {
	c := NewConverter()

	// Check all handlers registered
	for _, format := range SupportedFormats() {
		h, err := c.GetHandler(format)
		if err != nil {
			t.Errorf("handler not found for %s", format)
		}
		if h.Format() != format {
			t.Errorf("wrong handler for %s", format)
		}
	}
}

func TestConverter_Convert(t *testing.T) {
	c := NewConverter()

	jsonData := []byte(`{"id":"test","name":"Test Item"}`)

	// JSON to YAML
	yamlData, err := c.Convert(jsonData, FormatJSON, FormatYAML)
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}
	if !strings.Contains(string(yamlData), "id: test") {
		t.Error("expected YAML output")
	}

	// JSON to XML
	xmlData, err := c.Convert(jsonData, FormatJSON, FormatXML)
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}
	if !strings.Contains(string(xmlData), "<id>") {
		t.Error("expected XML output")
	}
}

func TestConverter_ConvertTo(t *testing.T) {
	c := NewConverter()

	// Convert struct to different formats
	jsonData, err := c.ConvertTo(sampleData, FormatJSON)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !IsValidJSON(jsonData) {
		t.Error("expected valid JSON")
	}

	yamlData, err := c.ConvertTo(sampleData, FormatYAML)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(string(yamlData), "id:") {
		t.Error("expected YAML output")
	}
}

func TestConverter_ConvertFrom(t *testing.T) {
	c := NewConverter()

	jsonData := []byte(`{"id":"test-123","name":"Test"}`)

	var result map[string]interface{}
	if err := c.ConvertFrom(jsonData, FormatJSON, &result); err != nil {
		t.Fatalf("error: %v", err)
	}

	if result["id"] != "test-123" {
		t.Error("expected id to be test-123")
	}
}

func TestConverter_DetectFormat(t *testing.T) {
	c := NewConverter()

	tests := []struct {
		name     string
		data     []byte
		expected Format
	}{
		{"JSON object", []byte(`{"key": "value"}`), FormatJSON},
		{"JSON array", []byte(`[1, 2, 3]`), FormatJSON},
		{"XML", []byte(`<?xml version="1.0"?><root/>`), FormatXML},
		{"YAML", []byte(`key: value`), FormatYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := c.DetectFormat(tt.data)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if format != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, format)
			}
		})
	}
}

func TestConverter_AutoConvert(t *testing.T) {
	c := NewConverter()

	jsonData := []byte(`{"key": "value"}`)
	yamlData, err := c.AutoConvert(jsonData, FormatYAML)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !strings.Contains(string(yamlData), "key:") {
		t.Error("expected YAML output")
	}
}

func TestConverter_SetOptions(t *testing.T) {
	c := NewConverter()

	opts := ConvertOptions{
		Indent:       false,
		IndentString: "\t",
	}
	c.SetOptions(opts)

	gotOpts := c.GetOptions()
	if gotOpts.Indent != false {
		t.Error("expected indent to be false")
	}
}

// ============= Global Function Tests =============

func TestGlobalConvert(t *testing.T) {
	jsonData := []byte(`{"a": 1}`)
	yamlData, err := Convert(jsonData, FormatJSON, FormatYAML)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(string(yamlData), "a:") {
		t.Error("expected YAML output")
	}
}

func TestGlobalConvertTo(t *testing.T) {
	data := map[string]int{"a": 1}
	jsonData, err := ConvertTo(data, FormatJSON)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !IsValidJSON(jsonData) {
		t.Error("expected valid JSON")
	}
}

func TestGlobalConvertFrom(t *testing.T) {
	jsonData := []byte(`{"a": 1}`)
	var result map[string]interface{}
	if err := ConvertFrom(jsonData, FormatJSON, &result); err != nil {
		t.Fatalf("error: %v", err)
	}
	if result["a"] != float64(1) {
		t.Error("expected a to be 1")
	}
}

func TestHelperFunctions(t *testing.T) {
	data := map[string]int{"x": 1}

	// ToJSON
	jsonData, err := ToJSON(data, true)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if !IsValidJSON(jsonData) {
		t.Error("expected valid JSON")
	}

	// ToYAML
	yamlData, err := ToYAML(data)
	if err != nil {
		t.Fatalf("ToYAML error: %v", err)
	}
	if !strings.Contains(string(yamlData), "x:") {
		t.Error("expected YAML output")
	}

	// ToCSV
	sliceData := []map[string]int{{"x": 1}, {"x": 2}}
	csvData, err := ToCSV(sliceData)
	if err != nil {
		t.Fatalf("ToCSV error: %v", err)
	}
	if !strings.Contains(string(csvData), "x") {
		t.Error("expected CSV output")
	}

	// ToXML
	xmlData, err := ToXML(data, "root", true)
	if err != nil {
		t.Fatalf("ToXML error: %v", err)
	}
	if !strings.Contains(string(xmlData), "<root>") {
		t.Error("expected XML output")
	}
}

func TestConversionHelpers(t *testing.T) {
	jsonData := []byte(`{"key": "value"}`)

	// JSONToYAML
	yamlData, err := JSONToYAML(jsonData)
	if err != nil {
		t.Fatalf("JSONToYAML error: %v", err)
	}

	// YAMLToJSON
	jsonBack, err := YAMLToJSON(yamlData)
	if err != nil {
		t.Fatalf("YAMLToJSON error: %v", err)
	}
	if !IsValidJSON(jsonBack) {
		t.Error("expected valid JSON")
	}

	// JSONToXML
	xmlData, err := JSONToXML(jsonData)
	if err != nil {
		t.Fatalf("JSONToXML error: %v", err)
	}
	if !IsValidXML(xmlData) {
		t.Error("expected valid XML")
	}

	// JSONToCSV
	jsonArray := []byte(`[{"a": "1"}, {"a": "2"}]`)
	csvData, err := JSONToCSV(jsonArray)
	if err != nil {
		t.Fatalf("JSONToCSV error: %v", err)
	}

	// CSVToJSON
	jsonFromCSV, err := CSVToJSON(csvData)
	if err != nil {
		t.Fatalf("CSVToJSON error: %v", err)
	}
	if !IsValidJSON(jsonFromCSV) {
		t.Error("expected valid JSON")
	}
}

// ============= Error Handling Tests =============

func TestErrorHandling(t *testing.T) {
	c := NewConverter()

	// Empty input
	_, err := c.Convert([]byte{}, FormatJSON, FormatYAML)
	if err == nil {
		t.Error("expected error for empty input")
	}

	// Unsupported format
	_, err = c.GetHandler("invalid")
	if err != ErrUnsupportedFormat {
		t.Error("expected ErrUnsupportedFormat")
	}

	// Nil value
	_, err = c.ConvertTo(nil, FormatJSON)
	if err != ErrNilValue {
		t.Error("expected ErrNilValue")
	}
}

// ============= Benchmark Tests =============

func BenchmarkJSONMarshal(b *testing.B) {
	h := NewJSONHandler()
	for i := 0; i < b.N; i++ {
		h.Marshal(sampleData)
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	h := NewJSONHandler()
	data, _ := h.Marshal(sampleData)
	var result TestData
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Unmarshal(data, &result)
	}
}

func BenchmarkConvert_JSONToYAML(b *testing.B) {
	c := NewConverter()
	data, _ := ToJSON(sampleData, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Convert(data, FormatJSON, FormatYAML)
	}
}

func BenchmarkConvert_JSONToXML(b *testing.B) {
	c := NewConverter()
	data, _ := ToJSON(sampleData, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Convert(data, FormatJSON, FormatXML)
	}
}
