package pingrelay

import "strings"

// StrSliceContainsStr returns a boolean if s is found in slice
func StrSliceContainsStr(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func checkMessageForSpam(msg string, triggers []string) bool {
	parts := strings.Split(strings.ToLower(msg), " ")
	for _, part := range parts {
		if StrSliceContainsStr(part, triggers) {
			return true
		}
	}
	return false
}
