package amount

import (
	"fmt"
	"math/big"
	"testing"
)

func TestParseString(t *testing.T) {
	cases := []struct {
		amount string
		result *big.Int
	}{
		{
			amount: "1",
			result: One,
		},
		{
			amount: "1.00000000",
			result: One,
		},
		{
			amount: "0.1",
			result: mustIntFromString("100000000000000000000000"),
		},
		{
			amount: "1.23",
			result: mustIntFromString("1230000000000000000000000"),
		},
		{
			amount: "1.0000000000000000000000001",
			result: One,
		},
	}

	for _, c := range cases {
		t.Run(c.amount, func(t *testing.T) {
			a, err := NewFromString(c.amount)
			if err != nil {
				t.Fatal(err)
			}
			if a.Int().Cmp(c.result) != 0 {
				t.Fatalf("expected %s, got %s", c.result, a.Int())
			}
		})
	}
}

func TestToString(t *testing.T) {
	cases := []struct {
		amount Amount
		result string
	}{
		{
			amount: NewFromInt(1000000000000000000),
			result: "0.000001000000000000000000",
		},
		{
			amount: NewFromInt(100000000000000000),
			result: "0.000000100000000000000000",
		},
		{
			amount: NewFromInt(1230000000000000000),
			result: "0.000001230000000000000000",
		},
	}

	for _, c := range cases {
		t.Run(c.result, func(t *testing.T) {
			if c.amount.String() != c.result {
				t.Fatalf("expected %s, got %s", c.result, c.amount.String())
			}
		})
	}
}

func TestJSONDecode(t *testing.T) {
	cases := []struct {
		amount string
		result *big.Int
	}{
		{
			amount: "1",
			result: One,
		},
		{
			amount: "1.00000000",
			result: One,
		},
		{
			amount: "0.1",
			result: mustIntFromString("100000000000000000000000"),
		},
		{
			amount: "1.23",
			result: mustIntFromString("1230000000000000000000000"),
		},
		{
			amount: "1.0000000000000000000000001",
			result: One,
		},
	}

	for _, c := range cases {
		t.Run(c.amount, func(t *testing.T) {
			var a Amount
			if err := a.UnmarshalJSON([]byte(fmt.Sprintf("\"%s\"", c.amount))); err != nil {
				t.Fatal(err)
			}
			if a.Int().Cmp(c.result) != 0 {
				t.Fatalf("expected %s, got %s", c.result, a.Int())
			}
		})
	}
}

func TestJSONEncode(t *testing.T) {
	cases := []struct {
		amount Amount
		result string
	}{
		{
			amount: NewFromInt(1000000000000000000),
			result: "0.000001000000000000000000",
		},
		{
			amount: NewFromInt(100000000000000000),
			result: "0.000000100000000000000000",
		},
		{
			amount: NewFromInt(1230000000000000000),
			result: "0.000001230000000000000000",
		},
	}

	for _, c := range cases {
		t.Run(c.result, func(t *testing.T) {
			b, err := c.amount.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != fmt.Sprintf("\"%s\"", c.result) {
				t.Fatalf("expected %s, got %s", c.result, string(b))
			}
		})
	}
}

func TestDB(t *testing.T) {
	cases := []struct {
		amount Amount
		result string
	}{
		{
			amount: NewFromInt(1000000000000000000),
			result: "0.000001000000000000000000",
		},
	}

	for _, c := range cases {
		t.Run(c.result, func(t *testing.T) {
			value, err := c.amount.Value()
			if err != nil {
				t.Fatal(err)
			}
			if value.(string) != c.result {
				t.Fatalf("writing to db failed: expected %s, got %s", c.result, value)
			}

			a := &Amount{}
			err = a.Scan(value)
			if err != nil {
				t.Fatal(err)
			}
			if a.Int().Cmp(c.amount.Int()) != 0 {
				t.Fatalf("reading from db failed: expected %s, got %s", c.amount.String(), a.String())
			}
		})
	}
}

func TestMath(t *testing.T) {
	cases := []struct {
		A, B Amount
		Add  string
		Sub  string
		Mul  string
		Div  string
		Pow  string
	}{
		{
			A:   MustNewFromString("3"),
			B:   MustNewFromString("2"),
			Add: "5.000000000000000000000000",
			Sub: "1.000000000000000000000000",
			Mul: "6.000000000000000000000000",
			Div: "5.000000000000000000000000",
			Pow: "9.000000000000000000000000",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s+%s", c.A.String(), c.B.String()), func(t *testing.T) {
			a := Add(c.A, c.B)
			if a.String() != c.Add {
				t.Fatalf("expected %s, got %s", c.Add, a.String())
			}
		})
		t.Run(fmt.Sprintf("%s-%s", c.A.String(), c.B.String()), func(t *testing.T) {
			a := Sub(c.A, c.B)
			if a.String() != c.Sub {
				t.Fatalf("expected %s, got %s", c.Sub, a.String())
			}
		})
		t.Run(fmt.Sprintf("%s*%s", c.A.String(), c.B.String()), func(t *testing.T) {
			a := Mul(c.A, c.B)
			if a.String() != c.Mul {
				t.Fatalf("expected %s, got %s", c.Mul, a.String())
			}
		})
		t.Run(fmt.Sprintf("Sum(%s, %s)", c.B.String(), c.A.String()), func(t *testing.T) {
			a := Sum(c.B, c.A)
			if a.String() != c.Add {
				t.Fatalf("expected %s, got %s", c.Add, a.String())
			}
		})
	}
}

func mustIntFromString(str string) *big.Int {
	i, ok := new(big.Int).SetString(str, 10)
	if !ok {
		panic("invalid int")
	}
	return i
}
