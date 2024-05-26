---
title: 跟風寫了個使用 OpenAI API 的 Telegram Chat Bot
date: 2023-03-17
categories:
 - develop
tags:
 - golang
 - telegram
---

# 前言

放在最前面 [Github repo](https://github.com/omegaatt36/chatelegram)

原本打算使用別人寫好的 bot，免費的版本均是透過儲存 cookies 的方式，直接透過 ChatGPT 網頁建立新的 「Chat」。好景不常，OpenAI 馬上把登入頁面到聊天頁面中間安插了類似 reCaptcha，來驗證是否為真人。看到有好心人提供了 go 版本的 OpenAi API，就打算自己寫一寫，順便當作 [《Clean Architecture》](https://www.tenlong.com.tw/products/9789864342945) 的讀後作業。

# 依賴注入

~~架構圖待補~~

我需要對 OpenAI 詢問，並用 stream(逐字)的方式回給我，也需要對 Telegram 逐字回答。

在[這個 8931bf 版本](https://github.com/omegaatt36/chatelegram/tree/8931bfbf9e0e00891c07aaaf0c12aa730e34dd39)中，很明確定義對於 OpenAI 與 Telegram 的 usecase

- `chatgpt_usecase.go`
    ```go
    // ChatGPTUseCase defines ChatGPT send question use case.
    type ChatGPTUseCase interface {
    	Stream(ctx context.Context, question string) (<-chan string, <-chan error)
    }
    ```
- `telegram_usecase.go`
    ```go
    // TelegramUseCase defines telegram send message use case.
    type TelegramUseCase interface {
    	SendAsLiveOutput(chatID int64, feed <-chan string) error
    }
    ```

在 `*_repository.go` 進行實作。並在 `main.go` 中依賴反轉，將實作注入近 bot service。
```go
func Main(ctx context.Context) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  config.telegramBotToken})
    client := gpt3.NewClient(
        config.apiKey)
    
    service := chatgpttelegram.NewService(
        bot,
        telegram.NewTelegramBot(bot),
        chatgpt.NewChatGPTClient(client),
    )
    
    service.Start(ctx)
}
```

最後在 telegram bot 進行 `OnText` 指令時呼叫 `ChatGPTUseCase.Stream()` 與 `TelegramUseCase.SendAsLiveOutput()`
```go
// type Service struct {
// 	telegram telegram.TelegramUseCase
// 	gpt      chatgpt.ChatGPTUseCase
// }

func (s Service) processChatGPTQuestion(chatID int64, question string) error {
	messageCh, errChatGPTCh := s.gpt.Stream(s.ctx, question)

	errCh := make(chan error, 1)
	defer close(errCh)

	done := make(chan struct{}, 1)
	go func() {
		if err := s.telegram.SendAsLiveOutput(chatID, messageCh); err != nil {
			errCh <- err
		}
		done <- struct{}{}
	}()
}
```

# 更換 OpenAI API Package

在[這個 7d81f9 版本](https://github.com/omegaatt36/chatelegram/tree/7d81f921345f765097298c8aa54e8ded6dcf995f)中，我發現 `github.com/sashabaranov/go-openai` 比我原本用的 `github.com/PullRequestInc/go-gpt3` 具有更全面的支援，於是打算更換 package。

僅需要在 repository 中新增新的實作
```go
// OLD "github.com/PullRequestInc/go-gpt3"
type GPT3Client struct {
	client gpt3.Client
}

// NEW "github.com/sashabaranov/go-openai"
type OpenAIClient struct {
	client *openai.Client
}
```

並在 `main.go` 中更換我想要注入的依賴
```go
func Main(ctx context.Context) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  config.telegramBotToken})
	client := openai.NewClientWithConfig(
        openai.DefaultConfig(config.apiKey))
    
    service := chatgpttelegram.NewService(
        bot,
        telegram.NewTelegramBot(bot),
        chatgpt.NewChatGPTClient(client),
    )
    
    service.Start(ctx)
}
```

It Works.

## 總結

依賴注入用的好，重構程式沒煩惱。
