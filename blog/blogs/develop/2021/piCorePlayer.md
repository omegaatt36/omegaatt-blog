---
title: Raspberry pi 4 + piCorePlayer 7.0.0 折騰筆記
date: 2021-05-02
tags:
  - raspberry_pi
  - piCorePlayer
categories:
 - develop
---

# 前言

鑒於想要剛好逛到 Raspberry pi 4 出了(2019)，剛好手邊又有兩顆閒置的喇叭，於是產生了做一套**網路串流音樂撥放器**。用到的硬體如下：

- [Raspberry pi 4b 4G ver 1.2](https://www.raspberrypi.org/products/raspberry-pi-4-model-b/)  
  樹莓派的部分選用 4G 版本，由於 4G 跟 8G 版本差價有點大，僅作為多媒體撥放器，應該不需要那麼大量的記憶體，故降梯成本僅選用 4G 版本。
  至於 1.2 版本差在哪裡可以參考[這篇文章](https://www.raspberrypi.com.tw/30359/news-raspberry-pi4-new-version-fixes-usb-type-c-problem)，主要修正了 type-C 供電問題。
- [Argon ONE m.2](https://www.argon40.com/argon-one-m-2-case-for-raspberry-pi-4.html)  
  由於 pi 4b 的 cpu 較 3b+ 升級很多，同時[發熱量也提升](https://www.reddit.com/r/raspberry_pi/comments/cbqt19/raspberry_pi_4_heatsink_testing/)了，故得選用一個同時具有散熱、美觀、擴充的外殼，剛好手邊有汰換下來的 m.2 SATA SSD，於是選擇了 Argon ONE m.2。
- 直流電源供應器  
  儘管線性電源如 Keces P3 或便宜一點 ifi iPower X 或是老虎魚都是很好的穩定 5V 直流電源供應器，但由於使用樹莓派就是不想花大錢，用比系統本身還高價如此多的周邊，似乎有點本末倒置，魚是只用了手機的 type-C 5V/3A 變壓器作為電源供應器。


# 硬體安裝

硬體安裝部分不在這次探討範圍，由於每個人使用的外殼均不相同，文章中就假設是裸機設定。

# 系統選擇

由於是第一次接觸 Raspberry pi as audio player，一開始是看到 Volumio 與 Moode 等等，也實際安裝了一次 Volumio，但由於慧根不夠(X)，或是不知道撞到了什麼，只要不小心設定錯音訊，便會導致系統無法撥音樂，也懶得 debug 了，於是想說乾脆來挑戰看看 piCorePlayer。

其實 piCorePlayer 官方本身的 [Tutorial](https://docs.picoreplayer.org/how-to/) 已經十分完善了，猴子照著弄都會，本文中若有說明不完善的可以到官方的教學尋找。

# 燒錄

首先到[官網下載頁](https://docs.picoreplayer.org/downloads/)下載對應的 piCorePlayer 壓縮包，雖然網路上看到的資料是說要找 Experimental RealTime Kernel (RT)版本，但稍微看一下 7.0.0 版本的 Release Note:
> 64bit Kernel 5.4.83
> 
> Known Issues  
> - No support for RT kernels. I don't see support for these kernels continuing.  

官方已經不會提供實驗性版本的 Kernel，不過如果是高級玩家應該還是會自己 compile kernel 吧XD。

總之在 win 環境下直接下載 [7.0.0](https://repo.picoreplayer.org/insitu/piCorePlayer7.0.0/piCorePlayer7.0.0-64Bit.zip) 版本 zip 後，解壓縮會解出一個 `piCorePlayer7.0.0-64Bit.img` 映像檔，可以透過 Win32DiskImager 進行燒錄。

![](/assets/dev/20210502/Win32DiskImager_rcv0s3SzgH.png)

燒錄完成後，插上 pi 的 sd 卡座，插電~開機~輕鬆秒殺。

在本機瀏覽器上輸入 http://pcp.local/ 後，便會看到歡迎畫面：
![](/assets/dev/20210502/chrome_JeTbbPYCfT.png)

# 透過 ssh 連進去

到 Gateway 上，可以找到 host name 為 pCP 對應的 IP。

![](/assets/dev/20210502/chrome_53jekDRsT7.png)

可以透過 `ssh tc@{ip}` 連進去，密碼預設為 `piCore`

![](/assets/dev/20210502/Terminus_7JJucowIZE.png)

# 安裝 LMS Logitech Media Server

## Resize File System

1. 首先先對檔案系統擴容，到 Main Page 找到 Resize FS。  
  ![](/assets/dev/20210502/chrome_65DMPFFeMS.png)
2. 如果 SD 卡夠大的話，可以選擇 `Whole SD Card`，並點選 Resize 按鈕。  
  ![](/assets/dev/20210502/chrome_XMv1jGZBiu.png)
3. 結束之後會提示重啟，並倒數 90 秒後會回到主畫面。  
  ![](/assets/dev/20210502/chrome_5QTwIMK697.png)

## Install LMS

1. 點選到 LMS 的分頁，會看到尚未安裝也尚未執行。  
  ![](/assets/dev/20210502/chrome_zqFZOYVoC1.png)
2. 在安裝 LMS 前，由於我是用 Argon ONE m.2，他有一個 USB 的存儲裝置，可以先將 USB mount 上去，路徑設定 `LMSfiles`。如果沒有自動導入 UUID 的話，可以[參考文章下方的操作](#usb-storage)。
  ![](/assets/dev/20210502/chrome_JKl9VJzg4d.png)
3. 再拉到最下面，點選 Beta 用來開啟更多選項。  
  ![](/assets/dev/20210502/chrome_2YekO6nVox.png)
4. 接著可以拉上去點選 Install，完成之後會看到以下畫面。  
  ![](/assets/dev/20210502/OBeF9EuEqK.png)
5. 現在可以點選 Start LMS 來啟動 service。  
  ![](/assets/dev/20210502/chrome_ZeZscW9ANI.png)
6. 完成後就會看到 LMS is running。  
  ![](/assets/dev/20210502/chrome_uylNTrgadI.png)
7. 接著可以把 LMS 移動到剛剛 mount 的 USB Disk 上。  
  ![](/assets/dev/20210502/chrome_lAbHu6yTwl.png)

## Setup LMS

1. 在網址輸入 `http://pcp.local:9000/` 進入 LMS 的頁面，第一次進去需要設定一些資訊，首先如果沒有使用 mysqueezebox 服務的話，可以直接點選右下角 Next。
  ![](/assets/dev/20210502/chrome_4tKqjEgvFF.png)
2. 再來會需要設定 Local Music Folder，點選 `/mnt/LMSfiles/music` 後再次確認左下角是這個路徑後，點選 Next。
  ![](/assets/dev/20210502/chrome_4JPsbz2Rmr.png)
3. 接著會需要設定 Playlist Folder，點選 `/mnt/LMSfiles/playlists` 後再次確認左下角是這個路徑後，點選 Next。
  ![](/assets/dev/20210502/chrome_Jxpl1wnSYd.png)
4. 完成後就會看到撥放的畫面(此時我已經在 `/mnt/LMSfiles/music` 放入一些音樂了，於是可以看到右邊有音樂)。
  ![](/assets/dev/20210502/chrome_dpnI18i5FL.png)

## LMS SKIN

1. 進入 LMS 的設定頁，點選 Plugins 分頁，往下可以找到 Material Skin，點下左邊的 check box 後按下右下角 Apply。
  ![](/assets/dev/20210502/chrome_phy3LEYWQo.png)
2. (重新整理後)到頁面最上面，會看到可以更新，再次點選 Material Skin 左邊的 check box 後 Apply。
  ![](/assets/dev/20210502/chrome_UOLtxDg5mm.png)
3. 切換至 Interface 頁面，把 Web Interface 換成 Material Skin。
  ![](/assets/dev/20210502/chrome_hB38pIEyyB.png)
4. 重新進入 `http://pcp.local:9000/` 便會看到已經換成稍微好看一點的 skin。
  ![](/assets/dev/20210502/chrome_2mMgYwQ1ne.png)

# Setting USB Audio

都設定好 LMS 後，接著就需要設定輸出音訊的部分，我是用 USB DAC，可以到[官方支援的 USB DAC 列表](https://sites.google.com/site/picoreplayer/home/List-of-USB-DACs)檢查是否支援自己的設備。


1. ssh 連進去樹莓派後，下指令 `aplay -l`，我的設備就是下面的 `card 1: Audio`
  ```bash
  tc@pCP:~$ aplay -l
  **** List of PLAYBACK Hardware Devices ****
  card 0: Headphones [bcm2835 Headphones], device 0: bcm2835 Headphones [bcm2835 Headphones]
    Subdevices: 7/8
    Subdevice #0: subdevice #0
    Subdevice #1: subdevice #1
    Subdevice #2: subdevice #2
    Subdevice #3: subdevice #3
    Subdevice #4: subdevice #4
    Subdevice #5: subdevice #5
    Subdevice #6: subdevice #6
    Subdevice #7: subdevice #7
  card 1: Audio [TX-384Khz Hifi Type-C Audio], device 0: USB Audio [USB Audio]
    Subdevices: 1/1
    Subdevice #0: subdevice #0
  tc@pCP:~$
  ```
2. 再來到 Squeezelite Settings 分頁中，選擇 USB audio。
  ![](/assets/dev/20210502/chrome_npiRBoQ8kw.png)
3. 會提示須要重啟，於是重啟後稍等。
  ![](/assets/dev/20210502/chrome_RwtD6QjqQM.png)
4. 在下方 Output setting 點選右邊紅色的 more>，會展開所以有輸出設備，選擇 `hw: CARD=Audio,DEV-0`，Audio 即為方才 aplay -l 列出的 USB 音訊設備。
  ![](/assets/dev/20210502/chrome_kYTfL82jOV.png)
5. 接著設定一些 buffer 設定後儲存。
  ![](/assets/dev/20210502/chrome_36a67pTpPY.png)

都完成後就可以到 LMS 的撥放頁面去享受音樂囉~~

# Argon ONE Settings

[可以先參考文章](https://forums.slimdevices.com/showthread.php?113575-How-To-Using-the-Argon-One-case-for-the-Pi-4B-together-with-piCorePlayer-(7-x)&p=1002008&viewfull=1#post1002008)，細節待補。

# LMS Yourube Plugin

待補。

# Something More

## USB Storage

piCorePlayer 會自動識別 ext4 格式的 USB 儲存，若非 ext4 可以先進行格式化。

1. 輸入 `fdisk -l` 來列出找到額外的儲存設備，此處找到為一張 256GB 的硬碟。
  ```bash
  tc@pCP:~$ fdisk -l
  Disk /dev/mmcblk0: 59 GB, 63864569856 bytes, 124735488 sectors
  1948992 cylinders, 4 heads, 16 sectors/track
  Units: sectors of 1 * 512 = 512 bytes

  Device       Boot StartCHS    EndCHS        StartLBA     EndLBA    Sectors  Size Id Type
  /dev/mmcblk0p1    128,0,1     127,3,16          8192     139263     131072 64.0M  c Win95 FAT32 (LBA)
  /dev/mmcblk0p2    1023,3,16   1023,3,16       139264  124538879  124399616 59.3G 83 Linux
  Disk /dev/sda: 233 GB, 250059350016 bytes, 488397168 sectors
  30401 cylinders, 255 heads, 63 sectors/track
  Units: sectors of 1 * 512 = 512 bytes

  Device  Boot StartCHS    EndCHS        StartLBA     EndLBA    Sectors  Size Id Type
  /dev/sda1    0,1,1       1023,254,63         63  488397167  488397105  232G 83 Linux
  ```
2. 使用指令 `sudo fdisk /dev/xxxxx`，而 xxxx 為 `/dev/sda1` 中的 sda，後面那個 `1` 不需要輸入，該數字為分區(partition)。  
  ```bash
  tc@pCP:~$ sudo fdisk /dev/sda
  ```
3. 下一個指令 `d`，刪除分區，後接著輸入指令 `w` 儲存這次操作。
  ```bash
  The number of cylinders for this disk is set to 30401.
  There is nothing wrong with that, but this is larger than 1024,
  and could in certain setups cause problems with:
  1) software that runs at boot time (e.g., old versions of LILO)
  2) booting and partitioning software from other OSs
    (e.g., DOS FDISK, OS/2 FDISK)

  Command (m for help): d
  Selected partition 1  

  Command (m for help): w
  The partition table has been altered.     
  Calling ioctl() to re-read partition table
  tc@pCP:~$ 
  ```
4. 再次輸入 `sudo fdisk /dev/sda`，下指令 `n` 新增分區，輸入 `p` 來設定主要分區，並輸入 `1` 來告入系統分割為一個分區就好，接著連按兩次 enter 使用預設的分區頭尾，並輸入 `w` 儲存這次操作。
  ```bash
  tc@pCP:~$ sudo fdisk /dev/sda

  The number of cylinders for this disk is set to 30401.
  There is nothing wrong with that, but this is larger than 1024,
  and could in certain setups cause problems with:
  1) software that runs at boot time (e.g., old versions of LILO)
  2) booting and partitioning software from other OSs
    (e.g., DOS FDISK, OS/2 FDISK)

  Command (m for help): n
  Partition type
    p   primary partition (1-4)
    e   extended
  p
  Partition number (1-4): 1
  First sector (63-488397167, default 63):
  Using default value 63
  Last sector or +size{,K,M,G,T} (63-488397167, default 488397167):
  Using default value 488397167

  Command (m for help): p
  Disk /dev/sda: 233 GB, 250059350016 bytes, 488397168 sectors
  30401 cylinders, 255 heads, 63 sectors/track
  Units: sectors of 1 * 512 = 512 bytes

  Device  Boot StartCHS    EndCHS        StartLBA     EndLBA    Sectors  Size Id Type
  /dev/sda1    0,1,1       1023,254,63         63  488397167  488397105  232G 83 Linux

  Command (m for help): w
  The partition table has been altered.
  Calling ioctl() to re-read partition table
  tc@pCP:~$ 
  ```

此時該磁碟就會以 ext4 存在於系統中，就可以進行掛載的操作了。