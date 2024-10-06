package util

func OrdinalSuffix(num int) string {
	switch num % 100 {
	case 11, 12, 13:
		return "th"

	default:
		switch num % 10 {
		case 1:
			return "st"

		case 2:
			return "nd"

		case 3:
			return "rd"

		default:
			return "th"
		}
	}
}
