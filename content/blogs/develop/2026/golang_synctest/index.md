---
title: Golang 1.25 testing/synctest 初體驗：告別在測試中寫 time.Sleep 的日子
date: 2026-01-18
categories:
  - develop
tags:
  - golang
  - optimization
  - testing
---

## 前言

在我們目前的應用程式架構中，由於高度與 Kubernetes 耦合，服務啟動與運作期間需要頻繁地去讀取 K8s 中的 ConfigMap。為了達成配置熱更新（Hot Reload），我們引入了 Kubernetes client-go 中的 [informers](https://pkg.go.dev/k8s.io/client-go/informers) 機制來監聽 ConfigMap 的 CRUD 事件。

雖然 K8s 官方提供了 fake client 讓我們能測試 informers 的邏輯，但在 Service Code 的層級，我們往往需要封裝一層更適合業務邏輯的 `ConfigWatcher`。Golang 引以為傲的輕量級 Goroutine 與 Channel 搭配非常適合用來處理這種非同步的事件傳遞。

然而，一旦涉及到 Goroutine 的非同步測試，「時間」往往就成了最大的敵人。

## 遇到的問題：不穩定的測試與魔法數字

為了模擬 ConfigMap 的變更通知，我們定義了一個 `ConfigMapWatcher` 介面與對應的 Event 結構：

```go
const (
	ConfigMapUpdateEventTypeAdded ConfigMapUpdateEventType = iota
	ConfigMapUpdateEventTypeModified
	ConfigMapUpdateEventTypeDeleted
)

type ConfigMapUpdateEvent struct {
	Name  string
	Type  ConfigMapUpdateEventType
	Value map[string]string
}

type ConfigMapWatcher interface {
	Watch(ctx context.Context, eventCh chan<- ConfigMapUpdateEvent) error
}
```

接著，我們很自然地在 testing 中實作了一個 fake 物件來模擬事件發送：

```go
type fakeConfigMapWatcher struct {
	injectCh  chan ConfigMapUpdateEvent
	watchErr  error
	watchOnce sync.Once
}

func newFakeConfigMapWatcher() *fakeConfigMapWatcher {
	return &fakeConfigMapWatcher{
		injectCh: make(chan ConfigMapUpdateEvent),
	}
}

func (f *fakeConfigMapWatcher) sendEvent(event ConfigMapUpdateEvent) {
	f.injectCh <- event
}

func (f *fakeConfigMapWatcher) Watch(ctx context.Context, eventCh chan<- ConfigMapUpdateEvent) error {
	if f.watchErr != nil {
		return f.watchErr
	}

	f.watchOnce.Do(func() {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case e := <-f.injectCh:
					eventCh <- e
				}
			}
		}()
	})
	return nil
}
```

問題來了，當我們在寫單元測試時，呼叫 `sendEvent` 將事件送入 channel 後，消費者端（也就是我們的業務邏輯 Goroutine）並不會「立刻」收到並處理完成。為了確保 `assert` 斷言執行時，業務邏輯已經跑完了，我們被迫在測試中加入 `time.Sleep`：

```go
func TestHandleConfigMapUpdate(t *testing.T) {
    fakeWatcher := newFakeConfigMapWatcher()
    
    // ... other init & start watching

    fakeWatcher.sendEvent(ConfigMapUpdateEvent{
        Type:  ConfigMapUpdateEventTypeAdded,
        Name:  "test-configmap",
        Value: map[string]string{"config": scannerYAML},
    })

    // 這裡的 100ms 就是所谓的 "Magic Number"
    time.Sleep(100 * time.Millisecond)

    assert.True(t, handlerCalled, "handler should be called after event")
}
```

這種作法有兩個顯著的缺點：
1. **測試變慢**：每個測試都要等 100ms，累積起來 CI 的時間會顯著增加。
2. **Flaky Tests**：在本地跑可能 100ms 夠用，但到了資源吃緊的 CI Runner 上，CPU 稍微忙一點，100ms 可能就不夠了，導致測試偶發性失敗。

雖然我們可以透過在 fake 物件中增加 `<-done` channel 來通知測試程式說「我處理好了」，但這會讓測試用的 fake 物件邏輯變得複雜，甚至為了測試而入侵產品代碼的設計，這並不是我們樂見的。

## 解決方案：Golang 1.24+ `testing/synctest`

在 Golang 1.24 中，官方引入了一個實驗性 package `testing/synctest`，這正是為了解決非同步測試難題而生的。而在隨後的 **Golang 1.25** 中，API 進行了一次 Breaking Change，將原本的 `synctest.Run` 改為與 `testing.T` 綁定更深的 `synctest.Test`，以提供更完整的測試整合。

它的核心概念是引入了一個「Bubble（氣泡）」環境。在這個氣泡中，時間是虛擬的，而且 `synctest` 能夠感知到所有 Goroutine 的狀態。

我們只需要用 `synctest.Test` 將測試邏輯包起來，並將原本的 `time.Sleep` 替換成 `synctest.Wait()`：

```go
import "testing/synctest"

func TestHandleConfigMapUpdate(t *testing.T) {
    // 使用 synctest.Test 建立一個隔離的 Bubble (Go 1.25+)
    synctest.Test(t, func(t *testing.T) {
        fakeWatcher := newFakeConfigMapWatcher()
        
        // ... other init

        fakeWatcher.sendEvent(ConfigMapUpdateEvent{
            Type:  ConfigMapUpdateEventTypeAdded,
            Name:  "test-configmap",
            Value: map[string]string{"config": scannerYAML},
        })

        // 移除 time.Sleep，改用 synctest.Wait()
        // time.Sleep(100 * time.Millisecond) 
        synctest.Wait()

        assert.True(t, handlerCalled, "handler should be called after event")
    })
}
```

### 為什麼這樣就不用 Sleep 了？

`synctest.Wait()` 的機制非常聰明，它會暫停當前 Goroutine，直到 Bubble 內的所有其他 Goroutine 都進入 **Durably Blocked**（持久阻塞）狀態。

所謂 Durably Blocked，指的是 Goroutine 正在等待某些只能由 Bubble 內其他 Goroutine 觸發的事件（例如等待 channel 接收、select 等）。當所有人都卡住了，表示目前的非同步任務都已經「推進」到極限了，這時 `Wait()` 就會返回，我們就可以放心地進行斷言。

更棒的是，Bubble 內的 `time` package 是被 mock 過的。如果你的代碼裡有 `time.Sleep(5 * time.Second)`，在 `synctest` 的環境下，它不會真的睡 5 秒，而是直接快轉時間，這讓測試速度有了質的飛躍。

## 怎樣可以更好 & 反思

`synctest` 在 1.24 還是一個實驗性功能，需要透過 `GOEXPERIMENT=synctest` 環境變數來啟用。[1.25 才正式釋出](https://go.dev/doc/go1.25#new-testingsynctest-package)。

### Go 1.24 vs 1.25 的 API 變動

值得注意的是，如果你是在 Go 1.24 剛推出的時候關注這個功能，你可能看過 `synctest.Run(func() { ... })` 這樣的用法。但在 Go 1.25 中，為了更好地整合測試框架（例如在測試失敗時正確清理資源、支援 subtest 等），官方將 API 修改為 `synctest.Test(t, func(t *testing.T) { ... })`。原本的 `synctest.Run` 已被標記為 deprecated 並且[預計在 Go 1.26 移除](https://github.com/golang/go/issues/74012)。這提醒我們在使用實驗性功能時，必須隨時準備好應對 Breaking Changes。

此外，使用 `synctest` 也有一些限制需要注意：

1. **外部 I/O 的不確定性**：如果你的 Goroutine 阻塞在網路 I/O（例如真實的 HTTP 請求）或 System Call 上，`synctest` 可能無法準確判斷這是否為 Durably Blocked，因為外部事件隨時可能喚醒它。因此，使用 `synctest` 時，盡量搭配 Mock 物件（如 `net.Pipe` 或 fake client）來隔離外部依賴。
2. **Mutex 的處理**：標準庫的 `sync.Mutex` 目前不被視為 Durably Blocked 的條件。這是因為 Mutex 通常持有時間很短，且可能被 Bubble 外的 Goroutine 影響。

### 總結

`testing/synctest` 的出現，填補了 Go 語言在複雜併發測試上的一塊拼圖。它讓我們不再需要在「寫死 Sleep 時間」與「撰寫複雜同步邏輯」之間做痛苦的抉擇。

對於像我們這種高度依賴 Event-Driven 架構與 Kubernetes Informer 的應用來說，這無疑是一個巨大的優化。測試變得更快、更穩，代碼也更乾淨了。

建議大家可以在一些非核心的測試中先嘗試引入，體驗一下「瞬間完成非同步測試」的快感。

這篇文章內的程式碼可以到 [The Go Playground](https://go.dev/play/p/22vr_saI1Yx) 上查看，或是到 [demo code](main.go) 中查看。

```
❯ go run main.go
Starting real sleep test...
TestWithRealSleep exec time: 100.67325ms
Starting synctest...
TestWithSynctest exec time: 42.166µs
PASS
```

## 參考資料

- [Testing concurrent code with testing/synctest - The Go Blog](https://go.dev/blog/synctest)
- [Testing time: Changes from 1.24 to 1.25](https://go.dev/blog/testing-time#changes-from-124-to-125)
- [Go 1.24 Release Notes](https://tip.golang.org/doc/go1.24)
- [Go 1.25 Release Notes](https://tip.golang.org/doc/go1.25)
