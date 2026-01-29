package fonts

import (
	"strings"
	"testing"
)

// TestGenerateToUnicodeCMap tests ToUnicode CMap generation.
func TestGenerateToUnicodeCMap(t *testing.T) {
	// Create a mock TTFFont with CharToGlyph mapping.
	ttf := &TTFFont{
		CharToGlyph: map[rune]uint16{
			'A': 36,
			'B': 37,
			'C': 38,
			'а': 100, // Cyrillic 'а'
			'б': 101, // Cyrillic 'б'
		},
	}

	// Create a font subset.
	subset := NewFontSubset(ttf)
	subset.UseChar('A')
	subset.UseChar('B')
	subset.UseChar('а')

	// Generate ToUnicode CMap.
	cmap, err := GenerateToUnicodeCMap(subset)
	if err != nil {
		t.Fatalf("GenerateToUnicodeCMap failed: %v", err)
	}

	cmapStr := string(cmap)

	// Check CMap header.
	if !strings.Contains(cmapStr, "/CIDInit /ProcSet findresource begin") {
		t.Error("CMap should contain standard header")
	}

	// Check code space range (2-byte for glyph IDs).
	if !strings.Contains(cmapStr, "<0000> <FFFF>") {
		t.Error("CMap should have 2-byte code space range")
	}

	// Check that mappings are glyphID -> unicode (not unicode -> unicode).
	// 'A' = U+0041, glyph ID = 36 = 0x0024
	// Mapping should be: <0024> <0041>
	if !strings.Contains(cmapStr, "<0024> <0041>") {
		t.Error("CMap should map glyph ID 0x0024 to Unicode U+0041 ('A')")
	}

	// 'а' (Cyrillic) = U+0430, glyph ID = 100 = 0x0064
	// Mapping should be: <0064> <0430>
	if !strings.Contains(cmapStr, "<0064> <0430>") {
		t.Error("CMap should map glyph ID 0x0064 to Unicode U+0430 ('а')")
	}

	// Check CMap footer.
	if !strings.Contains(cmapStr, "endcmap") {
		t.Error("CMap should contain footer")
	}

	t.Logf("Generated CMap:\n%s", cmapStr)
}

// TestGenerateToUnicodeCMapEmpty tests CMap generation with empty subset.
func TestGenerateToUnicodeCMapEmpty(t *testing.T) {
	ttf := &TTFFont{
		CharToGlyph: make(map[rune]uint16),
	}
	subset := NewFontSubset(ttf)

	cmap, err := GenerateToUnicodeCMap(subset)
	if err != nil {
		t.Fatalf("GenerateToUnicodeCMap failed: %v", err)
	}

	// Should still produce valid CMap structure.
	cmapStr := string(cmap)
	if !strings.Contains(cmapStr, "begincodespacerange") {
		t.Error("empty CMap should still have code space range")
	}
}

// TestGenerateToUnicodeCMapLargeBatch tests CMap with >100 characters.
func TestGenerateToUnicodeCMapLargeBatch(t *testing.T) {
	ttf := &TTFFont{
		CharToGlyph: make(map[rune]uint16),
	}

	// Add 150 characters to test batching (max 100 per batch).
	subset := NewFontSubset(ttf)
	for i := 0; i < 150; i++ {
		ch := rune('A' + i)
		glyph := uint16(36 + i)
		ttf.CharToGlyph[ch] = glyph
		subset.UseChar(ch)
	}

	cmap, err := GenerateToUnicodeCMap(subset)
	if err != nil {
		t.Fatalf("GenerateToUnicodeCMap failed: %v", err)
	}

	cmapStr := string(cmap)

	// Should have at least 2 beginbfchar sections.
	count := strings.Count(cmapStr, "beginbfchar")
	if count < 2 {
		t.Errorf("expected at least 2 beginbfchar sections, got %d", count)
	}

	t.Logf("CMap has %d beginbfchar sections for %d characters", count, len(subset.UsedChars))
}

// TestGlyphMappingOrder tests that mappings are sorted by glyph ID.
func TestGlyphMappingOrder(t *testing.T) {
	ttf := &TTFFont{
		CharToGlyph: map[rune]uint16{
			'Z': 90,  // Higher glyph ID
			'A': 36,  // Lower glyph ID
			'M': 77,  // Middle glyph ID
		},
	}

	subset := NewFontSubset(ttf)
	subset.UseChar('Z')
	subset.UseChar('A')
	subset.UseChar('M')

	cmap, err := GenerateToUnicodeCMap(subset)
	if err != nil {
		t.Fatalf("GenerateToUnicodeCMap failed: %v", err)
	}

	cmapStr := string(cmap)

	// Find positions of each mapping.
	posA := strings.Index(cmapStr, "<0024>") // glyph ID 36
	posM := strings.Index(cmapStr, "<004D>") // glyph ID 77
	posZ := strings.Index(cmapStr, "<005A>") // glyph ID 90

	// Should be sorted by glyph ID: A < M < Z.
	if posA > posM || posM > posZ {
		t.Error("mappings should be sorted by glyph ID")
	}
}
