package haleakala

// StrSliceContainsStr returns a boolean if s is found in slice
func StrSliceContainsStr(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
