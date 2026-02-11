package utils

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// NormalizeUnicode applies NFC normalization to handle special characters (umlauts, accents)
func NormalizeUnicode(text string) string {
	return norm.NFC.String(strings.TrimSpace(text))
}
