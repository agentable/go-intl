package listformat

const listFormatOwner = "listformat"

var (
	listTypeValues = [...]string{
		string(Conjunction),
		string(Disjunction),
		string(Unit),
	}
	listStyleValues = [...]string{
		string(LongStyle),
		string(ShortStyle),
		string(NarrowStyle),
	}
)

type Options struct {
	LocaleMatcher *string
	Type          *string
	Style         *string
}
