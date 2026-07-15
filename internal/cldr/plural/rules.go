package plural

import (
	"fmt"

	pluralop "github.com/agentable/go-intl/internal/plural"
)

// Rule returns the exact generated plural-rule family for a data locale.
func Rule(loc, typ string) (func(pluralop.OperandsRecord) pluralop.Category, error) {
	var rule func(pluralop.OperandsRecord) pluralop.Category
	var ok bool
	switch typ {
	case "cardinal":
		rule, ok = CardinalRule(loc)
	case "ordinal":
		rule, ok = OrdinalRule(loc)
	default:
		return nil, fmt.Errorf("plural: unknown rule family %q", typ)
	}
	if !ok {
		return nil, fmt.Errorf("plural: missing %s rule for data locale %q", typ, loc)
	}
	return rule, nil
}
