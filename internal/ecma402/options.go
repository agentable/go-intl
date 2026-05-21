package ecma402

import "slices"

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
