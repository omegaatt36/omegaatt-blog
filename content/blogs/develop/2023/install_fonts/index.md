---
title: 如何優雅地批次安裝 Nerd font 字型
date: 2023-05-03
categories:
 - develop
tags:
 - linux
 - wsl
aliases:
 - "/blogs/develop/2023/install_fonts.html"
---

[fonts repo](https://github.com/ryanoasis/nerd-fonts)

接觸 wsl/ubuntu 的這幾年，經常會需要安裝 zsh 以及字型，因此誕生了 [cheat sheet](/blogs/develop/2022/zsh-cheat-sheet)，久而久之連下載->解壓縮->安裝這個過程，都懶了。

於是寫了一個批次下載並安裝最新版本的 Nerd Fond 字型的 shell script，讓 wsl 可以安裝 windows 的字型，linux 可以安裝 linux 字型。

```shell
#!/bin/bash

declare target="linux"
declare repo="ryanoasis/nerd-fonts"
declare -a fonts=(
  BitstreamVeraSansMono
  CodeNewRoman
  DroidSansMono
  FiraCode
  FiraMono
  Go-Mono
  Hack
  Hermit
  JetBrainsMono
  Meslo
  Noto
  Overpass
  ProggyClean
  RobotoMono
  SourceCodePro
  SpaceMono
  Ubuntu
  UbuntuMono
)


ARGS=`getopt -o p --long windows -- "$@"`
if [ $? -ne 0 ]; then
  echo "getopt failed: " $ARGS
  exit 1
fi

eval set -- "${ARGS}"

while true
do
case "$1" in
  -p|--windows)
    target=windows
    shift
    ;;
  --)
    shift
    break
    ;;
esac
done


get_latest_release() {
  echo $(curl --silent "https://api.github.com/repos/$@/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/')
}

function install_linux_fonts() {
  local fonts_dir
  fonts_dir="${HOME}/.local/share/fonts"
  if [[ ! -d "$fonts_dir" ]]; then
    mkdir -p "$fonts_dir"
  fi
  mv $@/*.ttf $fonts_dir/

  # fc-cache -fv
}

# Install font for the current user. It'll appear in "Font settings".
function install_windows_fonts() {
  local dst_dir
  dst_dir=$(wslpath $(cmd.exe /c "echo %LOCALAPPDATA%\Microsoft\\Windows\\Fonts" 2>/dev/null | sed 's/\r$//'))
  mkdir -p "$dst_dir"
  local src
  for src in "$@"; do
    local file=$(basename "$src")
    test -f "$dst_dir/$file" || cp -f "$src" "$dst_dir/"
    local win_path
    win_path=$(wslpath -w "$dst_dir/$file")
    echo $win_path
    reg.exe add                                                      \
      "HKCU\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Fonts" \
      /v "${file%.*} (TrueType)"  /t REG_SZ /d "$win_path" /f 2>/dev/null
  done
}

function main() {
  local version
  version=$( get_latest_release $repo )
  local tmp_dir
  tmp_dir="$(mktemp -d)"

  trap "rm -rf ${tmp_dir@Q}" INT TERM EXIT

  for font in "${fonts[@]}"; do
    zip_file="${font}.zip"
    download_url="https://github.com/$repo/releases/download/${version}/${zip_file}"
    echo "Downloading $download_url"
    wget -q --show-progress -P "$tmp_dir" "$download_url"
    # unzip -o means replace file without asking.
    unzip -o "$tmp_dir/$zip_file" -d "$tmp_dir"
    rm "$tmp_dir/$zip_file"
  done
  find "$tmp_dir" -name '*Windows Compatible*' -delete

  if [ "$target" = "linux" ]
  then
    install_linux_fonts $tmp_dir
  elif [ "$target" = "windows" ]
  then
    install_windows_fonts $tmp_dir/*.ttf
  fi
}

main

echo -e '\033[0;32m'
echo 'Fonts successfully installed.'
echo -e '\033[0m'
```

若沒有要改目標字型也可以直接透過 github 上的檔案來安裝
- wsl:
    ```sh
    bash -c "$(curl -fsSL https://raw.githubusercontent.com/omegaatt36/dotfiles/main/install_fonts.sh)" --windows
    ```
- linux:
    ```sh
    bash -c "$(curl -fsSL https://raw.githubusercontent.com/omegaatt36/dotfiles/main/install_fonts.sh)"
    ```

#### ref
- [install_meslo_wsl](https://gist.githubusercontent.com/romkatv/aa7a70fe656d8b655e3c324eb10f6a8b/raw/install_meslo_wsl.sh)
- [install nerd fonts](https://gist.github.com/matthewjberger/7dd7e079f282f8138a9dc3b045ebefa0)
