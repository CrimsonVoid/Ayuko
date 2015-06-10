package roll

import (
	"math/rand"
	"time"

	"github.com/crimsonvoid/irclib/module"
)

var (
	rng    = rand.New(rand.NewSource(time.Now().UnixNano()))
	Module *module.Module
)
