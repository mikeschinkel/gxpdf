package forms

import (
	"testing"

	"github.com/coregx/gxpdf/internal/parser"
)

// mockReader creates a mock parser.Reader for testing.
func mockReader() *parser.Reader {
	// Create a minimal mock reader
	// In real usage, this would be a full PDF parser
	return nil
}

func TestNewWriter(t *testing.T) {
	reader := mockReader()
	writer := NewWriter(reader)

	if writer == nil {
		t.Fatal("NewWriter returned nil")
	}

	if writer.pdfReader != reader {
		t.Error("Writer.pdfReader not set correctly")
	}

	if writer.updates == nil {
		t.Error("Writer.updates map not initialized")
	}

	if len(writer.updates) != 0 {
		t.Errorf("Writer.updates length = %d, want 0", len(writer.updates))
	}
}

func TestWriter_HasUpdates(t *testing.T) {
	writer := NewWriter(mockReader())

	if writer.HasUpdates() {
		t.Error("HasUpdates() = true, want false for new writer")
	}

	// Add an update
	writer.updates["test"] = "value"

	if !writer.HasUpdates() {
		t.Error("HasUpdates() = false, want true after adding update")
	}
}

func TestWriter_GetUpdates(t *testing.T) {
	writer := NewWriter(mockReader())

	updates := writer.GetUpdates()
	if updates == nil {
		t.Fatal("GetUpdates() returned nil")
	}

	if len(updates) != 0 {
		t.Errorf("GetUpdates() length = %d, want 0", len(updates))
	}

	// Add updates
	writer.updates["field1"] = "value1"
	writer.updates["field2"] = true

	updates = writer.GetUpdates()
	if len(updates) != 2 {
		t.Errorf("GetUpdates() length = %d, want 2", len(updates))
	}

	if updates["field1"] != "value1" {
		t.Errorf("GetUpdates()[field1] = %v, want %q", updates["field1"], "value1")
	}

	if updates["field2"] != true {
		t.Errorf("GetUpdates()[field2] = %v, want true", updates["field2"])
	}
}

func TestWriter_ValidateFieldValue_TextType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "valid string for text field",
			fieldType: FieldTypeText,
			value:     "test value",
			wantErr:   false,
		},
		{
			name:      "invalid int for text field",
			fieldType: FieldTypeText,
			value:     123,
			wantErr:   true,
		},
		{
			name:      "invalid bool for text field",
			fieldType: FieldTypeText,
			value:     true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock field with validation
			field := &FieldInfo{
				Name: "test_field",
				Type: tt.fieldType,
			}

			// Validate directly
			err := validateFieldValueType(field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldValueType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_ValidateFieldValue_ButtonType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "valid bool for button field",
			fieldType: FieldTypeButton,
			value:     true,
			wantErr:   false,
		},
		{
			name:      "valid string for button field",
			fieldType: FieldTypeButton,
			value:     "Yes",
			wantErr:   false,
		},
		{
			name:      "invalid int for button field",
			fieldType: FieldTypeButton,
			value:     1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &FieldInfo{
				Name: "checkbox_field",
				Type: tt.fieldType,
			}

			err := validateFieldValueType(field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldValueType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriter_ValidateFieldValue_ChoiceType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		value     interface{}
		options   []string
		wantErr   bool
	}{
		{
			name:      "valid string option",
			fieldType: FieldTypeChoice,
			value:     "Option A",
			options:   []string{"Option A", "Option B"},
			wantErr:   false,
		},
		{
			name:      "invalid string option",
			fieldType: FieldTypeChoice,
			value:     "Option C",
			options:   []string{"Option A", "Option B"},
			wantErr:   true,
		},
		{
			name:      "valid string array options",
			fieldType: FieldTypeChoice,
			value:     []string{"Option A", "Option B"},
			options:   []string{"Option A", "Option B"},
			wantErr:   false,
		},
		{
			name:      "invalid string array with bad option",
			fieldType: FieldTypeChoice,
			value:     []string{"Option A", "Option C"},
			options:   []string{"Option A", "Option B"},
			wantErr:   true,
		},
		{
			name:      "no options defined - allow any",
			fieldType: FieldTypeChoice,
			value:     "Any Value",
			options:   nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewWriter(mockReader())
			field := &FieldInfo{
				Name:    "choice_field",
				Type:    tt.fieldType,
				Options: tt.options,
			}

			// Test the isValidOption method
			if tt.value != nil {
				switch v := tt.value.(type) {
				case string:
					valid := writer.isValidOption(field, v)
					if tt.wantErr && valid {
						t.Error("isValidOption() = true, want false")
					}
					if !tt.wantErr && !valid && len(tt.options) > 0 {
						t.Error("isValidOption() = false, want true")
					}
				case []string:
					allValid := true
					for _, val := range v {
						if !writer.isValidOption(field, val) {
							allValid = false
							break
						}
					}
					if tt.wantErr && allValid {
						t.Error("All options valid, want at least one invalid")
					}
					if !tt.wantErr && !allValid && len(tt.options) > 0 {
						t.Error("Some options invalid, want all valid")
					}
				}
			}
		})
	}
}

func TestWriter_ValidateFieldValue_SignatureType(t *testing.T) {
	field := &FieldInfo{
		Name: "signature_field",
		Type: FieldTypeSignature,
	}

	err := validateFieldValueType(field, "any_value")
	if err == nil {
		t.Error("Expected error for signature field, got nil")
	}
}

func TestWriter_setValueInDict_Text(t *testing.T) {
	writer := NewWriter(mockReader())
	dict := parser.NewDictionary()

	writer.setValueInDict(dict, FieldTypeText, "test value")

	valueObj := dict.Get("V")
	if valueObj == nil {
		t.Fatal("Value not set in dictionary")
	}

	strObj, ok := valueObj.(*parser.String)
	if !ok {
		t.Fatalf("Value is not a String, got %T", valueObj)
	}

	if strObj.Value() != "test value" {
		t.Errorf("String value = %q, want %q", strObj.Value(), "test value")
	}
}

func TestWriter_setValueInDict_ButtonBool(t *testing.T) {
	writer := NewWriter(mockReader())
	dict := parser.NewDictionary()

	// Test true value
	writer.setValueInDict(dict, FieldTypeButton, true)

	valueObj := dict.Get("V")
	if valueObj == nil {
		t.Fatal("Value not set in dictionary")
	}

	nameObj, ok := valueObj.(*parser.Name)
	if !ok {
		t.Fatalf("Value is not a Name, got %T", valueObj)
	}

	if nameObj.Value() != "Yes" {
		t.Errorf("Name value = %q, want %q", nameObj.Value(), "Yes")
	}

	asObj := dict.Get("AS")
	if asObj == nil {
		t.Fatal("AS not set in dictionary")
	}

	// Test false value
	dict = parser.NewDictionary()
	writer.setValueInDict(dict, FieldTypeButton, false)

	valueObj = dict.Get("V")
	nameObj, ok = valueObj.(*parser.Name)
	if !ok {
		t.Fatalf("Value is not a Name, got %T", valueObj)
	}

	if nameObj.Value() != "Off" {
		t.Errorf("Name value = %q, want %q", nameObj.Value(), "Off")
	}
}

func TestWriter_setValueInDict_ButtonString(t *testing.T) {
	writer := NewWriter(mockReader())
	dict := parser.NewDictionary()

	writer.setValueInDict(dict, FieldTypeButton, "CustomValue")

	valueObj := dict.Get("V")
	nameObj, ok := valueObj.(*parser.Name)
	if !ok {
		t.Fatalf("Value is not a Name, got %T", valueObj)
	}

	if nameObj.Value() != "CustomValue" {
		t.Errorf("Name value = %q, want %q", nameObj.Value(), "CustomValue")
	}
}

func TestWriter_setValueInDict_ChoiceString(t *testing.T) {
	writer := NewWriter(mockReader())
	dict := parser.NewDictionary()

	writer.setValueInDict(dict, FieldTypeChoice, "Option A")

	valueObj := dict.Get("V")
	strObj, ok := valueObj.(*parser.String)
	if !ok {
		t.Fatalf("Value is not a String, got %T", valueObj)
	}

	if strObj.Value() != "Option A" {
		t.Errorf("String value = %q, want %q", strObj.Value(), "Option A")
	}
}

func TestWriter_setValueInDict_ChoiceArray(t *testing.T) {
	writer := NewWriter(mockReader())
	dict := parser.NewDictionary()

	writer.setValueInDict(dict, FieldTypeChoice, []string{"Option A", "Option B"})

	valueObj := dict.Get("V")
	arrObj, ok := valueObj.(*parser.Array)
	if !ok {
		t.Fatalf("Value is not an Array, got %T", valueObj)
	}

	if arrObj.Len() != 2 {
		t.Errorf("Array length = %d, want 2", arrObj.Len())
	}

	// Check first element
	elem0 := arrObj.Get(0)
	str0, ok := elem0.(*parser.String)
	if !ok {
		t.Fatalf("Array[0] is not a String, got %T", elem0)
	}

	if str0.Value() != "Option A" {
		t.Errorf("Array[0] value = %q, want %q", str0.Value(), "Option A")
	}

	// Check second element
	elem1 := arrObj.Get(1)
	str1, ok := elem1.(*parser.String)
	if !ok {
		t.Fatalf("Array[1] is not a String, got %T", elem1)
	}

	if str1.Value() != "Option B" {
		t.Errorf("Array[1] value = %q, want %q", str1.Value(), "Option B")
	}
}

// Helper function for validation without needing a full reader
func validateFieldValueType(field *FieldInfo, value interface{}) error {
	switch field.Type {
	case FieldTypeText:
		if _, ok := value.(string); !ok {
			return &validationError{fieldName: field.Name, expectedType: "string"}
		}
	case FieldTypeButton:
		switch value.(type) {
		case bool, string:
			// OK
		default:
			return &validationError{fieldName: field.Name, expectedType: "bool or string"}
		}
	case FieldTypeSignature:
		return &validationError{fieldName: field.Name, expectedType: "signature fields cannot be set"}
	}
	return nil
}

type validationError struct {
	fieldName    string
	expectedType string
}

func (e *validationError) Error() string {
	return "validation error for field " + e.fieldName + ": expected " + e.expectedType
}
