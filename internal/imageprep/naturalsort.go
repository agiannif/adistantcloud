package imageprep

import (
	"cmp"
	"strings"
)

// naturalCompare orders strings so that runs of digits compare by numeric value,
// putting img-2 before img-10. Strings that are numerically equal, such as img-01
// and img-1, fall back to plain string order so sorting stays deterministic.
func naturalCompare(a, b string) int {
	remainderA, remainderB := a, b
	for remainderA != "" && remainderB != "" {
		digitsA, digitsB := leadingDigits(remainderA), leadingDigits(remainderB)
		if digitsA != "" && digitsB != "" {
			// compare by value without converting, so long digit runs cannot overflow
			numberA, numberB := strings.TrimLeft(digitsA, "0"), strings.TrimLeft(digitsB, "0")
			if len(numberA) != len(numberB) {
				return cmp.Compare(len(numberA), len(numberB))
			}
			if c := strings.Compare(numberA, numberB); c != 0 {
				return c
			}
			remainderA, remainderB = remainderA[len(digitsA):], remainderB[len(digitsB):]
			continue
		}
		if remainderA[0] != remainderB[0] {
			return cmp.Compare(int(remainderA[0]), int(remainderB[0]))
		}
		remainderA, remainderB = remainderA[1:], remainderB[1:]
	}

	if c := cmp.Compare(len(remainderA), len(remainderB)); c != 0 {
		return c
	}
	return strings.Compare(a, b)
}

func leadingDigits(s string) string {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	return s[:end]
}
