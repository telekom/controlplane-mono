package labelutil

import "strings"

var (
	unwantedChars = []string{" ", "", "_", "-", "/", "-", "\\", "-"}
	replacer      = strings.NewReplacer(unwantedChars...)
)

func NormalizeValue(value string) string {
	value = strings.ToLower(replacer.Replace(value))
	return strings.Trim(value, "-")
}
