package reminds

import (
	"fmt"
	"regexp"
)

const (
	timeFormat   = "02 Jan 2006 15:04 MST"
	pprintFormat = "02 Jan 2006 15:04"

	nickR     = `(?P<to>[\w{}\[\]^|` + "`" + `-]+)`
	timeR     = `(?P<time>\d+)`
	durationR = `(?P<duration>` +
		`s(econd(s)?)?|` +
		`m(inute(s)?)?|` +
		`h(our(s)?)?|` +
		`d(ay(s)?)?|` +
		`w(eek(s)?)?|` +
		`mo(nth(s)?)?|` +
		`y(ear(s)?)?` +
		`)`
)

var (
	remindsR = regexp.MustCompile(fmt.Sprintf("(?i)^-remind %v (in )?(%v ?%v )?(that )?%v$",
		nickR,
		timeR,
		durationR,
		`(?P<message>.*)`))

	alertsR = regexp.MustCompile(fmt.Sprintf("(?i)^-(hi(gh)?light|alert) %v (in )?(%v ?%v )?(that )?%v$",
		nickR,
		timeR,
		durationR,
		`(?P<message>.*)`))
)

var (
	reminds = NewReminds()
	alerts  = newAlerts()
)
