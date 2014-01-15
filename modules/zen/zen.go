package zen

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

var Module *module.Module

func init() {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	var err error

	Module, err = module.New(fmt.Sprintf("%v%c%v.json",
		filepath.Dir(file), filepath.Separator, base[:len(base)-len(ext)]))
	if err != nil {
		panic(err)
	}

	zenChan := make(chan string, 10)
	quit := make(chan bool)

	Module.Preconnect = func() error {
		go func() {
			for {
				// TODO - Potential block
				resp, err := http.Get("https://api.github.com/zen")
				if err != nil {
					Module.Logger.Errorln(err)
					continue
				}

				zen, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					Module.Logger.Errorln(err)
					continue
				}

				select {
				case zenChan <- string(zen):
				case <-quit:
					return
				}
				resp.Body.Close()
			}
		}()

		return nil
	}

	Module.Register(module.E_PRIVMSG, ".zen", func(line *irc.Line) {
		zen := <-zenChan

		Module.Logger.Infoln(fmt.Sprintf("%s - %s", line.Target(), zen))
		Module.Conn.Privmsg(line.Target(), zen)
	})

	Module.Disconnect = func() error {
		quit <- true

		return nil
	}
}
