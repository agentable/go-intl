package decimal

type RoundingPriority int

const (
	PriorityAuto RoundingPriority = iota
	PriorityMorePrecision
	PriorityLessPrecision
)

type RoundingType int

const (
	RoundingTypeFractionDigits RoundingType = iota
	RoundingTypeSignificantDigits
	RoundingTypeMorePrecision
	RoundingTypeLessPrecision
)

func ApplyRoundingPriority(hasSD, hasFD bool, priority RoundingPriority) RoundingType {
	switch priority {
	case PriorityAuto:
	case PriorityMorePrecision:
		return RoundingTypeMorePrecision
	case PriorityLessPrecision:
		return RoundingTypeLessPrecision
	}
	if hasSD {
		return RoundingTypeSignificantDigits
	}
	_ = hasFD
	return RoundingTypeFractionDigits
}
