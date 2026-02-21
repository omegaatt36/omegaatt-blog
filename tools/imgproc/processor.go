package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"
)

type ImageMeta struct {
	Src         string  `json:"src"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	AspectRatio float64 `json:"aspectRatio"`
}

type GalleryJSON struct {
	Images []ImageMeta `json:"images"`
}

type ProcessConfig struct {
	MaxWidth  int
	Quality   int
	Threshold int64 // bytes
	DryRun    bool
}

// resizeImage resizes the image to fit within maxWidth while maintaining aspect ratio.
// If the larger dimension exceeds maxWidth, scale down proportionally.
// If both dimensions are within maxWidth, return unchanged.
func resizeImage(img image.Image, maxWidth int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check if resizing is needed
	if width <= maxWidth && height <= maxWidth {
		return img
	}

	// Determine scaling factor based on the larger dimension
	var scale float64
	if width > height {
		scale = float64(maxWidth) / float64(width)
	} else {
		scale = float64(maxWidth) / float64(height)
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Create the resized image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

// extractMetadata extracts metadata from an image
func extractMetadata(filename string, img image.Image) ImageMeta {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	aspectRatio := float64(width) / float64(height)

	return ImageMeta{
		Src:         filename,
		Width:       width,
		Height:      height,
		AspectRatio: aspectRatio,
	}
}

// decodeImage decodes a JPEG or PNG from file path
func decodeImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Decode(file)
	case ".png":
		return png.Decode(file)
	case ".webp":
		return webp.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
}

// encodeWebP encodes an image to WebP format
func encodeWebP(path string, img image.Image, quality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return webp.Encode(file, img, &webp.Options{Quality: float32(quality)})
}

// isImageFile checks if a file is a supported image format (case-insensitive)
func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

// replaceExt replaces the file extension
func replaceExt(name, newExt string) string {
	return strings.TrimSuffix(name, filepath.Ext(name)) + newExt
}

// formatBytes returns a human-readable size string
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		return fmt.Sprintf("%.1f PB", float64(b)/float64(div*unit))
	}

	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), units[exp])
}

// processSubdir processes all images in a single directory
func processSubdir(dir string, cfg ProcessConfig) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var metadata []ImageMeta

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Process image files (jpg, jpeg, png)
		if isImageFile(filename) {
			imagePath := filepath.Join(dir, filename)

			if cfg.DryRun {
				info, _ := entry.Info()
				size := int64(0)
				if info != nil {
					size = info.Size()
				}
				fmt.Printf("  [DRY-RUN] Would convert: %s (%s)\n", imagePath, formatBytes(size))
				continue
			}

			// Decode image
			img, err := decodeImage(imagePath)
			if err != nil {
				fmt.Printf("Warning: failed to decode %s: %v\n", imagePath, err)
				continue
			}

			// Resize image
			resized := resizeImage(img, cfg.MaxWidth)

			// Extract metadata before encoding
			webpName := replaceExt(filename, ".webp")
			meta := extractMetadata(webpName, resized)

			// Encode to WebP
			webpPath := filepath.Join(dir, webpName)
			if err := encodeWebP(webpPath, resized, cfg.Quality); err != nil {
				fmt.Printf("Warning: failed to encode %s to WebP: %v\n", imagePath, err)
				continue
			}

			metadata = append(metadata, meta)

			// Delete original file
			if err := os.Remove(imagePath); err != nil {
				fmt.Printf("Warning: failed to delete original %s: %v\n", imagePath, err)
			}

			fmt.Printf("  Processed: %s -> %s\n", imagePath, webpPath)
		} else if strings.ToLower(filepath.Ext(filename)) == ".webp" {
			// Include existing WebP files for idempotency
			meta, err := readWebPMeta(dir, filename)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				continue
			}
			metadata = append(metadata, meta)
		}
	}

	if cfg.DryRun {
		return nil
	}

	return writeGalleryJSON(dir, metadata)
}

// regenSubdir regenerates gallery.json from existing WebP files without any image processing
func regenSubdir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var metadata []ImageMeta

	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".webp" {
			continue
		}

		meta, err := readWebPMeta(dir, entry.Name())
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			continue
		}
		metadata = append(metadata, meta)
	}

	return writeGalleryJSON(dir, metadata)
}

// readWebPMeta reads dimensions from an existing WebP file and returns metadata
func readWebPMeta(dir, filename string) (ImageMeta, error) {
	webpPath := filepath.Join(dir, filename)
	img, err := decodeImage(webpPath)
	if err != nil {
		return ImageMeta{}, fmt.Errorf("failed to decode %s: %w", webpPath, err)
	}
	return extractMetadata(filename, img), nil
}

// writeGalleryJSON sorts metadata and writes gallery.json to the directory
func writeGalleryJSON(dir string, metadata []ImageMeta) error {
	if len(metadata) == 0 {
		return nil
	}

	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Src < metadata[j].Src
	})

	gallery := GalleryJSON{Images: metadata}
	data, err := json.MarshalIndent(gallery, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal gallery JSON: %w", err)
	}

	jsonPath := filepath.Join(dir, "gallery.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write gallery.json: %w", err)
	}

	fmt.Printf("  Wrote: %s (%d images)\n", jsonPath, len(metadata))
	return nil
}

// findImageDirs finds all subdirectories containing images (jpg/png/webp)
func findImageDirs(dir string) ([]string, error) {
	var imageDirs []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if isImageFile(entry.Name()) || ext == ".webp" {
					imageDirs = append(imageDirs, path)
					break
				}
			}
		}
		return nil
	})

	return imageDirs, err
}

// runConcurrent runs fn on each directory concurrently with max 4 goroutines
func runConcurrent(dirs []string, fn func(string) error) error {
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	errChan := make(chan error, len(dirs))

	for _, d := range dirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := fn(dir); err != nil {
				errChan <- fmt.Errorf("%s: %w", dir, err)
			}
		}(d)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

// processDirectory walks the directory tree and processes image subdirectories concurrently
func processDirectory(dir string, cfg ProcessConfig) error {
	imageDirs, err := findImageDirs(dir)
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	return runConcurrent(imageDirs, func(d string) error {
		return processSubdir(d, cfg)
	})
}

// regenDirectory regenerates gallery.json for all subdirectories containing WebP files
func regenDirectory(dir string) error {
	imageDirs, err := findImageDirs(dir)
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	return runConcurrent(imageDirs, regenSubdir)
}
