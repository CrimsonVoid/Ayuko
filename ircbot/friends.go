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
	codesFile, err := os.Open("codes.gob")

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	defer codesFile.Close()

	codesDec := gob.NewDecoder(codesFile)
	fcMut.Lock()
	err = codesDec.Decode(&friendCodes)
	fcMut.Unlock()

	return err
}

func Exit() error {
	timeStamp := time.Now().UTC()

	timedPath := fmt.Sprintf("codes/%d-%s", timeStamp.Year(), timeStamp.Month())
	if err := os.MkdirAll(timedPath, 755); err != nil {
		return err
	}

	timedFileName := fmt.Sprintf("%s/%02d_(%02d.%02d).gob", timedPath, timeStamp.Day(),
		timeStamp.Hour(), timeStamp.Minute())
	timedFile, err := os.Create(timedFileName)
	if err != nil {
		return err
	}

	codesFile, err := os.Create("codes.gob")
	if err != nil {
		return err
	}

	defer timedFile.Close()
	defer codesFile.Close()

	log.Print("Saving friend codes... ")

	fcMut.RLock()
	defer fcMut.RUnlock()

	codesEnc := gob.NewEncoder(codesFile)
	if err = codesEnc.Encode(friendCodes); err != nil {
		return err
	}

	codesEnc = gob.NewEncoder(timedFile)
	if err = codesEnc.Encode(friendCodes); err != nil {
		return err
	}

	log.Println("Done")

	return nil
}

// Logs friendCodes
func Print() {
	fcMut.RLock()
	for nick, codes := range friendCodes {
		log.Printf("%-25s%#v\n", nick, codes)
	}
	fcMut.RUnlock()
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
