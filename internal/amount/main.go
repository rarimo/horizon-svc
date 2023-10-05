package amount

import (
	"fmt"
	"math/big"
	"strconv"
)

// Amount is a wrapper around big.Int that implements decimal numbers with max precision of 18 digits
type Amount big.Int

const DefaultPrecision = 24

var One, _ = big.NewInt(0).SetString("1000000000000000000000000", 10)

func NewFromString(s string) (Amount, error) {
	i, err := parseU(s)
	if err != nil {
		return Amount{}, err
	}
	return Amount(*i), nil
}

func MustNewFromString(v string) Amount {
	a, err := NewFromString(v)
	if err != nil {
		panic(err)
	}
	return a
}

func NewFromBigInt(v *big.Int) Amount {
	return Amount(*big.NewInt(0).Set(v))
}

func NewFromInt(v int64) Amount {
	return Amount(*big.NewInt(v))
}

// NewFromIntWithPrecision creates a new Amount from an integer with the specified precision
// For example, if v = 100000 and precision = 3 the result amount will be 100.000
func NewFromIntWithPrecision(v *big.Int, precision int) Amount {
	if precision > DefaultPrecision {
		panic(fmt.Sprintf("precision %d is greater than default precision %d", precision, DefaultPrecision))
	}

	precisionDiff := DefaultPrecision - precision
	i := big.NewInt(0).Mul(v, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(precisionDiff)), nil))
	return Amount(*i)
}

func (a Amount) Int() *big.Int {
	i := big.Int(a)
	return big.NewInt(0).Set(&i)
}

// IntWithPrecision returns a big.Int representation of number with the specified precision
// For example, if v = 100.000 and precision = 3 the result amount will be 100000
// if v = 100.000 and precision = 5 the result amount will be 10000000
func (a Amount) IntWithPrecision(precision int) *big.Int {
	if precision > DefaultPrecision {
		panic(fmt.Sprintf("precision %d is greater than default precision %d", precision, DefaultPrecision))
	}

	precisionDiff := DefaultPrecision - precision
	return big.NewInt(0).Div(a.Int(), big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(precisionDiff)), nil))
}

func (a Amount) Float() float64 {
	str := a.String()
	res, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err)
	}
	return res
}

func (a Amount) String() string {
	return stringU(a.Int())
}

func parseU(v string) (*big.Int, error) {
	var f, o, r big.Rat

	_, ok := f.SetString(v)
	if !ok {
		return nil, fmt.Errorf("cannot parse amount: %s", v)
	}

	o.SetInt(One)
	r.Mul(&f, &o)

	is := r.FloatString(0)
	amount := big.NewInt(0)
	amount.SetString(is, 10)
	return amount, nil
}

func stringU(v *big.Int) string {
	var f, o, r big.Rat

	f.SetInt(v)
	o.SetInt(One)
	r.Quo(&f, &o)

	return r.FloatString(DefaultPrecision)
}
