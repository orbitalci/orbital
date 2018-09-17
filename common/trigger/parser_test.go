package trigger

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
)

func TestParseSpacing(t *testing.T) {
	branchUnit := &BranchCondition{acceptedBranches: []string{"master", "develop", "release.*"}, logical: Or}
	textUnit := &TextCondition{acceptedTexts: []string{"schema_changed"}, logical: CNone}
	full := &ConditionalDirective{Conditions: []Section{branchUnit, textUnit}, Logical: And}
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
		err       bool
		directive string
		parsed    *ConditionalDirective
	}{
		{
			err:       false,
			directive: "branch: fix.* and text: buildme and filepath: GisCommon",
			parsed: &ConditionalDirective{
				Logical: And,
				Conditions: []Section{
					&BranchCondition{acceptedBranches: []string{"fix.*"}, logical: CNone},
					&TextCondition{acceptedTexts: []string{"buildme"}, logical: CNone},
					&FilepathCondition{acceptedFilepaths: []string{"GisCommon"}, logical: CNone},
				},
			},
		},
		{
			err:       false,
			directive: "branch: master||develop and filepath: src/test && src/main",
			parsed: &ConditionalDirective{
				Logical: And,
				Conditions: []Section{
					&BranchCondition{acceptedBranches: []string{"master", "develop"}, logical: Or},
					&FilepathCondition{acceptedFilepaths: []string{"src/test", "src/main"}, logical: And},
				},
			},
		},
		{
			err:       false,
			directive: "branch: master or text: force_build || buildBetch",
			parsed: &ConditionalDirective{
				Logical: Or,
				Conditions: []Section{
					&BranchCondition{acceptedBranches: []string{"master"}, logical: CNone},
					&TextCondition{acceptedTexts: []string{"force_build", "buildBetch"}, logical: Or},
				},
			},
		},
		{
			err:       true,
			directive: "branch master or text: force_build || buildBetch",
		},
		{
			err:       true,
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
