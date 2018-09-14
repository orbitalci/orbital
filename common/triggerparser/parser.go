package triggerparser

import (
	"bufio"
	"strings"
)

/*
branch: master||develop||release.* and text: schema_changed
	> branch is master or develop or release.* and commit text includes [schema_changed]
branch: fix.* and text: buildme
	> branch is fix.* and commit text includes [buildme]
branch: master||develop and filepath: src/test && src/main
	> branch is master or develop and commit file changelist includes changes in src/test and src/main
branch: master or text: force_build
	> branch is master or commit text includes [force_build]
text: schema_changed and filepath: deploy/schema/flyway||test/schema/flyway
	> commit text includes [schema_changed] and commit file changelist includes changes in EITHER deploy/schema/flyway OR test/schema/flyway

ENUM trigType
	BRANCH = 1;
	FILEPATH = 2;
	TEXT = 3;

ENUM conditional
	OR = 1;
	AND = 2;

*/

type parser struct {
	sc *bufio.Scanner
	scanned []*ConditionalDirective
	activeUnits *ConditionalDirective
	activeUnit *ConditionalSection
}

func Parse(directive string) (*ConditionalDirective, error) {
	//scanner := bufio.NewScanner(strings.NewReader(directive))
	//scanner.Split(bufio.ScanWords)

	//count := 0
	p := newParser(directive)
	err := p.scan()
	return p.activeUnits, err
}

func newParser(src string) *parser {
	scanner := bufio.NewScanner(strings.NewReader(src))
	scanner.Split(bufio.ScanWords)
	return &parser{
		sc: scanner,
	}
}

func (p *parser) scan() error {
	p.activeUnits = &ConditionalDirective{}
	for p.sc.Scan() {
		text := p.sc.Text()
		switch {
		// this means that a logical directive was just started, ie branch: or text:
		case ConvertTriggerType(text) != TNone:
			p.activeUnit = &ConditionalSection{Ttype: ConvertTriggerType(text)}

		// in between two directives, ie branch: master *and* text: help
		case ConvertConditionalWord(text) != CNone:
			aor := ConvertConditionalWord(text)
			if p.activeUnits.Logical == CNone {
				p.activeUnits.Logical = aor
			}
			// if different separators, start a new ConditionalDirective to describe the new logical behavior
			if p.activeUnits.Logical != aor {
				p.scanned = append(p.scanned, p.activeUnits)
				p.activeUnits = &ConditionalDirective{Logical: aor, Conditions: []*ConditionalSection{}}
			}
			p.activeUnits.Conditions = append(p.activeUnits.Conditions, p.activeUnit)
		case ConvertConditionalSymbol(text) != CNone:
			aor := ConvertConditionalSymbol(text)
			if p.activeUnit.Logical == CNone {
				p.activeUnit.Logical = aor
				continue
			}
			// if different separators in same active unit, bail out you aren't ready for that yet
			if aor != p.activeUnit.Logical {
				return CannotCombineSymbols()
			}
		default:
			// check to make sure the whole directive string starts with one of the action words
			if p.activeUnit == nil {
				return MustStartWithAction()
			}
			// make sure it can't be split any further
			if ContainsConditionalSymbol(text) {
				values, aor, err := SplitConditionals(text)
				if err != nil {
					return err
				}
				if p.activeUnit.Logical == CNone {
					p.activeUnit.Logical = aor
				}
				if p.activeUnit.Logical != aor {
					return CannotCombineSymbols()
				}
				for _, val := range values {
					if val != "" {
						p.activeUnit.Values = append(p.activeUnit.Values, val)
					}
				}
			} else {
				p.activeUnit.Values = append(p.activeUnit.Values, text)
			}
		}
	}
	p.activeUnits.Conditions = append(p.activeUnits.Conditions, p.activeUnit)
	return nil
}

func MustStartWithAction() *ErrNotSupported {
	return NotSupported("Directive must start with one of: 'branch:', 'text:', 'filepath:'")
}

func CannotCombineSymbols() *ErrNotSupported {
	return NotSupported("|| and && cannot yet be combined in one directive. sorry.")
}

func NotSupported(msg string) (*ErrNotSupported) {
	return &ErrNotSupported{msg:msg}
}

type ErrNotSupported struct {
	msg string
}

func (e *ErrNotSupported) Error() string {
	return e.msg
}