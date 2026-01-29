package fonts

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// HeadTable represents the 'head' (font header) table.
//
// The head table contains global information about the font:
//   - Font version, creation date
//   - Units per em (scaling factor)
//   - Bounding box
//   - Index format flags
//
// Reference: TrueType specification, 'head' table.
type HeadTable struct {
	UnitsPerEm uint16 // Units per em square (typically 1000 or 2048)
	XMin       int16  // Minimum X coordinate
	YMin       int16  // Minimum Y coordinate
	XMax       int16  // Maximum X coordinate
	YMax       int16  // Maximum Y coordinate
}

// HheaTable represents the 'hhea' (horizontal header) table.
//
// The hhea table contains metrics for horizontal layout:
//   - Ascender, descender, line gap
//   - Number of horizontal metrics
//
// Reference: TrueType specification, 'hhea' table.
type HheaTable struct {
	Ascender            int16  // Typographic ascender
	Descender           int16  // Typographic descender
	LineGap             int16  // Typographic line gap
	NumOfLongHorMetrics uint16 // Number of hMetric entries in hmtx table
}

// HmtxTable represents the 'hmtx' (horizontal metrics) table.
//
// The hmtx table contains advance widths and left side bearings
// for all glyphs in the font.
//
// Reference: TrueType specification, 'hmtx' table.
type HmtxTable struct {
	Metrics []HMetric // Horizontal metrics for each glyph
}

// HMetric represents horizontal metrics for a single glyph.
type HMetric struct {
	AdvanceWidth    uint16 // Advance width in font units
	LeftSideBearing int16  // Left side bearing in font units
}

// CmapTable represents the 'cmap' (character to glyph mapping) table.
//
// The cmap table maps character codes to glyph indices.
// It may contain multiple subtables for different platforms/encodings.
//
// Reference: TrueType specification, 'cmap' table.
type CmapTable struct {
	Subtables []*CmapSubtable // Platform-specific subtables
}

// CmapSubtable represents a single cmap subtable.
type CmapSubtable struct {
	PlatformID uint16          // Platform ID (0=Unicode, 3=Windows)
	EncodingID uint16          // Encoding ID
	Format     uint16          // Format number (0, 4, 6, 12, etc.)
	Mapping    map[rune]uint16 // Character to glyph ID mapping
}

// parseRequiredTables parses all required font tables.
func (f *TTFFont) parseRequiredTables() error {
	// Parse head table (required).
	if err := f.parseHeadTable(); err != nil {
		return fmt.Errorf("parse head table: %w", err)
	}

	// Parse hhea table (required).
	if err := f.parseHheaTable(); err != nil {
		return fmt.Errorf("parse hhea table: %w", err)
	}

	// Parse hmtx table (required).
	if err := f.parseHmtxTable(); err != nil {
		return fmt.Errorf("parse hmtx table: %w", err)
	}

	// Parse cmap table (required).
	if err := f.parseCmapTable(); err != nil {
		return fmt.Errorf("parse cmap table: %w", err)
	}

	// Parse optional tables for PDF embedding.
	// These tables provide additional metrics for FontDescriptor.

	// Parse post table (optional but recommended).
	if _, ok := f.Tables["post"]; ok {
		if err := f.parsePostTable(); err != nil {
			// Non-fatal: use defaults.
			f.ItalicAngle = 0
		}
	}

	// Parse OS/2 table (optional but recommended).
	if _, ok := f.Tables["OS/2"]; ok {
		if err := f.parseOS2Table(); err != nil {
			// Non-fatal: use defaults.
			f.CapHeight = f.Ascender
		}
	}

	// Parse name table for PostScript name (optional).
	if _, ok := f.Tables["name"]; ok {
		_ = f.parseNameTable() // Best effort.
	}

	// Calculate derived values.
	f.calculateDerivedMetrics()

	return nil
}

// parseHeadTable parses the 'head' table.
func (f *TTFFont) parseHeadTable() error {
	table, ok := f.Tables["head"]
	if !ok {
		return fmt.Errorf("head table not found")
	}

	r := bytes.NewReader(table.Data)

	// Skip version (4 bytes) and fontRevision (4 bytes).
	if err := skipBytes(r, 8); err != nil {
		return err
	}

	// Skip checksumAdjustment (4 bytes) and magicNumber (4 bytes).
	if err := skipBytes(r, 8); err != nil {
		return err
	}

	// Skip flags (2 bytes).
	if err := skipBytes(r, 2); err != nil {
		return err
	}

	// Read unitsPerEm.
	if err := binary.Read(r, binary.BigEndian, &f.UnitsPerEm); err != nil {
		return fmt.Errorf("read unitsPerEm: %w", err)
	}

	// Skip created and modified timestamps (8 bytes each = 16 bytes).
	if err := skipBytes(r, 16); err != nil {
		return err
	}

	// Read font bounding box: xMin, yMin, xMax, yMax (8 bytes total).
	if err := binary.Read(r, binary.BigEndian, &f.FontBBox[0]); err != nil {
		return fmt.Errorf("read xMin: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.FontBBox[1]); err != nil {
		return fmt.Errorf("read yMin: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.FontBBox[2]); err != nil {
		return fmt.Errorf("read xMax: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.FontBBox[3]); err != nil {
		return fmt.Errorf("read yMax: %w", err)
	}

	return nil
}

// parseHheaTable parses the 'hhea' table.
func (f *TTFFont) parseHheaTable() error {
	table, ok := f.Tables["hhea"]
	if !ok {
		return fmt.Errorf("hhea table not found")
	}

	r := bytes.NewReader(table.Data)

	// Skip version (4 bytes).
	if err := skipBytes(r, 4); err != nil {
		return err
	}

	// Read ascender, descender, lineGap (6 bytes).
	if err := binary.Read(r, binary.BigEndian, &f.Ascender); err != nil {
		return fmt.Errorf("read ascender: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.Descender); err != nil {
		return fmt.Errorf("read descender: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.LineGap); err != nil {
		return fmt.Errorf("read lineGap: %w", err)
	}

	// Skip other fields until numOfLongHorMetrics (24 bytes).
	// hhea table structure after lineGap:
	//   advanceWidthMax (2), minLeftSideBearing (2), minRightSideBearing (2),
	//   xMaxExtent (2), caretSlopeRise (2), caretSlopeRun (2), caretOffset (2),
	//   reserved1-4 (8), metricDataFormat (2) = 24 bytes total.
	if err := skipBytes(r, 24); err != nil {
		return err
	}

	// Read numOfLongHorMetrics (at offset 34).
	// Note: This value is already in table.Data at offset 34, we just need to
	// verify the read completes successfully. The hmtx parser will read it
	// directly from the table data.
	var numHMetrics uint16
	if err := binary.Read(r, binary.BigEndian, &numHMetrics); err != nil {
		return fmt.Errorf("read numOfLongHorMetrics: %w", err)
	}

	return nil
}

// parseHmtxTable parses the 'hmtx' table.
func (f *TTFFont) parseHmtxTable() error {
	hmtxTable, ok := f.Tables["hmtx"]
	if !ok {
		return fmt.Errorf("hmtx table not found")
	}

	hheaTable, ok := f.Tables["hhea"]
	if !ok {
		return fmt.Errorf("hhea table required for hmtx parsing")
	}

	// Get numOfLongHorMetrics from hhea.
	numHMetrics := binary.BigEndian.Uint16(hheaTable.Data[34:])

	r := bytes.NewReader(hmtxTable.Data)

	// Read long horizontal metrics (4 bytes each: advanceWidth + lsb).
	for i := uint16(0); i < numHMetrics; i++ {
		var advanceWidth uint16
		if err := binary.Read(r, binary.BigEndian, &advanceWidth); err != nil {
			return fmt.Errorf("read advanceWidth: %w", err)
		}

		// Skip left side bearing (2 bytes).
		if err := skipBytes(r, 2); err != nil {
			return err
		}

		f.GlyphWidths[i] = advanceWidth
	}

	return nil
}

// parseCmapTable parses the 'cmap' table.
func (f *TTFFont) parseCmapTable() error {
	table, ok := f.Tables["cmap"]
	if !ok {
		return fmt.Errorf("cmap table not found")
	}

	// Read cmap header.
	numTables, err := f.readCmapHeader(table.Data)
	if err != nil {
		return fmt.Errorf("read cmap header: %w", err)
	}

	// Find best subtable offset.
	bestOffset, err := f.findBestCmapSubtable(table.Data, numTables)
	if err != nil {
		return fmt.Errorf("find best subtable: %w", err)
	}

	// Parse the selected subtable.
	return f.parseCmapSubtable(table.Data, bestOffset)
}

// readCmapHeader reads the cmap table header.
func (f *TTFFont) readCmapHeader(data []byte) (uint16, error) {
	r := bytes.NewReader(data)

	// Read version.
	var version uint16
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return 0, fmt.Errorf("read version: %w", err)
	}

	// Read numTables.
	var numTables uint16
	if err := binary.Read(r, binary.BigEndian, &numTables); err != nil {
		return 0, fmt.Errorf("read numTables: %w", err)
	}

	return numTables, nil
}

// findBestCmapSubtable finds the best cmap subtable offset.
func (f *TTFFont) findBestCmapSubtable(data []byte, numTables uint16) (uint32, error) {
	r := bytes.NewReader(data[4:]) // Skip version and numTables.

	for i := uint16(0); i < numTables; i++ {
		var platformID, encodingID uint16
		var offset uint32

		if err := binary.Read(r, binary.BigEndian, &platformID); err != nil {
			return 0, fmt.Errorf("read platformID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &encodingID); err != nil {
			return 0, fmt.Errorf("read encodingID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &offset); err != nil {
			return 0, fmt.Errorf("read offset: %w", err)
		}

		// Prefer Windows Unicode BMP (platformID=3, encodingID=1).
		if platformID == 3 && encodingID == 1 {
			return offset, nil
		}
	}

	return 0, fmt.Errorf("no suitable cmap subtable found")
}

// parseCmapSubtable parses a cmap subtable (format 4 or 12).
func (f *TTFFont) parseCmapSubtable(data []byte, offset uint32) error {
	r := bytes.NewReader(data[offset:])

	var format uint16
	if err := binary.Read(r, binary.BigEndian, &format); err != nil {
		return fmt.Errorf("read format: %w", err)
	}

	switch format {
	case 4:
		return f.parseCmapFormat4(data, offset)
	case 12:
		return f.parseCmapFormat12(data, offset)
	default:
		return fmt.Errorf("unsupported cmap format: %d", format)
	}
}

// parseCmapFormat4 parses cmap format 4 (segment mapping).
func (f *TTFFont) parseCmapFormat4(data []byte, offset uint32) error {
	// Read format 4 header.
	segCount, err := f.readFormat4Header(data, offset)
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	// Read segment arrays.
	arrays, err := f.readFormat4Segments(data, offset, segCount)
	if err != nil {
		return fmt.Errorf("read segments: %w", err)
	}

	// Build character to glyph mapping.
	f.buildCharToGlyphMapping(segCount, arrays)

	return nil
}

// readFormat4Header reads the cmap format 4 header.
func (f *TTFFont) readFormat4Header(data []byte, offset uint32) (uint16, error) {
	r := bytes.NewReader(data[offset:])

	// Skip format (2) + length (2) + language (2) = 6 bytes.
	if err := skipBytes(r, 6); err != nil {
		return 0, err
	}

	// Read segCountX2.
	var segCountX2 uint16
	if err := binary.Read(r, binary.BigEndian, &segCountX2); err != nil {
		return 0, fmt.Errorf("read segCountX2: %w", err)
	}

	return segCountX2 / 2, nil
}

// format4Arrays holds the segment arrays from format 4.
type format4Arrays struct {
	endCode       []uint16
	startCode     []uint16
	idDelta       []int16
	idRangeOffset []uint16
	glyphIDArray  []uint16
}

// readFormat4Segments reads the segment arrays from format 4.
func (f *TTFFont) readFormat4Segments(
	data []byte,
	offset uint32,
	segCount uint16,
) (*format4Arrays, error) {
	r := bytes.NewReader(data[offset+14:]) // Skip to endCode array.
	arrays := &format4Arrays{}

	// Read endCode array.
	arrays.endCode = make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &arrays.endCode[i]); err != nil {
			return nil, fmt.Errorf("read endCode: %w", err)
		}
	}

	// Skip reservedPad (2 bytes).
	if err := skipBytes(r, 2); err != nil {
		return nil, err
	}

	// Read startCode array.
	arrays.startCode = make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &arrays.startCode[i]); err != nil {
			return nil, fmt.Errorf("read startCode: %w", err)
		}
	}

	// Read idDelta array.
	arrays.idDelta = make([]int16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &arrays.idDelta[i]); err != nil {
			return nil, fmt.Errorf("read idDelta: %w", err)
		}
	}

	// Read idRangeOffset array.
	arrays.idRangeOffset = make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &arrays.idRangeOffset[i]); err != nil {
			return nil, fmt.Errorf("read idRangeOffset: %w", err)
		}
	}

	// Read remaining bytes as glyphIDArray.
	// Calculate remaining bytes.
	remaining := data[offset:]
	headerAndArraysLen := 14 + (segCount * 8) + 2 // header + 4 arrays + pad
	if int(headerAndArraysLen) < len(remaining) {
		glyphDataLen := (len(remaining) - int(headerAndArraysLen)) / 2
		arrays.glyphIDArray = make([]uint16, glyphDataLen)
		for i := 0; i < glyphDataLen; i++ {
			if err := binary.Read(r, binary.BigEndian, &arrays.glyphIDArray[i]); err != nil {
				break // End of data
			}
		}
	}

	return arrays, nil
}

// buildCharToGlyphMapping builds the character to glyph mapping.
func (f *TTFFont) buildCharToGlyphMapping(segCount uint16, arrays *format4Arrays) {
	for i := uint16(0); i < segCount; i++ {
		for charCode := arrays.startCode[i]; charCode <= arrays.endCode[i]; charCode++ {
			if charCode == 0xFFFF {
				break
			}

			var glyphID uint16
			if arrays.idRangeOffset[i] == 0 {
				// Simple case: glyph = charCode + idDelta
				//nolint:gosec // Character code is uint16, fits in int32.
				glyphID = uint16((int32(charCode) + int32(arrays.idDelta[i])) & 0xFFFF)
			} else {
				// Complex case: look up from glyph ID array.
				// idRangeOffset[i] is the byte offset from &idRangeOffset[i] to the glyph index.
				// Formula from TrueType spec:
				// glyphIndex = *( &idRangeOffset[i] + idRangeOffset[i]/2 + (charCode - startCode[i]) )
				//
				// Since idRangeOffset[i] is at index i in the array, and glyphIDArray starts
				// right after idRangeOffset array, the index into glyphIDArray is:
				// idx = idRangeOffset[i]/2 - (segCount - i) + (charCode - startCode[i])
				//nolint:gosec // Safe integer conversion within bounds.
				idx := int(arrays.idRangeOffset[i]/2) - int(segCount-i) + int(charCode-arrays.startCode[i])
				if idx >= 0 && idx < len(arrays.glyphIDArray) {
					glyphID = arrays.glyphIDArray[idx]
					if glyphID != 0 {
						//nolint:gosec // Glyph ID is uint16, fits in int32.
						glyphID = uint16((int32(glyphID) + int32(arrays.idDelta[i])) & 0xFFFF)
					}
				}
			}

			if glyphID != 0 {
				f.CharToGlyph[rune(charCode)] = glyphID
			}
		}
	}
}

// parseCmapFormat12 parses cmap format 12 (segmented coverage).
func (f *TTFFont) parseCmapFormat12(_ []byte, _ uint32) error {
	// Format 12 is more complex, but less common for basic fonts.
	// For MVP, we'll focus on format 4 support.
	return fmt.Errorf("cmap format 12 not yet implemented")
}

// skipBytes skips n bytes in the reader.
func skipBytes(r *bytes.Reader, n int64) error {
	_, err := r.Seek(n, 1) // Seek relative to current position.
	if err != nil {
		return fmt.Errorf("skip %d bytes: %w", n, err)
	}
	return nil
}

// parsePostTable parses the 'post' (PostScript) table.
//
// The post table contains:
//   - ItalicAngle: Angle of italic text
//   - UnderlinePosition: Position of underline
//   - UnderlineThickness: Thickness of underline
//   - IsFixedPitch: Whether font is monospaced
//
// Reference: TrueType specification, 'post' table.
func (f *TTFFont) parsePostTable() error {
	table, ok := f.Tables["post"]
	if !ok {
		return fmt.Errorf("post table not found")
	}

	if len(table.Data) < 32 {
		return fmt.Errorf("post table too short: %d bytes", len(table.Data))
	}

	r := bytes.NewReader(table.Data)

	// Skip version (4 bytes).
	if err := skipBytes(r, 4); err != nil {
		return err
	}

	// Read italicAngle as Fixed (16.16 format).
	var italicAngleFixed int32
	if err := binary.Read(r, binary.BigEndian, &italicAngleFixed); err != nil {
		return fmt.Errorf("read italicAngle: %w", err)
	}
	// Convert Fixed 16.16 to float64.
	f.ItalicAngle = float64(italicAngleFixed) / 65536.0

	// Read underlinePosition (FWord = int16).
	if err := binary.Read(r, binary.BigEndian, &f.UnderlinePosition); err != nil {
		return fmt.Errorf("read underlinePosition: %w", err)
	}

	// Read underlineThickness (FWord = int16).
	if err := binary.Read(r, binary.BigEndian, &f.UnderlineThickness); err != nil {
		return fmt.Errorf("read underlineThickness: %w", err)
	}

	// Read isFixedPitch (uint32).
	var isFixedPitch uint32
	if err := binary.Read(r, binary.BigEndian, &isFixedPitch); err != nil {
		return fmt.Errorf("read isFixedPitch: %w", err)
	}
	f.IsFixedPitch = isFixedPitch != 0

	return nil
}

// parseOS2Table parses the 'OS/2' (OS/2 and Windows metrics) table.
//
// The OS/2 table contains:
//   - WeightClass: Font weight (100-900)
//   - WidthClass: Font width (1-9)
//   - FSType: Embedding licensing rights
//   - CapHeight: Height of capital letters
//   - XHeight: Height of lowercase x
//
// Reference: TrueType specification, 'OS/2' table.
func (f *TTFFont) parseOS2Table() error {
	table, ok := f.Tables["OS/2"]
	if !ok {
		return fmt.Errorf("OS/2 table not found")
	}

	if len(table.Data) < 78 {
		return fmt.Errorf("OS/2 table too short: %d bytes", len(table.Data))
	}

	r := bytes.NewReader(table.Data)

	// Read version (2 bytes).
	var version uint16
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return fmt.Errorf("read version: %w", err)
	}

	// Skip xAvgCharWidth (2 bytes).
	if err := skipBytes(r, 2); err != nil {
		return err
	}

	// Read usWeightClass (2 bytes).
	if err := binary.Read(r, binary.BigEndian, &f.WeightClass); err != nil {
		return fmt.Errorf("read usWeightClass: %w", err)
	}

	// Read usWidthClass (2 bytes).
	if err := binary.Read(r, binary.BigEndian, &f.WidthClass); err != nil {
		return fmt.Errorf("read usWidthClass: %w", err)
	}

	// Read fsType (2 bytes).
	if err := binary.Read(r, binary.BigEndian, &f.FSType); err != nil {
		return fmt.Errorf("read fsType: %w", err)
	}

	// Skip to sTypoAscender (offset 68).
	// Current position is 12, need to skip to 68.
	if err := skipBytes(r, 56); err != nil {
		return err
	}

	// Read sTypoAscender (2 bytes).
	if err := binary.Read(r, binary.BigEndian, &f.TypoAscender); err != nil {
		return fmt.Errorf("read sTypoAscender: %w", err)
	}

	// Read sTypoDescender (2 bytes).
	if err := binary.Read(r, binary.BigEndian, &f.TypoDescender); err != nil {
		return fmt.Errorf("read sTypoDescender: %w", err)
	}

	// Skip sTypoLineGap (2 bytes).
	if err := skipBytes(r, 2); err != nil {
		return err
	}

	// Skip usWinAscent, usWinDescent (4 bytes).
	if err := skipBytes(r, 4); err != nil {
		return err
	}

	// For version >= 2, read sxHeight and sCapHeight.
	if version >= 2 && len(table.Data) >= 96 {
		// Skip ulCodePageRange1, ulCodePageRange2 (8 bytes).
		if err := skipBytes(r, 8); err != nil {
			return err
		}

		// Read sxHeight (2 bytes).
		if err := binary.Read(r, binary.BigEndian, &f.XHeight); err != nil {
			return fmt.Errorf("read sxHeight: %w", err)
		}

		// Read sCapHeight (2 bytes).
		if err := binary.Read(r, binary.BigEndian, &f.CapHeight); err != nil {
			return fmt.Errorf("read sCapHeight: %w", err)
		}
	} else {
		// Estimate CapHeight as 70% of Ascender.
		f.CapHeight = int16(float64(f.Ascender) * 0.7)
		f.XHeight = int16(float64(f.Ascender) * 0.5)
	}

	return nil
}

// parseNameTable parses the 'name' table to extract PostScript name.
//
// Reference: TrueType specification, 'name' table.
func (f *TTFFont) parseNameTable() error {
	table, ok := f.Tables["name"]
	if !ok {
		return fmt.Errorf("name table not found")
	}

	if len(table.Data) < 6 {
		return fmt.Errorf("name table too short")
	}

	r := bytes.NewReader(table.Data)

	// Read format (2 bytes).
	var format uint16
	if err := binary.Read(r, binary.BigEndian, &format); err != nil {
		return fmt.Errorf("read format: %w", err)
	}

	// Read count (2 bytes).
	var count uint16
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return fmt.Errorf("read count: %w", err)
	}

	// Read stringOffset (2 bytes).
	var stringOffset uint16
	if err := binary.Read(r, binary.BigEndian, &stringOffset); err != nil {
		return fmt.Errorf("read stringOffset: %w", err)
	}

	// Search for PostScript name (nameID = 6).
	for i := uint16(0); i < count; i++ {
		var platformID, encodingID, languageID, nameID, length, offset uint16

		if err := binary.Read(r, binary.BigEndian, &platformID); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &encodingID); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &languageID); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &nameID); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &offset); err != nil {
			return err
		}

		// PostScript name has nameID = 6.
		if nameID == 6 {
			strStart := uint32(stringOffset) + uint32(offset)
			strEnd := strStart + uint32(length)
			if strEnd <= uint32(len(table.Data)) {
				nameData := table.Data[strStart:strEnd]

				// Platform 3 (Windows) uses UTF-16BE.
				if platformID == 3 {
					f.PostScriptName = decodeUTF16BE(nameData)
				} else {
					// Platform 1 (Mac) uses ASCII/MacRoman.
					f.PostScriptName = string(nameData)
				}
				return nil
			}
		}
	}

	// If PostScript name not found, use filename.
	return nil
}

// decodeUTF16BE decodes UTF-16 Big Endian bytes to string.
func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		return ""
	}

	runes := make([]rune, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		r := rune(binary.BigEndian.Uint16(data[i:]))
		runes = append(runes, r)
	}
	return string(runes)
}

// calculateDerivedMetrics calculates derived font metrics.
//
// This includes:
//   - StemV: Estimated vertical stem width
//   - Flags: PDF font flags bitmap
func (f *TTFFont) calculateDerivedMetrics() {
	// Estimate StemV from weight class.
	// Light fonts (100-300): 50-70
	// Normal fonts (400): 80
	// Medium fonts (500): 90
	// Bold fonts (600-700): 100-120
	// Black fonts (800-900): 130-150
	switch {
	case f.WeightClass <= 300:
		f.StemV = 50 + int16(f.WeightClass/10)
	case f.WeightClass <= 500:
		f.StemV = 80 + int16((f.WeightClass-400)/5)
	case f.WeightClass <= 700:
		f.StemV = 100 + int16((f.WeightClass-500)/5)
	default:
		f.StemV = 130 + int16((f.WeightClass-700)/10)
	}

	// Default StemV if weight class is 0.
	if f.StemV == 0 || f.WeightClass == 0 {
		f.StemV = 80 // Normal weight default.
	}

	// Calculate PDF font flags.
	// Bit 1 (1): FixedPitch
	// Bit 2 (2): Serif (not easily detectable, skip)
	// Bit 3 (4): Symbolic (skip for now)
	// Bit 4 (8): Script (skip for now)
	// Bit 6 (32): Nonsymbolic (standard Latin font)
	// Bit 7 (64): Italic
	// Bit 17 (65536): AllCap (skip)
	// Bit 18 (131072): SmallCap (skip)
	// Bit 19 (262144): ForceBold (skip)

	f.Flags = 32 // Nonsymbolic by default.

	if f.IsFixedPitch {
		f.Flags |= 1
	}

	if f.ItalicAngle != 0 {
		f.Flags |= 64
	}
}
