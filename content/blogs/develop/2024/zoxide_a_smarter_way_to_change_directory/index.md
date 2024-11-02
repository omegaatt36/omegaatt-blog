---
title: "A smarter way to change directory: zoxide"
date: 2024-09-24
categories:
  - develop
tags:
  - linux
  - software
  - open_source
---

在日常的開發工作中，我們經常需要在不同的目錄間切換。雖然 `cd` 命令已經足夠好用，但如果有一個更聰明的工具能記住我們最常用的目錄，並讓我們用最少的按鍵就能快速跳轉，那豈不是更棒?

這就是 zoxide 的用武之地。zoxide 是一個由 Rust 編寫的「更聰明的 cd 命令」，靈感來自 z 和 autojump。它會記住你最常使用的目錄，讓你只需輸入幾個字符就能快速跳轉。

## zoxide 的主要特性

- 自動匹配: 不需要輸入完整路徑，zoxide 會根據輸入自動匹配最相關的目錄：
- 自動紀錄過去的目錄: zoxide 會記住你最常使用的目錄，讓你只需輸入幾個字符就能快速跳轉。這些資料可以使用 `zoxide query` 來查詢，或是 `zoxide edit` 來管理。
- 互動式選擇: 結合 fzf，可以互動式地選擇目標目錄
- 輕量快速: 用 Rust 編寫，啟動迅速，幾乎不會影響 shell 的啟動時間

## 安裝與配置

zoxide 的安裝可以參考 [官方文件](https://github.com/ajeetdsouza/zoxide?tab=readme-ov-file#installation)，我們就不多贅述了。

安裝完成後，我們需要在 shell 的配置文件中添加初始化命令。以 zsh 為例，在 `~/.zshrc` 的末尾添加:

```bash
eval "$(zoxide init zsh)"
```

重新打開終端或執行 `source ~/.zshrc` 後，zoxide 就可以使用了。

## 使用方法

zoxide 的基本用法非常直觀，經過自動執行 `zoxide init zsh` 後，我們可以使用 `z` 和 `zi`

- 自動匹配:
  ```bash
  ❯ ls
    boo bar baz
  ❯ z bo
  ❯ pwd
    /home/raiven/boo
  ```
- 即便沒有輸入完整路徑，也能夠自動匹配最相關的目錄
  ```bash
  ❯ pwd
    /home/raiven/dev/omegaatt-blog
  ❯ z
  ❯ pwd
    /home/raiven
  ❯ z blog
  ❯ pwd
    /home/raiven/dev/omegaatt-blog
  ```
- 互動式選擇:
  ```bash
  ❯ zi
  ~ < 98/98(0)
  296.0 /home/raiven/dev/omegaatt-blog
  124.0 /home/raiven/dev/bookly
   50.0 /home/raiven/dev
   24.0 /home/raiven/dev/test
   12.0 /home/raiven/
  ```

這裡的 `z` 命令就像一個更聰明的 `cd`，而 `zi` 則是互動式版本。

## 進階配置

zoxide 還提供了一些進階配置選項，讓你可以根據自己的需求進行調整:

- 通過 `--cmd` 參數修改命令前綴，比如改成 `j` 或直接替換 `cd`
- 使用環境變量來自定義數據存儲位置、排除特定目錄等

## 結語

zoxide 是一個小巧但強大的工具，能極大地提升我們在終端中的工作效率。它智能地學習我們的使用習慣，讓目錄導航變得輕而易舉。如果你經常在終端中工作，不妨試試 zoxide。
