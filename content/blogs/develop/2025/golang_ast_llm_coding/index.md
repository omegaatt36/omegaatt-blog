---
title: 如何利用 Golang AST 助攻 LLM 省 token 又高效
date: 2025-06-16
categories:
  - develop
tags:
  - golang
cover:
  image: "images/cover.png"
---

## 前言

近來大型語言模型（LLM）的發展可謂一日千里，特別是在程式碼理解、生成與輔助開發方面，展現出了驚人的潛力。許多開發者開始嘗試將 LLM 融入到日常工作中，期望能提昇開發效率，甚至實現所謂的「vibe coding」——讓 LLM 理解程式碼的整體風格與意圖，並在此基礎上進行協作。

然而，當我們試圖讓 LLM 直接「閱讀」整個大型專案的程式碼庫時，往往會碰到一些現實的挑戰。上下文長度限制、高昂的 token 消耗以及潛在的雜訊干擾，都可能讓 LLM 的表現不盡如人意。這時候，我們就需要更聰明的方法來為 LLM「提煉」程式碼的精華。

在這篇文章中，我想分享一個在 Golang 專案中可能被忽略的利器：抽象語法樹（Abstract Syntax Tree, AST）。透過 Golang AST，我們可以更精準地提取程式碼的結構資訊，為 LLM 提供一份濃縮且高效的上下文，既能節省寶貴的 token，又能幫助 LLM 更好地把握「Code Vibe」。

### LLM 直接消化大型 Code Base 的「痛」

想像一下，你正在開發一個頗具規模的 Golang 後端服務，裡面包含了數十個套件、數百個檔案。現在，你想讓 LLM 幫你新增一個功能，或者重構某個模組。如果直接把所有相關的程式碼一股腦地丟給 LLM，可能會遇到以下這些令人頭痛的問題：

- token 消耗「爆表」：LLM 的使用成本與輸入輸出的 token 數量直接相關。將大量原始碼作為輸入，無疑會產生巨額的 token 費用，對於個人開發者或小型團隊來說，這可能難以承受。
- 「腦容量」不足的上下文限制：即使是目前頂尖的 LLM，其能夠處理的上下文長度也是有限的。面對龐大的程式碼庫，LLM 可能無法一次「看」全所有必要的資訊，導致理解片面或生成結果不佳。
- 資訊過載與雜訊干擾：完整的程式碼中，充斥著各種細節——註解、空行、詳細的錯誤處理邏輯、暫時用不到的私有函式等等。這些資訊對於 LLM 理解程式碼的「vibe」或執行特定高層次任務（例如「模仿現有風格新增一個 API 端點」）來說，有時反而會成為雜訊，影響其判斷。
- 龜速的回應：通常情況下，輸入給 LLM 的資訊越多，它處理並生成回應所需的時間就越長。在追求高效開發的今天，漫長的等待顯然不是我們想要的。

面對這些挑戰，我們不禁要問：有沒有一種方法，可以只給 LLM「剛剛好」的資訊，讓它既能理解我們的意圖，又能高效地完成任務呢？Golang AST 或許就是答案的一部分。

### Golang AST 如何「助攻」

在我們深入探討 AST 如何幫助 LLM 之前，先快速回顧一下什麼是 Golang AST。

#### Golang AST 簡介

AST，即抽象語法樹，是原始碼語法結構的一種樹狀表示。它以樹狀的形式表現程式碼中的文法結構，樹上的每個節點都表示原始碼中的一個結構。例如，一個函式定義、一個型別宣告、一個 `if` 陳述式，或者一個變數賦值，都可以是 AST 中的一個節點。

在 Golang 中，標準庫 `go/parser` 提供了將 Golang 原始碼解析成 AST 的功能，而 `go/ast` 套件則定義了構成 AST 的各種節點型別。透過這些工具，我們可以程式化地分析和操作 Golang 程式碼的結構。

AST 的一個重要特性是它「抽象」掉了原始碼中的許多非本質性細節，比如多餘的空格、括號的具體寫法、甚至是註解（雖然也可以選擇保留）。它更專注於程式的邏輯骨架。

#### 用 AST 精煉 LLM 的「飼料」

了解了 AST 是什麼之後，我們來看看如何利用它來為 LLM 準備更精煉的上下文，幫助 LLM 更好地理解「Code Vibe」。

##### 提取函式簽名與型別定義

當我們希望 LLM 遵循專案現有的設計模式（例如，新增一個符合現有風格的服務介面或資料處理函式）時，提供關鍵的函式簽名 (function/method signatures) 和相關的型別定義 (structs, interfaces) 往往比提供完整的函式實作更為高效。

想像一下，你的專案中有一個 `UserService` 介面和相關的 `User` struct：

```go
package user

// User represents a user in the system.
type User struct {
    ID   string
    Name string
    // ... other fields
}

// UserService defines operations for managing users.
type UserService interface {
    GetUser(id string) (*User, error)
    CreateUser(name string) (*User, error)
    // ... other methods
}
```

如果想讓 LLM 擴充這個 `UserService`，新增一個 `DeleteUser` 方法。與其把 `UserService` 所有實現的完整程式碼都給 LLM，不如只提供 `User` struct 的定義和 `UserService` 介面的定義。LLM 可以從這些簽名中學習到參數型別、回傳型別、錯誤處理模式等「vibe」，然後生成一個風格一致的新方法簽名，甚至初步的實現框架。

使用 `go/parser` 和 `ast.Inspect`，我們可以遍歷 AST，篩選出所有的 `ast.TypeSpec` (型別定義) 和 `ast.FuncDecl` (函式/方法宣告)，並只提取它們的名稱和簽名部分。

##### 描繪依賴輪廓

專案的 `import` 宣告揭示了其外部依賴和內部模組的組織方式。將這些匯入路徑列表提供給 LLM，可以幫助它快速了解專案的技術棧（例如，是用了 `gin` 還是 `echo` 作為 Web 框架？是用 `gorm` 還是 `sqlx` 操作資料庫？）以及模組間的大致依賴關係。

例如，當 LLM 看到大量的 `import "github.com/gin-gonic/gin"`，它就能推斷出接下來生成的 Web API 程式碼應該使用 `gin` 的風格。

##### 聚焦特定範圍的程式碼結構

如果任務是修改某個特定函式或檔案，我們也沒必要提供整個專案的 AST。可以只解析目標檔案，提取該檔案內的頂層宣告（函式、型別、常數、變數），或者更進一步，只提取目標函式的 AST 子樹。

假設我們要修改一個複雜函式 `processOrder` 中的某個錯誤處理邏輯。與其讓 LLM閱讀數百行夾雜著業務邏輯的完整程式碼，不如提供 `processOrder` 函式的簽名，以及其函式體內關鍵的控制流程結構（例如 `if-else` 分支、`for` 迴圈的骨架），並高亮標註出需要修改的部分。這樣 LLM 就能在保持對函式整體結構理解的同時，專注於解決核心問題。

##### 自訂遍歷，按需提取

`ast.Inspect` 函式非常強大，它允許我們遍歷 AST 的每一個節點。透過編寫自訂的訪問者 (visitor) 函式，我們可以精確地提取任何我們感興趣的資訊。

例如，如果我們想讓 LLM 生成一段符合專案日誌記錄風格的程式碼，我們可以先遍歷現有程式碼的 AST，找出所有呼叫日誌函式（比如 `log.Printf` 或自訂的 logger 函式）的地方，提取它們的呼叫模式（例如，日誌級別、訊息格式、記錄的上下文變數等），然後將這些模式作為範例提供給 LLM。

```go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

func main() {
	src := `
package main

import "log"

func main() {
	log.Println("Application started")
	userID := 123
	process(userID)
}

func process(id int) {
	log.Printf("Processing user ID: %d", id)
	if id == 0 {
		log.Fatalf("Fatal error: user ID is zero")
	}
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example.go", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// We want to extract how log functions are called.
	fmt.Println("Log call patterns found:")
	ast.Inspect(f, func(n ast.Node) bool {
		// Check if the node is a function call.
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true // Continue traversal.
		}

		// Check if the function being called is from the "log" package.
		// This is a simplified check; a more robust check would use type information.
		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != "log" {
			return true
		}

		// Found a call to a "log" package function.
		logFunctionName := selExpr.Sel.Name
		var args []string
		for _, arg := range callExpr.Args {
			// For simplicity, convert args to string.
			// In a real scenario, you might want to preserve types or structure.
			if basicLit, ok := arg.(*ast.BasicLit); ok {
				args = append(args, basicLit.Value)
			} else if ident, ok := arg.(*ast.Ident); ok {
				args = append(args, fmt.Sprintf("<variable:%s>", ident.Name))
			} else {
				args = append(args, "<complex_expression>")
			}
		}
		fmt.Printf("- log.%s(%s)\n", logFunctionName, strings.Join(args, ", "))
		return true
	})
}
```
執行上述程式碼，將會輸出：
```sh
Log call patterns found:
- log.Println("Application started")
- log.Printf("Processing user ID: %d", <variable:id>)
- log.Fatalf("Fatal error: user ID is zero")
```
這段精簡的日誌呼叫模式，對 LLM 理解和模仿專案的日誌風格非常有幫助。

### 如何更好：AST 應用的一些思考

雖然 Golang AST 為我們提供了一種強大的程式碼分析手段，但在實際應用中，還有一些方面值得我們進一步思考和探索：

- AST 資訊的粒度：我們到底需要多細緻的 AST 資訊？有時候，僅僅是函式和型別的簽名就足夠了。但在其他情況下，比如要 LLM 理解一個複雜演算法的內部邏輯，可能就需要提供到表達式層級的 AST 結構。這需要根據具體的任務需求來權衡。
- 融合語意資訊：AST 主要反映的是程式碼的語法結構。如果想讓 LLM 更深入地理解程式碼的語意（例如，一個變數的確切型別，一個函式呼叫會解析到哪個具體的定義），我們可能需要結合 `go/types` 套件來進行型別檢查和資訊提取。這無疑會增加處理的複雜度，但也能提供更豐富的上下文。
- AST 的「呈現」方式：提取出來的 AST 資訊需要轉換成一種 LLM 容易「消化」的格式。這可能是簡化的 JSON、XML，或者是某種偽程式碼或自然語言描述。如何設計這種呈現方式，使其既能保留關鍵結構資訊，又不至於過於冗長，是一個值得研究的問題。
- 與 LLM Agent 的協同工作流：一個理想的基於 AST 的 LLM Agent 工作流程可能是這樣的：
  1.  開發者用自然語言描述需求。
  2.  Agent 根據需求，初步判斷可能涉及的程式碼範圍。
  3.  Agent 自動調用 AST 解析工具，從相關程式碼中提取精簡的結構化上下文。
  4.  Agent 將這個精簡上下文連同原始需求一起提交給 LLM。
  5.  LLM 基於這些資訊，生成程式碼片段、修改建議或進一步的提問。
  6.  Agent 將 LLM 的輸出整合回開發環境，或呈現給開發者。
- 認識 AST 的侷限：雖然 AST 功能強大，但它並不能完全取代對原始碼的細緻閱讀。對於那些隱含在程式碼細節中的複雜業務邏 tộc或特定演算法的精妙之處，LLM 可能仍需要更直接的原始碼片段。AST 更適合提供一個結構性的概覽和「vibe」。
- 工具化與自動化：為了方便地在日常開發中利用 AST，我們可以將上述的提取和轉換邏輯封裝成可重用的工具或腳本，甚至整合到 IDE 或 CI/CD 流程中，例如包裝成 MCP Server。

### 總結與展望

在 LLM 席捲軟體開發領域的今天，如何更有效地利用這些強大的模型，是我們每個開發者都需要思考的問題。Golang AST 以其對程式碼結構的精確描述能力，為我們提供了一條與 LLM 高效協作的新路徑。

透過 AST，我們可以從龐雜的程式碼庫中提煉出核心的結構與「vibe」，為 LLM 提供一份「量身打造」的上下文，不僅能顯著降低 token 消耗，提高回應速度，還有可能提昇 LLM 理解和生成程式碼的品質。

當然，AST 的應用仍有廣闊的探索空間。未來，我們或許能看到更智慧化的 AST 分析工具，它們能夠根據 LLM 的即時回饋動態調整提供的上下文深度和廣度，甚至與 LLM 形成更緊密的互動迴圈。

## Ref

- https://yuroyoro.github.io/goast-viewer/
- https://pkg.go.dev/go/ast
- https://www.reddit.com/r/golang/comments/15tfj9y/go_ast_tools
