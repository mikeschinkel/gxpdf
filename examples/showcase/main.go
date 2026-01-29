// Package main demonstrates GxPDF enterprise capabilities with Unicode font support.
//
// This showcase creates a professional PDF document demonstrating:
// - Embedded TrueType fonts with full Unicode support
// - Cyrillic, Chinese, and special symbols
// - Professional layout with custom fonts
// - Graphics, gradients, and colors
package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/coregx/gxpdf/creator"
)

// Corporate color palette.
var (
	PrimaryDark = creator.Color{R: 0.05, G: 0.10, B: 0.20}
	PrimaryBlue = creator.Color{R: 0.00, G: 0.47, B: 0.75}
	AccentGold  = creator.Color{R: 0.85, G: 0.65, B: 0.13}
	TextDark    = creator.Color{R: 0.15, G: 0.15, B: 0.15}
	TextGray    = creator.Color{R: 0.40, G: 0.40, B: 0.40}
	LightGray   = creator.Color{R: 0.95, G: 0.95, B: 0.95}
	White       = creator.Color{R: 1.00, G: 1.00, B: 1.00}
)

func main() {
	// Find a Unicode-capable font.
	fontPath := findUnicodeFont()
	if fontPath == "" {
		log.Fatal("No Unicode font found. Please install Arial or similar TTF font.")
	}

	fmt.Printf("Using font: %s\n", fontPath)

	// Load the font.
	font, err := creator.LoadFont(fontPath)
	if err != nil {
		log.Fatalf("Failed to load font: %v", err)
	}

	// Try to load bold variant.
	boldPath := findBoldFont()
	var fontBold *creator.CustomFont
	if boldPath != "" {
		fontBold, err = creator.LoadFont(boldPath)
		if err != nil {
			fontBold = font // Fallback to regular.
		}
	} else {
		fontBold = font
	}

	c := creator.New()
	c.SetTitle("GxPDF Unicode Showcase")
	c.SetAuthor("CoreGX Technologies")
	c.SetSubject("Enterprise PDF with Unicode Support")

	// Page 1: Hero page with Unicode text.
	if err := createHeroPage(c, font, fontBold); err != nil {
		log.Fatalf("Failed to create hero page: %v", err)
	}

	// Page 2: Features with multilingual text.
	if err := createFeaturesPage(c, font, fontBold); err != nil {
		log.Fatalf("Failed to create features page: %v", err)
	}

	// Page 3: Technical specifications.
	if err := createSpecsPage(c, font, fontBold); err != nil {
		log.Fatalf("Failed to create specs page: %v", err)
	}

	// Write PDF.
	outputPath := "assets/gxpdf_enterprise_brochure.pdf"
	if err := c.WriteToFile(outputPath); err != nil {
		log.Fatalf("Failed to write PDF: %v", err)
	}

	fmt.Printf("Created: %s\n", outputPath)
}

// findUnicodeFont finds a TTF font that supports Unicode.
func findUnicodeFont() string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = []string{
			"C:/Windows/Fonts/arial.ttf",
			"C:/Windows/Fonts/calibri.ttf",
			"C:/Windows/Fonts/segoeui.ttf",
			"C:/Windows/Fonts/tahoma.ttf",
		}
	case "darwin":
		paths = []string{
			"/System/Library/Fonts/Helvetica.ttc",
			"/Library/Fonts/Arial.ttf",
			"/System/Library/Fonts/SFNS.ttf",
		}
	default: // Linux
		paths = []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/TTF/DejaVuSans.ttf",
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// findBoldFont finds a bold TTF font.
func findBoldFont() string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = []string{
			"C:/Windows/Fonts/arialbd.ttf",
			"C:/Windows/Fonts/calibrib.ttf",
		}
	case "darwin":
		paths = []string{
			"/Library/Fonts/Arial Bold.ttf",
		}
	default:
		paths = []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// createHeroPage creates the main hero page with Unicode text.
func createHeroPage(c *creator.Creator, font, fontBold *creator.CustomFont) error {
	page, err := c.NewPage()
	if err != nil {
		return err
	}

	width := page.Width()
	height := page.Height()

	// Dark header background.
	if err := page.DrawRect(0, height-200, width, 200, &creator.RectOptions{
		FillColor: &PrimaryDark,
	}); err != nil {
		return err
	}

	// Main title - English.
	if err := page.AddTextCustomFontColor("GxPDF", 50, height-80, fontBold, 48, White); err != nil {
		return err
	}

	// Subtitle with Unicode.
	if err := page.AddTextCustomFontColor("Enterprise PDF Library for Go", 50, height-120, font, 18, White); err != nil {
		return err
	}

	// Tagline in multiple languages.
	y := height - 160.0
	if err := page.AddTextCustomFontColor("Professional PDFs in any language", 50, y, font, 14, AccentGold); err != nil {
		return err
	}

	// Feature highlights section.
	y = height - 280.0

	// Section title.
	if err := page.AddTextCustomFontColor("Unicode Support Demonstration", 50, y, fontBold, 24, PrimaryBlue); err != nil {
		return err
	}

	// Multilingual examples (scripts supported by Arial).
	// Note: CJK (Chinese, Japanese, Korean) and RTL (Arabic, Hebrew) require
	// specialized fonts. Use Noto Sans CJK for CJK support.
	examples := []struct {
		lang string
		text string
	}{
		{"English", "Hello, World! Professional PDF generation."},
		{"Russian", "Привет, мир! Профессиональная генерация PDF."},
		{"Ukrainian", "Привіт, світ! Професійна генерація PDF."},
		{"Bulgarian", "Здравей, свят! Професионално генериране на PDF."},
		{"Greek", "Γειά σου κόσμε! Επαγγελματική δημιουργία PDF."},
		{"Polish", "Cześć, świecie! Profesjonalne generowanie PDF."},
		{"Czech", "Ahoj, světe! Profesionální generování PDF."},
		{"German", "Hallo, Welt! Professionelle PDF-Erstellung."},
		{"French", "Bonjour le monde! Génération PDF professionnelle."},
		{"Spanish", "¡Hola, mundo! Generación de PDF profesional."},
		{"Turkish", "Merhaba dünya! Profesyonel PDF oluşturma."},
	}

	y -= 50
	for _, ex := range examples {
		// Language label.
		if err := page.AddTextCustomFontColor(ex.lang+":", 50, y, fontBold, 11, TextDark); err != nil {
			return err
		}
		// Text in that language.
		if err := page.AddTextCustomFontColor(ex.text, 130, y, font, 11, TextGray); err != nil {
			return err
		}
		y -= 25
	}

	// Symbols section.
	y -= 30
	if err := page.AddTextCustomFontColor("Special Symbols:", 50, y, fontBold, 14, PrimaryBlue); err != nil {
		return err
	}

	y -= 25
	symbols := "Mathematical: ∑ ∏ ∫ √ ∞ ≠ ≤ ≥ ± × ÷ π θ α β γ δ"
	if err := page.AddTextCustomFontColor(symbols, 50, y, font, 11, TextDark); err != nil {
		return err
	}

	y -= 20
	// Currency symbols supported by Arial.
	symbols2 := "Currency: $ € £ ¥ ₽ ₴ ¢ ₹ ₱"
	if err := page.AddTextCustomFontColor(symbols2, 50, y, font, 11, TextDark); err != nil {
		return err
	}

	y -= 20
	symbols3 := "Arrows: → ← ↑ ↓ ↔ ⇒ ⇐ ⇑ ⇓ ⇔"
	if err := page.AddTextCustomFontColor(symbols3, 50, y, font, 11, TextDark); err != nil {
		return err
	}

	// Footer.
	drawFooter(page, font, 1)

	return nil
}

// createFeaturesPage creates the features page.
func createFeaturesPage(c *creator.Creator, font, fontBold *creator.CustomFont) error {
	page, err := c.NewPage()
	if err != nil {
		return err
	}

	height := page.Height()
	y := height - 60.0

	// Title.
	if err := page.AddTextCustomFontColor("Key Features", 50, y, fontBold, 28, PrimaryDark); err != nil {
		return err
	}

	// Feature list.
	features := []struct {
		title string
		desc  string
	}{
		{
			"Full Unicode Support",
			"Latin, Cyrillic, Greek, and 65,000+ characters with proper fonts",
		},
		{
			"Font Embedding",
			"TrueType/OpenType fonts with automatic subsetting",
		},
		{
			"Professional Graphics",
			"Lines, rectangles, circles, polygons, Bezier curves",
		},
		{
			"Gradient Fills",
			"Linear and radial gradients with multiple color stops",
		},
		{
			"Document Security",
			"RC4 and AES encryption (40/128/256-bit)",
		},
		{
			"Table Extraction",
			"100% accuracy on complex documents (740/740 transactions)",
		},
		{
			"Interactive Forms",
			"Text fields, checkboxes, radio buttons, dropdowns",
		},
		{
			"Zero Dependencies",
			"Pure Go implementation, standard library only",
		},
	}

	y -= 50
	for _, f := range features {
		// Feature title.
		if err := page.AddTextCustomFontColor("• "+f.title, 50, y, fontBold, 14, PrimaryBlue); err != nil {
			return err
		}
		y -= 20
		// Feature description.
		if err := page.AddTextCustomFontColor("  "+f.desc, 50, y, font, 11, TextGray); err != nil {
			return err
		}
		y -= 35
	}

	// Code example section.
	y -= 20
	if err := page.AddTextCustomFontColor("Quick Start Example:", 50, y, fontBold, 16, PrimaryDark); err != nil {
		return err
	}

	// Code background.
	y -= 10
	if err := page.DrawRect(50, y-100, 500, 100, &creator.RectOptions{
		FillColor: &PrimaryDark,
	}); err != nil {
		return err
	}

	// Code lines.
	codeLines := []string{
		`font, _ := creator.LoadFont("fonts/Arial.ttf")`,
		`page.AddTextCustomFont("Привет мир!", 100, 700, font, 24)`,
		`c.WriteToFile("output.pdf")`,
	}

	codeY := y - 25.0
	for _, line := range codeLines {
		if err := page.AddTextCustomFontColor(line, 60, codeY, font, 10, White); err != nil {
			return err
		}
		codeY -= 18
	}

	// Footer.
	drawFooter(page, font, 2)

	return nil
}

// createSpecsPage creates the technical specifications page.
//
//nolint:gocyclo // Example function intentionally comprehensive.
func createSpecsPage(c *creator.Creator, font, fontBold *creator.CustomFont) error {
	page, err := c.NewPage()
	if err != nil {
		return err
	}

	height := page.Height()
	y := height - 60.0

	// Title.
	if err := page.AddTextCustomFontColor("Technical Specifications", 50, y, fontBold, 28, PrimaryDark); err != nil {
		return err
	}

	// Specs table.
	specs := [][]string{
		{"PDF Version", "1.7 (ISO 32000-1:2008)"},
		{"Go Version", "1.25+"},
		{"Dependencies", "Standard library only"},
		{"Font Support", "Standard 14 + TTF/OTF embedding"},
		{"Unicode", "Full BMP support (U+0000 to U+FFFF)"},
		{"Encryption", "RC4 (40/128-bit), AES (128/256-bit)"},
		{"Compression", "FlateDecode (zlib)"},
		{"Images", "JPEG, PNG with alpha"},
		{"License", "MIT"},
	}

	y -= 50
	for _, row := range specs {
		// Label.
		if err := page.AddTextCustomFontColor(row[0]+":", 50, y, fontBold, 11, TextDark); err != nil {
			return err
		}
		// Value.
		if err := page.AddTextCustomFontColor(row[1], 200, y, font, 11, TextGray); err != nil {
			return err
		}
		y -= 25
	}

	// Performance section.
	y -= 30
	if err := page.AddTextCustomFontColor("Performance Metrics", 50, y, fontBold, 18, PrimaryBlue); err != nil {
		return err
	}

	perfData := [][]string{
		{"Table extraction", "~500 pages/second"},
		{"Text extraction", "~1000 pages/second"},
		{"PDF generation", "~100 pages/second"},
		{"Memory usage", "<50MB for 1000-page documents"},
	}

	y -= 30
	for _, row := range perfData {
		if err := page.AddTextCustomFontColor("• "+row[0]+": ", 50, y, font, 11, TextDark); err != nil {
			return err
		}
		if err := page.AddTextCustomFontColor(row[1], 200, y, fontBold, 11, PrimaryBlue); err != nil {
			return err
		}
		y -= 22
	}

	// Links section.
	y -= 30
	if err := page.AddTextCustomFontColor("Resources", 50, y, fontBold, 18, PrimaryBlue); err != nil {
		return err
	}

	y -= 25
	if err := page.AddTextCustomFontColor("GitHub: github.com/coregx/gxpdf", 50, y, font, 11, TextDark); err != nil {
		return err
	}
	y -= 20
	if err := page.AddTextCustomFontColor("Documentation: pkg.go.dev/github.com/coregx/gxpdf", 50, y, font, 11, TextDark); err != nil {
		return err
	}

	// Footer.
	drawFooter(page, font, 3)

	return nil
}

// drawFooter draws the page footer.
func drawFooter(page *creator.Page, font *creator.CustomFont, pageNum int) {
	width := page.Width()

	// Footer line.
	_ = page.DrawLine(50, 50, width-50, 50, &creator.LineOptions{
		Color: LightGray,
		Width: 1,
	})

	// Page number.
	text := fmt.Sprintf("Page %d", pageNum)
	_ = page.AddTextCustomFontColor(text, width/2-20, 35, font, 9, TextGray)

	// Copyright.
	_ = page.AddTextCustomFontColor("© 2026 CoreGX Technologies", 50, 35, font, 9, TextGray)
}
