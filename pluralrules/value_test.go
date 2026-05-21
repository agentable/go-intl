package pluralrules

func mustCategory(category Category, err error) Category {
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
