package url

import (
	"fmt"
	"strings"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func registerCommands() {
	regComParse()
}

func regComParse() {
	Module.RegisterRegexp(module.E_PRIVMSG, urlRe, func(line *irc.Line) {
		lineText := strings.ToLower(line.Text())
		url := urlRe.FindString(lineText)

		if url == "" {
			return
		}

		title, err := ParseTitle(url)
		if err != nil {
			Module.Logger.Errorf("[%v] - %v", url, err)
			return
		}

		Module.Conn.Privmsg(line.Target(), fmt.Sprintf("[%v] %v", url, title))
	})
}
