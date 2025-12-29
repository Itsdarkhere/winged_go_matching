package numrange

type Numbers interface {
	int | float64 // add more types as needed
}

func Between[T Numbers](value, min, max T) bool {
	return value >= min && value <= max
}

func NotBetween[T Numbers](value, min, max T) bool {
	return !Between(value, min, max)
}
