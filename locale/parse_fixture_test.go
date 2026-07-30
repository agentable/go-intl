package locale

func parseLocaleForTest(tag string) Locale {
	loc, err := Parse(tag)
	if err != nil {
		panic(err)
	}
	return loc
}
