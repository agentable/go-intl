package numberformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

// Style selects the formatter's overall presentation. Mirrors Intl.NumberFormat option "style".
type Style string

// CurrencyDisplay selects how currency values are displayed. Mirrors Intl.NumberFormat option "currencyDisplay".
type CurrencyDisplay string

// CurrencySign selects standard or accounting currency signs. Mirrors Intl.NumberFormat option "currencySign".
type CurrencySign string

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
	Locale          locale.Locale `json:"locale"`
	NumberingSystem string        `json:"numberingSystem"`
	Style           Style         `json:"style"`
	// Nil when the resolved style is not "currency".
	Currency *string `json:"currency,omitempty"`
	// Nil when the resolved style is not "currency".
	CurrencyDisplay *CurrencyDisplay `json:"currencyDisplay,omitempty"`
	// Nil when the resolved style is not "currency".
	CurrencySign *CurrencySign `json:"currencySign,omitempty"`
	// Nil when the resolved style is not "unit".
	Unit *string `json:"unit,omitempty"`
	// Nil when the resolved style is not "unit".
	UnitDisplay          *UnitDisplay `json:"unitDisplay,omitempty"`
	MinimumIntegerDigits int          `json:"minimumIntegerDigits"`
	// Nil when ECMA-402 omits fraction digit properties for significant-digit rounding.
	MinimumFractionDigits *int `json:"minimumFractionDigits,omitempty"`
	// Nil when ECMA-402 omits fraction digit properties for significant-digit rounding.
	MaximumFractionDigits *int `json:"maximumFractionDigits,omitempty"`
	// Nil when ECMA-402 omits significant digit properties for fraction-digit rounding.
	MinimumSignificantDigits *int `json:"minimumSignificantDigits,omitempty"`
	// Nil when ECMA-402 omits significant digit properties for fraction-digit rounding.
	MaximumSignificantDigits *int        `json:"maximumSignificantDigits,omitempty"`
	UseGrouping              UseGrouping `json:"useGrouping"`
	Notation                 Notation    `json:"notation"`
	// Nil when the resolved notation is not "compact".
	CompactDisplay      *CompactDisplay     `json:"compactDisplay,omitempty"`
	SignDisplay         SignDisplay         `json:"signDisplay"`
	RoundingIncrement   int                 `json:"roundingIncrement"`
	RoundingMode        RoundingMode        `json:"roundingMode"`
	RoundingPriority    RoundingPriority    `json:"roundingPriority"`
	TrailingZeroDisplay TrailingZeroDisplay `json:"trailingZeroDisplay"`
}

func (f *NumberFormat) ResolvedOptions() ResolvedOptions {
	resolved := f.formatState.resolved
	resolved.Currency = ecma402.CloneResolvedScalar(resolved.Currency)
	resolved.CurrencyDisplay = ecma402.CloneResolvedScalar(resolved.CurrencyDisplay)
	resolved.CurrencySign = ecma402.CloneResolvedScalar(resolved.CurrencySign)
	resolved.Unit = ecma402.CloneResolvedScalar(resolved.Unit)
	resolved.UnitDisplay = ecma402.CloneResolvedScalar(resolved.UnitDisplay)
	resolved.MinimumFractionDigits = ecma402.CloneResolvedScalar(resolved.MinimumFractionDigits)
	resolved.MaximumFractionDigits = ecma402.CloneResolvedScalar(resolved.MaximumFractionDigits)
	resolved.MinimumSignificantDigits = ecma402.CloneResolvedScalar(resolved.MinimumSignificantDigits)
	resolved.MaximumSignificantDigits = ecma402.CloneResolvedScalar(resolved.MaximumSignificantDigits)
	resolved.CompactDisplay = ecma402.CloneResolvedScalar(resolved.CompactDisplay)
	return resolved
}
