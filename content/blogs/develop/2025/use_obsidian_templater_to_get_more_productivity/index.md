---
title: 使用 Templater 來提昇 Obsidian 的生產力
date: 2025-01-02
categories:
  - develop
tags:
  - software
  - obsidian
  - productivity
---

> 這篇文章不涉及 [Obsidian](https://obsidian.md/) 完整使用說明以及 [Templater](https://silentvoid13.github.io/Templater/) 完整使用教學，僅以一個 use case 來展現如何使用 Templater 來提昇 Obsidian 的使用效率。

## 前言：

工作日誌，是一個甜蜜的負擔，有些人是因為公司要求，開始寫工作日誌，逐漸流於形式。

最初我也是因為工作需要「彙報」，才開始寫工作日誌。直到開始跑 scrum，開始明白 standup meeting 的意義，開始在意如何讓 stackholder 放心把事情交給我做，於是

參考了《[SCRUM：用一半的時間做兩倍的事](https://www.books.com.tw/products/0010785434)》以及實際執行的經驗，以下是我認為良好的彙報需要注意的關鍵因素：

- 昨天的進度：簡要說明昨天完成的具體工作，重點在成果而非過程。
- 今天的計劃：說明今天的工作重點，聚焦在需要達成的具體目標。如果是要跨超過一天的工作，也需要設定一個預計能完成的時間。
- 遇到的障礙：明確指出當前的問題或挑戰，並說明需要的幫助。當然也沒有必要什麼事都等到 standup meeting 才提出問題，有障礙應當即時反應。
- 與其他團隊或成員的依賴：說明需要其他人或團隊支持的地方，確保協作順暢，例如後端需要寄送的 email 內容，需要前端協助產生 html 等等。
- 以終為始、以成果為導向：聚焦於交付成果的進展，而非細節，例如：目前已完成 80% 的登入模組開發，預計明天可以驗收。

詳細的舉例不是這篇文章的重點，有空可以再寫成其他文章 D:

上述幾個要點，在工作日誌中被我濃縮成 markdown 的格式：

```
XXXX-XX-XX

昨日：
- 昨日1
- 昨日2
瓶頸：
- 瓶頸
今日：
- 今日1
- 今日2
```

## Obsidian

高中時接觸了 markdown，大學接觸了 Hackpad，畢業後延續這個習慣，使用 Hackmd 來管理自己的工作日誌。

隨著時間的增長，開始發現 Hackmd 無論是官方 hosting 的免費版與企業版，或是 self-hosted 的 CodeMD，在單一大檔案編輯文件時，都會有明顯的延遲。若我將工作日誌拆成多個月份，又不好管理。

在 2023 年轉換工作之際，開始尋找其他 markdown 文字編輯器。

- [Notion](https://www.notion.com/)

  從 2020 年開始接觸 notion，後來無 block 限制，加上原本就是 SaaS，提供了同步、共同編輯等機制，是一個非常適合拿來紀錄工作日誌的編輯器。

  短版也另我頭痛不已：

    1. web-based，即使有桌面版本，也是一個 electron 的網頁應用程式，實用時間一長，操作延遲大幅度上升。
    2. 過於混亂的介面，在手機版上特別明顯。
    3. 無離線模式，我一定要找到網路，才能對我的筆記進行存取

- [VScode](https://code.visualstudio.com/)

  雖然說是 VScode，但這邊均指任何 code editor，要用 zed or vim or nano 或是 geditor 任均挑選。完全存在本機，同步的方式我是使用 git 來管理。

  也是有短版存在：

    1. 缺少 Mobile 應用，應該不會有人用手機還在開 termux 來編輯文件吧...
    2. 無法即時的進行共同編輯
    3. 需要熟悉 git

還有其他優秀的軟體，因為文章篇幅，不具體展開：

- Google Docs
- [Anytype](https://anytype.io/)
- [logseq](https://logseq.com/)
- [Joplin](https://joplinapp.org/)

如果你 Apple 全家桶，大可以使用內建的記事本配合 iCloud 就能解決上述軟體的短版，可惜我不是。

當時的環境背景，我幾乎所有軟體都會優先考慮開源的解決方案，於是看向了 Obsidian 這個社群廣大的...軟體（應用太廣泛似乎有點難被定義成純編輯器）。

雖然他是基於 electron 開發的，卻沒有明顯能被拿來抱怨的效能問題，最吸引人的是豐富的 plugins，讓他可以變形並取代不少 SaaS。

同步的方式，我是先使用 [obsidian-livesync](https://github.com/vrtmrz/obsidian-livesync) 再轉用 [syncthings](https://syncthing.net/)。

## 問題：

落落長介紹了為何使用 Obsidian，文章終於來到這邊部落格的出發點，生產力。

效率就像一個詛咒，只要嚐到一次甜頭，會想要改變生活中每一個能被提昇的節點（動素）。

有不少 Youtuber 透過 Obsidian 來分享他們「第二大腦」的實踐過程，內容涵蓋這篇文章的百倍萬倍，以下僅僅是我的一個小小效率提昇

每個工作日，上工前，我會打開 Obsidian 來記錄工作日誌。

我會為每個月建立一個獨立的 note，並使用 H2 標題來區分每一天，例如：

```markdown
## 2025-01-02

昨日：
- 完成承暖步道
- 完成週末花蓮三天兩夜行程規劃
瓶頸：
今日：
- 完成 Obsidian 與 Templater 的介紹部落格
- 查看 2024 Recap 在 GA4 上的表現
```

雖然這種方式可以有效地記錄工作進度，但每次都要手動輸入這些重複的格式，實在是太浪費時間了。這也是我開始尋找解決方案的原因。

## 解決方案：

Templater 是一個 Obsidian 的模板插件，它可以讓使用者透過模板的方式快速插入預先定義好的內容。

透過 Templater，我們可以輕鬆地建立各種模板，並透過快捷鍵快速插入。

### 建立模板

首先，在 valut 中建立用於存放模板的資料夾，我是直接在根目錄中新增 `Templates`

接著需要在 Obsidian 的社群插件中搜尋並安裝 Templater 插件，可以參考[官方文件](https://silentvoid13.github.io/Templater/installation.html)。

安裝完畢後，在 Templater 的設定中指定模板資料夾為 `Templates`

在 Templates 中，建立一個新的 note，將其命名為 `Daily Work`（或其他你喜歡的名字），並將以下內容貼入：

```
## <% tp.date.now("YYYY-MM-DD") %>
昨日：
-
瓶頸：
-
今日：
-
```

[`tp.date.now("YYYY-MM-DD")`](https://silentvoid13.github.io/Templater/internal-functions/internal-modules/date-module.html) 可以取得當前日期並格式化為 YYYY-MM-DD。

### 使用模板

1. 在 Obsidian 的任意文件中按下 `Ctrl+P` 或 `Cmd+P` 打開命令面板
2. 輸入 `Daily Work` 應該會看到兩個選項分別為：
  - `Templater: Insert Templates/Daily Work.md`
  - `Templater: Create Templates/Daily Work.md`
3. 由於我們是在文件中插入區段，於是選擇 `Insert`
4. 檢查成果

  ```markdown
  ## 2025-01-02

  昨日：
  -
  瓶頸：
  -
  今日：
  -
  ```

### 修改模板

在我們的團隊中，每週一會進行當週的計畫會議，因此我需要在每週一的日誌中，加入「上週」和「本週」的總結。

將 `Daily Work` 的內容修改為以下內容：

```
## <% tp.date.now("YYYY-MM-DD") %>
  <%* let monday = tp.date.weekday("YYYY-MM-DD", 0)
if ( tp.date.now("YYYY-MM-DD") === monday ) { %>
上週：
-
本週：
-  <%* } else { %>
昨日：
- <%* } %>
瓶頸：
-
今日：
-
```

- `tp.date.weekday("YYYY-MM-DD", 0)`:  取得當週星期一的日期
- `if ( tp.date.now("YYYY-MM-DD") === monday ) { ... } else { ... }`:  判斷今天是否為星期一，如果是則顯示「上週」和「本週」，否則顯示「昨日」。

例如：

```mardkown
## 2024-12-30

上週：
-
本週：
-
瓶頸：
-
今日：
-
```

### 進階應用

除了每日工作日誌，Templater 還有許多其他應用，例如：

  - 快速插入會議記錄
  - 建立專案進度模板
  - 自動插入檔案資訊
  - 綁訂快捷鍵，不再需要使用 `Ctrl+P`

## 總結

Templater 作為一個強大的 Obsidian 插件，可以幫助我們提高工作效率。透過建立客製化的模板，我們可以快速插入重複性的內容，專注於真正重要的任務。
