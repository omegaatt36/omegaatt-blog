---
title: "Unrounded Scaling：Go 1.27 浮點數轉換演算法的深度拆解"
date: 2026-08-26
categories:
  - develop
tags:
  - golang
  - algorithm
  - optimization
cover:
  image: "images/cover.webp"
---

## 前言

開始之前先看這支[影片](https://www.youtube.com/watch?v=g1r3iLejTw0&list=PLwy0WTzBokTMHqq_-TQRNmaHWEuPgsGUc)，講的是 Quake III 那個 `0x5f3759df`。它跟本文沒有直接關係，但用的是同一類招式：把浮點數的 bit pattern 當成整數讀，用整數的乘加移位取代超越函數。差別在於那一招接受 3.4% 的誤差，而本文接下來要講的東西，一個 bit 的誤差都不能有。這個對照後面[定點對數](#定點對數先確定答案有幾位數)那節會再回來收。

<iframe width="560" height="315" src="https://www.youtube.com/embed/g1r3iLejTw0?si=bRcAcBeUxO5bfX-m" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

TL;DR 中間一大段是演算法拆解，有興趣再啃。只想知道「Go 1.27 實際上改了什麼、變多快、輸出有沒有跑掉」的話，可以直接跳到 [Go 1.27 到底改了什麼](#go-127-到底改了什麼)，結論是 878 行三套演算法變成 290 行一套、快了 11–21%、輸出零變化。如果你在意的是「那我平常記帳該用什麼」，跳到 [附錄](#附錄不用-float64-的話代價是什麼)，那一節比較了 `float64`、`math/big` 三兄弟與 `shopspring/decimal`。

我把 Go 升到 1.27 之後，照慣例翻了一遍 [release notes](https://go.dev/doc/go1.27)，也寫了[新特性整理](/blogs/develop/2026/golang_1_27_new_features/)。整份文件從頭到尾沒有一個字提到 `strconv`。

但是 Go 1.27 把 `fmt.Sprintf("%v", f)`、`strconv.ParseFloat`、`json.Marshal` 底下那顆浮點數轉換引擎整個換掉了。舊的實作是三套演算法各自為政：最短列印用 Dragonbox、固定位數列印用 Ryū 風格的程式碼、解析用 Eisel-Lemire，新的實作是一個叫 unrounded scaling 的東西，作者是 Russ Cox，論文是 2026 年 1 月才發表的。

這件事之所以值得寫一篇，是因為 Russ Cox 在 2011 年寫過一篇 [Floating Point to Decimal Conversion is Easy](https://research.swtch.com/ftoa)，副標題是 Part 1。文章的論點是：

> Floating point to decimal conversions have a reputation for being difficult. At heart, they're really very simple and straightforward.

那篇的實作用的是十進位字串上的長除法，每次把 buffer 乘 2 或除 2，跑 `e` 次。對 `float64` 最小的次正規數來說 `e` 是 1074，所以要跑一千多輪，每輪掃過幾百個 byte。簡單，但是慢得離譜。

十五年後的 2026 年 1 月，Part 3 [Floating-Point Printing and Parsing Can Be Simple And Fast](https://research.swtch.com/fp) 開場第一段就是自己打自己臉：

> My 2011 post "Floating Point to Decimal Conversion is Easy" argued that these conversions can be simple as long as you don't care about them being fast. But I was wrong: fast converters can be simple too, and this post shows how.

這篇文章要做的事情，是把 unrounded scaling 這個演算法從頭拆到尾：它的數學基礎是什麼、為什麼可以只用一次 64 位乘法、Go 1.27 到底改了哪些檔案，以及我在自己機器上實測出來的數字。中間會繞去看一下 IEEE 754 之前的浮點數長什麼樣、最早的人怎麼做轉換，最後有一節附錄比較 `math/big` 與 `shopspring/decimal`，說明「不用 `float64`」這條路的代價。文章裡所有的程式碼我都跑過，所有的數字都是實測的，範例程式碼放在 [uscale.go](uscale.go)、[main.go](main.go) 與 [appendix_demo.go.txt](appendix_demo.go.txt)。

## 這個問題到底難在哪

一個 `float64` 的形式是 `f = m · 2^e`，`m` 是 53 位的整數 mantissa，`e` 是指數。人類要讀的是十進位，所以需要在 `m · 2^e` 與 `d · 10^p` 之間互轉。

聽起來像是國中數學，難點在於要同時滿足三件事：

一、正確捨入（correct rounding）。輸出必須是「最接近原值的那個十進位數」，不能差半個 [ULP(Unit in the Last Place)](https://en.wikipedia.org/wiki/Unit_in_the_last_place)。

二、最短（shortest）。`0.1` 應該印成 `0.1`，不是 `0.1000000000000000055511151231257827`。而且最短的那個結果 parse 回來必須還是原本那個 `float64`，這叫 round-trip。

三、快（fast）。`json.Marshal` 一秒可能要做幾百萬次。

### 這三件事，哪些是 IEEE 754 規定的

我原本以為三件事都出自 IEEE 754，翻了標準才發現只有一件半。

正確捨入是硬性規定。[IEEE 754-2019](https://standards.ieee.org/ieee/754/6210/) 的 §5.12.2「External decimal character sequences representing finite numbers」講得很直接：

> Within the limits stated in this clause, conversions in both directions shall preserve the value of a number unless rounding is necessary and shall preserve its sign. If rounding is necessary, they shall use correct rounding and shall correctly signal the inexact and other exceptions.

round-trip 也算標準保證的，但它的講法不是「最短」，而是「位數給夠就保證拿得回來」。同一節的 NOTE 1：

> Conversions from a supported binary format bf to an external character sequence and back again results in a copy of the original number so long as there are at least Pmin (bf) significant digits specified […]

標準直接列出 binary64（也就是 `float64`）的 `Pmin` 是 17，其他二進位格式則給公式：


$$
P_{\min}(bf) = 1 + \left\lceil p \cdot \log_{10} 2 \right\rceil
$$

<br/>

代入 `float64` 的 53 位 mantissa 得到 `1 + ⌈15.95⌉ = 17`，跟列出來的值一致。這就是後面講最短列印時「17 位一定夠」的出處。

至於「最短」，IEEE 754 完全沒有要求。我把 754-2019 全文用 `pdftotext` 抽出來搜過，`shortest` 出現 0 次，連 `round-trip` 這個詞本身都沒出現過，標準用的是上面那句「converted ... and back again results in a copy of the original number」。

所以「最短」是額外加碼的目標，來源是 Steele 與 White 1990 年在 PLDI 發表的 [How to Print Floating-Point Numbers Accurately](https://dl.acm.org/doi/10.1145/93542.93559)，也就是 Dragon4 那篇。Go 把它寫進了 API 契約，[`strconv.FormatFloat`](https://pkg.go.dev/strconv#FormatFloat) 對 `prec = -1` 的說明是：

> The special precision -1 uses the smallest number of digits necessary such that ParseFloat will return f exactly.

「快」則不在任何標準裡，純粹是現實需求。

還有一個轉折：正確捨入是到 2008 年版才變成無條件要求的。1985 年初版的 §5.6 只在有限範圍內要求正確捨入，超出範圍時允許誤差比正確答案多 0.47 ULP：

> for rounding to nearest, the error in the converted result shall not exceed by more than 0.47 units in the destination's least significant digit the error that is incurred by the rounding specifications of Section 4

那個 0.47 的出處，1985 版的註腳寫得很明白，是 Coonen 1984 年的博士論文。這正好接上後面時間軸裡 1984 年那一列。換句話說，「浮點數轉字串必須完全正確」這件事被寫死進標準，比大部分人以為的晚了 23 年。

只滿足前兩個很簡單，用大整數硬幹就好，這就是 2011 年那篇的做法。困難的是加上第三個條件，因為當 `p` 是負數時，`10^p` 在二進位裡是無限循環小數，任何有限精度的近似都會引入誤差，而你必須證明這個誤差不會影響最後捨入的那一位。

這個「證明」花了業界大約 79 年。

## 在 IEEE 754 之前，浮點數不是一種格式

前一節講到「正確捨入到 2008 年才變成無條件要求」，單看很像標準委員會怠惰。把時間軸往前拉就會發現，1985 年之前根本沒有一個叫做「浮點數」的東西可以標準化，只有一堆長得不太一樣的浮點數。

### 1941：Z3 已經在用隱含最高位

Konrad Zuse 的 [Z3](https://zuse.zib.de/reconstructionZ3)（1941）通常被算作第一台能運作的可程式化數位計算機，而它從一開始就是用二進位浮點數的。字長 22 bit，1 bit 符號、7 bit 指數、14 bit 尾數。

真正有意思的是它已經用了「隱含最高位」這個技巧：二進位正規化之後尾數的最高位必然是 1，那就不必存，14 bit 的欄位實際上表達 15 bit 的精度。八十五年後 `float64` 用 52 bit 的欄位表達 53 bit 尾數，用的是同一招。本文後面會看到 `unpack64` 回傳的 `m` 保證最高位是 1、`bits(m)` 恆為 64，也是同一件事的延伸。

### 1954：浮點數才開始有硬體

Z3 之後有很長一段時間，浮點數是用副程式做出來的：整數硬體加上一段負責對齊指數與正規化的程式碼。[IBM 704](https://en.wikipedia.org/wiki/IBM_704)（1954）是第一台把浮點運算做進硬體而且量產的機器，36 bit 字長，FORTRAN 跟 LISP 都是先在它上面誕生的。

浮點數進了硬體之後，事情反而變複雜，因為每家廠商都自己決定要怎麼做。

### 1964：底數之爭，以及 wobbling precision

IBM System/360（1964）用的不是 base-2，是 base-16。值的形式是 `0.f · 16^(e-64)`，短格式的 `f` 有 24 bit。

用 16 當底數的好處是指數欄位可以短一點、對齊時的移位次數少（一次移 4 bit），壞處是正規化只能保證「最高的那個十六進位數字非零」，而一個非零的十六進位數字可能長得像 `0001`。也就是說尾數最前面可能白白浪費掉 3 個 bit。

我把幾個值在這個格式底下的實際有效位數算出來：

```
值        十六進位尾數   有效 bit 數
1         0x100000       21
1.9       0x1E6666       21
2         0x200000       22
4         0x400000       23
8         0x800000       24
15.9      0xFE6666       24
```

同一個 24 bit 的欄位，有效精度在 21 到 24 bit 之間游走，這就是所謂的 wobbling precision，相對誤差會隨著數值落在哪個十六進位區間而變動 16 倍。IEEE 754 用 base-2 就沒有這個問題，正規化之後最高位必定是 1，24 bit 就是 24 bit。

同一台機器上還有另一個著名的坑：雙精度格式一開始沒有 guard bit。Goldberg 那篇 [What Every Computer Scientist Should Know About Floating-Point Arithmetic](https://docs.oracle.com/cd/E19957-01/806-3568/ncg_goldberg.html) 記了這件事：

> Guard digits were considered sufficiently important by IBM that in 1968 it added a guard digit to the double precision format in the System/360 architecture (single precision already had a guard digit), and retrofitted all existing machines in the field.

回收所有已出貨的機器改硬體。這個代價大概可以說明那幾個 bit 有多要緊，而本文接下來要講的 unrounded number，偷的正是 guard／round／sticky 這一組 bit 的概念。

### 所以「正確的十進位輸出」在當時沒有定義

把上面幾件事放在一起看，1985 年那個 0.47 ULP 的容忍度就沒那麼難理解了。底數不一樣、guard bit 有沒有不一樣、捨入規則不一樣，你連「原本那個數到底是多少」都不見得能在兩台機器之間講清楚，要求轉字串完全正確在當時是奢侈的。要等到大家都用同一個格式二十幾年之後，2008 年才有本錢把那句話改成無條件。

## 最早的人是怎麼轉換的

那在這片混亂裡，最早的人是怎麼把二進位印成十進位的？

Russ Cox 追溯到的最早一筆是 Goldstine 與 von Neumann 1947 年的《Planning and Coding of Problems for an Electronic Computing Instrument》：

> They converted one digit at a time by repeated multiplication by 10 and modulo by 1, as did many conversions that followed.

（順帶一提，原文把名字寫成 Goldstein，正確拼法是 Goldstine，就是那位 ENIAC 的軍方聯絡官 Herman H. Goldstine。）

方法本身一句話講得完：把值壓進 `[0.1, 1)`，然後反覆「乘 10、取整數部分當作下一位、把整數部分減掉」，想要幾位就跑幾輪。我用 `big.Rat` 重現了一次：

```go
func gvn1947(f float64, n int) (digits []byte, muls int) {
    r := new(big.Rat).SetFloat64(f)
    ten := big.NewRat(10, 1)
    one, tenth := big.NewRat(1, 1), big.NewRat(1, 10)
    for r.Cmp(one) >= 0 { // 先壓進 [0.1, 1)
        r.Quo(r, ten)
        muls++
    }
    for r.Cmp(tenth) < 0 {
        r.Mul(r, ten)
        muls++
    }
    for i := 0; i < n; i++ {
        r.Mul(r, ten)
        muls++
        d := new(big.Int).Quo(r.Num(), r.Denom()) // 取整數部分
        digits = append(digits, byte('0'+d.Int64()))
        r.Sub(r, new(big.Rat).SetInt(d)) // modulo 1
    }
    return digits, muls
}
```

跑出來：

```
3.141592653589793 → 31415926535897931  （做了 18 次大數乘除法）
0.1               → 10000000000000000  （做了 17 次大數乘除法）
1e+300            → 10000000000000000  （做了 318 次大數乘除法）
1e-300            → 10000000000000000  （做了 316 次大數乘除法）
```

答案是對的，問題全在最後一欄。要的位數是常數 17，但把值壓進 `[0.1, 1)` 的前置作業要跑 `⌊log₁₀ x⌋` 輪，`1e300` 就是 300 輪，而且每一輪都在對一個越滾越大的分數做乘除。實測：

| 輸入 | 1947 的方法 | Go 1.27 `strconv` | 倍數 |
|---|---|---|---|
| π | 8342 ns/op、238 allocs | 47.68 ns/op、1 alloc | 175x |
| `1e300` | 762227 ns/op、3187 allocs | 51.62 ns/op、1 alloc | 14766x |
| `1e-300` | 817794 ns/op、2581 allocs | 63.83 ns/op、1 alloc | 12812x |

但真正的差別不是倍數，是形狀：1947 的方法成本跟指數的絕對值成正比，`strconv` 不管指數多大都差不多。

而它們在數學上做的是同一件事。`gvn1947` 每乘一次 10 擠出一位，`uscale(m, e, p)` 一次把 `10^p` 乘完，17 位一起擠出來。中間那 79 年的所有論文，本質上都在回答同一個問題：一次乘完的話，誤差還控制得住嗎？

## 一段被壓縮的歷史

那 79 年裡發生了什麼事，Russ Cox 的原文有一整節在追溯每個想法的來源，整理成時間軸：

| 年份 | 人物 | 貢獻 |
|---|---|---|
| 1947 | Goldstine、von Neumann | 最早的二進位／十進位轉換，重複乘 10 一位一位擠出來 |
| 1966 | Mancino | 指出查表法（預存 10 的冪次）兩個方向都能用 |
| 1970 | Len Harding / BCC Model 1 | sticky bit 誕生，硬體第一次做到正確捨入 |
| 1984 | Coonen | IEEE754 實作指南，用三次乘法組出任意 10 的冪次，給出誤差分析 |
| 1990 | Steele & White | Dragon2/Dragon4，第一個正確的最短列印演算法，用 bignum |
| 1990 | Clinger | 正確的解析演算法 |
| 1990 | Gay | `dtoa.c`，可能是史上被抄最多次的 C 檔案 |
| 1990 | Slishman | 查表法 + carry bit 檢查，精度不夠時 fallback 回 bignum |
| 2004 | Hack | 證明 128 位精度足以讓解析完全不需要 fallback |
| 2010 | Loitsch | Grisu3，快，但約 0.5% 的輸入要 fallback |
| 2016 | Andrysco 等 | Errol3，用 106 位 double-double，只有 45 個輸入要查表特判 |
| 2018 | Adams | Ryū |
| 2018 | Giulietti | Schubfach，發現「選對 `p` 時最多只有一個候選以 0 結尾」 |
| 2020 | Eisel、Lemire | 128 位版本的 Slishman，成為 `fast_float` |
| 2024 | Jeon | Dragonbox |
| 2026 | Russ Cox | unrounded scaling |

看完這張表我的第一個感想是：所有零件在 2020 年就都在桌上了，只是沒人把它們拼起來。Russ Cox 自己也是這樣講的：

> My contribution here is primarily a synthesis of all this prior work into a single unified framework with a simple explanation and relatively straightforward code.

第二個感想是，這裡面有兩條線一直沒有交會。Slishman 1990 年在 IBM 做的是「查表 + carry bit」，Hack 2004 年證明了 128 位就夠，但兩個人都只做解析，沒有推廣到列印。而 Grisu、Ryū、Schubfach、Dragonbox 這條線都在做最短列印，卻沒有人用 carry bit 這個最佳化。Russ Cox 做的事情就是把這兩條線接起來，然後發現接起來之後，程式碼反而變短了。

## Unrounded number：跟 IEEE754 硬體偷來的兩個 bit

整個演算法的第一塊基石，是一個叫「未捨入數」的表示法。

浮點數硬體在做加減乘除的時候，規格上是「先用無限精度算出來，再捨入到最近的浮點數」。實際的硬體當然不可能有無限精度，它只多留三個 bit：guard、round、sticky。sticky bit 的性質是「一旦被設為 1 就永遠是 1」，這樣就能記住「後面還有東西沒算完」。

Russ Cox 把這個技巧搬到軟體，而且只需要兩個 bit。一個實數 `x` 的未捨入形式 `⟨x⟩` 定義為：`⌊x⌋` 的整數部分，後面接兩個 bit。第一個 bit 表示小數部分是否 `≥ ½`（half bit），第二個 bit 表示小數部分是否不恰好等於 0 或 ½（sticky bit）。整個表示法的二進位版面是：

```
⟨x⟩ = [ ⌊x⌋ 的整數部分 ][ half bit ][ sticky bit ]
                         └── 這兩個 bit 就是「跟硬體偷來的」
```

寫成公式就是：

$$
\langle x \rangle = \lfloor 4x \rfloor \mathbin{|} (4x \neq \lfloor 4x \rfloor)
$$

<br/>

公式裡的 `|` 是位元 OR。要理解它，關鍵是注意 `⌊4x⌋` 的最低兩位已經把小數部分的四種狀態完整編碼了（`4(x − ⌊x⌋)` 落在 `[0, 4)`，取地板後餘數就是這兩個 bit）：

```
00 = 小數是 0     （例如 6.0    → 24 = ⟨6.0⟩）
01 = 有小數、還不到 ½（例如 6.3   → 25 = ⟨6.0+⟩，half 沒亮 sticky 亮）
10 = 剛好在 ½      （例如 6.5    → 26 = ⟨6.5⟩）
11 = 超過 ½        （例如 6.9    → 27 = ⟨6.5+⟩）
```

所以那個布林值成立時只要把 bit 0 補成 1 就夠了，`⌊4x⌋` 本身只會在「小數恰好為 0 或恰好為 ½」時落到偶數上，這時確實沒有小數殘留；一旦 `x` 略微偏離邊界，OR 就把 sticky 點亮。不能改成加法：`6.3` 的 `⌊4x⌋` 已經是 25，加 1 變 26 會把它誤判成 `⟨6.5⟩`。

```go
type unrounded uint64

// bool2[T](b) 把布林轉成 0 或 1。
func unround(x float64) unrounded {
    return unrounded(math.Floor(4*x)) | bool2[unrounded](math.Floor(4*x) != 4*x)
}
```

我把它印出來看比較有感覺，用 `n.h` 加上 `+` 的格式表示，`h` 是 0 或 5，`+` 代表 sticky bit：

```
x          raw    ⟨x⟩
6          24     ⟨6.0⟩
6.001      25     ⟨6.0+⟩
6.499      25     ⟨6.0+⟩
6.5        26     ⟨6.5⟩
6.501      27     ⟨6.5+⟩
6.999      27     ⟨6.5+⟩
7          28     ⟨7.0⟩
```

注意 `6.001` 跟 `6.499` 的未捨入形式完全一樣，都是 `⟨6.0+⟩`。這正是重點：未捨入數不記錄「小數部分是多少」，只記錄「捨入的時候需要知道的資訊」。而這些資訊恰好只要兩個 bit。

有了這兩個 bit，五種捨入模式全部退化成「加一個常數再右移兩位」：

```go
func (u unrounded) floor() uint64         { return uint64((u + 0) >> 2) }
func (u unrounded) roundHalfDown() uint64 { return uint64((u + 1) >> 2) }
func (u unrounded) round() uint64         { return uint64((u + 1 + (u>>2)&1) >> 2) }
func (u unrounded) roundHalfUp() uint64   { return uint64((u + 2) >> 2) }
func (u unrounded) ceil() uint64          { return uint64((u + 3) >> 2) }
```

直覺是：`⟨n.f⟩` 存的值等於 `4n` 加上小數編碼（0 到 3），右移兩位就是砍掉小數只留 `n`。「加常數」則是把小數編碼先推過不同的門檻，floor 加 0（什麼門檻都不要）、half down 加 1、half up 加 2、ceil 加 3。拿實際數字走一遍最清楚：`⟨6.5+⟩` 的原始值是 27，round half down 做 `(27+1)>>2 = 7`；`⟨6.5⟩` 是 26，`(26+1)>>2 = 6`。同樣是「6.5」，sticky 一亮一半沾到，捨入結果就不同。

`round()` 是 IEEE754 的預設模式「round half to even」，1.5 跟 2.5 都進位到 2。以 round half down 為基礎，`(u>>2)&1` 讀出整數部分的最低位（「是不是奇數」），奇數就再多加一，把卡在 half 邊界的值推上去，推完整數部分變偶數，正好符合 half to even。

實測結果：

```
⟨x⟩        floor   round½↓   round   round½↑   ceil
⟨6.0⟩      6       6         6       6         6
⟨6.0+⟩     6       6         6       6         7
⟨6.5⟩      6       6         6       7         7
⟨6.5+⟩     6       7         7       7         7
⟨7.0⟩      7       7         7       7         7
⟨7.5⟩      7       7         8       8         8
⟨8.5⟩      8       8         8       9         9
```

`⟨6.5⟩` 的 `round()` 是 6 而 `⟨7.5⟩` 的 `round()` 是 8，就是 round half to even 在動作。

### sticky bit 為什麼非要不可

未捨入數還要支援除法跟右移，這時 sticky bit 的價值才真正浮現。

考慮這個情境：`15.4 / 6`。如果先把 15.4 捨入成整數 15，再除以 6 得到 2.5，round half to even 會給你 2。但正確答案是 `15.4 / 6 = 2.5666...`，應該進位到 3。

未捨入除法保住了這個資訊：

```go
func (u unrounded) div(d uint64) unrounded {
    x := uint64(u)
    return unrounded(x/d) | u&1 | bool2[unrounded](x%d != 0)
}
```

`u&1` 把原本的 sticky bit 傳下去，`x%d != 0` 補上這次除法產生的餘數。實測：

```
⟨15.4⟩          = ⟨15.0+⟩ (raw 61)
⟨15.4⟩.div(6)   = ⟨2.5+⟩ → round() = 3
先捨入再除:      round(15.4)=15, 15/6=2.5 → round = 2
```

`⟨2.5+⟩` 那個 `+` 就是答案。它告訴 `round()`：「這個 2.5 不是剛好 2.5，是比 2.5 大一點」，所以往上進位。這兩個 bit 就是整個演算法能夠「先算再決定怎麼捨入」的原因。

## uscale：整個演算法只有這一個原語

第二塊基石是一個叫 unrounded scaling 的運算：

$$
\operatorname{uscale}(x, e, p) = \langle x \cdot 2^{e} \cdot 10^{p} \rangle
$$

<br/>

給一個整數 `x`，乘上 2 的 `e` 次方跟 10 的 `p` 次方，回傳未捨入形式。就這樣。

本文接下來要講的三個演算法，固定位數列印、解析、最短列印，全部都建立在這一個函式上。這是整篇論文最漂亮的地方：三個看起來完全不同的問題，其實是同一個運算的三種呼叫方式。

先用一個小例子把 `uscale` 走完一輪，後面看到它就不陌生了。取 `x = 5, e = -3, p = 2`：

```
x · 2^e · 10^p = 5 · 2^-3 · 100 = 62.5
⟨62.5⟩ 的 raw 值 = 250（62·4=248，再加 half bit 2）
    round()       = (250+1+(62&1))>>2 = 251>>2 = 62
    roundHalfUp() = (250+2)>>2 = 63
```

也就是說 `uscale` 回傳的東西帶著「捨入該怎麼做」的全部資訊，呼叫端挑一個模式 Shift 一下就是答案。

用大整數寫一個顯然正確但很慢的版本當作參考答案：

```go
func uscaleBig(x uint64, e, p int) unrounded {
    num := new(big.Int).SetUint64(x)
    num.Mul(num, big.NewInt(4))
    den := big.NewInt(1)
    if e > 0 {
        num.Lsh(num, uint(e))
    } else {
        den.Lsh(den, uint(-e))
    }
    ten := big.NewInt(10)
    if p > 0 {
        num.Mul(num, new(big.Int).Exp(ten, big.NewInt(int64(p)), nil))
    } else {
        den.Mul(den, new(big.Int).Exp(ten, big.NewInt(int64(-p)), nil))
    }
    q, r := new(big.Int).QuoRem(num, den, new(big.Int))
    return unrounded(q.Uint64()) | bool2[unrounded](r.Sign() != 0)
}
```

乘 4 是為了留出未捨入數的兩個 bit，`QuoRem` 的商就是 `⌊4x·2^e·10^p⌋`，餘數非零就設 sticky bit。這個版本在我機器上跑一次要 246.9 ns。快速版本要跑到 1.542 ns，差 160 倍。

### 定點對數：先確定答案有幾位數

在呼叫 `uscale` 之前要先算出 `p`，而 `p` 的公式裡有 `⌊log₁₀ 2^e⌋`。這種東西當然不能真的去呼叫 `math.Log10`，Go 1.27 的做法是定點乘法：

```go
// log10Pow2(x) returns ⌊log₁₀ 2**x⌋ = ⌊x * log₁₀ 2⌋.
func log10Pow2(x int) int {
    // log₁₀ 2 ≈ 0.30102999566 ≈ 78913 / 2^18
    return (x * 78913) >> 18
}

// log2Pow10(x) returns ⌊log₂ 10**x⌋ = ⌊x * log₂ 10⌋.
func log2Pow10(x int) int {
    // log₂ 10 ≈ 3.32192809489 ≈ 108853 / 2^15
    return (x * 108853) >> 15
}
```

這種魔術常數看了會有點不安，所以我用 `math/big` 的有理數精確算出正確答案，窮舉比對：

```
log10Pow2 在 x ∈ [-1200, 1200] 錯誤數: 0
log2Pow10 在 x ∈ [-400, 400] 錯誤數: 0
log10Pow2 第一個失效的正 x = 1651
log2Pow10 第一個失效的正 x = 643
```

`float64` 的指數範圍是 `e ∈ [-1074, 971]`，`p` 的範圍是 `[-343, 341]`，兩個常數的有效範圍都遠超實際需求。這種「近似值但在定義域內完全精確」的東西，不實際跑一遍窮舉我是不會相信的。

這裡可以把開場那支影片收回來了。Quake III 的 `i = 0x5f3759df - (i >> 1)` 之所以成立，是因為 IEEE 754 的 bit pattern 當成整數讀的時候，那個整數是 `log₂ x` 的仿射近似。把 `x = (1+m)·2^e` 的 biased exponent 寫成 `E = e + 127`，整數表示就是 `2^23·(e + 127 + m)`，而 `log₂ x = e + log₂(1+m) ≈ e + m + σ`，兩者只差一個常數與縮放。

換句話說 `0x5f3759df` 跟 `log10Pow2` 是同一個物種：都在用整數運算取代超越函數。真正的差別是對誤差的態度，我把兩邊量在一起：

```
0x5f3759df 初始猜測   最大相對誤差 3.4376%
一次牛頓迭代後         最大相對誤差 0.1752%
log10Pow2 在 e ∈ [-1074, 971] 的誤差筆數 0
```

Quake III 接受 3.4% 的誤差，因為它只是要算光照角度，靠一次牛頓法收到 0.18% 就夠用了。`log10Pow2` 一個 bit 的誤差都不能有，因為 `p` 算錯一格輸出就會多一位數字。同樣是「用魔術常數近似一個對數」，一邊是圖形學的夠用就好，一邊是要在定義域內證明到完全精確。

這個張力後面還會出現兩次：`pm` 對 `10^p` 的 128 位近似，以及丟掉低 64 位的那個決定。兩次都是「引入誤差，然後證明它影響不到答案」。

同樣的手法也用在 skewed footprint（後面會講到）的常數上：

```go
func skewed(e int) int {
    return (e*631305 - 261663) >> 21
}
```

我一樣窮舉驗證過：`skewed 在 e ∈ [-1200, 1000] 錯誤數: 0`。

## 固定位數列印

第一個應用。給 `f = m · 2^e`，要輸出剛好 `n` 位十進位數字。

推導很直白。要求 `d = m · 2^e · 10^p ∈ [10^(n-1), 10^n)`，兩邊取 `log₁₀` 得到 `n-1 ≤ log₁₀(m·2^e) + p < n`，把 `p` 解出來：

$$
p = n - 1 - \left\lfloor \log_{10} (m \cdot 2^{e}) \right\rfloor
$$

<br/>

問題只剩怎麼算出那個 floor。`unpack64` 回傳的 `m` 保證最高位是 1，也就是 `bits(m)` 恆為 64，所以 `log₁₀ m` 可以直接用 `log₁₀ 2^63` 代替、併進 `log10Pow2(e+63)`。這個代替是「把 `m` 往下壓成同位數的最小值」，真實的 `⌊log₁₀(m·2^e)⌋` 因此可能比估計值再大一，換句話說算出來的 `p` 可能偏小一格，輸出會多出一位數字，所以程式碼最後檢查 `d ≥ 10^n`，有超過就用 `u.div(10)` 除掉一次。整段程式碼是：

```go
func FixedWidth(f float64, n int) (d uint64, p int) {
    m, e := unpack64(f)
    p = n - 1 - log10Pow2(e+63)
    var pre scaler
    prescale(&pre, e, p, log2Pow10(p))
    u := uscale(m, &pre)
    d = u.round()
    if d >= uint64pow10[n] {
        d, p = u.div(10).round(), p-1
    }
    return d, -p
}
```

十行。這就是完整的固定位數浮點數列印。

拿 π 實測，取 15 位：

```
π = 0x1921fb54442d18 * 2**-51  (bits(m)=64)
p = n-1-⌊log₁₀ 2**(e+63)⌋ = 14
uscale(m, e=-62, p=14) = ⟨314159265358979.0+⟩
FixedWidth(π, 15) = 314159265358979 * 10**-14
strconv.FormatFloat(π,'e',14,64) = 3.14159265358979e+00
```

`⟨314159265358979.0+⟩` 那個 `+` 說的是「π 的 float64 值比 3.14159265358979 大一點點」，剛好對應 π 的下一位是 3，所以捨去。

## 解析：同一件事反過來做

解析用的是同一個 `uscale`，只是已知與未知對調。列印是給 `m, e` 求 `p`，解析是給 `d, p` 求 `e`：

$$
e = 52 - \left\lfloor \log_{2} (d \cdot 10^{p}) \right\rfloor
  \approx 53 - \operatorname{bits}(d) - \left\lfloor \log_{2} 10^{p} \right\rfloor
$$

<br/>

```go
func Parse(d uint64, p int) float64 {
    b := bits.Len64(d)
    lp := log2Pow10(p)
    e := min(1074, 53-b-lp)
    var pre scaler
    prescale(&pre, e-(64-b), p, lp)
    if pre.s >= 64 {
        return 0
    }
    u := uscale(d<<(64-b), &pre)
    s := bool2[int](u >= unmin(1<<53))
    u = u>>s | u&1
    e = e - s
    return pack64(u.round(), -e)
}
```

先講定義。`min(1074, ...)` 是處理次正規數，`e` 超過 1074 代表結果落在 `2^-1074` 這個最小刻度以下，要少留幾位；`prescale(&pre, e-(64-b), ...)` 是因為下一行先把 `d` 左移成 64 位對齊（跟快速 `uscale` 的要求一致），左移幾位就在指數裡扣回來。

中間的 `uscale` 之後到結尾是「結果可能多出一位 mantissa 就右移一次」的 branch-free 寫法，需要一點鋪墊：

- 目標是把未捨入數壓回 53 位以內。捨入前若 `u.round()` 會達到 `2^53`，就必須先右移一位、指數加回去。
- 問題是不能真的先 `round()` 再比較，那就多捨入一次了。解法是把「捨入後會等於 `x`」翻譯成未捨入數上的直接比較：要讓 `round(u) == x`，`u` 最小可以是多少？整數部分 `(x-1)` 配 half bit 為 1 時，`round()` 走 half to even：`x` 偶數時會進位上來、奇數時不會。所以只要針對偶數的 `2^53` 取門檻 `⟨…⟩ = (2^53 << 2) - 2`，也就是 `unmin(1<<53)`。
- `u >= unmin(1<<53)` 於是等價於「這個值捨入後會變成 `2^53` 或更多」，右移的 `s` 就是 1；`u>>s | u&1` 右移時不忘把 sticky bit 帶下來。

我覺得這段最值得玩味的地方，是它跟 `FixedWidth` 的對稱性。同一個表、同一個原語、同一組定點對數，方向相反而已。Go 1.26 之前這是兩套完全不同的程式碼：`ftoadbox.go` 是 Dragonbox，`atofeisel.go` 是 Eisel-Lemire，各自帶各自的邏輯。

## 最短列印：2^89 的悖論

最短列印是三個裡面最麻煩的。目標是「用最少的位數，而且 parse 回來要一模一樣」。

直覺的做法是：從 1 位開始遞增呼叫 `FixedWidth`，直到 `Parse(FixedWidth(f, n)) == f`。這個做法是錯的，而且有一個乾淨的反例：`f = 2^89`。

我把它跑出來：

```
f            = 2**89 = 618970019642690137449562112
前一個 float64 = 618970019642690068730085376
後一個 float64 = 618970019642690274888515584
下界中點      = 618970019642690103089823744.0
上界中點      = 618970019642690206169038848.0
正確捨入到 16 位 = 6.189700196426901e+26
  parse 回來 == f ? false
Short(f)      = 6189700196426902 * 10**11 (共 16 位)
strconv 最短   = 6.189700196426902e+26
  parse 回來 == f ? true
```

看出來了嗎。`2^89` 是 2 的冪次，浮點數的指數在這裡跳了一階，所以「前一個 float64」的距離只有「後一個 float64」的一半。下界中點跟 `f` 差 `¼·2^e`，上界中點差 `½·2^e`。畫成數線：

```
        prev              下界中點        f               上界中點       next
─────────┼──────────────────┼────────────┼──────────────────┼────────────────▶
                          ◄── ¼ULP ──► │ ◄──── ½ULP ────► 
                                          footprint（skewed：¼ + ½ = ¾ ULP）
```

所有能正確 parse 回 `f` 的十進位數都必須落在兩個中點之間，這個區間就是 footprint。

`f` 的下一位是 3，所以正確捨入到 16 位是 `...901`。但是 `...901` 落在下界中點 `...103089823744` 之外，parse 回來會變成前一個 float64，round-trip 失敗。而 `...902` 落在區間內，parse 回來是對的，也是 16 位。

所以：存在一個 16 位的正確答案，但它不是 16 位的正確捨入結果。遞增呼叫 `FixedWidth` 的迴圈會在 16 位失敗，跳到 17 位，輸出比必要長一位。

Russ Cox 把 `f` 到左右兩個中點的距離叫做 footprint。正常情況是對稱的 `2^e`，在 2 的冪次上會變成 skewed 的 `¾·2^e`。這也是 `skewed()` 那個魔術常數存在的理由。

正確做法是直接算出兩個中點對應的十進位邊界，然後在 `[dmin, dmax]` 這個整數區間裡挑：

```go
func Short(f float64) (d uint64, p int) {
    const minExp = -1085
    m, e := unpack64(f)

    var mn, mx uint64
    z := 11
    if m == 1<<63 && e > minExp {
        p = -skewed(e + z)
        mn = m - 1<<(z-2) // min = m - 1/4 * 2**(e+z)
        mx = m + 1<<(z-1) // max = m + 1/2 * 2**(e+z)
    } else {
        if e < minExp {
            z = 11 + (minExp - e)
        }
        p = -log10Pow2(e + z)
        mn = m - 1<<(z-1)
        mx = m + 1<<(z-1)
    }
    odd := int(m>>z) & 1

    var pre scaler
    prescale(&pre, e, p, log2Pow10(p))
    dmin := uscale(mn, &pre).nudge(+odd).ceil()
    dmax := uscale(mx, &pre).nudge(-odd).floor()

    if d = dmax / 10; d*10 >= dmin {
        return trimZeros(d, -(p - 1))
    }
    if d = dmin; d < dmax {
        d = uscale(m, &pre).round()
    }
    return d, -p
}
```

程式碼裡有兩個魔術常數，先解釋清楚再看判斷邏輯。

`z = 11`：`unpack64` 把 mantissa 左移成 64 位、最高位恆為 1，但 `float64` 真正有的只有 53 位有效數字，多出來的就是低 11 位。所以在這個表示法下，相鄰兩個 `float64` 的距離是 `2^(e+11)`，也就是一個 ULP。正常情況的 footprint 是「到左右鄰居各半」，也就是半個 ULP：`mn = m - 1<<(z-1)`、`mx = m + 1<<(z-1)` 裡的 `z-1 = 10` 就是這樣來的。

`minExp = -1085`：最極端的次正規數沒有更低的位可以擴張了，footprint 的下界會撞到零以下，所以把 `z` 加大、犧牲一點精度讓區間不越界。

`p` 的選法讓 footprint 剛好落在 `[1, 10)` 個十進位整數之間，所以區間裡最多只有 10 個候選。這個觀察來自 Schubfach，價值在於：10 個連續整數裡最多只有一個以 0 結尾。

於是判斷只有三種情況：

一、區間裡有一個以 0 結尾的候選，用它，砍掉尾零之後位數最少（`dmax/10` 那行）。

二、區間裡只有一個候選，就用它。這正是 `2^89` 的情況，那唯一的候選不是正確捨入的結果，但它是正確答案。

三、有多個候選但都不以 0 結尾，那就用正確捨入的那個（第三次 `uscale`）。

`nudge(±odd)` 處理的是捨入邊界的開閉區間問題。中點本身是 half 值，而 round half to even 的規則是「half 進位到偶數」：mantissa 為偶數時，中點會被捨回 `f` 這一邊，所以邊界是閉的、可以包含；mantissa 為奇數時，中點會被推到另一側，邊界是開的、必須排除。用「往區間內加減 1 再取 ceil/floor」來實現開區間，可以完全不用分支。

## 快速 uscale：128 位查表，然後把低 64 位丟掉

前面所有東西都建立在 `uscale` 上，但到目前為止 `uscale` 還是大整數版本。現在來看它怎麼變快。

核心想法是：把 `10^p` 近似成一個 128 位的浮點數 `pm · 2^pe`，其中

$$
\begin{aligned}
\mathit{pe} &= \left\lfloor \log_{2} 10^{p} \right\rfloor - 127 \\
\mathit{pm} &= \left\lceil 10^{p} / 2^{\mathit{pe}} \right\rceil
\end{aligned}
$$

<br/>

`pm` 那一行是天花板函數而不是四捨五入，這個細節後面「省下不必要的乘法」那節會變成關鍵。

這樣 `pm ∈ [2^127, 2^128)`，存成 `hi`、`lo` 兩個 `uint64` 查表。Go 1.27 的表在 `pow10tab.go`，範圍 `p ∈ [-348, 347]`，共 696 筆，每筆 16 byte，總共 10.9 KiB。

於是 `x · 2^e · 10^p ≈ (x · pm) >> -(e + pe)`。`x` 是 64 位，`pm` 是 128 位，乘出來是 192 位，用兩次 `bits.Mul64` 完成。

關鍵問題來了：`pm` 因為取了 ceiling，比真值大了一個誤差 `ε₀ < 1`，所以 `x · pm` 比真值大了 `ε₁ = x · ε₀ < x`。這個誤差最多影響乘積的低 `bits(x)` 位。

而 Russ Cox 要求呼叫端把 `x` 左移到最高位為 1，也就是 `bits(x) = 64`。所以誤差最多影響 192 位乘積的最低 64 位。

那就別算那 64 位了。

這句話就是整個最佳化的全部內容。中間 64 位跟高 64 位是可信的，低 64 位不可信但我們也不需要，因為我們要的只是「高位的整數部分」加上「後面還有沒有東西」這兩個資訊，而後者可以從中間 64 位讀出來。

位移量的推導可以拆成三步。先把整個乘積右移到「小數點左邊只剩我們要的東西」：

一、`x · pm · 2^(e+pe)` 才是真正的值，但 192 位乘積存的是 `x·pm`，所以要先抵掉 `e+pe` 個位。

二、192 位裡只有最高的 64 位是我們保留的，砍掉低的 128 位：再右移 128。

三、這 64 位裡還要留兩個 bit 給未捨入形式：再右移 2。

合計：

$$
s = -(e + \mathit{pe}) - 128 - 2
  = -\left(e + \left\lfloor \log_{2} 10^{p} \right\rfloor + 3\right)
$$

```go
func prescale(pre *scaler, e, p, lp int) {
    pre.pmHi = pow10Tab[p-pow10Min].hi
    pre.pmLo = pow10Tab[p-pow10Min].lo
    pre.s = -(e + lp + 3)
}
```

### 正確性怎麼證

論文把證明拆成三種情況，另外開了一篇 [Proof by Ivy](https://research.swtch.com/fp-proof) 用 APL 風格的 Ivy 語言寫成可執行的證明。大意是：

`p ∈ [0, 27)`：`5^p` 完全塞得進 `pm` 的高 64 位，`lo` 全是 0，丟掉的低位本來就是 0，所以精確。

`p ∈ [-27, -1]`：`x · pm` 在做的其實是除以 `5^(-p)`。當 `x` 是 `5^(-p)` 的倍數時只有最低 `bits(x)` 位非零，丟掉不影響。當 `x` 不是倍數時，分三步看：(a) 除法餘數的結構以 `5^(-p)` 為週期，小數部分的最小刻度離整數至少有 `1/5^(-p)` 那麼遠，這個距離遠大於 `pm` 的近似誤差，所以不會發生「誤差把值推過整數邊界」；(b) 這個小數部分是循環小數，循環表示裡必然有非零位，否則它就是精確整數了，跟「不是倍數」矛盾；(c) 64 位的寬度足以涵蓋多個週期，所以被丟掉的低 64 位（連同中間位）一定可以找到一個 1 把 sticky bit 點亮。

大的 `p`：這是最有意思的一段。誤差加進去可能產生進位，進位會從右往左傳播：一路把 1 變成 0，直到遇到第一個 0 把它變成 1 然後停下來。所以判斷規則是：只要中間 64 位裡能看到任何一個原本就是 1 的位，就代表進位鏈在那裡斷掉了、沒有繼續往高位走，高 64 位完全不受誤差影響；反過來如果中間 64 位全被進位洗成了 0，才需要擔心。Russ Cox 用程式分析表裡每一筆 `pm`，證明所有可能的 `x · pm` 中間位都至少有一個 1，所以永遠可以在中間位看到進位鏈停下的證據。

我沒有能力驗證他的證明，但我可以做交叉比對。用 200 萬個隨機 `float64`，把快速版跟大整數版的 `uscale` 輸出逐一比對：

```
比對 1999062 組隨機輸入，不一致 0 組
```

## 省下不必要的乘法

到這裡 `uscale` 已經是兩次 `bits.Mul64`。Russ Cox 又砍掉一次。

想法是：如果第一次乘法 `x · pm.hi` 的結果，被移掉的那 `s` 個低位裡有任何一個 1，那就代表後續的修正（借位）不可能傳到我們要的高位，也代表 sticky bit 一定是 1。既然如此，第二次乘法根本不用算。

為了讓這個推論成立，`pm` 必須是「往上取整」的，所以修正是減法（借位）而不是加法（進位）。Go 1.27 把表的表示法從 `hi<<64 + lo` 改成 `hi<<64 - lo`：

```go
// A pmHiLo represents hi<<64 - lo.
type pmHiLo struct {
    hi uint64
    lo uint64
}
```

這件事在原始碼裡看得到。同一筆 `1e-348`，Go 1.26 的表是：

```go
{0xfa8fd5a0081c0288, 0x1732c869cd60e453}, // hi<<64 + lo, rounded down
```

Go 1.27 的表是：

```go
{0xfa8fd5a0081c0289, 0xe8cd3796329f1bac}, // hi<<64 - lo, rounded up
```

我用 Python 的 `Fraction` 精確算過，`10^-348 · 2^1284` 的 floor 是 `0xfa8fd5a0081c02881732c869cd60e453`，ceil 是 `...e454`，兩張表差剛好 1 ULP。Go 1.26 存 floor，Go 1.27 存 ceil。

最佳化後的 `uscale` 長這樣：

```go
func uscale(x uint64, c *scaler) unrounded {
    hi, mid := bits.Mul64(x, c.pmHi)
    s := c.s & 63 // make shifts cheaper
    if hi>>s<<s != hi {
        // The shift dropped a 1 bit. No fix is needed, and sticky is 1.
        return unrounded(hi>>s | 1)
    }
    mid2, _ := bits.Mul64(x, c.pmLo)
    hi -= bool2[uint64](mid < mid2)
    return unrounded(hi>>s | bool2[uint64](mid-mid2 > 1))
}
```

`hi>>s<<s != hi` 就是「被移掉的 `s` 個低位不全為 0」的寫法。

slow path 的三行值得逐行看。`pm = hi<<64 - lo`，所以乘積要取 `(x·pmHi)<<64 - x·pmLo`，第二次 `Mul64` 算出的是被減掉的低半部，`mid < mid2` 表示不夠減、要從高位借位。最後那個 `mid-mid2 > 1` 是全段最微妙的地方：我們想知道「扣掉 half bit 之後還有沒有殘留」（sticky）。中間位存著 `(x·pm)>>64` 的低半部，如果它跟 `x·pmLo` 恰好差 1，代表多出來的值剛好是 1 個單位，這一個單位正是把 `pm` 從精確值推到 ceiling 的那一格誤差，換算回真實的 `x·10^p·2^e` 是完全精確的整數，所以 sticky 是 0；差超過 1 才代表 half 下方真的還有東西。抄程式碼時這裡最容易手癢寫成 `!= 0`。

另外 `s := c.s & 63` 那行：Go 對「位移數 ≥ 位寬」有明確定義（結果為 0），但硬體的變動位移指令只看低 6 位，編譯器想維持 spec 語意就得多生成一段檢查。遮罩成低 6 位讓它化簡成一條指令，代價是 `s ≥ 64` 時結果不對，這由呼叫端擋掉（`Parse` 裡的 `pre.s >= 64` 直接回傳 0；`Short`、`FixedWidth` 的參數範圍內 `s` 不會超過 63）。

那這條 fast path 到底多常走到？論文只說「大部分時候」，沒有數字。我加了計數器實測 10 萬個隨機 `float64`：

```
Short:      共 245887 次 uscale，單次乘法就結束 242365 次 (98.57%)
            平均每個 float64 呼叫 2.46 次 uscale
FixedWidth(17): 共 100000 次 uscale，單次乘法就結束 96979 次 (96.98%)
FixedWidth(6):  共 100000 次 uscale，單次乘法就結束 100000 次 (100.00%)
```

`FixedWidth(6)` 的 100% 是有道理的：位數要求越少，`s` 越大，被移掉的低位越多，裡面有 1 的機率越高。6 位輸出的情況下 10 萬次全部命中。

`Short` 的 98.57% 也符合論文的說明。`Short` 會用同一個 `pm` 乘兩個只差一個 bit 的 `x`，就算其中一個湊巧把低 `s` 位全部清成 0，另一個也很難同時清乾淨。

換算成實際的寬乘法次數：每個 `float64` 的最短列印平均做 2.49 次 64×64 寬乘法。

## Go 1.27 到底改了什麼

講完演算法，來看 Go 的原始碼實際上動了哪些地方。

Go 1.26 的 `internal/strconv` 裡，浮點數演算法分散在四個檔案：

| 檔案 | 行數 | 演算法 |
|---|---|---|
| `ftoadbox.go` | 349 | Dragonbox（最短列印） |
| `ftoafixed.go` | 184 | Ryū 風格的固定位數列印 |
| `atofeisel.go` | 166 | Eisel-Lemire（解析） |
| `math.go` | 179 | 共用工具：`umul128`／`umul192`、`pow10` 查表、定點對數、`divisiblePow5`、`trimZeros` |

Go 1.27 把這四個檔案全部刪掉，換成一個 `uscale.go`，290 行。878 行變 290 行，而且原本三套各自為政的演算法變成三個共用同一個原語的函式。

`pow10tab.go` 兩個版本都是 715 行、696 筆，但意義變了。Go 1.26 的表是 `uint128` 型別、往下取整，服務 Dragonbox 跟 Eisel-Lemire；Go 1.27 的表是 `pmHiLo` 型別、往上取整，列印與解析共用。HN 上有人（`e4m2`）點出這件事才是 uscale 真正的賣點：

> uscale's main strength isn't its speed, but rather its simplicity and, more importantly, the fact that it does both formatting and parsing using a single ~11 KiB table, which no other state-of-the-art algorithm offers

### 效能實測

我在同一台機器（AMD Ryzen 9 5900X，Linux）上，用同一份 benchmark 程式碼分別跑 Go 1.26.0 與 Go 1.27.0，`-count 5` 取中位數。輸入是 1 萬個隨機 bit pattern 的 `float64`：

| Benchmark | Go 1.26 | Go 1.27 | 變化 |
|---|---|---|---|
| `AppendFloat` 最短 | 51.27 ns/op | 42.88 ns/op | -16.4% |
| `AppendFloat` 17 位 | 45.92 ns/op | 38.41 ns/op | -16.4% |
| `AppendFloat` 6 位 | 37.16 ns/op | 32.98 ns/op | -11.2% |
| `ParseFloat` | 69.20 ns/op | 60.01 ns/op | -13.3% |
| `fmt.Sprintf("%v", f)` | 137.2 ns/op | 107.9 ns/op | -21.4% |

需要說清楚的是，這個比較不是純演算法比較。Go 1.26 到 1.27 之間編譯器跟 runtime 也有改動，而且 benchmark 裡包含了 digit formatting 跟字串處理的成本，把演算法本身的差距稀釋掉了。單看 `uscale` 這個原語，我量到的是 1.542 ns/op，大整數版本是 246.9 ns/op。

我原本也量了 `json.Marshal` 跟 `json.Unmarshal`，但 Go 1.27 的 `encoding/json` 底層整個換成 v2 實作，數字被那個改動蓋過去，跟 `strconv` 無關，所以不列。

### 輸出有沒有變

換演算法最怕的是輸出跑掉。我寫了一個差異測試：300 萬個隨機 `float64`，每個跑 7 種格式（`e` 最短／16 位／5 位，`f` 最短／6 位，`g` 最短／10 位），加上 `float32` 的最短輸出，全部餵進 SHA-256：

```
--- go1.26 ---
go1.26.0  樣本 2998543  sha256=52d16cd63eee86d2cc044a24665f087c
--- go1.27 ---
go1.27.0  樣本 2998543  sha256=52d16cd63eee86d2cc044a24665f087c
```

完全一致，而且所有最短輸出的 round-trip 都成功。我另外用自己實作的 `Short` 跟 `Parse` 跑 100 萬個隨機 `float64`：

```
測試 999533 個隨機 float64：round-trip 失敗 0，與標準庫不同 0
```

換句話說，這是一次純粹的效能與程式碼複雜度改善，行為零變化。這大概也是它沒被寫進 release notes 的原因。

## 這是終點嗎

Russ Cox 在文末寫下「At long last, the dragons have been vanquished」，從 Dragon4 開始的龍終於被屠了。

不過 HN 上的討論給了比較節制的看法。C++ `{fmt}` 函式庫的作者 Victor Zverovich 留言指出：

> Shortest uscale is basically Schubfach or, rather, it's variant called Teju Jagua and has 2-3 wide multiplications compared to 1 for newer methods.

2026 年已經有 Zmij、xjb、yy 這些更新的演算法，在 `dtoa-benchmark` 上跑得比 uscale 快。我實測的 2.49 次寬乘法也印證了這個說法，新方法可以做到 1 次。

但是這些新方法快在哪，`e4m2` 講得很清楚：主要快在 digit formatting，也就是把十進位整數變成字串那一段，不是浮點數轉換本身。而 uscale 的價值在於它用同一個 11 KiB 的表、同一個原語，同時解決固定位數列印、最短列印、解析三個問題，程式碼還比任何一個競爭者短。

對標準函式庫來說，這個取捨很划算。`strconv` 是要維護十年以上的東西，878 行三套演算法變成 290 行一套，這個帳我覺得算得過來。

另外值得一提的是，論文發表之後已經有人做了 [Rust 的獨立實作](https://github.com/biantaishabi2/fp-convert)，MIT / Apache-2.0 授權，把 unrounded number、預算表、`format_shortest`、`fixed_width` 都做齊了。演算法本身跟語言無關，只需要 64×64→128 的寬乘法，任何有這個原語的語言都能照抄。

## 附錄：不用 float64 的話，代價是什麼

這篇寫到這裡，有一個很合理的問題還沒回答：既然 `float64` 印出來這麼麻煩，那不要用它不就好了？Go 標準庫有 `math/big`，社群有 `github.com/shopspring/decimal`，這些東西跟 `float64` 到底差在哪。

這個問題值得一節，因為答案剛好從另一個角度說明 uscale 在解什麼。

### 先看一眼 0.1

一切的起點是這個數：

```
float64 的 0.1 = 0.1000000000000000055511151231257827021181583404541015625
              = 3602879701896397 / 36028797018963968
```

分母是 `2^55`。`float64` 只能表達分母是 2 的冪次的分數，而 `1/10` 的分母有因數 5，所以永遠塞不進去。這不是實作的問題，是二進位這個底數的問題。

離開這個限制有三條路：位數加大但底數不變、改用有理數、換一個底數。Go 生態裡剛好各有一個代表，而它們全都長在 `big.Int` 上面，所以先從它開始看。

### big.Int：任意精度的整數

`math/big` 最底層的東西，結構單純到有點無趣：

```go
type Int struct {
    neg bool // 符號
    abs nat  // 絕對值
}
```

`nat` 就是 `[]Word`，`Word` 是 `uint`，在 amd64 上是 64 bit。所以一個 `big.Int` 是「一串 64 bit 的數字，小端序，加一個符號」。加法逐 word 帶進位，乘法在短的時候用課本乘法，較短的那個運算元超過 `karatsubaThreshold`（40 個 word，也就是 2560 bit）才換 [Karatsuba](https://www.youtube.com/watch?v=H8MFw0qxY84)。門檻存在的理由是 Karatsuba 雖然是 `O(n^1.585)`，遞迴跟切割的常數項卻大得多，短的時候反而輸給 `O(n²)`。Go 對這個數字沒有硬猜，`math/big` 裡有一支 `TestCalibrate`，用 `go test -run=Calibrate -calibrate` 跑出來的資料重新校準門檻。

`big.Int` 本身沒有小數的概念，但它是下面兩個東西的地基，本文的 `uscaleBig` 也是靠它算出參考答案的。

`unsafe.Sizeof(big.Int{})` 是 32 byte，再加上堆上的 `[]Word`。一次 `Add` 7.5 ns、`Mul` 9.5 ns，跟 `float64` 的 0.47 ns 比是 16 到 20 倍。考慮到它處理的是可變長度的東西，其實不算慢。

### big.Float：任意精度，但底數還是 2

這是最容易被誤會的一個。

```go
type Float struct {
    prec uint32
    mode RoundingMode
    acc  Accuracy
    form form
    neg  bool
    mant nat
    exp  int32
}
```

看到 `mant` 跟 `exp` 就知道它是浮點數，重點在 `prec` 這個欄位的單位是二進位位數，不是十進位位數。換句話說 `big.Float` 是一個可以自訂尾數長度的 IEEE 754，底數仍然是 2。

它沒有解決 0.1 的問題。我實測過：

```
prec=53    十進位有效位數 ≈ 16.0，與 1/10 的差 ≈ 5.551e-18
prec=100   十進位有效位數 ≈ 30.1，與 1/10 的差 ≈ 1.972e-32
prec=200   十進位有效位數 ≈ 60.2，與 1/10 的差 ≈ 1.556e-62
```

誤差變小了，但沒有變成零，而且永遠不會變成零。把 0.1 累加十次：

```
float64        : 0.1 累加十次 == 1.0 ? false（差 -1.11022e-16）
big.Float(200) : 0.1 累加十次 == 1.0 ? false（差 1.24460e-60）
```

`prec = 53` 的 `big.Float` 我驗證過跟 `float64` 是同一個值，一個 bit 都不多（差別只在指數欄位是 `int32`，範圍大得多，而且沒有 NaN，會產生 NaN 的運算直接 panic）。所以如果有人說「改用 `big.Float` 就不會有浮點數誤差了」，那是誤解：買到的是更小的誤差，不是沒有誤差。

還有一個更陰險的地方。我把 `0.1 + 0.2 == 0.3` 在各種精度下都跑了一次：

```
prec=53   false
prec=60   true
prec=64   true
prec=100  true
prec=200  true
prec=500  true
```

`prec` 一調大就變成 `true` 了。這種「隨機通過」正是這類 bug 最難查的原因，測試會過，然後在某個湊巧的輸入上炸掉。`big.Float` 該用的場合是需要高精度數值計算的地方，例如算 π 的一百萬位，不是記帳。

### big.Rat：唯一真的精確的

```go
type Rat struct {
    a, b Int
}
```

分子分母兩個 `big.Int`，每次運算之後約分。`1/10` 就是 `1/10`，`(1/3)·3` 就是 `1`，0.1 累加十次就是 `1`，全部精確。

代價是分母會爆炸。加法要通分，兩個互質的分母通分之後就是它們的乘積，而每次運算都在做 GCD。我拿調和級數試：

```
調和級數前  10 項：分子 13 bits，分母 12 bits
調和級數前  50 項：分子 74 bits，分母 72 bits
調和級數前 100 項：分子 134 bits，分母 132 bits
調和級數前 200 項：分子 296 bits，分母 293 bits
```

200 項就要 296 bit，而這還是分母高度重疊的好情況，真實資料通常更糟。另外它沒有「小數點後幾位」這個概念，要印出來一定得指定位數並且捨入。

`big.Rat` 適合當參考實作，不適合當生產環境的數值型別。本文前面驗證 `log10Pow2` 那幾個魔術常數的時候，用的就是它。

### shopspring/decimal：換一個底數

現在來看 `github.com/shopspring/decimal`（v1.4.0）：

```go
type Decimal struct {
    value *big.Int
    exp   int32
}
```

值是 `value · 10^exp`。跟 `big.Float` 的差別只有一個字：底數從 2 換成 10。

就這一個字，0.1 變成精確的：

```
shopspring 的 0.1：value=1 exp=-1（= 1 · 10^-1）
```

金額、稅率、匯率這些東西本來就是十進位定義的，用十進位底數存就不會有轉換損失。這是它在金融領域被廣泛使用的全部理由。

有三件事值得知道。

一、加減乘是精確的，除法不是。`Add` 跟 `Sub` 先把兩邊 rescale 到同一個 `exp` 再做 `big.Int` 加減，`Mul` 直接把 `value` 相乘、`exp` 相加，全部無損。但 `Div` 沒辦法無損，`1/3` 在十進位裡也是無限循環，所以它有一個全域變數：

```go
var DivisionPrecision = 16
```

實測：

```
1/3     = 0.3333333333333333
(1/3)*3 = 0.9999999999999999
```

`big.Rat` 在這裡會給你精確的 `1`。所以 `decimal` 不是「精確運算」，是「十進位精確加上除法可控捨入」。對記帳來說這正是想要的，因為會計本來就要求除不盡的時候按規則捨入並且記下來。

二、`Add` 比 `Mul` 慢四倍，這很反直覺，原因在 rescale：

```go
expScale := new(big.Int).Exp(tenInt, big.NewInt(int64(diff)), nil)
```

兩個 `exp` 不同的數相加，要先把其中一個乘上 `10^diff`，而 `10^diff` 是當場用 `big.Int.Exp` 算出來的。實測 `Add` 202 ns、8 allocs，`Mul` 49 ns、2 allocs。如果一段程式碼裡都是同一種幣別、同一個小數位數，讓所有 `Decimal` 的 `exp` 保持一致就能省掉這段。

三、`NewFromFloat` 有陷阱：

```
NewFromFloat(0.1)                  = 0.1
NewFromFloatWithExponent(0.1, -30) = 0.100000000000000005551115123126
```

`NewFromFloat` 走的是「最短列印」，也就是本文 `Short()` 那個演算法，所以它會把 `float64` 已經有的誤差抹掉，給你一個看起來很乾淨的 `0.1`。這在大部分時候是你要的，但它代表 `float64 → Decimal` 這一步是有語意的，不是單純的型別轉換。真的想看 `float64` 的原貌得用 `NewFromFloatWithExponent`。

最保險的做法是根本不要讓值經過 `float64`：從資料庫、JSON、API 讀進來的時候直接 `NewFromString`。

### 放在一起比

| | 底數 | 0.1 精確 | 除法精確 | 一次加法 | 一次乘法 | struct 大小 |
|---|---|---|---|---|---|---|
| `float64` | 2 | 否 | 否 | 0.47 ns | 0.46 ns | 8 B |
| `big.Int` | 無小數 | 不適用 | 不適用 | 7.5 ns | 9.5 ns | 32 B + heap |
| `big.Float`（prec=200） | 2 | 否 | 否 | 35.2 ns | 43.6 ns | 40 B + heap |
| `big.Rat` | 有理數 | 是 | 是 | 203 ns | 103 ns | 64 B + heap |
| `decimal` | 10 | 是 | 否（16 位） | 202 ns | 49 ns | 16 B + heap |

同一台機器（AMD Ryzen 9 5900X）、Go 1.27.0、`-count 5` 取中位數。

### 回到本文的主題：它們各自怎麼列印

這一段是我覺得最有意思的部分。同樣是「把數字變成字串」，這四個東西走的是完全不同的路。

`strconv.FormatFloat` 走的是本文整篇在講的 uscale，290 行、一到兩次寬乘法。

`big.Float.Text` 走的是 `math/big/decimal.go`，那個檔案的開頭寫著：

> This file implements multi-precision decimal numbers. The implementation is for float to decimal conversion only; not general purpose use. […] The key observation and some code (shr) is borrowed from strconv/decimal.go

也就是說 `big.Float` 的列印用的還是十進位 buffer 上的長除法，本質上就是 Russ Cox 2011 年那篇文章的做法。Go 1.27 只換掉了 `strconv`，`math/big` 沒動。實測：

```
strconv.FormatFloat(x, 'f', -1, 64)    51.89 ns/op    16 B/op    1 allocs/op
big.Float.Text('f', -1)              3991    ns/op  2144 B/op   22 allocs/op
```

77 倍。`big.Float` 的尾數長四倍，不能直接比，但那 22 次配置說明它做的是完全不同層級的工作。

`decimal.Decimal.String()` 最有趣，因為它根本沒有浮點數轉換這個步驟：

```go
func (d Decimal) string(trimTrailingZeros bool) string {
    if d.exp >= 0 {
        return d.rescale(0).value.String()
    }
    abs := new(big.Int).Abs(d.value)
    str := abs.String()
    // ... 找位置插一個小數點
}
```

`big.Int.String()` 之後找位置插一個小數點，結束。值本來就是十進位存的，列印只是換個排版。實測 145 ns，比 `strconv` 慢將近三倍，但那三倍花在 `big.Int` 的十進位轉換跟四次記憶體配置上，不是花在演算法上。

而 `decimal.NewFromFloat` 是反過來的，它需要「把 `float64` 印成最短的十進位」，所以 repo 裡有一個 415 行的 `decimal-go.go`，開頭是：

```
// Copyright 2009 The Go Authors. All rights reserved.
// Multiprecision decimal numbers.
// For floating-point formatting only; not general purpose.
```

那是 2009 年 `strconv/decimal.go` 的複製。Go 1.27 把 `strconv` 換成 uscale 了，但這份複製不會跟著更新。實測 `NewFromFloat` 362 ns/op，其中大部分就花在這裡。

把四條路排在一起，uscale 的定位就清楚了。它是給「決定繼續用 `float64`」的人的最佳化。真的需要精確十進位的人早就離開 `float64` 了，代價是每次運算 100 到 400 倍的成本；而留下來的人付的代價，是每次列印要跑一次本文講了大半篇的演算法。Russ Cox 做的事情，是把這個代價從 51 ns 壓到 43 ns，順便把程式碼從 878 行壓到 290 行。

### 實務上怎麼選

我自己的分法很粗暴。

金額、稅、匯率、任何會出現在對帳單上的數字，用 `decimal`，而且值不要經過 `float64`。

科學計算、統計、機器學習、圖形，用 `float64`。這些領域的輸入本來就有量測誤差，15 位有效數字遠遠夠用，而且只有它有硬體支援。

需要超過 15 位有效數字的數值計算（大數運算、密碼學、高精度常數），用 `big.Float` 或 `big.Int`，並且記得 `prec` 的單位是 bit。

寫測試要一個絕對正確的參考答案，用 `big.Rat`。本文驗證那幾個魔術常數的時候就是這樣做的。

## 寫在最後

寫這篇的過程裡，讓我停下來最久的不是數學，是那張時間表。

Slishman 在 1990 年就用了 carry bit 檢查，但只做解析。Hack 在 2004 年就證明了 128 位夠用，但只證明解析。Giulietti 在 2018 年就重新發明了未捨入形式，但沒用 carry bit。Eisel 跟 Lemire 在 2020 年獨立地重新走了一次 Slishman 的路，還留著一個永遠不會被執行到的 fallback 分支，因為當時無法確定 128 位到底夠不夠。

零件全部都在。從 2020 年到 2026 年，缺的只是有人把它們放在同一張桌子上，然後發現這些東西其實是同一件事。

我自己在工作上也常遇到這種事情的小型版本：兩個模組各自解決一半的問題，各自演化了兩年，直到有人重寫的時候才發現它們可以共用同一個抽象，而且共用之後程式碼還變少了。差別只是規模，Russ Cox 那張桌子上放的是 79 年份的論文。

Go 1.27 升上去之後，你的 `fmt.Sprintf("%v", 3.14)` 就已經在跑這套演算法了，不需要改任何一行程式碼。這種免費的午餐不常有，值得知道它是怎麼來的。


本文的範例程式碼分成兩份。主文的演算法拆解放在 [uscale.go](uscale.go) 與 [main.go](main.go)，只用標準庫；附錄的 `math/big` 與 `shopspring/decimal` 比較放在 [appendix_demo.go.txt](appendix_demo.go.txt) 與 [appendix_bench_test.go.txt](appendix_bench_test.go.txt)，需要另外 `go get`。執行方式見 [README.txt](README.txt)。需要 Go 1.27 以上，因為 `pow10tab.go` 是直接從 `GOROOT` 抄過來的，1.26 的表是往下取整的版本，混用會算錯。

## 參考資料

- [Russ Cox: Floating-Point Printing and Parsing Can Be Simple And Fast](https://research.swtch.com/fp)
- [Russ Cox: Fast Unrounded Scaling: Proof by Ivy](https://research.swtch.com/fp-proof)
- [Russ Cox: Pulling a New Proof from Knuth's Fixed-Point Printer](https://research.swtch.com/fp-knuth)
- [Russ Cox: Floating Point to Decimal Conversion is Easy (2011)](https://research.swtch.com/ftoa)
- [Floating Point Formatting 系列索引](https://research.swtch.com/fp-all)
- [Go 1.27 原始碼 internal/strconv/uscale.go](https://github.com/golang/go/blob/go1.27.0/src/internal/strconv/uscale.go)
- [IEEE 754-2019 標準](https://standards.ieee.org/ieee/754/6210/)（§5.12.2 定義轉換的正確捨入與 `Pmin` 要求）
- [IEEE 754-1985 初版全文](https://www.ime.unicamp.br/~biloti/download/ieee_754-1985.pdf)（§5.6 的 0.47 ULP 誤差容忍與 Coonen 註腳）
- [Steele & White: How to Print Floating-Point Numbers Accurately (PLDI 1990)](https://dl.acm.org/doi/10.1145/93542.93559)
- [Wikipedia: Unit in the last place](https://en.wikipedia.org/wiki/Unit_in_the_last_place)
- [Hacker News: Go 1.27 討論串](https://news.ycombinator.com/item?id=49365405)
- [biantaishabi2/fp-convert：unrounded scaling 的 Rust 實作](https://github.com/biantaishabi2/fp-convert)
- [Victor Zverovich: yy-dtoa](https://vitaut.net/posts/2026/yy-dtoa/)
- [fmtlib/dtoa-benchmark 結果](https://fmtlib.github.io/dtoa-benchmark/results/)
- [tonybai: Russ Cox 与浮点数转换的十五年之战](https://tonybai.com/2026/02/03/russ-cox-15-year-war-on-floating-point-conversion/)
- [Goldstine & von Neumann: Planning and Coding of Problems for an Electronic Computing Instrument (1947)](https://www.historyofinformation.com/detail.php?id=658)
- [Konrad Zuse Internet Archive：Z3 的重建與規格](https://zuse.zib.de/reconstructionZ3)
- [Wikipedia: IBM 704](https://en.wikipedia.org/wiki/IBM_704)（第一台量產的浮點數硬體）
- [Wikipedia: Hexadecimal floating point](https://en.wikipedia.org/wiki/Hexadecimal_floating_point)（System/360 的 base-16 與 wobbling precision）
- [Goldberg: What Every Computer Scientist Should Know About Floating-Point Arithmetic](https://docs.oracle.com/cd/E19957-01/806-3568/ncg_goldberg.html)（guard digit 與 1968 年的硬體回收）
- [Go 原始碼 math/big/decimal.go](https://github.com/golang/go/blob/go1.27.0/src/math/big/decimal.go)（`big.Float` 列印用的十進位長除法）
- [shopspring/decimal](https://github.com/shopspring/decimal)（v1.4.0，`Decimal` 的 `value · 10^exp` 表示法）
