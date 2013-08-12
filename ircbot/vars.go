package ircbot

const (
	PUBLIC = "@"
	PRIV   = "."
)

const (
	nickR = "(?P<nick>[\\w{}\\[\\]^|`-]+)"
	modeR = `(?P<mode>[\` + PUBLIC + PRIV + `])`
)
