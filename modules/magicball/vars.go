package magicball

import (
	"math/rand"
	"time"

	"github.com/crimsonvoid/irclib/module"
)

var (
	Module *module.Module
	rng    = rand.New(rand.NewSource(time.Now().UnixNano()))

	replies = []string{
		"Yes",
		"No",
		"no fuck that",
		"u wot m8",
		"sure whatev",
		"idklol",
		"uh nope",
		"hell yea",
	}
)
