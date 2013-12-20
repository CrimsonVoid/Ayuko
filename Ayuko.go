package main

import (
	"github.com/crimsonvoid/irclib"

	"github.com/crimsonvoid/ayuko/modules/fcode"
	"github.com/crimsonvoid/ayuko/modules/reminds"
	"github.com/crimsonvoid/ayuko/modules/zen"
)

func main() {
	m, err := irclib.New("config.json")
	if err != nil {
		panic(err)
	}

	m.Register(fcode.Module)
	m.Register(reminds.Module)
	m.Register(zen.Module)

	m.Connect()

	succExit := false
	for !succExit {
		select {
		case succExit = <-m.Quit:
		}
	}
}
