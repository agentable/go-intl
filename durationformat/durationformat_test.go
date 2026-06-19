package durationformat

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDurationFormatResolvedOptionsDefault(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got := format.ResolvedOptions()
	want := ResolvedOptions{
		Locale:              intltest.Locale(t, "en"),
		NumberingSystem:     "latn",
		Style:               ShortStyle,
		Years:               ShortUnitStyle,
		YearsDisplay:        AutoDisplay,
		Months:              ShortUnitStyle,
		MonthsDisplay:       AutoDisplay,
		Weeks:               ShortUnitStyle,
		WeeksDisplay:        AutoDisplay,
		Days:                ShortUnitStyle,
		DaysDisplay:         AutoDisplay,
		Hours:               ShortUnitStyle,
		HoursDisplay:        AutoDisplay,
		Minutes:             ShortUnitStyle,
		MinutesDisplay:      AutoDisplay,
		Seconds:             ShortUnitStyle,
		SecondsDisplay:      AutoDisplay,
		Milliseconds:        ShortUnitStyle,
		MillisecondsDisplay: AutoDisplay,
		Microseconds:        ShortUnitStyle,
		MicrosecondsDisplay: AutoDisplay,
		Nanoseconds:         ShortUnitStyle,
		NanosecondsDisplay:  AutoDisplay,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolvedOptions() = %#v, want %#v", got, want)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "hi"), intltest.Locale(t, "zh-Hans-CN")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "hi"), intltest.Locale(t, "zh-Hans-CN")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}

	_, err = SupportedLocalesOf(requested, Options{LocaleMatcher: LocaleMatcher("bad")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
	}

	_, err = SupportedLocalesOf(requested, Options{Style: Style("bad")})
	if err != nil {
		t.Fatalf("SupportedLocalesOf(invalid formatting option) error = %v, want nil", err)
	}
}

func TestDurationFormatFormatDefaultShort(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{
		Years:        1,
		Months:       2,
		Days:         3,
		Hours:        4,
		Minutes:      5,
		Seconds:      6,
		Milliseconds: 7,
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "1 yr, 2 mths, 3 days, 4 hr, 5 min, 6 sec, 7 ms"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatFormatEmptyDuration(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{})
	if err != nil {
		t.Fatalf("Format(empty) error = %v", err)
	}
	if got != "" {
		t.Fatalf("Format(empty) = %q, want empty string", got)
	}
	parts, err := format.FormatToParts(Duration{})
	if err != nil {
		t.Fatalf("FormatToParts(empty) error = %v", err)
	}
	if parts != nil {
		t.Fatalf("FormatToParts(empty) = %#v, want nil", parts)
	}
}

func TestDurationFormatFormatDigital(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: DigitalStyle})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}

	got, err := format.Format(Duration{Years: 1})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "1 yr, 0:00:00"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatRejectsInvalidDurationValues(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	tests := []struct {
		name     string
		duration Duration
	}{
		{name: "mixed signs", duration: Duration{Hours: 1, Minutes: -1}},
		{name: "years too large", duration: Duration{Years: 1 << 32}},
		{name: "months too negative", duration: Duration{Months: -(1 << 32)}},
		{name: "normalized seconds too large", duration: Duration{Days: 1 << 40}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := format.Format(tc.duration); !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("Format(%s) error = %v, want intlerr.ErrInvalidValue", tc.name, err)
			}
			if _, err := format.FormatToParts(tc.duration); !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("FormatToParts(%s) error = %v, want intlerr.ErrInvalidValue", tc.name, err)
			}
		})
	}
}

func TestDurationFormatRejectsInvalidFractionalDigits(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		digits int
	}{
		{name: "negative", digits: -1},
		{name: "above max", digits: 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en")}, Options{FractionalDigits: intPtr(tc.digits)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(fractionalDigits=%d) error = %v, want intlerr.ErrInvalidOption", tc.digits, err)
			}
		})
	}
}

func TestDurationFormatFractionalDigitsExplicitZero(t *testing.T) {
	t.Parallel()

	format, err := New(intltest.LocaleList(t, "en"), Options{FractionalDigits: intPtr(0)})
	if err != nil {
		t.Fatalf("New(fractionalDigits=0) error = %v, want nil", err)
	}
	resolved := format.ResolvedOptions()
	if resolved.FractionalDigits == nil || *resolved.FractionalDigits != 0 {
		t.Fatalf("ResolvedOptions().FractionalDigits = %v, want pointer to 0", resolved.FractionalDigits)
	}
}

func TestDurationFormatRejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	en := locale.List{intltest.Locale(t, "en")}
	tests := []struct {
		name string
		opts Options
	}{
		{name: "locale matcher", opts: Options{LocaleMatcher: LocaleMatcher("bad")}},
		{name: "style", opts: Options{Style: Style("bad")}},
		{name: "numbering system", opts: Options{NumberingSystem: "bad!"}},
		{name: "date unit numeric style", opts: Options{Years: NumericUnitStyle}},
		{name: "unit display", opts: Options{HoursDisplay: Display("sometimes")}},
		{name: "fractional unit always display", opts: Options{Milliseconds: NumericUnitStyle, MillisecondsDisplay: AlwaysDisplay}},
		{name: "fractional chain broken", opts: Options{Milliseconds: NumericUnitStyle, Microseconds: ShortUnitStyle}},
		{name: "numeric chain broken", opts: Options{Hours: NumericUnitStyle, Minutes: LongUnitStyle}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(en, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
		})
	}
}

func TestDurationFormatSubsecondRollupIsExact(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:        NarrowStyle,
		Milliseconds: NumericUnitStyle,
	})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{Seconds: 1, Milliseconds: 473})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "1.473s"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatDigitalNegativeSignAppearsOnce(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: DigitalStyle})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}
	got, err := format.Format(Duration{Hours: -1, Minutes: -2, Seconds: -3})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != "-1:02:03" {
		t.Fatalf("Format() = %q, want -1:02:03", got)
	}
}

func TestDurationFormatSubsecondRollupCarriesToParentUnit(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:        NarrowStyle,
		Milliseconds: NumericUnitStyle,
	})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{Seconds: 1, Milliseconds: 1_000})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "2s"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatFormatToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.FormatToParts(Duration{Hours: 1, Minutes: 2})
	if err != nil {
		t.Fatalf("FormatToParts() error = %v", err)
	}
	want := []Part{
		{Type: PartInteger, Value: "1", Unit: Hour},
		{Type: PartLiteral, Value: " ", Unit: Hour},
		{Type: PartUnit, Value: "hr", Unit: Hour},
		{Type: PartLiteral, Value: ", "},
		{Type: PartInteger, Value: "2", Unit: Minute},
		{Type: PartLiteral, Value: " ", Unit: Minute},
		{Type: PartUnit, Value: "min", Unit: Minute},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, want)
	}
}

func TestDurationFormatFormatEqualsFormatToPartsJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  Options
		duration Duration
	}{
		{
			name:    "default short",
			options: Options{},
			duration: Duration{
				Years:        1,
				Months:       2,
				Days:         3,
				Hours:        4,
				Minutes:      5,
				Seconds:      6,
				Milliseconds: 7,
			},
		},
		{
			name:     "empty",
			options:  Options{},
			duration: Duration{},
		},
		{
			name:     "digital",
			options:  Options{Style: DigitalStyle},
			duration: Duration{Hours: 1, Minutes: 2, Seconds: 3},
		},
		{
			name:     "negative digital",
			options:  Options{Style: DigitalStyle},
			duration: Duration{Hours: -1, Minutes: -2, Seconds: -3},
		},
		{
			name: "fractional subsecond",
			options: Options{
				Style:        NarrowStyle,
				Milliseconds: NumericUnitStyle,
			},
			duration: Duration{Seconds: 1, Milliseconds: 473},
		},
		{
			name: "explicit fractional digits",
			options: Options{
				Style:            NarrowStyle,
				Milliseconds:     NumericUnitStyle,
				FractionalDigits: intPtr(2),
			},
			duration: Duration{Seconds: 1, Milliseconds: 230},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.options)
			if err != nil {
				t.Fatalf("New(en) error = %v", err)
			}

			got, err := format.Format(tc.duration)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			parts, err := format.FormatToParts(tc.duration)
			if err != nil {
				t.Fatalf("FormatToParts() error = %v", err)
			}
			if text := durationPartsText(parts); text != got {
				t.Fatalf("Format() = %q, joined FormatToParts() = %q", got, text)
			}
		})
	}
}

func TestDurationFormatFormatUsesPartsOwner(t *testing.T) {
	t.Parallel()

	parsed, err := parser.ParseFile(token.NewFileSet(), "format.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	format := durationMethodDecl(parsed, "Format")
	if format == nil {
		t.Fatal("Format method not found")
	}
	if !durationMethodCalls(format, "FormatToParts") || !durationMethodCalls(format, "joinParts") {
		t.Fatal("Format must derive output from FormatToParts and joinParts to avoid string/parts drift")
	}
}

func TestDurationFormatRejectsMixedSigns(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}
	if _, err := format.Format(Duration{Hours: 1, Minutes: -1}); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("Format(mixed signs) error = %v, want intlerr.ErrInvalidValue", err)
	}
}

func TestDurationFormatRejectsInvalidFractionalDigitsAboveMax(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en")}, Options{FractionalDigits: intPtr(10)})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(fractionalDigits=10) error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func durationPartsText(parts []Part) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func durationMethodDecl(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		decl, ok := decl.(*ast.FuncDecl)
		if ok && decl.Recv != nil && decl.Name.Name == name {
			return decl
		}
	}
	return nil
}

func durationMethodCalls(fn *ast.FuncDecl, name string) bool {
	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if fun.Name != name {
				return true
			}
		case *ast.SelectorExpr:
			if fun.Sel.Name != name {
				return true
			}
		default:
			return true
		}
		found = true
		return false
	})
	return found
}
