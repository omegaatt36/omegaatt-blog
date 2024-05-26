---
title: golang 中使用 Line LIFF 實作 Single Sign-On
date: 2023-04-22
categories:
 - develop
tags:
 - golang
---

文章中的程式碼放在[https://github.com/omegaatt36/line-verify-id-token](https://github.com/omegaatt36/line-verify-id-token)中。

### Requirements

- [Line Login Channel](https://developers.line.biz/zh-hant/)
    - channel id
    - channel secret
- [Line LIFF APP](https://github.com/line/line-liff-v2-starter)
    - 用 liff.getIDToken() 獲取 [ID Token](https://developers.line.biz/en/docs/line-login/verify-id-token/#get-an-id-token)

在現代網站中，單一登錄 (Single Sign-On, SSO) 已經成為了一個普遍存在的功能，它能夠讓使用者在不同的應用程式和服務之間自動地登錄，而不需要再輸入帳號和密碼。這樣可以方便使用者的使用，並且也能夠增加安全性，減少帳號密碼被盜用的風險。

我們將使用 Golang 語言來實作單一登錄功能，並且使用 Line LIFF 來進行身份驗證。在此之前，我們需要先了解幾個概念。

[Line LIFF (Line Front-end Framework)](https://developers.line.biz/en/docs/liff/overview/) 是一個由 Line 提供的 Web 應用程式框架，開發者可以使用它來建立 Line 的客戶端應用程式。使用 Line LIFF 開發的應用程式可以在 Line 客戶端中被直接執行，而不需要額外安裝或下載。此外，Line LIFF 還提供了一些功能，例如使用者的身份驗證、分享資料等等。

JWT (JSON Web Token) 則是一種開放標準，用於在不同系統之間安全地傳輸訊息。它通常用於認證和授權，因為它可以確保傳輸的訊息是可信的，而且在傳輸過程中不會被竄改。

我們可以使用 Line 的 Verify API，同時也可以選擇後端[自己驗證](https://developers.line.biz/en/docs/line-login/verify-id-token/#write-original-code)，解出 jwt 中的資訊。

我們的目的是拿到每個使用者在 Line 的 UserID，做為身分識別。於是根據 [Line 提供的 JWT 欄位對應](https://developers.line.biz/en/docs/line-login/verify-id-token/#payload)，定義一個結構用來存放驗證後的資訊。

```go
// DecodedIDToken defines decoded payload by id token.
type DecodedIDToken struct {
	Amr       []string
	ChannelID string
	Email     string
	ExpiredAt int64
	IssuedAt  int64
	Issuer    string
	Name      string
	Picture   string
	UserID    string
}
```

接著撰寫一個驗證 ID Token 的 function 用來驗證
```go
// VerifyIDToken verify id token by using HS256 or ES256.
// It checks: signature, issuer, time related fields, channel ID.
func VerifyIDToken(ctx context.Context, idToken string) (*entity.DecodedIDToken, error) {
    // verify

	return &entity.DecodedIDToken{
	}, nil
}
```

最一開始的版本我們使用了標準的 jwt 解碼寫法:
```go
claims := jwt.StandardClaims{}
jwt.ParseWithClaims(idToken, claims, func(token *jwt.Token) (any, error) {
    return channelSecret, nil
})
```

發現解失敗了，原因是用了不同的 Hash Function。此時我們再回去看[官方的文件](https://developers.line.biz/en/docs/line-login/verify-id-token/#header)，ID Token 的 `alg`(signing algorithm)，有可能為 `HS256` 與 `ES256`。

若 `alg` 為 `ES256`，我們需要拿到公鑰 `kid`(Key ID) 才能進行解密。

於是我們用了 `gopkg.in/square/go-jose.v2` 來協助我們獲取公鑰本

```go
func fetchJSONWebKeySet(ctx context.Context) (*jose.JSONWebKeySet, error) {
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet,
		"https://api.line.me/oauth2/v2.1/certs", nil)
	if err != nil {
		return nil, errors.Wrap(err, "can't gen request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "can't fetch line oauth cert keys")
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed request, status: %d", resp.StatusCode)
	}

	var jsonWebKeySet jose.JSONWebKeySet
	if err = json.NewDecoder(resp.Body).Decode(&jsonWebKeySet); err != nil {
		return nil, err
	}

	return &jsonWebKeySet, err
}

func findKey(keySet *jose.JSONWebKeySet, keyID string) (*jose.JSONWebKey, error) {
	for _, key := range keySet.Key(keyID) {
		return &key, nil
	}

	return nil, errors.New("jwk not found")
}
```

接著就可以使用該公鑰本來進行解密了
```go
parser := jwt.Parser{
    ValidMethods: []string{
        jwt.SigningMethodHS256.Alg(),
        jwt.SigningMethodES256.Alg(),
    },
}
claims := jwt.MapClaims{}
token, err := parser.ParseWithClaims(idToken, claims,
    func(token *jwt.Token) (any, error) {
        alg := token.Method.Alg()
        switch alg {
        case jwt.SigningMethodES256.Alg():
            keySet, err := fetchJSONWebKeySet(ctx)
            if err != nil {
                return nil, err
            }

            kid, ok := token.Header["kid"]
            if !ok {
                return nil, errors.New("can't find kid in header")
            }

            kidStr, ok := kid.(string)
            if !ok {
                return nil, fmt.Errorf("kid type assertion failed (%T)", kid)
            }

            jwkKey, err := findKey(keySet, kidStr)
            if err != nil {
                return nil, err
            }

            return jwkKey.Key, nil
        case jwt.SigningMethodHS256.Alg():
            return []byte(*channelSecret), nil
        default:
            return nil, fmt.Errorf("illegal id token alg: %v", alg)
        }
    })
if err != nil {
    return nil, err
}
```

最後再對解出來的 claims 進行驗證

```go
// validate token.
if !token.Valid {
    return nil, errors.New("id token is invalid")
}

if !claims.VerifyIssuer("https://access.line.me", true) {
    return nil, errors.New("not be signed by line")
}

if !claims.VerifyAudience(*channelID, true) {
    return nil, fmt.Errorf("audience is not match channel ID(%v)", *channelID)
}

// convert to user friendly struct.
bs, err := json.Marshal(token.Claims)
if err != nil {
    return nil, errors.Wrapf(err, "json.Marshal() failed, %s", token.Claims)
}

var id struct {
    Amr       []string `json:"amr"`
    ChannelID string   `json:"aud"`
    Email     string   `json:"email"`
    ExpiredAt int64    `json:"exp"`
    IssuedAt  int64    `json:"iat"`
    Issuer    string   `json:"iss"`
    Name      string   `json:"name"`
    Picture   string   `json:"picture"`
    UserID    string   `json:"sub"`
}
if err := json.Unmarshal(bs, &id); err != nil {
    return nil, errors.Wrapf(err, "json.Unmarshal() to DecodedIDToken failed, %s", string(bs))
}
```

完成驗證，回傳驗證後的 ID Token
```go
return &entity.DecodedIDToken{
    Amr:       id.Amr,
    ChannelID: id.ChannelID,
    Email:     id.Email,
    ExpiredAt: id.ExpiredAt,
    IssuedAt:  id.IssuedAt,
    Issuer:    id.Issuer,
    Name:      id.Name,
    Picture:   id.Picture,
    UserID:    id.UserID,
}, nil
```

後續只要將 UserID 存進站內的資料庫，並與用戶關聯，下次登入就可以區分為:
- `POST /api/user/v1/login` 使用帳號密碼登入，並回傳站內的 JWT。
- `POST /api/user/v1/login/line` 使用 LINE ID Token 登入，並回傳站內的 JWT。

總結一下，這篇主要是介紹了如何使用 Golang 實作 Line 的 Single Sign-On，並且介紹了如何透過驗證 ID Token 來確認使用者是否已經登入 Line 帳號。
