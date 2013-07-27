package irclibrary

import (
	"encoding/json"
	"errors"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"io/ioutil"
	"time"
)

type ServerInfo struct {
	Nick, Ident, Name string

	Pass, Server string
	SSL          bool
	Port         int

	Version, QuitMessage string
	PingFreq, SplitLen   int
	Flood, Tracking      bool

	Botinfo BotInfo
}

type BotInfo struct {
	Chans  []string
	Access access
}

func (serverInfo *ServerInfo) configServer() (*irc.Config, error) {
	// Check there is enough info to set up a server
	switch {
	case serverInfo.Nick == "":
		return nil, errors.New("Specify a Nick in the config file")
	case serverInfo.Server == "":
		return nil, errors.New("Specify a Server in the config file")
	}

	if serverInfo.Ident == "" {
		serverInfo.Ident = serverInfo.Nick
	}
	if serverInfo.Name == "" {
		serverInfo.Name = serverInfo.Nick
	}

	cfg := irc.NewConfig(serverInfo.Nick, serverInfo.Ident, serverInfo.Name)
	cfg.Pass = serverInfo.Pass
	// serverInfo.SSL == True probably wont work without extra configuration
	cfg.SSL = serverInfo.SSL
	cfg.Server = fmt.Sprintf("%s:%d", serverInfo.Server, serverInfo.Port)

	if serverInfo.Version != "" {
		cfg.Version = serverInfo.Version
	}
	if serverInfo.QuitMessage != "" {
		cfg.QuitMessage = serverInfo.QuitMessage
	}
	if serverInfo.PingFreq > 0 {
		cfg.PingFreq = time.Duration(serverInfo.PingFreq) * time.Second
	}
	if serverInfo.SplitLen > 0 {
		cfg.SplitLen = serverInfo.SplitLen
	}
	cfg.Flood = serverInfo.Flood

	return cfg, nil
}

func New(fileName string) (con *irc.Conn, botInfo *BotInfo, err error) {
	cfg, serverInfo, err := NewConfig(fileName)
	if err != nil {
		return
	}

	con = irc.Client(cfg)
	if serverInfo.Tracking == true {
		con.EnableStateTracking()
	}

	botInfo = &serverInfo.Botinfo
	return
}

func NewConfig(fileName string) (cfg *irc.Config, serverInfo *ServerInfo, err error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	serverInfo = new(ServerInfo)
	err = json.Unmarshal(file, serverInfo)
	if err != nil {
		return
	}

	cfg, err = serverInfo.configServer()
	return
}
