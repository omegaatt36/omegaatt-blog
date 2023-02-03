---
title: Golang 學習筆記 Day 0 - Why Golang & 安裝篇
date: 2020-08-25
categories:
 - develop
tags:
 - golang
---

``` 
免責聲明，此篇為個人學習筆記，非正式教學文，僅供參考感謝指教，有任何內容錯誤歡迎發 issue。
```

![image source: https://medium.com/nordcloud-engineering/how-to-build-pluggable-golang-application-and-benefit-from-aws-lambda-layers-154c8117df9b](https://miro.medium.com/max/2946/1*TArmNwMaoXjR1MegmucwyQ.png)

## Why Golang

&emsp; &emsp; 在高中時期，除了一年級上學期是用 C 打開程式大門之外，大家都知道的，畫金字塔、 `cin` 、 `cout` 等等基本的邏輯觀念，一年級下學期開始就是 C#、HTML、PHP，也玩轉在 CPLD、Arduino、焊三角錐、兩根同軸焊接，總之就是全部吃掉就對了，雖然都是皮毛但不至於完全沒印象。

&emsp; &emsp; 然而到了大學，敝人不才沒有考上一般大學，也沒考上工科第一二志願科大，在整個大學過程碰了最多的反而是 python、machine learning。然而也都是碰皮毛~~因為都在玩社團、學攝影~~，事實上在大學時期程式功力已荒廢大半。

&emsp; &emsp; 所幸在實習期間，公司願意提供環境讓我邊學習邊成長後做出貢獻，而使用的後端語言是 PHP，然而就一年以來的使用 PHP 有以下優缺點：

* 學習曲線~~極~~低，但坑很多
* 解釋型腳本語言部署極方便
* 非靜態、非強行別、萬用 array 導致容易出現很多偷懶的 code
* 窮人聖經 LAMP、LNMP
* 歷史悠久 framework 眾多、工作好找，甚至有許多 C 底層的高校框架可以使用。
* 可以很 OO 也可以不 OO，寫 code 彈性極大
* 在 php 7.0 改用 zend 引擎後效能改頭換面，「原來 PHP 也可以寫遊戲」。搭配一些 C 底層框架，甚至可以跟 nodeJS 一較高下，而期望中 8.0 效能將再更高。

![](https://i.imgur.com/CkKvVrE.jpg)

&emsp; &emsp; 但是，再快也只是 AE86，腳本語言終究與「高效」存在矛盾，同時，與其在 PHP 中配合 IDE 使用 type hinting 自我約束，何不直接寫強行別語言呢?? 然而在看過許多主流 web service 語言後，最後決定來學學 Golang，而關於 golang 的優勢，可以點擊下方圖片看更多詳細。

[![](https://cdn-media-1.freecodecamp.org/images/1*NDXd5I87VZG0Z74N7dog0g.png)](https://www.freecodecamp.org/news/here-are-some-amazing-advantages-of-go-that-you-dont-hear-much-about-1af99de3b23a/)

[![](/assets/dev/20200825/golang_download_page.png)](https://golang.org/dl/)

## Install for windows

Step.1 首先到[官網](https://golang.org/dl/)點擊 Featured downloads 中的 Micoro\$oft Windows，下載完後雙擊開啟 => 下一步直到安裝完成 Finish。

Step.2 若安裝正常則 go 已經加進 windows 的環境變數中了。可以開啟 CMD 輸入 `go` 可以看到 go 的 cli 指令，而輸入 `go version` 則可查看目前安裝的版本。

Step.3 更改環境變數 GOPATH

而目錄空格有時候會搞出很多不必要的毛(在 python 被深深傷害)，於是去系統環境變數找到 GOPATH 並更改為一個沒有空格的目錄。也將做為未來撰寫 golang 時的 worksapce。

## Install for linux

``` 
這邊環境為 wsl2 / ubuntu 20.04
```

Step.1 同樣到[官網](https://golang.org/dl/)下載 binary 檔並解壓縮:

``` bash
$ wget  https://golang.org/dl/go1.15.linux-amd64.tar.gz
```

Step.2 解壓縮後放到你要放的資料夾

``` bash
$ sudo tar -xvf go1.15.linux-amd64.tar.gz
# $ sudo mv go [PATH]
$ sudo mv go /usr/local
```

Step.3 在 `.profile` 或諸如 `.zshrc` 中設定環境變數， `source` 後即可使用

``` bash
# $ vi .profile
$ vi .zshrc
```

``` vim
# in the .zshrc
export GOPATH=/usr/local
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$GOPATH/bin
```

``` bash
# $ source .profile
$ source .zshrc
```

最後測試一下，輸入 `go version` 確認一下是否能正常執行 golang 指令

``` bash
$ go version
'go version go1.14.4 linux/amd64
```

## Hello world!

許多語言或許都在爭誰能夠「最快」印出 Hello world，我想 golang 在這方面絕對是沒有 python 來的「方便」呢。

首先先創建一個 `hello-world` 資料夾，並在裡面建立一個 `main.go` ，並複製貼上下面的程式碼。

``` go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, playground")
}
```

儲存後在 cmd 中輸入 `go run main.go` 就會看到字串 `Hello, playground` 被打印出來囉。同樣的程式也可以在 [playground](https://play.golang.org/) 中執行。

## Hello world 步驟解析

這段落給初學者們，若您是超級初心者，可以先落過這段落去看下一篇。

golang 中一個資料夾就是一個 `package` (若不熟悉什麼是 `package` 的話日後會再詳細說明)，於是在程式碼第一行中一定要宣告這支檔案是什麼 `package` 。由於 golang 中一個資料夾中只允許存在一個 `main package` ，於是新建立一個 `hello-world` 資料夾是為了避免在同一個資料夾內有好幾個 `main package` 。

golang 中所有的程式碼檔案都是以 `.go` 做為副檔名，於是建立一個 `main.go` 做為 `main package` 的「入口」，換句話說，也可以命名成 `hello-world.go` ，只要在執行時替換成 `go run hello-world.go` 就可以了。

接著來看程式碼：

``` go
// 這支檔案是 main package
package main

// 我們要打印字串，於是需要使用到 fmt 這個 libary，於是在這裡「引用」
import (
	"fmt"
)

// 程式的「入口」會從 main() 「主函式」開始執行
func main() {
    // 透過 fmt 這個 libary 中的 Println 來打印字串
	fmt.Println("Hello, playground")
}
```
