package main

import (
	"flag"

	"github.com/crimsonvoid/irclib"
	"github.com/crimsonvoid/irclib/module"

	"github.com/CrimsonVoid/ayuko/modules/choices"
	"github.com/CrimsonVoid/ayuko/modules/magicball"
	"github.com/CrimsonVoid/ayuko/modules/roll"
	"github.com/crimsonvoid/ayuko/modules/fcode"
	"github.com/crimsonvoid/ayuko/modules/reminds"
	"github.com/crimsonvoid/ayuko/modules/url"
	// "github.com/crimsonvoid/ayuko/modules/zen"
)

func main() {
	configFile := flag.String("config", "data/confs/config.toml", "Set a config file")
	flag.Parse()

	module.SetLogDir("./data/logs/")

	m, err := irclib.New(*configFile)
	if err != nil {
		panic(err)
	}

	m.Register(fcode.Module)
	m.Register(reminds.Module)
	m.Register(url.Module)
	m.Register(roll.Module)
	m.Register(magicball.Module)
	m.Register(choices.Module)
	// m.Register(zen.Module)

	m.Connect()

	succExit := false
	for !succExit {
		select {
		case succExit = <-m.Quit:
		}
	}
}
