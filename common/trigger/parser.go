package trigger

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
*/

type parser struct {
	sc            *bufio.Scanner
	scanned       []*ConditionalDirective
	directive     *ConditionalDirective
	activeSection Section
}

func Parse(directive string) (*ConditionalDirective, error) {
	//scanner := bufio.NewScanner(strings.NewReader(directive))
	//scanner.Split(bufio.ScanWords)

	//count := 0
	p := newParser(directive)
	err := p.scan()
	return p.directive, err
}

func newParser(src string) *parser {
	scanner := bufio.NewScanner(strings.NewReader(src))
	scanner.Split(bufio.ScanWords)
	return &parser{
		sc: scanner,
	}
}

func (p *parser) scan() error {
	p.directive = &ConditionalDirective{}
	for p.sc.Scan() {
		text := p.sc.Text()
		switch {
		// this means that a logical directive was just started, ie branch: or text:
		case convertTriggerType(text) != TNone:
			p.activeSection = convertTriggerType(text).Spawn()

		// in between two directives, ie branch: master *and* text: help
		case convertConditionalWord(text) != CNone:
			aor := convertConditionalWord(text)
			if p.directive.Logical == CNone {
				p.directive.Logical = aor
			}
			// if different separators, start a new ConditionalDirective to describe the new logical behavior
			if p.directive.Logical != aor {
				p.scanned = append(p.scanned, p.directive)
				p.directive = &ConditionalDirective{Logical: aor}
			}
			p.directive.Conditions = append(p.directive.Conditions, p.activeSection)
		case convertConditionalSymbol(text) != CNone:
			aor := convertConditionalSymbol(text)
			if p.activeSection.GetLogical() == CNone {
				p.activeSection.SetLogical(aor)
				continue
			}
			// if different separators in same active unit, bail out you aren't ready for that yet
			if aor != p.activeSection.GetLogical() {
				return CannotCombineSymbols()
			}
		default:
			// check to make sure the whole directive string starts with one of the action words
			if p.activeSection == nil {
				return MustStartWithAction()
			}
			// make sure it can't be split any further
			if containsConditionalSymbol(text) {
				values, aor, err := splitConditionals(text)
				if err != nil {
					return err
				}
				if p.activeSection.GetLogical() == CNone {
					p.activeSection.SetLogical(aor)
				}
				if p.activeSection.GetLogical() != aor {
					return CannotCombineSymbols()
				}
				for _, val := range values {
					if val != "" {
						p.activeSection.AddConditionValue(val)
					}
				}
			} else {
				p.activeSection.AddConditionValue(text)
			}
		}
	}
	p.directive.Conditions = append(p.directive.Conditions, p.activeSection)
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