package choices

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func registerCommands() {
	regChoices()
}

func regChoices() {
	re := regexp.MustCompile(`^(-|\.)pick (?P<choice>.*)`)

	Module.Register(module.E_PRIVMSG, re, func(line *irc.Line) {
		res := re.FindStringSubmatch(line.Text())

		choices := []string{}
		for _, cs := range strings.Split(res[2], " OR ") {
			choices = append(choices, strings.Split(cs, " or ")...)
		}
		index := rng.Intn(len(choices))

		Module.Conn.Privmsg(
			line.Target(),
			fmt.Sprintf("%v, %v", line.Nick, strings.TrimSpace(choices[index])))
	})
}
