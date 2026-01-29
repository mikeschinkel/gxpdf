// Package writer provides PDF writing infrastructure for generating PDF files.
package writer

import (
	"bytes"
	"fmt"
	"sort"
)

// ResourceDictionary manages PDF page resources (fonts, images, graphics states, etc.).
//
// Resources are referenced in content streams by name (e.g., /F1 for fonts, /Im1 for images).
// This struct tracks resource names and their corresponding PDF object numbers.
//
// PDF Dictionary Format:
//
//	/Resources <<
//	  /Font << /F1 5 0 R /F2 6 0 R >>
//	  /XObject << /Im1 7 0 R >>
//	  /ExtGState << /GS1 8 0 R >>
//	  /ProcSet [/PDF /Text /ImageB /ImageC /ImageI]
//	>>
//
// Thread Safety: Not thread-safe. Caller must synchronize if needed.
type ResourceDictionary struct {
	fonts      map[string]int    // Font resource name -> object number (e.g., "F1" -> 5)
	fontIDs    map[string]string // Font ID -> resource name (e.g., "custom:font_1" -> "F1")
	xobjects   map[string]int    // XObject resource name -> object number (e.g., "Im1" -> 10)
	extgstates map[string]int    // ExtGState resource name -> object number (e.g., "GS1" -> 15)
}

// NewResourceDictionary creates a new empty resource dictionary.
func NewResourceDictionary() *ResourceDictionary {
	return &ResourceDictionary{
		fonts:      make(map[string]int),
		fontIDs:    make(map[string]string),
		xobjects:   make(map[string]int),
		extgstates: make(map[string]int),
	}
}

// AddFont adds a font resource and returns its resource name.
//
// Fonts are named sequentially: F1, F2, F3, etc.
//
// Parameters:
//   - objNum: PDF object number of the font dictionary
//
// Returns:
//   - Resource name (e.g., "F1")
//
// Example:
//
//	rd := NewResourceDictionary()
//	name := rd.AddFont(5)  // Returns "F1"
//	// In content stream: /F1 12 Tf (set font F1 at 12pt)
func (rd *ResourceDictionary) AddFont(objNum int) string {
	name := fmt.Sprintf("F%d", len(rd.fonts)+1)
	rd.fonts[name] = objNum
	return name
}

// AddFontWithID adds a font resource with an associated ID and returns its resource name.
//
// The fontID is used to later set the correct object number via SetFontObjNumByID.
// This enables correct font object assignment when fonts are created after content streams.
//
// Parameters:
//   - objNum: PDF object number (can be 0 as placeholder)
//   - fontID: Unique identifier for this font (e.g., "custom:font_1" or "std:Helvetica")
//
// Returns:
//   - Resource name (e.g., "F1")
func (rd *ResourceDictionary) AddFontWithID(objNum int, fontID string) string {
	name := fmt.Sprintf("F%d", len(rd.fonts)+1)
	rd.fonts[name] = objNum
	rd.fontIDs[fontID] = name
	return name
}

// SetFontObjNumByID sets the object number for a font identified by its ID.
//
// This is used after font objects are created to update the placeholder object numbers.
//
// Returns true if the font was found and updated, false otherwise.
func (rd *ResourceDictionary) SetFontObjNumByID(fontID string, objNum int) bool {
	resName, ok := rd.fontIDs[fontID]
	if !ok {
		return false
	}
	rd.fonts[resName] = objNum
	return true
}

// GetFontIDMapping returns a copy of the font ID to resource name mapping.
//
// This is useful for debugging and testing.
func (rd *ResourceDictionary) GetFontIDMapping() map[string]string {
	result := make(map[string]string, len(rd.fontIDs))
	for k, v := range rd.fontIDs {
		result[k] = v
	}
	return result
}

// AddImage adds an image XObject resource and returns its resource name.
//
// Images are named sequentially: Im1, Im2, Im3, etc.
//
// Parameters:
//   - objNum: PDF object number of the image XObject
//
// Returns:
//   - Resource name (e.g., "Im1")
//
// Example:
//
//	rd := NewResourceDictionary()
//	name := rd.AddImage(10)  // Returns "Im1"
//	// In content stream: /Im1 Do (draw image Im1)
func (rd *ResourceDictionary) AddImage(objNum int) string {
	name := fmt.Sprintf("Im%d", len(rd.xobjects)+1)
	rd.xobjects[name] = objNum
	return name
}

// AddExtGState adds a graphics state resource and returns its resource name.
//
// Graphics states are named sequentially: GS1, GS2, GS3, etc.
//
// Parameters:
//   - objNum: PDF object number of the ExtGState dictionary
//
// Returns:
//   - Resource name (e.g., "GS1")
//
// Example:
//
//	rd := NewResourceDictionary()
//	name := rd.AddExtGState(15)  // Returns "GS1"
//	// In content stream: /GS1 gs (apply graphics state GS1)
func (rd *ResourceDictionary) AddExtGState(objNum int) string {
	name := fmt.Sprintf("GS%d", len(rd.extgstates)+1)
	rd.extgstates[name] = objNum
	return name
}

// HasResources returns true if any resources are registered.
//
// Use this to check if the resource dictionary is empty before writing.
func (rd *ResourceDictionary) HasResources() bool {
	return len(rd.fonts) > 0 || len(rd.xobjects) > 0 || len(rd.extgstates) > 0
}

// Bytes returns the resource dictionary as PDF bytes.
//
// Format:
//
//	<< /Font << /F1 5 0 R >> /XObject << /Im1 10 0 R >> /ProcSet [/PDF /Text /ImageB /ImageC /ImageI] >>
//
// Returns empty dictionary if no resources are registered:
//
//	<< >>
//
// Note: Resource names are sorted alphabetically for consistent output.
func (rd *ResourceDictionary) Bytes() []byte {
	var buf bytes.Buffer
	buf.WriteString("<<")

	// Font resources.
	if len(rd.fonts) > 0 {
		buf.WriteString(" /Font <<")
		rd.writeSortedResources(&buf, rd.fonts)
		buf.WriteString(" >>")
	}

	// XObject resources (images, forms).
	if len(rd.xobjects) > 0 {
		buf.WriteString(" /XObject <<")
		rd.writeSortedResources(&buf, rd.xobjects)
		buf.WriteString(" >>")
	}

	// ExtGState resources (graphics states).
	if len(rd.extgstates) > 0 {
		buf.WriteString(" /ExtGState <<")
		rd.writeSortedResources(&buf, rd.extgstates)
		buf.WriteString(" >>")
	}

	// ProcSet (procedure set) - required for compatibility with old PDF readers.
	// Modern readers ignore this, but it's recommended for maximum compatibility.
	if rd.HasResources() {
		buf.WriteString(" /ProcSet [/PDF /Text /ImageB /ImageC /ImageI]")
	}

	buf.WriteString(" >>")

	return buf.Bytes()
}

// String returns the resource dictionary as a PDF string.
//
// Convenience method for debugging and testing.
func (rd *ResourceDictionary) String() string {
	return string(rd.Bytes())
}

// writeSortedResources writes resources to buffer in sorted order.
//
// Resources are sorted by name (F1, F2, F3, ...) for consistent output.
// This makes diffs more readable and tests more reliable.
//
// Format: /Name ObjNum 0 R.
func (rd *ResourceDictionary) writeSortedResources(buf *bytes.Buffer, resources map[string]int) {
	// Sort resource names for consistent output.
	names := make([]string, 0, len(resources))
	for name := range resources {
		names = append(names, name)
	}
	sort.Strings(names)

	// Write resources in sorted order.
	for _, name := range names {
		objNum := resources[name]
		fmt.Fprintf(buf, " /%s %d 0 R", name, objNum)
	}
}
