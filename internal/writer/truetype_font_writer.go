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
// This writer creates all required objects for embedding a TrueType font:
//   - Font dictionary (/Type /Font /Subtype /TrueType)
//   - FontDescriptor (font metrics)
//   - ToUnicode CMap (for text extraction)
//   - FontFile2 stream (embedded font data)
//
// Reference: PDF 1.7, Section 9.6 (Simple Fonts) and 9.8 (FontDescriptor).
type TrueTypeFontWriter struct {
	ttf       *fonts.TTFFont
	subset    *fonts.FontSubset
	objNumGen func() int // Function to generate next object number
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

	objects := make([]*IndirectObject, 0, 4)

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

	// 4. Create Font dictionary.
	fontObj, err := w.createFontObject(fontObjNum, descriptorObjNum, toUnicodeObjNum)
	if err != nil {
		return nil, nil, fmt.Errorf("create font dictionary: %w", err)
	}
	objects = append(objects, fontObj)

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

// createFontObject creates the main Font dictionary.
func (w *TrueTypeFontWriter) createFontObject(objNum, descriptorObjNum, toUnicodeObjNum int) (*IndirectObject, error) {
	// Generate FontDescriptor for name.
	fd := fonts.GenerateFontDescriptor(w.ttf)
	usedChars := make([]rune, 0, len(w.subset.UsedChars))
	for ch := range w.subset.UsedChars {
		usedChars = append(usedChars, ch)
	}
	subsetName := fonts.SubsetFontName(fd.FontName, usedChars)

	// Calculate FirstChar and LastChar.
	firstChar, lastChar := w.getCharRange()

	// Generate Widths array.
	widths := w.generateWidthsArray(firstChar, lastChar)

	// Create font dictionary.
	var buf bytes.Buffer
	buf.WriteString("<<\n")
	buf.WriteString("/Type /Font\n")
	buf.WriteString("/Subtype /TrueType\n")
	buf.WriteString(fmt.Sprintf("/BaseFont /%s\n", subsetName))
	buf.WriteString(fmt.Sprintf("/FirstChar %d\n", firstChar))
	buf.WriteString(fmt.Sprintf("/LastChar %d\n", lastChar))
	buf.WriteString(fmt.Sprintf("/Widths %s\n", widths))
	buf.WriteString(fmt.Sprintf("/FontDescriptor %d 0 R\n", descriptorObjNum))
	buf.WriteString(fmt.Sprintf("/ToUnicode %d 0 R\n", toUnicodeObjNum))
	buf.WriteString("/Encoding /WinAnsiEncoding\n") // For basic Latin; Unicode uses ToUnicode
	buf.WriteString(">>")

	return &IndirectObject{
		Number:     objNum,
		Generation: 0,
		Data:       buf.Bytes(),
	}, nil
}

// getCharRange returns the FirstChar and LastChar values.
func (w *TrueTypeFontWriter) getCharRange() (int, int) {
	if len(w.subset.UsedChars) == 0 {
		return 0, 0
	}

	// Collect character codes.
	chars := make([]int, 0, len(w.subset.UsedChars))
	for ch := range w.subset.UsedChars {
		// For simple TrueType fonts, use character code directly.
		// Limit to single-byte range for WinAnsiEncoding compatibility.
		if ch <= 255 {
			chars = append(chars, int(ch))
		}
	}

	if len(chars) == 0 {
		return 32, 126 // Default ASCII range
	}

	sort.Ints(chars)
	return chars[0], chars[len(chars)-1]
}

// generateWidthsArray generates the /Widths array for the font.
func (w *TrueTypeFontWriter) generateWidthsArray(firstChar, lastChar int) string {
	var buf bytes.Buffer
	buf.WriteString("[")

	// Scale factor for PDF units (1000 per em).
	scale := 1000.0 / float64(w.ttf.UnitsPerEm)

	for i := firstChar; i <= lastChar; i++ {
		if i > firstChar {
			buf.WriteString(" ")
		}

		// Get glyph ID for character.
		glyphID, ok := w.ttf.CharToGlyph[rune(i)]
		if !ok {
			// Character not in font, use 0 width.
			buf.WriteString("0")
			continue
		}

		// Get advance width for glyph.
		width, ok := w.ttf.GlyphWidths[glyphID]
		if !ok {
			buf.WriteString("0")
			continue
		}

		// Scale to PDF units.
		scaledWidth := int(float64(width) * scale)
		buf.WriteString(fmt.Sprintf("%d", scaledWidth))
	}

	buf.WriteString("]")
	return buf.String()
}
