package irclibrary

type access map[string][]string

func (a access) InGroup(group, nick string) bool {
	users, ok := a[group]
	if !ok {
		return false
	}

	for _, u := range users {
		if u == nick {
			return true
		}
	}

	return false
}

func (a access) InGroups(groups []string, nick string) bool {
	for _, g := range groups {
		if a.InGroup(g, nick) {
			return true
		}
	}

	return false
}
