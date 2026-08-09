package ecma402

import (
	"errors"
	"testing"
)

func TestInvalidStringOption(t *testing.T) {
	t.Parallel()

	checks := []StringOption{
		RequiredStringOption("required", "a", "a", "b"),
		OptionalStringOption("optional", "", "x"),
		RequiredStringOption("bad", "z", "x", "y"),
	}
	got, ok := InvalidStringOption(checks...)
	if !ok {
		t.Fatal("InvalidStringOption() ok=false, want true")
	}
	if got.Name != "bad" || got.Value != "z" {
		t.Fatalf("InvalidStringOption() = %+v, want bad/z", got)
	}
}

func TestInvalidStringOptionAllValid(t *testing.T) {
	t.Parallel()

	_, ok := InvalidStringOption(
		RequiredStringOption("required", "b", "a", "b"),
		OptionalStringOption("optional", "", "x"),
	)
	if ok {
		t.Fatal("InvalidStringOption() ok=true, want false")
	}
}

func TestOptionalStringOptionInputDistinguishesPresentEmpty(t *testing.T) {
	t.Parallel()

	if _, ok := InvalidStringOption(OptionalStringOptionInput("caseFirst", "", false, "false", "upper", "lower")); ok {
		t.Fatal("OptionalStringOptionInput(omitted empty) invalid = true, want false")
	}

	got, ok := InvalidStringOption(OptionalStringOptionInput("caseFirst", "", true, "false", "upper", "lower"))
	if !ok {
		t.Fatal("OptionalStringOptionInput(present empty) invalid = false, want true")
	}
	if got.Name != "caseFirst" || got.Value != "" {
		t.Fatalf("InvalidStringOption() = %+v, want caseFirst empty", got)
	}
}

func TestApplyScalarOptions(t *testing.T) {
	t.Parallel()

	gotString := "default"
	ApplyOption(&gotString, nil)
	if gotString != "default" {
		t.Fatalf("ApplyOption(nil) = %q, want default", gotString)
	}

	stringValue := "short"
	ApplyOption(&gotString, &stringValue)
	if gotString != "short" {
		t.Fatalf("ApplyOption(value) = %q, want short", gotString)
	}

	gotInt := 1
	ApplyOption(&gotInt, nil)
	if gotInt != 1 {
		t.Fatalf("ApplyOption(nil) = %d, want 1", gotInt)
	}

	intValue := 5
	ApplyOption(&gotInt, &intValue)
	if gotInt != 5 {
		t.Fatalf("ApplyOption(value) = %d, want 5", gotInt)
	}

	gotBool := true
	ApplyOption(&gotBool, nil)
	if !gotBool {
		t.Fatal("ApplyOption(nil) = false, want true")
	}

	boolValue := false
	ApplyOption(&gotBool, &boolValue)
	if gotBool {
		t.Fatal("ApplyOption(value) = true, want false")
	}
}

func TestApplyOptionInputsCopyValueAndPresence(t *testing.T) {
	t.Parallel()

	stringValue := "lookup"
	gotString := "best fit"
	stringPresent := false
	ApplyOptionInput(&gotString, &stringPresent, nil)
	if gotString != "best fit" || stringPresent {
		t.Fatalf("ApplyOptionInput(nil) = %q/%v, want best fit/false", gotString, stringPresent)
	}
	ApplyOptionInput(&gotString, &stringPresent, &stringValue)
	if gotString != "lookup" || !stringPresent {
		t.Fatalf("ApplyOptionInput(value) = %q/%v, want lookup/true", gotString, stringPresent)
	}

	intValue := 3
	gotInt := 0
	intPresent := false
	ApplyOptionInput(&gotInt, &intPresent, nil)
	if gotInt != 0 || intPresent {
		t.Fatalf("ApplyOptionInput(nil) = %d/%v, want 0/false", gotInt, intPresent)
	}
	ApplyOptionInput(&gotInt, &intPresent, &intValue)
	if gotInt != 3 || !intPresent {
		t.Fatalf("ApplyOptionInput(value) = %d/%v, want 3/true", gotInt, intPresent)
	}

	boolValue := true
	gotBool := false
	boolPresent := false
	ApplyOptionInput(&gotBool, &boolPresent, nil)
	if gotBool || boolPresent {
		t.Fatalf("ApplyOptionInput(nil) = %v/%v, want false/false", gotBool, boolPresent)
	}
	ApplyOptionInput(&gotBool, &boolPresent, &boolValue)
	if !gotBool || !boolPresent {
		t.Fatalf("ApplyOptionInput(value) = %v/%v, want true/true", gotBool, boolPresent)
	}
}

func TestInvalidStringOptionErrorCarriesAllowedValues(t *testing.T) {
	t.Parallel()

	check := RequiredStringOption("style", "wide", "short", "long")
	err := InvalidStringOptionError("listformat", check, "en-US")
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("InvalidStringOptionError() error = %v, want ErrInvalidOption", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("InvalidStringOptionError() error = %T, want Error", err)
	}
	if detail.Owner != "listformat" || detail.Name != "style" || detail.Value != "wide" || detail.Locale != "en-US" {
		t.Fatalf("Error = %+v, want listformat style wide en-US", detail)
	}
	if detail.Expected != `one of "short", "long"` {
		t.Fatalf("Error.Expected = %q, want allowed values", detail.Expected)
	}
}

func TestValidateStringOptions(t *testing.T) {
	t.Parallel()

	err := ValidateStringOptions(
		"listformat",
		"en-US",
		RequiredStringOption("type", "conjunction", "conjunction", "disjunction", "unit"),
		OptionalStringOption("style", "", "long", "short", "narrow"),
	)
	if err != nil {
		t.Fatalf("ValidateStringOptions() error = %v, want nil", err)
	}

	err = ValidateStringOptions(
		"listformat",
		"en-US",
		RequiredStringOption("type", "bad", "conjunction", "disjunction", "unit"),
	)
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("ValidateStringOptions() error = %v, want ErrInvalidOption", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("ValidateStringOptions() error = %T, want Error", err)
	}
	if detail.Owner != "listformat" || detail.Name != "type" || detail.Value != "bad" || detail.Locale != "en-US" {
		t.Fatalf("Error = %+v, want listformat type bad en-US", detail)
	}
}

func TestStringOptionExpectedUsesPreciseSingleValueCopy(t *testing.T) {
	t.Parallel()

	check := RequiredStringOption("usage", "search", "sort")
	if got := check.Expected(); got != `"sort"` {
		t.Fatalf("StringOption.Expected() = %q, want %q", got, `"sort"`)
	}
}

func TestValidateSupportedStringOptions(t *testing.T) {
	t.Parallel()

	if err := ValidateSupportedStringOptions("formatter", "en-US", RequiredStringOption("mode", "sort", "sort")); err != nil {
		t.Fatalf("ValidateSupportedStringOptions(sort) error = %v, want nil", err)
	}

	err := ValidateSupportedStringOptions("formatter", "en-US", RequiredStringOption("mode", "search", "sort"))
	if !errors.Is(err, ErrUnsupportedOption) {
		t.Fatalf("ValidateSupportedStringOptions(search) error = %v, want ErrUnsupportedOption", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("ValidateSupportedStringOptions(search) error = %T, want Error", err)
	}
	if detail.Owner != "formatter" || detail.Kind != "unsupportedOption" || detail.Name != "mode" || detail.Value != "search" || detail.Locale != "en-US" {
		t.Fatalf("Error = %+v, want formatter unsupported mode search en-US", detail)
	}
	if detail.Expected != `"sort"` {
		t.Fatalf("Error.Expected = %q, want %q", detail.Expected, `"sort"`)
	}
}

func TestInvalidIntegerOption(t *testing.T) {
	t.Parallel()

	checks := []IntegerOption{
		{Name: "unset", Value: 100, Min: 1, Max: 10},
		{Name: "ok", Value: 5, Min: 1, Max: 10, Set: true},
		{Name: "bad", Value: 11, Min: 1, Max: 10, Set: true},
	}
	got, ok := InvalidIntegerOption(checks...)
	if !ok {
		t.Fatal("InvalidIntegerOption() ok=false, want true")
	}
	if got.Name != "bad" || got.Value != 11 {
		t.Fatalf("InvalidIntegerOption() = %+v, want bad/11", got)
	}
}

func TestInvalidIntegerOptionErrorCarriesRange(t *testing.T) {
	t.Parallel()

	check := IntegerOption{Name: "fractionalDigits", Value: 10, Min: 0, Max: 9, Set: true}
	err := InvalidIntegerOptionError("durationformat", check, "en-US")
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("InvalidIntegerOptionError() error = %v, want ErrInvalidOption", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("InvalidIntegerOptionError() error = %T, want Error", err)
	}
	if detail.Owner != "durationformat" || detail.Name != "fractionalDigits" || detail.Value != "10" || detail.Locale != "en-US" {
		t.Fatalf("Error = %+v, want durationformat fractionalDigits 10 en-US", detail)
	}
	if detail.Expected != "an integer from 0 through 9" {
		t.Fatalf("Error.Expected = %q, want range guidance", detail.Expected)
	}
}

func TestValidateIntegerOptions(t *testing.T) {
	t.Parallel()

	err := ValidateIntegerOptions(
		"datetimeformat",
		"en-US",
		IntegerOption{Name: "fractionalSecondDigits", Value: 2, Min: 1, Max: 3, Set: true},
	)
	if err != nil {
		t.Fatalf("ValidateIntegerOptions() error = %v, want nil", err)
	}

	err = ValidateIntegerOptions(
		"datetimeformat",
		"en-US",
		IntegerOption{Name: "fractionalSecondDigits", Value: 4, Min: 1, Max: 3, Set: true},
	)
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("ValidateIntegerOptions() error = %v, want ErrInvalidOption", err)
	}
	detail, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("ValidateIntegerOptions() error = %T, want Error", err)
	}
	if detail.Owner != "datetimeformat" || detail.Name != "fractionalSecondDigits" || detail.Value != "4" || detail.Locale != "en-US" {
		t.Fatalf("Error = %+v, want datetimeformat fractionalSecondDigits 4 en-US", detail)
	}
}

func TestInvalidIntegerOptionAllValid(t *testing.T) {
	t.Parallel()

	_, ok := InvalidIntegerOption(
		IntegerOption{Name: "unset", Value: 100, Min: 1, Max: 10},
		IntegerOption{Name: "ok", Value: 10, Min: 1, Max: 10, Set: true},
	)
	if ok {
		t.Fatal("InvalidIntegerOption() ok=true, want false")
	}
}
