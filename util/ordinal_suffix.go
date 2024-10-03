package util

import "strconv"

func OrdinalSuffix(num int) string {
	s := strconv.Itoa(num)

	switch num % 100 {
	case 11, 12, 13:
		return s + "th"

	default:
		switch num % 10 {
		case 1:
			return s + "st"

		case 2:
			return s + "nd"

		case 3:
			return s + "rd"

		default:
			return s + "th"
		}
	}
}
