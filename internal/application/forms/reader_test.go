package forms

import (
	"testing"
)

func TestFieldType(t *testing.T) {
	tests := []struct {
		name string
		ft   FieldType
		want string
	}{
		{"Text", FieldTypeText, "Tx"},
		{"Button", FieldTypeButton, "Btn"},
		{"Choice", FieldTypeChoice, "Ch"},
		{"Signature", FieldTypeSignature, "Sig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.ft) != tt.want {
				t.Errorf("FieldType = %q, want %q", tt.ft, tt.want)
			}
		})
	}
}

func TestFieldInfo(t *testing.T) {
	info := &FieldInfo{
		Name:         "test_field",
		Type:         FieldTypeText,
		Value:        "Hello",
		DefaultValue: "Default",
		Flags:        3, // ReadOnly + Required
		Rect:         [4]float64{100, 200, 300, 220},
		Options:      nil,
	}

	if info.Name != "test_field" {
		t.Errorf("Name = %q, want %q", info.Name, "test_field")
	}

	if info.Type != FieldTypeText {
		t.Errorf("Type = %q, want %q", info.Type, FieldTypeText)
	}

	if info.Value != "Hello" {
		t.Errorf("Value = %v, want %q", info.Value, "Hello")
	}

	if info.Flags != 3 {
		t.Errorf("Flags = %d, want %d", info.Flags, 3)
	}

	// Check ReadOnly flag (bit 0)
	isReadOnly := info.Flags&1 != 0
	if !isReadOnly {
		t.Error("Expected ReadOnly flag to be set")
	}

	// Check Required flag (bit 1)
	isRequired := info.Flags&2 != 0
	if !isRequired {
		t.Error("Expected Required flag to be set")
	}
}

func TestFieldInfoWithOptions(t *testing.T) {
	info := &FieldInfo{
		Name:    "dropdown",
		Type:    FieldTypeChoice,
		Value:   "Option B",
		Options: []string{"Option A", "Option B", "Option C"},
	}

	if len(info.Options) != 3 {
		t.Errorf("Options length = %d, want %d", len(info.Options), 3)
	}

	if info.Options[1] != "Option B" {
		t.Errorf("Options[1] = %q, want %q", info.Options[1], "Option B")
	}
}
