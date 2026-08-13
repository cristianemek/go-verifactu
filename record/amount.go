package record

import (
	"fmt"
	"strconv"
	"strings"
)

type Amount int64

func ParseAmount(s string) (Amount, error) {
	cleanString := strings.TrimSpace(s)

	if cleanString == "" {
		return 0, fmt.Errorf("empty string is not a valid amount")
	}

	isNegative := false

	if cleanString[0] == '-' {
		isNegative = true
	}

	if cleanString[0] == '-' || cleanString[0] == '+' {
		cleanString = cleanString[1:]
		if cleanString == "" {
			return 0, fmt.Errorf("invalid amount string")
		}
	}

	if strings.Contains(cleanString, "-") || strings.Contains(cleanString, "+") {
		return 0, fmt.Errorf("invalid character in amount string")
	}

	integerPart, decimalPart, hasDecimal := strings.Cut(cleanString, ".")

	if !hasDecimal {
		decimalPart = "00"
	}

	if len(integerPart) == 0 {
		integerPart = "0"
	}

	decimalPart, err := normalizeDecimalPart(decimalPart)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal part: %w", err)
	}

	return parseStringParts(integerPart, decimalPart, isNegative)
}

func normalizeDecimalPart(decimalPart string) (string, error) {

	switch {
	case len(decimalPart) == 0:
		return "00", nil

	case len(decimalPart) == 1:
		return decimalPart + "0", nil

	case len(decimalPart) == 2:
		return decimalPart, nil

	case len(decimalPart) > 2:
		return "", fmt.Errorf("2 maximum decimal places allowed, got %d", len(decimalPart))

	default:
		return "", fmt.Errorf("unexpected decimal part length: %d", len(decimalPart))
	}

}

func parseStringParts(integerPart, decimalPart string, isNegative bool) (Amount, error) {

	integerValue, err := strconv.ParseInt(integerPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer part: %w", err)
	}

	decimalValue, err := strconv.ParseInt(decimalPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal part: %w", err)
	}

	result := integerValue*100 + decimalValue
	if isNegative {
		result = -result
	}

	return Amount(result), nil
}
