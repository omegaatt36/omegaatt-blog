---
title: 使用 Bitwarden 與自架後端 Vaultwarden 來管理密碼
date: 2023-11-12
categories:
 - develop
tags:
 - linux
 - self-hosted
---

在尋覓有哪些 self-hosted 專案好玩時，偶然發現了 1password、LastPass 的開源替代方案，甚至後端資料庫能自架，決定架來用用看。

# 使用 Bitwarden 來管理密碼

![Bitwarden_logo](https://upload.wikimedia.org/wikipedia/commons/thumb/c/cc/Bitwarden_logo.svg/1200px-Bitwarden_logo.svg.png)

Bitwarden 是一款流行且功能強大的密碼管理工具，它提供了一個安全的方法來存儲和管理所有密碼。作為一個開源產品，Bitwarden 允許用戶選擇自行托管其服務，這意味著用戶可以在自己的服務器上運行 Bitwarden，從而更好地控制自己的數據安全。

## Bitwarden 的特點

- **安全性**: Bitwarden 使用端到端加密，確保只有您可以訪問您的密碼。
- **跨平台支持**: 支持 Windows、macOS、Linux、Android 和 iOS。
- **易於使用**: 提供直觀的用戶界面和簡單的操作流程。
- **開源**: 開源，增加了透明度和安全性。

Bitwarden 同時[支援基於時間的一次性密碼](https://bitwarden.com/help/authenticator-keys/)，讓 TOTP 也能自動填入。

[Bitwarden 開源了 client 與 server](https://github.com/bitwarden)，在 server 端的選擇有以下：

1. 使用 Bitwarden 提供的官方服務，又分為免費跟付費，但這個選擇就跟 1p 沒太多區別。
2. 自架 Bitwarden 提供的 [open source server](https://github.com/bitwarden/server)，由於是使用 C# 與 mssql，吃的記憶體著實太多。
3. 自架 Bitwarden 相容的後端，我採用的是 rust 實做的 [Vaultwarden](https://github.com/dani-garcia/vaultwarden)，搭配 sqlite，記憶體使用量與官方的 C# 不是一個量級的。

## 使用 Vaultwarden 作為自托管後端

[Vaultwarden](https://github.com/dani-garcia/vaultwarden) 是一個 Bitwarden 的非官方後端實現，它使用 Rust 編寫，更輕量且易於部署。使用 Vaultwarden，您可以在自己的服務器上部署 Bitwarden，這樣您就可以完全控制您的密碼數據。

### 步驟一：Vaultwarden 服務器

可以參考 [Vaultwarden 的 wiki](https://github.com/dani-garcia/vaultwarden/wiki/Deployment-examples)，官方有現成的 docker image。
愛折騰的我還是選擇將他包進 [helm chart](https://github.com/omegaatt36/lab/tree/main/k8s/vaultwarden) 內部署到 k8s cluster 內方便管理。
vaultwarden 的 wiki 頁十分完整，smtp 等設定都是透過環境變數控制。

### 步驟二：Bitwarden 客戶端

到 Bitwarden 上查看所有[客戶端](https://bitwarden.com/download/)

我自己是只在瀏覽器、手機上安裝 Bitwarden 客戶端：

- [chrome extension](https://chromewebstore.google.com/detail/bitwarden-free-password-m/nngceckbapebfimnlniiiahkandclblb)
- [firefox addon](https://addons.mozilla.org/zh-TW/firefox/addon/bitwarden-password-manager/)
- [Android](https://play.google.com/store/apps/details?id=com.x8bit.bitwarden)

需要設定它連接到自己的 Vaultwarden 服務器。

1. 在 Bitwarden 客戶端中，這邊以 firefox extension 為例，點擊「正在登入到」，並選擇 self-hosted。
    ![self-hosted](/assets/dev/20231112/Screenshot_20231112_125439.webp)
2. 在 "服務器 URL" 中輸入 Vaultwarden 服務器地址。
3. 登錄或創建一個新的帳戶。

### 步驟三：開始使用

設定完成後，就可以開始使用來管理密碼了，所有的密碼都會儲存在自己的服務器上。

## 結論

使用 Bitwarden 和 Vaultwarden 為自己提供一個安全、可靠且完全控制的密碼管理方案。透過這種方法，不僅擁有了強大的密碼管理工具，還確保了數據的私密性和安全性。

由於此 server 是公開的，需要自行處理 fail2ban 與 DDOS 攻擊與備份管理，若此站被攻破，也只能怪自己沒有做好完整的資安防護。
