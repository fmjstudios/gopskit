package helpers

import "unicode"

func OnlyLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

func StrPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func Int(i int) *int {
	return &i
}
