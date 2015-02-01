package url

import (
	"fmt"
	"regexp"
)

func matchGroups(reg *regexp.Regexp, s string) (map[string]string, error) {
	groups := make(map[string]string)
	res := reg.FindStringSubmatch(s)
	if res == nil {
		return nil, fmt.Errorf("%s did not match regexp", s)
	}

	groupNames := reg.SubexpNames()
	for k, v := range groupNames {
		if v != "" {
			groups[v] = res[k]
		}
	}

	return groups, nil
}

func cleanHTML(com string) string {
	return htmlCleanerR.ReplaceAllLiteralString(com, " ")
}

func takeWhile(s string, f func(rune) bool) string {
	end := 0

	for _, c := range s {
		if f(c) {
			end++
		} else {
			break
		}
	}

	return s[:end]
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }
func isAlpha(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }

func isAlphaNum(r rune) bool { return isDigit(r) || isAlpha(r) }
func isHex(r rune) bool      { return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') }
