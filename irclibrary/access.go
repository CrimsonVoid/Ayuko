package irclibrary

type access map[string][]string

func (a *access) inGroup(nick, group string) bool {
	users, ok := (*a)[group]
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

func (a *access) InGroups(nick string, groups ...string) bool {
	for _, g := range groups {
		if a.inGroup(g, nick) {
			return true
		}
	}

	return false
}
