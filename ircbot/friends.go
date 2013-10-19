package ircbot

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
)

type friendCode struct {
	Wii, Wiiu string
	Ds, Ds3   string
	Live, Psn string
	Steam     string
	Bnet      string
}

var (
	friendCodes = make(map[string]*friendCode)
	fcMut       = new(sync.RWMutex)
)

func Start() error {
	return Load("codes.gob")
}

func Exit() error {
	timeStamp := time.Now().UTC()

	timedPath := fmt.Sprintf("codes/%d-%s", timeStamp.Year(), timeStamp.Month())
	if err := os.MkdirAll(timedPath, 755); err != nil {
		return err
	}
	timedFileName := fmt.Sprintf("%s/%02d_(%02d.%02d).gob", timedPath, timeStamp.Day(),
		timeStamp.Hour(), timeStamp.Minute())

	log.Print("Saving friend codes... ")

	if err := Save("codes.gob"); err != nil {
		return err
	}
	if err := Save(timedFileName); err != nil {
		return err
	}

	log.Println("Done")

	return nil
}

func Load(fileName string) error {
	file, err := os.Open(fileName)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	codesDec := gob.NewDecoder(file)
	fcMut.Lock()
	defer fcMut.Unlock()

	return codesDec.Decode(&friendCodes)
}

func Save(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	fcMut.RLock()
	defer fcMut.RUnlock()

	codesEnc := gob.NewEncoder(file)

	return codesEnc.Encode(friendCodes)
}

// Logs friendCodes
func Print() {
	fcMut.RLock()
	defer fcMut.RUnlock()

	for nick, codes := range friendCodes {
		log.Printf("%-25s%#v\n", nick, codes)
	}
}

func matchGroups(reg *regexp.Regexp, s string) (map[string]string, error) {
	groups := make(map[string]string)
	res := reg.FindStringSubmatch(s)
	if res == nil {
		return nil, fmt.Errorf("%s did not match regexp", s)
	}

	groupNames := reg.SubexpNames()
	for k, v := range groupNames {
		if v != "" {
			groups[v] = res[k]
		}
	}

	return groups, nil
}
