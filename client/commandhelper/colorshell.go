package commandhelper

import (
	"runtime"
)

const (
	gitIdentifier = `\[\033[`
	normal = gitIdentifier + `0;`
	bold = gitIdentifier + `1;`
	Black = `30m\]`
	Red = `31m\]`
	Green = `32m\]`
	Yellow = `33m\]`
	Blue = `34m\]`
	Purple = `35m\]`
	Cyan = `36m\]`
	White = `37m\]`
	Reset = `\[\033[0m\]`
)

type ColorDefs struct {
	Failed      	 string
	Queued 			 string
	Passed           string
	Done  			 string
	Running          string
	Reset      		 string
	Normal			 string
	Info			 string
	Error            string
	Warning          string
	Status           string
	TableHeaderColor string
}

func Default() *ColorDefs {
	if runtime.GOOS != "windows" {
		return &ColorDefs{
			Failed: bold+Red,
			Queued: normal+Black,
			Passed: normal+Green,
			Done: normal+Black,
			Info: normal+Blue,
			Warning: normal+Yellow,
			Normal: normal+Black,
			Reset: Reset,
			Running: normal+Purple,
			Error: bold+Red,
		}
	} else {
		return &ColorDefs{}
	}
}