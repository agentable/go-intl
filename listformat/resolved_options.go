package listformat

import "github.com/agentable/go-intl/locale"

type LocaleMatcher string
type Type string
type Style string
type PartType string

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	Conjunction Type = "conjunction"
	Disjunction Type = "disjunction"
	Unit        Type = "unit"

	LongStyle   Style = "long"
	ShortStyle  Style = "short"
	NarrowStyle Style = "narrow"

	PartElement PartType = "element"
	PartLiteral PartType = "literal"
)

type ResolvedOptions struct {
	Locale locale.Locale `json:"locale"`
	Type   Type          `json:"type"`
	Style  Style         `json:"style"`
}

func (f *ListFormat) ResolvedOptions() ResolvedOptions {
	return f.resolved
}
