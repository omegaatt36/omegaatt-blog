package main

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"strconv"
)

func main() {
	demoUnrounded()
	demoRounding()
	demoStickyDiv()
	verifyLogConstants()
	verifySkewed()
	demoFixedWidth()
	demoPowerOfTwoParadox()
	crossCheckUscale()
	roundTripCheck()
	fastPathRate()
}

func demoUnrounded() {
	fmt.Println("### 1. 未捨入數的表示")
	fmt.Printf("%-10s %-6s %s\n", "x", "raw", "⟨x⟩")
	for _, x := range []float64{6, 6.001, 6.499, 6.5, 6.501, 6.999, 7} {
		u := unround(x)
		fmt.Printf("%-10g %-6d %s\n", x, uint64(u), u)
	}
	fmt.Println()
}

func demoRounding() {
	fmt.Println("### 2. 五種捨入全部來自同一份未捨入數")
	fmt.Printf("%-10s %-7s %-9s %-7s %-9s %s\n", "⟨x⟩", "floor", "round½↓", "round", "round½↑", "ceil")
	for _, x := range []float64{6, 6.25, 6.5, 6.75, 7, 7.5, 8.5} {
		u := unround(x)
		fmt.Printf("%-10s %-7d %-9d %-7d %-9d %d\n",
			u, u.floor(), u.roundHalfDown(), u.round(), u.roundHalfUp(), u.ceil())
	}
	fmt.Println()
}

func demoStickyDiv() {
	fmt.Println("### 3. sticky bit 讓除法保住正確捨入")
	u := unround(15.4)
	fmt.Printf("⟨15.4⟩          = %s (raw %d)\n", u, uint64(u))
	fmt.Printf("⟨15.4⟩.div(6)   = %s → round() = %d\n", u.div(6), u.div(6).round())
	fmt.Printf("先捨入再除:      round(15.4)=15, 15/6=2.5 → round = %d\n", 2)
	fmt.Println()
}

// verifyLogConstants compares log10Pow2 and log2Pow10 against exact math for
// every x in the range that float64 conversion uses.
func verifyLogConstants() {
	fmt.Println("### 4. 定點對數近似的窮舉驗證")
	exactLog10Pow2 := func(x int) int {
		// The largest k with 10**k <= 2**x.
		return floorLogRatio(x, 2, 10)
	}
	exactLog2Pow10 := func(x int) int {
		return floorLogRatio(x, 10, 2)
	}
	bad := 0
	lo, hi := -1200, 1200
	for x := lo; x <= hi; x++ {
		if log10Pow2(x) != exactLog10Pow2(x) {
			bad++
			if bad < 5 {
				fmt.Printf("  log10Pow2(%d) = %d, 正確答案 %d\n", x, log10Pow2(x), exactLog10Pow2(x))
			}
		}
	}
	fmt.Printf("log10Pow2 在 x ∈ [%d, %d] 錯誤數: %d\n", lo, hi, bad)

	bad = 0
	lo, hi = -400, 400
	for x := lo; x <= hi; x++ {
		if log2Pow10(x) != exactLog2Pow10(x) {
			bad++
			if bad < 5 {
				fmt.Printf("  log2Pow10(%d) = %d, 正確答案 %d\n", x, log2Pow10(x), exactLog2Pow10(x))
			}
		}
	}
	fmt.Printf("log2Pow10 在 x ∈ [%d, %d] 錯誤數: %d\n", lo, hi, bad)

	for x := 1200; x < 200000; x++ {
		if log10Pow2(x) != exactLog10Pow2(x) {
			fmt.Printf("log10Pow2 第一個失效的正 x = %d\n", x)
			break
		}
	}
	for x := 400; x < 200000; x++ {
		if log2Pow10(x) != exactLog2Pow10(x) {
			fmt.Printf("log2Pow10 第一個失效的正 x = %d\n", x)
			break
		}
	}
	fmt.Println()
}

// floorLogRatio returns ⌊log_to(from**x)⌋ with exact rational math.
func floorLogRatio(x, from, to int) int {
	// Find the largest k with to**k <= from**x.
	// A big.Rat handles a negative x. A big.Int cannot.
	num := big.NewRat(1, 1)
	f := big.NewRat(int64(from), 1)
	if x >= 0 {
		for i := 0; i < x; i++ {
			num.Mul(num, f)
		}
	} else {
		for i := 0; i < -x; i++ {
			num.Quo(num, f)
		}
	}
	t := big.NewRat(int64(to), 1)
	k := 0
	cur := big.NewRat(1, 1)
	if num.Cmp(cur) >= 0 {
		for {
			next := new(big.Rat).Mul(cur, t)
			if next.Cmp(num) > 0 {
				break
			}
			cur = next
			k++
		}
	} else {
		for cur.Cmp(num) > 0 {
			cur.Quo(cur, t)
			k--
		}
	}
	return k
}

func verifySkewed() {
	fmt.Println("### 5. skewed footprint 常數驗證")
	bad := 0
	for e := -1200; e <= 1000; e++ {
		// v = 3/4 * 2**e, the skewed footprint.
		v := new(big.Rat).SetFrac64(3, 4)
		two := big.NewRat(2, 1)
		if e >= 0 {
			for i := 0; i < e; i++ {
				v.Mul(v, two)
			}
		} else {
			for i := 0; i < -e; i++ {
				v.Quo(v, two)
			}
		}
		// Find ⌊log10 v⌋.
		k := 0
		ten := big.NewRat(10, 1)
		cur := big.NewRat(1, 1)
		if v.Cmp(cur) >= 0 {
			for {
				next := new(big.Rat).Mul(cur, ten)
				if next.Cmp(v) > 0 {
					break
				}
				cur = next
				k++
			}
		} else {
			for cur.Cmp(v) > 0 {
				cur.Quo(cur, ten)
				k--
			}
		}
		if skewed(e) != k {
			bad++
			if bad < 4 {
				fmt.Printf("  skewed(%d) = %d, 正確 %d\n", e, skewed(e), k)
			}
		}
	}
	fmt.Printf("skewed 在 e ∈ [-1200, 1000] 錯誤數: %d\n", bad)
	fmt.Println()
}

func demoFixedWidth() {
	fmt.Println("### 6. 固定位數列印：π 的 15 位")
	f := math.Pi
	m, e := unpack64(f)
	fmt.Printf("π = 0x%x * 2**%d  (bits(m)=%d)\n", m>>11, e+11, bits.Len64(m))
	n := 15
	p := n - 1 - log10Pow2(e+63)
	fmt.Printf("p = n-1-⌊log₁₀ 2**(e+63)⌋ = %d\n", p)
	var pre scaler
	prescale(&pre, e, p, log2Pow10(p))
	u := uscale(m, &pre)
	fmt.Printf("uscale(m, e=%d, p=%d) = %s\n", e, p, u)
	d, pp := FixedWidth(f, n)
	fmt.Printf("FixedWidth(π, 15) = %d * 10**%d\n", d, pp)
	fmt.Printf("strconv.FormatFloat(π,'e',14,64) = %s\n", strconv.FormatFloat(f, 'e', 14, 64))
	fmt.Println()
}

func demoPowerOfTwoParadox() {
	fmt.Println("### 7. 2**89 的最短列印悖論")
	f := math.Ldexp(1, 89)
	fmt.Printf("f            = 2**89 = %s\n", new(big.Float).SetFloat64(f).Text('f', 0))
	prev := math.Nextafter(f, 0)
	next := math.Nextafter(f, math.Inf(1))
	fmt.Printf("前一個 float64 = %s\n", new(big.Float).SetFloat64(prev).Text('f', 0))
	fmt.Printf("後一個 float64 = %s\n", new(big.Float).SetFloat64(next).Text('f', 0))
	fmt.Printf("下界中點      = %s\n", new(big.Float).Quo(new(big.Float).Add(bf(prev), bf(f)), big.NewFloat(2)).Text('f', 1))
	fmt.Printf("上界中點      = %s\n", new(big.Float).Quo(new(big.Float).Add(bf(next), bf(f)), big.NewFloat(2)).Text('f', 1))

	// The correctly rounded 16-digit form does not always parse back to f.
	r16 := strconv.FormatFloat(f, 'e', 15, 64)
	fmt.Printf("正確捨入到 16 位 = %s\n", r16)
	back, _ := strconv.ParseFloat(r16, 64)
	fmt.Printf("  parse 回來 == f ? %v\n", back == f)

	d, p := Short(f)
	fmt.Printf("Short(f)      = %d * 10**%d (共 %d 位)\n", d, p, len(fmt.Sprint(d)))
	fmt.Printf("strconv 最短   = %s\n", strconv.FormatFloat(f, 'e', -1, 64))
	back2, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'e', -1, 64), 64)
	fmt.Printf("  parse 回來 == f ? %v\n", back2 == f)
	fmt.Println()
}

func bf(f float64) *big.Float { return new(big.Float).SetPrec(200).SetFloat64(f) }

// crossCheckUscale compares the fast uscale against the big integer uscale.
func crossCheckUscale() {
	fmt.Println("### 8. 快速 uscale vs 大整數 uscale 交叉驗證")
	r := rand.New(rand.NewChaCha8([32]byte{7}))
	bad, n := 0, 0
	for i := 0; i < 2_000_000; i++ {
		f := math.Float64frombits(r.Uint64() % (1 << 63))
		if math.IsNaN(f) || math.IsInf(f, 0) || f == 0 {
			continue
		}
		m, e := unpack64(f)
		if m == 0 {
			continue
		}
		nd := 17
		p := nd - 1 - log10Pow2(e+63)
		if p < pow10Min || p > pow10Max {
			continue
		}
		var pre scaler
		prescale(&pre, e, p, log2Pow10(p))
		if pre.s < 0 || pre.s >= 64 {
			continue
		}
		n++
		got := uscale(m, &pre)
		want := uscaleBig(m, e, p)
		if got != want {
			bad++
			if bad < 4 {
				fmt.Printf("  x=%#x e=%d p=%d got=%d want=%d\n", m, e, p, got, want)
			}
		}
	}
	fmt.Printf("比對 %d 組隨機輸入，不一致 %d 組\n\n", n, bad)
}

func roundTripCheck() {
	fmt.Println("### 9. round-trip 驗證（自寫 Short/Parse vs 標準庫）")
	r := rand.New(rand.NewChaCha8([32]byte{42}))
	bad, n, diff := 0, 0, 0
	for i := 0; i < 1_000_000; i++ {
		f := math.Float64frombits(r.Uint64() % (1 << 63))
		if math.IsNaN(f) || math.IsInf(f, 0) || f == 0 {
			continue
		}
		n++
		d, p := Short(f)
		if Parse(d, p) != f {
			bad++
		}
		std := strconv.FormatFloat(f, 'e', -1, 64)
		mine := fmt.Sprintf("%de%d", d, p)
		sf, _ := strconv.ParseFloat(std, 64)
		mf, _ := strconv.ParseFloat(mine, 64)
		if sf != mf || digitsOf(d) != sigDigits(std) {
			diff++
			if diff < 4 {
				fmt.Printf("  f=%v std=%s mine=%s\n", f, std, mine)
			}
		}
	}
	fmt.Printf("測試 %d 個隨機 float64：round-trip 失敗 %d，與標準庫不同 %d\n\n", n, bad, diff)
}

func digitsOf(d uint64) int { return len(fmt.Sprint(d)) }

func sigDigits(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == 'e' {
			break
		}
		if s[i] >= '0' && s[i] <= '9' {
			n++
		}
	}
	// The shortest form from strconv never ends in a zero, so n is correct.
	return n
}

// fastPathRate measures how often uscale needs only one multiplication.
func fastPathRate() {
	fmt.Println("### 10. 「省下不必要的乘法」命中率")
	r := rand.New(rand.NewChaCha8([32]byte{99}))
	fs := make([]float64, 0, 100000)
	for len(fs) < 100000 {
		f := math.Float64frombits(r.Uint64() % (1 << 63))
		if math.IsNaN(f) || math.IsInf(f, 0) || f == 0 {
			continue
		}
		fs = append(fs, f)
	}

	fastPathHits, slowPathHits = 0, 0
	for _, f := range fs {
		Short(f)
	}
	tot := fastPathHits + slowPathHits
	fmt.Printf("Short:      共 %d 次 uscale，單次乘法就結束 %d 次 (%.2f%%)\n",
		tot, fastPathHits, 100*float64(fastPathHits)/float64(tot))
	fmt.Printf("            平均每個 float64 呼叫 %.2f 次 uscale\n", float64(tot)/float64(len(fs)))

	fastPathHits, slowPathHits = 0, 0
	for _, f := range fs {
		FixedWidth(f, 17)
	}
	tot = fastPathHits + slowPathHits
	fmt.Printf("FixedWidth(17): 共 %d 次 uscale，單次乘法就結束 %d 次 (%.2f%%)\n",
		tot, fastPathHits, 100*float64(fastPathHits)/float64(tot))

	fastPathHits, slowPathHits = 0, 0
	for _, f := range fs {
		FixedWidth(f, 6)
	}
	tot = fastPathHits + slowPathHits
	fmt.Printf("FixedWidth(6):  共 %d 次 uscale，單次乘法就結束 %d 次 (%.2f%%)\n",
		tot, fastPathHits, 100*float64(fastPathHits)/float64(tot))
	fmt.Println()
}
