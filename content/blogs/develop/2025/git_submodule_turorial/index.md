---
title: 在習慣 go mod 後重新學習 git submodule
date: 2025-01-24
categories:
  - develop
tags:
  - git
  - golang
  - python
---

## 前言

使用 Golang 作為主要開發已經有五年的時間。最近因工作需要接觸到 Python 專案，並且該專案使用 git submodule 的方式來引用共同函式庫 `common-lib-python`。對於長久使用 Golang 的 `go mod` 的我來說，submodule 是一個相對陌生的概念。藉這個機會撰寫一篇文章，整理並紀錄一下 git submodule 的用法，也作為未來的參考。

這篇文章會假設已經對 git 的基本操作有一定程度的了解，並著重在 submodule 的概念、使用情境以及與 Golang 的 `go mod` 的差異比較。文章內容會以下列流程來呈現：

1. 建立一個新的 Python 專案 `my-python-repo`
2. 建立一個 Python 模組 `my-python-module` 作為 submodule
3. 在 `my-python-repo` 中使用 `my-python-module` 作為 submodule
4. 模擬需求變更，同時修改 `my-python-repo` 與 `my-python-module`，並分別發送 PR
5. 與 Golang 的 `go mod` 進行比較

## 建立 Python 專案與模組

先建立兩個新的 git repo，分別是 `my-python-repo` 與 `my-python-module`。

```bash
# 建立 my-python-repo
mkdir my-python-repo
cd my-python-repo
git init
touch main.py
git add .
git commit -m "Initial commit"

# 建立 my-python-module
cd ..
mkdir my-python-module
cd my-python-module
git init
touch my_module.py
git add .
git commit -m "Initial commit"

# 將兩個專案 push 到 Github/Gitlab 上
```

`my-python-repo` 與 `my-python-module` 都已經是一個獨立的 git repo。

## 將模組加入 submodule

將 `my-python-module` 作為 submodule 加入到 `my-python-repo` 中。

```bash
cd my-python-repo

# 將 my-python-module 加入為 submodule，並將其放在 lib 資料夾下
git submodule add <my-python-module 的 git 網址> lib/my-python-module

git status
# On branch main
# Changes to be committed:
#   (use "git restore --staged <file>..." to unstage)
#         new file:   .gitmodules
#         new file:   lib/my-python-module

git commit -m "Add my-python-module as submodule"
```

可以看到 `my-python-repo` 中多了一個 `.gitmodules` 檔案以及 `lib/my-python-module` 資料夾。`.gitmodules` 檔案紀錄了 submodule 的相關訊息，例如路徑與網址。`lib/my-python-module` 資料夾則是一個指向 `my-python-module` repo的特殊指標。

## 使用 submodule

已經將 `my-python-module` 作為 submodule 加入到 `my-python-repo` 中，可以在 `main.py` 中 import 並使用 `my-python-module` 中的程式碼。

```python
# my-python-repo/main.py
from lib.my_python_module.my_module import hello

hello()
```

```python
# my-python-repo/lib/my_python_module/my_module.py
def hello():
    print("Hello from my-python-module!")
```

## 同時修改專案與 submodule

假設今天需要修改 `my-python-module` 中的 `hello()` 函式，並同時在 `my-python-repo` 中使用新的函式。

```python
# my-python-repo/lib/my_python_module/my_module.py
def hello():
    print("Hello from my-python-module! (v2)")
```

```python
# my-python-repo/main.py
from lib.my_python_module.my_module import hello

hello() # 應該要印出 Hello from my-python-module! (v2)
```

在 `my-python-repo` 中，可以直接修改 `lib/my-python-module` 中的程式碼，這些修改會被視為 `my-python-repo` 的修改。

```bash
cd my-python-repo
git status
# On branch main
# Changes not staged for commit:
#   (use "git add <file>..." to update what will be committed)
#   (use "git restore <file>..." to discard changes in working directory)
#         modified:   lib/my_python_module (modified content)
#
# Untracked files:
#   (use "git add <file>..." to include in what will be committed)
#         main.py
#
no changes added to commit (use "git add" and/or "git commit -a")

git add .
git commit -m "Update my-python-module and use new hello() function"
```

需要分別到 `my-python-module` 中提交修改。

```bash
cd my-python-repo/lib/my_python_module
git add .
git commit -m "Update hello() function to v2"

# 將修改 push 到 my-python-module 的遠端repo
git push origin main
```

需要回到 `my-python-repo` 中，更新 submodule 的指標。

```bash
cd my-python-repo
git status
# On branch main
# Changes to be committed:
#   (use "git restore --staged <file>..." to unstage)
#         modified:   lib/my_python_module

git add lib/my_python_module
git commit -m "Update my-python-module submodule to latest commit"

# 將修改 push 到 my-python-repo 的遠端repo
git push origin main
```

這時，`my-python-repo` 中的 `lib/my_python_module` 就會指向 `my-python-module` 的最新 commit。

## 與 `go mod` 的比較

在 Golang 中，使用 `go mod` 來管理依賴，每個模組都有明確的版本號，並且會紀錄在 `go.mod` 與 `go.sum` 檔案中。`go mod` 會自動下載並管理依賴的版本，相較於 submodule，`go mod` 更加方便與直觀。

然而，`go mod` 的缺點在於，如果需要修改依賴的程式碼，會需要 fork 該模組，並修改 `go.mod` 中的依賴路徑，或是使用 [go workspace](https://go.dev/blog/get-familiar-with-workspaces)。這對於需要頻繁修改依賴的開發流程來說，可能會有些不便。

相較於 Golang 內建的 `go mod` 模組管理方案，Python 生態系的模組管理工具則顯得更加多元且蓬勃發展。從早期的 `virtualenv` 搭配 `pip` 手動管理虛擬環境與依賴，到 `pipenv`、`poetry` 等工具的出現，Python 的模組管理方式不斷在演進。而最近備受矚目的 `uv`，更是將效能推向了新的高度。

以下簡述 Python 模組管理工具的演進歷程：

1. **venv + pip**：這是 Python 早期常用的組合。`venv` 用於創建隔離的虛擬環境，避免不同專案間的依賴衝突；而 `pip` 則用於安裝與管理套件。然而，這種方式需要手動管理 `requirements.txt` 檔案，並且在處理多個環境或複雜依賴時，容易變得混亂且難以維護。
2. **pipenv**：`pipenv` 的出現，可以說是為了解決上述問題。它整合了 `virtualenv` 與 `pip` 的功能，並引入了 `Pipfile` 與 `Pipfile.lock` 來管理專案的依賴。`Pipfile` 類似於 `package.json`，用於定義專案的依賴；而 `Pipfile.lock` 則鎖定了所有依賴的具體版本，確保開發與部署環境的一致性。
3. **poetry**：`poetry` 則更進一步，它不僅提供了依賴管理功能，還包含了建構、打包與發布等完整的專案管理功能。`poetry` 使用 `pyproject.toml` 檔案來管理專案設定與依賴，這個檔案也是 Python 社群近年來推動的標準化設定檔。
4. **uv**：`uv` 是由 [Astral](https://astral.sh/) 開發的全新 Python 包管理器，以速度和正確性作為主要目標。其使用 Rust 語言編寫，相較於 `pip`、`pipenv` 或 `poetry`，`uv` 在安裝與解析依賴的速度上有著顯著的提升。

這些工具的出現，讓 Python 的模組管理變得更加方便與高效。`pipenv`、`poetry` 簡化了虛擬環境與依賴的管理，而 `uv` 則大幅提升了安裝與解析依賴的速度。然而，這也帶來了另一個問題：選擇太多，反而讓人眼花繚亂。

對於新專案來說，`uv` 或許是一個不錯的選擇，它提供了更快的速度與簡潔的介面。但對於既有專案，特別是已經使用 `pipenv` 或 `poetry` 的專案，遷移到 `uv` 可能需要一些額外的工作。

- 使用虛擬環境，避免不同專案間的依賴衝突。
- 使用 `requirements.txt`、`Pipfile`、`pyproject.toml` 等檔案來管理專案的依賴。
- 鎖定依賴的具體版本，確保開發與部署環境的一致性。
- 定期更新依賴，避免安全漏洞與效能問題。

而 git submodule，則提供了另一種（較原始。）不同於上述工具的依賴管理方式。它更著重於程式碼的模組化與可重用性，適合用於管理大型專案或組織內部的共同函式庫。然而，submodule 的複雜性與學習曲線也相對較高，需要開發者根據實際情況做出權衡。

我最終開發時採取的方式為 `uv` 建立 venv 並安裝 pipenv，並使用 pipenv 來無痛銜接既有專案，避免在 Onboard 時就發起過多的挑戰。
