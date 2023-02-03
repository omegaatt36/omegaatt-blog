---
title: gin 搭配 html/template 包實現動態生成 HTML 文件
date: 2023-01-26
categories:
 - develop
tags:
 - golang
---

## 起因

在網站註冊流程中，若是用信箱驗證，網站會寄送一封驗證信到指定的電子信箱。信中可能含有 verify token 或是直接是寫好的 verify URL。至於實作方面我們可以透過 Go 1.16 推出的 Embedding Files 搭配 html/template，實現動態生成 HTML 信件，用以寄送至指定信箱。

本篇內容的詳細程式碼可以到 [github](https://github.com/omegaatt36/gin-embed-template) 頁面查看。

## 實作
資料夾結構:
```
project/
├─ templates/
│  ├─ template.go
│  ├─ success.tmpl
│  ├─ verify.tmpl
├─ main.go

```

### templates
> project/templates/template.go

- 在包內宣告私有變數，透過 Embedding Files 讀出目錄內的所有檔案。
    ```go
    package template

    //go:embed *
    var f embed.FS
    ```
- 宣告需要被轉換成 HTML 模板的檔案名 `TemplateName`，可以用 [go-enum](https://github.com/abice/go-enum) 來自動生成變數。並將其註冊進陣列 `_TemplateNameNames` 內。
    ```go
    package template

    // ENUM(
    // success.tmpl
    // verify.tmpl
    // )
    type TemplateName string
    
    func (x TemplateName) String() string {
    	return string(x)
    }
    
    var _TemplateNameNames = []string{
    	string(TemplateNameVerifyTmpl),
    	string(TemplateNameSuccessTmpl),
    }
    
    // TemplateNameNames returns a list of possible string values of TemplateName.
    func TemplateNameNames() []string {
    	tmp := make([]string, len(_TemplateNameNames))
    	copy(tmp, _TemplateNameNames)
    	return tmp
    }
    
    const (
    	TemplateNameVerifyTmpl  TemplateName = "verify.tmpl"
    	TemplateNameSuccessTmpl TemplateName = "success.tmpl"
    )
    ```
- 最終將 embed files 與檔案名丟給 `html/template` 包請他們我們轉換成 template。
    ```go
    package template
    
    var templates = template.Must(template.New("").ParseFS(f, TemplateNameNames()...))
    ```

### gin
> project/main.go

在 API 中我們可以透過模板產生 HTML 文件(`string`)，也可以讓 gin 幫我們直接透過模板回傳 HTML 文件。golang 中的 `html/template` 包主要是透過 `map[string]any{"Var": var}` 來映射參數。

> 模板部分可以參考[教學網站](https://gowebexamples.com/templates/)，此文章是基於 Vuepress，寫模板語言會被 parser 認為是模板而建置失敗 orZ。

#### 透過模板產生 HTML 文件
- 我們的成功驗證(`verify.tmpl`)長這樣:
    ```html
    <!DOCTYPE html>
    <html>
    <title>verify</title>
    <head>
    </head>
    <body>
    <h1>Hello {{.Name}}</h1>
    <div>your verify token: {{.VerifyToken}}</div>
    </body>
    </html>
    ```
    所以我們需要填入使用者名稱 `Name` 與驗證碼 `VerifyToken`。
- 在 `project/templates/template.go` 中新增公開方法，用以透過模板與填充物來產生 HTML 文件。
    ```go
    package template
    
    // GenerateHTML returns html with filler.
    func GenerateHTML(n TemplateName, filler any) (string, error) {
    	buf := new(bytes.Buffer)
    	if err := templates.ExecuteTemplate(buf, n.String(), filler); err != nil {
    		return "", err
    	}
    
    	return buf.String(), nil
    }
    ```
- 呼叫 templates 中的 `GenerateHTML` 來產生文件
    ```go
    package main

    /* option 1: from struct
    var req struct {
		Name string
        VerifyToken string
	}{
        Name: "Raiven",
        VerifyToken: "abcde",
    }
    */

    /* option 2: from map
        req := map[string]any{
            "Name": "Raiven",
            "VerifyToken": "abcde",
        }
    */

    mailText, err := templates.GenerateHTML(templates.TemplateNameGeneralTmpl, req)
    if err != nil {
        c.AbortWithError(http.StatusInternalServerError, err)
        return
    }
    ```
    就能拿 `mailText` 透過 mail package 寄送郵件了。

#### 讓 gin 幫我們直接透過模板回傳 HTML 文件

- 在 `project/templates/template.go` 新增公開方法，提供讓 gin 註冊模板。
    ```go
    package template

    // SetHTMLTemplate set templates into gin engine.
    func SetHTMLTemplate(r *gin.Engine) {
    	r.SetHTMLTemplate(templates)
    }
    ```
- 在 gin serve http server 前註冊模板
    ```go
    router := gin.Default()
	templates.SetHTMLTemplate(router)

	if err := srv.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %s\n", err)
	}
    ```
- handler 中可以透過 `gin.Context.HTML()` 來直接透過 gin 呼叫模板轉換
    ```go
	c.HTML(
		http.StatusOK,
		templates.TemplateNameSuccessTmpl.String(),
		map[string]any{"Name": "Raiven"},
	)
    ```