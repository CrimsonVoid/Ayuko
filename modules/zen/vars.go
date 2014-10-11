package zen

import (
	"time"

	"github.com/crimsonvoid/irclib/module"
)

const (
	timeoutMin = 10
	timeoutInc = 10
	timeoutMax = 60 // minutes

	timeoutResetFact = 1.5
)

var (
	timeoutMult = timeoutMin
	lastTimeout = time.Now()

	zenChan = make(chan string, 10)
	quit    = make(chan bool)

	Module *module.Module
)
