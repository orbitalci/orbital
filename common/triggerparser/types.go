package triggerparser

import "strings"

//go:generate stringer -type=TriggerType,Conditional



type TriggerType int

const (
	TNone TriggerType = iota
	Branch
	Filepath
	Text
)

func ConvertTriggerType(str string) TriggerType {
	switch strings.ToLower(str) {
	case "branch:":   return Branch
	case "filepath:": return Filepath
	case "text:":     return Text
	default:          return TNone
	}
}


type Conditional int

const (
	CNone Conditional = iota
	Or
	And
)

func ConvertConditionalWord(str string) Conditional {
	switch strings.ToUpper(str) {
	case "OR":  return Or
	case "AND": return And
	default:    return CNone
	}
}

func ConvertConditionalSymbol(str string) Conditional {
	switch strings.TrimSpace(str) {
	case "||": return Or
	case "&&": return And
	default:   return CNone
	}
}

func ContainsConditionalSymbol(str string) bool {
	if !strings.Contains(str, "||") && !strings.Contains(str, "&&") {
		return false
	}
	return true
}

func SplitConditionals(str string) ([]string, Conditional, error) {
	if strings.Contains(str, "||") && strings.Contains(str, "&&") {
		return nil, CNone, CannotCombineSymbols()
	}

	values := strings.Split(str, "||")
	if len(values) > 1 {
		return values, Or, nil
	}
	values = strings.Split(str, "&&")
	if len(values) > 1 {
		return values, And, nil
	}
	return values, CNone, nil
}

type ConditionalSection struct {
	Ttype   TriggerType
	Values  []string
	Logical Conditional
	index   int
}

type ConditionalDirective struct {
	Conditions  []*ConditionalSection
	Logical     Conditional
}
