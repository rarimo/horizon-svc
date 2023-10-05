package rpc

import (
	"encoding/json"
	"fmt"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Trait struct {
	TraitType string      `json:"trait_type,omitempty"`
	Value     *TraitValue `json:"value,omitempty"`
}

type Traits struct {
	Value []Trait
}

func (t *Traits) UnmarshalJSON(data []byte) error {
	var value []Trait
	err := json.Unmarshal(data, &value)
	if err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
			return errors.Wrap(err, "can't unmarshal traits")
		}

		value, err = unmarshalTraitsFromMap(data)
		if err != nil {
			return errors.Wrap(err, "can't unmarshal traits")
		}
	}

	*t = Traits{Value: value}

	return nil
}

func unmarshalTraitsFromMap(data []byte) ([]Trait, error) {
	var valueMap map[string]json.RawMessage
	err := json.Unmarshal(data, &valueMap)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal traits")
	}

	var traits []Trait
	for key, value := range valueMap {
		var tv *TraitValue
		err = json.Unmarshal(value, &tv)
		if err != nil {
			return nil, errors.Wrap(err, "can't unmarshal trait value")
		}
		traits = append(traits, Trait{
			TraitType: key,
			Value:     tv,
		})
	}

	return traits, nil
}

type TraitValue string

func (t *TraitValue) UnmarshalJSON(data []byte) error {
	var value string
	err := json.Unmarshal(data, &value)
	if err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
			return errors.Wrap(err, "can't unmarshal trait value")
		}

		value, err = unmarshalNumberedTraitValue(data)
		if err != nil {
			return err
		}
	}

	*t = TraitValue(value)
	return errors.Wrap(err, "can't unmarshal trait value")
}

func unmarshalNumberedTraitValue(data []byte) (string, error) {
	var valueInt int
	if err := json.Unmarshal(data, &valueInt); err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
			return "", errors.Wrap(err, "can't unmarshal trait value")
		}

		var valueFloat float64
		if err = json.Unmarshal(data, &valueFloat); err != nil {
			if err != nil {
				return "", errors.Wrap(err, "can't unmarshal trait value")
			}
		}
		return fmt.Sprintf("%f", valueFloat), nil
	}

	return fmt.Sprintf("%d", valueInt), nil
}
