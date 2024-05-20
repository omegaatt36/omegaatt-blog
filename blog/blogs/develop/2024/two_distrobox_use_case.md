---
title: 透過 Distrobox 解決 Linux 環境依賴問題
date: 2024-05-18
categories:
 - develope
tags:
 - linux
---

## 前言

自從一年前嘗試使用 [KDE Neon](/blogs/develop/2023/kde_neon_experience) 後，主要的桌面開發環境就從 wsl 整個遷移到 Linux 上。這一年來感受到，KDE Neon 可能正如 reddit 上的留言闡述的那樣，只是 KDE 團隊的「實驗場」，可以享受到「最」新的 Plasma 與「較」新的 Kernel，卻始終是基於 Ubuntu 22.04 LTS，即便 Ubuntu 24.04 LTS 推出之際，也不再讓我感受性感。

因緣際會，在 Ubuntu 24.04 LTS 發布之際，我決定從 KDE Neon distro hopping 到 OpenSUSE Tumbleweed，這是一個由 SUSE 公司推出的滾動式發行版，軟體包在 SUSE 測試過後便會推送到 OpenSUSE 的 Repo（其他 OpenSUSE 的功能省略）。這次的轉換主要是為了獲得最前沿的 Linux 體驗。OpenSUSE Tumbleweed 的滾動更新模式能確保我始終使用最新的軟體和技術，這對我來說是一個有趣的選擇，即便有可能收到有問題的 xz 更新 XD。

```shell
❯ fastfetch
                                     ......             raiven@raiven-suse
     .,cdxxxoc,.               .:kKMMMNWMMMNk:.         ------------------
    cKMMN0OOOKWMMXo. A        ;0MWk:'      ':OMMk.      OS: openSUSE Tumbleweed 20240517 x86_64
  ;WMK;'       'lKMMNM,     :NMK'             'OMW;     Host: FMV UH-X
 cMW;             WMMMN   ,XMK'                 oMM.    Kernel: 6.8.9-1-default
.MMc             ''^*~l. xMN:                    KM0    Uptime: 1 hour, 14 mins
'MM.                   .NMO                      oMM    Packages: 3211 (rpm), 37 (flatpak)
.MM,                 .kMMl                       xMN    Shell: zsh 5.9
 KM0               .kMM0' .dl>~,.               .WMd    DE: KDE Plasma 6.0.4
 'XM0.           ,OMMK'    OMMM7'              .XMK     WM: KWin (Wayland)
   *WMO:.    .;xNMMk'       NNNMKl.          .xWMx      WM Theme: Nordic
     ^ONMMNXMMMKx;          V  'xNMWKkxllox0NMWk'       Theme: Breeze (NordicDarker) [QT], Breeze-Dark [GTK2], Breeze [GTK3]
         '''''                    ':dOOXXKOxl'          Icons: Zafiro-Nord-Dark-Black [QT], Zafiro-Nord-Dark-Black [GTK2/3/4]
                                                        Font: Noto Sans (10pt) [QT], Noto Sans (10pt) [GTK2/3/4]
                                                        Cursor: Nordic (24px)
                                                        Terminal: tmux 3.4
                                                        CPU: 13th Gen Intel(R) Core(TM) i5-1335U (12) @ 4.60 GHz
                                                        GPU: Intel Iris Xe Graphics @ 1.25 GHz [Integrated]
                                                        Memory: 9.86 GiB / 15.27 GiB (65%)
                                                        Swap: 677.00 MiB / 2.00 GiB (33%)
                                                        Disk (/): 88.81 GiB / 1.82 TiB (5%) - btrfs
```

即便是已經足夠穩定的 OpenSUSE，也是可以很輕易的被搞的不穩定，正如 debian 在 [DontBreakDebian](https://wiki.debian.org/DontBreakDebian) 一文中提到的：

> On Debian installing software from random websites is a bad habit.

文章中建議我們使用 chroot, containers, vm 等等技術來與主機隔離，這也是 [Never Install Locally](https://www.youtube.com/watch?v=J0NuOlA2xDc) 的精隨，於是我就想起了 Distrobox 這個工具。

## Distrobox 是什麼

在日常開發環境中，我們時常會面臨到不同軟體和工具對作業系統版本的需求差異。為了解決這些問題，許多人選擇使用虛擬機或 Docker 來管理不同的開發環境。然而，這些解決方案要麼過於繁瑣，要麼資源佔用過多。這時候，Distrobox 成為了一個理想的選擇。

### 什麼是 Distrobox？

[Distrobox](https://distrobox.it/compatibility/) 是一個基於 Podman 或 Docker 的工具，允許我們在現有 Linux 發行版上運行不同的 Linux 發行版，並且這些容器可以無縫地與主機系統互動。它的設計目的是為了提供一個輕量且靈活的方式來管理多種 Linux 環境，特別適合那些需要在不同發行版之間進行測試和開發的人。

### 為什麼選擇 Distrobox？

- 輕量且快速：與 VM 相比，Distrobox 使用 container 來執行不同的 Linux 發行版，這意味著它佔用的資源更少，啟動速度更快。
- 無縫整合：Distrobox 容器與主機系統無縫整合，可以直接訪問主機的文件系統、網絡和設備。這使得在容器中運行的應用程序可以像本地應用一樣操作。
- 靈活性高：我們可以在同一台機器上運行多個不同的 Linux 發行版，這對於需要在多種環境中進行測試的開發者來說非常方便。
- 簡單易用：通過簡單的命令即可創建、管理和銷毀容器。這大大降低了學習曲線，讓更多人能夠輕鬆上手。

### 如何使用

詳細的安裝與教學都可以到 [GitHub](https://github.com/89luca89/distrobox) 上找到，官網內也有很多 [useful tips](https://distrobox.it/useful_tips/)。

## UseCase

透過兩個我實際遇到的問題，來展示 Distrobox 的使用案例

### 案例一：想要的可執行檔沒有在發行版上發布

前言所述，我跳到了 OpenSUSE Tumbleweed，過去我會使用 [`pandoc`](https://pandoc.org/) 從 markdown 轉 html，接著夠過 [`wkhtmltopdf`](https://wkhtmltopdf.org/) 來將 html 轉成 pdf。但當我們嘗試安裝 `wkhtmltopdf` 時，會發生沒有在 OpenSUSE Tumbleweed 上發布的情況：

```shell
❯ sudo zypper in wkhtmltopdf
Loading repository data...
Reading installed packages...
'wkhtmltopdf' not found in package names. Trying capabilities.
No provider of 'wkhtmltopdf' found.
Resolving package dependencies...
Nothing to do.
```

此時我的選擇是，透過 Distrobox 建立一個 Ubuntu 的 container，且剛好是 Ubuntu 24.04 LTS 發布，於是我使用的發行版為 Ubuntu Noble。

#### 透過 Distrobox 建立容器

這個操作主要是呼叫 [`distrobox-create`](https://github.com/89luca89/distrobox?tab=readme-ov-file#what-it-does) 來建立一個 rootless 的 container(根據設定決定是 docker 或是 podman，以及是 rootful 或是 rootless)。

```shell
❯ distrobox create --name ubuntu-noble --image ubuntu:noble
Creating 'ubuntu-noble' using image ubuntu:noble        [ OK ]
Distrobox 'ubuntu-noble' successfully created.
To enter, run:

distrobox enter ubuntu-noble
```

#### 進入容器

根據 create 指令結束時的回應，我們可以進入容器內，此時會安裝一些基本的套件，並且綁定一些目錄（例如家目錄），這正是使用 Distrobox 與直接使用 Container 的不同之處。

```shell
❯ distrobox enter ubuntu-noble
Starting container...                            [ OK ]
Installing basic packages...                     [ OK ]
Setting up devpts mounts...                      [ OK ]
Setting up read-only mounts...                   [ OK ]
Setting up read-write mounts...                  [ OK ]
Setting up host's sockets integration...         [ OK ]
Integrating host's themes, icons, fonts...       [ OK ]
Setting up package manager exceptions...         [ OK ]
Setting up package manager hooks...              [ OK ]
Setting up dpkg exceptions...                    [ OK ]
Setting up apt hooks...                          [ OK ]
Setting up distrobox profile...                  [ OK ]
Setting up sudo...                               [ OK ]
Setting up user groups...                        [ OK ]
Setting up kerberos integration...               [ OK ]
Setting up user's group list...                  [ OK ]
Setting up existing user...                      [ OK ]
Setting up user home...                          [ OK ]
Ensuring user's access...                        [ OK ]

Container Setup Complete!
```

在 zsh 中 command line 的最左側會顯示目前的發行版，或是可以查看 `/etc/os-release`

```shell
❯ source /etc/os-release && echo $ID
ubuntu

╭─    ~                                ✔  in ubuntu-noble   0.87   5.62G   at 13:44:25 
╰─
```

#### 安裝需要的執行檔

可以直接透過 Debian/Ubuntu 的包管理器 Aptitude 來進行安裝

```shell
❯ sudo apt install pandoc wkhtmltopdf
Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
...
0 upgraded, 108 newly installed, 0 to remove and 0 not upgraded.
Need to get 80.8 MB of archives.
After this operation, 400 MB of additional disk space will be used.
Do you want to continue? [Y/n] Y
...
❯ which pandoc
/usr/bin/pandoc
❯ which wkhtmltopdf
/usr/bin/wkhtmltopdf
```

#### 驗收

在容器內執行我的 Makefile 中的 pdf_en 段落：

```shell
❯ cat Makefile
html_en:
    pandoc --standalone --include-in-header $(STYLES_DIR)/$(STYLE).css \
           --lua-filter=pdc-links-target-blank.lua \
           --from markdown --to html \
           --template template.html \
           --output $(OUT_DIR)/index.html $(ORG_EN) metadata.yaml

❯ make pdf_en
mkdir -p build
pandoc --standalone --include-in-header styles/chmduquesne.css \
           --lua-filter=pdc-links-target-blank.lua \
           --from markdown --to html \
           --template template.html \
           --output build/index.html index.md metadata.yaml
wkhtmltopdf build/index.html --disable-javascript build/index.pdf
Loading page (1/2)
Printing pages (2/2)
Done
```

#### 這個_我不要了

當我們不再需要這個容器時，可以透過 Distrobox 來砍掉容器。由於我的 distrobox 後端為 podman，我們可以先使用 podman container ls 來查看，並且最後再驗收容器使否真的有被殺掉。

```shell
❯ podman container ls -a | grep ubuntu-noble
9dd11776ce28  docker.io/library/ubuntu:noble           --verbose --name ...  Up 4 minutes   Created                                          ubuntu-noble

❯ distrobox stop ubuntu-noble

❯ distrobox rm ubuntu-noble
Removing exported binaries...
Removing container...
ubuntu-noble

❯ podman container ls -a | grep ubuntu-noble
```

#### 案例總結

在這個案例中，我們可以透過 Distrobox 來透過其他發行版的套件管理，來安裝其他發行版的套件，有別於直接使用 docker/podman，他會將指定的路徑直接 mount 在 HOME 上（預設是本地的家目錄），同時隔離了系統層的文件。

### 案例二：我的環境太新的，無法安裝舊的依賴

某一天我想要安裝某個桌面應用，例如[非官方的 ChatGPT 桌面版](https://github.com/lencx/ChatGPT)，這個在 linux 上只有提供 `AppImage` 與 `.deb`。

#### 嘗試使用 AppImage

首先嘗試使用 AppImage 版本，會發生缺少套件的情況：

```shell
❯ ./chat-gpt_1.1.0_amd64.AppImage
Gtk-Message: 14:05:34.065: Failed to load module "appmenu-gtk-module"
[2024-05-19][06:05:34][chatgpt::app::setup][INFO] stepup
[2024-05-19][06:05:34][chatgpt::app::setup][INFO] global_shortcut_unregister
[2024-05-19][06:05:34][chatgpt::app::setup][INFO] stepup_tray
[2024-05-19][06:05:34][chatgpt::app::setup][INFO] run_check_update
[2024-05-19][06:05:34][chatgpt::utils][INFO] run_check_update: silent=false has_msg=None
[2024-05-19][06:05:34][attohttpc][DEBUG] trying to connect to lencx.github.io:443
[2024-05-19][06:05:34][attohttpc][DEBUG] trying to connect to [2606:50c0:8000::153]:443
[2024-05-19][06:05:34][attohttpc][DEBUG] failed to connect to [2606:50c0:8000::153]:443: Network is unreachable (os error 101)
[2024-05-19][06:05:34][attohttpc][DEBUG] trying to connect to 185.199.109.153:443
[2024-05-19][06:05:34][reqwest::connect][DEBUG] starting new connection: https://raw.githubusercontent.com/
[2024-05-19][06:05:34][attohttpc][DEBUG] successfully connected to 185.199.109.153:443, took 65ms
[2024-05-19][06:05:34][attohttpc][DEBUG] GET /ChatGPT/install.json HTTP/1.1
Gtk-Message: 14:05:34.330: Failed to load module "appmenu-gtk-module"
Gtk-Message: 14:05:34.383: Failed to load module "appmenu-gtk-module"
[2024-05-19][06:05:34][attohttpc][DEBUG] creating a length body reader
[2024-05-19][06:05:34][attohttpc][DEBUG] creating gzip decoder
[2024-05-19][06:05:34][attohttpc][DEBUG] status code
```

當我想要安裝缺少的 module 時，又會發生依賴找不到的問題：

```shell
❯ sudo zypper install appmenu-gtk-module
[sudo] password for root:
Loading repository data...
Reading installed packages...
'appmenu-gtk-module' not found in package names. Trying capabilities.
No provider of 'appmenu-gtk-module' found.
Resolving package dependencies...
Nothing to do.
```

#### 嘗試使用 Distrobox 解除依賴問題

於是我把目光轉向 `.deb`，我啟動了剛才建立的 Ubuntu Noble，並嘗試從 `.deb` 來安裝，卻同樣會發生缺少依賴：

```shell
❯ sudo apt install ./ChatGPT_1.1.0_linux_x86_64.deb
Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
Note, selecting 'chat-gpt' instead of './ChatGPT_1.1.0_linux_x86_64.deb'
Some packages could not be installed. This may mean that you have
requested an impossible situation or if you are using the unstable
distribution that some required packages have not yet been created
or been moved out of Incoming.
The following information may help to resolve the situation:

The following packages have unmet dependencies:
 chat-gpt : Depends: libwebkit2gtk-4.0-37 but it is not installable
E: Unable to correct problems, you have held broken packages.
```

再次嘗試搜尋依賴，發現 Ubuntu Noble 只收錄了 `libwebkit2gtk-4.1`

```shell
❯ sudo apt search libwebkit2gtk-4.0-37
Sorting... Done
Full Text Search... Done
```

#### 使用另一個發行版

於是我把目光轉向「更穩定」（套件更舊）的 Debian（但是使用較新的 Debian12），並且再次嘗試搜尋是否具有該套件：

```shell
# 此時是在 host
❯ distrobox create --name debian --image debian:12
❯ distrobox enter debian
... # 以下省略


# 此時是在 debian 容器內
❯ sudo apt search libwebkit2gtk-4.0-37
Sorting... Done
Full Text Search... Done
libwebkit2gtk-4.0-37/stable-security,now 2.44.1-1~deb12u1 amd64 [installed,automatic]
  Web content engine library for GTK
```

#### 安裝目標桌面應用

```shell
❯ sudo apt install ./ChatGPT_1.1.0_linux_x86_64.deb
Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
...
```

我們可以嘗試在 container 開啟應用程式，可以順利打開。

```shell
# container 內
❯ chat-gpt
[2024-05-19][06:15:20][chatgpt::app::setup][INFO] stepup
[2024-05-19][06:15:20][chatgpt::app::setup][INFO] global_shortcut_unregister
[2024-05-19][06:15:20][chatgpt::app::setup][INFO] run_check_update
[2024-05-19][06:15:20][chatgpt::app::setup][INFO] stepup_tray
[2024-05-19][06:15:20][chatgpt::utils][INFO] run_check_update: silent=false has_msg=None
[2024-05-19][06:15:20][attohttpc][DEBUG] trying to connect to lencx.github.io:443
[2024-05-19][06:15:20][attohttpc][DEBUG] trying to connect to [2606:50c0:8000::153]:443
[2024-05-19][06:15:20][attohttpc][DEBUG] failed to connect to [2606:50c0:8000::153]:443: Network is unreachable (os error 101)
[2024-05-19][06:15:20][attohttpc][DEBUG] trying to connect to 185.199.109.153:443
[2024-05-19][06:15:20][reqwest::connect][DEBUG] starting new connection: https://raw.githubusercontent.com/
[2024-05-19][06:15:20][attohttpc][DEBUG] successfully connected to 185.199.109.153:443, took 66ms
...
```

也能在主機透過 cli 快速開啟應用。

```shell
# 主機端
❯ distrobox enter debian -- chat-gpt
```

#### 在主機安裝應用程式的入口(entry)

在 container 內使用 `distrobox-export` 可以直接在主機的應用程式列表內註冊該應用。

```shell
# 切記要在容器內呼叫，否則不會有任何事情發生
❯ distrobox-export --app chat-gpt
```

此時我們就能在應用程式列表中看到該應用，也很好心得幫我們把 Distrobox 的容器名稱一併住記上去。

![kwin_open_chat-gpt](/assets/dev/20240518/20240519_141927.png)

## 長處與侷限

Distrobox 是一個強大且靈活的工具，適合需要多種 Linux 環境的開發者。它的輕量級和高效能使其成為虛擬機和 Docker 的理想替代方案。同時也支援在 MacOS 上使用（雖然 MacOS 的 container 仍是跑在 vm 內，資源使用巨大）。

雖然看起來很萬能，但畢竟是基於容器的操作，若是 Kernel 本身不支援，使用 Distrobox 也無法達到不支援的功能。

例如 OpenSUSE 的 Kernel 不具有 binder，即便照著 [useful tips](https://distrobox.it/useful_tips/#using-waydroid-inside-a-distrobox) 進行設定，仍會在啟動時發生 module not found 的錯誤。
