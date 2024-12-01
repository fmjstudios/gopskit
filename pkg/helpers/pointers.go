package helpers

func StrPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func Int(i int) *int {
	return &i
}
