package numberformat

import "github.com/agentable/go-intl/locale"

// Style selects the formatter's overall presentation. Mirrors Intl.NumberFormat option "style".
type Style string

// Currency identifies an ISO 4217 currency code. Mirrors Intl.NumberFormat option "currency".
type Currency string

// CurrencyDisplay selects how currency values are displayed. Mirrors Intl.NumberFormat option "currencyDisplay".
type CurrencyDisplay string

// CurrencySign selects standard or accounting currency signs. Mirrors Intl.NumberFormat option "currencySign".
type CurrencySign string

// Unit identifies a sanctioned unit identifier. Mirrors Intl.NumberFormat option "unit".
type Unit string

// UnitDisplay selects how unit names are displayed. Mirrors Intl.NumberFormat option "unitDisplay".
type UnitDisplay string

// UseGrouping selects integer grouping behavior. Mirrors Intl.NumberFormat option "useGrouping".
type UseGrouping string

// Notation selects standard, scientific, engineering, or compact notation. Mirrors Intl.NumberFormat option "notation".
type Notation string

// CompactDisplay selects compact notation width. Mirrors Intl.NumberFormat option "compactDisplay".
type CompactDisplay string

// SignDisplay selects when signs are shown. Mirrors Intl.NumberFormat option "signDisplay".
type SignDisplay string

// RoundingMode selects the decimal rounding mode. Mirrors Intl.NumberFormat option "roundingMode".
type RoundingMode string

// RoundingPriority selects fraction-vs-significant digit conflict resolution. Mirrors Intl.NumberFormat option "roundingPriority".
type RoundingPriority string

// TrailingZeroDisplay selects whether integer zeros are stripped. Mirrors Intl.NumberFormat option "trailingZeroDisplay".
type TrailingZeroDisplay string

// LocaleMatcher selects locale negotiation behavior. Mirrors Intl.NumberFormat option "localeMatcher".
type LocaleMatcher string

const (
	DecimalStyle  Style = "decimal"
	PercentStyle  Style = "percent"
	CurrencyStyle Style = "currency"
	UnitStyle     Style = "unit"

	CurrencyDisplayCode         CurrencyDisplay = "code"
	CurrencyDisplaySymbol       CurrencyDisplay = "symbol"
	CurrencyDisplayNarrowSymbol CurrencyDisplay = "narrowSymbol"
	CurrencyDisplayName         CurrencyDisplay = "name"

	StandardCurrencySign   CurrencySign = "standard"
	AccountingCurrencySign CurrencySign = "accounting"

	ShortUnitDisplay  UnitDisplay = "short"
	NarrowUnitDisplay UnitDisplay = "narrow"
	LongUnitDisplay   UnitDisplay = "long"

	UseGroupingMin2   UseGrouping = "min2"
	UseGroupingAuto   UseGrouping = "auto"
	UseGroupingAlways UseGrouping = "always"
	UseGroupingFalse  UseGrouping = "false"

	StandardNotation    Notation = "standard"
	ScientificNotation  Notation = "scientific"
	EngineeringNotation Notation = "engineering"
	CompactNotation     Notation = "compact"

	ShortCompactDisplay CompactDisplay = "short"
	LongCompactDisplay  CompactDisplay = "long"

	AutoSignDisplay       SignDisplay = "auto"
	AlwaysSignDisplay     SignDisplay = "always"
	ExceptZeroSignDisplay SignDisplay = "exceptZero"
	NegativeSignDisplay   SignDisplay = "negative"
	NeverSignDisplay      SignDisplay = "never"

	CeilRoundingMode       RoundingMode = "ceil"
	FloorRoundingMode      RoundingMode = "floor"
	ExpandRoundingMode     RoundingMode = "expand"
	TruncRoundingMode      RoundingMode = "trunc"
	HalfCeilRoundingMode   RoundingMode = "halfCeil"
	HalfFloorRoundingMode  RoundingMode = "halfFloor"
	HalfExpandRoundingMode RoundingMode = "halfExpand"
	HalfTruncRoundingMode  RoundingMode = "halfTrunc"
	HalfEvenRoundingMode   RoundingMode = "halfEven"

	AutoRoundingPriority          RoundingPriority = "auto"
	MorePrecisionRoundingPriority RoundingPriority = "morePrecision"
	LessPrecisionRoundingPriority RoundingPriority = "lessPrecision"

	AutoTrailingZeroDisplay           TrailingZeroDisplay = "auto"
	StripIfIntegerTrailingZeroDisplay TrailingZeroDisplay = "stripIfInteger"

	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"
)

type ResolvedOptions struct {
	// Locale is the resolved locale. Mirrors Intl.NumberFormat resolved option "locale".
	Locale locale.Locale `json:"locale"`
	// NumberingSystem is the resolved numbering system. Mirrors Intl.NumberFormat resolved option "numberingSystem".
	NumberingSystem string `json:"numberingSystem"`
	// Style is the resolved presentation style. Mirrors Intl.NumberFormat resolved option "style".
	Style Style `json:"style"`
	// Currency is the resolved currency code. Mirrors Intl.NumberFormat resolved option "currency".
	// Empty when the resolved style is not "currency".
	Currency string `json:"currency,omitempty"`
	// CurrencyDisplay is the resolved currency display. Mirrors Intl.NumberFormat resolved option "currencyDisplay".
	// Empty when the resolved style is not "currency".
	CurrencyDisplay CurrencyDisplay `json:"currencyDisplay,omitempty"`
	// CurrencySign is the resolved currency sign. Mirrors Intl.NumberFormat resolved option "currencySign".
	// Empty when the resolved style is not "currency".
	CurrencySign CurrencySign `json:"currencySign,omitempty"`
	// Unit is the resolved unit identifier. Mirrors Intl.NumberFormat resolved option "unit".
	// Empty when the resolved style is not "unit".
	Unit string `json:"unit,omitempty"`
	// UnitDisplay is the resolved unit display. Mirrors Intl.NumberFormat resolved option "unitDisplay".
	// Empty when the resolved style is not "unit".
	UnitDisplay UnitDisplay `json:"unitDisplay,omitempty"`
	// MinimumIntegerDigits is the resolved minimum integer digits. Mirrors Intl.NumberFormat resolved option "minimumIntegerDigits".
	MinimumIntegerDigits int `json:"minimumIntegerDigits"`
	// MinimumFractionDigits is the resolved minimum fraction digits. Mirrors Intl.NumberFormat resolved option "minimumFractionDigits".
	// Nil when ECMA-402 omits fraction digit properties for significant-digit rounding.
	MinimumFractionDigits *int `json:"minimumFractionDigits,omitempty"`
	// MaximumFractionDigits is the resolved maximum fraction digits. Mirrors Intl.NumberFormat resolved option "maximumFractionDigits".
	// Nil when ECMA-402 omits fraction digit properties for significant-digit rounding.
	MaximumFractionDigits *int `json:"maximumFractionDigits,omitempty"`
	// MinimumSignificantDigits is the resolved minimum significant digits. Mirrors Intl.NumberFormat resolved option "minimumSignificantDigits".
	// Nil when ECMA-402 omits significant digit properties for fraction-digit rounding.
	MinimumSignificantDigits *int `json:"minimumSignificantDigits,omitempty"`
	// MaximumSignificantDigits is the resolved maximum significant digits. Mirrors Intl.NumberFormat resolved option "maximumSignificantDigits".
	// Nil when ECMA-402 omits significant digit properties for fraction-digit rounding.
	MaximumSignificantDigits *int `json:"maximumSignificantDigits,omitempty"`
	// UseGrouping is the resolved grouping behavior. Mirrors Intl.NumberFormat resolved option "useGrouping".
	UseGrouping UseGrouping `json:"useGrouping"`
	// Notation is the resolved notation. Mirrors Intl.NumberFormat resolved option "notation".
	Notation Notation `json:"notation"`
	// CompactDisplay is the resolved compact display. Mirrors Intl.NumberFormat resolved option "compactDisplay".
	// Empty when the resolved notation is not "compact".
	CompactDisplay CompactDisplay `json:"compactDisplay,omitempty"`
	// SignDisplay is the resolved sign display. Mirrors Intl.NumberFormat resolved option "signDisplay".
	SignDisplay SignDisplay `json:"signDisplay"`
	// RoundingIncrement is the resolved rounding increment. Mirrors Intl.NumberFormat resolved option "roundingIncrement".
	RoundingIncrement int `json:"roundingIncrement"`
	// RoundingMode is the resolved rounding mode. Mirrors Intl.NumberFormat resolved option "roundingMode".
	RoundingMode RoundingMode `json:"roundingMode"`
	// RoundingPriority is the resolved rounding priority. Mirrors Intl.NumberFormat resolved option "roundingPriority".
	RoundingPriority RoundingPriority `json:"roundingPriority"`
	// TrailingZeroDisplay is the resolved trailing zero display. Mirrors Intl.NumberFormat resolved option "trailingZeroDisplay".
	TrailingZeroDisplay TrailingZeroDisplay `json:"trailingZeroDisplay"`
}

func (f *NumberFormat) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
