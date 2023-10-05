package handlers

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	minPageSize = uint64(1)
	maxPageSize = uint64(100)
)

func validatePageSize(value uint64) error {
	err := validation.Validate(value, validation.Required, validation.Min(minPageSize), validation.Max(maxPageSize))
	if err != nil {
		return fmt.Errorf("value must be within [%d,%d], but got %d",
			minPageSize, maxPageSize, value)
	}

	return nil
}
