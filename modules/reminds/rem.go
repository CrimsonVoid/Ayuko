package reminds

import (
	"github.com/crimsonvoid/irclib/module"
)

var Module *module.Module

func init() {
	var err error
	Module, err = module.New("./modules/reminds/reminds.json")
	if err != nil {
		panic(err)
	}

	// TODO - Remove empty map keys
	registerCommands()
}
