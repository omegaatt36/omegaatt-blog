---
title: 使用內建的 rsync 備份 Truenas Scale 到 Proxmox Backup Server
date: 2024-05-19
categories:
 - develope
tags:
 - linux
---

## 前言

過去我會使用 backup script 配合 crontab 來定期的備份 nas 的資料，這次更換了 Proxmox Backup Server 的物理機後，多了一個硬碟的空間好讓我實驗 Truenas Scale 的備份機制。

### 差異

rsync 本身並沒有 server/client 的概念，只有 source 與 destination。
過去我會在**備份主機**上透過 samba 來 mount Nas 到資料夾內，檢查有沒有 mount 成功才在**備份主機**上使用 rsync。
而 Truenas Scale 提供的 Data Protection 功能中，內建了 Rsync Tasks 模組，透過預先建立好的 ssh credential 來呼叫**備份主機**進行 rsync。
同樣都是由**備份主機**來進行 rsync，主要是任務的執行呼叫是 Truenas Scale 本身還是**備份主基本身**。

## 在 Truenas Scale 上建立備份任務

### 建立 SSH Pair

首先到 Credentials -> Backup Credentials 的頁面

![](/assets/dev/20240519/20240519_163627.png)

點擊 SSH Configurations 中的 Add

![](/assets/dev/20240519/20240519_163732.png)

並且填上對應的 ip, port 與 username，Remote Host Key 則可以先複製起來備用，待會兒會需要貼到 PBS 的主機內。

![](/assets/dev/20240519/20240519_163821.png)

### 建立 Rsync Task

到 Data Protection 中，點擊 Rsync Tasks 的 Add 按鈕

![](/assets/dev/20240519/20240519_163855.png)

在 Rsync Mode 中選擇 SSH，並選擇剛剛建立好的 SSH Connection。Remote Path 則可能事：
- PBS 上透過 WebUI 建立 Directory 或是 ZFS -> 會綁在 PBS 的 Storate 下面所以 path 中會有 `/mnt/datastore/{name}`。
- PBS 或任意主機上透過 cli 建立的 directory -> 就可能是任意位置，例如 `/root/nas-backup` 等等。

![](/assets/dev/20240519/20240519_163941.png)

## 在 Proxmox Backup Server 上建立備份目標

### 建立 Storage

這邊選擇在 WebUI 上建立 Directory。

![](/assets/dev/20240519/20240519_164048.png)

### 儲存 authorized key

到 PBS 的 cli 下，透過 `cat 'authorized key' > ~/.ssh/authorized_keys` 來儲存剛才產生的 ssh remote key。

![](/assets/dev/20240519/20240519_164332.png)

## 驗收

回到 Truenas Scale，點擊 Rsync Tasks 中建立好的任務的 Run Now 按鈕，就可以看到任務開始跑了。

![](/assets/dev/20240519/20240519_170938.png)

## 錯誤處理

PBS 預設沒有安裝 rsync，可以在 PBS 中使用 `apt install rsync` 來安裝。
