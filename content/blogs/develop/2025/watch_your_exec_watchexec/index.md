---
title: "監控你的執行檔：初探 watchexec"
date: 2025-09-06
categories:
  - develop
tags:
  - golang
  - tui
  - bubbletea
cover:
  image: "images/cover.jpeg"
summary: "前端開發有 liveserver，後端開發有 air，那 TUI 開發呢？本文記錄了我在開發 Bubbletea 應用時，從 air 轉向 watchexec 的心路歷程，以及如何使用這個通用工具來優雅地實現終端機應用的熱重載。"
---

## 前言

身為一個開發者，我們總是在追求更高的效率，而「熱重載」（Hot Reloading）無疑是現代開發流程中不可或缺的一環。

當我們在開發前端應用時，`liveserver` 這類工具能讓我們在存檔的瞬間，立即在瀏覽器上看到變更，省去了手動刷新的繁瑣。將目光轉向後端，特別是在我熟悉的 Golang 生態中，[air-verse/air](https://github.com/air-verse/air) 則是我過去相當依賴的好夥伴。它能監控整個專案的檔案變動，並在需要時自動重新編譯和啟動服務，對於開發 Web API 來說，體驗非常流暢。

然而，當我最近一頭栽進 [Bubbletea](/blogs/develop/2025/golang_bubbletea_experience) 的世界，試圖打造互動式終端機應用（TUI）時，卻發現 `air` 似乎有些力不從心。

## TUI 開發的挑戰：當 `air` 不再香

將 `air` 設定好，準備在我的 Bubbletea 專案中使用時，問題出現了。每當我修改程式碼，`air` 確實偵測到了變動並嘗試重啟應用，但終端機的畫面卻只會停在 `running...`，並不會進到 tui。

```
❯ air --build.cmd "go build -o build/clish cmd/clish/main.go" --build.bin "./build/clish"

  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.62.0, built with Go go1.25.0

watching .
watching api
watching repository/k8s
watching service
!exclude tmp
building...
running...
```

起初我以為是自己的程式碼有問題，但在一番折騰後，我在 Bubbletea 的 GitHub Issues 中找到了答案。在 [issue #150](https://github.com/charmbracelet/bubbletea/issues/150) 中，許多開發者都回報了類似的問題。根本原因在於，Bubbletea 這類 TUI 應用會完全接管終端機的控制權（TTY），包括其輸入模式和畫面渲染。而 `air` 這類專為 Web 服務設計的工具，在重啟進程時，可能沒有正確地處理終端機狀態的清理與恢復，導致新舊進程的終端機控制發生衝突。

這讓我意識到，我需要一個更通用、更底層的檔案監控與進程管理工具。

## 柳暗花明：watchexec

在尋找替代方案的過程中，[watchexec/watchexec](https://github.com/watchexec/watchexec) 進入了我的視野。它是一個用 Rust 編寫的通用檔案監控工具，其設計理念非常純粹：**監控檔案變動，然後執行你指定的任何命令**。

它不關心你正在開發的是什麼語言、什麼框架，無論是 Go、Rust、Python 腳本，還是前端專案的建構命令，它都能勝任。這種語言無關的通用性，以及對進程執行的直接控制，恰好解決了 TUI 開發中遇到的終端機狀態問題。

## 如何使用 watchexec

`watchexec` 的使用非常直觀，主要透過命令列參數進行設定。

### 安裝

在 macOS 上，可以透過 Homebrew 輕鬆安裝：

```bash
brew install watchexec
```

其他平台可以參考[官方文件](https://watchexec.github.io/downloads.html)的安裝方式。

### 基本用法

最簡單的用法是，讓它在任何檔案變動時，重新執行 `go run ./cmd/clish/main.go`：

```bash
# -r: --restart 的縮寫，表示檔案變動時重啟命令
# --: 分隔 watchexec 的參數與要執行的命令
watchexec -r -- go run ./cmd/clish/main.go
```

### 進階用法

為了更精準地控制，我們可以指定要監控的檔案類型或目錄：

```bash
# -w .: 監控當前目錄及其子目錄
# -e go,mod: 只監控 .go 和 .mod 檔案的變動
# -c: 清空螢幕
watchexec -w . -e go,mod -c -- go run ./cmd/clish/main.go
```

這個命令會：
1. 監控 (`-w`) 當前目錄 (`.`)。
2. 只關心 (`-e`) `.go` 和 `.mod` 檔案的變動。
3. 每次重啟前清空 (`-c`) 螢幕。
4. 執行 (`--`) `go run ./cmd/clish/main.go` 命令。

現在，每當我修改 Bubbletea 應用的 Go 原始碼，`watchexec` 都會乾淨俐落地終止舊進程、清空螢幕，然後啟動一個全新的進程。終端機的渲染再也沒有出現過問題，開發體驗如絲般滑順。

## `watchexec` vs. `air`

| 特性 | watchexec | air |
| :--- | :--- | :--- |
| 定位 | 通用型檔案監控與任務執行器 | 專為 Golang 熱重載設計 |
| 語言 | Rust | Golang |
| 設定方式 | 命令列參數 | `.air.toml` 設定檔 |
| 適用場景 | 任何需要檔案監控的場景，特別適合 TUI、腳本 | Golang Web 服務、API 開發 |
| 優點 | 極度靈活、輕量、語言無關 | 功能豐富、開箱即用、為 Go Web 開發優化 |
| 缺點 | 功能相對單純，需自行組合命令 | 對非 Web 服務的 TUI 應用支援不佳 |

## 寫在最後

這次從 `air` 轉換到 `watchexec` 的經驗，讓我再次體會到「沒有銀彈，只有最適合的工具」這個道理。`air` 在它擅長的領域（Go Web 開發）依然是一個非常優秀的工具，但當場景轉換到 TUI 開發時，更通用、更底層的 `watchexec` 則展現出了它的價值。

選擇工具，如同選擇武器，關鍵在於理解問題的本質。對於需要精準控制進程與終端機環境的 TUI 開發來說，`watchexec` 無疑是我目前找到的最佳拍檔。
