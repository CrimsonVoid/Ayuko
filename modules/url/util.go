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
