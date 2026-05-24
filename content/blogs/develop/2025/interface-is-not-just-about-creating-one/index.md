---
title: Interface 不是有開就好：從一個 PR 來看抽象化的重要性
date: 2025-10-04
categories:
  - develop
tags:
  - golang
  - architecture
---

## 前言

最近團隊正在開發一個新產品，其中一個核心功能需要 client 與 server 之間進行即時、雙向的溝通。經過一番技術評估，我們決定採用 WebSocket 來實現這個需求。

身為一個良好習慣的開發團隊，我們在開發初期就導入了依賴注入（Dependency Injection），希望透過界面（Interface）來解耦商業邏輯與具體的實作，這樣不僅能提高程式碼的可測試性，未來在更換底層實作時也能更加輕鬆。

一切聽起來都很美好，直到我在一次 Code Review 中，看到了一段熟悉的程式碼。

## 一個 PR 的故事

在我們的 Domain Layer，也就是處理核心商業邏輯的地方，我看到同事定義了下面這個 interface：

```go
// package/to/domain/service.go

// WebSocketService defines the interface for websocket communication.
type WebSocketService interface {
    // StartAndLinsten starts the service and listens for incoming messages.
    StartAndLinsten(ctx context.Context) error

    // Send sends a message to the client.
    Send(ctx context.Context, message any) error

    // ... other methods
}
```

第一眼看過去，好像沒什麼大問題。有名稱、有方法、也確實是個 interface。然而，當我細看 `WebSocketService` 這個命名時，總覺得哪裡怪怪的。

於是我在 PR 上留下了這樣的 comment：

> 這個界面主要是抽象化 client 與 server 間的互動，不應該侷限於 WebSocket 這個 Protocol。假如我們未來要換成使用 socket.io 或是 gRPC stream，是不是連 domain 層的 interface 也要跟著改動？
>
> domain layer 的界面應該要更專注在他提供的「功能」，而不是「實作細節」。以命名來說，稱作 `AgentCommunicator` 或 `ServerCommunicator` 都好過 `WebSocketService`。

## 抽象化的真正意義

我們之所以要使用 interface，是為了實現「依賴反轉原則」（Dependency Inversion Principle）。簡單來說，就是高層次的模組（我們的商業邏輯）不應該依賴於低層次的模組（WebSocket 的具體實作），兩者都應該依賴於「抽象」。

在這個例子中，`WebSocketService` 這個 interface 就是那個「抽象」。但問題是，這個抽象「洩漏」了底層的實作細節。它的名字直接告訴我們：「我就是一個 WebSocket 的服務」。

這會導致什麼問題呢？

1.  **限制了未來的可能性**：就像我在 comment 中提到的，如果有一天我們發現 WebSocket 不敷使用，想換成 socket.io 或甚至是 MQTT，那我們就必須回來修改這個位在 domain layer 的 interface。這就違反了我們當初想達到的「輕鬆替換實作」的目標。
2.  **模糊了焦點**：在 domain layer，我們真正關心的是「能夠與另一端進行溝通」這件事，而不是「如何溝通」。`WebSocketService` 這個名字會讓後續的維護者（甚至是未來的自己）將思維侷限在 WebSocket 的框架內，而忽略了這個 interface 真正的職責。

一個好的 interface 應該是描述「做了什麼」（What），而不是「怎麼做」（How）。

## 一個更好的 Interface

那麼，一個更恰當的 interface 應該長什麼樣子？我們可以將它重新命名，讓它更能反映其商業意涵。

```go
// package/to/domain/service.go

// Communicator defines the contract for communication between agent and server.
type Communicator interface {
    // Start begins the communication channel and listens for incoming messages.
    Start(ctx context.Context) error

    // Send transmits a message to the other party.
    Send(ctx context.Context, message any) error

    // ... other methods related to communication logic
}
```

我們將它命名為 `Communicator`，並且在註解中清楚說明它的職責是「定義 agent 與 server 之間的溝通契約」。方法名稱也變得更通用，`StartAndLinsten` 變成了 `Start`。

如此一來，我們的商業邏輯就可以安心地依賴 `Communicator` 這個 interface，完全不用知道底層到底是跑 WebSocket、Socket.IO 還是用鴿子在傳訊息。

```go
// package/to/domain/usecase.go

type SomeUseCase struct {
    comm Communicator // Depend on the abstraction
}

func (uc *SomeUseCase) DoSomething() {
    // ... business logic
    uc.comm.Send(context.Background(), "Hello from use case!")
    // ...
}
```

而 `WebSocket` 的具體實作則會放在 infrastructure layer，並且實作 `Communicator` interface。

```go
// package/to/infrastructure/websocket.go

type WebSocketCommunicator struct {
    // ... fields like connection, etc.
}

// Ensure WebSocketCommunicator implements the domain.Communicator interface.
var _ domain.Communicator = (*WebSocketCommunicator)(nil)

func (wsc *WebSocketCommunicator) Start(ctx context.Context) error {
    // ... implementation detail for starting a websocket connection
    return nil
}

func (wsc *WebSocketCommunicator) Send(ctx context.Context, message any) error {
    // ... implementation detail for sending a message via websocket
    return nil
}

// NewWebSocketCommunicator creates a new websocket communicator.
func NewWebSocketCommunicator() *WebSocketCommunicator {
    return &WebSocketCommunicator{}
}
```

## 後話

「開 interface」在現代軟體開發中似乎已經成了一種政治正確。為了可測試性、為了可擴充性，我們到處開 interface。但有時候，我們會為了開而開，卻忽略了 interface 設計的初衷——「抽象化」。

一個設計不良的 interface，即便語法正確，也只是將具體的實作換個名字包起來而已，並沒有真正達到解耦的效果。它像是一把貼著「萬用鑰匙」標籤，卻只能開一扇門的鑰匙。

這次的經驗提醒了我，在定義任何抽象層時，都應該多問自己一句：「我抽象化的到底是什麼？是功能，還是某個特定的技術？」。只有當我們專注於抽象「商業能力」而非「技術實現」時，才能真正享受到架構設計帶來的好處。
