package reminds

import (
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/crimsonvoid/console"
)

type Message struct {
	From    string
	Message string

	Set      time.Time
	Expire   time.Time
	duration <-chan time.Time
}

type ChanNick struct {
	Channel string
	Nick    string
}

type Reminds struct {
	msgMap map[ChanNick][]*Message
	mut    sync.RWMutex
}

func NewReminds() Reminds {
	rems := Reminds{
		msgMap: make(map[ChanNick][]*Message),
	}

	return rems
}

func ParseMessage(from, duration, msg string, timeN int) (*Message, error) {
	now := time.Now().UTC()
	var expire time.Time

	switch duration {
	case "":
		expire = now.Add(time.Duration(timeN))
	case "s", "second", "seconds":
		expire = now.Add(time.Second * time.Duration(timeN))
	case "m", "minute", "minutes":
		expire = now.Add(time.Minute * time.Duration(timeN))
	case "h", "hour", "hours":
		expire = now.Add(time.Hour * time.Duration(timeN))
	case "d", "day", "days":
		expire = now.Add(time.Hour * 24 * time.Duration(timeN))
	case "mo", "month", "months": // 1 month == 30 days
		expire = now.Add(time.Hour * 24 * 30 * time.Duration(timeN))
	case "y", "year", "years": // 8765.81 hours in a year according to Google
		expire = now.Add(time.Hour * 8766 * time.Duration(timeN))
	default:
		// Module.Logger.Errorf("Error parsing duration: `%v`", duration)

		return nil, fmt.Errorf("Error parsing duration: `%v`", duration)
	}

	remind := &Message{
		From:    from,
		Message: msg,

		Set:      now,
		Expire:   expire,
		duration: time.After(expire.Sub(now)),
	}

	return remind, nil
}

func (self *Reminds) Start() error {
	return self.Load("reminds.gob")
}

func (self *Reminds) Exit() error {
	timeStamp := time.Now().UTC()

	timedPath := fmt.Sprintf("reminds/%v-%v", timeStamp.Year(), timeStamp.Month())
	if err := os.MkdirAll(timedPath, 755); err != nil {
		return err
	}
	timedFileName := fmt.Sprintf("%v/%02v_(%02v.%02v).gob",
		timedPath, timeStamp.Day(), timeStamp.Hour(), timeStamp.Minute())

	if err := self.Save("reminds.gob"); err != nil {
		return err
	}
	if err := self.Save(timedFileName); err != nil {
		return err
	}

	return nil
}

func (self *Reminds) Save(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	self.mut.RLock()
	defer self.mut.RUnlock()

	codesEnc := gob.NewEncoder(file)
	err = codesEnc.Encode(self.msgMap)

	return err
}

func (self *Reminds) Load(fileName string) error {
	file, err := os.Open(fileName)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	self.mut.Lock()
	defer self.mut.Unlock()

	codesDec := gob.NewDecoder(file)
	err = codesDec.Decode(&self.msgMap)

	for _, msgs := range self.msgMap {
		for _, msg := range msgs {
			msg.duration = time.After(msg.Expire.Sub(msg.Set))
		}
	}

	return err
}

func (self *Reminds) Add(key ChanNick, msg *Message) {
	if msg.duration == nil {
		msg.duration = time.After(msg.Expire.Sub(msg.Set))
	}

	self.mut.Lock()
	defer self.mut.Unlock()

	msgList := self.msgMap[key]
	msgList = append(msgList, msg)
	self.msgMap[key] = msgList
}

func (self *Reminds) Remove(key ChanNick, indices ...int) {
	self.mut.Lock()
	defer self.mut.Unlock()

	sort.Ints(indices)

	msgList, ok := self.msgMap[key]
	if !ok {
		return
	}

	elemRem := 0
	for _, i := range indices {
		i -= elemRem
		if i >= len(msgList) {
			continue
		}

		copy(msgList[i:], msgList[i+1:])
		msgList = msgList[:len(msgList)-1]
		elemRem++
	}

	if len(msgList) == 0 {
		delete(self.msgMap, key)
	} else {
		self.msgMap[key] = msgList
	}

}

func (self *Reminds) GetExpired(key ChanNick) []*Message {
	self.mut.RLock()
	defer self.mut.RUnlock()

	expiredList := make([]*Message, 0, 5)
	expiredInd := make([]int, 0, 5)

	msgList, ok := self.msgMap[key]
	if !ok {
		return expiredList
	}

	for i, rem := range msgList {
		select {
		case <-rem.duration:
			expiredList = append(expiredList, rem)
			expiredInd = append(expiredInd, i)
		default:
		}
	}

	go self.Remove(key, expiredInd...)

	return expiredList
}

func (self *Reminds) String() string {
	self.mut.RLock()
	defer self.mut.RUnlock()

	now := time.Now().UTC()
	out := ""
	remMap := make(map[string]map[string][]string)

	for chnNick, msgList := range self.msgMap {
		nickMap, ok := remMap[chnNick.Channel]
		if !ok {
			nickMap = make(map[string][]string)
		}

		nickList, ok := nickMap[chnNick.Nick]
		if !ok {
			nickList = make([]string, 0, 5)
		}

		for _, msg := range msgList {
			// Green - Expired
			// Red   - Active

			statusColor := console.C_FgGreen
			if now.Before(msg.Expire) {
				statusColor = console.C_FgRed
			}

			nickList = append(nickList, fmt.Sprintf("%v%v%v - %v: %v",
				statusColor, msg.Expire.Format(pprintFormat), console.C_Reset,
				msg.From, msg.Message))
		}

		nickMap[chnNick.Nick] = nickList
		remMap[chnNick.Channel] = nickMap
	}

	for chn, nickMap := range remMap {
		out += fmt.Sprintf("%v\n", chn)

		for nick, rems := range nickMap {
			out += fmt.Sprintf("  %v\n    %v\n", nick, strings.Join(rems, "\n    "))
		}

		out += "\n"
	}

	return out
}

func (self *Reminds) Copy() map[ChanNick][]Message {
	self.mut.RLock()
	defer self.mut.RUnlock()

	rems := make(map[ChanNick][]Message)

	for nick, msgList := range self.msgMap {
		msgs := make([]Message, 0, len(msgList))

		for _, m := range msgList {
			msgs = append(msgs, Message{
				From:    m.From,
				Message: m.Message,

				Set:      m.Set,
				Expire:   m.Expire,
				duration: nil,
			})
		}

		rems[nick] = msgs
	}

	return rems
}
