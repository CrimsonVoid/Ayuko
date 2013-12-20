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
	fcAdd = regexp.MustCompile(fmt.Sprintf(`^%vfcode add %v (?P<fcode>.*)`, modeR, systemsR))
	fcRem = regexp.MustCompile(fmt.Sprintf(`^%vfcode rem %v$`,
		modeR, `(?P<system>`+systemsL+`|\*)`))
	fcGet  = regexp.MustCompile(fmt.Sprintf(`^%vfcode %v\s?$`, modeR, nickR))
	fcList = regexp.MustCompile(fmt.Sprintf(`^%vfcode list %v$`, modeR, systemsR))
	fcHelp = regexp.MustCompile(fmt.Sprintf(`^%vfcodehelp$`, modeR))

	fCodes = NewfcManager()
)

var (
	Module *module.Module
)
