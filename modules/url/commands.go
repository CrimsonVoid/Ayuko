package url

import (
	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func registerCommands() {
	regComParse()
}

func regComParse() {
	Module.RegisterRegexp(module.E_PRIVMSG, urlRe, func(line *irc.Line) {
		url := urlRe.FindString(line.Text())
		if url == "" {
			return
		}

		title, err := Parse(url)
		if err != nil {
			Module.Logger.Errorf("[%v] - %v", url, err)
			return
		}

		Module.Conn.Privmsg(line.Target(), title)
	})
}
