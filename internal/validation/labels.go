package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	labelKeyPrefixRegexDef = "^[a-z]{1,}(([.][a-z]){0,}([-]?[a-z0-9]{1,}){0,}){0,}$" // "prefix" part of k8s label key (before "/")
	labelKeyNameRegexDef   = "^[a-zA-Z0-9]([-A-Za-z0-9_.]{0,61}[a-zA-Z0-9]){0,}$"    // "name" part of k8s label key (after "/")
)

var (
	labelKeyPrefixRegexp = regexp.MustCompile(labelKeyPrefixRegexDef)
	labelKeyNameRegexp   = regexp.MustCompile(labelKeyNameRegexDef)
	labelValueRegexp     = labelKeyNameRegexp
)

// VerifyLabelKey returns error if the provided string is not a proper k8s label key.
func VerifyLabelKey(key string) error {
	return validateLabelKey(key)
}

// VerifyLabelValue returns error if the provided string is not a proper k8s label value.
func VerifyLabelValue(value string) error {
	return validateLabelValue(value)
}

func validateLabelValue(value string) error {
	var labelValue = strings.TrimSpace(value)

	if len(labelValue) == 0 {
		return nil
	}

	if len(labelValue) > 63 {
		return fmt.Errorf("label value too long: %d", len(labelValue))
	}

	if labelValueRegexp.MatchString(labelValue) {
		return nil
	}

	return fmt.Errorf("invalid label value: \"%s\"", labelValue)
}

func validateLabelKey(value string) error {
	var labelKey = strings.TrimSpace(value)

	//max: 253 + 1 + 63
	if (len(labelKey) < 1) || (len(labelKey) > 317) {
		return fmt.Errorf("invalid label key length: %d", len(labelKey))
	}

	// "/" can be only in the middle
	if strings.HasSuffix(labelKey, "/") || strings.HasPrefix(labelKey, "/") {
		return errors.New("invalid position of '/' character")
	}

	prefixAndName := strings.Split(labelKey, "/")

	if len(prefixAndName) == 1 {
		return validateLabelKeyName(prefixAndName[0])
	}
	if len(prefixAndName) == 2 {
		if err := validateLabelKeyPrefix(prefixAndName[0]); err != nil {
			return err
		}
		return validateLabelKeyName(prefixAndName[1])
	}
	return errors.New("too many '/' characters")
}

func validateLabelKeyPrefix(value string) error {
	if len(value) > 253 {
		return fmt.Errorf("label key prefix too long: %d", len(value))
	}

	if labelKeyPrefixRegexp.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid label key prefix: \"%s\"", value)
}

func validateLabelKeyName(value string) error {
	if len(value) > 63 {
		return fmt.Errorf("label key name too long: %d", len(value))
	}

	if labelKeyNameRegexp.MatchString(value) {
		return nil
	}

	return fmt.Errorf("invalid label key name: \"%s\"", value)
}
