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
		nicks := getNicks(groups["ids"])
		duration, message := strings.ToLower(groups["duration"]), groups["message"]

		timeN, err := strconv.Atoi(groups["time"])
		if groups["time"] == "" {
			timeN = 0
		} else if err != nil {
			Module.Logger.Errorf("Could not convert `%v` to an int: %v\n  %v\n",
				groups["time"], err, line)
			Module.Conn.Notice(line.Nick, "I'm sorry, there was an error parsing your remind")

			return
		}

		// map[To]*Message
		msgs := make(map[string]*Message, len(nicks))

		for _, nick := range nicks {
			nick = strings.ToLower(nick)

			to, from := nick, line.Nick
			if nick == "me" {
				to, from = strings.ToLower(line.Nick), "You"
			}

			rem, err := ParseMessage(from, duration, message, timeN)
			if err != nil {
				Module.Logger.Errorf("Error parsing remind: %v\n  %v\n",
					err, lineText)

				break
			}

			msgs[to] = rem
		}

		// Only add if all Messages were parsed without an error
		if len(msgs) != len(nicks) {
			Module.Conn.Notice(line.Nick, "I'm sorry, there was an error parsing your remind")

			return
		}

		toS := make([]string, 0, len(msgs))
		chn := strings.ToLower(line.Target())
		var to string
		var msg *Message

		for to, msg = range msgs {
			reminds.Add(ChanNick{chn, to}, msg)

			if to == strings.ToLower(line.Nick) {
				to = "you"
			}

			toS = append(toS, to)
		}

		timeMsg := "unkown (nil)"
		if msg != nil {
			timeMsg = fmt.Sprintf("%v (%v)",
				msg.Expire.Sub(msg.Set), msg.Expire.Format(timeFormat),
			)
		} else {
			Module.Logger.Errorf("`msg` is nil, this should not happen!\n  Line: %v\n  Parsed Messages: %v\n",
				lineText, msgs)
		}

		whom := ""
		switch len(toS) {
		case 1, 2:
			whom = strings.Join(toS, " and ")
		default:
			toLen := len(toS) - 1
			whom = fmt.Sprintf("%v, and %v", strings.Join(toS[:toLen], ", "), toS[toLen])
		}

		Module.Conn.Privmsg(line.Target(), fmt.Sprintf("Okay I'll remind %v about that in %v.",
			whom, timeMsg,
		))
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

func getNicks(ids string) []string {
	// Case if ends with " and " due to regexp
	ids = strings.TrimSuffix(ids, " and ")

	// Split on ","
	rawIds := strings.Split(ids, ",")

	// If not a comma separated list, split on "and"
	// Case if ids separated by "and"; "me and you and self"
	if len(rawIds) == 1 && rawIds[0] != "and" {
		rawIds = strings.Split(ids, "and")
	} else { // Case if list split with "," and "and"; "me, you and self"
		tmpIds := make([]string, 0, len(rawIds))

		for _, id := range rawIds {
			id = strings.Trim(id, " ")

			if !strings.Contains(id, " ") { // Single 'word'
				tmpIds = append(tmpIds, id)

				continue
			}

			if strings.Contains(id, " and ") {
				tmpIds = append(tmpIds, strings.Split(id, " and ")...)
			} else {
				tmpIds = append(tmpIds, id)
			}
		}

		rawIds = tmpIds
	}

	// Pseudo set
	set := make(map[string]bool, len(rawIds))

	for _, id := range rawIds {
		// Case if an id is "and"; split on "and" removes the id
		if id == " " {
			id = "and"
		} else {
			id = strings.Trim(id, " ")
		}

		// Case for "me, and you, and self"; remove the "and "
		// Impossible for (id == "and ") because of earlier Trim()
		id = strings.TrimPrefix(id, "and ")

		if strings.Trim(id, " ") == "" {
			continue
		}

		set[id] = true
	}

	cleanIds := make([]string, 0, len(set))
	for k, _ := range set {
		cleanIds = append(cleanIds, k)
	}

	return cleanIds
}
