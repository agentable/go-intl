package displaynames

type LocaleMatcher string
type Type string
type Style string
type Fallback string
type LanguageDisplay string

const displayNamesOwner = "displaynames"

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	Language      Type = "language"
	Region        Type = "region"
	Script        Type = "script"
	Currency      Type = "currency"
	Calendar      Type = "calendar"
	DateTimeField Type = "dateTimeField"

	LongStyle   Style = "long"
	ShortStyle  Style = "short"
	NarrowStyle Style = "narrow"

	CodeFallback Fallback = "code"
	NoneFallback Fallback = "none"

	DialectLanguageDisplay  LanguageDisplay = "dialect"
	StandardLanguageDisplay LanguageDisplay = "standard"
)

var (
	displayNameTypeValues = [...]string{
		string(Language),
		string(Region),
		string(Script),
		string(Currency),
		string(Calendar),
		string(DateTimeField),
	}
	displayNameStyleValues = [...]string{
		string(LongStyle),
		string(ShortStyle),
		string(NarrowStyle),
	}
	displayNameFallbackValues = [...]string{
		string(CodeFallback),
		string(NoneFallback),
	}
	displayNameLanguageDisplayValues = [...]string{
		string(DialectLanguageDisplay),
		string(StandardLanguageDisplay),
	}
)

type Options struct {
	LocaleMatcher   *string
	Type            *string
	Style           *string
	Fallback        *string
	LanguageDisplay *string
}
