// Package forms provides form field reading and manipulation functionality.
package forms

import (
	"fmt"

	"github.com/coregx/gxpdf/internal/parser"
)

// FieldType represents the type of a form field.
type FieldType string

const (
	FieldTypeText      FieldType = "Tx"  // Text field
	FieldTypeButton    FieldType = "Btn" // Button (checkbox, radio, pushbutton)
	FieldTypeChoice    FieldType = "Ch"  // Choice (list, combo)
	FieldTypeSignature FieldType = "Sig" // Signature
)

// FieldInfo contains information about a form field.
type FieldInfo struct {
	// Name is the fully qualified field name.
	Name string

	// Type is the field type (Tx, Btn, Ch, Sig).
	Type FieldType

	// Value is the current field value.
	Value interface{}

	// DefaultValue is the default field value.
	DefaultValue interface{}

	// Flags is the field flags bitmask (Ff).
	Flags int

	// Rect is the field rectangle [x1, y1, x2, y2].
	Rect [4]float64

	// Options contains choice field options.
	Options []string

	// ObjectNum is the PDF object number for this field.
	ObjectNum int

	// PageIndex is the page where this field appears (0-based).
	PageIndex int
}

// Reader reads form fields from a PDF document.
type Reader struct {
	pdfReader *parser.Reader
}

// NewReader creates a new form field reader.
func NewReader(pdfReader *parser.Reader) *Reader {
	return &Reader{pdfReader: pdfReader}
}

// GetFields returns all form fields in the document.
func (r *Reader) GetFields() ([]*FieldInfo, error) {
	acroForm, err := r.pdfReader.GetAcroForm()
	if err != nil {
		return nil, fmt.Errorf("failed to get AcroForm: %w", err)
	}

	if acroForm == nil {
		return nil, nil // No form fields
	}

	// Get the Fields array
	fieldsObj := acroForm.Get("Fields")
	if fieldsObj == nil {
		return nil, nil // No fields
	}

	fieldsArray, err := r.pdfReader.ResolveArray(fieldsObj)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Fields array: %w", err)
	}

	var fields []*FieldInfo
	for i := 0; i < fieldsArray.Len(); i++ {
		fieldObj := fieldsArray.Get(i)
		fieldInfos, err := r.parseField(fieldObj, "")
		if err != nil {
			continue // Skip invalid fields
		}
		fields = append(fields, fieldInfos...)
	}

	return fields, nil
}

// GetFieldByName returns a specific field by its fully qualified name.
func (r *Reader) GetFieldByName(name string) (*FieldInfo, error) {
	fields, err := r.GetFields()
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		if field.Name == name {
			return field, nil
		}
	}

	return nil, fmt.Errorf("field not found: %s", name)
}

// parseField parses a field dictionary and its children.
func (r *Reader) parseField(obj parser.PdfObject, parentName string) ([]*FieldInfo, error) {
	// Resolve indirect reference
	obj = r.pdfReader.ResolveReferences(obj)

	dict, ok := obj.(*parser.Dictionary)
	if !ok {
		return nil, fmt.Errorf("field is not a dictionary")
	}

	// Get object number if indirect
	objectNum := 0
	if ref, ok := obj.(*parser.IndirectReference); ok {
		objectNum = ref.Number
	}

	// Build field name
	fieldName := parentName
	if nameObj := dict.Get("T"); nameObj != nil {
		if nameStr, ok := r.pdfReader.ResolveReferences(nameObj).(*parser.String); ok {
			if fieldName != "" {
				fieldName += "."
			}
			fieldName += nameStr.Value()
		}
	}

	// Check for Kids (child fields)
	if kidsObj := dict.Get("Kids"); kidsObj != nil {
		kidsArray, err := r.pdfReader.ResolveArray(kidsObj)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve Kids: %w", err)
		}

		var fields []*FieldInfo
		for i := 0; i < kidsArray.Len(); i++ {
			kidObj := kidsArray.Get(i)
			kidFields, err := r.parseField(kidObj, fieldName)
			if err != nil {
				continue
			}
			fields = append(fields, kidFields...)
		}
		return fields, nil
	}

	// This is a terminal field (no Kids)
	info := &FieldInfo{
		Name:      fieldName,
		ObjectNum: objectNum,
	}

	// Get field type
	if ftObj := dict.Get("FT"); ftObj != nil {
		if ftName, ok := r.pdfReader.ResolveReferences(ftObj).(*parser.Name); ok {
			info.Type = FieldType(ftName.Value())
		}
	}

	// Get field flags
	if ffObj := dict.Get("Ff"); ffObj != nil {
		if ffInt, ok := r.pdfReader.ResolveReferences(ffObj).(*parser.Integer); ok {
			info.Flags = int(ffInt.Value())
		}
	}

	// Get value
	info.Value = r.extractValue(dict, "V")

	// Get default value
	info.DefaultValue = r.extractValue(dict, "DV")

	// Get rectangle
	if rectObj := dict.Get("Rect"); rectObj != nil {
		if rectArray, err := r.pdfReader.ResolveArray(rectObj); err == nil && rectArray.Len() == 4 {
			for i := 0; i < 4; i++ {
				if num := r.extractNumber(rectArray.Get(i)); num != nil {
					info.Rect[i] = *num
				}
			}
		}
	}

	// Get options for choice fields
	if info.Type == FieldTypeChoice {
		if optObj := dict.Get("Opt"); optObj != nil {
			if optArray, err := r.pdfReader.ResolveArray(optObj); err == nil {
				for i := 0; i < optArray.Len(); i++ {
					optItem := r.pdfReader.ResolveReferences(optArray.Get(i))
					if optStr, ok := optItem.(*parser.String); ok {
						info.Options = append(info.Options, optStr.Value())
					} else if optArr, ok := optItem.(*parser.Array); ok && optArr.Len() >= 2 {
						// [export_value, display_value]
						if displayStr, ok := r.pdfReader.ResolveReferences(optArr.Get(1)).(*parser.String); ok {
							info.Options = append(info.Options, displayStr.Value())
						}
					}
				}
			}
		}
	}

	return []*FieldInfo{info}, nil
}

// extractValue extracts a field value from a dictionary.
func (r *Reader) extractValue(dict *parser.Dictionary, key string) interface{} {
	obj := dict.Get(key)
	if obj == nil {
		return nil
	}

	obj = r.pdfReader.ResolveReferences(obj)

	switch v := obj.(type) {
	case *parser.String:
		return v.Value()
	case *parser.Name:
		return v.Value()
	case *parser.Integer:
		return v.Value()
	case *parser.Real:
		return v.Value()
	case *parser.Boolean:
		return v.Value()
	case *parser.Array:
		// Multiple selection
		var values []string
		for i := 0; i < v.Len(); i++ {
			item := r.pdfReader.ResolveReferences(v.Get(i))
			if str, ok := item.(*parser.String); ok {
				values = append(values, str.Value())
			}
		}
		return values
	default:
		return nil
	}
}

// extractNumber extracts a numeric value from a PDF object.
func (r *Reader) extractNumber(obj parser.PdfObject) *float64 {
	obj = r.pdfReader.ResolveReferences(obj)

	switch v := obj.(type) {
	case *parser.Integer:
		val := float64(v.Value())
		return &val
	case *parser.Real:
		val := v.Value()
		return &val
	default:
		return nil
	}
}
