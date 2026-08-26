package main

import (
	"math"
	"math/big"
	"math/bits"
)

type unrounded uint64

func bool2[T ~int | ~uint64](b bool) T {
	if b {
		return 1
	}
	return 0
}

// unround makes the unrounded form of x. The algorithms in this file never
// call it. It shows what an unrounded number holds.
func unround(x float64) unrounded {
	return unrounded(math.Floor(4*x)) | bool2[unrounded](math.Floor(4*x) != 4*x)
}

func (u unrounded) floor() uint64         { return uint64((u + 0) >> 2) }
func (u unrounded) roundHalfDown() uint64 { return uint64((u + 1) >> 2) }
func (u unrounded) round() uint64         { return uint64((u + 1 + (u>>2)&1) >> 2) }
func (u unrounded) roundHalfUp() uint64   { return uint64((u + 2) >> 2) }
func (u unrounded) ceil() uint64          { return uint64((u + 3) >> 2) }
func (u unrounded) nudge(d int) unrounded { return u + unrounded(d) }

func (u unrounded) div(d uint64) unrounded {
	x := uint64(u)
	return unrounded(x/d) | u&1 | bool2[unrounded](x%d != 0)
}

func (u unrounded) String() string {
	half := "0"
	if (u>>1)&1 == 1 {
		half = "5"
	}
	sticky := ""
	if u&1 == 1 {
		sticky = "+"
	}
	return "⟨" + itoa(uint64(u>>2)) + "." + half + sticky + "⟩"
}

func itoa(u uint64) string {
	if u == 0 {
		return "0"
	}
	var b [24]byte
	i := len(b)
	for u > 0 {
		i--
		b[i] = byte('0' + u%10)
		u /= 10
	}
	return string(b[i:])
}

// uscaleBig is a slow uscale that is easy to trust. It uses big integers, so
// it has no error. The fast uscale must agree with it for every input.
func uscaleBig(x uint64, e, p int) unrounded {
	num := new(big.Int).SetUint64(x)
	num.Mul(num, big.NewInt(4))
	den := big.NewInt(1)
	if e > 0 {
		num.Lsh(num, uint(e))
	} else {
		den.Lsh(den, uint(-e))
	}
	ten := big.NewInt(10)
	if p > 0 {
		num.Mul(num, new(big.Int).Exp(ten, big.NewInt(int64(p)), nil))
	} else {
		den.Mul(den, new(big.Int).Exp(ten, big.NewInt(int64(-p)), nil))
	}
	q, r := new(big.Int).QuoRem(num, den, new(big.Int))
	return unrounded(q.Uint64()) | bool2[unrounded](r.Sign() != 0)
}

func log10Pow2(x int) int { return (x * 78913) >> 18 }
func log2Pow10(x int) int { return (x * 108853) >> 15 }
func skewed(e int) int    { return (e*631305 - 261663) >> 21 }

// A pmHiLo holds hi<<64 - lo. The minus sign makes the table value a ceiling.
type pmHiLo struct{ hi, lo uint64 }

type scaler struct {
	pmHi, pmLo uint64
	s          int
}

func prescale(pre *scaler, e, p, lp int) {
	pre.pmHi = pow10Tab[p-pow10Min].hi
	pre.pmLo = pow10Tab[p-pow10Min].lo
	pre.s = -(e + lp + 3)
}

// These counters record how often uscale gets an answer from one multiplication.
var fastPathHits, slowPathHits uint64

func uscale(x uint64, c *scaler) unrounded {
	hi, mid := bits.Mul64(x, c.pmHi)
	s := c.s & 63
	if hi>>s<<s != hi {
		fastPathHits++
		return unrounded(hi>>s | 1)
	}
	slowPathHits++
	mid2, _ := bits.Mul64(x, c.pmLo)
	hi -= bool2[uint64](mid < mid2)
	return unrounded(hi>>s | bool2[uint64](mid-mid2 > 1))
}

func unpack64(f float64) (uint64, int) {
	const shift = 64 - 53
	const minExp = -(1074 + shift)
	b := math.Float64bits(f)
	m := 1<<63 | (b&(1<<52-1))<<shift
	e := int((b >> 52) & (1<<11 - 1))
	if e == 0 {
		m &^= 1 << 63
		e = minExp
		s := 64 - bits.Len64(m)
		return m << s, e - s
	}
	return m, (e - 1) + minExp
}

func pack64(m uint64, e int) float64 {
	if m&(1<<52) == 0 {
		return math.Float64frombits(m)
	}
	return math.Float64frombits(m&^(1<<52) | uint64(1075+e)<<52)
}

var uint64pow10 = [...]uint64{
	1, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
}

// FixedWidth returns f as d * 10**p. d has exactly n digits. n must be 18 or less.
func FixedWidth(f float64, n int) (d uint64, p int) {
	m, e := unpack64(f)
	p = n - 1 - log10Pow2(e+63)
	var pre scaler
	prescale(&pre, e, p, log2Pow10(p))
	u := uscale(m, &pre)
	d = u.round()
	if d >= uint64pow10[n] {
		d, p = u.div(10).round(), p-1
	}
	return d, -p
}

// Parse rounds d * 10**p to the nearest float64. d can have 19 digits or fewer.
func Parse(d uint64, p int) float64 {
	b := bits.Len64(d)
	lp := log2Pow10(p)
	e := min(1074, 53-b-lp)
	var pre scaler
	prescale(&pre, e-(64-b), p, lp)
	if pre.s >= 64 {
		return 0
	}
	u := uscale(d<<(64-b), &pre)
	s := bool2[int](u >= unmin(1<<53))
	u = u>>s | u&1
	e = e - s
	return pack64(u.round(), -e)
}

func unmin(x uint64) unrounded { return unrounded(x<<2 - 2) }

// Short returns the shortest d * 10**p that parses back to f.
func Short(f float64) (d uint64, p int) {
	const minExp = -1085
	m, e := unpack64(f)

	var mn, mx uint64
	z := 11
	if m == 1<<63 && e > minExp {
		p = -skewed(e + z)
		mn = m - 1<<(z-2) // min = m - 1/4 * 2**(e+z)
		mx = m + 1<<(z-1) // max = m + 1/2 * 2**(e+z)
	} else {
		if e < minExp {
			z = 11 + (minExp - e)
		}
		p = -log10Pow2(e + z)
		mn = m - 1<<(z-1)
		mx = m + 1<<(z-1)
	}
	odd := int(m>>z) & 1

	var pre scaler
	prescale(&pre, e, p, log2Pow10(p))
	dmin := uscale(mn, &pre).nudge(+odd).ceil()
	dmax := uscale(mx, &pre).nudge(-odd).floor()

	if d = dmax / 10; d*10 >= dmin {
		return trimZeros(d, -(p - 1))
	}
	if d = dmin; d < dmax {
		d = uscale(m, &pre).round()
	}
	return d, -p
}

func trimZeros(x uint64, p int) (uint64, int) {
	for x%10 == 0 && x != 0 {
		x /= 10
		p++
	}
	return x, p
}
