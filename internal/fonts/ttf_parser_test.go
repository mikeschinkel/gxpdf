package fonts

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestParseFontDirectory tests parsing of font directory.
func TestParseFontDirectory(t *testing.T) {
	// Create minimal font directory for testing.
	var buf bytes.Buffer

	// Write sfnt version (TrueType = 0x00010000).
	_ = binary.Write(&buf, binary.BigEndian, uint32(0x00010000))

	// Write numTables = 2.
	_ = binary.Write(&buf, binary.BigEndian, uint16(2))

	// Write searchRange, entrySelector, rangeShift.
	_ = binary.Write(&buf, binary.BigEndian, uint16(32)) // searchRange.
	_ = binary.Write(&buf, binary.BigEndian, uint16(1))  // entrySelector.
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))  // rangeShift.

	// Write table entry 1: "head".
	buf.WriteString("head")
	_ = binary.Write(&buf, binary.BigEndian, uint32(0x12345678)) // checksum.
	_ = binary.Write(&buf, binary.BigEndian, uint32(100))        // offset.
	_ = binary.Write(&buf, binary.BigEndian, uint32(54))         // length.

	// Write table entry 2: "hhea".
	buf.WriteString("hhea")
	_ = binary.Write(&buf, binary.BigEndian, uint32(0x87654321)) // checksum.
	_ = binary.Write(&buf, binary.BigEndian, uint32(200))        // offset.
	_ = binary.Write(&buf, binary.BigEndian, uint32(36))         // length.

	// Parse font directory.
	font := &TTFFont{
		Tables:      make(map[string]*TTFTable),
		GlyphWidths: make(map[uint16]uint16),
		CharToGlyph: make(map[rune]uint16),
	}

	err := font.parseFontDirectory(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("parseFontDirectory failed: %v", err)
	}

	// Verify tables were parsed.
	if len(font.Tables) != 2 {
		t.Errorf("expected 2 tables, got %d", len(font.Tables))
	}

	// Verify "head" table.
	headTable, ok := font.Tables["head"]
	if !ok {
		t.Fatal("head table not found")
	}
	if headTable.Tag != "head" {
		t.Errorf("expected tag 'head', got %q", headTable.Tag)
	}
	if headTable.Offset != 100 {
		t.Errorf("expected offset 100, got %d", headTable.Offset)
	}
	if headTable.Length != 54 {
		t.Errorf("expected length 54, got %d", headTable.Length)
	}

	// Verify "hhea" table.
	hheaTable, ok := font.Tables["hhea"]
	if !ok {
		t.Fatal("hhea table not found")
	}
	if hheaTable.Tag != "hhea" {
		t.Errorf("expected tag 'hhea', got %q", hheaTable.Tag)
	}
}

// TestParseTableEntry tests parsing of a single table entry.
func TestParseTableEntry(t *testing.T) {
	var buf bytes.Buffer

	// Write table entry: "test".
	buf.WriteString("test")
	_ = binary.Write(&buf, binary.BigEndian, uint32(0xAABBCCDD)) // checksum.
	_ = binary.Write(&buf, binary.BigEndian, uint32(1000))       // offset.
	_ = binary.Write(&buf, binary.BigEndian, uint32(500))        // length.

	// Parse entry.
	font := &TTFFont{}
	entry, err := font.parseTableEntry(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("parseTableEntry failed: %v", err)
	}

	// Verify fields.
	if entry.Tag != "test" {
		t.Errorf("expected tag 'test', got %q", entry.Tag)
	}
	if entry.Checksum != 0xAABBCCDD {
		t.Errorf("expected checksum 0xAABBCCDD, got 0x%08X", entry.Checksum)
	}
	if entry.Offset != 1000 {
		t.Errorf("expected offset 1000, got %d", entry.Offset)
	}
	if entry.Length != 500 {
		t.Errorf("expected length 500, got %d", entry.Length)
	}
}

// TestLoadTable tests loading table data.
func TestLoadTable(t *testing.T) {
	// Create test data.
	data := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	}

	// Create table entry.
	table := &TTFTable{
		Tag:    "test",
		Offset: 4,
		Length: 8,
	}

	// Load table data.
	font := &TTFFont{}
	err := font.loadTable(data, table)
	if err != nil {
		t.Fatalf("loadTable failed: %v", err)
	}

	// Verify loaded data.
	expected := []byte{0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B}
	if !bytes.Equal(table.Data, expected) {
		t.Errorf("expected %v, got %v", expected, table.Data)
	}
}

// TestLoadTableOutOfBounds tests error handling for invalid offsets.
func TestLoadTableOutOfBounds(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}

	tests := []struct {
		name   string
		offset uint32
		length uint32
	}{
		{"offset too large", 100, 10},
		{"length too large", 0, 100},
		{"offset + length overflow", 2, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := &TTFTable{
				Tag:    "test",
				Offset: tt.offset,
				Length: tt.length,
			}

			font := &TTFFont{}
			err := font.loadTable(data, table)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// TestLoadTTFGlyphWidths tests that glyph widths are properly loaded from hmtx table.
// This is a regression test for the bug where numOfLongHorMetrics was read
// from the wrong offset in the hhea table.
func TestLoadTTFGlyphWidths(t *testing.T) {
	ttf, err := LoadTTF("C:/Windows/Fonts/arial.ttf")
	if err != nil {
		t.Skipf("test font not available: %v", err)
	}

	// GlyphWidths must be populated from hmtx table.
	if len(ttf.GlyphWidths) == 0 {
		t.Fatal("GlyphWidths should not be empty - hmtx table not properly parsed")
	}

	// Arial has thousands of glyphs.
	if len(ttf.GlyphWidths) < 100 {
		t.Errorf("expected more than 100 glyph widths, got %d", len(ttf.GlyphWidths))
	}

	// Glyph 0 (.notdef) should exist.
	if _, ok := ttf.GlyphWidths[0]; !ok {
		t.Error("glyph 0 (.notdef) should have a width entry")
	}

	// Check that widths are reasonable (not all zeros).
	nonZero := 0
	for _, w := range ttf.GlyphWidths {
		if w > 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Error("all glyph widths are zero - hmtx data corrupted")
	}

	t.Logf("GlyphWidths: %d total, %d non-zero", len(ttf.GlyphWidths), nonZero)
}

// TestLoadTTFCharToGlyph tests that character-to-glyph mapping is properly built.
func TestLoadTTFCharToGlyph(t *testing.T) {
	ttf, err := LoadTTF("C:/Windows/Fonts/arial.ttf")
	if err != nil {
		t.Skipf("test font not available: %v", err)
	}

	// CharToGlyph must be populated from cmap table.
	if len(ttf.CharToGlyph) == 0 {
		t.Fatal("CharToGlyph should not be empty - cmap table not properly parsed")
	}

	// Test basic ASCII characters.
	asciiChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for _, ch := range asciiChars {
		if _, ok := ttf.CharToGlyph[ch]; !ok {
			t.Errorf("ASCII character %q (U+%04X) not in CharToGlyph", ch, ch)
		}
	}

	// Test Cyrillic characters (Arial supports Cyrillic).
	cyrillicChars := "АБВГДЕЖЗИКЛМНОПРСТУФХЦЧШЩЫЭЮЯабвгдежзиклмнопрстуфхцчшщыэюя"
	cyrillicFound := 0
	for _, ch := range cyrillicChars {
		if _, ok := ttf.CharToGlyph[ch]; ok {
			cyrillicFound++
		}
	}
	if cyrillicFound == 0 {
		t.Error("no Cyrillic characters found - cmap format 4 parsing error")
	}

	t.Logf("CharToGlyph: %d entries, %d Cyrillic chars found", len(ttf.CharToGlyph), cyrillicFound)
}

// TestLoadTTFFontMetrics tests that font metrics are properly extracted.
func TestLoadTTFFontMetrics(t *testing.T) {
	ttf, err := LoadTTF("C:/Windows/Fonts/arial.ttf")
	if err != nil {
		t.Skipf("test font not available: %v", err)
	}

	// UnitsPerEm should be 2048 for Arial.
	if ttf.UnitsPerEm != 2048 {
		t.Errorf("expected UnitsPerEm=2048 for Arial, got %d", ttf.UnitsPerEm)
	}

	// Ascender should be positive.
	if ttf.Ascender <= 0 {
		t.Errorf("Ascender should be positive, got %d", ttf.Ascender)
	}

	// Descender should be negative.
	if ttf.Descender >= 0 {
		t.Errorf("Descender should be negative, got %d", ttf.Descender)
	}

	// FontBBox should have valid dimensions.
	if ttf.FontBBox[2] <= ttf.FontBBox[0] {
		t.Error("FontBBox xMax should be greater than xMin")
	}
	if ttf.FontBBox[3] <= ttf.FontBBox[1] {
		t.Error("FontBBox yMax should be greater than yMin")
	}

	// PostScriptName should be "ArialMT" for Arial.
	if ttf.PostScriptName == "" {
		t.Error("PostScriptName should not be empty")
	}

	t.Logf("PostScriptName: %s", ttf.PostScriptName)
	t.Logf("UnitsPerEm: %d, Ascender: %d, Descender: %d", ttf.UnitsPerEm, ttf.Ascender, ttf.Descender)
}

// TestBuildCharToGlyphMapping tests the cmap format 4 parsing logic.
func TestBuildCharToGlyphMapping(t *testing.T) {
	// Create a simple format4Arrays for testing.
	// Segment 1: map 'A'-'Z' to glyph IDs 65-90 (idDelta = 0)
	// Segment 2: end marker 0xFFFF
	arrays := &format4Arrays{
		endCode:       []uint16{0x5A, 0xFFFF}, // 'Z', end marker
		startCode:     []uint16{0x41, 0xFFFF}, // 'A', end marker
		idDelta:       []int16{0, 1},
		idRangeOffset: []uint16{0, 0},
		glyphIDArray:  nil,
	}

	font := &TTFFont{
		CharToGlyph: make(map[rune]uint16),
	}

	font.buildCharToGlyphMapping(2, arrays)

	// Check that A-Z are mapped correctly.
	for ch := 'A'; ch <= 'Z'; ch++ {
		glyph, ok := font.CharToGlyph[ch]
		if !ok {
			t.Errorf("character %q not in CharToGlyph", ch)
			continue
		}
		// With idDelta=0, glyphID = charCode
		if glyph != uint16(ch) {
			t.Errorf("character %q: expected glyph %d, got %d", ch, ch, glyph)
		}
	}
}

// TestBuildCharToGlyphMappingWithIdRangeOffset tests cmap format 4 with idRangeOffset.
func TestBuildCharToGlyphMappingWithIdRangeOffset(t *testing.T) {
	// Test case: segment with idRangeOffset (indirect glyph lookup).
	// This is the more complex case where glyphs are looked up from an array.
	segCount := uint16(2)

	// Segment 0: charCodes 0x30-0x32 ('0'-'2'), using glyphIDArray.
	// idRangeOffset points into glyphIDArray.
	// Formula: idx = idRangeOffset[i]/2 - (segCount-i) + (charCode - startCode[i])
	// For i=0, segCount=2: idx = offset/2 - 2 + (charCode - 0x30)
	// We want idx=0,1,2 for charCodes 0x30,0x31,0x32.
	// offset/2 - 2 + 0 = 0 => offset/2 = 2 => offset = 4.
	arrays := &format4Arrays{
		endCode:       []uint16{0x32, 0xFFFF},
		startCode:     []uint16{0x30, 0xFFFF},
		idDelta:       []int16{0, 1},
		idRangeOffset: []uint16{4, 0}, // 4 bytes = 2 uint16 positions
		glyphIDArray:  []uint16{100, 101, 102},
	}

	font := &TTFFont{
		CharToGlyph: make(map[rune]uint16),
	}

	font.buildCharToGlyphMapping(segCount, arrays)

	// Check that '0', '1', '2' map to glyphs 100, 101, 102.
	expected := map[rune]uint16{
		'0': 100,
		'1': 101,
		'2': 102,
	}

	for ch, expectedGlyph := range expected {
		glyph, ok := font.CharToGlyph[ch]
		if !ok {
			t.Errorf("character %q not in CharToGlyph", ch)
			continue
		}
		if glyph != expectedGlyph {
			t.Errorf("character %q: expected glyph %d, got %d", ch, expectedGlyph, glyph)
		}
	}
}
