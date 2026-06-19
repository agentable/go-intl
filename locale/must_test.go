package locale

func parseLocaleForTest(tag string) Locale {
	loc, err := Parse(tag)
	if err != nil {
		panic(err)
	}
	return loc
}

func mustParseListForTest(tags ...string) List {
	locales, err := ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}
