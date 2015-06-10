package roll

import (
	"fmt"
	"regexp"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func registerCommands() {
	regCommandRoll()
}

func regCommandRoll() {
	re := regexp.MustCompile(`^-roll `)

	Module.Register(module.E_PRIVMSG, re, func(line *irc.Line) {
		num := rng.Intn(101)

		Module.Conn.Privmsg(
			line.Target(),
			fmt.Sprintf("%v: %v%%", line.Nick, num))
	})
}
