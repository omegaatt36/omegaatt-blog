---
title: Codewars Find X II in golang
date: 2020-03-03
tags:
  - golang
  - codewars
categories:
 - develop
---

想不到一眨眼半個月過去了，自從面試趨勢之後整個人彷彿被電到失去信心，在離職與等待面試消息期間，也只好振作補足不足的地方繼續向前。

而 docker 學習心得何時會產出呢？在學習過程中發現除了一般操作(day1 內容)外，底層或多 container 溝通都是大量的網路觀念與知識，需要更多知識量才能整理成筆記，不急不急。

由於學習過程發現而這段時間在 coding 能力上主要還是靠 codewars 上的題目，畢竟解完後能看別人的 code 來檢視自己的 code 能不能更好。其中不免遇到一些覺得很有趣的題目，好比這題 [Find X II](https://www.codewars.com/kata/5d339b01496f8d001054887f)，他並不是什麼過分難的題目，邏輯也不是很難通，畢竟只是一個 6 kyu 的題目，但為何快一年了只有 44 個人解完呢，這就牽扯到下面要說的。

# 題目概要

有下一段程式碼:

``` go
func FindX(n int) int {
  if n == 0 {
    return 0
  }
  x := 0
  for i:=1; i<=n; i++ {
    x += FindX(i-1) + 3*i
  }
  return x
}
```

但 x 範圍為 1 <= n <= `10**6(1e6)` ，請試著重構。且當 n 愈來試大時，可能會超過 int64 表達，故需要將結果對 `10**9 + 7` 取模

標記 `FUNDAMENTALS`  `OPTIMIZATION`
# 解題邏輯

若沒有取模限制的話這題可以化簡為 O(1)，但很明顯不行。於是先將原題目化簡為 O(n):

``` go
func FindX(n int) int {
    mod := int(math.Pow10(9)) + 7
	m := 0
	for i := 0; i <= n; i++ {
		m = (m*2 + 3*i) % mod
	}
	return m
}
```

好測試過，就算是輸入 1e6 進去也很快就算出來了，於是提交答案

 `Execution Timed Out (12000 ms)`
什麼，怎麼會超時，於是，我卡關了。

# 最終優化

想盡辦法針對該程式優化，還去查了 golang 中有沒有提升 mod 效能的作法，多次嘗試後仍然超時...。

![](https://img.itw01.com/images/2018/03/11/08/1629_Np7DWm_PM3HETA.jpg!r800x0.jpg)

會不會是因為測資不只一個 1e6，假設有幾千個隨機值，於是做了一個簡易的快取機制:

``` go
var m = []int{0}
var mod = int(math.Pow10(9)) + 7

func FindX(n int) int {
  if len(m) > n {
    return m[n]
  }
  for i := len(m); i <= n; i++ {
    m = append(m, (m[i-1]*2+3*i)%mod)
  }
  return m[len(m)-1]
}
```

測試通過，在看一下其他兩個用 golang 做出來的，其實解法也就大同小異了。

# 後記

從 php 轉學 golang 會許效能會提升，但不會讓你把 O(n**2) 的問題變成 O(n)，更甚說不會讓 NP 問題變簡單。

這題的觀念回頭過來看十分簡單，不就是一個全域變數嗎？但很多時候並不能每次都用空間換時間，特別是最近學 bitwise 後更有感。

但就單論這題，再如何優化，不靠快取機制，絕對會超時，也並非這題作者想讓大家學到得觀念了。
