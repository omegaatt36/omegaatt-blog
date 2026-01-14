---
title: Container Image Optimization 那些年我們寫錯的 Dockerfile
date: 2026-01-14
categories:
  - develop
tags:
  - docker
  - optimization
  - devops
---

最近在檢視公司內部的專案時，針對其中一個 container image 進行了優化。在一個簡單的 commit 後，我們的 image size 從 **1.82GB** 修正到了 **1.18GB**。

透過 [dive](https://github.com/wagoodman/dive) 查看，Image Efficiency 更是從不及格的 **69%** 飆升到了 **99%**。

這讓我回想起過去寫 Dockerfile 時，常常因為不了解 Docker Layer 的機制，或是為了寫起來「方便」，而踩到了許多效能與安全的地雷。

## 致命的 `chown`

這次優化的核心，其實源自於一個非常常見的操作：修改檔案權限。

在我們的案例中，Dockerfile 原本是這樣寫的：

```dockerfile
# Bad Practice: recursive chown after copy
FROM ubuntu:22.04

WORKDIR /app
COPY . .

# ... install dependencies ...
RUN dpkg -i packages/*.deb

# Change ownership for security reasons
RUN groupadd -r appuser && useradd -r -g appuser appuser
RUN chown -R appuser:appuser /app
```

看起來邏輯很正確：把檔案複製進去，安裝套件，最後為了安全性將檔案權限交給非 root 使用者。

但在 Docker 的世界裡，`RUN` 指令會產生新的 Layer。`chown -R` 這個指令會遞迴修改目錄底下所有檔案的 metadata。

**即使你沒有修改檔案內容 (file content)，但因為 metadata 變了，Docker 的 OverlayFS 視為新檔案。**

因此，目錄內原本佔用的 369MB 檔案，全部被從上一層 (COPY layer) **複製**了一份到這一層 (RUN chown layer) 中。這就是為什麼我們的 image size 會莫名其妙暴增，且 Efficiency 低落的主因。

### 如何修正？

修正的方法其實很簡單，盡量在產生檔案的當下就決定好權限，或是將操作合併在同一個 Layer 中。

**方法一：使用 `COPY --chown`**

如果是單純複製原始碼，Docker `COPY` 指令原生支援 `--chown` flag：

```dockerfile
# Best Practice: use --chown flag
WORKDIR /app
RUN groupadd -r appuser && useradd -r -g appuser appuser

# Set ownership directly during copy, avoiding duplication in a new layer
COPY --chown=appuser:appuser . .
```

**方法二：Multi-stage Build 搭配解壓縮**

在我們的案例中，因為涉及到 `dpkg -i` 安裝 `.deb` 檔，這些檔案預設會安裝到系統目錄且擁有者為 root。如果在安裝後才執行 `chown`，就會發生上述的 Layer 膨脹問題。

我們採用的進階解法是利用 Multi-stage build，在一個臨時的 Stage 中將 `.deb` 解壓縮 (`dpkg -x`) 並處理好相關設定 (如 symlink)，最後再將處理好的檔案 `COPY --chown` 到最終的 Image 中。

```dockerfile
# Stage 1: Extractor
FROM ubuntu:22.04 AS deb_extractor

COPY packages /tmp/deb
# 使用 dpkg -x 解壓縮，而不直接安裝
RUN mkdir -p /extracted && \
    for deb in /tmp/deb/*.deb; do dpkg -x "$deb" /extracted; done && \
    # 模擬 postinst script 的動作 (例如建立 symlink)
    cd /extracted/usr/lib/ && \
    ln -s libapp.so.1.0.0 libapp.so

# Stage 2: Final Image
FROM ubuntu:22.04

ARG USERNAME=appuser
RUN groupadd -r $USERNAME && useradd -r -g $USERNAME $USERNAME

# 從 extractor stage 複製檔案，並同時修改權限
COPY --from=deb_extractor --chown=$USERNAME:$USERNAME /extracted/usr/lib /usr/lib
COPY --from=deb_extractor --chown=$USERNAME:$USERNAME /extracted/app /app
```

這樣做的好處是：
1. **Zero Wasted Layer**: `dpkg -x` 解壓的過程發生在 `deb_extractor` stage，不會帶入最終 Image。
2. **Correct Ownership**: 透過 `COPY --chown` 一次到位，沒有額外的 `chown` layer。

## 其他常見的 Dockerfile 陷阱

除了 `chown` 之外，還有許多細節會影響 Image 的大小與建構速度。

### 1. `apt-get update` 與 `install` 分家

```dockerfile
# Bad Practice: separating update and install
RUN apt-get update
RUN apt-get install -y python3
```

這會導致 Layer Caching 的問題。如果 `apt-get update` 的 Layer 被 Cache 住了，當你修改下方的 `install` 指令（例如新增一個 package）時，Docker 可能會直接使用舊的 `update` Layer，導致你安裝到舊版本的軟體，甚至是找不到 package (404 Not Found)。

**正確做法：**

```dockerfile
# Best Practice: combine update, install, and cleanup
RUN apt-get update && apt-get install -y \
    python3 \
    && rm -rf /var/lib/apt/lists/*
```

同時，記得在同一個 Layer 清除 apt cache (`/var/lib/apt/lists/*`)，否則這些暫存檔會永遠留在這一層 Layer 中佔用空間。

### 2. 忽略 `.dockerignore`

就像 git 有 `.gitignore`，Docker 也有 `.dockerignore`。

如果執行 `COPY . .` 卻沒有設定 `.dockerignore`，會將 `.git` 目錄、本地的 `node_modules`、測試報告、甚至是敏感的 `.env` 檔案全部複製進 Image。這不僅增加了 Image 大小，更有可能洩漏機敏資訊。

### 3. 無效的 Layer 順序

Docker 的 Cache 機制是基於 Layer 的。一旦某一層發生變化，其後的所有 Layer Cache 都會失效。

```dockerfile
# Bad Practice: copying source code before installing dependencies
COPY . .
RUN npm install
```

只要你的 source code 修改了一個字，`COPY . .` 的 hash 就會改變，導致後面的 `npm install` 全部重跑一遍，浪費了大量的 build time。

**正確做法：**

```dockerfile
# Best Practice: copy dependency definitions first
COPY package.json package-lock.json .
RUN npm install
COPY . .
```

先複製 dependency definition 檔案並安裝，最後才複製 source code。這樣只有在 dependency 改變時，才需要重新執行 install。

## 善用工具

這次能發現問題，歸功於 [dive](https://github.com/wagoodman/dive) 這個工具。它能視覺化每一層 Layer 到底增加了什麼檔案。

```shell
dive <your-image-tag>
```

在 UI 中，你可以清楚看到哪些檔案在不同的 Layer 中被重複複製（會有黃色/紅色的標示），這時就是進行優化的最佳時機。

如果你對於 Image 的 Layer 結構感興趣，甚至想知道如何手動分析這些 Layer，可以參考我之前的文章：[如何不啟動 container 從 image 中提取可執行檔](/blogs/develop/2024/extract_executable_from_docker_image)。

## 總結

撰寫 Dockerfile 雖然容易入門，但要寫出「精簡、安全、好維護」的 Dockerfile 卻需要對底層機制有一定的了解。

- 合併相關的指令。
- 留意 `chown`, `chmod` 等修改 metadata 的操作。
- 將變動最少的指令放在最前面。
- 善用 dive 等分析工具。

下次覺得 Image 太肥大時，不妨用 dive 檢查一下，說不定你也複製了幾百 MB 的隱形檔案。
