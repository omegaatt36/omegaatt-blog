---
title: Golang Iterator 簡介與 samber/lo 比較
date: 2025-05-31
categories:
  - develop
tags:
  - golang
  - optimization
  - iterator
cover:
  image: "images/cover.png"
---

自從 Golang 1.18 版本引入泛型（Generics）後，Go 語言的生態系統迎來了許多令人興奮的變化。其中，Golang 1.23 版本對 Iterator（迭代器）的標準化，以及 `iter` 套件的加入，無疑是近期改動中相當重要的一環。本文將淺談 Golang Iterator 的基本概念，深入探討 Pure Iterator 與 Impure Iterator 之間的區別與設計考量，並與社群中流行的 `samber/lo` 工具庫進行比較。

## 什麼是 Iterator？

Iterator Pattern（迭代器模式）是一種常見的設計模式，它提供了一種循序存取集合物件中各個元素的方法，而又無需暴露該物件的內部表示。簡單來說，Iterator 就像一個指針，可以依序指向集合中的下一個元素，直到遍歷完所有元素為止。

## Golang 中的 Iterator

在 Golang 1.23 之前，我們通常透過 `for-range` 迴圈來迭代 array、slice、string、map、channel 等內建資料結構。然而，對於自訂的資料結構或複雜的序列生成邏輯，缺乏一個統一的迭代標準。

Golang 1.23 版本正式將 Iterator 標準化，並在標準庫中加入了 `iter` 套件。同時，`slices` 和 `maps` 套件也增加了一些回傳 Iterator 的工廠函數（Iterator Factories）。到了 Golang 1.24，更有如 `strings.SplitSeq` 等函數加入，進一步豐富了 Iterator 的應用場景。

```go
// strings.SplitSeq 回傳一個迭代器，用於遍歷由 sep 分隔的 s 子字串。
// 此迭代器產生的字串與 Split(s, sep) 回傳的相同，但不會建構整個 slice。
// 它回傳一個單次使用的迭代器。
func SplitSeq(s, sep string) iter.Seq[string]
```

如果對 Golang 1.23+ 中 Iterator 的語法和語義還不熟悉，建議可以閱讀 Ian Lance Taylor 在 Go 官方部落格發表的[介紹文章](https://go.dev/blog/range-functions)

Iterator 的核心是一個函數，它接受一個 `yield` 函數作為參數。`yield` 函數用於產生序列中的下一個元素。當 `yield` 回傳 `false` 時，表示迭代終止。

以下是一個簡單的例子，`fibonacci` 函數回傳一個產生費氏數列的 Iterator：

```go
package main

import (
  "fmt"
  "iter"
)

// fibonacci 回傳一個費氏數列的 Iterator
func fibonacci() iter.Seq[int] {
  return func(yield func(int) bool) {
    a, b := 0, 1
    for yield(a) { // 當 yield(a) 回傳 true 時繼續迭代
      a, b = b, a+b
    }
  }
}

func main() {
  for n := range fibonacci() {
    if n > 100 {
      break
    }
    fmt.Printf("%d ", n)
  }
  fmt.Println()
  // 輸出: 0 1 1 2 3 5 8 13 21 34 55 89
}
```

在這個例子中，`fibonacci()` 回傳的匿名函數就是一個 Iterator。每次 `for-range` 迴圈迭代時，這個匿名函數會被調用，並透過 `yield(a)` 產生下一個費氏數。

## Iterator 的優點

標準化的 Iterator 為 Go 語言帶來了許多好處：

- 彈性與關注點分離 (Flexibility and Separation of Concerns)：呼叫者無需關心序列是如何產生的，只需專注於如何處理資料。例如，一個從 GitHub API 分頁讀取資料的 Iterator。
- 封裝性 (Encapsulation)：Iterator 將資料暴露為序列，這些序列不像 slice 或 map 那樣可以直接被外部修改。
- 效能潛力 (Performance Potential)：Iterator 按需產生元素，而不是一次性將所有資料載入記憶體。這在許多情況下能降低延遲並減少記憶體分配。相較於基於 channel 的 Iterator 實現，其效能也更好。
- 支援無限序列 (Infinite Sequences)：可以表示無限長的序列（例如質數序列），這是像 slice 或 map 這樣的有限資料結構無法做到的。

## Pure vs. Impure Iterators

`iter` 套件的[文件](https://pkg.go.dev/iter)中提到了 single-use iterator 的概念：

> Most iterators provide the ability to walk an entire sequence: when called, the iterator does any setup necessary to start the sequence, then calls yield on successive elements of the sequence, and then cleans up before returning. Calling the iterator again walks the sequence again.
>
> Some iterators break that convention, providing the ability to walk a sequence only once. These “single-use iterators” typically report values from a data stream that cannot be rewound to start over. Calling the iterator again after stopping early may continue the stream, but calling it again after the sequence is finished will yield no values at all. Doc comments for functions or methods that return single-use iterators should document this fact.

這段描述似乎將 Iterator 分為兩類。參考 Julien 的文章 [Pure vs. impure iterators in Go](https://jub0bs.com/posts/2025-05-29-pure-vs-impure-iterators-in-go)，我們可以用 Pure 和 Impure 來更清晰地描述這兩類 Iterator。

### Pure Iterator

Pure Iterator 的行為類似於純函數（Pure Function）。每次調用 Iterator 時，它都會從頭開始重新產生整個序列，並且不會產生外部可觀察的副作用。

我們上面定義的 `fibonacci` 函數產生的就是一個 Pure Iterator。如果我們多次遍歷它，每次都會得到從 0 開始的費氏數列：

```go
package main

import (
  "fmt"
  "iter"
)

func fibonacciPure() iter.Seq[int] {
  return func(yield func(int) bool) {
    // a, b 在 Iterator 函數內部定義
    for a, b := 0, 1; yield(a); a, b = b, a+b {
      // deliberately empty
    }
  }
}

func main() {
  seq := fibonacciPure()

  fmt.Println("First iteration:")
  for n := range seq {
    if n > 10 {
      break
    }
    fmt.Printf("%d ", n)
  }
  fmt.Println() // Output: 0 1 1 2 3 5 8

  fmt.Println("Second iteration:")
  for n := range seq {
    if n > 20 {
      break
    }
    fmt.Printf("%d ", n)
  }
  fmt.Println() // Output: 0 1 1 2 3 5 8 13 21
}
```

`fibonacciPure` 中的變數 `a` 和 `b` 是在回傳的 Iterator 函數內部宣告的。因此，每次 `range seq` 開始時，`a` 和 `b` 都會被重新初始化為 `0` 和 `1`。

### Impure Iterator (or Single-Use Iterator)

Impure Iterator 則不同，它們通常會「記住」上次迭代停止的位置。當再次調用（或繼續迭代）時，它們會從中斷的地方開始，而不是從頭開始。這種類型的 Iterator 通常與無法「倒帶」的資料流（如網路請求、檔案讀取）或需要在多次調用間保持狀態的場景相關。

`iter` 套件文件中的 single-use iterator 很大程度上描述了 Impure Iterator 的一種特性。

考慮以下 `fibonacciImpure` 的例子：

```go
package main

import (
  "fmt"
  "iter"
)

func fibonacciImpure() iter.Seq[int] {
  // a, b 在 Iterator 函數外部定義，成為 Iterator 的自由變數 (free variables)
  a, b := 0, 1
  return func(yield func(int) bool) {
    for ; yield(a); a, b = b, a+b { // 注意這裡 a, b 的狀態會被保留
      // deliberately empty
    }
  }
}

func main() {
  seq := fibonacciImpure()

  fmt.Println("First iteration:")
  for n := range seq {
    if n > 10 {
      break
    }
    fmt.Printf("%d ", n)
  }
  fmt.Println() // Output: 0 1 1 2 3 5 8

  fmt.Println("Second iteration (resumes):")
  for n := range seq {
    if n > 100 { // 假設我們想繼續迭代到更大的數
      break
    }
    fmt.Printf("%d ", n)
  }
  fmt.Println() // Output: 13 21 34 55 89
}
```

在 `fibonacciImpure` 中，變數 `a` 和 `b` 是在回傳的 Iterator 函數之外宣告的。這使得 Iterator 成為一個閉包（Closure），它捕獲了 `a` 和 `b`。因此，當第二次 `range seq` 時，迭代會從 `a` 和 `b` 上次保留的狀態（即 `a=13, b=21`）繼續。

這種 Impure Iterator 可以被描述為「可恢復的 (resumable)」。

#### Single-Use 的模糊性

Julien 的文章指出，官方文件對 Iterator 的分類有些模糊。Single-use 這個詞可能無法涵蓋所有 Impure Iterator 的行為。例如，我們可以設計出：

- Usable twice and non-resumable (可使用兩次但不可恢復): 第一次完整迭代，第二次完整迭代，第三次無輸出。
- Usable twice and resumable (可使用兩次且可恢復): 第一次迭代一部分，第二次從中斷處繼續，但總共只能啟動兩次迭代過程。

這些例子顯示，一旦 Iterator 具有內部狀態（即 Impure），其行為模式可以有很多種。

## 設計考量：Pure or Impure?

那麼，我們應該盡可能設計 Pure Iterator 嗎？這取決於具體的場景和設計目標。

### 效能考量

Pure Iterator 通常更容易推理，因為它們沒有隱藏的狀態。在某些情況下，它們也可能具有更好的效能。Julien 的文章以 `strings.Lines` 為例：

Go 1.24.3 中的 `strings.Lines` 原始碼如下，它回傳一個 Impure Iterator，因為它修改了其自由變數 `s`：

```go
// strings/iter.go
func Lines(s string) iter.Seq[string] {
  return func(yield func(string) bool) {
    for len(s) > 0 {
      var line string
      if i := IndexByte(s, '\n'); i >= 0 {
        line, s = s[:i+1], s[i+1:] // s 被修改
      } else {
        line, s = s, ""           // s 被修改
      }
      if !yield(line) {
        return
      }
    }
    return
  }
}
```

由於 `s` 在閉包中被修改，它會逃逸到 heap 上 (heap allocation)。

如果將其改為 Pure Iterator，在 Iterator 內部操作 `s` 的副本：

```go
  import "strings"

  func LinesPure(s string) iter.Seq[string] {
    return func(yield func(string) bool) {
      sCopy := s // 操作 s 的副本
        for len(sCopy) > 0 {
          var line string
          if i := strings.IndexByte(sCopy, '\n'); i >= 0 {
            line, sCopy = sCopy[:i+1], sCopy[i+1:]
          } else {
            line, sCopy = sCopy, ""
          }
          if !yield(line) {
              return
          }
        }
        return
    }
  }
```

這樣，原始的 `s` 不會逃逸到 heap 上，可能減少一次記憶體分配。

### 一致性考量

然而，效能並非唯一的考量。與相關 API 的行為保持一致性也很重要。例如，`bytes.Lines` 與 `strings.Lines` 功能類似，但操作的是 `[]byte`。

```go
  import "bytes"

// bytes/iter.go
func BytesLines(s []byte) iter.Seq[[]byte] { // 函數名修改以避免與 strings.Lines 衝突
  return func(yield func([]byte) bool) {
    for len(s) > 0 {
      var line []byte
      if i := bytes.IndexByte(s, '\n'); i >= 0 {
        line, s = s[:i+1], s[i+1:] // s 被修改
      } else {
        line, s = s, nil          // s 被修改
      }
      // line[:len(line):len(line)] 確保回傳的 slice 不會與原始 s 的底層陣列有意外的共享
      if !yield(line[:len(line):len(line)]) {
        return
      }
    }
    return
  }
}
```

由於 slice 是可變的，即使在 Iterator 內部創建 `s` 的副本 (淺拷貝)，如果外部仍然持有原始 slice 的引用並修改它，Pure Iterator 的行為也可能受到影響。要實現 `bytes.Lines` 的 Pure Iterator，可能需要對底層陣列進行深拷貝，這通常會違背使用 Iterator 以提升效能的初衷。

因此，如果 `bytes.Lines` 難以設計為高效的 Pure Iterator，那麼 `strings.Lines` 保持 Impure 以維持 API 的一致性，也是一個合理的設計選擇。

## Golang Iterator 與 `samber/lo` 等工具庫的比較

在 Go 1.18+ 泛型出現後，除了標準庫的 `iter` 套件，社群也出現了如 [`samber/lo`](https://github.com/samber/lo) 這樣強大的工具庫，它提供了大量類似 [Lodash](https://lodash.com/) 風格的輔助函數，用於操作 slice、map 等集合。

### `samber/lo` 的特色

- 豐富的 API：`lo` 提供了如 `Map`, `Filter`, `Reduce`, `Uniq`, `GroupBy` 等數十種常用的集合操作函數，極大簡化了程式碼。
- 基於泛型：充分利用 Go 1.18+ 的泛型特性，提供型別安全的集合操作。
- 立即求值 (Eager Evaluation)：`lo` 中的函數通常會直接處理輸入的集合，並立即回傳一個新的、經過處理的集合。例如，`lo.Map` 會遍歷整個輸入 slice，並回傳一個包含所有映射結果的新 slice。

```go
import (
  "fmt"
  "github.com/samber/lo"
)

func main() {
  numbers := []int{1, 2, 3, 4}

  // 使用 lo.Map 將數字轉為字串
  strs := lo.Map(numbers, func(x int, index int) string {
    return fmt.Sprintf("item-%d", x)
  })
  fmt.Println(strs) // Output: [item-1 item-2 item-3 item-4]

  // 使用 lo.Filter 過濾偶數
  evens := lo.Filter(numbers, func(x int, index int) bool {
    return x%2 == 0
  })
  fmt.Println(evens) // Output: [2 4]
}
```

### 與 Golang Iterator 的核心差異

1. 求值策略：
  - `samber/lo`：通常是立即求值。整個集合被處理，結果立即產生。
  - Golang Iterator (`iter.Seq`)：是惰性求值 (Lazy Evaluation)。元素僅在迭代過程中被請求時才逐個產生。這對於大型資料集或無限序列非常重要，因為不需要一次將所有資料載入記憶體。

2. 核心目標：
  - `samber/lo`：提供一套豐富的、便捷的、用於轉換和操作現有集合的工具函數。
  - Golang Iterator：提供一個標準化的「迭代」機制和協議。它更側重於如何「產生」和「消耗」序列，而不是直接提供大量的轉換函數。標準庫 `iter` 套件本身提供的轉換函數相對較少，但其設計允許在其之上構建更複雜的惰性操作。

3. 資源消耗：
  - `samber/lo`：由於是立即求值並通常會創建新的集合來存放結果，對於非常大的集合，可能會消耗較多記憶體。
  - Golang Iterator：由於是惰性求值，可以逐個處理元素，潛在地減少峰值記憶體使用，特別是在鏈式操作中，中間結果不需要完全物化。

### 何時選擇？

- 選擇 `samber/lo` (或其他類似工具庫)：
  - 當我們已經有一個具體的集合 (如 slice 或 map)。
  - 需要進行常見的集合轉換 (map, filter, reduce 等)，並且希望程式碼簡潔易讀。
  - 操作的資料集大小可控，立即求值的記憶體開銷可以接受。
  - 追求開發效率，快速完成集合操作邏輯。

- 選擇 Golang Iterator (`iter.Seq`)：
  - 需要處理可能非常大或無限的序列。
  - 希望實現惰性計算，按需產生資料，以優化效能和資源使用。
  - 需要自訂複雜的序列生成邏輯。
  - 設計可組合的資料處理管道，其中每個步驟都是惰性的。
  - 期望 `iter.Seq` 作為輸入或輸出的 API 整合。

實際上，兩者並非完全互斥。`samber/lo` 可以用於準備或最終處理 Iterator 產生的資料。例如，我們可以使用 Iterator 高效地從資料來源讀取和初步過濾資料，然後將結果分塊收集到 slice 中，再使用 `samber/lo` 進行更複雜的轉換或分組。

總結來說，`samber/lo` 提供了豐富的「瑞士刀」般的集合操作工具，而 Golang Iterator 則提供了一種更底層、更具彈性的惰性序列處理機制。理解它們的設計哲學和核心差異，可以幫助我們在不同場景下做出更明智的技術選型。

## 結論

Golang 中 Iterator 的標準化為開發者提供了更強大、更靈活的工具來處理序列資料。理解 Pure Iterator 和 Impure Iterator 的差異及其各自的適用場景，以及與 `samber/lo` 這類工具庫的比較，有助於我們根據具體需求做出更合適的設計決策。

目前圍繞 Iterator 的慣例仍在發展中，隨著社群的實踐與探索，相信未來會有更多清晰的最佳實踐浮現。

## 參考資料

- [Go Blog: Iterators](https://go.dev/blog/range-functions)
- [Julien: Pure vs. impure iterators in Go](https://jub0bs.com/posts/2025-05-29-pure-vs-impure-iterators-in-go)
- [Go `iter` package documentation](https://pkg.go.dev/iter)
- [GitHub: samber/lo](https://github.com/samber/lo)
