---
title: 基於 Golang 的 Grafana Dashboard 與 JWT 認證的前後端實作
date: 2024-06-10
categories:
 - develop
tags:
 - golang
---

在工作上有一個需求是需要做一些 [OLAP](https://aws.amazon.com/tw/what-is/olap/)，原訂計畫是使用 Google Looker(ver. Google Cloud Core)，礙於量小不符合經濟效益，決定用 Grafana 這個較熟悉的開源套件來幫助我們做視覺化的處理。

這篇文章的範例可以在 [omegaatt36/grafana-embed-example](https://github.com/omegaatt36/grafana-embed-example) 中找到所有的 source code

## 流程

```mermaid
sequenceDiagram
    autonumber
    participant U as User
    participant B as Browser
    participant I as iframe
    participant S as Server
    participant G as Grafana

    rect rgb(245, 245, 245)
    U->>B: Open Web Page
    B->>S: Request JWT Token
    S->>B: Return JWT Token
    B->>I: Load iframe
    I->>G: Request Dashboard with JWT Token
    G->>I: Return Dashboard
    I->>B: Display Dashboard in iframe
    B->>U: Show Dashboard
    end
```

## Grafana

### 配置

主要是針對 `grafana.ini` 做修改

#### [security]

- `allow_embedding = true`：允許將 Grafana 儀表板嵌入到其他網頁中，這對於我們需要將儀表板嵌入到自定義前端頁面是必要的。
- `cookie_samesite = none`：允許跨站點請求攜帶 cookie，這在嵌入的情況下特別有用，因為瀏覽器會阻止跨站點的 cookie 請求。

#### [auth]

- `whitelisted_domains = localhost`：允許來自這些域名的請求繞過某些安全檢查。在開發和測試環境中，這有助於簡化流程。

#### [auth.jwt]

- `enabled = true`：啟用 JWT 認證。
- `header_name = X-JWT-Assertion`：指定用於攜帶 JWT 的 HTTP 標頭名稱。
- `enable_login_token = true：允許使用 JWT 進行登入。
- `username_claim = user 和 email_claim = email：指定 JWT 中對應使用者名稱和電子郵件的字段。
- `key_file = /etc/grafana/rsa_pub.pem`：指定用於驗證 JWT 簽名的公鑰文件路徑。
- `cache_ttl = 60m`：指定 JWT 快取的有效期。
- `expect_claims = {}`：設定期望的 JWT 權限。
- `role_attribute_path = role`：指定 JWT 中對應角色的字段。
- `role_attribute_strict = false`：是否嚴格匹配角色屬性。
- `auto_sign_up = true`：允許自動註冊新用戶。
- `url_login = true`：允許通過 URL 進行登入。
- `allow_assign_grafana_admin = false`：不允許自動分配 Grafana 管理員角色。
- `skip_org_role_sync = false`：不跳過組織角色同步。

#### [auth.anonymous]

- `enabled = false`：禁用匿名訪問，確保所有訪問都需要經過認證。這點對於後續驗證十分重要，有很多其他部落格的這個區域都是設成 `true`，根本沒有經過 auth，相對的就是沒有安全性可言。

{{< details title="完整 config：" >}}

```ini
[security]
allow_embedding = true
cookie_samesite = none

[auth]
whitelisted_domains = localhost

#################################### Auth JWT ##########################
[auth.jwt]
enabled = true
header_name = X-JWT-Assertion
enable_login_token = true
username_claim = user
email_claim = email
key_file = /etc/grafana/rsa_pub.pem
cache_ttl = 60m
expect_claims = {}
; key_id =
role_attribute_path = role
role_attribute_strict = false
auto_sign_up = true
url_login = true
allow_assign_grafana_admin = false
skip_org_role_sync = false

[auth.anonymous]
enabled = false

[log.console]
level = info
```

{{< /details >}}

### JWT

JWT (JSON Web Token) 是一種開放標準，用於在不同系統之間安全地傳輸訊息。在這個實作中，我們使用 JWT 來認證和授權使用者訪問嵌入的 Grafana 儀表板。

我們可以產生一組 Public Key 與 Private Key 的 Pair，簽署用的私鑰放在後端，驗證用的公鑰放在 grafana 伺服器內。

後端會生成 JWT，前端（由於是 iframe）透過 URL query string 將其傳遞給 Grafana，Grafana 會使用公鑰驗證 JWT 的有效性。

### 資料來源

資料來源定義了 Grafana 如何連接到數據庫或其他數據來源，以便從中提取數據進行可視化。這些設置可以在 Grafana 的 UI 中配置，並且需要確保數據源的連接憑證和訪問權限正確配置。在這個 Example 中我們會連接到 MySQL 資料庫中，並且存取 `users` & `orders` 等等資料表。

### User

使用者配置包括創建和管理 Grafana 使用者帳號，設定其訪問權限和角色。JWT 認證允許自動創建和分配使用者角色，這使得管理大量使用者更加便捷。我們會在 Grafana 中新增一個 Role 為 Viewer 的使用者，並且後端會使用這組帳號來簽署 JWT。

### Dashboard

儀表板是 Grafana 的核心組件，用於可視化數據。儀表板可以包括多個面板，每個面板展示特定數據源的數據。可以通過 JSON 定義儀表板的結構和內容。我們是直接使用剛才建立的 MySQL Datasource 作為資料源。

### Dashboard Varable

儀表板變數允許使用者動態更改儀表板的內容，例如選擇不同的數據範圍或過濾條件。這些變數可以在儀表板設計時定義，並通過 URL 參數或 UI 控件進行設置。

### Dashboard Permission

儀表板許可權控制哪些使用者或角色可以訪問或編輯儀表板。可以在 Grafana 的 UI 中配置許可權，以確保只有授權的使用者可以查看或修改敏感數據。我們將設定僅有剛才建立的 Viewer 能夠查看這個 Dashboard。

### Dashboard Embed link

嵌入鏈接允許將 Grafana 儀表板嵌入到其他網頁或應用中。配置嵌入鏈接時，需要確保允許嵌入並使用合適的 JWT 進行認證，以確保安全性和數據隱私。

## 前端

由於我的前端技術非常薄弱，這段程式碼是請 ChatGPT 產生的，實現了一個簡單的前端介面，允許用戶輸入起始日期、結束日期和組織 ID，主要包括以下幾個部分：

### HTML 部分

{{< details title="完整 code:" >}}

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Embed Grafana Dashboard</title>
</head>
<body>
    <h1>Grafana Dashboard</h1>
    <label for="start">Start Date:</label>
    <input type="date" id="start" name="start" value="2024-05-01">
    <label for="end">End Date:</label>
    <input type="date" id="end" name="end" value="2024-05-31">
    <label for="organization_id">Organization ID:</label>
    <input type="text" id="organization_id" name="organization_id" value="1">
    <button onclick="loadGrafanaDashboard()">Load Dashboard</button>
    <iframe id="grafanaFrame" src="" width="100%" height="450px"></iframe>

    <script>
        async function fetchToken() {
            try {
                const response = await fetch(`/token`);
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                const token = await response.text();
                return token;
            } catch (error) {
                document.getElementById('error').style.display = 'block';
                console.error('Error fetching token:', error);
            }
        }

        async function loadGrafanaDashboard() {
            const token = await fetchToken();
            if (token) {
                const startDate = document.getElementById('start').value;
                const endDate = document.getElementById('end').value;
                const organizationId = document.getElementById('organization_id').value;

                if (!startDate || !endDate || !organizationId) {
                    document.getElementById('error').innerText = 'Please select start date, end date, and enter organization ID.';
                    document.getElementById('error').style.display = 'block';
                    return;
                }

                const from = new Date(`${startDate}T00:00:00`).getTime();
                const to = new Date(`${endDate}T23:59:59`).getTime();

                const grafana1BaseURL = `{{ .DashboardURL }}`;
                const grafana1URL = `${grafana1BaseURL}?auth_token=${token}&kiosk&from=${from}&to=${to}&var-org_id=${organizationId}`;
                document.getElementById('grafanaFrame').src = grafana1URL;
            }
        }
    </script>
</body>
</html>
```

{{< /details >}}

- `input` 用於選擇起始和結束日期，以及輸入組織 ID。
- `button` 用於觸發加載儀表板的操作。
- `iframe` 用於嵌入並顯示 Grafana 儀表板。

### JavaScript 部分

- `fetchToken` 函數從後端獲取 JWT token。
- `loadGrafanaDashboard` 函數使用獲取的 token 和用戶輸入的參數生成 Grafana 儀表板的 URL，並將其設置到 iframe 中以加載儀表板。需要特別提到的是 `const grafana1BaseURL = {{ .DashboardURL }};` 為模板語言，目的是在 example 中透過 golang 的 `html/template` 來動態產生 URL。

這樣的設計確保了前端介面簡潔且易於使用，同時利用 JWT 認證確保了安全性。

## 後端

### HTTP server

啟動一個 http server，主要是 handle `/` 與 `/token`

```go
func main() {
    flag.Parse()
    if grafanaDashboardURL == nil || *grafanaDashboardURL == "" {
        log.Fatal("grafana-dashboard-url is required")
    }

    router := http.NewServeMux()
    router.HandleFunc("GET /token", generateJWT)
    router.HandleFunc("GET /", serveIndex)

    fmt.Println("Server is running on port 8000")
    if err := http.ListenAndServe(":8000", router); err != nil {
        fmt.Println(err)
    }

    fmt.Println("exiting...")
}
```

### Serve index

使用 embed 嵌入 index.html，並且動態地將 DashboardURL 交給 `html/template` 渲染

```go
//go:embed index.html
var indexHTML string

var grafanaDashboardURL *string = flag.String("grafana-dashboard-url", "", "grafana dashboard url")

func getIndexHtml() (*template.Template, error) {
    index, err := template.New("index").Parse(indexHTML)
    if err != nil {
        return nil, fmt.Errorf("failed to parse index.html: %w", err)
    }

    return index, nil
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
    index, err := getIndexHtml()
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err := index.Execute(w, map[string]any{
        "DashboardURL": *grafanaDashboardURL,
    }); err != nil {
        log.Println(err)
    }
}

```

### 簽署 JWT

一樣透過 embed 的方式讀取私鑰，並且在每個請求都能夠使用我們設計好的 claim 來將 user 與 email 給簽進 JWT 內。

```go
//go:embed rsa.pem
var privateKey string

func getPrivateKey() (*rsa.PrivateKey, error) {
    block, _ := pem.Decode([]byte(privateKey))
    if block == nil || block.Type != "RSA PRIVATE KEY" {
        panic("Failed to decode PEM block containing private key")
    }

    pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        panic(err)
    }

    return pk, nil
}

func generateJWT(w http.ResponseWriter, r *http.Request) {
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
        "user":  "demo",
        "email": "user@demo.com",
        "iat":   time.Now().Unix(),
        "exp":   time.Now().Add(time.Hour * 1).Unix(),
    })

    pk, err := getPrivateKey()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    tokenString, err := token.SignedString(pk)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if _, err := w.Write([]byte(tokenString)); err != nil {
        log.Println(err)
    }
}
```

## 驗證

這邊需要強調的是**務必使用無痕模式來驗證**，否則 iframe 會使用已經登入的 session 來開啟，故可能無法驗證嵌入的效果。

照著 [How to use](https://github.com/omegaatt36/grafana-embed-example?tab=readme-ov-file#how-to-use) 段落一步一步往下做，就能夠在 iframe 內看到透過 JWT 進行登入的 Grafana Dashboard。
