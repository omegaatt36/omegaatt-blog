---
title: 使用 go work 在本地開發解決同時開發 module 的問題
date: 2025-02-20
categories:
  - develop
tags:
  - golang
---

在 [Golang 1.18 中 go workspace 的提案](https://go.dev/blog/get-familiar-with-workspaces)釋出後，golang 的官方文件或多或少也提到應該要怎麼做 multi module 的開發。相較於過去需要不斷的替換 `go.mod` 內的 `replace` 指令，`go work` 大幅改善了 multi module 的開發體驗。

## 為什麼需要 go work

### 專案逐漸變大

當你在維護一個小工具 side project 時，單一的 module 就能夠滿足所有需求，但當專案逐漸變大，會需要將專案拆分成多個 module。
可以透過一個例子來理解：
假設我們有一個專案，拆分成以下兩個 module：

- `common-lib-golang`: 存放所有專案都會用到的 function，例如 retry, logger, tracer 等等
- `backend`: 實際提供 http api 的程式碼

最開始的檔案架構為：

```text
.
├── backend
│   ├── go.mod
│   ├── go.sum
│   └── main.go
└── common-lib-golang
    ├── go.mod
    ├── go.sum
    └── util.go
```

隨著專案變大，我們在開發 `backend` 時，時常會需要修改 `common-lib-golang` 內的程式碼，這時候就需要 `go work` 來協助我們。

### 更好的本地開發體驗

有了這些 module 後，本地測試會變得麻煩，需要一個個 cd 進去跑測試，如果可以有一個檔案來統一管理這些 module，就可以更方便的進行測試。

## go work 做了什麼

### 跨 module 開發

```go
module github.com/omegaatt36/backend

go 1.23.0

require (
    github.com/omegaatt36/common-lib-golang v1.0.0
)

// go.mod
```

在沒有 `go.work` 的情況下，會需要透過 `replace` 來指向本地的 module，但這會導致上版控時發生問題，因為團隊成員不一定會將程式碼放在同一個路徑下。

```go
module github.com/omegaatt36/backend

go 1.23.0

require (
    github.com/omegaatt36/common-lib-golang v1.0.0
)

replace github.com/omegaatt36/common-lib-golang => ../common-lib-golang

// go.mod
```

此時可以使用 `go work` 來解決這個問題，在專案根目錄下建立一個 `go.work` 檔案：

```go
go 1.23.0

use (
    ./backend
    ./common-lib-golang
)
```

或是直接在跟目錄使用

```shell
go work init backend common-lib-golang
go work sync
```

接著執行 `go mod tidy`，就可以將這些 module 加入到 `go.work` 中，並且在程式碼中直接 import 本地的 module，同時不需要在 `go.mod` 中加入 `replace` 指令。

### 避免 dependency 地獄

過去如果沒有使用 container 或是 vm，而是直接在本機開發，很容易發生 dependency 地獄，但 go work 本身並不能解決 dependency 地獄的問題，他主要是可以讓我們在開發時，更方便的切換不同的 module。

## go work 的侷限

### 僅限開發環境

如同[官方文件](https://go.dev/ref/mod#workspaces)描述，`go work` 的目的僅僅是用於本地開發，在 production 環境下，仍然需要使用 `go.mod` 來管理依賴。

### 增加複雜度

如果專案過於龐大，module 數量過多，會導致 `go.work` 檔案過於龐大，難以管理，也可能會有其他潛在的問題。

例如若是根目錄內有時個 repo，僅使用 go work 來引入兩個專案的時候，在其他目錄使用 `go build xxx` 等等的會發現某些 package 的缺失，便是由於 go work 強制複寫了依賴。

需要刪除 `go.work` 與 `go.work.sum`，一來一回又增加了 DX 上的中斷點。
