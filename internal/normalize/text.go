package normalize

import "strings"

// CollapseSpace trims a string and replaces every run of Unicode whitespace
// with one ordinary space.
func CollapseSpace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// VisibleText normalizes extracted Word text for display.
func VisibleText(value string) string {
	return CollapseSpace(value)
}

// CanonicalName produces a comparison key while leaving display spelling to
// the caller. Russian Е and Ё compare as the same letter for deduplication.
func CanonicalName(value string) string {
	key := strings.ToUpper(CollapseSpace(value))
	return strings.ReplaceAll(key, "Ё", "Е")
}
