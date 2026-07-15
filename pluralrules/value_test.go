package pluralrules

var _ func(*PluralRules, Value) Category = (*PluralRules).Select

func mustRangeCategory(category Category, err error) Category {
	if err != nil {
		panic(err)
	}
	return category
}

func decimalValue(s string) Value {
	value, err := Decimal(s)
	if err != nil {
		panic(err)
	}
	return value
}
