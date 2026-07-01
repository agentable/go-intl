package decimal

import "github.com/cockroachdb/apd/v3"

type Decimal struct {
	inner    apd.Decimal
	form     Form
	negative bool
}

type Form uint8

const (
	Finite Form = iota
	Infinite
	NaN
	NaNSignaling
)

var (
	Zero        Decimal
	NaNValue    = Decimal{form: NaN}
	PosInfinity = Decimal{form: Infinite}
	NegInfinity = Decimal{form: Infinite, negative: true}
)

func (d Decimal) Sign() int {
	if d.IsZero() || d.IsNaN() {
		return 0
	}
	if d.negative {
		return -1
	}
	return 1
}

func (d Decimal) IsZero() bool { return d.form == Finite && d.inner.Coeff.Sign() == 0 }

func (d Decimal) IsNaN() bool { return d.form == NaN || d.form == NaNSignaling }

func (d Decimal) IsInf() bool { return d.form == Infinite }

func (d Decimal) IsFinite() bool { return d.form == Finite }

func (d Decimal) Negative() bool { return d.negative }

func (d Decimal) Cmp(other Decimal) int {
	return d.inner.Cmp(&other.inner)
}

func AbsDiffCmp(base, a, b Decimal) int {
	left := Abs(sub(base, a))
	right := Abs(sub(base, b))
	return left.Cmp(right)
}

func (d Decimal) String() string {
	if d.IsNaN() {
		return "NaN"
	}
	if d.IsInf() {
		if d.negative {
			return "-Infinity"
		}
		return "Infinity"
	}
	if d.IsZero() {
		return "0"
	}
	return d.inner.Text('f')
}
