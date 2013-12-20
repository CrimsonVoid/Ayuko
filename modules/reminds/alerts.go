package reminds

import (
	"sync"
)

type alert struct {
	alert []*Message
	mut   sync.RWMutex

	quit    chan bool
	Expired chan *Message // Public chan to route messages out
	expChan chan *Message // For internal use. <-Expired <-expChan <-*Message
}

func newAlerts() map[string]*alert {
	alerts := make(map[string]*alert)

	return alerts
}

// TODO - Map locks
func startAlert(channel string) <-chan *Message {
	a, ok := alerts[channel]
	if ok {
		return a.Expired
	}

	a = &alert{
		alert: make([]*Message, 0, 5),

		quit:    make(chan bool),
		Expired: make(chan *Message, 5),
		expChan: make(chan *Message, 5),
	}

	go func() {
		for {
			select {
			case alrt := <-a.expChan:
				a.Expired <- alrt
				// TODO - Delete alrt
			case <-a.quit:
				return
			}
		}
	}()

	alerts[channel] = a

	return a.Expired
}

func stopAlert(channel string) {
	a := alerts[channel]
	a.quit <- true
	close(a.Expired)
	alerts[channel] = a
}
