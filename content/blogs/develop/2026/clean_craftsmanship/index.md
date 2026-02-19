---
title: Clean Craftsmanship：在 LLM 時代重拾軟體工匠精神
date: 2026-02-19
categories:
  - develop
tags:
  - reading
  - golang
cover:
  image: "images/cover.png"
---

## 為什麼在這個時間點讀這本書

最近對於如何在 LLM 時代帶領團隊一起提昇生產力感到困惑。當 AI 能在幾秒鐘內生成幾百行程式碼，「寫程式」這件事的門檻似乎降到了前所未有的低點。但門檻低了，品質呢？

當「每個人」都在養龍蝦，有了生產力焦慮，TypeScript 寫的 [OpenClaw](https://github.com/openclaw/openclaw) 才剛出現，Python 的 [NanoBot](https://github.com/HKUDS/nanobot) 馬上跟上，接著是 Golang 的 [picoclaw](https://github.com/sipeed/picoclaw)，然後 Rust 的 [ZeroClaw](https://github.com/zeroclaw-labs/zeroclaw) 也來了。

根據 ZeroClaw 做的 [benchmark](https://github.com/zeroclaw-labs/zeroclaw?tab=readme-ov-file#benchmark-snapshot-zeroclaw-vs-openclaw-reproducible)，這些 Agent 已經降到小於 5MB 的記憶體佔用與 10ms 的啟動時間，「種族」為人類的我們對於產出軟體來說還剩什麼？

帶著這個問題，我決定先從自身出發，找出在 LLM 時代還能保持軟體工程「工匠精神」的誘因。於是翻開了 Robert C. Martin 的 [*Clean Craftsmanship*](https://www.tenlong.com.tw/products/9786263339941)。 

這本書分為三個部份：**紀律、標準、道德**。一半以上的篇幅在講述 TDD 這個老生常談的開發方式，但透過 TDD，我們更能知道何謂軟體的「品質」。以下是我特別書籤的幾個段落。

## 紀律

### 童子軍規則

> 離開營地時，要比你來時更乾淨。

這也是在 *Clean Code* 一書就提到的概念。每次微小的重構，都能小程度的減少技術債的產生。不需要一次大刀闊斧，只要每次經過一段程式碼時，順手讓它變得更好一點。

讓我想到 [Claude is not a senior engineer (yet)](https://www.approachwithalacrity.com/claude-ne/) 這篇文章中提到的 Sweeks——一位被稱為「園丁」的 distinguished engineer，他不斷地重寫、收緊抽象，讓經過他手的程式碼都變得更乾淨。我們都想成為 Sweeks，對吧？

在 LLM 時代，AI 擅長的是「組裝」現有解決方案，但它缺乏 Sweeks 那種「看到可以更好的地方就會忍不住動手」的靈魂。童子軍規則提醒我們：這份靈魂不能丟。

### Test Doubles 的正名

書中透過實戰的例子講述了所有 test double：**Dummy、Stub、Spy、Mock、Fake**。

坦白說，我曾經在諸多 repo 中看到這些名詞卻沒有實際使用它們。頂多在 DI 時製作了一個「用於模擬 database repository 的 implement」，或是使用了 `gomock` 這種套件來產生 mock，然後把所有替身都統稱為 mock。

文中詳盡的敘述讓我知道它們各自的用途與邊界：

- **Dummy**：只是為了填滿參數列表，從不被使用
- **Stub**：始終回傳單一固定值。例如 `PromiscuousAuthenticator` 永遠回傳 true、`RejectingAuthenticator` 永遠回傳 false——你需要兩個不同的 struct 來測不同情境
- **Spy**：可編程的 Stub。透過 `setResult(bool)` 之類的輔助函數，用一個 struct 就能切換回傳值，不必為每種情境各寫一個替身
- **Mock**：會驗證行為的替身，它知道自己預期被怎麼呼叫，不符合預期時主動失敗
- **Fake**：有簡化但可運作的實作，例如 in-memory database

知道這些區分後，未來在 Go 專案中就能更精準地選擇替身，而不是什麼都丟給 `gomock` 了事。

### 持續重構、果斷重構

不必害怕程式碼的改變，勇敢修改——這正是「軟體」software 中 **soft** 的好處。如果軟體不能被輕易改變，那它就不配叫軟體，只能叫硬體。

書中也提到一個我很認同的心態：留條出路。如果測試、重構已經到了死胡同，勇敢地 `git reset --hard` 重新來過。沒有什麼比在錯誤的方向上越走越遠更浪費時間的了。

這個觀念在實務中很受用。我們常常會因為沉沒成本而不願意放棄已經寫了一半的重構，但有時候，承認走錯路、砍掉重練，反而是最有效率的做法。

## 標準

### YAGNI

> You Aren't Gonna Need It.

可以在每次想要引入新的元件時好好想想「如果你不再需要它，會怎樣？」

例如我不喜歡遇到非同步操作就加上 queue 來保障一致性。即便這是在面試的系統設計時能令面試官「為之一亮」的名詞，但卻引入了複雜度——無論是架構上或是維護上。緊接著就是 queue 滿了該怎麼處理、dead letter queue 的策略、消費者掛掉後的重試邏輯⋯⋯每一個「為之一亮」背後都藏著一坨維護成本。

YAGNI 不是叫你不做設計，而是叫你誠實面對當下的需求。

### 永遠不交付 SHIT

這點在 LLM 時代特別有感。當 vibe programming 興起，我們能在極短時間內產出「能跑」的程式碼，但我們很難跟同事說明這行會 work 的 code 其實不怎麼「好看」。

書中列舉了什麼是 SHIT，每一條都值得反覆檢視：

- 你寫出的每個缺陷都是 SHIT
- 你沒測試過的每個函數都是 SHIT
- 你沒有好好寫的每個函數都是 SHIT
- 對細節的每個依賴都是 SHIT
- 每個不必要的耦合都是 SHIT
- 在 GUI 裡出現的 SQL 語句是 SHIT
- 業務規則裡出現的資料庫 Schema 是 SHIT

（雖然 mobile 端使用 SQLite 來處理狀態早已是業界生態，不必把每句話都當成聖旨。）

但這份清單的核心精神是：**你對交出去的東西有沒有感到驕傲？** 如果沒有，那就是 SHIT。在 LLM 能幫你秒出程式碼的年代，這個問題比以往更值得問自己。

### 我們不把問題留給 QA

> 我們希望 QA 什麼問題都不會發現。

這句話的重點不是讓 QA 失業，而是對自己交付品質的要求。我在之前的 [Agile Testing 閱讀筆記](/blogs/develop/2024/agile_testing/) 中也提過類似的觀點：品質是團隊共同的責任，不是丟給測試人員的最後防線。

如果你的程式碼需要 QA 來「發現」問題，那表示你在開發階段就已經失職了。

## 道德

### 教導

> 我們希望所有軟體工程師都成為導師，我們希望你能幫助他人學習。

這也是我仍然在寫部落格的原因。

隨著在團隊中開始承擔 mentor 的角色，愈來愈體會到「教學相長」不是客套話。每一次試圖向別人解釋一個概念，都是在重新檢驗自己是否真的理解。每一篇文章的撰寫，都是在逼自己把模糊的直覺轉化成清晰的邏輯。

在 LLM 的時代，PR review 的回饋有時會被直接轉化為餵給模型的 token，發出去的 comment 是否能讓對方成長，從我的角度是愈來愈難得到回饋感的。但即便如此，教導的責任不會因為工具的改變而消失。

---

看到了這裡，你也從我，變成了我們。

Ref:
- [童子軍規則](https://www.in-com.com/zh-TW/blog/the-boy-scout-rule-the-secret-to-effortless-refactoring-that-scales/)
- [Claude is not a senior engineer (yet)](https://www.approachwithalacrity.com/claude-ne/)
