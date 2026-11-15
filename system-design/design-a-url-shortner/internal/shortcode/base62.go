package shortcode

import "errors"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Encode(value int64) (string, error) {
	if value < 0 {
		return "", errors.New("id must not be negative")
	}
	if value == 0 {
		return "0", nil
	}
	result := make([]byte, 0, 11)
	for value > 0 {
		result = append(result, alphabet[value%62])
		value /= 62
	}
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return string(result), nil
}
