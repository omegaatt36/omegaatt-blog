---
title: Helm Smart Resource：讓你的 Chart 學會與既有資源和平共處
date: 2025-11-22
categories:
  - develop
  - devops
tags:
  - helm
  - kubernetes
  - infrastructure
---

## 前言

在 Kubernetes 的世界裡，Helm 無疑是管理應用程式部署的霸主。它標準化了資源的定義，讓我們可以用宣告式的方式管理整套系統。然而，在真實世界的維運場景中，事情往往沒那麼單純。

我們常遇到一種尷尬的情況：某些資源（例如 Database 的 Secret、外部系統的 ConfigMap）可能在 Helm Chart 安裝之前就已經由維運人員手動建立，或是由另一個流程（如 Terraform）預先準備好了。

這時候，如果直接執行 `helm install`，往往會收到 "resource already exists" 的錯誤；如果使用 `helm upgrade --install`，又擔心 Helm 會覆蓋掉這些既有設定。

這篇文章將分享一種「Smart Resource」的設計模式，透過 Helm 的 `lookup` 函數與樣板邏輯，讓你的 Chart 能夠聰明地判斷：「這東西是我管的嗎？如果是，我才動它；如果不是，我就尊重現狀。」

## 核心難題：Ownership

在 Kubernetes 中，資源的「所有權」觀念至關重要。Helm 預設認為它 release 中的所有資源都應該由它全權管理。但當我們需要與外部資源協作時，我們需要更細緻的控制。

我們的目標很明確：
1. 若資源不存在：建立它，並標記為 Helm 管理。
2. 若資源已存在且由 Helm 管理：更新它（Patch/Merge）。
3. 若資源已存在但由外部管理：保持原狀，不進行覆蓋或刪除。

為了達成這個目標，我們需要一個輔助函數來判斷資源的歸屬權。

## 實作細節

### 1. 定義所有權檢查

首先，我們在 `_helpers.tpl` 中定義一個檢查函數。Helm 會在它建立的資源上打上特定的 Annotations（`meta.helm.sh/release-name` 和 `meta.helm.sh/release-namespace`）。我們可以利用這一點來判斷資源是否屬於當前的 Release。

```yaml
{{/*
Check if a resource is owned by this Helm release
Returns "true" or "false" as string for stable piping
*/}}
{{- define "visionone-filesecurity.isOwnedByRelease" -}}
{{- $resource := .resource -}}
{{- $releaseName := .releaseName -}}
{{- $releaseNamespace := .releaseNamespace -}}
{{- $owned := and $resource
    (hasKey $resource.metadata "annotations")
    (eq (get $resource.metadata.annotations "meta.helm.sh/release-name") $releaseName)
    (eq (get $resource.metadata.annotations "meta.helm.sh/release-namespace") $releaseNamespace)
-}}
{{ printf "%t" $owned }}
{{- end -}}
```

這段程式碼邏輯很簡單：只有當資源存在，且其 Annotations 中的 Release Name 與 Namespace 都與當前 Release 相符時，才視為「Owned」。

### 2. 實作 Smart ConfigMap

接下來，我們利用 `lookup` 函數來實作 Smart ConfigMap。`lookup` 允許我們在 Template 渲染期間查詢 Cluster 內的實際狀態。

```yaml
{{/*
Render a ConfigMap with smart lookup and merge strategy
*/}}
{{- define "visionone-filesecurity.smartConfigMap" -}}
{{- $name := .name -}}
{{- $ns := .namespace -}}
{{- $labels := .labels | default (dict) -}}
{{- $data := .data | default (dict) -}}
{{- $context := .context -}}
{{- $existing := lookup "v1" "ConfigMap" $ns $name -}}

{{- /* Check if resource is owned by this release */ -}}
{{- $owned := include "visionone-filesecurity.isOwnedByRelease" (dict "resource" $existing "releaseName" $context.Release.Name "releaseNamespace" $context.Release.Namespace) | eq "true" -}}

{{- if and $existing (not $owned) -}}
  {{- /* Resource exists but not owned by this release - skip rendering */ -}}
{{- else -}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ $name }}
  namespace: {{ $ns }}
  labels:
{{- toYaml (merge $labels (dict "app.kubernetes.io/managed-by" "Helm")) | nindent 4 }}
  annotations:
    meta.helm.sh/release-name: {{ $context.Release.Name | quote }}
    meta.helm.sh/release-namespace: {{ $context.Release.Namespace | quote }}
data:
  {{- /* Merge existing data with desired data if owned by this release */ -}}
  {{- $liveData := (ternary (get $existing "data") (dict) $owned) | default (dict) -}}
  {{- $mergedData := mergeOverwrite (deepCopy $liveData) $data -}}
  {{- range $k, $v := $mergedData }}
  {{ $k }}: {{ $v | quote }}
  {{- end }}
{{- end -}}
{{- end -}}
```

這裡的關鍵邏輯在於 `if and $existing (not $owned)`。如果資源存在但不是我管的，Helm 就會完全跳過這段 YAML 的渲染。這意味著 Helm 不會嘗試去 Apply 這個資源，從而避開了 "already exists" 的錯誤，也保護了外部資源不被修改。

反之，如果資源是我管的（或是全新的），我們就會渲染出完整的 ConfigMap 定義，並且利用 `mergeOverwrite` 來保留可能被其他 Controller 動態修改過的 data 欄位（如果你的架構允許這種操作的話），或是確保我們定義的值被正確寫入。

### 3. 處理更棘手的 Secret

Secret 的處理邏輯與 ConfigMap 類似，但因為涉及 Base64 編碼，處理起來稍微繁瑣一些。我們需要同時支援明文輸入 (`stringData`) 和既有的編碼資料 (`data`)。

```yaml
{{/*
Render a Secret with smart lookup and merge strategy
*/}}
{{- define "visionone-filesecurity.smartSecret" -}}
{{- $name := .name -}}
{{- $ns := .namespace -}}
{{- $context := .context -}}
{{- /* ... 略過部分變數宣告 ... */ -}}

{{- $existing := lookup "v1" "Secret" $ns $name -}}
{{- $owned := include "visionone-filesecurity.isOwnedByRelease" (dict "resource" $existing "releaseName" $context.Release.Name "releaseNamespace" $context.Release.Namespace) | eq "true" -}}

{{- if and $existing (not $owned) -}}
  {{- /* Skip rendering */ -}}
{{- else }}
---
apiVersion: v1
kind: Secret
# ... metadata ...
type: {{ $type }}

{{- /* Live data (base64 encoded) */ -}}
{{- $liveData := (ternary (get $existing "data") (dict) $owned) | default (dict) -}}

{{- /* Convert stringData to base64 and merge with data */ -}}
{{- $encodedStringData := dict -}}
{{- range $k, $v := $inStringData }}
  {{- $_ := set $encodedStringData $k (b64enc (toString $v)) -}}
{{- end -}}
{{- $allDesiredData := merge $encodedStringData $inData -}}

{{- $mergedData := mergeOverwrite (deepCopy $liveData) $allDesiredData -}}
{{- if $mergedData }}
data:
  {{- range $k, $v := $mergedData }}
  {{ $k }}: {{ $v | quote }}
  {{- end }}
{{- end }}
{{- end }}
{{- end }}
```

這段程式碼展示了如何在樣板層級處理 Secret 的編碼轉換，確保無論使用者傳入的是明文還是 Base64，最終都能與 Cluster 上現有的資料正確合併。

### 4. 實際應用

定義好這些 template 後，我們在實際的資源檔（例如 `configmap.yaml`）中就可以這樣使用：

```yaml
{{/* Validate configuration before creating ConfigMaps */}}
{{- include "visionone-filesecurity.validateScannerConfig" . -}}

{{/* ConfigMap for Scanner configuration using smart lookup strategy */}}
{{- $baseLabels := include "visionone-filesecurity.managementService.labels" . | fromYaml -}}

{{- include "visionone-filesecurity.smartConfigMap" (dict
  "name" .Values.scanner.configMapName
  "namespace" .Release.Namespace
  "labels" $baseLabels
  "data" (dict "log_level" .Values.scanner.logLevel)
  "context" .
) -}}
```

透過這種方式，我們的 YAML 檔案變得很乾淨，複雜的判斷邏輯都被封裝在 `_helpers.tpl` 中。

## 注意事項與限制

雖然 Smart Resource 模式解決了很多問題，但使用時仍需注意以下幾點：

1. Dry Run 的限制：Helm 的 `lookup` 函數在 `helm install --dry-run` 或 `helm template` 時是不會執行實際查詢的（因為沒有連接到 Cluster）。在這種情況下，`$existing` 會是空的，因此樣板會預設渲染出「建立新資源」的 YAML。這在 CI/CD pipeline 做 linting 時需要特別留意。
2. 權限問題：執行 Helm 的 User 或 ServiceAccount 必須擁有讀取該 Namespace 下 ConfigMap/Secret 的權限，否則 `lookup` 會失敗。
3. 複雜度管理：雖然我們封裝了邏輯，但這畢竟是在 Template 層級寫程式。如果邏輯過於複雜，可能會導致維護困難。建議只在真正需要處理「混合管理」的資源上使用此模式，普通的資源還是維持標準寫法即可。

## 結語

「基礎設施即程式碼」（IaC）的理想是所有狀態都由 Code 定義，但在真實的遷移過程或混合雲環境中，我們往往需要與「既有狀態」妥協。

透過 Helm 的 `lookup` 與 Smart Resource 模式，我們賦予了 Chart 更高的彈性與適應力。它不再是一個只會盲目覆蓋的推土機，而是一位懂得尊重現狀、優雅協作的管理者。這不僅減少了佈署時的衝突，也讓我們的維運工作更加從容。
