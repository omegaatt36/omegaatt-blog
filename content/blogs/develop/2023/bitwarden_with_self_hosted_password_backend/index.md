---
title: 使用 Bitwarden 與自架後端 Vaultwarden 來管理密碼與 2FA Authenticator
date: 2024-09-01
categories:
 - develop
tags:
 - linux
 - self-hosted
---

在尋覓有哪些 self-hosted 專案好玩時，偶然發現了 1password、LastPass 的開源替代方案，甚至後端資料庫能自架，決定架來用用看。

# 使用 Bitwarden 來管理密碼

[![Bitwarden_logo](https://upload.wikimedia.org/wikipedia/commons/thumb/c/cc/Bitwarden_logo.svg/1200px-Bitwarden_logo.svg.png)](https://bitwarden.com/products/personal/)

Bitwarden 是一款流行且功能強大的密碼管理工具，它提供了一個安全的方法來存儲和管理所有密碼。作為一個開源產品，Bitwarden 允許用戶選擇自行托管其服務，這意味著用戶可以在自己的服務器上運行 Bitwarden，從而更好地控制自己的數據安全。

## Bitwarden 的特點

![bitwarden-demo](images/Screenshot_20240901_101852.png)

- **安全性**: Bitwarden 使用端到端加密，確保只有您可以訪問您的密碼。
- **跨平台支持**: 支持 Windows、macOS、Linux、Android 和 iOS。
- **易於使用**: 提供直觀的用戶界面和簡單的操作流程。
- **開源**: 開源，增加了透明度和安全性。

Bitwarden 同時[支援基於時間的一次性密碼](https://bitwarden.com/help/authenticator-keys/)，讓 TOTP 也能自動填入。也能透過 [github.com/scito/extract_otp_secrets](https://github.com/scito/extract_otp_secrets) 來提取 Google Authenticator 內的 2FA 資訊，儲存進 Bitwarden 中。

[Bitwarden 開源了 client 與 server](https://github.com/bitwarden)，在 server 端的選擇有以下：

1. 使用 Bitwarden 提供的官方服務，又分為免費跟付費，但這個選擇就跟 1p 沒太多區別。
2. 自架 Bitwarden 提供的 [open source server](https://github.com/bitwarden/server)，由於是使用 C# 與 mssql，吃的記憶體著實太多。
3. 自架 Bitwarden 相容的後端，我採用的是 rust 實做的 [Vaultwarden](https://github.com/dani-garcia/vaultwarden)，搭配 sqlite，記憶體使用量與官方的 C# 不是一個量級的。

## Bitwarden 架構

bitwarden 的 local storage 都是儲存加密後的密碼資料，不會使用明碼儲存，故上傳到 server 上的也僅僅是加密後的密碼資料。

### master password

master password 是 bitwarden 的基礎密碼，可以用來加密其他密碼資料，也可以用來登入 bitwarden 的 web 介面。每一個密碼加密均會有 master password 的參與，故 master password 必須選用容易記住、難以攻剋的密碼。可以使用官方的[密碼強度檢查工具](https://bitwarden.com/password-strength/)。

### 儲存密碼

```mermaid
sequenceDiagram
    participant Client
    participant Server

    rect rgb(236,239,244)
    Client ->> Client: 使用 Master Password 生成加密密鑰
    Client ->> Client: 加密密碼資料
    Client ->> Server: 發送新增密碼請求 (加密後的密碼資料)
    Server ->> Server: 儲存加密後的密碼資料
    Server -->> Client: 回應成功 (確認訊息)
    end
```

### 新的 client 登入

在 bitwarden 內，可以由多個 client 登入，使用同一個 master password 即可登入。

```mermaid
sequenceDiagram
    participant Client
    participant Server

    rect rgb(236,239,244)
    Client ->> Client: 用戶輸入帳號和 Master Password
    Client ->> Server: 發送身份驗證請求 (帳號, 雜湊後的驗證資訊)
    Server ->> Server: 驗證用戶身份
    Server -->> Client: 回傳加密的 vault 資料
    Client ->> Client: 使用 Master Password 生成解密密鑰
    Client ->> Client: 解密加密的 vault 資料
    Client ->> Client: 本地存儲解密後的 vault
    end
```

### 更改 master password

```mermaid
sequenceDiagram
    participant Client
    participant Server

    rect rgb(236,239,244)
    Client ->> Client: 用戶輸入舊的和新的 Master Password
    Client ->> Client: 使用舊 Master Password 解密本地 vault
    Client ->> Client: 使用新 Master Password 生成新加密密鑰
    Client ->> Client: 使用新密鑰重新加密 vault
    Client ->> Server: 上傳重新加密的 vault
    Server ->> Server: 替換舊的加密資料
    Server -->> Client: 確認成功
    end

    rect rgb(236,239,244)
    loop 每個已登入的客戶端
        Client ->> Server: 同步檢查更新
        Server -->> Client: 傳送新加密的 vault
        Client ->> Client: 用戶輸入新的 Master Password
        Client ->> Client: 生成新解密密鑰
        Client ->> Client: 解密並本地更新
    end
    end
```


## Pricing

個人使用的話，若僅需儲存個人密碼與信用卡、身份等純文字資訊，則免費版本已經夠用了，不需要付費。若需要額外的檔案加密、硬體加密、多個組織等等，則需要支付一定的費用。

但如此佛心的開源項目，也鼓勵大家即便自架後端，也仍可以付費訂閱以已支持團隊持續開發與維運，Premium Account 一年 10USD，也並不是一個很大的費用。

![bitwarden-pricing](images/Screenshot_20240901_103441.png)

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
- [iOS](https://apps.apple.com/us/app/bitwarden-password-manager/id1137397744)
- [Linux Flatpak](https://flathub.org/apps/com.bitwarden.desktop)

需要設定它連接到自己的 Vaultwarden 服務器。

1. 在 Bitwarden 客戶端中，這邊以 firefox extension 為例，點擊「正在登入到」，並選擇 self-hosted。
    ![self-hosted](images/Screenshot_20231112_125439.webp)
2. 在 "服務器 URL" 中輸入 Vaultwarden 服務器地址。
3. 登錄或創建一個新的帳戶。

### 步驟三：開始使用

設定完成後，就可以開始使用來管理密碼了，所有的密碼都會儲存在自己的服務器上。在 bitwarden 內唯一可以用來辨識是否已經套用 vaultwarden 的方式，是打開帳戶->進階會員，若是看到已經是升級帳戶，則表達已經套用。

![bitwarden-advanced-member](images/Screenshot_20240901_104005.png)

### 步驟四：備份

#### 通用備份

無論是使用 bitwarden 官方 server 還是自架的 vaultwarden，都可以在應用程式內匯出「明碼」或「加密後」的密碼備份檔案，從檔案->會出密碼庫，並輸入 master password 來解密，即可將備份檔案匯出至硬碟。

![bitwarden-export-backup](images/Screenshot_20240901_104940.png)

```json
{
  "encrypted": false,
  "folders": [],
  "items": [
    {
      "passwordHistory": null,
      "revisionDate": "2023-11-11T07:52:23.560Z",
      "creationDate": "2023-11-11T07:52:23.560Z",
      "deletedDate": null,
      "id": "xxxxxxxx-xxxx-4xxx-xxxx-xxxxxxxxxxxx", // UUID v4
      "organizationId": null,
      "folderId": null,
      "type": 1,
      "reprompt": 0,
      "name": "Google Raiven 55688",
      "notes": null,
      "favorite": false,
      "login": {
        "fido2Credentials": [],
        "uris": [
          {
            "match": null,
            "uri": "https://www.google.com"
          }
        ],
        "username": "raiven55688",
        "password": "raiven55688", // 明碼
        "totp": null
      },
      "collectionIds": null
    }
  ]
}
```

#### vaultwarden backup

Valutwarden 會將密碼儲存在 sqlite 資料庫中，可以使用 sqlite 工具來匯出備份。或是使用 [github.com/ttionya/vaultwarden-backup](https://github.com/ttionya/vaultwarden-backup) 來協助我們自動備份密碼。

會使用 Rclone 來備份，也可以直接使用 docker 來 mount vaultwarden 的資料庫上，可以參考[文件](https://github.com/ttionya/vaultwarden-backup?tab=readme-ov-file#automatic-backups)：

```bash
docker run -d \
  --restart=always \
  --name vaultwarden_backup \
  --volumes-from=vaultwarden \
  --mount type=volume,source=vaultwarden-rclone-data,target=/config/ \
  -e DATA_DIR="/data" \
  ttionya/vaultwarden-backup:latest
```

## 結論

使用 Bitwarden 和 Vaultwarden 為自己提供一個安全、可靠且完全控制的密碼管理方案。透過這種方法，不僅擁有了強大的密碼管理工具，還確保了數據的私密性和安全性。

由於此 server 是公開的，需要自行處理 fail2ban 與 DDOS 攻擊與備份管理，若此站被攻破，也只能怪自己沒有做好完整的資安防護。
