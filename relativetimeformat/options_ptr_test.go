package relativetimeformat

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}
