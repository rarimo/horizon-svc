package data

import (
	"fmt"
	"sort"
	"strings"
)

// DecodeSorting - transforms sorting attribute specified by user into sorting column
func DecodeSorting(rawSort *string, columns map[string]string) (string, bool, error) {
	if rawSort == nil || *rawSort == "" {
		return "", false, nil
	}

	ascending := !strings.HasPrefix(*rawSort, "-")
	sortStr := strings.ToLower(strings.TrimPrefix(*rawSort, "-"))

	sortParamParts := strings.Split(sortStr, ".")
	sortParam := sortParamParts[len(sortParamParts)-1]
	param, ok := columns[sortParam]
	if !ok {
		return "", false, fmt.Errorf(`"%s" is not supported sorting value. It's only possible to sort by: %s'`, *rawSort, getSupportedSortingValues(columns))
	}

	return param, ascending, nil
}

func getSupportedSortingValues(columns map[string]string) string {
	supportedSortValues := make([]string, 0, len(columns))
	for key := range columns {
		supportedSortValues = append(supportedSortValues, key)
	}
	sort.Strings(supportedSortValues)
	return strings.Join(supportedSortValues, ", ")
}
