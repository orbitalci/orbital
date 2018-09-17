package trigger

import "strings"

//go:generate stringer -type=TriggerType,Conditional



type TriggerType int


const (
	TNone TriggerType = iota
	Branch
	Filepath
	Text
)

func (t TriggerType) Spawn() Section {
	switch t {
	case Branch:
		return &BranchCondition{}
	case Filepath:
		return &FilepathCondition{}
	case Text:
		return &TextCondition{}
	default:
		panic("shouldn't happen")
	}
}

func convertTriggerType(str string) TriggerType {
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

func convertConditionalWord(str string) Conditional {
	switch strings.ToUpper(str) {
	case "OR":  return Or
	case "AND": return And
	default:    return CNone
	}
}

func convertConditionalSymbol(str string) Conditional {
	switch strings.TrimSpace(str) {
	case "||": return Or
	case "&&": return And
	default:   return CNone
	}
}

func containsConditionalSymbol(str string) bool {
	if !strings.Contains(str, "||") && !strings.Contains(str, "&&") {
		return false
	}
	return true
}

func splitConditionals(str string) ([]string, Conditional, error) {
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
