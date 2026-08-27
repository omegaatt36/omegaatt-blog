本文的範例程式碼，分成兩份。

一、主文的演算法拆解（uscale.go / main.go），只用標準庫：

    mkdir fpdemo && cd fpdemo
    cp /path/to/uscale.go /path/to/main.go .
    cp go.mod.txt go.mod
    # 從 GOROOT 抄一份 128 位 10 的冪次表過來
    sed 's/^package strconv$/package main/' \
        "$(go env GOROOT)/src/internal/strconv/pow10tab.go" > pow10tab.go
    go run .

需要 Go 1.27.0 以上（pow10tab.go 的 pmHiLo 在 1.26 是 uint128 且採用 floor 表示法，
1.27 才改成 hi<<64 - lo 的 ceiling 表示法，兩者不可混用）。

二、附錄的 math/big 與 shopspring/decimal 比較，需要第三方套件，所以另開一個目錄：

    mkdir fpappendix && cd fpappendix
    cp /path/to/appendix_demo.go.txt main.go
    cp /path/to/appendix_bench_test.go.txt main_test.go
    go mod init fpappendix
    go get github.com/shopspring/decimal@v1.4.0
    go run .
    go test -bench=. -benchmem -count=5

文中的數字都是在 AMD Ryzen 9 5900X / Linux / Go 1.27.0 上，-count 5 取中位數量到的。
