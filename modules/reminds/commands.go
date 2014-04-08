package reminds

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func registerCommands() {
	Module.Preconnect = reminds.Start
	Module.Disconnect = reminds.Exit

	regComAddRemind()
	regComGetRemind()

	errFns := []func() error{
		regConsPrintRems,
	}

	for _, errFn := range errFns {
		if err := errFn(); err != nil {
			panic(err)
		}
	}
}

func regComAddRemind() {
	Module.RegisterRegexp(module.E_PRIVMSG, remindsR, func(line *irc.Line) {
		lineText := line.Text()
		groups, _ := matchGroups(remindsR, lineText)

		groups["to"] = strings.ToLower(groups["to"])
		groups["duration"] = strings.ToLower(groups["duration"])

		timeN, err := strconv.Atoi(groups["time"])
		if groups["time"] == "" {
			timeN = 0
		} else if err != nil {
			Module.Logger.Errorf("Could not convert `%v` to an int: %v\n  %v\n",
				groups["time"], err, line)
			Module.Conn.Notice(line.Nick, "I'm sorry, but there was an error parsing your remind")

			return
		}

		from := line.Nick
		if groups["to"] == "me" {
			from = "You"
		}

		rem, err := ParseMessage(from, groups["duration"], groups["message"], timeN)
		if err != nil {
			Module.Logger.Errorf("Error parsing remind: %v\n  %v\n",
				err, lineText)
			Module.Conn.Notice(line.Nick, "I'm sorry, but there was an error parsing your remind")

			return
		}

		to := groups["to"]
		if to == "me" {
			to = strings.ToLower(line.Nick)
			groups["to"] = "you"
		}

		reminds.Add(ChanNick{strings.ToLower(line.Target()), to}, rem)

		Module.Conn.Privmsg(line.Target(), fmt.Sprintf("Okay! I'll remind %v about that in %v (%v).",
			groups["to"], rem.Expire.Sub(rem.Set), rem.Expire.Format(timeFormat)))
	})
}

func regComGetRemind() {
	re := regexp.MustCompile(`.*`)

	Module.RegisterRegexp(module.E_PRIVMSG, re, func(line *irc.Line) {
		rems := reminds.GetExpired(ChanNick{strings.ToLower(line.Target()),
			strings.ToLower(line.Nick)})

		for _, rem := range rems {
			Module.Conn.Privmsg(line.Target(), fmt.Sprintf("Oh %s! %s wanted me to remind you %s",
				line.Nick, rem.From, rem.Message))
		}
	})
}

func regConsPrintRems() error {
	err := Module.Console.Register("list", func(string) {
		log.Println(reminds.String())
	})

	return err
}
