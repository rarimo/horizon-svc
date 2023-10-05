package amount

import (
	"database/sql/driver"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (a Amount) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a *Amount) Scan(src interface{}) error {
	source, ok := src.(string)
	if !ok {
		return errors.New("type assertion .(string) failed")
	}

	parsed, err := NewFromString(source)
	*a = parsed
	return errors.Wrap(err, "failed to parse amount")
}
