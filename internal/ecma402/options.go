package ecma402

import (
	"slices"
	"strconv"
	"strings"
)

// StringOption describes one string-backed ECMA-402 option validation rule.
type StringOption struct {
	Name       string
	Value      string
	Values     []string
	AllowEmpty bool
}

// RequiredStringOption returns a string option rule whose value must be one of
// the allowed values.
func RequiredStringOption(name, value string, values ...string) StringOption {
	return StringOption{Name: name, Value: value, Values: values}
}

// OptionalStringOption returns a string option rule whose empty value means the
// option was omitted.
func OptionalStringOption(name, value string, values ...string) StringOption {
	return StringOption{Name: name, Value: value, Values: values, AllowEmpty: true}
}

// OptionalStringOptionInput returns the rule for an optional string option
// after the caller's options bag has preserved whether the option was supplied.
func OptionalStringOptionInput(name, value string, present bool, values ...string) StringOption {
	if present {
		return RequiredStringOption(name, value, values...)
	}
	return OptionalStringOption(name, value, values...)
}

// ApplyOption copies a present scalar option into formatter config.
func ApplyOption[T any](dst *T, value *T) {
	if value == nil {
		return
	}
	*dst = *value
}

// ApplyOptionInput copies a present scalar option and records that the caller
// supplied it explicitly.
func ApplyOptionInput[T any](dst *T, present *bool, value *T) {
	if value == nil {
		return
	}
	*dst = *value
	*present = true
}

// IntegerOption describes one integer ECMA-402 option range validation rule.
type IntegerOption struct {
	Name  string
	Value int
	Min   int
	Max   int
	Set   bool
}

// InvalidStringOption returns the first invalid string-backed option.
func InvalidStringOption(checks ...StringOption) (StringOption, bool) {
	for _, check := range checks {
		if check.AllowEmpty && check.Value == "" {
			continue
		}
		if !slices.Contains(check.Values, check.Value) {
			return check, true
		}
	}
	return StringOption{}, false
}

// InvalidStringOptionError returns an invalid-option error for a failed
// StringOption rule.
func InvalidStringOptionError(owner string, check StringOption, loc string) error {
	return InvalidOptionErrorExpected(owner, check.Name, check.Value, loc, check.Expected(), nil)
}

// ValidateStringOptions returns an invalid-option error for the first failed
// StringOption rule.
func ValidateStringOptions(owner, loc string, checks ...StringOption) error {
	if check, ok := InvalidStringOption(checks...); ok {
		return InvalidStringOptionError(owner, check, loc)
	}
	return nil
}

// UnsupportedStringOptionError returns an unsupported-option error for a valid
// StringOption value outside the active implementation's supported subset.
func UnsupportedStringOptionError(owner string, check StringOption, loc string) error {
	return UnsupportedOptionErrorExpected(owner, check.Name, check.Value, loc, check.Expected(), nil)
}

// ValidateSupportedStringOptions returns an unsupported-option error for the
// first valid option value outside the active implementation's supported subset.
func ValidateSupportedStringOptions(owner, loc string, checks ...StringOption) error {
	if check, ok := InvalidStringOption(checks...); ok {
		return UnsupportedStringOptionError(owner, check, loc)
	}
	return nil
}

// InvalidIntegerOption returns the first integer option outside its range.
func InvalidIntegerOption(checks ...IntegerOption) (IntegerOption, bool) {
	for _, check := range checks {
		if !check.Set {
			continue
		}
		if check.Value < check.Min || check.Value > check.Max {
			return check, true
		}
	}
	return IntegerOption{}, false
}

// InvalidIntegerOptionError returns an invalid-option error for a failed
// IntegerOption rule.
func InvalidIntegerOptionError(owner string, check IntegerOption, loc string) error {
	return InvalidOptionErrorExpected(owner, check.Name, strconv.Itoa(check.Value), loc, check.Expected(), nil)
}

// ValidateIntegerOptions returns an invalid-option error for the first failed
// IntegerOption rule.
func ValidateIntegerOptions(owner, loc string, checks ...IntegerOption) error {
	if check, ok := InvalidIntegerOption(checks...); ok {
		return InvalidIntegerOptionError(owner, check, loc)
	}
	return nil
}

// Expected returns user-facing guidance for the allowed string values.
func (o StringOption) Expected() string {
	switch len(o.Values) {
	case 0:
		return ""
	case 1:
		return strconv.Quote(o.Values[0])
	default:
		return "one of " + quotedValues(o.Values)
	}
}

// Expected returns user-facing guidance for the allowed integer range.
func (o IntegerOption) Expected() string {
	return "an integer from " + strconv.Itoa(o.Min) + " through " + strconv.Itoa(o.Max)
}

func quotedValues(values []string) string {
	var b strings.Builder
	for i, value := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Quote(value))
	}
	return b.String()
}
