package fcode

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type friendCode struct {
	Nid       string
	Wii, Wiiu string
	Ds, Ds3   string
	Live, Psn string
	Steam     string
	Bnet      string
}

type fcManager struct {
	friendCodes map[string]*friendCode
	mut         sync.RWMutex
}

func NewfcManager() *fcManager {
	return &fcManager{
		friendCodes: make(map[string]*friendCode),
	}
}

func (self *fcManager) Start() error {
	return self.Load("codes.gob")
}

func (self *fcManager) Exit() error {
	timeStamp := time.Now().UTC()

	timedPath := fmt.Sprintf("%v-%02[2]d %[2]v", timeStamp.Year(), timeStamp.Month())
	if err := os.MkdirAll(dataDir+timedPath, 755); err != nil {
		return err
	}
	timedFileName := fmt.Sprintf("%v/%02v_(%02v.%02v).gob",
		timedPath, timeStamp.Day(), timeStamp.Hour(), timeStamp.Minute())

	if err := self.Save("codes.gob"); err != nil {
		return err
	}
	if err := self.Save(timedFileName); err != nil {
		return err
	}

	return nil
}

func (self *fcManager) Load(fileName string) error {
	file, err := os.Open(dataDir + fileName)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	codesDec := gob.NewDecoder(file)

	self.mut.Lock()
	defer self.mut.Unlock()

	return codesDec.Decode(&self.friendCodes)
}

func (self *fcManager) Save(fileName string) error {
	file, err := os.Create(dataDir + fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	self.mut.RLock()
	defer self.mut.RUnlock()

	codesEnc := gob.NewEncoder(file)

	return codesEnc.Encode(self.friendCodes)
}

func (self *fcManager) String() string {
	self.mut.RLock()
	defer self.mut.RUnlock()

	out := ""
	maxNickLen := 0

	for nick, _ := range self.friendCodes {
		if nickLen := len(nick); nickLen > maxNickLen {
			maxNickLen = nickLen
		}
	}

	outFmt := "%" + strconv.Itoa(maxNickLen+2) + "v: %v\n"

	for nick, fCode := range self.friendCodes {
		if len(nick) > maxNickLen {
			maxNickLen = len(nick)
		}

		out += fmt.Sprintf(outFmt, nick, fCode.String())
	}

	return out
}

func (self *friendCode) String() string {
	out := ""

	if self.Nid != "" {
		out += fmt.Sprintf("(Nid: %v) ", self.Nid)
	}
	if self.Wii != "" {
		out += fmt.Sprintf("(Wii: %v) ", self.Wii)
	}
	if self.Wiiu != "" {
		out += fmt.Sprintf("(WiiU: %v) ", self.Wiiu)
	}
	if self.Ds != "" {
		out += fmt.Sprintf("(DS: %v) ", self.Ds)
	}
	if self.Ds3 != "" {
		out += fmt.Sprintf("(3DS: %v) ", self.Ds3)
	}
	if self.Live != "" {
		out += fmt.Sprintf("(Live: %v) ", self.Live)
	}
	if self.Psn != "" {
		out += fmt.Sprintf("(PSN: %v) ", self.Psn)
	}
	if self.Steam != "" {
		out += fmt.Sprintf("(Steam: %v) ", self.Steam)
	}
	if self.Bnet != "" {
		out += fmt.Sprintf("(Bnet: %v) ", self.Bnet)
	}

	return out
}

func (self *fcManager) Strings() map[string]string {
	self.mut.RLock()
	defer self.mut.RUnlock()

	fcMap := make(map[string]string, len(self.friendCodes))

	for nick, fCode := range self.friendCodes {
		fcMap[nick] = fCode.String()
	}

	return fcMap
}

func (self *fcManager) Add(nick, system, code string) bool {
	self.mut.Lock()
	defer self.mut.Unlock()

	var (
		save *string
		re   *regexp.Regexp
	)

	fCode, ok := self.friendCodes[nick]
	if !ok {
		fCode = new(friendCode)
	}

	switch system {
	case "wii":
		re = regexp.MustCompile(`\d{4}-\d{4}-\d{4}-\d{4}$`)
		save = &fCode.Wii
	case "wiiu":
		re = regexp.MustCompile(`^.{6,16}\s?$`)
		save = &fCode.Wiiu
	case "nid":
		re = regexp.MustCompile(`^.{6,16}\s?$`)
		save = &fCode.Nid
	case "ds":
		re = regexp.MustCompile(`^\d{4}-\d{4}-\d{4}\s?$`)
		save = &fCode.Ds
	case "3ds":
		re = regexp.MustCompile(`^\d{4}-\d{4}-\d{4}\s?$`)
		save = &fCode.Ds3
	case "psn":
		re = regexp.MustCompile(`^.{6,16}\s?$`)
		save = &fCode.Psn
	case "live":
		re = regexp.MustCompile(`^.{6,15}\s?$`)
		save = &fCode.Live
	case "steam":
		// TODO|fcode - steam nick restrictions
		re = regexp.MustCompile(`^.*`)
		save = &fCode.Steam
	case "bnet":
		re = regexp.MustCompile(`^.*#\d{3,4}\s?$`)
		save = &fCode.Bnet
	default:
		return false
	}

	if re.FindStringSubmatch(code) == nil {
		return false
	}

	*save = code
	self.friendCodes[nick] = fCode

	return true
}

func (self *fcManager) Remove(nick, system string) error {
	self.mut.Lock()
	defer self.mut.Unlock()

	fCode, ok := self.friendCodes[nick]
	if !ok {
		return errors.New("Nick not in database")
	}

	switch system {
	case "wii":
		fCode.Wii = ""
	case "wiiu":
		fCode.Wiiu = ""
	case "nid":
		fCode.Nid = ""
	case "ds":
		fCode.Ds = ""
	case "3ds":
		fCode.Ds3 = ""
	case "psn":
		fCode.Psn = ""
	case "live":
		fCode.Live = ""
	case "steam":
		fCode.Steam = ""
	case "bnet":
		fCode.Bnet = ""
	case "*":
		delete(self.friendCodes, nick)
		return nil
	default:
		return errors.New("Unknown system")
	}
	self.friendCodes[nick] = fCode

	return nil
}

func (self *fcManager) GetUser(nick string) (map[string]string, error) {
	self.mut.RLock()
	defer self.mut.RUnlock()

	fCode, ok := self.friendCodes[nick]
	if !ok {
		return nil, errors.New("Nick not in database")
	}

	codes := make(map[string]string, 8)
	if c := fCode.Wii; c != "" {
		codes["Wii"] = c
	}
	if c := fCode.Wiiu; c != "" {
		codes["WiiU"] = c
	}
	if c := fCode.Nid; c != "" {
		codes["NID"] = c
	}
	if c := fCode.Ds; c != "" {
		codes["DS"] = c
	}
	if c := fCode.Ds3; c != "" {
		codes["3DS"] = c
	}
	if c := fCode.Live; c != "" {
		codes["Live"] = c
	}
	if c := fCode.Psn; c != "" {
		codes["PSN"] = c
	}
	if c := fCode.Steam; c != "" {
		codes["Steam"] = c
	}
	if c := fCode.Bnet; c != "" {
		codes["Bnet"] = c
	}

	return codes, nil
}

func (self *fcManager) GetSystem(system string) map[string]string {
	self.mut.RLock()
	defer self.mut.RUnlock()

	codeList := make(map[string]string, 8)

	for nick, fCode := range self.friendCodes {
		code := ""

		switch system {
		case "wii":
			code = fCode.Wii
		case "wiiu":
			code = fCode.Wiiu
		case "nid":
			code = fCode.Nid
		case "ds":
			code = fCode.Ds
		case "3ds":
			code = fCode.Ds3
		case "psn":
			code = fCode.Psn
		case "live":
			code = fCode.Live
		case "steam":
			code = fCode.Steam
		case "bnet":
			code = fCode.Bnet
		}

		if code != "" {
			codeList[nick] = code
		}
	}

	return codeList
}

func FcHelp() string {
	return fmt.Sprintf("Save and retrieve gaming identities. "+
		"Syntax: [%v%v]fcode [add|rem|list] [%v] (code) || .fcode nick",
		PUBLIC, PRIV, systemsL)
}
