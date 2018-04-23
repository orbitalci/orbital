package commandhelper

import (
	//"runtime"

	"runtime"

	"github.com/fatih/color"
)

type SprintfFunc func(format string, a ...interface{}) string
type Color struct {
	*color.Color
}

type ColorDefs struct {
	Failed           *Color
	Queued           *Color
	Passed           *Color
	Done             *Color
	Running          *Color
	Normal           *Color
	Info             *Color
	Error            *Color
	Warning          *Color
	Status           *Color
	TableHeaderColor *Color
	NoColor 		 bool
}

func Default(noColor bool) *ColorDefs {
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	black := color.New(color.FgBlack)
	green := color.New(color.FgGreen)
	purple := color.New(color.FgMagenta)
	//blue := color.New(color.FgBlue).SprintfFunc()
	cyan := color.New(color.FgCyan)
	theme := &ColorDefs{
		Failed: &Color{red},
		Queued: &Color{black},
		Passed: &Color{green},
		Done: &Color{black},
		Info: &Color{cyan},
		Warning: &Color{yellow},
		Normal: &Color{black},
		Running: &Color{purple},
		Error: &Color{red},
	}
	if runtime.GOOS == "windows" || noColor {
		color.NoColor = true
		theme.NoColor = true
	}
	return theme
}
