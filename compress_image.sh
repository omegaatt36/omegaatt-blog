find . -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" \) -print0 |
  while IFS= read -r -d $'\0' file; do
    if [[ $(stat -c "%s" "$file") -gt 1048576 ]]; then
      echo "Compressing $file"
      ext="${file##*.}"
      width=$(identify -format "%w" "$file")
      height=$(identify -format "%h" "$file")
      if [[ "$width" -gt 2000 || "$height" -gt 2000 ]]; then
         if [[ "$ext" == "png" ]]; then
          magick convert "$file" -resize 50%  "$file"
         else
          magick convert "$file" -resize 50% -sampling-factor 4:2:0 -strip -quality 85% "${file%.*}_compressed.""$ext" && rm "$file" && mv "${file%.*}_compressed.""$ext" "$file"
         fi
      else
        if [[ "$ext" == "png" ]]; then
          optipng -o7 "$file"
        else
          magick convert "$file" -sampling-factor 4:2:0 -strip -quality 85% "${file%.*}_compressed.""$ext" && rm "$file" && mv "${file%.*}_compressed.""$ext" "$file"
        fi
      fi
    fi
  done
