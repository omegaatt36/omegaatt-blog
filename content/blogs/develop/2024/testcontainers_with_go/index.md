---
title: 用 testcontainers 在本地開發 Go 應用程式
date: 2024-05-19
categories:
 - develope
tags:
 - golang
aliases:
 - "/blogs/develop/2024/testcontainers_with_go.html"
---

## 介紹

使用 [testcontainers](https://testcontainers.com/) 是在本地開發 Golang 應用程式的一個高效方式。這可以讓我們在不需要依賴外部環境的情況下，模擬應用程式在實際生產環境中的運行狀況。

### 安裝 testcontainers

在 Go 專案中，我們可以通過以下指令來導入 testcontainers

```shell
go get github.com/testcontainers/testcontainers-go
```

### 透過 Redis 實踐一個 rate limiter

```go
package user

type Limiter struct {
    client *redis.Client

    limit         int
    limitPeriod   time.Duration // 1 hour for limitPeriod
    counterWindow time.Duration // 1 minute for example, 1/60 of the period
}

func NewLimiter(client *redis.Client, limit int, period, expiry time.Duration) *Limiter {
    return &Limiter{
        client: client,

        limit:         limit,
        limitPeriod:   period,
        counterWindow: expiry,
    }
}

func (r *Limiter) AllowRequest(ctx context.Context, key string, incr int) error {
    now := time.Now()
    timestamp := fmt.Sprint(now.Truncate(r.counterWindow).Unix())

    val, err := r.client.HIncrBy(ctx, key, timestamp, int64(incr)).Result()
    if err != nil {
        return err
    }

    if val >= int64(r.limit) {
        return ErrRateLimitExceeded(0, r.limit, r.limitPeriod, now.Add(r.limitPeriod))
    }

    r.client.Expire(ctx, key, r.limitPeriod)

    result, err := r.client.HGetAll(ctx, key).Result()
    if err != nil {
        return err
    }

    threshold := fmt.Sprint(now.Add(-r.limitPeriod).Unix())

    total := 0
    for k, v := range result {
        if k > threshold {
            i, _ := strconv.Atoi(v)
            total += i
        } else {
            r.client.HDel(ctx, key, k)
        }
    }

    if total >= int(r.limit) {
        return ErrRateLimitExceeded(0, r.limit, r.limitPeriod, now.Add(r.limitPeriod))
    }

    return nil
}

type RateLimitExceeded struct {
    Remaining int
    Limit     int
    Period    time.Duration
    Reset     time.Time
}

func ErrRateLimitExceeded(remaining int, limit int, period time.Duration, reset time.Time) error {
    return RateLimitExceeded{
        Remaining: remaining,
        Limit:     limit,
        Period:    period,
        Reset:     reset,
    }
}

func (e RateLimitExceeded) Error() string {
    return fmt.Sprintf(
        "rate limit of %d per %v has been exceeded and resets at %v",
        e.Limit, e.Period, e.Reset)
}
```

### 創建和啟動 Redis 容器

正式的產品通常會使用 config 來管理 Redis 位置，這個 demo 中直接使用 `localhost:6379` 來展示。

```go
package cache

import "github.com/redis/go-redis/v9"

func NewRedisClient() *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    return client
}
```

### 寫點測試

#### 建立 test suite

```go
package user_test

type LimiterTestSuite struct {
    suite.Suite
}

func TestLimiter(t *testing.T) {
    suite.Run(t, new(LimiterTestSuite))
}
```

#### 透過 testcontainers 啟動 Redis 來進行測試

使用 testcontainers 來創建和啟動 Redis 容器，並將其用於測試限流器。執行測試時會呼叫本地的 docker socket 來啟動 container，並透過 `endpoint, err := container.Endpoint(ctx, "")` 來獲取連線位置。

```go
func (s *LimiterTestSuite) TestLimiterWithTestContainers() {
    ctx := context.Background()

    request := testcontainers.ContainerRequest{
        Image:        "redis:latest",
        ExposedPorts: []string{"6379/tcp"},
        WaitingFor:   wait.ForLog("Ready to accept connections"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: request,
        Started:          true,
    })
    s.NoError(err)

    endpoint, err := container.Endpoint(ctx, "")
    s.NoError(err)

    client := redis.NewClient(&redis.Options{
        Addr: endpoint,
    })

    limiter := user.NewLimiter(client, 10, time.Second*5, time.Second)

    for i := 0; i < 9; i++ {
        s.NoError(limiter.AllowRequest(ctx, "55688", 1), "request %d should be allowed", i+1)
        time.Sleep(time.Millisecond)
    }

    s.Error(limiter.AllowRequest(ctx, "55688", 1), "request should be denied")
}
```

#### 透過真實的連線來測試

我們仍然可以透過 `docker run -p 6379:6379 redis:latest` 來創見一個真實的 redis 實例。

```go
func (s *LimiterTestSuite) TestLimiterWithRealConn() {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    limiter := user.NewLimiter(client, 10, time.Second*5, time.Second)

    ctx := context.Background()

    for i := 0; i < 9; i++ {
        s.NoError(limiter.AllowRequest(ctx, "55688", 1), "request %d should be allowed", i+1)
        time.Sleep(time.Millisecond)
    }

    s.Error(limiter.AllowRequest(ctx, "55688", 1), "request should be denied")
}
```

## 寫在最後

testcontaienrs 的理想是「每一個」測試都會有最乾淨的依賴，假如專案內有 1000 個需要用到 cache/db 的 test function，就會至少有多 1000 個 container 的資源佔用。

若是我們只創建一個 container，並透過 random db name 來給不同的 test case，或許可以解決資源佔用的問題，但就違反了 testcontaienrs 得初衷了。

對於資源有限的 CI/CD 機器更是難以負荷大量的 container 創建/刪除等等，或許可以在專案初期採用，直到專案複雜度提昇，總是會需要在架構乾淨與成長性等等做權衡。
