package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/crimsonvoid/anyabot/ircbot"
	lib "github.com/crimsonvoid/anyabot/irclibrary"
	"github.com/crimsonvoid/console"
	irc "github.com/fluffle/goirc/client"
)

var quit = make(chan bool)

func main() {
	configFile := flag.String("config", "config.json", "Server configuration")
	flag.Parse()

	client, chanConfig, err := lib.New(*configFile)
	checkError(err, log.Fatalln)

	setupHandlers(client, chanConfig)
	err = client.Connect()
	checkError(err, log.Fatalln)

	log.Printf("%s connected to %s\n", client.Config().Me.Nick, client.Config().Server)
	err = ircbot.Start()
	checkError(err, log.Fatalln)

	// Console Input
	cn := console.New()
	cn.Register(":q", func(s string) {
		client.Quit()
	})
	cn.Register(":s", func(s string) {
		err = ircbot.Save("codes.gob")
		checkError(err, log.Println)
	})
	cn.Register(":l", func(s string) {
		err = ircbot.Load("codes.gob")
		checkError(err, log.Println)
	})
	cn.Register(":codes", func(s string) {
		ircbot.Print()
	})

	go cn.Monitor()
	defer cn.Stop()

	<-quit
	<-time.After(time.Second)
}

func setupHandlers(con *irc.Conn, chanConfig *lib.BotInfo) {
	serverInfo := con.Config()

	// Identify with NickServ and connect to channels
	con.HandleFunc(irc.CONNECTED, func(con *irc.Conn, line *irc.Line) {
		con.Privmsg("NickServ", "IDENTIFY "+serverInfo.Pass)
		for _, ch := range chanConfig.Chans {
			con.Join(ch)
		}
	})

	// Save friend codes and exit
	con.HandleFunc(irc.DISCONNECTED, func(con *irc.Conn, line *irc.Line) {
		err := ircbot.Exit()
		checkError(err, log.Println)

		quit <- true
	})

	// Accept invites from authorized groups
	con.HandleFunc(irc.INVITE, func(con *irc.Conn, line *irc.Line) {
		if chanConfig.Access.InGroups(strings.ToLower(line.Nick), "admin") {
			log.Printf("Invited to %s by %s\n", line.Text(), line.Nick)
			con.Join(line.Text())
		} else {
			con.Notice(line.Nick, "You cannot invite me to channels")
		}
	})

	// Friend Codes
	con.HandleFunc(irc.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		lineText := strings.ToLower(line.Text())

		groups, msg, err := ircbot.MatchFC(line.Nick, lineText)
		if err != nil {
			// log.Println(err)
			return
		}

		switch groups["mode"] {
		case ircbot.PUBLIC:
			con.Privmsg(line.Target(), msg)
		case ircbot.PRIV:
			con.Notice(line.Nick, msg)
		}
	})

	// Leave commands
	con.HandleFunc(irc.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		// Sanitize nick
		nick := escapeRegexp(strings.ToLower(con.Me().Ident))
		leaveR := regexp.MustCompile(fmt.Sprintf(`^(?P<cmd>leave|quit) %s\s?$`, nick))
		groups, err := matchGroups(leaveR, strings.ToLower(line.Text()))

		if err != nil {
			// log.Println(err)
			return
		}

		if chanConfig.Access.InGroups(strings.ToLower(line.Nick), "admin") {
			switch groups["cmd"] {
			case "leave":
				con.Part(line.Target(), fmt.Sprintf("Bye %s", line.Nick))
			case "quit":
				con.Quit(fmt.Sprintf("Bye %s", line.Nick))
			}
		}
	})

	// Log notices
	con.HandleFunc(irc.NOTICE, func(con *irc.Conn, line *irc.Line) {
		log.Printf("Notice from %s\n\t%s\n", line.Ident, line.Text())
	})
}

func checkError(err error, f func(v ...interface{})) {
	if err != nil {
		f(err)
	}
}

func matchGroups(reg *regexp.Regexp, s string) (map[string]string, error) {
	groups := make(map[string]string)
	res := reg.FindStringSubmatch(s)
	if res == nil {
		return nil, fmt.Errorf("%s did not match regexp", s)
	}

	groupNames := reg.SubexpNames()
	for k, v := range groupNames {
		if v != "" {
			groups[v] = res[k]
		}
	}

	return groups, nil
}

func escapeRegexp(s string) string {
	s = strings.Replace(s, `[`, `\[`, -1)
	s = strings.Replace(s, `]`, `\]`, -1)

	s = strings.Replace(s, `{`, `\{`, -1)
	s = strings.Replace(s, `}`, `\}`, -1)

	s = strings.Replace(s, `(`, `\(`, -1)
	s = strings.Replace(s, `)`, `\)`, -1)

	s = strings.Replace(s, `^`, `\^`, -1)
	s = strings.Replace(s, `$`, `\$`, -1)

	s = strings.Replace(s, `*`, `\*`, -1)
	s = strings.Replace(s, `?`, `\?`, -1)
	s = strings.Replace(s, `.`, `\.`, -1)

	return s
}
