package numberformat

import "github.com/agentable/go-intl/locale"

type Style string
type Currency string
type CurrencyDisplay string
type CurrencySign string
type Unit string
type UnitDisplay string
type UseGrouping string
type Notation string
type CompactDisplay string
type SignDisplay string
type RoundingMode string
type RoundingPriority string
type TrailingZeroDisplay string
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
	Locale                   locale.Locale
	NumberingSystem          string
	Style                    Style
	Currency                 string
	CurrencyDisplay          CurrencyDisplay
	CurrencySign             CurrencySign
	Unit                     string
	UnitDisplay              UnitDisplay
	MinimumIntegerDigits     int
	MinimumFractionDigits    int
	MaximumFractionDigits    int
	MinimumSignificantDigits int
	MaximumSignificantDigits int
	UseGrouping              UseGrouping
	Notation                 Notation
	CompactDisplay           CompactDisplay
	SignDisplay              SignDisplay
	RoundingIncrement        int
	RoundingMode             RoundingMode
	RoundingPriority         RoundingPriority
	TrailingZeroDisplay      TrailingZeroDisplay
}

func (f *NumberFormat) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
