---
title: 一次核心模組的重構經驗
date: 2025-06-14
categories:
  - develop
tags:
  - golang
---

## 前言

在軟體開發的漫漫長路中，我們時常會接手一些充滿「歷史印記」的專案。這些專案的核心模組，往往因為業務的快速迭代與時間的無情沖刷，逐漸演化成難以觸碰的「史前巨獸」。近期，我便有幸（或許該說是不幸地）參與了一次這樣核心模組的重構之旅，其核心是我們產品線廣泛使用的 Golang gRPC 認證攔截器 (Interceptor)。這段經歷充滿挑戰，但也收穫良多，希望能藉此分享一些心得。

## 歷史的塵埃：核心模組的演進悲歌

我接手的這個核心認證模組，在專案初期或許設計簡潔明瞭，但隨著產品線的不斷擴展和需求的堆疊，其複雜度已然失控。追溯其演進的脈絡，彷彿能看到一部小型技術債的形成史。

### 最初的起點：單純的 gRPC Interceptor

可以想見，專案伊始，對於 gRPC 服務的認證需求相對單純。一個通用的攔截器或許就能滿足所有需求，程式碼結構清晰可見：

```go
package main

func SimpleAuthInterceptor(...) {
    log.Println("Performing basic authentication via SimpleAuthInterceptor")
    return handler(ctx, req)
}

func SimpleStreamAuthInterceptor(...) error {
    log.Println("Performing basic stream authentication via SimpleStreamAuthInterceptor")
    return handler(srv, ss)
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer(
        grpc.UnaryInterceptor(SimpleAuthInterceptor),
        grpc.StreamInterceptor(SimpleStreamAuthInterceptor),
    )

    log.Println("gRPC server listening on :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

在這個階段，一切看起來是那麼的美好與純粹。

### 第一次的妥協：Interceptor 內嵌認證邏輯

隨著業務的發展，真正的認證需求浮現。例如，需要開始對請求中的 token 進行驗證。很自然地，這些邏輯被直接添加到了原有的 `Interceptor` 之中：

```go
package interceptor

func extractToken(ctx context.Context) (string, error) {
    return "dummy-token-for-v1", nil
}

func authenticate(ctx context.Context, token string) error {
    if token == "secret-token-v1" || token == "dummy-token-for-v1" {
        return nil
    }
    return errors.New("invalid token")
}

func UnaryAuthInterceptorV1(...) {
    token, err := extractToken(ctx)
    if err != nil {
        return nil, status.Errorf(codes.Unauthenticated, "token extraction failed: %v", err)
    }

    if err := authenticate(ctx, token); err != nil {
        return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
    }
    return handler(ctx, req)
}

func StreamAuthInterceptorV1(...) error {
    token, err := extractToken(ss.Context())
    if err != nil {
        return status.Errorf(codes.Unauthenticated, "token extraction failed: %v", err)
    }

    if err := authenticate(ss.Context(), token); err != nil {
        return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
    }
    return handler(srv, ss)
}
```
此時，`Interceptor` 的職責開始變得不那麼單一，但尚在可控範圍。

### 失控的擴張：Build Tag 與多產品線的糾葛

災難的序幕往往在於「特殊需求」的出現。當我們的核心服務需要支援多條產品線，而某些產品線（例如 `product2`）可能不需要認證，或者有其獨特的認證方式時，Build Tag 方案似乎成了一個「快速」的解決方案：

```go
//go:build product1
// +build product1

package interceptor

func authenticateProduct1(ctx context.Context, token string) error {
    if token == "product1-secret" {
        return nil
    }
    return errors.New("invalid token for product1")
}

func UnaryAuthInterceptor(...) {
    token := "product1-secret"
    if err := authenticateProduct1(ctx, token); err != nil {
        return nil, status.Errorf(codes.Unauthenticated, "product1 auth failed: %v", err)
    }
    return handler(ctx, req)
}
```

```go
//go:build product2
// +build product2

package interceptor

func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    return handler(ctx, req)
}
```

這種方式的直接後果是程式碼庫的碎片化。一個看似通用的 `Interceptor` 功能，其實散落在多個帶有 Build Tag 的檔案中。任何對核心攔截邏輯的通用性修改，都可能需要在多個檔案中同步，極易出錯且難以追蹤。開發者在不同產品線的編譯環境下，看到的可能是完全不同的程式碼行為，這對於理解和除錯造成了極大困擾。

### 最終的混亂：全域狀態與複雜的初始化

隨著更多產品線的接入，以及對既有認證邏輯（legacy auth）的兼容需求，Build Tag 顯然已不堪重負。於是，系統演進到了下一個階段：將認證方式的選擇權上移到 `main` 函數的初始化階段，並透過某種形式（可能是全域變數，或是早期不成熟的依賴注入）將認證策略「滲透」到 `Interceptor` 中。

```go
package main

func main() {
    auth, authURL, newAuth, newAuthURL, err := auth.InitConfig()
    if err != nil {
    	panic(err)
    }

    if err := auth.InitializeGlobalAuth(auth, authURL, newAuth, newAuthURL); err != nil {
        log.Fatalf("Failed to initialize global auth: %v", err)
    }

    s := grpc.NewServer(
        grpc.UnaryInterceptor(interceptor.UnaryAuthInterceptorWithGlobalState),
        grpc.StreamInterceptor(interceptor.StreamAuthInterceptorWithGlobalState),
    )

    log.Println("gRPC server listening on :50051")
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

```go
//go:build product1
// +build product1

package auth

func InitConfig() (...) {
	authURL := os.Getenv("AUTH_URL")
	newAuthURL := os.Getenv("AUTH_NEW_URL")

	if authURL != "" && newAuthURL != "" {
		return errors.New("support only one auth")
	}

	return authURL != "", authURL, newAuthURL != "", newAuthURL
}
```

```go
//go:build product2
// +build product2


package auth

func InitConfig() (...) {
	return false, "", false, "", nil
}
```

```go
package auth
var (
    globalAuth       bool
    globalAuthURL    string
    globalNewAuth    bool
    globalNewAuthURL string
)

func InitializeGlobalAuth(auth bool, authURL string, newAuth bool, newAuthURL string) error {
    globalAuth = auth
    globalAuthURL = authURL
    globalNewAuth = newAuth
    globalNewAuthURL = newAuthURL
    return nil
}

func ValidateTokenUsingGlobalState(ctx context.Context, token string) error {
    log.Printf("Validating token '%s' with global settings: useNewAuth=%t, productID=%s", token, globalUseNewAuth, globalProductID)

		var x, y string
		var err error
    if globalAuth {
  			x, y, err = validateToken(ctx, authURL, token)
    } else {
       x, err = validateNewToken(ctx, newAuthURL, token)
    }

    log.Println(x, y)

    return err
}
```

```go
package interceptor

func UnaryAuthInterceptorWithGlobalState(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    log.Println("UnaryAuthInterceptorWithGlobalState called")
    token := "some-token"

    if err := auth.ValidateTokenUsingGlobalState(ctx, token); err != nil {
        return nil, status.Errorf(codes.Unauthenticated, "global state auth failed: %v", err)
    }

    return handler(ctx, req)
}

func StreamAuthInterceptorWithGlobalState(...) error {
    log.Println("StreamAuthInterceptorWithGlobalState called")
    ctx := ss.Context()
    token := "some-token"

    if err := auth.ValidateTokenUsingGlobalState(ctx, token); err != nil {
        return status.Errorf(codes.Unauthenticated, "global state stream auth failed: %v", err)
    }
    return handler(srv, ss)
}
```

程式碼至此，`Interceptor` 內部充滿了對全域狀態的依賴和複雜的分支判斷。不僅可讀性極差，單元測試也變得異常困難，任何小小的改動都可能牽一髮而動全身，引發不可預期的錯誤。這正是我們決定進行重構的臨界點。

## 重構的曙光：尋找更優雅的解決方案

面對如此混亂的局面，我們深知必須對其進行一次徹底的「手術」。重構的核心目標非常明確：

- 首先，簡化 `Interceptor` 的核心邏輯。它應該只負責攔截請求、提取必要資訊，並將認證的職責委派出去，而不是自己包攬所有產品線的認證細節。
- 其次，大幅提高程式碼的可讀性與可維護性。未來的開發者應該能夠輕鬆理解認證流程，並在需要時安全地擴展新產品線的認證邏R邏輯。
- 再者，消除重複的程式碼。不同產品線之間相似的認證步驟應該被抽象和複用。
- 最後，確保整個架構能夠優雅地支援未來更多產品線的平滑接入。

基於這些目標，我們決定採用**依賴反轉原則**，引入一個 `Authorizer` 介面來抽象化認證行為。

### 核心思路：抽象化與依賴反轉

我們定義了一個清晰的 `Authorizer` 介面：

```go
package auth

type Authorizer interface {
    Authorize(ctx context.Context, token string) error
}

type Product1Authorizer struct {
}

func NewProduct1Authorizer() *Product1Authorizer {
    return &Product1Authorizer{}
}

func (a *Product1Authorizer) Authorize(ctx context.Context, token string) error {
    if token == "valid-token-for-product1" {
        return nil
    }
    return errors.New("invalid token for product 1")
}

type NoOpAuthorizer struct{}

func NewNoOpAuthorizer() *NoOpAuthorizer {
    return &NoOpAuthorizer{}
}

func (a *NoOpAuthorizer) Authorize(ctx context.Context, token string) error {
    return nil
}
```

這個介面非常簡單，只有一個 `Authorize` 方法。任何需要執行認證的邏輯，都可以實作這個介面。

接著，重構後的 `Interceptor` 不再關心具體的認證細節，它只依賴於注入的 `Authorizer` 實例：

```go
package interceptor

type AuthInterceptor struct {
    authorizer auth.Authorizer
}

func NewAuthInterceptor(authorizer auth.Authorizer) *AuthInterceptor {
    if authorizer == nil {
        log.Println("Warning: No authorizer provided to AuthInterceptor, using a default deny-all authorizer.")
        authorizer = &defaultDenyAuthorizer{}
    }
    return &AuthInterceptor{authorizer: authorizer}
}

func (ai *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        token, err := extractTokenFromCtx(ctx)
        if err != nil {
            return nil, status.Errorf(codes.Unauthenticated, "request unauthenticated, token extraction error: %v", err)
        }

        if err := ai.authorizer.Authorize(ctx, token); err != nil {
            return nil, status.Errorf(codes.PermissionDenied, "authorization failed: %v", err)
        }

        return handler(ctx, req)
    }
}

func (ai *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
    return func(...) error {
        ctx := ss.Context()
        token, err := extractTokenFromCtx(ctx)
        if err != nil {
            return status.Errorf(codes.Unauthenticated, "request unauthenticated, token extraction error: %v", err)
        }

        if err := ai.authorizer.Authorize(ctx, token); err != nil {
            return status.Errorf(codes.PermissionDenied, "authorization failed: %v", err)
        }

        return handler(srv, ss)
    }
}

func extractTokenFromCtx(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", errors.New("missing metadata from context")
    }

    authHeaders := md.Get("authorization")
    if len(authHeaders) == 0 {
        if p, ok := credentials.FromContext(ctx); ok {
            if p.AuthType() == "tls" {
            }
        }
        return "", errors.New("missing 'authorization' token in metadata")
    }
    return authHeaders[0], nil
}

type defaultDenyAuthorizer struct{}

func (a *defaultDenyAuthorizer) Authorize(ctx context.Context, token string) error {
    return errors.New("access denied by default policy")
}
```

不同產品線的認證邏輯，現在可以作為 `Authorizer` 介面的不同實作，清晰地隔離開來。例如，`product1` 可能有 `auth_product1.Product1Authorizer`，而 `product2` 如果不需要認證，則可以使用 `auth.NoOpAuthorizer`。

在 `main` 函數中，我們根據配置（例如環境變數、命令列參數或設定檔）來決定實例化哪個 `Authorizer`，並將其注入到 `AuthInterceptor` 中：

```go
package main

func main() {
    productLine := os.Getenv("PRODUCT_LINE")
    var chosenAuthorizer auth.Authorizer

    log.Printf("Configuring server for PRODUCT_LINE: '%s'", productLine)

    switch productLine {
    case "product1":
        chosenAuthorizer = auth.NewProduct1Authorizer()
        log.Println("Using Product1Authorizer.")
    case "product2":
        chosenAuthorizer = auth.NewNoOpAuthorizer()
        log.Println("Using NoOpAuthorizer for Product 2 (no auth).")
    default:
        log.Printf("PRODUCT_LINE '%s' unrecognized or not set. Falling back to default (Product1) authorizer.", productLine)
        chosenAuthorizer = auth.NewProduct1Authorizer()
    }

    authInterceptor := interceptor.NewAuthInterceptor(chosenAuthorizer)

    serverOptions := []grpc.ServerOption{
        grpc.UnaryInterceptor(authInterceptor.Unary()),
        grpc.StreamInterceptor(authInterceptor.Stream()),
    }
    grpcServer := grpc.NewServer(serverOptions...)

    port := ":50051"
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("Failed to listen on port %s: %v", port, err)
    }

    log.Printf("gRPC server starting on port %s with %T", port, chosenAuthorizer)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve gRPC server: %v", err)
    }
}
```

這樣的結構不僅職責分明，也極大地簡化了為新產品線添加認證邏輯的過程——只需要實作新的 `Authorizer` 並在初始化時選擇它即可，無需再改動核心的 `Interceptor` 程式碼。

### 測試的守護：不可或缺的安全網

在進行如此核心且影響廣泛的模組重構時，一套可靠且覆蓋全面的自動化測試是我們的生命線。尤其是在我們沒有充足時間進行大規模人工回歸測試的情況下，CI/CD 流程中整合的端對端（E2E）測試和單元測試，確保了每一次提交都不會破壞現有任何一條產品線的功能和安全。

對於 `Authorizer` 的不同實作，我們可以編寫精確的單元測試來驗證其邏輯的正確性。而對於 `Interceptor` 本身，也可以透過 Mock `Authorizer` 的方式來測試其攔截和委派行為。更重要的是，E2E 測試能夠從使用者角度驗證整個認證流程在不同產品線配置下的實際表現。正是這些測試，給了我們大刀闊斧進行重構的信心。

## 如何更好：重構之後的思考與展望

這次重構顯著改善了原核心認證模組的狀況。程式碼的清晰度、可維護性和可擴展性都得到了大幅提升。每條產品線的認證邏輯被有效地隔離，使得團隊成員能夠更專注於各自負責的部分，降低了互相干擾的風險。

然而，技術的演進永無止境，總有可以做得「更好」的地方：


- 動態配置與載入 `Authorizer`：目前 `Authorizer` 的選擇是在服務啟動時基於環境配置決定的。在某些進階場景下，或許可以探索更動態的方式，例如透過設定檔熱載入，甚至基於請求的某些特徵（如來源 IP、特定標頭）動態選擇不同的 `Authorizer` 實例（當然，這會增加系統複雜度，需要謹慎評估）。

這次重構也再次提醒我，面對歷史遺留的技術債，逃避不是辦法。積極地識別問題，審慎地規劃方案，並在充分測試的保障下果斷執行，才能讓系統重新煥發生機。這不僅是對技術的錘鍊，更是對工程素養的提升。
