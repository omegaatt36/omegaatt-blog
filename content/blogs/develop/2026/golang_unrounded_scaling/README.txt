本文的範例程式碼。執行方式：

    mkdir fpdemo && cd fpdemo
    cp /path/to/uscale.go /path/to/main.go .
    cp go.mod.txt go.mod
    # 從 GOROOT 抄一份 128 位 10 的冪次表過來
    sed 's/^package strconv$/package main/' \
        "$(go env GOROOT)/src/internal/strconv/pow10tab.go" > pow10tab.go
    go run .

需要 Go 1.27.0 以上（pow10tab.go 的 pmHiLo 在 1.26 是 uint128 且採用 floor 表示法，
1.27 才改成 hi<<64 - lo 的 ceiling 表示法，兩者不可混用）。
