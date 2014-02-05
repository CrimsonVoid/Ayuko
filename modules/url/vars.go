package url

import (
	"regexp"

	"github.com/crimsonvoid/irclib/module"
)

var (
	urlRe = regexp.MustCompile(`(http|https)\://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(:[a-zA-Z0-9]*)?/?([a-zA-Z0-9\-\._\?\,\'/\\\+&amp;%\$#\=~])*`)

	Module *module.Module
)
