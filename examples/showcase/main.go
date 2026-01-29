// Package main demonstrates enterprise-grade PDF creation with GxPDF.
//
// This example creates a professional corporate brochure showcasing:
// - Professional header with branding
// - Styled typography with multiple fonts
// - Color scheme and graphics
// - Multi-column layouts
// - Footer with page numbers
//
// Run: go run examples/showcase/main.go
package main

import (
	"fmt"
	"log"

	"github.com/coregx/gxpdf/creator"
)

// Corporate color palette
var (
	// Primary colors
	PrimaryDark = creator.Color{R: 0.05, G: 0.10, B: 0.20} // Deep navy
	PrimaryBlue = creator.Color{R: 0.00, G: 0.47, B: 0.75} // Corporate blue
	AccentGold  = creator.Color{R: 0.85, G: 0.65, B: 0.13} // Premium gold

	// Neutral colors
	TextDark  = creator.Color{R: 0.15, G: 0.15, B: 0.15}
	TextGray  = creator.Color{R: 0.40, G: 0.40, B: 0.40}
	LightGray = creator.Color{R: 0.95, G: 0.95, B: 0.95}
	White     = creator.Color{R: 1.00, G: 1.00, B: 1.00}
)

func main() {
	c := creator.New()
	c.SetTitle("GxPDF Enterprise Capabilities")
	c.SetAuthor("CoreGX Technologies")
	c.SetSubject("Professional PDF Generation for Go")

	// Page 1: Cover
	if err := createCoverPage(c); err != nil {
		log.Fatalf("Cover page failed: %v", err)
	}

	// Page 2: Features Overview
	if err := createFeaturesPage(c); err != nil {
		log.Fatalf("Features page failed: %v", err)
	}

	// Page 3: Technical Specifications
	if err := createSpecsPage(c); err != nil {
		log.Fatalf("Specs page failed: %v", err)
	}

	// Save
	outputPath := "gxpdf_enterprise_brochure.pdf"
	if err := c.WriteToFile(outputPath); err != nil {
		log.Fatalf("Failed to write PDF: %v", err)
	}

	fmt.Printf("✓ Enterprise brochure created: %s\n", outputPath)
	fmt.Println("\nThis PDF demonstrates:")
	fmt.Println("  • Professional cover page with branding")
	fmt.Println("  • Corporate color scheme")
	fmt.Println("  • Typography hierarchy")
	fmt.Println("  • Vector graphics and shapes")
	fmt.Println("  • Multi-page document structure")
}

// createCoverPage creates an impressive cover page
func createCoverPage(c *creator.Creator) error {
	page, err := c.NewPage()
	if err != nil {
		return err
	}

	// Full-page dark background header (top 40%)
	if err := page.DrawRect(0, 505, 595, 337, &creator.RectOptions{
		FillColor: &PrimaryDark,
	}); err != nil {
		return err
	}

	// Accent gold line
	if err := page.DrawRect(0, 500, 595, 5, &creator.RectOptions{
		FillColor: &AccentGold,
	}); err != nil {
		return err
	}

	// Company logo area (simplified geometric logo)
	// Outer square
	if err := page.DrawRect(267, 720, 60, 60, &creator.RectOptions{
		FillColor:   &PrimaryBlue,
		StrokeColor: &White,
		StrokeWidth: 2,
	}); err != nil {
		return err
	}
	// Inner accent
	if err := page.DrawRect(282, 735, 30, 30, &creator.RectOptions{
		FillColor: &AccentGold,
	}); err != nil {
		return err
	}

	// Main title
	if err := page.AddTextColor("GxPDF", 200, 650, creator.HelveticaBold, 48, White); err != nil {
		return err
	}

	// Subtitle
	if err := page.AddTextColor("Enterprise PDF Library for Go", 150, 610, creator.Helvetica, 18, White); err != nil {
		return err
	}

	// Tagline
	if err := page.AddTextColor("Pure Go  |  Zero Dependencies  |  Production Ready", 120, 540, creator.HelveticaOblique, 14, AccentGold); err != nil {
		return err
	}

	// Feature highlights section (below gold line)
	features := []struct {
		title string
		desc  string
		y     float64
	}{
		{"100% Accuracy", "Table extraction tested on 740+ transactions", 420},
		{"Enterprise Security", "AES-256 & RC4 encryption support", 340},
		{"Full Featured", "Create, read, merge, split, encrypt", 260},
	}

	for _, f := range features {
		// Icon - diamond shape using rectangle
		if err := page.DrawRect(75, f.y-5, 12, 12, &creator.RectOptions{
			FillColor: &PrimaryBlue,
		}); err != nil {
			return err
		}
		// Title
		if err := page.AddTextColor(f.title, 100, f.y, creator.HelveticaBold, 16, TextDark); err != nil {
			return err
		}
		// Description
		if err := page.AddTextColor(f.desc, 100, f.y-22, creator.Helvetica, 11, TextGray); err != nil {
			return err
		}
	}

	// Bottom section - call to action
	if err := page.DrawRect(0, 50, 595, 80, &creator.RectOptions{
		FillColor: &LightGray,
	}); err != nil {
		return err
	}

	if err := page.AddTextColor("github.com/coregx/gxpdf", 200, 100, creator.CourierBold, 14, PrimaryBlue); err != nil {
		return err
	}
	if err := page.AddTextColor("MIT Licensed  -  Open Source  -  Free Forever", 175, 75, creator.Helvetica, 11, TextGray); err != nil {
		return err
	}

	return nil
}

// createFeaturesPage creates the features overview page
func createFeaturesPage(c *creator.Creator) error {
	page, err := c.NewPage()
	if err != nil {
		return err
	}

	// Header bar
	if err := page.DrawRect(0, 792, 595, 50, &creator.RectOptions{
		FillColor: &PrimaryDark,
	}); err != nil {
		return err
	}
	if err := page.AddTextColor("FEATURES", 250, 810, creator.HelveticaBold, 18, White); err != nil {
		return err
	}

	// Gold accent line under header
	if err := page.DrawRect(0, 790, 595, 3, &creator.RectOptions{
		FillColor: &AccentGold,
	}); err != nil {
		return err
	}

	// Section: PDF Reading
	y := 720.0
	if err := drawFeatureSection(page, "PDF Reading & Extraction", y, []string{
		"- Parse any PDF 1.0 - 2.0 document",
		"- Extract text with position information",
		"- Extract tables with 4-Pass Hybrid algorithm",
		"- Extract images (JPEG, PNG)",
		"- Read document metadata and properties",
	}); err != nil {
		return err
	}

	// Section: PDF Creation
	y = 550.0
	if err := drawFeatureSection(page, "PDF Creation & Editing", y, []string{
		"- Create new PDFs from scratch",
		"- Standard 14 fonts + TTF/OTF embedding",
		"- Graphics: lines, rectangles, circles, polygons",
		"- Images: embed JPEG and PNG",
		"- Merge, split, and rotate pages",
	}); err != nil {
		return err
	}

	// Section: Security
	y = 380.0
	if err := drawFeatureSection(page, "Security & Encryption", y, []string{
		"- AES-128 and AES-256 encryption",
		"- RC4 40-bit and 128-bit encryption",
		"- Password protection (user & owner)",
		"- Permission controls (print, copy, edit)",
	}); err != nil {
		return err
	}

	// Section: Export
	y = 230.0
	if err := drawFeatureSection(page, "Export & Integration", y, []string{
		"- Export tables to CSV, JSON, Excel",
		"- CLI tool for scripting and automation",
		"- Clean Go API with comprehensive docs",
		"- Context support for cancellation",
	}); err != nil {
		return err
	}

	// Footer
	drawFooter(page, 2)

	return nil
}

// createSpecsPage creates technical specifications page
//
//nolint:gocyclo // Example function intentionally comprehensive to demonstrate all features.
func createSpecsPage(c *creator.Creator) error {
	page, err := c.NewPage()
	if err != nil {
		return err
	}

	// Header bar
	if err := page.DrawRect(0, 792, 595, 50, &creator.RectOptions{
		FillColor: &PrimaryDark,
	}); err != nil {
		return err
	}
	if err := page.AddTextColor("TECHNICAL SPECIFICATIONS", 180, 810, creator.HelveticaBold, 18, White); err != nil {
		return err
	}
	if err := page.DrawRect(0, 790, 595, 3, &creator.RectOptions{
		FillColor: &AccentGold,
	}); err != nil {
		return err
	}

	// Specs table
	specs := [][]string{
		{"Language", "Pure Go (no CGO)"},
		{"Go Version", "1.21+"},
		{"License", "MIT"},
		{"Dependencies", "Zero external dependencies"},
		{"PDF Versions", "1.0 - 2.0"},
		{"Platforms", "Linux, macOS, Windows"},
		{"Architectures", "amd64, arm64, 386"},
	}

	y := 700.0
	for i, spec := range specs {
		// Alternating row background
		if i%2 == 0 {
			if err := page.DrawRect(50, y-5, 495, 25, &creator.RectOptions{
				FillColor: &LightGray,
			}); err != nil {
				return err
			}
		}
		// Label
		if err := page.AddTextColor(spec[0], 60, y, creator.HelveticaBold, 11, TextDark); err != nil {
			return err
		}
		// Value
		if err := page.AddTextColor(spec[1], 250, y, creator.Helvetica, 11, TextGray); err != nil {
			return err
		}
		y -= 30
	}

	// Performance section
	y = 450.0
	if err := page.AddTextColor("Performance Benchmarks", 50, y, creator.HelveticaBold, 16, PrimaryBlue); err != nil {
		return err
	}
	if err := page.DrawLine(50, y-5, 300, y-5, &creator.LineOptions{Color: AccentGold, Width: 2}); err != nil {
		return err
	}

	benchmarks := []string{
		"Table extraction: ~200ms for 15-page document",
		"PDF creation: 28.4 us/page",
		"Text rendering: 11.2 us/operation",
		"Memory: ~15MB peak for complex documents",
	}

	y = 410.0
	for _, b := range benchmarks {
		if err := page.AddTextColor("> "+b, 60, y, creator.Helvetica, 11, TextDark); err != nil {
			return err
		}
		y -= 22
	}

	// Code example box
	y = 280.0
	if err := page.DrawRect(50, y-80, 495, 100, &creator.RectOptions{
		FillColor:   &PrimaryDark,
		StrokeColor: &AccentGold,
		StrokeWidth: 1,
	}); err != nil {
		return err
	}

	if err := page.AddTextColor("Quick Start", 60, y, creator.HelveticaBold, 12, AccentGold); err != nil {
		return err
	}

	codeLines := []string{
		"doc, _ := gxpdf.Open(\"document.pdf\")",
		"tables := doc.ExtractTables()",
		"csv, _ := tables[0].ToCSV()",
	}
	y -= 25
	for _, line := range codeLines {
		if err := page.AddTextColor(line, 70, y, creator.Courier, 10, White); err != nil {
			return err
		}
		y -= 18
	}

	// Contact/CTA section
	y = 120.0
	if err := page.DrawRect(50, y-40, 495, 60, &creator.RectOptions{
		StrokeColor: &PrimaryBlue,
		StrokeWidth: 2,
	}); err != nil {
		return err
	}

	if err := page.AddTextColor("Ready to Get Started?", 200, y, creator.HelveticaBold, 14, PrimaryBlue); err != nil {
		return err
	}
	if err := page.AddTextColor("go get github.com/coregx/gxpdf@v0.1.0", 170, y-25, creator.Courier, 12, TextDark); err != nil {
		return err
	}

	// Footer
	drawFooter(page, 3)

	return nil
}

// drawFeatureSection draws a feature section with title and bullet points
func drawFeatureSection(page *creator.Page, title string, y float64, items []string) error {
	// Section title with blue accent
	if err := page.AddTextColor(title, 50, y, creator.HelveticaBold, 16, PrimaryBlue); err != nil {
		return err
	}

	// Underline
	if err := page.DrawLine(50, y-5, 250, y-5, &creator.LineOptions{
		Color: AccentGold,
		Width: 2,
	}); err != nil {
		return err
	}

	// Items
	itemY := y - 30
	for _, item := range items {
		if err := page.AddTextColor(item, 60, itemY, creator.Helvetica, 11, TextDark); err != nil {
			return err
		}
		itemY -= 20
	}

	return nil
}

// drawFooter draws the page footer
func drawFooter(page *creator.Page, pageNum int) {
	// Footer line
	_ = page.DrawLine(50, 50, 545, 50, &creator.LineOptions{
		Color: LightGray,
		Width: 1,
	})

	// Company name
	_ = page.AddTextColor("CoreGX Technologies", 50, 30, creator.Helvetica, 9, TextGray)

	// Page number
	_ = page.AddTextColor(fmt.Sprintf("Page %d", pageNum), 520, 30, creator.Helvetica, 9, TextGray)

	// Confidential mark
	_ = page.AddTextColor("CONFIDENTIAL", 270, 30, creator.HelveticaBold, 8, PrimaryBlue)
}
