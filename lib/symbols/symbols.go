package symbols

import (
	"fmt"
	"strings"
)

const TargetUnknown = "UNKNOWN"

// ToTickerString cleans a given string to a valid string we expect for connect.
func ToTickerString(s string) (string, error) {
	const forbiddenCharacters = "$#%/"
	// remove forbidden characters
	s = strings.Trim(s, forbiddenCharacters)

	// make upper case and remove whitespace and characters from the edges
	s = strings.ToUpper(strings.TrimLeft(strings.TrimRight(s, " ,"), " ,"))
	if s == "" {
		return "", fmt.Errorf("symbols must contain at least one non-whitespace character")
	}

	// change commas to periods so that tickers are valid
	return strings.ReplaceAll(s, ",", "."), nil
}
