package writer

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/coregx/gxpdf/internal/fonts"
)

// EmbeddedFontRefs holds object numbers for an embedded TrueType font.
//
// These references are used to link font objects together in the PDF.
type EmbeddedFontRefs struct {
	FontObjNum       int // Font dictionary object number
	DescriptorObjNum int // FontDescriptor object number
	ToUnicodeObjNum  int // ToUnicode CMap object number
	FontFileObjNum   int // FontFile2 stream object number
}

// TrueTypeFontWriter generates PDF objects for TrueType/OpenType fonts.
//
// This writer creates all required objects for embedding a TrueType font
// as a Type 0 Composite Font (for full Unicode support):
//   - Type 0 Font dictionary
//   - CIDFontType2 descendant font
//   - FontDescriptor (font metrics)
//   - ToUnicode CMap (for text extraction)
//   - FontFile2 stream (embedded font data)
//
// Reference: PDF 1.7, Section 9.7 (Composite Fonts) and 9.8 (FontDescriptor).
type TrueTypeFontWriter struct {
	ttf        *fonts.TTFFont
	subset     *fonts.FontSubset
	objNumGen  func() int      // Function to generate next object number
	cidFontObj *IndirectObject // CIDFont object (set during createFontObject)
}

// NewTrueTypeFontWriter creates a new TrueType font writer.
//
// Parameters:
//   - ttf: Parsed TrueType font
//   - subset: Font subset with used characters
//   - objNumGen: Function that returns next available object number
func NewTrueTypeFontWriter(ttf *fonts.TTFFont, subset *fonts.FontSubset, objNumGen func() int) *TrueTypeFontWriter {
	return &TrueTypeFontWriter{
		ttf:       ttf,
		subset:    subset,
		objNumGen: objNumGen,
	}
}

// WriteFont generates all PDF objects for the embedded font.
//
// Returns:
//   - objects: List of IndirectObjects to add to the PDF
//   - refs: Object numbers for cross-referencing
//   - error: If font generation fails
func (w *TrueTypeFontWriter) WriteFont() ([]*IndirectObject, *EmbeddedFontRefs, error) {
	// Allocate object numbers.
	fontObjNum := w.objNumGen()
	descriptorObjNum := w.objNumGen()
	toUnicodeObjNum := w.objNumGen()
	fontFileObjNum := w.objNumGen()

	refs := &EmbeddedFontRefs{
		FontObjNum:       fontObjNum,
		DescriptorObjNum: descriptorObjNum,
		ToUnicodeObjNum:  toUnicodeObjNum,
		FontFileObjNum:   fontFileObjNum,
	}

	objects := make([]*IndirectObject, 0, 5)

	// 1. Create FontFile2 stream (compressed font data).
	fontFileObj, err := w.createFontFileObject(fontFileObjNum)
	if err != nil {
		return nil, nil, fmt.Errorf("create font file: %w", err)
	}
	objects = append(objects, fontFileObj)

	// 2. Create FontDescriptor object.
	descriptorObj, err := w.createFontDescriptorObject(descriptorObjNum, fontFileObjNum)
	if err != nil {
		return nil, nil, fmt.Errorf("create font descriptor: %w", err)
	}
	objects = append(objects, descriptorObj)

	// 3. Create ToUnicode CMap stream.
	toUnicodeObj, err := w.createToUnicodeObject(toUnicodeObjNum)
	if err != nil {
		return nil, nil, fmt.Errorf("create ToUnicode: %w", err)
	}
	objects = append(objects, toUnicodeObj)

	// 4. Create Font dictionary (Type 0 Composite Font).
	// This also creates the CIDFont object internally.
	fontObj, err := w.createFontObject(fontObjNum, descriptorObjNum, toUnicodeObjNum)
	if err != nil {
		return nil, nil, fmt.Errorf("create font dictionary: %w", err)
	}
	objects = append(objects, fontObj)

	// 5. Add CIDFont object (created by createFontObject).
	if w.cidFontObj != nil {
		objects = append(objects, w.cidFontObj)
	}

	return objects, refs, nil
}

// createFontFileObject creates the FontFile2 stream with compressed font data.
func (w *TrueTypeFontWriter) createFontFileObject(objNum int) (*IndirectObject, error) {
	// Get compressed font data from subset.
	compressedData := w.subset.SubsetData
	originalLength := len(w.ttf.FontData)

	// If not already compressed, compress it.
	if len(compressedData) == 0 || len(compressedData) >= originalLength {
		var err error
		compressedData, err = CompressStream(w.ttf.FontData, DefaultCompression)
		if err != nil {
			return nil, fmt.Errorf("compress font data: %w", err)
		}
	}

	// Create stream dictionary.
	var buf bytes.Buffer
	buf.WriteString("<<\n")
	buf.WriteString(fmt.Sprintf("/Length %d\n", len(compressedData)))
	buf.WriteString(fmt.Sprintf("/Length1 %d\n", originalLength))
	buf.WriteString("/Filter /FlateDecode\n")
	buf.WriteString(">>\n")
	buf.WriteString("stream\n")
	buf.Write(compressedData)
	buf.WriteString("\nendstream")

	return &IndirectObject{
		Number:     objNum,
		Generation: 0,
		Data:       buf.Bytes(),
	}, nil
}

// createFontDescriptorObject creates the FontDescriptor dictionary.
func (w *TrueTypeFontWriter) createFontDescriptorObject(objNum, fontFileObjNum int) (*IndirectObject, error) {
	// Generate FontDescriptor from TTF data.
	fd := fonts.GenerateFontDescriptor(w.ttf)
	if fd == nil {
		return nil, fmt.Errorf("failed to generate font descriptor")
	}

	// Generate subset font name.
	usedChars := make([]rune, 0, len(w.subset.UsedChars))
	for ch := range w.subset.UsedChars {
		usedChars = append(usedChars, ch)
	}
	subsetName := fonts.SubsetFontName(fd.FontName, usedChars)

	// Create descriptor dictionary.
	var buf bytes.Buffer
	buf.WriteString("<<\n")
	buf.WriteString("/Type /FontDescriptor\n")
	buf.WriteString(fmt.Sprintf("/FontName /%s\n", subsetName))
	buf.WriteString(fmt.Sprintf("/Flags %d\n", fd.Flags))
	buf.WriteString(fmt.Sprintf("/FontBBox [%d %d %d %d]\n",
		fd.FontBBox[0], fd.FontBBox[1], fd.FontBBox[2], fd.FontBBox[3]))
	buf.WriteString(fmt.Sprintf("/ItalicAngle %.1f\n", fd.ItalicAngle))
	buf.WriteString(fmt.Sprintf("/Ascent %d\n", fd.Ascent))
	buf.WriteString(fmt.Sprintf("/Descent %d\n", fd.Descent))
	buf.WriteString(fmt.Sprintf("/CapHeight %d\n", fd.CapHeight))
	buf.WriteString(fmt.Sprintf("/StemV %d\n", fd.StemV))
	buf.WriteString(fmt.Sprintf("/FontFile2 %d 0 R\n", fontFileObjNum))
	buf.WriteString(">>")

	return &IndirectObject{
		Number:     objNum,
		Generation: 0,
		Data:       buf.Bytes(),
	}, nil
}

// createToUnicodeObject creates the ToUnicode CMap stream.
func (w *TrueTypeFontWriter) createToUnicodeObject(objNum int) (*IndirectObject, error) {
	// Generate ToUnicode CMap.
	cmapData, err := fonts.GenerateToUnicodeCMap(w.subset)
	if err != nil {
		return nil, fmt.Errorf("generate ToUnicode CMap: %w", err)
	}

	// Compress CMap data.
	compressedData, err := CompressStream(cmapData, DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("compress ToUnicode: %w", err)
	}

	// Create stream.
	var buf bytes.Buffer
	buf.WriteString("<<\n")
	buf.WriteString(fmt.Sprintf("/Length %d\n", len(compressedData)))
	buf.WriteString("/Filter /FlateDecode\n")
	buf.WriteString(">>\n")
	buf.WriteString("stream\n")
	buf.Write(compressedData)
	buf.WriteString("\nendstream")

	return &IndirectObject{
		Number:     objNum,
		Generation: 0,
		Data:       buf.Bytes(),
	}, nil
}

// createFontObject creates the main Font dictionary (Type 0 Composite Font).
//
// For full Unicode support, we use Type 0 (Composite) font structure:
// - Type 0 font with Identity-H encoding
// - CIDFontType2 descendant font (TrueType-based CID font)
// - Identity CIDToGIDMap
//
// This allows encoding any glyph ID directly in the content stream.
func (w *TrueTypeFontWriter) createFontObject(objNum, descriptorObjNum, toUnicodeObjNum int) (*IndirectObject, error) {
	// Generate subset font name.
	fd := fonts.GenerateFontDescriptor(w.ttf)
	usedChars := make([]rune, 0, len(w.subset.UsedChars))
	for ch := range w.subset.UsedChars {
		usedChars = append(usedChars, ch)
	}
	subsetName := fonts.SubsetFontName(fd.FontName, usedChars)

	// Allocate object number for CIDFont (descendant font).
	cidFontObjNum := w.objNumGen()

	// Generate W (Widths) array for CIDFont.
	widthsArray := w.generateCIDWidthsArray()

	// Create CIDFont dictionary (descendant font).
	var cidBuf bytes.Buffer
	cidBuf.WriteString("<<\n")
	cidBuf.WriteString("/Type /Font\n")
	cidBuf.WriteString("/Subtype /CIDFontType2\n")
	cidBuf.WriteString(fmt.Sprintf("/BaseFont /%s\n", subsetName))
	cidBuf.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >>\n")
	cidBuf.WriteString(fmt.Sprintf("/FontDescriptor %d 0 R\n", descriptorObjNum))
	cidBuf.WriteString("/CIDToGIDMap /Identity\n")
	cidBuf.WriteString(fmt.Sprintf("/DW %d\n", w.getDefaultWidth()))
	if widthsArray != "" {
		cidBuf.WriteString(fmt.Sprintf("/W %s\n", widthsArray))
	}
	cidBuf.WriteString(">>")

	cidFontObj := &IndirectObject{
		Number:     cidFontObjNum,
		Generation: 0,
		Data:       cidBuf.Bytes(),
	}

	// Create Type 0 font dictionary (composite font).
	var buf bytes.Buffer
	buf.WriteString("<<\n")
	buf.WriteString("/Type /Font\n")
	buf.WriteString("/Subtype /Type0\n")
	buf.WriteString(fmt.Sprintf("/BaseFont /%s\n", subsetName))
	buf.WriteString("/Encoding /Identity-H\n")
	buf.WriteString(fmt.Sprintf("/DescendantFonts [%d 0 R]\n", cidFontObjNum))
	buf.WriteString(fmt.Sprintf("/ToUnicode %d 0 R\n", toUnicodeObjNum))
	buf.WriteString(">>")

	fontObj := &IndirectObject{
		Number:     objNum,
		Generation: 0,
		Data:       buf.Bytes(),
	}

	// Store CIDFont object for later addition.
	w.cidFontObj = cidFontObj

	return fontObj, nil
}

// getDefaultWidth returns the default glyph width in PDF units.
func (w *TrueTypeFontWriter) getDefaultWidth() int {
	// Use advance width of space character if available.
	if glyphID, ok := w.ttf.CharToGlyph[' ']; ok {
		if width, ok := w.ttf.GlyphWidths[glyphID]; ok {
			scale := 1000.0 / float64(w.ttf.UnitsPerEm)
			return int(float64(width) * scale)
		}
	}
	// Default to 1000 (full em width).
	return 1000
}

// generateCIDWidthsArray generates the /W (Widths) array for CIDFont.
//
// The /W array format allows sparse representation of glyph widths:
// [startGID [w1 w2 w3 ...]] or [startGID endGID width]
//
// We use the first format for compactness.
func (w *TrueTypeFontWriter) generateCIDWidthsArray() string {
	if len(w.subset.UsedChars) == 0 {
		return ""
	}

	// Collect all used glyph IDs with their widths.
	type glyphWidth struct {
		gid   uint16
		width int
	}

	glyphs := make([]glyphWidth, 0, len(w.subset.UsedChars))
	scale := 1000.0 / float64(w.ttf.UnitsPerEm)

	for ch := range w.subset.UsedChars {
		gid, ok := w.ttf.CharToGlyph[ch]
		if !ok {
			continue
		}
		width, ok := w.ttf.GlyphWidths[gid]
		if !ok {
			continue
		}
		scaledWidth := int(float64(width) * scale)
		glyphs = append(glyphs, glyphWidth{gid: gid, width: scaledWidth})
	}

	if len(glyphs) == 0 {
		return ""
	}

	// Sort by glyph ID.
	sort.Slice(glyphs, func(i, j int) bool {
		return glyphs[i].gid < glyphs[j].gid
	})

	// Generate W array.
	var buf bytes.Buffer
	buf.WriteString("[")

	i := 0
	for i < len(glyphs) {
		// Find consecutive glyph IDs.
		start := i
		for i < len(glyphs)-1 && glyphs[i+1].gid == glyphs[i].gid+1 {
			i++
		}

		// Write this range.
		buf.WriteString(fmt.Sprintf("%d [", glyphs[start].gid))
		for j := start; j <= i; j++ {
			if j > start {
				buf.WriteString(" ")
			}
			buf.WriteString(fmt.Sprintf("%d", glyphs[j].width))
		}
		buf.WriteString("] ")

		i++
	}

	buf.WriteString("]")
	return buf.String()
}
