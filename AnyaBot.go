package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/crimsonvoid/anyabot/ircbot"
	lib "github.com/crimsonvoid/anyabot/irclibrary"
	irc "github.com/fluffle/goirc/client"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	configFile = flag.String("config", "config.json", "Server config file")
	quit       = make(chan bool)
)

func main() {
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
	in := make(chan string)
	go func() {
		inp := bufio.NewReader(os.Stdin)
		for {
			s, err := inp.ReadString('\n')
			if err != nil {
				close(in)
				client.Quit()
				break
			}
			if len(s) > 2 {
				in <- s[0 : len(s)-1]
			}
		}
	}()
	go func() {
		for cmd := range in {
			switch cmd {
			case ":q":
				client.Quit()
			case ":s":
				err = ircbot.Start()
				checkError(err, log.Println)
			case ":l":
				err = ircbot.Exit()
				checkError(err, log.Println)
			case ":codes":
				ircbot.Print()
			}
		}
	}()

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
		if chanConfig.Access.InGroup("admin", strings.ToLower(line.Nick)) {
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
		nick := escapeRegexp(strings.ToLower(con.Me().Nick))
		leaveR := regexp.MustCompile(fmt.Sprintf(`^(?P<cmd>leave|quit) %s\s?$`, nick))
		groups, err := matchGroups(leaveR, strings.ToLower(line.Text()))

		if err != nil {
			// log.Println(err)
			return
		}

		if chanConfig.Access.InGroup("admin", strings.ToLower(line.Nick)) {
			switch groups["cmd"] {
			case "leave":
				con.Part(line.Target(), fmt.Sprintf("Bye %s"))
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
