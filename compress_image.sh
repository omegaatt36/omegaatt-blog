#!/bin/bash

# Image Compression Script
# Usage: ./compress_image.sh [path] [options]
# Options:
#   -d, --dry-run    Show what would be compressed without making changes
#   -t, --threshold  Size threshold in KB (default: 500)
#   -h, --help       Show this help message

set -euo pipefail

# Default values
SEARCH_PATH="."
DRY_RUN=false
THRESHOLD_KB=500
THRESHOLD_BYTES=$((THRESHOLD_KB * 1024))

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

show_help() {
    echo "Usage: $0 [path] [options]"
    echo ""
    echo "Arguments:"
    echo "  path              Directory to search for images (default: current directory)"
    echo ""
    echo "Options:"
    echo "  -d, --dry-run     Show what would be compressed without making changes"
    echo "  -t, --threshold   Size threshold in KB (default: 500)"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Compress images in current directory"
    echo "  $0 ./content/images                   # Compress images in specific path"
    echo "  $0 ./images -d                        # Dry run to preview changes"
    echo "  $0 ./images -t 200                    # Use 200KB threshold"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -t|--threshold)
            THRESHOLD_KB="$2"
            THRESHOLD_BYTES=$((THRESHOLD_KB * 1024))
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        -*)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
        *)
            SEARCH_PATH="$1"
            shift
            ;;
    esac
done

# Validate search path
if [[ ! -d "$SEARCH_PATH" ]]; then
    log_error "Directory not found: $SEARCH_PATH"
    exit 1
fi

# Check dependencies
check_dependency() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 is required but not installed."
        exit 1
    fi
}

check_dependency "magick"
check_dependency "identify"

# Check for optipng (optional)
HAS_OPTIPNG=false
if command -v optipng &> /dev/null; then
    HAS_OPTIPNG=true
fi

log_info "Searching for images in: $SEARCH_PATH"
log_info "Size threshold: ${THRESHOLD_KB}KB"
log_info "Dry run: $DRY_RUN"
echo ""

# Counters
total_files=0
compressed_files=0
total_saved=0

# Get file extension in lowercase
get_ext_lower() {
    local file="$1"
    local ext="${file##*.}"
    echo "$ext" | tr '[:upper:]' '[:lower:]'  # POSIX-compatible lowercase
}

# Get file size
get_file_size() {
    stat -f "%z" "$1" 2>/dev/null || stat --format="%s" "$1" 2>/dev/null
}

# Format bytes to human readable
format_size() {
    local bytes=$1
    if [[ $bytes -ge 1048576 ]]; then
        echo "$(echo "scale=1; $bytes / 1048576" | bc)MB"
    elif [[ $bytes -ge 1024 ]]; then
        echo "$(echo "scale=1; $bytes / 1024" | bc)KB"
    else
        echo "${bytes}B"
    fi
}

compress_image() {
    local file="$1"
    local original_size
    original_size=$(get_file_size "$file")

    if [[ $original_size -le $THRESHOLD_BYTES ]]; then
        return 0
    fi

    ((total_files++))

    local ext
    ext=$(get_ext_lower "$file")

    # Get image dimensions
    local width height
    width=$(identify -format "%w" "$file" 2>/dev/null) || { log_warning "Cannot read dimensions: $file"; return 1; }
    height=$(identify -format "%h" "$file" 2>/dev/null) || { log_warning "Cannot read dimensions: $file"; return 1; }

    local size_human
    size_human=$(format_size "$original_size")

    if [[ "$DRY_RUN" == true ]]; then
        log_info "[DRY-RUN] Would compress: $file (${width}x${height}, $size_human)"
        return 0
    fi

    log_info "Compressing: $file (${width}x${height}, $size_human)"

    local temp_file="${file}.tmp.${ext}"

    case "$ext" in
        png)
            if [[ $width -gt 1500 || $height -gt 1500 ]]; then
                # Resize large PNGs and optimize
                magick "$file" -resize "1500x1500>" -quality 95 "$temp_file"
            elif [[ "$HAS_OPTIPNG" == true ]]; then
                # Lossless optimization for smaller PNGs
                cp "$file" "$temp_file"
                optipng -o5 -quiet "$temp_file"
            else
                # Fallback: just re-encode with magick
                magick "$file" -quality 95 "$temp_file"
            fi
            ;;
        jpg|jpeg)
            if [[ $width -gt 2000 || $height -gt 2000 ]]; then
                # Resize very large images
                magick "$file" -resize "2000x2000>" -sampling-factor 4:2:0 -strip -quality 82 -interlace Plane "$temp_file"
            elif [[ $width -gt 1500 || $height -gt 1500 ]]; then
                # Moderate resize for large images
                magick "$file" -resize "1500x1500>" -sampling-factor 4:2:0 -strip -quality 85 -interlace Plane "$temp_file"
            else
                # Just compress without resize
                magick "$file" -sampling-factor 4:2:0 -strip -quality 85 -interlace Plane "$temp_file"
            fi
            ;;
        *)
            log_warning "Unsupported format: $ext"
            return 1
            ;;
    esac

    if [[ -f "$temp_file" ]]; then
        local new_size
        new_size=$(get_file_size "$temp_file")

        # Only replace if we actually saved space
        if [[ $new_size -lt $original_size ]]; then
            local saved=$((original_size - new_size))
            local new_size_human
            new_size_human=$(format_size "$new_size")
            local saved_human
            saved_human=$(format_size "$saved")

            mv "$temp_file" "$file"
            ((compressed_files++))
            ((total_saved += saved))

            log_success "Compressed: $size_human -> $new_size_human (saved $saved_human)"
        else
            rm "$temp_file"
            log_warning "Skipped (no size reduction): $file"
        fi
    else
        log_error "Failed to create compressed file: $file"
        return 1
    fi
}

# Find and process images using process substitution to avoid subshell variable scope issue
while IFS= read -r -d $'\0' file; do
    compress_image "$file" || true
done < <(find "$SEARCH_PATH" -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" \) -print0)

echo ""
echo "========================================"
log_info "Summary:"
echo "  Files checked: $total_files"
echo "  Files compressed: $compressed_files"
echo "  Total saved: $(format_size $total_saved)"
echo "========================================"
