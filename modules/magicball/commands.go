package magicball

import (
	"fmt"
	"regexp"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func registerCommands() {
	regMagicBall()
}

func regMagicBall() {
	re := regexp.MustCompile(`^(-|\.)8ball .*`)

	Module.Register(module.E_PRIVMSG, re, func(line *irc.Line) {
		index := rng.Intn(len(replies))

		Module.Conn.Privmsg(
			line.Target(),
			fmt.Sprintf("%v: %v", line.Nick, replies[index]))
	})
}
