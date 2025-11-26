package stringx

import (
	"strings"
	"unicode"
)

// ToLowerFirstLetter returns the given string with the first letter converted to lower case.
func ToLowerFirstLetter(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

// MakeSingular returns the given string but singular.
func MakeSingular(s string) string {
	if strings.HasSuffix(s, "Status") {
		return s
	}
	return strings.TrimSuffix(s, "s")
}

// MakePlural returns the given string but plural.
func MakePlural(s string) string {
	singular := MakeSingular(s)
	if strings.HasSuffix(singular, "s") {
		return singular + "es"
	}

	return singular + "s"
}
