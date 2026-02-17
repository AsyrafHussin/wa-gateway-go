package validator

import (
	"fmt"
	"regexp"
	"strings"
)

type Validator struct {
	countryCode string
	minLength   int
	maxLength   int
	phoneRegex  *regexp.Regexp
}

func New(countryCode string, minLength, maxLength int) *Validator {
	pattern := fmt.Sprintf(`^%s\d{%d,%d}$`, regexp.QuoteMeta(countryCode), minLength-len(countryCode), maxLength-len(countryCode))
	return &Validator{
		countryCode: countryCode,
		minLength:   minLength,
		maxLength:   maxLength,
		phoneRegex:  regexp.MustCompile(pattern),
	}
}

func (v *Validator) ValidatePhone(phone string) (string, error) {
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	if strings.HasPrefix(cleaned, "0") {
		cleaned = v.countryCode + cleaned[1:]
	}

	if !v.phoneRegex.MatchString(cleaned) {
		return "", fmt.Errorf("invalid phone number: must be %d-%d digits starting with %s", v.minLength, v.maxLength, v.countryCode)
	}

	return cleaned, nil
}

func (v *Validator) ValidateMessage(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("message text cannot be empty")
	}
	return nil
}
