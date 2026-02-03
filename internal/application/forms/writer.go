// Package forms provides form field reading and manipulation functionality.
package forms

import (
	"fmt"

	"github.com/coregx/gxpdf/internal/parser"
)

// FieldUpdate represents a form field value update.
type FieldUpdate struct {
	Name  string
	Value interface{}
}

// Writer writes form field values to a PDF document.
type Writer struct {
	pdfReader *parser.Reader
	updates   map[string]interface{}
}

// NewWriter creates a new form field writer.
func NewWriter(pdfReader *parser.Reader) *Writer {
	return &Writer{
		pdfReader: pdfReader,
		updates:   make(map[string]interface{}),
	}
}

// SetFieldValue sets a form field value by name.
//
// The value type depends on the field type:
//   - Text field: string
//   - Checkbox: bool or string (e.g., "Yes", "Off")
//   - Radio button: string (option name)
//   - Choice field: string or []string (for multi-select)
//
// Returns an error if the field is not found.
func (w *Writer) SetFieldValue(name string, value interface{}) error {
	// Verify field exists
	reader := NewReader(w.pdfReader)
	_, err := reader.GetFieldByName(name)
	if err != nil {
		return err
	}

	w.updates[name] = value
	return nil
}

// GetUpdates returns all pending field updates.
func (w *Writer) GetUpdates() map[string]interface{} {
	return w.updates
}

// HasUpdates returns true if there are pending updates.
func (w *Writer) HasUpdates() bool {
	return len(w.updates) > 0
}

// ValidateFieldValue validates a value against the field type.
func (w *Writer) ValidateFieldValue(name string, value interface{}) error {
	reader := NewReader(w.pdfReader)
	field, err := reader.GetFieldByName(name)
	if err != nil {
		return err
	}

	switch field.Type {
	case FieldTypeText:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("text field %q requires string value, got %T", name, value)
		}

	case FieldTypeButton:
		// Button fields accept bool or string
		switch value.(type) {
		case bool, string:
			// OK
		default:
			return fmt.Errorf("button field %q requires bool or string value, got %T", name, value)
		}

	case FieldTypeChoice:
		// Choice fields accept string or []string
		switch v := value.(type) {
		case string:
			// Validate option exists
			if !w.isValidOption(field, v) {
				return fmt.Errorf("value %q is not a valid option for field %q", v, name)
			}
		case []string:
			// Validate all options exist
			for _, opt := range v {
				if !w.isValidOption(field, opt) {
					return fmt.Errorf("value %q is not a valid option for field %q", opt, name)
				}
			}
		default:
			return fmt.Errorf("choice field %q requires string or []string value, got %T", name, value)
		}

	case FieldTypeSignature:
		return fmt.Errorf("cannot set value for signature field %q", name)

	default:
		// Unknown field type - allow any value
	}

	return nil
}

// isValidOption checks if a value is a valid option for a choice field.
func (w *Writer) isValidOption(field *FieldInfo, value string) bool {
	if len(field.Options) == 0 {
		// No options defined - allow any value
		return true
	}

	for _, opt := range field.Options {
		if opt == value {
			return true
		}
	}
	return false
}

// ApplyUpdatesToDict applies updates to a field dictionary.
//
// This creates a new dictionary with the /V entry updated.
// Used internally when writing the modified PDF.
func (w *Writer) ApplyUpdatesToDict(field *FieldInfo, dict *parser.Dictionary) *parser.Dictionary {
	value, exists := w.updates[field.Name]
	if !exists {
		return dict
	}

	// Create new dictionary with updated value
	newDict := parser.NewDictionary()

	// Copy all existing entries
	for _, key := range dict.Keys() {
		if key != "V" && key != "AS" {
			newDict.Set(key, dict.Get(key))
		}
	}

	// Set new value
	w.setValueInDict(newDict, field.Type, value)

	return newDict
}

// setValueInDict sets the /V entry in a dictionary based on field type.
func (w *Writer) setValueInDict(dict *parser.Dictionary, fieldType FieldType, value interface{}) {
	switch fieldType {
	case FieldTypeText:
		if s, ok := value.(string); ok {
			dict.Set("V", parser.NewString(s))
		}

	case FieldTypeButton:
		switch v := value.(type) {
		case bool:
			if v {
				dict.Set("V", parser.NewName("Yes"))
				dict.Set("AS", parser.NewName("Yes"))
			} else {
				dict.Set("V", parser.NewName("Off"))
				dict.Set("AS", parser.NewName("Off"))
			}
		case string:
			dict.Set("V", parser.NewName(v))
			dict.Set("AS", parser.NewName(v))
		}

	case FieldTypeChoice:
		switch v := value.(type) {
		case string:
			dict.Set("V", parser.NewString(v))
		case []string:
			arr := parser.NewArray()
			for _, s := range v {
				arr.Append(parser.NewString(s))
			}
			dict.Set("V", arr)
		}
	}
}

// GetFieldsToUpdate returns field info for all fields that have updates.
func (w *Writer) GetFieldsToUpdate() ([]*FieldInfo, error) {
	reader := NewReader(w.pdfReader)
	var result []*FieldInfo

	for name := range w.updates {
		field, err := reader.GetFieldByName(name)
		if err != nil {
			return nil, fmt.Errorf("field %q not found: %w", name, err)
		}
		result = append(result, field)
	}

	return result, nil
}
