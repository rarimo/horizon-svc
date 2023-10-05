package data

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/big"
)

// Int256 - represents
type Int256 struct {
	*big.Int
}

// Value implements the Valuer interface for Uint256
func (b Int256) Value() (driver.Value, error) {
	if b.Int != nil {
		return b.String(), nil
	}
	// int256 - does not allow null
	return "0", nil
}

// Scan implements the Scanner interface for Uint256
func (b *Int256) Scan(value interface{}) error {
	var i sql.NullString
	if err := i.Scan(value); err != nil {
		return err
	}
	b.Int = new(big.Int)
	var ok bool
	if b.Int, ok = b.SetString(i.String, 10); ok {
		return nil
	}
	return fmt.Errorf("could not scan type %T into BigInt", value)
}

func (b Int256) MarshalMsgpack() ([]byte, error) {
	if b.Int != nil {
		return []byte(b.String()), nil
	}
	// int256 - does not allow null
	return []byte("0"), nil
}

func (b *Int256) UnmarshalMsgpack(value []byte) error {
	var ok bool
	b.Int, ok = new(big.Int).SetString(string(value), 10)
	if !ok {
		panic(fmt.Errorf("could not scan type %T into BigInt", value))
	}
	return nil
}
