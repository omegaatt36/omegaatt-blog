---
title: Golang 1.26 新特性在數量上史無前例的多
date: 2026-01-13
categories:
  - develop
tags:
  - golang
  - release
---

隨著時間來到 2026 年初，Go 語言迎來了 1.26 版本的更新。如果說 Go 1.18 的泛型是語言層面的重大變革，那麼 Go 1.26 則是在「數量」與「廣度」上讓人感到驚艷的一次釋出。從語言特性的語法糖、標準庫的實用擴充，到 Runtime 效能的顯著提升（Green Tea GC），甚至是實驗性的 SIMD 支援，這次的更新內容豐富到讓人目不暇給。

本文將挑選其中幾個我認為對日常開發最重要、或最有趣的改動來進行介紹。

## 語言層面的改動

### `new(expr)`：終於不用再寫輔助函數了

在 Go 1.26 之前，如果我們想要取得一個基本型別（如 `int`, `bool`, `string`）的 pointer，通常需要宣告一個變數或者寫一個輔助函數。這在定義 struct 的 literal 時特別煩人，尤其是當 struct 欄位是 `*bool` 或 `*int` 用來區分「零值」與「未設定」的時候。

回顧過去，我們為了這個小需求付出了不少努力：

* 在 Go 1.18 泛型出現之前：我們經常需要定義一堆如 `Int64Ptr(v int64) *int64` 或 `Float64Ptr(v float64) *float64` 的輔助函數（AWS SDK 的使用者應該對此非常熟悉）。
* 匿名函數大法：如果不想要定義全域的輔助函數，有時甚至會看到像 `enabled := func(b bool) *bool { return &b }(true)` 這種冗長且難讀的寫法。
* 泛型時代：雖然可以用一個通用的 `ptr[T]` 解決，但還是需要額外的程式碼。

以前我們可能需要這樣做：

```go
func ptr[T any](v T) *T {
    return &v
}

type Config struct {
    Enabled *bool
}

conf := Config{
    Enabled: ptr(true),
}
```

在 Go 1.26 中，內建的 `new` 函數得到了增強，現在它不僅接受型別，還可以直接接受表達式（Expression）。

```go
// 直接取得指向數值 42 的 pointer
p := new(42)
fmt.Println(*p) // 42

type Cat struct {
    Name string `json:"name"`
    Fed  *bool  `json:"is_fed"`
}

// 直接在 struct literal 中使用
cat := Cat{
    Name: "Mittens",
    Fed:  new(true),
}
```

這個小小的改動極大地提升了開發體驗，特別是在處理 JSON 序列化或是 Optional 欄位時，程式碼變得乾淨許多。

### Recursive Type Constraints (遞迴型別約束)

泛型的約束現在可以遞迴地引用自身了。這讓定義像 `Ordered` 這樣的介面變得更直觀，允許型別 `T` 與同為 `T` 的值進行比較。

```go
type Ordered[T Ordered[T]] interface {
    Less(T) bool
}
```

雖然日常業務邏輯中不一定常寫這種 code，但對於設計泛型庫的開發者來說，這是一個補全泛型拼圖的重要功能。

例如可以這樣寫：

```go
type Tree[T Ordered[T]] struct {
    nodes []T
}

// netip.Addr has a Less method with the right signature,
// so it meets the requirements for Ordered[netip.Addr].
t := Tree[netip.Addr]{}
_ = t
```

## 標準庫的增強

### 更安全的錯誤檢查：`errors.AsType`

`errors.As` 一直以來都是檢查錯誤型別的標準做法，但它依賴 reflect 並且需要傳入 pointer 的 pointer，寫起來容易出錯（例如傳了 `nil` 或錯誤的型別）。

Go 1.26 引入了 `errors.AsType`，這是一個泛型函數，提供了編譯時期的型別安全檢查，而且效能更好。

```go
// 舊寫法：容易 runtime panic 或寫錯
// var target *AppError
// if errors.As(err, &target) { ... }

// 新寫法：Type-safe
if target, ok := errors.AsType[*AppError](err); ok {
    fmt.Println("application error:", target)
}
```

### Log 套件的進化：`slog.NewMultiHandler`

`log/slog` 在 Go 1.21 引入後已成為標準，我自己新的專案在效能允許的情況也都使用 slog 搭配自訂 Handler（極限性能還是會考慮使用 uber 的 zap）。

現在 Go 1.26 為它補上了一個常用的功能：`MultiHandler`。以往我們如果想同時將 log 輸出到 stdout 和檔案，往往需要自己實作 Handler 或依賴第三方庫。現在標準庫直接支援了：

```go
stdoutHandler := slog.NewTextHandler(os.Stdout, nil)
fileHandler := slog.NewJSONHandler(logFile, nil)

// 同時輸出到兩個地方
multiHandler := slog.NewMultiHandler(stdoutHandler, fileHandler)
logger := slog.New(multiHandler)

logger.Info("system started", slog.Int("pid", os.Getpid()))
```

### 其他值得注意的標準庫更新

* `net/http` Context-aware Dialing: `net.Dialer` 新增了 `DialTCP`, `DialUDP` 等方法，並且直接支援 `context.Context`，這讓網路連線的控制與超時管理更加一致且高效。
* `bytes.Buffer.Peek`: 終於可以直接偷看 Buffer 中的下 N 個 bytes 而不移動讀取指針了。
* `crypto/hpke`: 正式支援 RFC 9180 Hybrid Public Key Encryption (HPKE)，這是一種現代化的加密標準，結合了非對稱與對稱加密的優點。

## 開發與除錯：不再害怕 Goroutine 洩漏

對於高併發服務的開發者來說，Goroutine Leak 往往是夢魘。在 Go 1.26 之前，我們通常依賴 `uber-go/goleak` 在測試階段攔截，或是等到生產環境記憶體暴漲後，Dump 出成千上萬個 Goroutine stack trace 來大海撈針。

Go 1.26 在 `runtime/pprof` 中引入了全新的 `goroutineleak` profile。不同於一般的 `goroutine` profile 只是列出當前所有存活的 Goroutine，這個新工具會智慧地找出那些「卡在同步原語（如 channel 讀寫、Mutex 等待）」且「無法從其他 runnable goroutine 到達」的孤兒 Goroutine。

```go
import "runtime/pprof"

// 在測試或除錯 endpoint 中使用
// debug=1 會輸出可讀的文字格式
pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
```

這個功能目前可以透過 `GOEXPERIMENT=goroutineleakprofile` 在編譯時開啟（預計在 Go 1.27 會預設開啟），這絕對是排查「服務跑了一週後記憶體緩慢變大」這類問題的殺手鐧。

## Runtime 與效能提升

### Green Tea GC

綠茶聽起來很好喝？這其實是 [Go 1.25 實驗性引入](https://go.dev/blog/greenteagc)，並在 1.26 正式預設啟用的新垃圾回收器。Green Tea GC 專為現代多核心 CPU 設計，針對 8 KiB 的記憶體區塊（spans）進行掃描優化，並能同時處理多個物件以減少 cache miss。

根據官方數據，這能降低 10-40% 的 GC overhead。對於高併發的後端服務來說，這意味著在不改一行 code 的情況下，服務的延遲（latency）可能會顯著降低。

### Cgo 與 Syscall 的瘦身

Go 1.26 移除了 `_Psyscall` 的處理器狀態，簡化了 runtime 內部的路徑。這直接導致 cgo 的呼叫開銷降低了約 30%。如果你的專案大量依賴 C library（例如影像處理、加解密等），升級到 1.26 應該會有感。目前我們團隊的專案最核心的功能仰賴 cgo，希望能藉此獲得免費的效能提昇（雖然可執行檔大小會變大）。

```go
func BenchmarkSyscall(b *testing.B) {
    for b.Loop() {
        _, _, _ = syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0)
    }
}
```

## 實驗性功能 (Experimental)

Go 1.26 也帶來了一些令人興奮的實驗性套件，雖然 API 可能還會變動，但展示了 Go 未來的可能性。

### `simd/archsimd`：向量化運算

對於需要極致效能的場景，Go 終於開始提供低階的 SIMD 存取能力（目前針對 `amd64`）。不過即便有 SIMD 來加速矩陣運算，也無法撼動 Python/Pytorch/CUDA 目前在 LLM 的地位就是了（本來就不是同樣的東西），CPU 的 AVX-512 再快也快不過 GPU，但或許可以加速一些邊緣運算諸如 Vector DB 的 Embedding 與 Tokenization。

```go
// 概念範例
func Add(a, b []float32) []float32 {
    // 使用 archsimd 進行向量加法...
}
```

### `runtime/secret`：敏感資料保護

在處理私鑰或密碼時，我們希望能確保這些資料在使用後立即從記憶體中抹除，避免被 swap 到磁碟或是被 core dump 抓出來。`runtime/secret` 提供了 `secret.Do`，保證在函數執行完畢後，相關的暫存器與 stack 都會被清空。這在 ISO/PCI 等等 compliance 都十分重要，可以大膽的說，用 Go 更安全了(?)。

```go
secret.Do(func() {
    // 在這裡生成 ephemeral key
    // 離開後自動抹除痕跡
})
```

## 總結

Go 1.26 給我的感覺是「穩中求進，兼顧細節」。它沒有像泛型那樣徹底改變我們寫 Go 的方式，但在開發者體驗（`new(expr)`, `errors.AsType`）與底層效能（GC, Cgo, Allocator）上都做出了巨大的貢獻。

特別是 `new(expr)` 和 `errors.AsType`，我相信會迅速成為大家日常 coding 的標準配備。而效能的免費午餐（Free Lunch），對於維護大型 Go 服務的團隊來說，絕對是升級的最大動力。

準備好更新你的 `go.mod` 到 `go 1.26` 了嗎？

## 參考資料

- [Go 1.26 Release Notes](https://tip.golang.org/doc/go1.26)
- [Go 1.26 interactive tour - Anton Zhiyanov](https://antonz.org/go-1-26/)
- [The Green Tea Garbage Collector](https://go.dev/blog/greenteagc)
- [Go feature: Secret mode](https://antonz.org/accepted/runtime-secret/)
