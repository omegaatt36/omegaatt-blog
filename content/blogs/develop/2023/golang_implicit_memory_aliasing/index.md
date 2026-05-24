---
title: Golang 隱式記憶體別名 Implicit Memory Aliasing 與其檢測方法
date: 2023-11-12
categories:
  - develop
tags:
  - golang
---

在使用 Golang 時，我們可能會遇到一種稱為隱式記憶體別名（Implicit Memory Aliasing）的問題。這篇文章將探討這個問題以及如何使用工具和語言特性來解決它。

## 隱式記憶體別名的問題

隱式記憶體別名主要發生在 `range` 語句中。當使用 `range` 對一個切片 slice 或映射 map 進行迭代時，Iterator 在每次迭代中並不是唯一的實例。這可能導致意外的行為，特別是在並發環境或當迭代變量被指針引用時。

### [slice with Implicit Memory Aliasing](https://go.dev/play/p/E_0aBWZcsmZ)

下面示例可能是基本的 golang 面試題，會問你迭代後的 pointers 內的 Name 為何
答案是 Joe Joe Joe

```go
package main

import "fmt"

type Person struct {
	Name   string
	Gender string
}

func main() {
	persons := []Person{
		{Name: "John", Gender: "M"},
		{Name: "Jane", Gender: "F"},
		{Name: "Joe", Gender: "X"},
	}

	pointers := make([]*string, len(persons))
	for index, person := range persons {
		fmt.Printf("%p, %p, %p, %v\n", &person, &person.Name, &person.Gender, &persons[index].Name)
		pointers[index] = &person.Name
	}

	for index := range pointers {
		fmt.Println(pointers[index], *pointers[index])
	}
}
```

### [channel with Implicit Memory Aliasing](https://go.dev/play/p/N21dlpTr_9G)

同樣的情況也會發生在 channel 的迭代。

### 如何解決

1. [with get pointer function(call by value)](https://go.dev/play/p/64y15Y_aF7z)
   使用一個簡易的 to pointer func
   ```
   func ptr[T any](v T) *T {
       return &v
   }
   ```
1. [with closures](https://go.dev/play/p/12mCLF9-_4f)
   迭代時使用 closure
   ```go
   for index, person := range persons {
   	func(v string) {
   		pointers[index] = &v
      }(person.Name)
   }
   ```

### 如何診斷

1. gosec
   [`gosec`](https://github.com/securego/gosec) 是一個流行的 Golang 安全掃描工具，能夠幫助識別代碼中的安全漏洞。其中，G601檢查就是用來發現隱式記憶體別名問題的。通過使用 `gosec`，開發者可以自動檢測到潛在的隱式別名問題，從而提前預防可能的錯誤。

   ```bash
   ❯ gosec ./...
   [gosec] 2023/09/03 22:22:26 Including rules: default
   [gosec] 2023/09/03 22:22:26 Excluding rules: default
   [gosec] 2023/09/03 22:22:26 Import directory: /home/raiven/test
   [gosec] 2023/09/03 22:22:26 Checking package: main
   [gosec] 2023/09/03 22:22:26 Checking file: /home/raiven/test/main.go
   Results:


   [/home/raiven/test/main.go:20] - G601 (CWE-118): Implicit memory aliasing in for loop. (Confidence: MEDIUM, Severity: MEDIUM)
       19:                 fmt.Printf("%p, %p, %p, %v\n", &person, &person.Name, &person.Gender, &persons[index].Name)
   > 20:                 pointers[index] = &person.Name
       21:         }



   [/home/raiven/test/main.go:19] - G601 (CWE-118): Implicit memory aliasing in for loop. (Confidence: MEDIUM, Severity: MEDIUM)
       18:         for index, person := range persons {
   > 19:                 fmt.Printf("%p, %p, %p, %v\n", &person, &person.Name, &person.Gender, &persons[index].Name)
       20:                 pointers[index] = &person.Name



   [/home/raiven/test/main.go:19] - G601 (CWE-118): Implicit memory aliasing in for loop. (Confidence: MEDIUM, Severity: MEDIUM)
       18:         for index, person := range persons {
   > 19:                 fmt.Printf("%p, %p, %p, %v\n", &person, &person.Name, &person.Gender, &persons[index].Name)
       20:                 pointers[index] = &person.Name



   [/home/raiven/test/main.go:19] - G601 (CWE-118): Implicit memory aliasing in for loop. (Confidence: MEDIUM, Severity: MEDIUM)
       18:         for index, person := range persons {
   > 19:                 fmt.Printf("%p, %p, %p, %v\n", &person, &person.Name, &person.Gender, &persons[index].Name)
       20:                 pointers[index] = &person.Name



   Summary:
   Gosec  : dev
   Files  : 1
   Lines  : 26
   Nosec  : 0
   Issues : 4
   ```

1. [Go 1.22 的 loopclosure 特性](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/loopclosure)
   在 Go 1.22 版本中，引入了一項實驗性功能 `GOEXPERIMENT=loopvar`。這個特性旨在解決 `range` 迭代中的隱式記憶體別名問題。當啟用這個實驗性功能時，Go 編譯器會為每次迭代生成一個新的變量實例，從而避免因別名問題導致的錯誤。

## 總結

雖然 Golang 提供了高效的迭代機制，但隱式記憶體別名可能成為一個難題。幸運的是，通過使用如 `gosec` 這樣的工具和利用 Go 1.22 中的 `GOEXPERIMENT=loopvar` 特性，開發者可以有效地識別和解決這些問題，確保代碼的穩定性和安全性。

## 參考

- https://github.com/golang/go/wiki/Range
- https://github.com/golang/go/wiki/CommonMistakes
- https://forum.golangbridge.org/t/when-using-loop-result-is-set-as-only-last-value/11701
- https://www.uber.com/en-TW/blog/data-race-patterns-in-go/
