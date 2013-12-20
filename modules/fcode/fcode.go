package fcode

import (
	"github.com/crimsonvoid/irclib/module"
)

func init() {
	var err error
	Module, err = module.New("./modules/fcode/fcode.json")
	if err != nil {
		panic(err)
	}

	registerCommands()
}
