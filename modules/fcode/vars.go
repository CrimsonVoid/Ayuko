package fcode

import (
	"fmt"
	"regexp"

	"github.com/crimsonvoid/irclib/module"
)

const (
	PUBLIC = "@"
	PRIV   = "."

	nickR = `(?P<nick>[\w{}\[\]^|` + "`" + `-]+)`
	modeR = `(?P<mode>[` + PRIV + PUBLIC + `])`

	systemsL = `wii|wiiu|nid|ds|3ds|psn|live|steam|bnet`
	systemsR = `(?P<system>` + systemsL + `)`
)

var (
	fcAdd = regexp.MustCompile(fmt.Sprintf(`(?i)^%vfcode add %v (?P<fcode>.*)`, modeR, systemsR))
	fcRem = regexp.MustCompile(fmt.Sprintf(`(?i)^%vfcode rem %v$`,
		modeR, `(?P<system>`+systemsL+`|\*)`))
	fcGet  = regexp.MustCompile(fmt.Sprintf(`(?i)^%vfcode %v\s?$`, modeR, nickR))
	fcList = regexp.MustCompile(fmt.Sprintf(`(?i)^%vfcode list %v$`, modeR, systemsR))
	fcHelp = regexp.MustCompile(fmt.Sprintf(`(?i)^%vfcodehelp$`, modeR))

	fCodes = NewfcManager()
)

const (
	dataDir = "./data/fcode/"
)

var (
	Module *module.Module
)
