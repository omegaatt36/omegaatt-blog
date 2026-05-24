---
title: A=B 遊完心得、解法
date: 2021-10-13
categories:
  - develop
tags:
  - game
---

Steam 上推出了一款 [A=B](https://store.steampowered.com/app/1720850/AB/)，稍微玩了一下，下面提供自己的解法。

小試了一下，這個迷你語言每一行由左側、等號`=`、右側組成，左側會被取代成右側，並且每次取代完都會從第一行重新開始，有點像在寫遞迴。

# 第一章 A=B

- 1-1
  a取代成b，限定一行。
  ```
  a=b
  ```
- 1-2
  變成大寫，限定三行。
  ```
  a=A
  b=B
  c=C
  ```
- 1-3
  去掉重複的，限定三行。
  ```
  aa=a
  bb=b
  cc=c
  ```
- 1-4
  去掉連續的a，限定兩行。
  ```
  aaa=aa
  aa=
  ```
- 1-5
  只會給 a 與 b，回答最多的是 a 還是 b，最多四行。
  ```
  ab=
  ba=
  aa=a
  bb=b
  ```
- 1-6
  排序，最多三行。
  ```
  ca=ac
  cb=bc
  ba=ab
  ```

# 第二章 新的關鍵字

多了一個新的關鍵字 `return`，僅能修飾右側語法，並且要用`()`包起來

- 2-1
  回答 helloword，最多一行。
  ```
  =(return)helloworld
  ```
- 2-2
  是否至少包含三個 a，最多四行。
  ```
  b=
  c=
  aaa=(return)true
  =(return)false
  ```
- 2-3
  字數除以 3 的餘數，最多六行。
  ```
  c=a
  b=a
  aaa=
  aa=(return)2
  a=(return)1
  =(return)0
  ```
- 2-4
  ```
  ca=ac
  ba=ab
  cb=bc
  aaa=a
  bbb=b
  ccc=c
  aa=(return)false
  bb=(return)false
  cc=(return)false
  =(return)true
  ```
- 2-5
  ```
  aaa=aa
  aa=d
  bbb=bb
  bb=d
  ccc=cc
  cc=d
  ab=(return)false
  ac=(return)false
  ba=(return)false
  bc=(return)false
  ca=(return)false
  cb=(return)false
  d=
  a=(return)true
  b=(return)true
  c=(return)true
  =(return)false
  ```
- 2-6
  ```
  ba=ab
  ca=ac
  cb=dc
  bc=d
  bd=db
  ad=
  a=(return)false
  b=(return)false
  c=(return)true
  =(return)false
  ```
- 2-7
  ```
  ba=ab
  ca=ac
  cb=bc
  ab=d
  ad=da
  da=aa
  ac=
  bc=
  d=b
  cc=c
  bb=b
  aa=a
  ```
