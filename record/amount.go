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
		return 0, fmt.Errorf("%w: empty string", ErrInvalidAmount)
	}

	isNegative := false

	if cleanString[0] == '-' {
		isNegative = true
	}

	if cleanString[0] == '-' || cleanString[0] == '+' {
		cleanString = cleanString[1:]
		if cleanString == "" {
			return 0, ErrInvalidAmount
		}
	}

	if strings.Contains(cleanString, "-") || strings.Contains(cleanString, "+") {
		return 0, fmt.Errorf("%w: invalid character", ErrInvalidAmount)
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
		return 0, err
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

	default:
		return "", fmt.Errorf("%w: unexpected decimal part length: %d", ErrInvalidAmount, len(decimalPart))
	}

}

func parseStringParts(integerPart, decimalPart string, isNegative bool) (Amount, error) {

	integerValue, err := strconv.ParseInt(integerPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid integer part: %v", ErrInvalidAmount, err)
	}

	decimalValue, err := strconv.ParseInt(decimalPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid decimal part: %v", ErrInvalidAmount, err)
	}

	result := integerValue*100 + decimalValue
	if isNegative {
		result = -result
	}

	return Amount(result), nil
}

func (a Amount) Format() string {
	v := int64(a)
	sign := ""

	if v < 0 {
		sign = "-"
		v = -v
	}

	intPart := v / 100
	decimalPart := v % 100

	intPartStr := strconv.FormatInt(intPart, 10)
	decimalPartStr := strconv.FormatInt(decimalPart, 10)

	if len(decimalPartStr) == 1 {
		decimalPartStr = "0" + decimalPartStr
	}

	return sign + intPartStr + "." + decimalPartStr
}
