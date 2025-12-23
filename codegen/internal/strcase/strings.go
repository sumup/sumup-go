package strcase

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

// ToCamel converts the provided string into CamelCase and ensures well-known initialisms
// stay capitalized (e.g. app_id -> AppID).
func ToCamel(s string) string {
	tokens := tokenize(s)
	if len(tokens) == 0 {
		return ""
	}

	var b strings.Builder
	for _, token := range tokens {
		b.WriteString(capitalizeToken(token))
	}

	return b.String()
}

// ToLowerCamel converts the provided string into lowerCamelCase with the same
// initialism handling as [ToCamel].
func ToLowerCamel(s string) string {
	tokens := tokenize(s)
	if len(tokens) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(lowerToken(tokens[0]))
	for _, token := range tokens[1:] {
		b.WriteString(capitalizeToken(token))
	}

	return b.String()
}

// ToSnake converts any supported string into snake_case.
func ToSnake(s string) string {
	tokens := tokenize(s)
	if len(tokens) == 0 {
		return ""
	}

	return strings.Join(tokens, "_")
}

func capitalizeToken(token string) string {
	if token == "" {
		return ""
	}

	if isInitialism(token) {
		return strings.ToUpper(token)
	}

	var b strings.Builder
	for i, r := range token {
		if i == 0 {
			b.WriteRune(unicode.ToUpper(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func lowerToken(token string) string {
	if token == "" {
		return ""
	}

	if isInitialism(token) {
		return strings.ToLower(strings.ToUpper(token))
	}

	return token
}

func isInitialism(token string) bool {
	_, ok := commonInitialisms[strings.ToUpper(token)]
	return ok
}

func tokenize(s string) []string {
	if s == "" {
		return nil
	}

	var tokens []string
	var current []rune
	var last rune
	var hasLast bool

	flush := func() {
		if len(current) == 0 {
			return
		}
		tokens = append(tokens, string(current))
		current = current[:0]
	}

	runes := []rune(s)
	for i, r := range runes {
		if !isAlphaNumeric(r) {
			flush()
			hasLast = false
			continue
		}

		if hasLast {
			var next rune
			hasNext := false
			if i+1 < len(runes) {
				next = runes[i+1]
				hasNext = true
			}
			if isBoundary(last, r, next, hasNext) {
				flush()
			}
		}

		current = append(current, unicode.ToLower(r))
		last = r
		hasLast = true
	}

	flush()

	return tokens
}

func isBoundary(prev, curr rune, next rune, hasNext bool) bool {
	switch {
	case unicode.IsDigit(prev) && !unicode.IsDigit(curr):
		return true
	case !unicode.IsDigit(prev) && unicode.IsDigit(curr):
		return true
	case unicode.IsLower(prev) && unicode.IsUpper(curr):
		return true
	case unicode.IsUpper(prev) && unicode.IsUpper(curr) && hasNext && unicode.IsLower(next):
		return true
	default:
		return false
	}
}

func isAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

var commonInitialisms = map[string]struct{}{
	"ACL":   {},
	"AML":   {},
	"API":   {},
	"ASCII": {},
	"AUD":   {},
	"BGN":   {},
	"BIC":   {},
	"BRL":   {},
	"CAD":   {},
	"CHF":   {},
	"CLP":   {},
	"CPU":   {},
	"CSS":   {},
	"CZK":   {},
	"DKK":   {},
	"DNS":   {},
	"EOF":   {},
	"EUR":   {},
	"GBP":   {},
	"GUID":  {},
	"HTML":  {},
	"HTTP":  {},
	"HTTPS": {},
	"IBAN":  {},
	"ID":    {},
	"IP":    {},
	"JPY":   {},
	"JSON":  {},
	"JWT":   {},
	"KYB":   {},
	"KYC":   {},
	"LHS":   {},
	"NFC":   {},
	"NOK":   {},
	"NZD":   {},
	"OTP":   {},
	"PIN":   {},
	"PLN":   {},
	"POS":   {},
	"QPS":   {},
	"RAM":   {},
	"RHS":   {},
	"RON":   {},
	"RPC":   {},
	"RSD":   {},
	"SAR":   {},
	"SDK":   {},
	"SEK":   {},
	"SLA":   {},
	"SMS":   {},
	"SMTP":  {},
	"SQL":   {},
	"SSH":   {},
	"TCP":   {},
	"TLS":   {},
	"TTL":   {},
	"UDP":   {},
	"UI":    {},
	"UID":   {},
	"URI":   {},
	"URL":   {},
	"USD":   {},
	"UTF8":  {},
	"UUID":  {},
	"VAT":   {},
	"VM":    {},
	"XML":   {},
	"XMPP":  {},
	"XSRF":  {},
	"XSS":   {},
}
