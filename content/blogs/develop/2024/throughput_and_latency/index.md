---
title: 透過 throughput 與 latency 的關係解決高延遲問題
date: 2024-06-16
categories:
 - develope
tags:
 - backend
---


在軟體工程中，我們常會面臨服務 high latency 的問題，而要解決這個問題，理解 throughput 與 latency 的關係是關鍵。

## 名詞定義

什麼是 throughput 與 latency？

- latency：指從 server 收到請求到回應的時間，通常以毫秒為單位計算。latency 越低，代表回應速度越快。
- throughput：指 server 在特定時間內處理的請求數量，通常以每秒請求數（RPS 或 QPS）表示。throughput 越高，代表 server 處理能力越強。

## server 如何工作

server 接收請求後，由一個或多執行緒（threads）處理請求並生成回應。單執行緒 server 一次只能處理一個請求，而多執行緒 server 可以同時處理多個請求。

舉例說明：自助洗衣店

- latency：從顧客開始操作洗衣機到洗衣完成的時間。這包括顧客等候空閒洗衣機的時間，以及實際洗衣的時間。等候時間越長，latency 越高。
- throughput：洗衣店每小時完成的洗衣次數。洗衣機數量越多，洗衣店能處理的顧客數量越多，throughput 越高。

### latency 與 throughput 的關係

latency 與 throughput 之間存在緊密的關係，通常在系統負載增加時，latency 也會增加。當請求數量超過 server 的處理能力時，請求會被排隊等待處理，這就是 queueing latency。

例如，一個洗衣店有四台洗衣機，每台洗衣機需要30分鐘來完成一輪洗衣。當顧客數量增加，所有洗衣機都在運作時，新來的顧客需要等候空閒洗衣機，這樣就產生了 latency。如果每小時來的顧客數量超過洗衣機的處理能力，就會導致顧客排隊等候，latency 增加。

### 解決方法

1. 增加處理能力：增加洗衣機的數量，提升洗衣店的最大 throughput。
    增加洗衣機數量或開設更多分店，可以有效提升最大 throughput，減少顧客等候時間。
1. 優化內部處理：降低單個洗衣過程的時間，減少內部 latency。
    通過使用更高效的洗衣機，縮短洗衣時間，可以降低單次洗衣的時間。

## 實例分析

假設一個洗衣店有四台洗衣機，每台洗衣機需要30分鐘來完成一輪洗衣。當每小時來的顧客數量增加到5人時，洗衣店的最大 throughput 為每小時8人次，因此新來的顧客需要排隊等候空閒洗衣機，導致 latency 增加。通過增加洗衣機數量，可以提高洗衣店的最大 throughput，減少顧客排隊等候，降低 latency。

如果我們將洗衣機數量增加到8台，最大 throughput 將提升至每小時16人次，這樣可以清除積壓的顧客，減少排隊 latency，使 latency 回到正常水平。

### queueing latency vs. true latency

- queueing latency：顧客在洗衣機前等待的時間。
- true latency：實際洗衣的時間。

當洗衣店忙碌時，queueing latency 增加，總 latency 也隨之增加。通過增加洗衣機的數量或優化洗衣過程，可以減少總 latency。

例如，當洗衣機的洗衣時間從30分鐘優化到20分鐘，總 latency 將會減少，即使在高負載下也能保持較低的 latency。

### 排查 latency 問題

- 如果 throughput 穩定但 latency 增加，可能是排隊 latency 導致的，應考慮增加處理能力。
- 如果 throughput 下降，則應檢查本質 latency 是否增加，找出問題所在並進行優化。

透過理解 throughput 與 latency 的關係，我們可以有效地排查並解決 server 的高 latency 問題，確保系統的穩定運行。

## 實際案例分析

以下是兩個實際案例，說明如何應用這些概念來解決高 latency 問題：

### 案例一：電商網站的促銷活動

在一個電商網站中，某次促銷活動期間，網站的 latency 突然大幅增加。經過分析，發現網站的請求數量超過了 server 的處理能力，導致大量請求被排隊等待，造成高 latency。

為了解決這個問題，團隊首先增加了 server 的執行緒數量，提升了 server 的最大 throughput。其次，團隊優化了數據庫查詢，減少了單個請求的處理時間。最終，通過這些措施，網站的 latency 顯著降低，確保了促銷活動期間的良好用戶體驗。

### 案例二：SATA 與 NVMe 類型的 SSD

在資料存取方面，硬碟的 IOPS（Input/Output Operations Per Second）對 throughput 與 latency 有重大影響。SATA 與 NVMe 類型的 SSD 便是一個典型例子。

- SATA SSD：通常具有較低的 IOPS，這意味著每秒能處理的 I/O 操作數量較少。這會導致在高負載下，請求會被排隊等待處理，增加了 latency。
- NVMe SSD：具有更高的 IOPS，意味著每秒能處理更多的 I/O 操作數量，throughput 更高。在高負載下，NVMe SSD 能夠更快速地處理大量請求，顯著減少 latency。

假設一個數據庫 server 使用 SATA SSD，當同時有大量讀寫操作時，SATA SSD 的 IOPS 無法應付這些操作，導致請求被排隊，增加了 latency。通過將硬碟升級為 NVMe SSD，數據庫 server 的最大 IOPS 提升，能夠處理更多的同時操作，減少請求排隊等待的時間，降低了總體 latency。

#### 使用 RAID 或 ZFS 增加 IOPS

使用 RAID 或 ZFS 來增加 IOPS 也是提高 throughput 和降低 latency 的有效方法。

- RAID：
  - RAID 0：將數據分散寫入多個磁碟，增加 IOPS 和 Throughput，但沒有數據冗餘。
  - RAID 1：數據鏡像寫入多個磁碟，提高讀取 IOPS，但寫入性能與單個磁碟相同。
  - RAID 5/6：通過分散數據和奇偶校驗位來提高 IOPS 和提供數據冗餘。
- ZFS：
  - ZFS 的 vdev 結構允許將多個磁碟組合在一起，形成一個邏輯存儲池。這可以提高 IOPS 和 throughput，並提供強大的數據保護功能。
  - 使用 ZFS 的 RAID-Z 可以提高數據安全性，同時增加讀寫性能。

通過使用 RAID 或 ZFS，我們可以在一個存儲池中結合多個磁碟的 IOPS，提升整體系統的 throughput 和降低 latency。

## 總結

理解 throughput 與 latency 的關係對於解決高 latency 問題至關重要。透過增加處理能力和優化內部處理，可以有效減少 latency，提高系統的穩定性和性能。希望這篇文章能幫助你更好地理解和應用這些概念，解決實際工作中的 latency 問題。
