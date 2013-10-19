package ircbot

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type regMap struct {
	re *regexp.Regexp
	f  func(string, map[string]string) (string, error)
}

const (
	systemsL = `wii|wiiu|nid|ds|3ds|psn|live|steam|bnet`
	systemsR = `(?P<system>` + systemsL + `)`
)

var (
	fcAdd  = regexp.MustCompile(fmt.Sprintf(`^%sfcode add %s (?P<fcode>.*)`, modeR, systemsR))
	fcRem  = regexp.MustCompile(fmt.Sprintf(`^%sfcode rem (%s|\*)$`, modeR, systemsR))
	fcGet  = regexp.MustCompile(fmt.Sprintf(`^%sfcode %s\s?$`, modeR, nickR))
	fcList = regexp.MustCompile(fmt.Sprintf(`^%sfcode list %s$`, modeR, systemsR))
	fcHelp = regexp.MustCompile(fmt.Sprintf(`^%sfcodehelp$`, modeR))

	matches = []regMap{
		{fcAdd, fnAdd},
		{fcRem, fnRem},
		{fcGet, fnGet},
		{fcList, fnList},
		{fcHelp, fnHelp},
	}
)

func MatchFC(nick, line string) (map[string]string, string, error) {
	for _, m := range matches {
		groups, err := matchGroups(m.re, line)
		if err != nil {
			continue
		}

		msg, err := m.f(nick, groups)
		if err != nil {
			return nil, "", err
		}
		log.Println(m.re, "\t", groups, "\t", line)

		return groups, msg, nil
	}

	return nil, "", fmt.Errorf("No matches found for %s", line)
}

func fnAdd(nick string, groups map[string]string) (string, error) {
	groups["mode"] = PRIV

	nick = strings.ToLower(nick)
	var (
		save *string
		re   *regexp.Regexp
	)

	fcMut.Lock()
	defer fcMut.Unlock()

	fCode := friendCodes[nick]
	if fCode == nil {
		fCode = new(friendCode)
	}

	switch groups["system"] {
	case "wii":
		re = regexp.MustCompile(`\d{4}-\d{4}-\d{4}-\d{4}$`)
		save = &fCode.Wii
	case "wiiu", "nid":
		re = regexp.MustCompile(`^.{6,16}\s?$`)
		save = &fCode.Wiiu
	case "ds":
		re = regexp.MustCompile(`^\d{4}-\d{4}-\d{4}\s?$`)
		save = &fCode.Ds
	case "3ds":
		re = regexp.MustCompile(`^\d{4}-\d{4}-\d{4}\s?$`)
		save = &fCode.Ds3
	case "psn":
		re = regexp.MustCompile(`^.{6,16}\s?$`)
		save = &fCode.Psn
	case "live":
		re = regexp.MustCompile(`^.{6,15}\s?$`)
		save = &fCode.Psn
	case "steam":
		// TODO|fcode - steam nick restrictions
		re = regexp.MustCompile(`^.*`)
		save = &fCode.Steam
	case "bnet":
		re = regexp.MustCompile(`^.*\d{3}\s?$`)
		save = &fCode.Bnet
	}

	res := re.FindStringSubmatch(groups["fcode"])
	if res == nil {
		return fmt.Sprintf("%s is not a valid %s code", groups["fcode"], groups["system"]), nil
	} else {
		*save = groups["fcode"]
		friendCodes[nick] = fCode

		return fmt.Sprintf("Saved %s friend code %s", groups["system"], groups["fcode"]), nil
	}
}

func fnRem(nick string, groups map[string]string) (string, error) {
	groups["mode"] = PRIV

	nick = strings.ToLower(nick)

	fcMut.Lock()
	defer fcMut.Unlock()

	fc, ok := friendCodes[nick]
	if !ok {
		return "You have not saved and user names", nil
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
	default: // .fc rem *
		delete(friendCodes, nick)
		return "Removed you from the database", nil
	}
	friendCodes[nick] = fc

	return fmt.Sprintf("Removed nick for %s", groups["system"]), nil
}

func fnGet(nick string, groups map[string]string) (string, error) {
	fcMut.RLock()
	defer fcMut.RUnlock()

	fc, ok := friendCodes[strings.ToLower(groups["nick"])]
	if !ok {
		return fmt.Sprintf("%s has not saved any friend codes", groups["nick"]), nil
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

	return fmt.Sprintf("%s's friend codess are %s", groups["nick"], codes), nil
}

func fnList(nick string, groups map[string]string) (string, error) {
	codeList := ""

	fcMut.RLock()
	defer fcMut.RUnlock()

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

	return codeList, nil
}

func fnHelp(nick string, groups map[string]string) (string, error) {
	groups["mode"] = PRIV

	return fmt.Sprintf("Save and retrieve gaming identities."+
		" Syntax: [@.]fcode [add|rem|list] [%s] (code) || .fc nick", systemsL), nil
}
