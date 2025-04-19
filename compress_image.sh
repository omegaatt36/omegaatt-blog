find . -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" \) -print0 |
  while IFS= read -r -d $'\0' file; do
    # Use macOS/BSD compatible stat command to get file size
    if [[ $(stat -f "%z" "$file") -gt 1048576 ]]; then
      echo "Compressing $file"
      ext="${file##*.}"
      # Get image dimensions using identify
      width=$(identify -format "%w" "$file")
      height=$(identify -format "%h" "$file")
      if [[ "$width" -gt 1000 || "$height" -gt 1000 ]]; then
         if [[ "$ext" == "png" ]]; then
          # Resize PNG in place (consider using mogrify or temp file for safety)
          magick convert "$file" -resize 80%  "$file"
         else
          # Resize and compress JPG/JPEG
          # Using a temp file approach for safety during conversion
          temp_file="${file%.*}_compressed.$ext"
          magick convert "$file" -resize 50% -sampling-factor 4:2:0 -strip -quality 85% "$temp_file" && \
          rm "$file" && \
          mv "$temp_file" "$file"
         fi
      else
        if [[ "$ext" == "png" ]]; then
          # Optimize PNG losslessly
          optipng -o7 "$file"
        else
          # Compress JPG/JPEG without resizing
          temp_file="${file%.*}_compressed.$ext"
          magick convert "$file" -sampling-factor 4:2:0 -strip -quality 85% "$temp_file" && \
          rm "$file" && \
          mv "$temp_file" "$file"
        fi
      fi
    fi
  done
