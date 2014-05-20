package url

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/crimsonvoid/irclib/module"
)

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

	registerCommands()
}