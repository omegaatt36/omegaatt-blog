---
title: Golang 學習筆記 Day 1 - Getting started
date: 2020-08-25
categories:
 - develop
tags:
 - golang
---

``` 
免責聲明，此篇為個人學習筆記，非正式教學文，僅供參考感謝指教，有任何內容錯誤歡迎發 issue。
```

在前一天介紹了為何要使用 golang、安裝過程以及基本的 hello world，這天將重點放在最基本的 golang 程式碼特徵。

## packages

正如昨天關於 hello world 的解析中提到的，每個程式都是由一到多個 packages 組成的，也就是 `.go` 中必須宣告該檔案是哪個 package，且同一個資料夾中僅接受 `xxx` 以及 `xxx_test` 兩種 package名稱，也就是該 package 的名稱以及測試檔的「測試包」。

至於 package 要怎麼命名呢，首先要提到檔案該如何命名，雖然 `go format` 是對於程式碼風格十分嚴格的，但是對於檔案名稱似乎就沒那麼嚴謹，但整個社群中普遍會用 `snake_case` 或是 `dash-case` 作為檔案 (甚至是 package) 的命名。

* aaabbbccc.go
* time-machine.go
* battery_charger.go

但是需要注意的是 go build 會忽略所有以 `_` 或是以 `.` 作為結尾的後綴， [ `underline` 可能會導致在建構時被忽略](https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool)。

再來提到 package 到底要如何命名呢，[官方文檔](https://blog.golang.org/package-names)中明確的提到不要用 `camelCase` 以及 `snake_case` 。

> The style of names typical of another language might not be idiomatic in a Go program. Here are two examples of names that might be good style in other languages but do not fit well in Go:
> - computeServiceClient
> - priority_queue

於是須將諸如 `time-machine.go` 的 package 名稱命名為 `timemachine` 。

``` go
package timemachine
```

而似於 c/cpp，golang 中所有程式都會從 `main` 包中的 `main()` 方法開始執行，故不能像一些腳本語言，隨便一個檔案都能直接開來用，每支程式一定要有 main package 以及 main function，但是 main package 不一定要叫做 `main.go` ，但若是要 push 上 github 可能會不方便別人一眼就看出那個檔案是這支程式的入口點。

## import

golang 中導入其他 package 便是用 import 語句，通常直接宣告在 package 下面，可以使用逐行宣告：

``` go
import "fmt"
import "math"
```

也可以使用小括弧組合在一起：

``` go
import (
    "fmt"
    "math"
)
```

## public/private 

golang 中有一個十分有趣又簡潔的設計，熟悉 OOP 肯定會知道 `public` 、 `private` 以及 `protected` ，雖然 golang 並不是一個物件導向語言，沒有 `protected` ，但是仍有公有、私有變數以及方法的區別。

首先需要知道的是，golang 中的變數命名一律使用 `CamelCase` ，而公有私有就是手字大小寫：

* IsHuman() 是一個 public function
* queryUsers() 是一個 private function

這個公有私有的作用域僅限該 package，也就是說連同個資料夾中的 `package xxx_test` 測試包都無法使用該變數/方法。

## functions

方法的宣告也十分簡單，格式為：

``` go
func functionName1(param type) returnType {
    var xxx returnType
    ...
    return xxx
}

func functionName2(param type)  {
    ...
    return
}
```

或是可以先在宣告回傳時的變數名稱，同時，若是要回傳大於一個回傳值或是要先宣告回傳的變數名稱，需要使用括弧：

``` go
func functionName1(param type) (returnType, error) {
    var xxx returnType
    var err error
    ...
    return xxx, err
}

func functionName2(param type) (xxx returnType) {
    ...
    // 這邊就可以省略 xxx
    return
}
```

## variables

變數宣告使用關鍵字 `var` ，接著變數名稱與型別，與 import 相同，也可以使用括弧進行多行的宣告。

``` go
var name string

var name, email string

var (
    name, email string
    gender int
)
```

若要同時宣告變數的初始值，可以在行別後面加上 `= value` 或使用 `:=` 同時宣告型別與初始值

``` go
var name string = "raiven"

name := "raiven"

email, gender := "raiven@test.io", 1
```

## 基本型別

golang 的基本型別

``` go
bool

string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte // alias for uint8

rune // alias for int32
     // represents a Unicode code point

float32 float64

complex64 complex128
```

 
較特別的是 rune 是類似於 char 的存在，

### 預設值

若是今天宣告了一個變數，而不給他初始值，則會有各變數的預設值，

``` go
var i int     // 0
var f float64 // 0.0
var b bool    // false
var s string  // ""
```

### 轉換型別

可透過 `Type(value)` 來進行轉換，同時也可以應用在 `:=` 宣告時：

``` go
var i int = 10
f := float64(i) // f is 10.0

nInt := 10          // var nInt int = 10
nInt64 := int64(10) // var nInt64 int64 = 10
```

### constants

常數可以使用關鍵字 `const` 來宣告，但無法使用 `:=` 進行宣告

``` go
const Pi = 3.14

const Pi float64 = 3.14
```

<!-- 
準備被移至 gomodules 時介紹

而如果是在其他資料夾或是其他的 package/module 內，則也需要一併宣告路徑，假設這個包的專案目錄長這樣：

``` 
my-project
|    main.go
|
└────util
    |    util.go
    |    util_test.go
```

此時若要使用 `util.go` 內的 

``` go
import (

    "fmt"
    "math"

)
``` -->
