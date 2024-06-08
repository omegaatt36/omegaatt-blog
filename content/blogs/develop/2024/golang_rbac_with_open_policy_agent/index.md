---
title: 如何利用 Open Policy Agent 配合 Golang 建構彈性的 RBAC 模組
date: 2024-04-04
categories:
 - develop
tags:
 - golang
---

## RBAC 概念簡介

在我們探討如何利用 Open Policy Agent (以下簡稱 OPA) 和 Golang 建立一個彈性的 RBAC 模組之前，先讓我們來了解一下 RBAC的基本概念。

RBAC（Role-Based Access Control，基於角色的訪問控制）是一種廣泛應用的訪問控制策略，在軟體安全性領域尤為重要。其核心思想是將系統訪問權限與用戶的角色（職位、責任或職務）關聯起來，而不是直接與個別用戶關聯。這意味著訪問權限被捆綁到角色上，然後將用戶分配給這些角色。舉個例子，一個「管理員」角色可能有權訪問系統的所有資源，而「員工」角色則只能訪問特定部分的資源。

RBAC 的主要優勢在於其靈活性和簡化的權限管理。當需要變更權限時，只需修改角色的訪問權限，而不需要為每個用戶單獨設定。這不僅使權限管理更為高效，也減少了錯誤配置的可能性，提高了整體的系統安全性。

在實踐中，RBAC 允許創建精細且靈活的策略，以滿足複雜的商業和安全需求。無論是大型企業還是小型團隊，RBAC 都提供了一個可靠的框架，來確保正確的用戶擁有適當的訪問權限，從而保護關鍵資源免受未授權訪問。

可以參考 [Cloudflare 的文章](https://www.cloudflare.com/zh-tw/learning/access-management/role-based-access-control-rbac/)，簡單來說就是**什麼「角色」能夠對什麼「資源」做什麼「操作」**。

## Open Policy Agent (OPA) 介紹

[OPA](https://www.openpolicyagent.org/) 是一個「Strategy as Code」的開源專案，專門設計用於統一地管理和執行跨不同系統的策略。它不僅提供了一個高級的策略語言——Rego，還支援將策略作為代碼與應用程式的其他部分一同存儲、版本控制和部署。OPA 的這種設計使其能夠輕鬆集成到微服務、Kubernetes、CI/CD 管道、API 網關等多種環境中。顯著特點是其策略的編寫方式。Rego 是一種專門為策略和規則定制的查詢語言，它使開發者能夠以聲明式方式描述策略和規則，從而確保這些策略既容易理解又易於維護。這對於建立複雜的 RBAC 系統尤為重要，因為它允許策略的靈活性和可擴展性，同時又保持了清晰和易於審查的結構。

我們可以利用 OPA 提供的 API 來評估和執行這些策略。這意味著開發者可以在 Golang 程式碼中直接嵌入策略判斷的邏輯，從而實現動態、細粒度的訪問控制。這種方法的一個優點是，它支援在 runtime 動態更新策略，或是編譯進 binary，從而提供更大的靈活性和即時性。

## 整合 OPA 與 Golang

### 透過官方案例來了解如何使用

參考了 [OPA 官方的 rbac 章節](https://www.openpolicyagent.org/docs/latest/comparison-to-other-systems/)，並加以修改。使用最簡單的例子：

```plain
admin can read user
bob is admin
-------------------
bob can read user
```

轉化成 RBAC 模型即為：

- role: admin
- resource: user
- action: read

於是我們使用 rego 來撰寫出這個模型，並綁定 bob 到 admin 這個 role 上

```plain
# user-role assignments
user_roles := {
    "bob": ["admin"]
}

# role-permissions assignments
role_permissions := {
    "admin": [{"action": "read",  "resource": "user"}],
}
```

並完成 allow 的判斷「策略」：

```plain
# logic that implements RBAC.
default allow := false
allow if {
    # lookup the list of roles for the user
    roles := user_roles[input.user]
    # for each role in that list
    r := roles[_]
    # lookup the permissions list for role r
    permissions := role_permissions[r]
    # for each permission
    p := permissions[_]
    # check if the permission granted to r matches the user's request
    p == {"action": input.action, "resource": input.resource}
}
```

我們可以在 [playground](https://play.openpolicyagent.org/p/fWfYxyaFU5) 上查看結果

當我們 input 是

```json
{
    "action": "read",
    "resource": "user",
    "user": "bob"
}
```

輸出即為

```json
{
    "allow": true,
    "role_permissions": {
        "admin": [
            {
                "action": "read",
                "resource": "user"
            }
        ]
    },
    "user_roles": {
        "bob": [
            "admin"
        ]
    }
}
```

最終我們能在 output 中的 `allow` 得到目標結果。

### 如何彈性的輸入角色

將 `user_roles` 的部份抽成外部輸入，並稍微進行一些程式碼最佳化

```plain
# rbac.rego
package rbac

import future.keywords.contains
import future.keywords.if
import future.keywords.in

# role-permissions assignments
role_permissions := {
  "admin": [
    {"resource": "user", "action": "edit"},
    {"resource": "user", "action": "read"}
  ]
}

default allow := false

allow if {
  some grant in grants

  input.action == grant.action
  input.resource == grant.resource
}

grants contains grant if {
  some role in input.role
  some grant in role_permissions[role]
}

```

接著我們可以對這個策略寫一些「測試」

```plain
# rbac_test.rego
package rbac_test

import data.rbac.allow
import data.rbac.grants

import future.keywords.in

test_admin_with_incomplete_param {
  not allow with input as {"role": ["admin"]}
}

test_admin {
  not {"action": "A", "resource": "B"} in grants with input as {"role": ["admin"]}
  {"action": "read", "resource": "user"} in grants with input as {"role": ["admin"]}
}
```

再來我們就能透過 [opa 的 cli](https://www.openpolicyagent.org/docs/latest/cli/) 來跑測試 `opa test -v ./*.rego`

```shell
❯ opa test -v ./*.rego
./rbac_test.rego:
data.rbac_test.test_admin_with_incomplete_param: PASS (230.288µs)
data.rbac_test.test_admin: PASS (143.406µs)
--------------------------------------------------------------------------------
PASS: 2/2
```

更詳細的可以參考 [Policy Testing 章節](https://www.openpolicyagent.org/docs/latest/policy-testing/)，這篇文章著重在整合 Golang，詳細 opa 語法就不贅述。

### 透過 Golang 輸入 Role

我們可以在程式端來輸入 `role` 來查詢 `grants`，也可以輸入 `role`, `resource`, `action` 來查詢 `allow`，會用到 `github.com/open-policy-agent/opa/rego` 來實現。

定義一個 rbacService，並完成初始化，使用到 embed 來將 rbac 的檔案給嵌入到查詢裡，若要做成動態的策略，則可以由外部注入。

```go
package rbac

import (
    "context"
    _ "embed"
    "sync"

    "github.com/open-policy-agent/opa/rego"
    "github.com/pkg/errors"
)


//go:embed rbac.rego
var policy []byte

type rbacService struct {
    once        sync.Once
    allowQuery  rego.PreparedEvalQuery
    grantsQuery rego.PreparedEvalQuery
}

// RBACService defines the rbac service interface.
var RBACService rbacService

// Init initializes the rbac service.
func Init(ctx context.Context) error {
    module := rego.Module("policy", string(policy))

    var err1, err2 error
    RBACService.once.Do(func() {
        RBACService.allowQuery, err1 = rego.New(
            rego.Query("data.rbac.allow"),
            module,
        ).PrepareForEval(ctx)

        RBACService.grantsQuery, err2 = rego.New(
            rego.Query("grants = data.rbac.grants"),
            module,
        ).PrepareForEval(ctx)
    })

    if err1 != nil || err2 != nil {
        err := errors.New("failed to prepare rbac policy")
        if err1 != nil {
            err = errors.Wrap(err, err1.Error())
        }
        if err2 != nil {
            err = errors.Wrap(err, err2.Error())
        }
        return err
    }

    return nil
}
```

接著完成查詢 `allow` 與 `grants` 的兩個 function

```go
// IsGrantRequest defines the request for IsGrant.
type IsGrantRequest struct {
    Roles    []Role
    Action   Action
    Resource Resource
}

// IsGranted checks if the request is granted by the rbac policy.
func (s *rbacService) IsGranted(ctx context.Context, req IsGrantRequest) bool {
    results, err := s.allowQuery.Eval(ctx, rego.EvalInput(map[string]any{
        "role":     req.Roles,
        "action":   req.Action,
        "resource": req.Resource,
    }))
    if err != nil {
        log.Println(errors.Wrap(err, "failed to evaluate rbac policy"))
        return false
    } else if len(results) == 0 {
        log.Println("empty rbac policy result, we have wrong query string or policy")
    }

    return results.Allowed()
}

// Grant defines the grant of a role.
type Grant struct {
    Resource
    Action
}

// GetGrants returns the grants of the roles.
func (s *rbacService) GetGrants(ctx context.Context, roles ...Role) ([]Grant, error) {
    results, err := s.grantsQuery.Eval(ctx, rego.EvalInput(map[string]any{
        "role": roles,
    }))
    if err != nil {
        return nil, errors.Wrap(err, "failed to evaluate rbac policy")
    } else if len(results) == 0 {
        return nil, errors.New("empty rbac policy result, we have wrong query string or policy")
    }

    var grants []Grant
    for _, grantI := range results[0].Bindings["grants"].([]any) {
        grant := grantI.(map[string]any)
        grants = append(grants, Grant{
            Resource: Resource(grant["resource"].(string)),
            Action:   Action(grant["action"].(string)),
        })
    }

    return grants, nil
}
```

如此一來我們就完成了獨立於其他服務的 RBAC 模組，僅須輸入某個特定物件（可能是 user，也可能是某個 service）所擁有的 roles，就能查詢他是否擁有操作該資源的權限。

## 還可以更好

未來轉換到其他程式語言，仍可以沿用這套策略。若是將 embed 的部份給解耦，有額外的 role，僅須更改 policy 的 rego，並不需要重新編譯整個 binary。

透過導入 OPA，我們也能學習更現代的策略管理，不僅僅能運用在 RBAC，更可以用在一些 container service account 的權限控管。
