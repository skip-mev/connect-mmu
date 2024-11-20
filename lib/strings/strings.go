package gecko

import "strings"

func CommaSeparate(tokens []string) string {
	return strings.Join(tokens, ",")
}
