package fcode

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func init() {
	log.SetFlags(0)
}

func registerCommands() {
	// TODO - Write logs
	Module.Preconnect = fCodes.Start
	Module.Disconnect = fCodes.Exit

	regComAdd()
	regComRem()
	regComGet()
	regComGetSystem()

	if err := regConsSave(); err != nil {
		panic(err)
	}
	if err := regConsLoad(); err != nil {
		panic(err)
	}
	if err := regConsList(); err != nil {
		panic(err)
	}
}

func regComAdd() {
	Module.RegisterRegexp(fcAdd, module.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		lineText := strings.ToLower(line.Text())
		groups, _ := matchGroups(fcAdd, lineText)

		ok := fCodes.Add(strings.ToLower(line.Nick), groups["system"], groups["fcode"])
		if !ok {
			Module.Logger.Errorf("fcManager.Add(%v, %v, %v)\n  Line: %v\n",
				strings.ToLower(line.Nick), groups["system"], groups["fcode"], lineText)
			con.Notice(line.Nick, "There was a problem adding you.")

			return
		}

		Module.Logger.Infof("Added friendCode[%v].%v = %v\n",
			groups["nick"], groups["system"], groups["fcode"])
		con.Notice(line.Nick,
			fmt.Sprintf("Saved friend code %v for %v\n", groups["fcode"], groups["system"]))
	})
}

func regComRem() {
	Module.RegisterRegexp(fcRem, module.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		lineText := strings.ToLower(line.Text())
		groups, _ := matchGroups(fcRem, lineText)

		err := fCodes.Remove(strings.ToLower(line.Nick), groups["system"])
		if err != nil {
			Module.Logger.Errorf("fcManager.Remove(%v, %v) error: %v\n  Line: %v\n",
				strings.ToLower(line.Nick), groups["system"], err, lineText)
			con.Notice(line.Nick, err.Error())

			return
		}

		switch groups["system"] {
		case "*":
			Module.Logger.Infof("Deleted friendCode[%v]\n", groups["nick"])
			con.Notice(line.Nick, "Removed you from the database")
		default:
			Module.Logger.Infof("Removed friendCode[%v].%v = ''\n", groups["nick"], groups["system"])
			con.Notice(line.Nick, fmt.Sprintf("Removed nick for %s", groups["system"]))
		}
	})
}

func regComGet() {
	Module.RegisterRegexp(fcGet, module.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		lineText := strings.ToLower(line.Text())
		groups, _ := matchGroups(fcGet, lineText)

		fcMap, err := fCodes.Get(groups["nick"])
		if err != nil {
			Module.Logger.Errorf("fcManager.Get(%v): %v\n  Line: %v\n",
				groups["nick"], err, lineText)
			con.Notice(line.Nick,
				fmt.Sprintf("Sorry I could not find %v in the database", groups["nick"]))

			return
		}

		codes := fmt.Sprintf("%v's friend codes are ", groups["nick"])
		for system, code := range fcMap {
			codes += fmt.Sprintf("(%v: %v) ", system, code)
		}

		switch groups["mode"] {
		case PRIV:
			con.Notice(line.Nick, codes)
		case PUBLIC:
			con.Privmsg(line.Target(), codes)
		}
	})
}

func regComGetSystem() {
	Module.RegisterRegexp(fcList, module.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		lineText := strings.ToLower(line.Text())
		groups, _ := matchGroups(fcList, lineText)

		sysMap := fCodes.GetSystem(groups["system"])
		if len(sysMap) == 0 {
			con.Notice(line.Nick,
				fmt.Sprintf("No one has saved any codes for %v :<", groups["system"]))
		}

		codes := ""
		for nick, code := range sysMap {
			codes += fmt.Sprintf("(%v - %v) ", nick, code)
		}

		switch groups["mode"] {
		case PRIV:
			con.Notice(line.Nick, codes)
		case PUBLIC:
			con.Privmsg(line.Target(), codes)
		}
	})
}

func regComFcHelp() {
	Module.RegisterRegexp(fcHelp, module.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		con.Notice(line.Nick, FcHelp())
	})
}

func regConsSave() error {
	re := regexp.MustCompile(`^save ?(?P<file>.*)?$`)
	err := Module.Console.RegisterRegexp(re, func(s string) {
		groups, _ := matchGroups(re, s)

		if groups["file"] == "" {
			groups["file"] = "codes.gob"
		}

		if err := fCodes.Save(groups["file"]); err != nil {
			errMsg := fmt.Sprintf("Error saving %v: %v", groups["file"], err)
			Module.Logger.Errorln(errMsg)
			log.Println(errMsg)

			return
		}

		log.Printf("Saved codes to %v\n", groups["file"])
	})

	return err
}

func regConsLoad() error {
	re := regexp.MustCompile(`^load ?(?P<file>.*)$`)
	err := Module.Console.RegisterRegexp(re, func(s string) {
		groups, _ := matchGroups(re, s)

		if groups["file"] == "" {
			groups["file"] = "codes.gob"
		}

		if err := fCodes.Load(groups["file"]); err != nil {
			errMsg := fmt.Sprintf("Error loading %v: %v", groups["file"], err)
			Module.Logger.Errorln(errMsg)
			log.Println(errMsg)

			return
		}

		log.Printf("Loaded codes from %v\n", groups["file"])
	})

	return err
}

func regConsList() error {
	err := Module.Console.Register("list", func(s string) {
		fcMap := fCodes.Strings()

		out := ""
		for nick, fCode := range fcMap {
			out += fmt.Sprintf("%-25v %#v\n", nick, fCode)
		}

		log.Print(out)
	})

	return err
}
