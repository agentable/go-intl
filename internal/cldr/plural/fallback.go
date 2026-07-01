package plural

import pluralop "github.com/agentable/go-intl/internal/plural"

// CardinalRuleOrDefault returns loc's cardinal rule, then the default English
// cardinal rule, and finally the always-other rule if generated data is absent.
func CardinalRuleOrDefault(loc string) func(pluralop.OperandsRecord) pluralop.Category {
	if rule, ok := CardinalRule(loc); ok {
		return rule
	}
	if rule, ok := CardinalRule("en"); ok {
		return rule
	}
	return otherRule
}

func otherRule(pluralop.OperandsRecord) pluralop.Category {
	return pluralop.Other
}
