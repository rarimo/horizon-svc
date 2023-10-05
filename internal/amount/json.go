package amount

import (
	"encoding/json"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"math/big"
)

func (a Amount) MarshalJSON() ([]byte, error) {
	i := big.Int(a)
	return json.Marshal(stringU(&i))
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var rawAmount string
	err := json.Unmarshal(data, &rawAmount)
	if err != nil {
		return errors.Wrap(err, "can't unmarshal amount")
	}

	rawA, err := parseU(rawAmount)
	*a = Amount(*rawA)

	return errors.Wrap(err, "can't parse amount")
}
