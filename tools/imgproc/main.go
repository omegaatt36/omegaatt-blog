package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "process":
		processCmd()
	case "regen":
		regenCmd()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: imgproc <command> <dir> [flags]

Commands:
  process     Convert images to WebP and generate gallery.json
  regen       Regenerate gallery.json from existing WebP files (no conversion)

Process flags:
  --max-width    Maximum width for resized images (default: 2000)
  --quality      JPEG/WebP quality (default: 85)
  --threshold    Skip files smaller than threshold in bytes (default: 500000)
  --dry-run      Show what would be done without modifying files (default: false)
`)
}

func processCmd() {
	// Parse arguments
	fs := flag.NewFlagSet("process", flag.ExitOnError)

	maxWidth := fs.Int("max-width", 2000, "Maximum width for resized images")
	quality := fs.Int("quality", 85, "JPEG/WebP quality")
	threshold := fs.Int64("threshold", 500000, "Skip files smaller than threshold in bytes")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without modifying files")

	// Parse remaining arguments after "process"
	fs.Parse(os.Args[2:])

	// Get directory argument
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: directory argument required\n\n")
		printUsage()
		os.Exit(1)
	}

	dir := args[0]

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", dir)
		os.Exit(1)
	}

	// Create config
	cfg := ProcessConfig{
		MaxWidth:  *maxWidth,
		Quality:   *quality,
		Threshold: *threshold,
		DryRun:    *dryRun,
	}

	// Process directory
	if *dryRun {
		fmt.Printf("DRY RUN: Processing %s\n", dir)
	} else {
		fmt.Printf("Processing %s\n", dir)
	}

	if err := processDirectory(dir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}

func regenCmd() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: directory argument required\n\n")
		printUsage()
		os.Exit(1)
	}

	dir := os.Args[2]

	info, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", dir)
		os.Exit(1)
	}

	fmt.Printf("Regenerating gallery.json in %s\n", dir)

	if err := regenDirectory(dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}
