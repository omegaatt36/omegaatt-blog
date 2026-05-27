---
title: Go 1.27 即將支援 Generic Method：告別 package-level 函式的 workaround
date: 2026-05-27
categories:
  - develop
tags:
  - golang
  - generics
cover:
  image: "images/cover.webp"
---

## 前言

Go 1.18 帶來泛型之後，我們終於可以寫出型別安全的通用函式，不再需要到處 `interface{}` 加型別斷言。但一個眾所周知的限制一直存在到現在，method 沒辦法有自己的 type parameter。

換句話說，這段程式碼在目前的 Go 版本中是非法的：

```go
func (c *Client) doRequest[T any](ctx context.Context, method, path string, body any) (*T, error) {
    // 編譯錯誤：cannot use generic method doRequest without instantiation
}
```

為了繞過這個限制，大家只能退而求其次，把 receiver 當作普通參數傳進去，改用 package-level 的 generic function。

最近 [Go issue #77273](https://github.com/golang/go/issues/77273) 的進展讓這件事有了轉機，generic method 的設計已被接受，預計在 Go 1.27 正式釋出。

## 為什麼現在沒有 Generic Method？

這不是 Go team 沒想到，而是刻意的設計決策。

Go 的 interface 是核心功能，method 必須能夠被 dispatch 到具體的實作上。如果 method 自身帶有 type parameter，編譯器在 dispatch 時需要知道 `T` 是什麼，這在動態 dispatch（interface dispatch）的情境下是個難題，runtime 無法事先知道要生成哪個型別的實例化版本。

所以泛型剛推出時，generic method 被刻意排除在外，等待一個比較完整的設計方案。這個等待從 Go 1.18 一路等到了 1.27（暫定）。

## 現實中的痛點：以 HTTP Client 為例

我們的 API client 需要處理不同 response 結構，所有 endpoint 共用同一個 `Response[T]` wrapper：

```go
type Response[T any] struct {
    Data  T      `json:"data,omitzero"`
    Error string `json:"error,omitzero"`
}
```

理想的寫法是讓 `*Client` 直接持有 generic method：

```go
func (c *Client) doRequest[T any](ctx context.Context, method, path string, body any) (*T, error) {
    var response Response[T]
    err := c.doRequestWithQuery[T](ctx, method, path, nil, body)
    // ...
    return &response.Data, nil
}
```

但這在現在的 Go 裡行不通。於是現實的 code 長這樣，分成兩層：

```go
// 第一層：package-level generic function，把 client 當參數傳進來
func doRequest[T any](ctx context.Context, client *Client, method, path string, body any) (*T, error) {
    var response Response[T]
    err := client.doRequest(ctx, method, path, body, &response)
    if err != nil {
        return nil, err
    }
    return &response.Data, nil
}

func doRequestWithQuery[T any](ctx context.Context, client *Client, method, path string, query url.Values, body any) (*T, error) {
    var response Response[T]
    err := client.doRequestWithQuery(ctx, method, path, query, body, &response)
    if err != nil {
        return nil, err
    }
    return &response.Data, nil
}

// 第二層：真正的 method，但沒有 type parameter，只能接受 any
func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
    return c.doRequestWithQuery(ctx, method, path, nil, body, result)
}
```

一層負責型別安全的反序列化（package-level generic），一層負責實際的 HTTP 請求（non-generic method）。每次新增 endpoint，caller 得這樣呼叫：

```go
func (a *APIClient) GetResourceA(ctx context.Context, req RequestGetResourceA) (*ResponseGetResourceA, error) {
    path := fmt.Sprintf("/api/v1/a/%s", req.ID)
    response, err := doRequest[responseGetResourceA](ctx, a.client, http.MethodGet, path, body)
    // ...
}
```

第一眼看到這段的人通常都會問：「為什麼 `doRequest` 是一個吃 client 的獨立函式，不是 `client` 的 method？」這就是 generic method 缺失帶來的認知成本，你需要額外解釋一個不自然的設計決策。

此外，因為 package-level 的 `doRequest` 和 method `(c *Client) doRequest` 名稱相同，只靠大小寫和呼叫方式來區分，閱讀時容易混淆，命名上也很容易撞牆。

## Generic Method 到來之後

有了 generic method，兩層結構可以合併，讓 `*Client` 直接持有 generic method：

```go
func (c *Client) doRequest[T any](ctx context.Context, method, path string, body any) (*T, error) {
    var response Response[T]
    err := c.doRequestWithQuery(ctx, method, path, nil, body, &response)
    if err != nil {
        return nil, err
    }
    return &response.Data, nil
}

func (c *Client) doRequestWithQuery[T any](ctx context.Context, method, path string, query url.Values, body any) (*T, error) {
    var response Response[T]
    err := c.doRequestRaw(ctx, method, path, query, body, &response)
    if err != nil {
        return nil, err
    }
    return &response.Data, nil
}
```

caller 的寫法從「傳 client 進去的函式」變成「呼叫 client 的 method」：

```go
// before: package-level function，client 當參數
response, err := doRequest[responseGetResourceA](ctx, a.client, http.MethodGet, path, body)

// after: method call
response, err := a.client.doRequest[responseGetResourceA](ctx, http.MethodGet, path, body)
```

差別看起來小，但對 API 設計的影響不小。

一、可見性更清晰。`doRequest` 現在是 `*Client` 的 method，IDE 的 autocomplete 會列出它，不再需要記住有哪些 package-level helper 函式。
二、封裝更乾淨。之前 package-level `doRequest` 和 method `doRequest` 同名但型態不同，現在都是 method，命名空間清楚多了，也不會再有「等等，這個 `doRequest` 是 function 還是 method？」的困惑。
三、更符合 Go 的慣用法。Go 鼓勵把行為綁定在型別上，generic method 讓這個原則在泛型場景下也能成立，不再需要因為語言限制而繞道而行。

## 多型的角度：Generic Method 的潛力與邊界

有人可能會問，generic method 可以放進 interface 嗎？這樣就能對不同的 client 實作進行多型了？

答案是：有但有邊界。

Go 的 interface 以 runtime dynamic dispatch 為核心設計，如果 interface 裡有 generic method，runtime 無法事先知道該 dispatch 到哪個具體的型別實例化版本。所以帶有 generic method 的型別，無法直接用來滿足帶有同名 generic method 的「普通 interface 變數」。

如果你需要對 HTTP client 進行多型（例如區分 real client 和 mock），目前比較實際的做法還是讓 interface 維持 non-generic：

```go
// interface 仍然使用 any，不帶 type parameter
type Requester interface {
    doRequestRaw(ctx context.Context, method, path string, query url.Values, body any, result any) error
}
```

generic method 的多型價值更多展現在「type constraint」的層面。在 generic function 裡，可以透過 constraint 聲明「我需要一個支援某個 generic method 的型別」：

```go
type CanRequest[T any] interface {
    DoRequest[T](ctx context.Context, method, path string, body any) (*T, error)
}

func fetchAll[C CanRequest[T], T any](ctx context.Context, client C, paths []string) ([]*T, error) {
    results := make([]*T, 0, len(paths))
    for _, p := range paths {
        r, err := client.DoRequest[T](ctx, http.MethodGet, p, nil)
        if err != nil {
            return nil, err
        }
        results = append(results, r)
    }
    return results, nil
}
```

這讓 library 作者在設計通用工具函式時，能以更精確的 constraint 描述所需的能力，而不是傳一個吃 `any` 的函式指標進去。

## 目前的限制

Generic method 雖然呼之欲出，但還是有幾點值得注意。

Generic method 無法用於 interface 的 dynamic dispatch，這是架構性限制，不是 bug。如果你的設計高度依賴 interface abstraction 來做 dependency injection（例如在測試中替換 client），generic method 只是讓實作層更乾淨，interface 層還是得維持原本的 non-generic 設計。

另外，預計在 Go 1.27 釋出，但在正式 release 之前細節可能還會有調整，不建議現在就把生產代碼的設計押注在特定語法上。等 RC 出來再說。

## 寫在最後

Generic method 是 Go 泛型拼圖中一直缺席的一塊。它不會從根本上改變我們寫 Go 的方式，但會把那些現在感覺「明明應該這樣寫但寫不了」的 workaround 清掉，讓 API 設計更直覺，少解釋一些「為什麼這裡要這樣繞」的歷史包袱。

以前跟同事解釋「為什麼 `doRequest` 是一個吃 client 的函式而不是 client 的 method」這件事，希望從 Go 1.27 開始可以從解釋名單上劃掉。

---

關於「dynamic dispatch」這個術語，是物件導向程式語言中普遍存在的概念，簡單說就是「在 runtime 才決定要呼叫哪個函式的機制」。

在 Go 裡，interface 的方法呼叫就是 dynamic dispatch。當你透過 `io.Reader` 呼叫 `Read` 時，Go runtime 在執行期查找該變數底層的實際型別，再找到對應的 `Read` 實作來執行，這個查找是有成本的（通常透過 itab 指標）。

對應地，generic function 使用的是「static dispatch」，也叫 monomorphization：編譯器在編譯時根據每個具體的型別參數，各自生成獨立的程式碼，執行時不需要查找，效能更好，代價是二進制大小可能增加。

Rust 的 `dyn Trait` 對應 dynamic dispatch，`impl Trait` / `T: Trait` 對應 static dispatch，概念和 Go 一樣，只是語法不同。Go 的泛型走的就是 static dispatch 那條路，所以 generic method 的限制（無法用於 interface dynamic dispatch）也就不難理解了。

## 參考資料

- [proposal: spec: allow methods to have type parameters #49206](https://github.com/golang/go/issues/49206)
- [spec: generic methods #77273](https://github.com/golang/go/issues/77273)
- [Go Generics Specification](https://go.dev/ref/spec#Type_parameters)
- [Go 1.18 Release Notes - Generics](https://go.dev/doc/go1.18#generics)
- [本文範例程式碼 client.go](client.go)
