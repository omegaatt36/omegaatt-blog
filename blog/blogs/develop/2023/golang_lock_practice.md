---
title: golnag 樂觀鎖、悲觀鎖 學習筆記與實驗
date: 2023-05-14
categories:
 - develop
tags:
 - golang
---

## 介紹

改變數值的三個步驟
1. 取出
2. 修改
3. 保存

但這三者間的時間差在不同 process、不同 thread、不同 corutine/goroutine 中會造成競爭危害(race condition)。

可以使用多種發法確保並行(concurrency)處理時保持資料的一致性，這裡介紹的是最常使用的悲觀鎖與樂觀鎖。

- 悲觀鎖: 總可能發生問題
    ```
    lock
    (1) 取值
    (2) 修改
    (3) 保存
    unlock
    ```
- 樂觀鎖: 不會總是發生問題
    ```
    (1) 取值
    if *addr == old {
        (2) 修改
        (3) 保存
        return true
    }
    return false
    ```

## 悲觀鎖

golang 中主要使用 `sync.Mutex` 作為悲觀鎖，看似會阻塞住其他 goroutines，但其實 `sync.Mutex` 中也使用到了 CAS。

`sync.Mutex` 中有一個 `int32` 的 `state` 與 `uint32` 的 `sema`(semaphore)

```go
type Mutex struct {
	state int32
	sema  uint32
}

const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken
	mutexStarving
	mutexWaiterShift = iota
    starvationThresholdNs = 1e6
)

func (m *Mutex) Lock() {
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		if race.Enabled {
			race.Acquire(unsafe.Pointer(m))
		}
		return
	}
	m.lockSlow()
}
```

- `state`:
    |29bit|1bit|1bit|1bit|
    | :-:|:-:|:-:|:-:|
    |waiter|starving|woken|locked|

    會使用 CAS 嘗試搶鎖，搶失敗或進入自旋(配合 woken、starving、sema 實現較高效的自旋)
    - locked: 是否已鎖住
    - woken: lock 與 unlock 前設置為 1，告訴正在 unlock、lock 的減少操作
    - starving: 被阻塞超過 1ms。初次搶鎖失敗後的 lockSlow 會進行短暫的 CPU PAUSE 後再次嘗試搶鎖，這將會導致過去搶鎖失敗並進入自旋的 goroutine 發生飢餓。
    - waiter: 正在等候這個鎖的 goroutine 總數。
- `sema`: 同一時間只會發出或接收訊號

## 樂觀鎖

樂觀鎖看似無鎖(在程式中沒有 Lock)，但在多個 gorutine 同時進行自旋時將會耗費大量的 CPU 週期。

golang 中主要是 atomic 包提供樂觀鎖，諸如 `atomic.AddXXX`、`atomic.CompareAndSwapXXX`。

特別需要提到的是 CAP 操作，CAP 使用的是 cpu 提供的 atomic CAS

```assembly
go/src/runtime/internal/atomic/atomic_amd64.s
// bool	·Cas64(uint64 *val, uint64 old, uint64 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else {
//		return 0;
//	}
TEXT ·Cas64(SB), NOSPLIT, $0-25
	MOVQ	ptr+0(FP), BX
	MOVQ	old+8(FP), AX
	MOVQ	new+16(FP), CX
	LOCK
	CMPXCHGQ	CX, 0(BX)
	SETEQ	ret+24(FP)
	RET
```

中間可以看到 LOCK 會鎖 bus 以確保 CMPXCHGQ 不受影響
CMPXCHGQ %cx %bx:
拿 AX(old) 與 BX(share memory) 相比
- 相等: 修改 AX(old) 為 BX，並 ZX 設為 1 (return true)
- 不等: 修改 CX(new) 為 BX，並 ZX 設為 0 (return true)

CAS 的缺點為
1. 多個 goroutine 競爭時，大量 goroutine 都在自旋浪費時間。
2. ABA
    ```
    P1 讀取 A
    context switch 進 P2
    P2 修改 A -> B
    P2 修改 B -> A
    context switch 進 P1
    P1 CAS A -> X
    但其實值已經被 B 改過
    ```

## 實驗

創建一千個 goroutine 並對 num 進行累加至一千

```go
func main() {
	var num int32 = 0

	count := 1e3
	wg := sync.WaitGroup{}
	wg.Add(int(count))
	for index := 0; index < int(count); index++ {
		go increase(&wg, &num, int32(index))
	}

	wg.Wait()
	fmt.Println(num)
}
```

- 不鎖
    ```go
    func increase(wg *sync.WaitGroup, num *int32, old int32) {
        defer wg.Done()
        *num++
    }    
    ```

    最快，但顯而易見的答案是錯的

    ```shell
    ❯ time go run main.go
    908
    go run main.go  0.12s user 0.04s system 145% cpu 0.116 total
    ```

- 悲觀鎖: 使用 sync.Mutex
    ```go
    var m sync.Mutex

    func increase(wg *sync.WaitGroup, num *int32, old int32) {
        defer wg.Done()
        m.Lock()
        defer m.Unlock()
        *num++
    }
    ```

    ```shell
    ❯ time go run main.go
    1000
    go run main.go  0.11s user 0.04s system 138% cpu 0.106 total
    ```    

- 樂觀鎖: 使用 atomic.Add
    ```go
    func increase(wg *sync.WaitGroup, num *int32, old int32) {
        defer wg.Done()
        atomic.AddInt32(num, 1)
    }
    ```

    ```shell
    ❯ time go run main.go
    1000
    go run main.go  0.09s user 0.07s system 143% cpu 0.110 total
    ```

- 樂觀鎖: 使用 CAS 自旋鎖
    ```go
    func increase(wg *sync.WaitGroup, num *int32, old int32) {
        defer wg.Done()
        var retryCount int
        for {
            if atomic.CompareAndSwapInt32(num, old, old+1) {
                break
            }
            retryCount++
        }

        fmt.Printf("interval(%d) retry(%d) times \n", old, retryCount)
    }
    ```

    大部分時間都花在自旋 for-loop 吃滿 CPU 週期

    ```shell
    ❯ time go run main.go
    interval(0) retry(0) times
    ...
    interval(999) retry(85609040) times
    interval(998) retry(80088291) times 
    interval(997) retry(76717475) times 
    interval(995) retry(83328148) times 
    1000
    go run main.go  3829.08s user 20.38s system 797% cpu 8:02.84 total
    ```