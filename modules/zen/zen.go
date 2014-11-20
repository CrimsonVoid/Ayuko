package zen

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/crimsonvoid/irclib/module"
	irc "github.com/fluffle/goirc/client"
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	var err error

	Module, err = module.New(fmt.Sprintf("data%[1]cconfs%[1]c%v.toml",
		filepath.Separator, base[:len(base)-len(ext)]))
	if err != nil {
		panic(err)
	}

	Module.Preconnect = func() error {
		go preConnect()

		return nil
	}

	Module.Disconnect = func() error {
		quit <- true

		return nil
	}

	Module.Register(module.E_PRIVMSG, ".zen", func(line *irc.Line) {
		var zen string

		select {
		case zen = <-zenChan:
		case <-time.After(time.Second * 10):
			zen = "Timeout while waiting for zen"
		}

		Module.Logger.Infoln(fmt.Sprintf("%s - %s", line.Target(), zen))
		Module.Conn.Privmsg(line.Target(), zen)
	})
}

func preConnect() {
	abrt := make(chan bool)

	for {
		// TODO - Potential block
		resp, err := http.Get("https://api.github.com/zen")
		if err != nil {
			Module.Logger.Errorln(err)
			continue
		}
		// sendZen() closes resp.Body

		select {
		case <-sendZen(resp, abrt): // sendZen || timeout
		case <-quit:
			select {
			case abrt <- true:
			case <-time.After(time.Second):
			}

			return
		}
	}
}

func sendZen(resp *http.Response, abrt <-chan bool) <-chan error {
	errCh := make(chan error)

	go func() {
		defer resp.Body.Close()

		if err := respStatus(resp); err != nil {
			select {
			case <-time.After(time.Duration(int(time.Minute) * timeoutMult)):
				select {
				case errCh <- err:
				case <-abrt:
					return
				}
			case <-abrt:
				return
			}

			return
		}

		zen, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			select {
			case errCh <- err:
			case <-abrt:
				return
			}

			return
		}

		select {
		case zenChan <- string(zen):
			dur := int(float64(int(time.Minute)*timeoutMult) * timeoutResetFact)
			if time.Now().Sub(lastTimeout) > time.Duration(dur) {
				if timeoutMult -= timeoutInc; timeoutMult < timeoutMin {
					timeoutMult = timeoutMin
				}
			}
		case <-abrt:
			return
		}

		select {
		case errCh <- nil:
		case <-abrt:
			return
		}
	}()

	return errCh
}

func respStatus(resp *http.Response) error {
	// 200 <= resp < 400
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("Response status %v", resp.Status)
	}

	// 'Content-Type' response header contains (text|json)
	contentType := make([]string, 0, 1)

	for key, val := range resp.Header {
		if strings.ToLower(key) == "content-type" {
			contentType = append(contentType, val...)
			break
		}
	}

	// Return true if (text|json)
	for _, cntType := range contentType {
		if strings.Contains(cntType, "text") || strings.Contains(cntType, "json") {
			return nil
		}
	}

	return errors.New("Content-Type not text|json")
}
