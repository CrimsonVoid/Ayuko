package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"github.com/crimsonvoid/ircbot"
	irc "github.com/fluffle/goirc/client"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

// func (w *Writter)checkError(err error)
func checkError(err error, f func(v ...interface{})) {
	if err != nil {
		f(err)
	}
}

var configFile = flag.String("config", "config.json", "Server config file")

var quit = make(chan bool)

type FriendCode struct {
	Wii, Wiiu string
	Ds, Ds3   string
	Live, Psn string
	Steam     string
	Bnet      string
}

var friendCodes = make(map[string]FriendCode)

func main() {
	flag.Parse()

	client, chanConfig, err := ircbot.New(*configFile)
	checkError(err, log.Fatalln)

	// Initialize SSLConfig first
	// cfg.SSLConfig.InsecureSkipVerify = true

	setupHandlers(client, chanConfig)

	err = client.Connect()
	checkError(err, log.Fatalln)
	loadCodes()
	fmt.Printf("%s connected to %s\n", client.Config().Me.Nick, client.Config().Server)

	// Handle console input
	// TODO|Console - Unbuffered chan?
	in := make(chan string, 4)
	// global commands :[command]
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
				saveCodes()
			case ":l":
				loadCodes()
			case ":codes":
				for nick, codes := range friendCodes {
					fmt.Printf("%-25s%#v\n", nick, codes)
				}
			}
		}
	}()

	<-quit
	<-time.After(time.Second)
}

// This is kind of ugly, but wat do
func setupHandlers(con *irc.Conn, chanConfig *ircbot.BotInfo) {
	serverInfo := con.Config()

	con.HandleFunc(irc.CONNECTED, func(con *irc.Conn, line *irc.Line) {
		con.Privmsg("NickServ", "IDENTIFY "+serverInfo.Pass)
		for _, ch := range chanConfig.Chans {
			con.Join(ch)
		}
	})

	con.HandleFunc(irc.DISCONNECTED, func(con *irc.Conn, line *irc.Line) {
		saveCodes()
		quit <- true
	})

	con.HandleFunc(irc.INVITE, func(con *irc.Conn, line *irc.Line) {
		// TODO - Helper function to allow certain groups
		for _, acc := range chanConfig.Access {
			for _, nick := range acc {
				if line.Nick == nick {
					con.Join(line.Text())
					return
				}
			}
		}

		con.Notice(line.Nick, "You cannot invite me to channels :<")
	})

	// Friend Code
	con.HandleFunc(irc.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		systemsL := "wii|wiiu|nid|ds|3ds|psn|live|steam|bnet"
		systemsR := fmt.Sprintf("(?P<system>%s)", systemsL)
		nickR := "(?P<nick>[\\w{}\\[\\]^|`-]+)"

		fcAdd := regexp.MustCompile(fmt.Sprintf("^[.-]fc(ode)? add %s (?P<fcode>.*)", systemsR))
		fcRem := regexp.MustCompile(fmt.Sprintf("^[.-]fc(ode)? rem (%s|\\*)$", systemsR))
		fcGet := regexp.MustCompile(fmt.Sprintf("^[.-]fc(ode)? %s\\s?$", nickR))
		fcList := regexp.MustCompile(fmt.Sprintf("^[.-]fc(ode)? list %s$", systemsR))
		fcHelp := regexp.MustCompile("^[.-]fc(ode)? help$")

		lineText := strings.ToLower(line.Text())

		groups, err := MatchGroups(fcHelp, lineText)
		if err == nil {
			fmt.Println("fcHelp\t", groups, "\t", lineText)

			con.Notice(line.Nick, fmt.Sprintf("Save and retrieve gaming identities. Syntax: .fc(ode) add|rem %s (code) OR .fc [nick]", systemsL))
			return
		}

		groups, err = MatchGroups(fcAdd, lineText)
		if err == nil {
			fmt.Println("fcAdd\t", groups, "\t", lineText)

			fCode := friendCodes[strings.ToLower(line.Nick)]
			switch groups["system"] {
			case "wii":
				codeMatch := regexp.MustCompile("\\d{4}-\\d{4}-\\d{4}-\\d{4}$")
				res := codeMatch.FindStringSubmatch(groups["fcode"])
				if res == nil {
					con.Notice(line.Nick, fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]))
					return
				}

				fCode.Wii = groups["fcode"]
			case "wiiu", "nid":
				codeMatch := regexp.MustCompile("^.{6,16}\\s?$")
				res := codeMatch.FindStringSubmatch(groups["fcode"])

				if res == nil {
					con.Notice(line.Nick, fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]))
					return
				}

				fCode.Wiiu = groups["fcode"]
			case "ds":
				codeMatch := regexp.MustCompile("^\\d{4}-\\d{4}-\\d{4}\\s?$")
				res := codeMatch.FindStringSubmatch(groups["fcode"])
				if res == nil {
					con.Notice(line.Nick, fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]))
					return
				}

				fCode.Ds = groups["fcode"]
			case "3ds":
				codeMatch := regexp.MustCompile("^\\d{4}-\\d{4}-\\d{4}\\s?$")
				res := codeMatch.FindStringSubmatch(groups["fcode"])
				if res == nil {
					con.Notice(line.Nick, fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]))
					return
				}

				fCode.Ds3 = groups["fcode"]
			case "psn":
				codeMatch := regexp.MustCompile("^.{6,16}\\s?$")
				res := codeMatch.FindStringSubmatch(groups["fcode"])

				if res == nil {
					con.Notice(line.Nick, fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]))
					return
				}

				fCode.Psn = groups["fcode"]
			case "live":
				codeMatch := regexp.MustCompile("^.{6,15}\\s?$")
				res := codeMatch.FindStringSubmatch(groups["fcode"])

				if res == nil {
					con.Notice(line.Nick, fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]))
					return
				}

				fCode.Live = groups["fcode"]
			case "steam":
				// TODO|fcode - steam nick restrictions
				fCode.Steam = groups["fcode"]
			case "bnet":
				codeMatch := regexp.MustCompile("^.*\\d{3}\\s?$")
				res := codeMatch.FindStringSubmatch(groups["fcode"])

				if res == nil {
					con.Notice(line.Nick, fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]))
					return
				}

				fCode.Bnet = groups["fcode"]
			}

			friendCodes[strings.ToLower(line.Nick)] = fCode
			con.Notice(line.Nick, fmt.Sprintf("Saved %s friend code %s", groups["system"], groups["fcode"]))

			return
		}

		groups, err = MatchGroups(fcRem, lineText)
		if err == nil {
			fmt.Println("fcRem\t", groups, "\t", lineText)
			
			fc, ok := friendCodes[strings.ToLower(line.Nick)]
			if !ok {
				con.Notice(line.Nick, "You have not saved any user names")
				return
			}

			switch groups["system"] {
			case "wii":
				fc.Wii = ""
			case "wiiu", "nid":
				fc.Wiiu = ""
			case "ds":
				fc.Ds = ""
			case "3ds":
				fc.Ds3 = ""
			case "psn":
				fc.Psn = ""
			case "live":
				fc.Live = ""
			case "steam":
				fc.Steam = ""
			case "bnet":
				fc.Bnet = ""
			default:
				// .fc rem *
				delete(friendCodes, strings.ToLower(line.Nick))
				con.Notice(line.Nick, "Removed you from database")
				return
			}
			friendCodes[strings.ToLower(line.Nick)] = fc

			con.Notice(line.Nick, fmt.Sprintf("Removed nick for %s", groups["system"]))

			return
		}

		groups, err = MatchGroups(fcList, lineText)
		if err == nil {
			fmt.Println("fcList\t", groups, "\t", lineText)

			codeList := ""
			for nick, codes := range friendCodes {
				switch groups["system"] {
				case "wii":
					if codes.Wii != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Wii)
					}
				case "wiiu", "nid":
					if codes.Wiiu != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Wiiu)
					}
				case "ds":
					if codes.Ds != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Ds)
					}
				case "3ds":
					if codes.Ds3 != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Ds3)
					}
				case "psn":
					if codes.Psn != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Psn)
					}
				case "live":
					if codes.Live != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Live)
					}
				case "steam":
					if codes.Steam != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Steam)
					}
				case "bnet":
					if codes.Bnet != "" {
						codeList += fmt.Sprintf("(%s - %s) ", nick, codes.Bnet)
					}
				}
			}

			con.Notice(line.Nick, codeList)
            return
		}

		groups, err = MatchGroups(fcGet, lineText)
		if err == nil {
			fmt.Println("fcGet\t", groups, "\t", lineText)
			fc, ok := friendCodes[strings.ToLower(groups["nick"])]
			if !ok {
				con.Notice(line.Nick, fmt.Sprintf("%s has not saved any friend codes", groups["nick"]))
				return
			}

			codes := ""
			if fc.Wii != "" {
				codes += fmt.Sprintf("wii:(%s) ", fc.Wii)
			}
			if fc.Wiiu != "" {
				codes += fmt.Sprintf("wiiu:(%s) ", fc.Wiiu)
			}
			if fc.Ds != "" {
				codes += fmt.Sprintf("ds:(%s) ", fc.Ds)
			}
			if fc.Ds3 != "" {
				codes += fmt.Sprintf("3ds:(%s) ", fc.Ds3)
			}
			if fc.Live != "" {
				codes += fmt.Sprintf("Live:(%s) ", fc.Live)
			}
			if fc.Psn != "" {
				codes += fmt.Sprintf("PSN:(%s) ", fc.Psn)
			}
			if fc.Steam != "" {
				codes += fmt.Sprintf("Steam:(%s) ", fc.Steam)
			}
			if fc.Bnet != "" {
				codes += fmt.Sprintf("Bnet:(%s) ", fc.Bnet)
			}

			con.Notice(line.Nick, fmt.Sprintf("%s's friend codes are %s", groups["nick"], codes))
			return
		}
	})

	con.HandleFunc(irc.PRIVMSG, func(con *irc.Conn, line *irc.Line) {
		if strings.ToLower(line.Text()) == "leave anya-chan " || strings.ToLower(line.Text()) == "leave anya-chan" {
			if strings.ToLower(line.Nick) == "anya" {
				con.Quit(fmt.Sprintf("Bye %s", line.Nick))
			}
		}
	})

	con.HandleFunc(irc.NOTICE, func(con *irc.Conn, line *irc.Line) {
		fmt.Printf("Notice from %s\n\t%s\n", line.Ident, line.Text())
	})
}

func MatchGroups(reg *regexp.Regexp, s string) (map[string]string, error) {
	groups := make(map[string]string, 5)
	res := reg.FindStringSubmatch(s)

	if res == nil {
		return nil, errors.New(fmt.Sprintf("%s did not match regexp", s))
	}

	groupNames := reg.SubexpNames()

	for k, v := range groupNames {
		if v != "" {
			groups[v] = res[k]
		}
	}

	return groups, nil
}

func loadCodes() {
	codesFile, err := os.OpenFile("codes.gob", os.O_RDONLY, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer codesFile.Close()

	codesDec := gob.NewDecoder(codesFile)

	codes := make(map[string]FriendCode)
	err = codesDec.Decode(&codes)
	if err != nil {
		fmt.Println(err)
	}

	for nick, c := range codes {
		fc := friendCodes[nick]
		fc = c
		friendCodes[nick] = fc
	}

	fmt.Println(codes)
}

func saveCodes() {
	timeStamp := time.Now().UTC()

	timedFileName := fmt.Sprintf("codes/%02d-%02d-%v(%02d.%02d).gob", timeStamp.Day(), timeStamp.Month(), timeStamp.Year(), timeStamp.Hour(), timeStamp.Minute())
	timedFile, err := os.OpenFile(timedFileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}

	codesFile, err := os.OpenFile("codes.gob", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer timedFile.Close()
	defer codesFile.Close()

	fmt.Print("Saving friend codes... ")
	codesEnc := gob.NewEncoder(codesFile)
	timedEnc := gob.NewEncoder(timedFile)

	if codesEnc.Encode(friendCodes) != nil {
		fmt.Println("Friend codes gob error:", err)
		return
	}

	if timedEnc.Encode(friendCodes) != nil {
		fmt.Println("Friend codes gob error:", err)
		return
	}

	fmt.Println("Done")
}
