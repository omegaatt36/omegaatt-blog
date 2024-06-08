---
title: Golang 1.22 中 http routing 的改進
date: 2024-03-31
categories:
 - develop
tags:
 - golang
---

Golang 作為一個偏向 server 應用的程式語言，一般的 web server 並不會直接使用原生的 package `net/http`，而更多的使用 `gin-gonic/gin` 或是 `gorilla/mux`，後來也有 `labstack/echo` 以及 `go-chi/chi` 等等選擇，在效能、輕量、好維護、好擴充中，都能找到對應的 third party package，其中的原因不外乎是原生的 package 提供的功能過於簡潔。

好在 1.22 中，官方改進了 `net/http` 中對於多工器、路由，甚至出了一篇[部落格](https://go.dev/blog/routing-enhancements)，現在更可以「大膽的」直接使用 standard library。

## Path Parameter

若要將應用的 Web API 定義成 RESTful，我們會使用 `/資源/{資源唯一識別符}/子資源/{子資源唯一識別符}` 來定義路徑。假如要獲取一個使用者的訂單，則會使用 `GET /users/1/orders` 來獲取。在 1.22 以前，我們只能定義到 `/users`，再自行解析往後的 path：

```go
http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    subPath := strings.TrimPrefix(req.URL.Path, "/users/")
    if len(subPath) == 0 {
        xxx
    } else {
        ooo
    }
    ...
})
```

而在 1.22 中新增了 `net.http` 對 path parameter 的支持，我們可以直接使用 `(*http.Request).PathValue("xxx")` 來獲取：

```go
http.HandleFunc("/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("user_id")
    ...
})
```

不過也帶來一些限制

### 相容 go 1.x

眾所周知 Golang 是一個極度在意簡單性、向後相容的程式語言，為了不要因為升到 1.22 而發生非預期的錯誤，是可以讓 path parameter 的路由與一般路由並存的。

> The precedence rule is simple: the most specific pattern wins. This rule matches our intuition that posts/latests should be preferred to posts/{id}, and /users/{u}/posts/latest should be preferred to /users/{u}/posts/{id}. It also makes sense for methods. For example, GET /posts/{id} takes precedence over /posts/{id} because the first only matches GET and HEAD requests, while the second matches requests with any method.

舉例來說：

```go
mux.HandleFunc("/orders/{order_id}", xxx)
mux.HandleFunc("/orders/latest",xxx)
```

若是使用 gin 的話，這種路由註冊將在 runtime 出現 panic，也是由於 gin 是一個基於 `valyala/fasthttp` 的 package，而 `valyala/fasthttp` 又是基於 radix 這種資料結構，node 間發生了衝突才引發 panic。

`net/http` 則是使其相容於舊版本：

只有在發生模糊不清的路徑時，才會在 runtime 發生 panic：

```go
mux.HandleFunc("/orders/latest",xxx)
mux.HandleFunc("/{other_resource}/latest")
//  pattern "/{other_resource}/latest" (registered at /home/raiven/go-http-22/main.go:110) conflicts with pattern "/orders/{order_id}"
```

更進一步的相容可以打開 `GODEBUG=httpmuxgo121=1` 來使其單獨 rollback 回 1.21。

```go
package http

type ServeMux struct {
    mu       sync.RWMutex
    tree     routingNode
    index    routingIndex
    patterns []*pattern  // TODO(jba): remove if possible
    mux121   serveMux121 // used only when GODEBUG=httpmuxgo121=1
}
```

## 指定 Method

可以直接將 http method 寫在路由判斷內，這個改動相較簡單，卻又能大幅度的減少程式碼，直接寫範例：

before upgrading to 1.22:

```go
mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        xxx
    } else if r.Method == http.MethodPost {
        xxx
    } else {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
})
```

after upgrading to 1.22:

```go
mux.HandleFunc("GET /orders", func(w http.ResponseWriter, r *http.Request) {})
mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {})
```

## 比較編譯大小

如此一來有一些很小的 package 就不在需要引入碩大的 third party package，比較一下不同的 package 在 handle `localhost:8080/hello` 的 binary 大小：

全部基於 linux,amd64,go1.22.0

- `net/http`

```go
package main

import "net/http"

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })
    http.ListenAndServe(":8080", mux)
}
```

- `gin-gonic/gin`

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()

    router.GET("/hello", func(c *gin.Context) {
        c.JSON(http.StatusOK, "Hello, World!")
    })

    http.ListenAndServe(":8080", router)
}

```

- `go-chi/chi`

```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
)

func main() {
    r := chi.NewRouter()

    r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })

    http.ListenAndServe(":8080", r)
}
```

- `gorilla/mux`

```go
package main

import (
    "net/http"

    "github.com/gorilla/mux"
)

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })

    http.ListenAndServe(":8080", r)
}
```

- `labstack/echo`

```go
package main

import (
    "net/http"

    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()

    e.GET("/hello", func(c echo.Context) error {
        return c.String(http.StatusOK, "Hello, World!")
    })

    e.Logger.Fatal(e.Start(":8080"))
}
```

使用 `go build main.go` 來編譯成可執行檔，並使用 `du -sh main` 來查看執行檔大小：
| package  | size |
|----------|------|
| net/http | 6.8M |
| gin-gonic/gin | 11M |
| go-chi/chi | 7.1M |
| gorilla/mux | 7.1M |
| labstack/echo | 7.5M |

假如自己的 side project 每天都要編譯一個 nightly version 的 docker image，使用 gin 將比原生的 net/http 多出 1.5G 的存儲空間。
