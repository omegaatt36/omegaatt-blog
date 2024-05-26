---
title: 那些在 Backend Sharing 中出現的人事物
date: 2024-03-24
categories:
 - develop
tags:
 - personal
aliases:
 - "/blogs/develop/2024/backend_sharing_note.html"
---

《最高學以致用法》、《最高學習法》這兩本書是我在 2023 年上半年讀了覺得挺有意思的書，核心概念就是「產出」，例如唸書時能夠回答同學問題的，肯定都已經精通該知識點了。

加入 KryptoGO 後，因為團隊的成長，Leader 開始嘗試舉行兩週一次的 Backend Sharing，不僅是分享工作上遇到的疑難雜症，更可以交流不同的知識點。

起初，我可以我分享了一些過去用過的工具與知識點，隨著時間的流逝，開始感受到黔驢技窮，於是我也仿效了《刻意練習》，不斷的為了能有更好的分享品質而學習。

以下是這半年來我在 Backend Sharing 中或多或少提及或是討論到的，又分為解決方案、小工具、方法論。

### 小工具

由於喜歡折騰 Homelab，時不時會到 [r/selfhosted](https://www.reddit.com/r/selfhosted/)、[r/opensource/](https://www.reddit.com/r/opensource/) 尋找一些開源的自架方案或小工具，有一些大幅度的改善了我的開發流程，而有一些則漸漸的不再使用

#### exa & eza & bat

[eza](https://github.com/eza-community/eza)、[exa](https://github.com/ogham/exa)、[bat](https://github.com/sharkdp/bat) 都是基於 rust 寫成的 cli 替代品，`exa`、`eza` 對標 `cd`，而 `bat` 對標 `cat`，我會在 `.zshrc` 中寫上 alias。

```bash
if command -v bat &> /dev/null; then alias cat=bat; fi
if command -v eza &> /dev/null;
then
  alias ls="eza --icons"
  alias ll="eza --icons -lh"
  alias tree="eza --icons --tree"
fi
```

還有其他性質相同的 rust 寫的小工具諸如： [zoxide](https://github.com/ajeetdsouza/zoxide)、[topgrade](https://github.com/topgrade-rs/topgrade)、[alacritty](https://github.com/alacritty/alacritty)、[bottom](https://github.com/ClementTsang/bottom)

---

#### fzf

[fzf](https://github.com/junegunn/fzf) 主要是做模糊搜索，相關應用還挺多的，主要都是拿 fzf 來當作模糊搜尋模塊：

- [ytfzf](https://github.com/pystardust/ytfzf)：搭配 mpv，可以直接在 cli 播放音樂。
- [ani-cli](https://github.com/pystardust/ani-cli)：ytfzf 同個作者，可以直接在 cli 播放音樂。
- [kubectl-fzf](https://github.com/bonnefoa/kubectl-fzf)：取代 `k get pods` 等等操作。

---

#### Omnivore App

一個免費的 Read it later 服務，開源可以自架。目前仍有在使用，也有成功推廣給同事，~~特別是 Line 宣佈 Keep 功能下線後~~。

---

#### tmux

不用依賴 terminal emulator 的 split 功能。遠端管理伺服器執行 lone live 腳本時不用擔心 ssh 斷線，甚至可以看到同事在編輯哪個檔案。
順便分享我目前的[配置來源](https://www.youtube.com/watch?v=DzNmUNvnB04)

---

#### distrobox

[distrobox](https://github.com/89luca89/distrobox) 大幅降低了 distro hopping 的困擾，例如可以在 debian host 上，安裝最新的 yay(arch linux) 的執行檔，並且使用起來有「原生」的體驗（類似 WSL 之於 Windows）。

---

#### gopatch

一開始在大規模改寫時，尋找 patching 的替代方案，發現了 uber 開源的 [gopatch](https://github.com/uber-go/gopatch)，但後來覺得 GoLand 的重構功能更好用，就沒有再使用了。

---

#### [littlelink](https://github.com/sethcottle/littlelink)

在公司要做名片時，提出了可以放 QR code 在平片上。貪心的人怎麼可能只放一個 URL 到名片上，當然是「我全都要」。

---

### 解決方案

#### sync.singleflight

[singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight) 是一個 thread safe、避免 cache miss 後所有連線湧入 db 的防擊穿寫法，其中處理 panic or context dead 的寫法也是有一點含金量，~~以後會專門出一期節目。~~

---

#### gormigrate

database migration as code，[gormigrate](https://github.com/go-gormigrate/gormigrate) 是一個簡單的 gorm migrator 的 wrapper，方便在程式內寫 migration，例如某個版本要到某某網站拉某某資料等等。

---

#### Gomock

學會 DI 的下一步就是撰寫假實做，可以透過 stub,fake,mock 等等，而 [golang/mock](https://github.com/golang/mock) 則是官方推出的 mock code gen，經歷了一段時間沒人維護後，由 uber 接手 fork 了一份 [uber-go/mock](https://github.com/uber-go/mock)

---

#### go-enum

golang 是一個以簡單性為核心設計的語言，許多 proposal 討論了幾年後，也可能因為官方覺得不太好而 closed。enum 沒有出現在 golang 的原生支援中，但多人協作時仍可以使用 [go-enum](https://github.com/abice/go-enum) 來進行 code gen，幫助維持程式碼的一致性。

---

#### Golang Generics

舉個最簡單的例子，過去常常會需要寫 `util.IntPtr(position)、util.FloatPtr(position)` 等等，如今可以使用 `util.Ptr(xxx)` 來解決程式碼的複雜度。

相對的也有會有許多坑，可能導致的[效能 issue](https://hackmd.io/@fieliapm/BkHvJjYq3)

---

#### Go Module 和 Workspace

[go workspace](https://go.dev/blog/get-familiar-with-workspaces) 解決了多專案共用 module 卻需要反覆修改 go.mod 引發的種種困擾。

在編寫所有 repo 都會用到的 backend-common 時很好用，除此之外沒有特別需要使用的場景。

---

#### Docker 在 macOS 上的應用

單純是看到 [Colima](https://www.linkedin.com/pulse/colima-better-way-run-docker-macos-linux-asutosh-pandya)，希望使用 macOS 的同事可以在本地順暢地、更沒有負擔地使用 docker container 來加速開發與除錯。

---

#### Web GUI DB Admin 工具

[cloudbeaver](https://github.com/dbeaver/cloudbeaver) 取代了多個 admin 的功能，又比 [adminer](https://www.adminer.org/) 多出了連線管理功能，記憶體使用也比 Desktop App 低上不少。

---

#### Implicit Memory Aliasing in Go

有寫[一篇專門在講這塊](/blogs/develop/2023/golang_implicit_memory_aliasing)，不在此贅述。

---

#### Gin context data race

針對[這篇文章來討論](https://stackoverflow.com/questions/73762584/why-should-i-make-a-copy-of-a-context-for-goroutines-inside-handlers)，主要是 gin 使用了 radix tree 來提昇效能，但會綁 middleware 到上層的 node，或許可以看看 go 1.22 中最新的 http router 會如何解決這類問題。

---

#### wakatime

[追蹤自己寫成是的時間](/blogs/develop/2023/wakatime_experience)，不在此贅述。

---

### 方法論

#### Clean Architecture in Go

---

#### Build your own X（a.k.a. 造輪子）

來源：

- [https://www.youtube.com/watch?v=WITxT9Da5s8](https://www.youtube.com/watch?v=WITxT9Da5s8)
- [https://github.com/codecrafters-io/build-your-own-x](https://github.com/codecrafters-io/build-your-own-x)
- [https://app.codecrafters.io/tracks/go](https://app.codecrafters.io/tracks/go)

好處：

- 刻意練習（O）、白嫖 side projects（X）
- 可以強迫自己寫 MVP，也可以強迫自己練習「玩」演算法、資料結構、設計模式、架構
- 練習 TDD，再參考 code crafters 的測試案例
- 快速學習另一個程式語言（？）
  - 提供 guild

壞處：

- 「自己把問題拆細的能力」仍是依賴別人，自己要找其他方法練習
- 題目偏純後端，若要提昇全端能力，還是要其他方法
- build my own redis

嘗試自己實做 redis server：[https://github.com/omegaatt36/my-redis](https://github.com/omegaatt36/my-redis)

---

#### Code Complete II 書籍内容

透過閱讀與分享，共享這本堪稱軟體開發聖經的書籍。

---

知識無價，共勉之。
