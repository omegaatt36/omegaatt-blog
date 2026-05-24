---
title: Pragmatic Clean Architecture in Go
date: 2026-04-23
categories:
  - develop
tags:
  - golang
  - architecture
cover:
  image: "images/cover.webp"
---

## 前言

在大型系統或需要長期維護的產品中，導入 DDD（Domain-Driven Design）或 Clean Architecture 幾乎是業界的標準答案，分層清晰、職責明確、可測試性高，這些優點毋庸置疑。

但問題是，不是每個專案都是大型系統。

當你面對的是一個內部工具、side project、或是剛起步的新產品，Clean Architecture 的大全套 Use Case、Port、Adapter、Aggregate、Repository Interface 全放 domain，往往會讓你在還沒寫第一行商業邏輯之前，就先在資料夾結構裡迷路了三個小時，或是讓不熟悉的貢獻者花費大量時間在閱讀架構。

這篇文章想討論的是：在不犧牲可測試性與可維護性的前提下，哪些抽象可以捨棄、哪些值得保留，以及我自己踩過的一些坑。

## 捨棄什麼

### 獨立的 Port/Adapter 層

Clean Architecture 中，Use Case 透過明確定義的 input/output port 與外界溝通，搭配 Adapter 負責格式轉換。這在大型系統中確實有其價值，但它帶來的代價是：為了讓每一層都能獨立替換，你需要維護大量的介面與轉換邏輯，而這些轉換邏輯往往只是把 A struct 的欄位複製到 B struct。

在小專案中，這條轉換鏈可以大幅縮短。HTTP handler 本身就可以負責 DTO 的轉換，不需要再多一層 Adapter。

### DDD 的 Aggregate

Aggregate 是 DDD 中保護業務不變式（invariant）的邊界，透過 Aggregate Root 統一管理相關物件的存取。這個概念本身沒有問題，但在小專案中，如果業務規則還不夠複雜到需要嚴格的不變式保護，過早引入 Aggregate 反而會讓簡單的 CRUD 操作變得冗長。

## 保留什麼

捨棄了部分複雜度之後，剩下的四層架構已經足以應付絕大多數的小專案需求。

### `package domain`

這裡只放 **Domain Model** 與 **Value Object**。

值得特別說明 Domain Model 的定義。它不等同於 DDD 裡的 Domain，不是什麼深奧的設計概念，Domain Model 只是在描述「這個應用程式如何跟外部世界互動、邊界在哪」，也就是你的核心資料結構與它們身上的行為。有些「賣課」的人喜歡把 Domain Model 直接等同於 DDD 的 Domain，實際上它是個更基礎、更普遍的概念，可以參考這支[影片](https://www.youtube.com/watch?v=aSPu0WhVPew)的解釋。Domain Model 中不需要也不該出現任何技術細節，例如回傳是否是 JSON 等等，這些 Domain Model 甚至要能讓 PM 與非技術人員也「聽的懂」。

```go
// domain/expense.go

type Category string

func (c Category) Emoji() string { ... }

type Expense struct {
    ID        string
    Name      string
    Price     uint64
    Currency  Currency
    Category  Category
    ShoppedAt time.Time
}

func (e *Expense) TotalInTWD() decimal.Decimal { ... }
```

Domain Model 上可以有方法，例如 `TotalInTWD()` 這種純計算邏輯，不依賴任何外部資源，可以直接在不啟動任何 infrastructure 的情況下測試。

**這裡不放 interface。**

### `package service`

Service 負責最核心的流程控制：協調 Domain Model、透過 interface 呼叫外部依賴、組合業務規則。你可以把它理解成精簡版的 Use Case，我們沒有拿掉這層的概念，只是沒有再額外抽出獨立的 port 定義。

```go
// internal/service/user/service.go

type Service struct {
    repo UserRepo
}

func NewService(repo UserRepo) *Service {
    return &Service{repo: repo}
}

func (s *Service) IsAuthorized(telegramUserID int64) bool {
    user, err := s.repo.GetUser(GetUserRequest{TelegramID: &telegramUserID})
    if err != nil || user == nil {
        return false
    }
    return true
}
```

而 `UserRepo` 這個 interface，定義在**這裡**，而不是 `domain` package：

```go
// internal/service/user/repository.go

type UserRepo interface {
    GetUser(request GetUserRequest) (*domain.User, error)
    GetUsers() ([]domain.User, error)
}
```

這個調整很關鍵，也是我過去做不好的地方。把所有 interface 都丟進 `domain` 是個常見的錯誤，interface 應該定義在**需要它的地方**，也就是消費端（consumer）。`service/user` 需要存取 user 資料，那麼 `UserRepo` 就屬於 `service/user`。這個原則讓每個 package 的依賴關係更清晰，也符合 Go 的 interface 設計哲學：interface 由使用方定義，實作方只需滿足它即可。

### `package handler`

Handler（Presentation Layer）負責兩件事：

1. 將外部呼叫（HTTP request、gRPC call、Telegram message）轉換成內部的 service 呼叫
2. 將 service 回傳的 Domain Model 轉換成外部格式（JSON response、Telegram message）

DTO 的轉換就在這裡發生，這也是我們沒有額外抽出 Adapter 層的原因，handler 本身就兼做了這件事。

```go
// internal/handler/webapp/handler.go

type createExpenseRequest struct {
    Name     string `json:"name"`
    Price    uint64 `json:"price"`
    Category string `json:"category"`
}

func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
    var req createExpenseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := h.expenseSvc.CreateExpense(r.Context(), &domain.Expense{
        Name:     req.Name,
        Price:    req.Price,
        Category: domain.Category(req.Category),
    }); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}
```

### `package repository`

Repository（Infrastructure Layer）是 service 裡定義的 interface 的實作集合。Service 不需要知道資料是存在 Notion、PostgreSQL 還是 in-memory，只要實作了對應的 interface 就能被注入進去。

```go
// internal/repository/notion/expense.go

type ExpenseRepo struct {
    client *notionapi.Client
    dbID   string
}

// CreateExpense 實作了 service/expense.ExpenseRepo 介面
func (r *ExpenseRepo) CreateExpense(ctx context.Context, expense *domain.Expense) error {
    // Notion API 的具體實作
}
```

### `cmd/xxx/main.go`

最後，`main.go` 是整個應用的**組裝點**（composition root）。所有的依賴注入在這裡發生，所有 interface 與實作在這裡配對：

```go
// cmd/webapp/main.go

func main() {
    notionClient := setupNotionClient()

    expenseRepo := notion.NewExpenseRepo(notionClient, cfg.DatabaseID)
    userRepo    := memory.NewUserRepo(cfg.AuthorizedUsers)

    userSvc    := user.NewService(userRepo)
    expenseSvc := expense.NewService(expenseRepo)

    h := webapp.NewHandler(userSvc, expenseSvc)
    http.ListenAndServe(":8080", h.Router())
}
```

整個目錄結構看起來像這樣：

```
.
├── cmd/
│   ├── bot/
│   │   └── main.go
│   └── webapp/
│       └── main.go
├── domain/
│   ├── expense.go
│   ├── receipt.go
│   └── user.go
└── internal/
    ├── handler/
    │   ├── bot/
    │   │   └── handler.go
    │   └── webapp/
    │       └── handler.go
    ├── repository/
    │   ├── notion/
    │   │   └── expense.go
    │   └── memory/
    │       └── user.go
    └── service/
        ├── expense/
        │   ├── repository.go   ← interface 定義在這
        │   └── service.go
        └── user/
            ├── repository.go   ← interface 定義在這
            └── service.go
```

## 當這套架構開始不夠用

### 跨 repository 的 atomic 操作

這套架構第一個會碰壁的地方，通常是「我需要同時寫兩張表，而且要 atomic」。

以轉帳為例：扣款與入帳必須在同一個 transaction 內完成，任何一個失敗就要整體 rollback。但在這套架構裡，`AccountRepo` 和 `LedgerRepo` 是兩個獨立的 interface，service 沒辦法直接控制底層的 transaction 邊界。

這時候就需要引入額外的抽象。常見的做法是 **Unit of Work** 模式：由 UoW 持有 transaction context，並負責提供各個 repository，service 只需要對 UoW 操作：

```go
// internal/service/transfer/uow.go

type UnitOfWork interface {
    AccountRepo() AccountRepo
    LedgerRepo()  LedgerRepo
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

```go
// internal/service/transfer/service.go

func (s *Service) Transfer(ctx context.Context, from, to string, amount uint64) error {
    uow, err := s.uowFactory.Begin(ctx)
    if err != nil {
        return err
    }
    defer uow.Rollback(ctx)

    if err := uow.AccountRepo().Deduct(ctx, from, amount); err != nil {
        return err
    }
    if err := uow.LedgerRepo().Record(ctx, from, to, amount); err != nil {
        return err
    }

    return uow.Commit(ctx)
}
```

repository 層的實作則負責把 `*sql.Tx` 傳進去：

```go
// internal/repository/postgres/uow.go

type unitOfWork struct {
    tx *sql.Tx
}

func (u *unitOfWork) AccountRepo() transfer.AccountRepo {
    return &accountRepo{tx: u.tx}
}

func (u *unitOfWork) Commit(ctx context.Context) error {
    return u.tx.Commit()
}
```

service 完全不知道底層是 PostgreSQL 還是其他資料庫，transaction 的邊界被清楚地封裝在 repository 層內。

這個模式確實比原本的四層多了一個抽象，但它是**被業務需求逼出來的**，而不是一開始就套上去的。這也是為什麼不建議一開始就把 UoW 納入標配，在沒有跨 repository atomic 需求之前，它只是多餘的複雜度。

## 架構是活的

這套精簡架構適合起步，但不代表永遠夠用。

當業務規則開始複雜，同一個 `Expense` 的建立需要同時觸發通知、更新統計、記錄 audit log，你可能就需要引入 event-driven 的設計，或是認真考慮 DDD 的 Aggregate 了。當外部整合的數量多到需要嚴格管控輸入輸出格式，獨立的 Adapter 層也自然有其位置。

捨棄複雜性是為了降低認知負擔，不是為了以後更難重構。如果現在採用這套精簡架構，日後要往 Clean Architecture 靠攏，每一步都是加法，而不是大破大立，這才是「pragmatic」的意義所在。

## 參考資料

- [What is a Domain Model?](https://www.youtube.com/watch?v=aSPu0WhVPew)
- [這篇用到的 repo](https://github.com/omegaatt36/noccounting)
- [Golang Composition over Inheritance](/blogs/develop/2025/golang_composition_over_inheritance/)
- [Interface 不是有開就好](/blogs/develop/2025/interface-is-not-just-about-creating-one/)
