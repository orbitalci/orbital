package triggerparser

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
)


func TestParseSpacing(t *testing.T) {
	branchUnit := &ConditionalSection{Ttype: Branch, Values: []string{"master", "develop", "release.*"}, Logical: Or, index: 0}
	textUnit := &ConditionalSection{Ttype: Text, Values: []string{"schema_changed"}}
	full := &ConditionalDirective{Conditions: []*ConditionalSection{branchUnit, textUnit}, Logical: And}
	cases := []string{
		"branch: master||develop|| release.* and text: schema_changed",
		"branch: master || develop||release.* and text: schema_changed",
		"branch: master|| develop ||release.* and text: schema_changed",
		"branch: master  || develop ||   release.* and text: schema_changed",
	}
	for _, tc := range cases {
		live, err := Parse(tc)
		if err != nil {
			t.Error(err)
		}
		if diff := deep.Equal(full, live); diff != nil {
			marshaled, _ := json.Marshal(live)
			t.Logf(string(marshaled))
			t.Error(diff)
		}
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		err bool
		directive string
		parsed *ConditionalDirective
	}{
		{
			err: false,
			directive: "branch: fix.* and text: buildme and filepath: GisCommon",
			parsed: &ConditionalDirective{
				Logical: And,
				Conditions: []*ConditionalSection{
					{Ttype: Branch, Values: []string{"fix.*"}, Logical: CNone},
					{Ttype: Text, Values: []string{"buildme"}, Logical: CNone},
					{Ttype: Filepath, Values: []string{"GisCommon"}, Logical: CNone},
				},
			},
		},
		{
			err: false,
			directive: "branch: master||develop and filepath: src/test && src/main",
			parsed: &ConditionalDirective{
				Logical: And,
				Conditions: []*ConditionalSection{
					{Ttype: Branch, Values: []string{"master", "develop"}, Logical: Or},
					{Ttype: Filepath, Values: []string{"src/test", "src/main"}, Logical: And},
				},
			},
		},
		{
			err: false,
			directive: "branch: master or text: force_build || buildBetch",
			parsed: &ConditionalDirective{
				Logical: Or,
				Conditions: []*ConditionalSection{
					{Ttype: Branch, Values: []string{"master"}, Logical: CNone},
					{Ttype: Text, Values: []string{"force_build", "buildBetch"}, Logical: Or},
				},
			},
		},
		{
			err: true,
			directive: "branch master or text: force_build || buildBetch",
		},
		{
			err: true,
			directive: "branch: master or text: force_build || buildBetch && hereWeGOAgain",
		},

	}
	for _, tc := range cases {
		live, err := Parse(tc.directive)
		if err != nil {
			if !tc.err {
				t.Error(err)
			}
			continue
		}
		if diff := deep.Equal(tc.parsed, live); diff != nil {
			marshaled, _ := json.Marshal(live)
			t.Logf(string(marshaled))
			t.Error(diff)
		}
	}
}