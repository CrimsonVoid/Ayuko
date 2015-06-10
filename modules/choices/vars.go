package choices

import (
	"math/rand"
	"time"

	"github.com/crimsonvoid/irclib/module"
)

var (
	Module *module.Module
	rng    = rand.New(rand.NewSource(time.Now().UnixNano()))
)
