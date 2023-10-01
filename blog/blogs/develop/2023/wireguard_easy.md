---
title: 使用 wireguard 與 aws 搭建虛擬內網
date: 2023-10-01
categories:
 - develop
tags:
 - linux
---

先前在「[在 AWS 上使用 EC2 建立 FRP 玩玩內網穿透](/blogs/develop/2022/frp-tunnel)」一文中敘述了如何透過 AWS 實現虛擬穿透，也在內文中提到或許可以使用 wireguard 搭建內網，其原因也很簡單：每更新一個 port 都要重新設定 server side 實在是太麻煩了，拖更了進一年終於要開始寫 wireguard 的使用了。（AWS 免費也快到期了）

## 遇到了什麼問題

在使用 wireguard 來搭建 VPN 前，我是使用 [zerotier](https://www.zerotier.com/) 搭配 Mikrotik RB5009 所執行的 RouterOS 7.x 版本，讓外網可以連回家裡的網路環境

但 zerotier 的網路拓樸是存在他們官方伺服氣上，與其相信 zerotier，我想在 vps 上建立虛擬機，並只許特定 ip 登入，可能更加可靠（？

## 解決方法

於是我在 frp tunnel VM 上建立了 wireguard 節點，作為外網與內網溝通的橋樑。

[wireguard](https://www.wireguard.com/) 是一個高效的現代化 VPN，目標是比 IPsec 更快。在 2020 年時正式推出 1.0.0 版本。

wireguard 的拓樸實際上是 peer2peer，甚至可以達到 full mesh，但礙於錢錢不夠，單一個節點作為 server side 就足夠了。

### server

- 使用 docker-compose 能快速部署 wireguard 服務，我們使用的是 [wg-easy](https://github.com/wg-easy/wg-easy)
  簡單換掉一些參數:
  - `WG_HOST`: vps 的公網 IP
  - `WG_DEFAULT_ADDRESS`: 預設內網 ip 範圍，比如 `10.100.0.x`，比如新增一個 client 端就的預設會是 `10.100.0.1`
  - `PASSWORD`: 登入管理頁面時的密碼
  ```
  # source: https://github.com/wg-easy/wg-easy/blob/master/docker-compose.yml

  version: "3.8"
  services:
    wg-easy:
      environment:
        # ⚠️ Required:
        # Change this to your host's public address
        - WG_HOST=x.x.x.x

        # Optional:
        # - PASSWORD=foobar123
        # - WG_PORT=51820
        # - WG_DEFAULT_ADDRESS=10.8.0.x
        # - WG_DEFAULT_DNS=1.1.1.1
        # - WG_MTU=1420
        # - WG_ALLOWED_IPS=192.168.15.0/24, 10.0.1.0/24
        # - WG_PRE_UP=echo "Pre Up" > /etc/wireguard/pre-up.txt
        # - WG_POST_UP=echo "Post Up" > /etc/wireguard/post-up.txt
        # - WG_PRE_DOWN=echo "Pre Down" > /etc/wireguard/pre-down.txt
        # - WG_POST_DOWN=echo "Post Down" > /etc/wireguard/post-down.txt

      image: weejewel/wg-easy
      container_name: wg-easy
      volumes:
        - .:/etc/wireguard
      ports:
        - "51820:51820/udp"
        - "51821:51821/tcp"
      restart: unless-stopped
      cap_add:
        - NET_ADMIN
        - SYS_MODULE
      sysctls:
        - net.ipv4.ip_forward=1
        - net.ipv4.conf.all.src_valid_mark=1
  ```
- 隨後透過 `docker compose up -d` 來啟動
- 到 AWS 或任何 vps 的網路傳入規則中新增兩條規則
  1. udp 51820 允許 0.0.0.0/0（全開），讓全世界可以訪問 wireguard server
  2. tcp 51821 允許你家的 IP 地址，比如 1.2.3.4/32，用來登入 wg-easy 的管理頁面
  ![](/assets/dev/20231001/Screenshot_20231001_122839.png)
- 接著打開瀏覽器，進入 wg-easy 管理頁面，比如 x.x.x.x:51821，登入後新增 config，就能下載 config 了

### client

詳細可以看 https://www.wireguard.com/install/

- [Android](https://play.google.com/store/apps/details?id=com.wireguard.android&hl=zh_TW&gl=US&pli=1) 與 [iOS](https://apps.apple.com/us/app/wireguard/id1441195209) 上安裝 wireguard 的 app 後直接透過掃描 QR code 來存入設定檔就好。
- [Android TV 或是 Google TV] 下載完 app 後則可以透過載入設定檔或手動輸入來建立 client 設定檔。

linux 是我最常手動載入設定檔的，做個筆記。
```bash
sudo apt install wireguard resolvconf -y
sudo -i
sudo echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf
sysctl -p
exit

mkdir -p /etc/wireguard/
chmod 0777 /etc/wireguard

cd ${HOME}/Downloads
mv xxxxx.conf /etc/wireguard/wg0.conf

systemctl enable wg-quick@wg0
```

## 成效

使用場景：
  - 最簡單的，世界各地只要連的到 aws，就能連回家裡存取 nas 內的資料，家裡的 router 也不用打洞或設置 DMZ，甚至不用擔心裸奔被攻擊。
  - 外出到旅館後可以使用 android tv 連回家裡的 Jellyfin server 觀看動畫或電影。
  - 在公司可以連回家裡的 pve homelab 進行實驗性部署測試。
  - 在家架設 ReDroid 或是 WayDroid 或是 windows vm 上跑 Android 模擬器，掛著掛機遊戲，想起這件事情就連回家操作，然後繼續掛機。

## 如何更好

過去使用 zerotier 是直接將 router 當作節點，wireguard [也可以這麼做](https://help.mikrotik.com/docs/display/ROS/WireGuard)

或是未來有不同地點的租屋需求，可以讓兩邊直接使用 router 作為 wireguard tunnel

