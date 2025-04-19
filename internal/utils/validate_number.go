package utils

import "strconv"

func IsOrderNumberValid(number string) bool {
	if number == "" {
		return false
	}

	var sum int
	parity := len(number) % 2
	for i := range len(number) {
		digit, _ := strconv.Atoi(string(number[i]))
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
