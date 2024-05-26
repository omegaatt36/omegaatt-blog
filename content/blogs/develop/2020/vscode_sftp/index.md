---
title: VScode SFTP 快速同步文件到 Server
date: 2020-01-20
categories:
 - develop
tags:
 - linux
 - vscode
aliases:
 - "/blogs/develop/2020/vscode_sftp.html"
---

當 Server 環境使用白名單搭配 Reverce-Proxy 時，或是 dev、test 環境不供外部使用時，或許會將部分 code 存在只能靠內網連線的主機上，並透過 OpenSSH 供內部遠端操作。

而 PHP 快速部屬的便捷在於替換單一檔案不須重啟 Server (前提是並非把檔案載進 Memory 或轉化守護進程)，此時可以依靠 FTP 上傳單一檔案即好。但每次改個 code 還要想辦法上傳程式碼實在是太浪費時間了，如標題所述，如何快速同步 code 到 Server 上呢。

[vscode-sftp](https://github.com/liximomo/vscode-sftp) 是一個在 VScode 中非常方便的擴充套件，上傳/下載/同步、單一/多伺服器、ftp/sftp，僅需設置一份或多份的設定檔便能(偷懶不切視窗)上傳檔案到 Server。然而正是環境 (production) 還是走正常 CI/CD 比較好。

# 安裝

首先在 VScode 中的市集，或是快捷鍵 ctrl+shift+P (或 F1) 打開指令窗口後輸入 extension:install。並搜尋 SFTP，認名作者 liximomo 點一下玩一年，上傳不花一分錢(支持者可以拉到下面點選 donate)。

![](https://i.imgur.com/p0YZBxa.png)

# 設定上傳環境/參數

接著同樣打開指令窗口並輸入 SFTP: Config

![](https://i.imgur.com/ozI524w.png)

**注意由於是 JSON 格式，不能註解**

主要配置好host, port, username, privateKeyPath, remotePath, ignore这参数即可：
![](https://i.imgur.com/m03U2cm.png)

* name : 檔案暱稱
* host : server IP/URL
* port : 若 server 建置好 ssh 環境，可將 port 端口改為 22
* username : 在 server 中的使用者名稱
* remotePath : 欲上傳的目錄位置
* privateKeyPath : 私鑰
* password : 若有設私鑰則可有(null)可無
* uploadOnSave : 自動上傳，個人是沒有開啟(經常性按 ctrl + S)
* ignore : 

``` 
{
    "name": "my_work_space",
    "host": "192.168.X.X",
    "port": 22,
    "protocol": "sftp",
    "username": "centos",
    "remotePath": "\\home\\centos\\***\\",
    "privateKeyPath": "C:\\***.pem", 
    "uploadOnSave": false,
    "syncMode": "update",
    "ignore": [
        "**/.vscode/**",
        "**/.git/**",
        "**/.svn/**",
        "**/.DS_Store"
        ]
}
```

# 同步所有檔案

儲存後按下同樣打開指令窗口並輸入「SFTP: Sync Local -> Remote」，再選取剛剛設定中的 name(ex.my_work_space) 便可以重部整份資料夾/工作區到遠端伺服器上囉。

![](https://i.imgur.com/OuLh788.png)

# 同步單一檔案

僅需在左側目錄欄中對檔案右鍵並選取 Upload 便可以同步單一檔案了

![](https://i.imgur.com/JMKKSpL.png)

# 如何更快速(懶)

首先可以把 uploadOnSave 打開，便在儲存時會同步所有檔案。而單一檔案的話則可以更改快捷鍵，設定好 Upload Active File，則下次儲存完按下快速鍵後便已經同步完成囉。

![](https://i.imgur.com/keYG6oe.png)
