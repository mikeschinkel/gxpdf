package fonts

import (
	"bytes"
	"fmt"
	"sort"
)

// GenerateToUnicodeCMap generates a ToUnicode CMap for text extraction.
//
// A ToUnicode CMap allows PDF viewers to extract correct Unicode text
// from documents using embedded fonts.
//
// The CMap maps character codes (as used in the PDF content stream)
// to Unicode code points.
//
// Reference: PDF 1.7 specification, Section 9.10 (ToUnicode CMaps).
func GenerateToUnicodeCMap(subset *FontSubset) ([]byte, error) {
	var buf bytes.Buffer

	// Write CMap header.
	if err := writeCMapHeader(&buf); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}

	// Write character mappings.
	if err := writeCharMappings(&buf, subset); err != nil {
		return nil, fmt.Errorf("write mappings: %w", err)
	}

	// Write CMap footer.
	if err := writeCMapFooter(&buf); err != nil {
		return nil, fmt.Errorf("write footer: %w", err)
	}

	return buf.Bytes(), nil
}

// writeCMapHeader writes the CMap header.
func writeCMapHeader(buf *bytes.Buffer) error {
	// Code space range is 2 bytes (0000-FFFF) to accommodate 16-bit glyph IDs.
	header := `/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo
<< /Registry (Adobe)
/Ordering (UCS)
/Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
`
	_, err := buf.WriteString(header)
	return err
}

// glyphMapping represents a mapping from glyph ID to Unicode code point.
type glyphMapping struct {
	glyphID uint16
	unicode rune
}

// writeCharMappings writes glyph ID to Unicode mappings.
//
// For TrueType fonts, the content stream uses glyph IDs as character codes.
// This CMap maps those glyph IDs back to Unicode code points for text extraction.
func writeCharMappings(buf *bytes.Buffer, subset *FontSubset) error {
	// Build glyph ID → Unicode mappings.
	var mappings []glyphMapping

	for ch := range subset.UsedChars {
		glyphID, ok := subset.BaseFont.CharToGlyph[ch]
		if !ok {
			// Character not in font - skip.
			continue
		}

		mappings = append(mappings, glyphMapping{
			glyphID: glyphID,
			unicode: ch,
		})
	}

	// Sort by glyph ID for consistent output.
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].glyphID < mappings[j].glyphID
	})

	// Write mappings in batches of 100 (PDF spec limit).
	const maxBatchSize = 100
	for i := 0; i < len(mappings); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(mappings) {
			end = len(mappings)
		}

		if err := writeMappingBatch(buf, mappings[i:end]); err != nil {
			return fmt.Errorf("write batch: %w", err)
		}
	}

	return nil
}

// writeMappingBatch writes a batch of glyph ID to Unicode mappings.
func writeMappingBatch(buf *bytes.Buffer, mappings []glyphMapping) error {
	// Write batch header.
	if _, err := fmt.Fprintf(buf, "%d beginbfchar\n", len(mappings)); err != nil {
		return err
	}

	// Write each mapping: <glyphID> <unicode>
	for _, m := range mappings {
		// Glyph ID as 2-byte hex (TrueType uses 16-bit glyph IDs).
		glyphCode := fmt.Sprintf("<%04X>", m.glyphID)

		// Unicode code point as 4-digit hex.
		unicode := fmt.Sprintf("<%04X>", m.unicode)

		// Write mapping line.
		if _, err := fmt.Fprintf(buf, "%s %s\n", glyphCode, unicode); err != nil {
			return err
		}
	}

	// Write batch footer.
	if _, err := buf.WriteString("endbfchar\n"); err != nil {
		return err
	}

	return nil
}

// writeCMapFooter writes the CMap footer.
func writeCMapFooter(buf *bytes.Buffer) error {
	footer := `endcmap
CMapName currentdict /CMap defineresource pop
end
end
`
	_, err := buf.WriteString(footer)
	return err
}
