package entity

import (
	"strings"

	"github.com/shopspring/decimal"

	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type validationDetails map[string]string

func (d validationDetails) err() error {
	if len(d) == 0 {
		return nil
	}
	return sharedErrors.New(sharedErrors.CodeValidation, "validation failed", 400).WithDetails(map[string]string(d))
}

func hasAtMostDecimals(d decimal.Decimal, places int32) bool {
	return d.Equal(d.Round(places))
}

func NormalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}
