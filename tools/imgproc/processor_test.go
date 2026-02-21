package main

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestResizeImage_LargerThanMax(t *testing.T) {
	// Create 4000x3000 image
	bounds := image.Rect(0, 0, 4000, 3000)
	img := image.NewRGBA(bounds)

	resized := resizeImage(img, 2000)

	if resized.Bounds().Dx() != 2000 {
		t.Errorf("expected width 2000, got %d", resized.Bounds().Dx())
	}
	if resized.Bounds().Dy() != 1500 {
		t.Errorf("expected height 1500, got %d", resized.Bounds().Dy())
	}
}

func TestResizeImage_SmallerThanMax(t *testing.T) {
	// Create 800x600 image
	bounds := image.Rect(0, 0, 800, 600)
	img := image.NewRGBA(bounds)

	resized := resizeImage(img, 2000)

	if resized.Bounds().Dx() != 800 {
		t.Errorf("expected width 800, got %d", resized.Bounds().Dx())
	}
	if resized.Bounds().Dy() != 600 {
		t.Errorf("expected height 600, got %d", resized.Bounds().Dy())
	}
}

func TestResizeImage_Portrait(t *testing.T) {
	// Create 3000x4000 portrait image
	bounds := image.Rect(0, 0, 3000, 4000)
	img := image.NewRGBA(bounds)

	resized := resizeImage(img, 2000)

	if resized.Bounds().Dx() != 1500 {
		t.Errorf("expected width 1500, got %d", resized.Bounds().Dx())
	}
	if resized.Bounds().Dy() != 2000 {
		t.Errorf("expected height 2000, got %d", resized.Bounds().Dy())
	}
}

func TestImageMetadata(t *testing.T) {
	// Create 1600x1200 image
	bounds := image.Rect(0, 0, 1600, 1200)
	img := image.NewRGBA(bounds)

	meta := extractMetadata("test.jpg", img)

	if meta.Width != 1600 {
		t.Errorf("expected width 1600, got %d", meta.Width)
	}
	if meta.Height != 1200 {
		t.Errorf("expected height 1200, got %d", meta.Height)
	}
	expectedAspectRatio := 1600.0 / 1200.0
	if meta.AspectRatio != expectedAspectRatio {
		t.Errorf("expected aspect ratio %f, got %f", expectedAspectRatio, meta.AspectRatio)
	}
	if meta.Src != "test.jpg" {
		t.Errorf("expected src test.jpg, got %s", meta.Src)
	}
}

func TestDecodeImage_PNG(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.png")

	// Create a simple PNG
	bounds := image.Rect(0, 0, 100, 100)
	img := image.NewRGBA(bounds)
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}

	decoded, err := decodeImage(testPath)
	if err != nil {
		t.Fatalf("failed to decode image: %v", err)
	}

	if decoded.Bounds().Dx() != 100 || decoded.Bounds().Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestDecodeImage_JPEG(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.jpg")

	// Create a simple JPEG
	bounds := image.Rect(0, 0, 100, 100)
	img := image.NewRGBA(bounds)
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("failed to encode JPEG: %v", err)
	}

	decoded, err := decodeImage(testPath)
	if err != nil {
		t.Fatalf("failed to decode image: %v", err)
	}

	if decoded.Bounds().Dx() != 100 || decoded.Bounds().Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestEncodeWebP(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.webp")

	// Create a simple image
	bounds := image.Rect(0, 0, 100, 100)
	img := image.NewRGBA(bounds)

	err := encodeWebP(testPath, img, 85)
	if err != nil {
		t.Fatalf("failed to encode WebP: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testPath); err != nil {
		t.Errorf("WebP file was not created: %v", err)
	}

	// Verify it's readable
	decoded, err := decodeImage(testPath)
	if err != nil {
		t.Fatalf("failed to decode WebP: %v", err)
	}

	if decoded.Bounds().Dx() != 100 || decoded.Bounds().Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"image.jpg", true},
		{"image.jpeg", true},
		{"image.png", true},
		{"image.JPG", true},
		{"image.JPEG", true},
		{"image.PNG", true},
		{"image.gif", false},
		{"image.txt", false},
		{"noext", false},
	}

	for _, test := range tests {
		if got := isImageFile(test.filename); got != test.expected {
			t.Errorf("isImageFile(%q) = %v, expected %v", test.filename, got, test.expected)
		}
	}
}

func TestReplaceExt(t *testing.T) {
	tests := []struct {
		name     string
		newExt   string
		expected string
	}{
		{"image.jpg", ".webp", "image.webp"},
		{"image.jpeg", ".webp", "image.webp"},
		{"image.png", ".webp", "image.webp"},
		{"path/to/image.jpg", ".webp", "path/to/image.webp"},
	}

	for _, test := range tests {
		if got := replaceExt(test.name, test.newExt); got != test.expected {
			t.Errorf("replaceExt(%q, %q) = %q, expected %q", test.name, test.newExt, got, test.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{512, "512 B"},
	}

	for _, test := range tests {
		if got := formatBytes(test.bytes); got != test.expected {
			t.Errorf("formatBytes(%d) = %q, expected %q", test.bytes, got, test.expected)
		}
	}
}

func TestExtractMetadataWithRealFile(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.jpg")

	// Create a 800x600 JPEG
	bounds := image.Rect(0, 0, 800, 600)
	img := image.NewRGBA(bounds)
	file, err := os.Create(testPath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("failed to encode JPEG: %v", err)
	}

	decoded, err := decodeImage(testPath)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	meta := extractMetadata("test.jpg", decoded)

	if meta.Width != 800 {
		t.Errorf("expected width 800, got %d", meta.Width)
	}
	if meta.Height != 600 {
		t.Errorf("expected height 600, got %d", meta.Height)
	}
	expectedRatio := 800.0 / 600.0
	if meta.AspectRatio != expectedRatio {
		t.Errorf("expected ratio %f, got %f", expectedRatio, meta.AspectRatio)
	}
}
