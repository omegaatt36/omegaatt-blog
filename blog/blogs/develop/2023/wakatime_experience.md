---
title: 透過 WakaTime 幫助來紀錄自己做了哪些事，並製作獨特的 github profile
date: 2023-08-03
categories:
 - develop
tags:
 - github
---

## WakaTime 介紹

[WakaTime](https://wakatime.com) 是一款紀錄自己生產力的工具，透過客戶端插件、集成工具紀錄「行徑」並發送到官方｜非官方伺服器，可以分析花了多久時間在哪個專案、哪個程式語言、會議、code review。
![](/assets/dev/20230803/wakatime_dashboard_official-1.png)

記憶力不佳，過去常常無法想起某段時間做了哪些專案，甚至 daily standup 前忘記昨天做了什麼，發現了 wakatime 可以透過 vscode 插件、terminal 插件來查看自己在哪些時間變更了哪些專案、檔案，甚至可以紀錄下過得指令（僅 binary 的部份，不含參數不會洩漏資料）。

在使用 WakaTime 前，先到 [WakaTime 官網](https://wakatime.com)上註冊並登入，取得最重要的 API Key(API Token)：

![](/assets/dev/20230803/wakatime_dashboard_official-2.png)

安裝相應的客戶端插件。[wakatime 支援的插件、編輯器](https://wakatime.com/plugins)包括但不限於：
- Visual Studio Code
- Vim
- Excel
- Terminal
    - zsh
    - bash
    - fish
    - iTerm2

舉例來說我們可以安裝 [vscode 插件](https://marketplace.visualstudio.com/items?itemName=WakaTime.vscode-wakatime)，並跟著 Installation 輸入完 API Key 就可以開始寫點東西了。

## 查看報告

使用了一段時間後，可以回到官網查看 [dashboard](https://wakatime.com/dashboard)

查看自己上週、前兩週、前一個月每天花了（浪費）了多少時間，以前三十天的圖為例：
![](/assets/dev/20230803/wakatime_dashboard_official-3.png)

或是 YoY 的活動圖
![](/assets/dev/20230803/wakatime_dashboard_official-6.png)

以此 blog 的 repo 來作為舉例，可以查看某個 project 的細節

在該專案花了多少時間、什麼語言：
![](/assets/dev/20230803/wakatime_dashboard_official-4.png)

檔案、分支的時間分配：
![](/assets/dev/20230803/wakatime_dashboard_official-5.png)

## 更新你的 github profile

TODO

## 費用

免費版本的功能已經十分完整了，[完整費用方案在此](https://wakatime.com/pricing?utm_source=magic-panda-engineer)。
透過學生帳號可以申請教育折扣，年度付費 Premium 方案的話一年只要 59 美元，就可以享有完整的 WakaTime 服務。
雖然付費方案的功能幾乎都可以透過免費版本就有的 API 來完成，但仍可以花點小錢支持團隊，或是開啟公司計畫。


