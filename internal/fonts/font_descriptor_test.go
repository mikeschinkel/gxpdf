package fonts

import (
	"strings"
	"testing"
)

func TestGenerateFontDescriptor(t *testing.T) {
	// Create a mock TTFFont with typical values.
	ttf := &TTFFont{
		FilePath:       "/fonts/OpenSans-Regular.ttf",
		PostScriptName: "OpenSans-Regular",
		UnitsPerEm:     2048,
		FontBBox:       [4]int16{-550, -271, 1204, 1048},
		Ascender:       1069,
		Descender:      -293,
		LineGap:        0,
		ItalicAngle:    0,
		CapHeight:      714,
		XHeight:        519,
		StemV:          80,
		Flags:          32, // Nonsymbolic
	}

	fd := GenerateFontDescriptor(ttf)

	if fd == nil {
		t.Fatal("GenerateFontDescriptor returned nil")
	}

	// Check font name.
	if fd.FontName != "OpenSans-Regular" {
		t.Errorf("FontName = %q, want %q", fd.FontName, "OpenSans-Regular")
	}

	// Check flags.
	if fd.Flags != 32 {
		t.Errorf("Flags = %d, want %d", fd.Flags, 32)
	}

	// Check scaled metrics (1000/2048 scale).
	// Ascent: 1069 * 1000/2048 ≈ 522
	if fd.Ascent < 500 || fd.Ascent > 550 {
		t.Errorf("Ascent = %d, want ~522", fd.Ascent)
	}

	// Descent: -293 * 1000/2048 ≈ -143
	if fd.Descent > -100 || fd.Descent < -180 {
		t.Errorf("Descent = %d, want ~-143", fd.Descent)
	}

	// CapHeight: 714 * 1000/2048 ≈ 349
	if fd.CapHeight < 300 || fd.CapHeight > 400 {
		t.Errorf("CapHeight = %d, want ~349", fd.CapHeight)
	}
}

func TestGenerateFontDescriptor_DeriveNameFromPath(t *testing.T) {
	ttf := &TTFFont{
		FilePath:   "/fonts/MyFont-Bold.ttf",
		UnitsPerEm: 1000,
		FontBBox:   [4]int16{0, -200, 1000, 800},
		Ascender:   800,
		Descender:  -200,
		Flags:      32,
	}

	fd := GenerateFontDescriptor(ttf)

	if fd.FontName != "MyFont-Bold" {
		t.Errorf("FontName = %q, want %q", fd.FontName, "MyFont-Bold")
	}
}

func TestFontDescriptor_ToPDFDict(t *testing.T) {
	fd := &FontDescriptor{
		FontName:    "TestFont",
		Flags:       32,
		FontBBox:    [4]int{0, -200, 1000, 800},
		ItalicAngle: 0,
		Ascent:      800,
		Descent:     -200,
		CapHeight:   700,
		StemV:       80,
		XHeight:     500,
	}

	dict := fd.ToPDFDict(5)

	// Check required entries.
	if !strings.Contains(dict, "/Type /FontDescriptor") {
		t.Error("Missing /Type /FontDescriptor")
	}
	if !strings.Contains(dict, "/FontName /TestFont") {
		t.Error("Missing /FontName")
	}
	if !strings.Contains(dict, "/Flags 32") {
		t.Error("Missing /Flags")
	}
	if !strings.Contains(dict, "/FontBBox [0 -200 1000 800]") {
		t.Error("Missing /FontBBox")
	}
	if !strings.Contains(dict, "/Ascent 800") {
		t.Error("Missing /Ascent")
	}
	if !strings.Contains(dict, "/Descent -200") {
		t.Error("Missing /Descent")
	}
	if !strings.Contains(dict, "/CapHeight 700") {
		t.Error("Missing /CapHeight")
	}
	if !strings.Contains(dict, "/StemV 80") {
		t.Error("Missing /StemV")
	}
	if !strings.Contains(dict, "/FontFile2 5 0 R") {
		t.Error("Missing /FontFile2 reference")
	}
}

func TestSubsetFontName(t *testing.T) {
	name := SubsetFontName("OpenSans-Regular", []rune{'H', 'e', 'l', 'l', 'o'})

	// Should have format XXXXXX+FontName.
	if !strings.Contains(name, "+OpenSans-Regular") {
		t.Errorf("SubsetFontName = %q, missing base name", name)
	}

	// Prefix should be 6 uppercase letters.
	parts := strings.Split(name, "+")
	if len(parts) != 2 {
		t.Fatalf("SubsetFontName = %q, invalid format", name)
	}

	prefix := parts[0]
	if len(prefix) != 6 {
		t.Errorf("Prefix length = %d, want 6", len(prefix))
	}

	for _, c := range prefix {
		if c < 'A' || c > 'Z' {
			t.Errorf("Prefix contains non-uppercase letter: %q", prefix)
			break
		}
	}
}

func TestSubsetFontName_Deterministic(t *testing.T) {
	// Same characters should produce same prefix.
	name1 := SubsetFontName("Font", []rune{'A', 'B', 'C'})
	name2 := SubsetFontName("Font", []rune{'A', 'B', 'C'})

	if name1 != name2 {
		t.Errorf("SubsetFontName not deterministic: %q != %q", name1, name2)
	}

	// Different characters should produce different prefix.
	name3 := SubsetFontName("Font", []rune{'X', 'Y', 'Z'})
	if name1 == name3 {
		t.Errorf("SubsetFontName should differ for different chars: %q == %q", name1, name3)
	}
}
