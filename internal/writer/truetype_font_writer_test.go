package writer

import (
	"strings"
	"testing"

	"github.com/coregx/gxpdf/internal/fonts"
)

func TestTrueTypeFontWriter_WriteFont(t *testing.T) {
	// Create mock TTF font.
	ttf := &fonts.TTFFont{
		FilePath:       "/fonts/TestFont.ttf",
		PostScriptName: "TestFont-Regular",
		UnitsPerEm:     1000,
		FontBBox:       [4]int16{0, -200, 1000, 800},
		Ascender:       800,
		Descender:      -200,
		LineGap:        0,
		ItalicAngle:    0,
		CapHeight:      700,
		XHeight:        500,
		WeightClass:    400,
		StemV:          80,
		Flags:          32,
		GlyphWidths:    make(map[uint16]uint16),
		CharToGlyph:    make(map[rune]uint16),
		FontData:       []byte("mock font data for testing"),
	}

	// Add some character mappings.
	ttf.CharToGlyph['H'] = 1
	ttf.CharToGlyph['e'] = 2
	ttf.CharToGlyph['l'] = 3
	ttf.CharToGlyph['o'] = 4
	ttf.GlyphWidths[1] = 700 // H
	ttf.GlyphWidths[2] = 500 // e
	ttf.GlyphWidths[3] = 300 // l
	ttf.GlyphWidths[4] = 500 // o

	// Create subset.
	subset := fonts.NewFontSubset(ttf)
	subset.UseString("Hello")

	// Create object number generator.
	nextObjNum := 10
	objNumGen := func() int {
		num := nextObjNum
		nextObjNum++
		return num
	}

	// Create writer.
	writer := NewTrueTypeFontWriter(ttf, subset, objNumGen)

	// Generate font objects.
	objects, refs, err := writer.WriteFont()
	if err != nil {
		t.Fatalf("WriteFont failed: %v", err)
	}

	// Verify we got 5 objects (Type 0 font + CIDFont + FontDescriptor + ToUnicode + FontFile2).
	if len(objects) != 5 {
		t.Errorf("Expected 5 objects, got %d", len(objects))
	}

	// Verify object numbers.
	if refs.FontObjNum != 10 {
		t.Errorf("FontObjNum = %d, want 10", refs.FontObjNum)
	}
	if refs.DescriptorObjNum != 11 {
		t.Errorf("DescriptorObjNum = %d, want 11", refs.DescriptorObjNum)
	}
	if refs.ToUnicodeObjNum != 12 {
		t.Errorf("ToUnicodeObjNum = %d, want 12", refs.ToUnicodeObjNum)
	}
	if refs.FontFileObjNum != 13 {
		t.Errorf("FontFileObjNum = %d, want 13", refs.FontFileObjNum)
	}

	// Find and verify Type 0 font dictionary.
	var fontDict *IndirectObject
	for _, obj := range objects {
		if obj.Number == refs.FontObjNum {
			fontDict = obj
			break
		}
	}

	if fontDict == nil {
		t.Fatal("Font dictionary object not found")
	}

	fontData := string(fontDict.Data)

	// Verify Type 0 font dictionary contents.
	if !strings.Contains(fontData, "/Type /Font") {
		t.Error("Missing /Type /Font in font dictionary")
	}
	if !strings.Contains(fontData, "/Subtype /Type0") {
		t.Error("Missing /Subtype /Type0")
	}
	if !strings.Contains(fontData, "/BaseFont /") {
		t.Error("Missing /BaseFont")
	}
	if !strings.Contains(fontData, "/DescendantFonts") {
		t.Error("Missing /DescendantFonts reference")
	}
	if !strings.Contains(fontData, "/ToUnicode") {
		t.Error("Missing /ToUnicode reference")
	}
	if !strings.Contains(fontData, "/Encoding /Identity-H") {
		t.Error("Missing /Encoding /Identity-H")
	}

	// Find CIDFont (descendant font).
	var cidFontDict *IndirectObject
	for _, obj := range objects {
		data := string(obj.Data)
		if strings.Contains(data, "/Subtype /CIDFontType2") {
			cidFontDict = obj
			break
		}
	}

	if cidFontDict == nil {
		t.Fatal("CIDFont dictionary not found")
	}

	cidFontData := string(cidFontDict.Data)
	if !strings.Contains(cidFontData, "/FontDescriptor") {
		t.Error("Missing /FontDescriptor in CIDFont")
	}
	if !strings.Contains(cidFontData, "/CIDSystemInfo") {
		t.Error("Missing /CIDSystemInfo in CIDFont")
	}
}

func TestTrueTypeFontWriter_FontDescriptor(t *testing.T) {
	ttf := &fonts.TTFFont{
		PostScriptName: "Arial-Bold",
		UnitsPerEm:     2048,
		FontBBox:       [4]int16{-200, -500, 1500, 1200},
		Ascender:       1800,
		Descender:      -400,
		ItalicAngle:    -12.5,
		CapHeight:      1400,
		StemV:          120,
		Flags:          32 | 64, // Nonsymbolic + Italic
		GlyphWidths:    make(map[uint16]uint16),
		CharToGlyph:    make(map[rune]uint16),
		FontData:       []byte("test"),
	}

	ttf.CharToGlyph['A'] = 1
	ttf.GlyphWidths[1] = 1400

	subset := fonts.NewFontSubset(ttf)
	subset.UseString("A")

	nextObjNum := 1
	writer := NewTrueTypeFontWriter(ttf, subset, func() int {
		num := nextObjNum
		nextObjNum++
		return num
	})

	objects, refs, err := writer.WriteFont()
	if err != nil {
		t.Fatalf("WriteFont failed: %v", err)
	}

	// Find FontDescriptor.
	var descriptor *IndirectObject
	for _, obj := range objects {
		if obj.Number == refs.DescriptorObjNum {
			descriptor = obj
			break
		}
	}

	if descriptor == nil {
		t.Fatal("FontDescriptor not found")
	}

	data := string(descriptor.Data)

	// Check descriptor contents.
	if !strings.Contains(data, "/Type /FontDescriptor") {
		t.Error("Missing /Type /FontDescriptor")
	}
	if !strings.Contains(data, "/FontName /") {
		t.Error("Missing /FontName")
	}
	if !strings.Contains(data, "/ItalicAngle -12.5") {
		t.Error("Missing or wrong /ItalicAngle")
	}
	if !strings.Contains(data, "/FontFile2") {
		t.Error("Missing /FontFile2 reference")
	}
}

func TestTrueTypeFontWriter_ToUnicode(t *testing.T) {
	ttf := &fonts.TTFFont{
		PostScriptName: "TestFont",
		UnitsPerEm:     1000,
		Ascender:       800,
		Descender:      -200,
		Flags:          32,
		GlyphWidths:    make(map[uint16]uint16),
		CharToGlyph:    make(map[rune]uint16),
		FontData:       []byte("test"),
	}

	// Add Cyrillic characters.
	ttf.CharToGlyph['П'] = 10
	ttf.CharToGlyph['р'] = 11
	ttf.CharToGlyph['и'] = 12
	ttf.CharToGlyph['в'] = 13
	ttf.CharToGlyph['е'] = 14
	ttf.CharToGlyph['т'] = 15
	ttf.GlyphWidths[10] = 700
	ttf.GlyphWidths[11] = 500
	ttf.GlyphWidths[12] = 500
	ttf.GlyphWidths[13] = 500
	ttf.GlyphWidths[14] = 500
	ttf.GlyphWidths[15] = 500

	subset := fonts.NewFontSubset(ttf)
	subset.UseString("Привет")

	nextObjNum := 1
	writer := NewTrueTypeFontWriter(ttf, subset, func() int {
		num := nextObjNum
		nextObjNum++
		return num
	})

	objects, refs, err := writer.WriteFont()
	if err != nil {
		t.Fatalf("WriteFont failed: %v", err)
	}

	// Find ToUnicode stream.
	var toUnicode *IndirectObject
	for _, obj := range objects {
		if obj.Number == refs.ToUnicodeObjNum {
			toUnicode = obj
			break
		}
	}

	if toUnicode == nil {
		t.Fatal("ToUnicode stream not found")
	}

	// ToUnicode should be a compressed stream.
	data := string(toUnicode.Data)
	if !strings.Contains(data, "/Filter /FlateDecode") {
		t.Error("ToUnicode stream should be compressed")
	}
	if !strings.Contains(data, "stream") {
		t.Error("Missing stream keyword")
	}
}
