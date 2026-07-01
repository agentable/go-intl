package relativetime

import (
	"testing"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/internal/testprocess"
)

// narrowIndexSubprocessEnv gates the narrow-index Once assertion so it runs in a
// freshly started process where no field decode has happened yet.
const narrowIndexSubprocessEnv = "GO_INTL_RELATIVETIME_NARROW_INDEX_SUBPROCESS"

// TestSupportedLocalesDoesNotDecodeFieldBlob asserts the narrow-index rule:
// SupportedLocales reads only the supported blob and must never trigger the
// field blob decode.
//
// The assertion runs in a fresh process so other relative-time tests cannot
// populate the package-level Once state first.
func TestSupportedLocalesDoesNotDecodeFieldBlob(t *testing.T) {
	t.Parallel()

	if !testprocess.RunInFreshProcess(t, narrowIndexSubprocessEnv) {
		return
	}
	testcontract.AssertNarrowStringIndexDoesNotLoad(t, "SupportedLocales", SupportedLocales,
		testcontract.LoadProbe{Name: "relative-time field blob", Loaded: func() bool { return fieldsByLocale != nil }},
	)
}

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedLocales", SupportedLocales)
}

func TestFieldsForReturnsDeepCopy(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}
	fields := FieldsFor(loc)
	dayLong := fields["day"]["long"]
	monthLong := fields["month"]["long"]
	if dayLong.Future == nil || dayLong.Past == nil || dayLong.Relative == nil || monthLong.Future == nil {
		t.Fatal("FieldsFor(en) missing expected day/month long fields")
	}
	wantFuture := dayLong.Future["other"]
	wantPast := dayLong.Past["one"]
	wantRelative := dayLong.Relative["0"]
	wantMonthFuture := monthLong.Future["other"]

	dayLong.Future["other"] = "mutated future"
	dayLong.Past["one"] = "mutated past"
	dayLong.Relative["0"] = "mutated relative"
	fields["day"]["long"] = RelativeTimeField{}
	delete(fields, "month")

	again := FieldsFor(loc)
	if got := again["day"]["long"].Future["other"]; got != wantFuture {
		t.Errorf("FieldsFor(en).day.long.Future[other] = %q, want %q", got, wantFuture)
	}
	if got := again["day"]["long"].Past["one"]; got != wantPast {
		t.Errorf("FieldsFor(en).day.long.Past[one] = %q, want %q", got, wantPast)
	}
	if got := again["day"]["long"].Relative["0"]; got != wantRelative {
		t.Errorf("FieldsFor(en).day.long.Relative[0] = %q, want %q", got, wantRelative)
	}
	if got := again["month"]["long"].Future["other"]; got != wantMonthFuture {
		t.Errorf("FieldsFor(en).month.long.Future[other] = %q, want %q", got, wantMonthFuture)
	}
}

func TestFieldsForMissingLocaleReturnsNil(t *testing.T) {
	t.Parallel()

	if got := FieldsFor(Locale(65535)); got != nil {
		t.Errorf("FieldsFor(missing) = %v, want nil", got)
	}
}

// TestSmokeKnownFields is a checkout-independent smoke test: it asserts a few
// known (unit, style, section, key) tuples resolved through the kernel "en"
// handle return the strings recorded from the committed data.go. These values
// are intentionally hard-coded so a silent encoder/decoder regression fails here
// even when the FormatJS fixtures are unavailable.
func TestSmokeKnownFields(t *testing.T) {
	t.Parallel()

	loc, ok := cldrlocale.ResolveLocale("en")
	if !ok {
		t.Fatal(`ResolveLocale("en") = false, want true`)
	}

	fields := FieldsFor(loc)
	if len(fields) == 0 {
		t.Fatal("FieldsFor(en) returned no fields")
	}

	day := fields["day"]["long"]
	if got, want := day.Future["other"], "in {0} days"; got != want {
		t.Errorf("day/long future[other] = %q, want %q", got, want)
	}
	if got, want := day.Past["one"], "{0} day ago"; got != want {
		t.Errorf("day/long past[one] = %q, want %q", got, want)
	}
	if got, want := day.Relative["0"], "today"; got != want {
		t.Errorf("day/long relative[0] = %q, want %q", got, want)
	}
}

// TestSmokeSupportedLocalesWithinProfile asserts every SupportedLocales tag is a
// member of the kernel locale profile subset. This mirrors the deleted root
// snapshot subset assertion, scoped to the relativetime domain's borrowed
// kernel.
func TestSmokeSupportedLocalesWithinProfile(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceSubset(t, "SupportedLocales", SupportedLocales(), "kernel locale profile", cldrlocale.AvailableLocales())
}
