package trigger

import (
	"fmt"
	//"testing"
	"testing"

	"github.com/shankj3/ocelot/models/pb"
)

/*
- branch: master||develop|| release.* and text: schema_changed
	changelist:
	- filepaths:
		- src/main/java/com/o8/kontrol.java
	- text:
		- test1
		- i did a thing
		- [schema_changed]
	- branch: release/update_schema
- branch: fix.* and text: buildme and filepath: GisCommon
	changelist:
	- filepaths:
		- GisCommon/src/main/java/com/...
	- text:
		- i did a big thing
		- [buildme]
	- branch: fix/memory_leak
- branch: master||develop and filepath: src/test && src/main
	changelist:
	- filepaths:
		- src/main/java/com/hilhil.java
		- src/test/java/com/hilhil.java
	- branch: develop
	- text:
		- [sizzurp]
*/

func TestConditionalDirective_Passes(t *testing.T) {
	td := &pb.ChangesetData{
		FilesChanged: []string{
			"src/main/java/com/o8/controller.java",
			"src/main/test/com/o8/controllerTest.java",
			"deploy/helm/upgrade.yaml",
			"sql/__SQL/__SQL_1.sql",
			"hummuna",
		},
		CommitTexts: []string{
			"i changed a controller",
			"i updated the schema [schema_changed]",
			"i added a controller test",
			"i added to deployment",
		},
		Branch: "fix/thisiskewl",
	}

	var directives = []struct{
		cd         *ConditionalDirective
		shouldPass bool
	}{
		{
			cd: &ConditionalDirective{
					Conditions: []Section{
						&BranchCondition{acceptedBranches: []string{`fix\/.*`},
					},
				},
			},
			shouldPass: true,
		},
		{
			cd: &ConditionalDirective{
				Conditions: []Section{
					&BranchCondition{acceptedBranches: []string{`fix\/.*`, `master`, `develop`}, logical: Or},
					},
				},
			shouldPass: true,
		},
		{
			cd: &ConditionalDirective{
				Conditions: []Section{
					&FilepathCondition{acceptedFilepaths: []string{"src/main/test", "src/main/java"}, logical: And},
				},
			},
			shouldPass: true,
		},
		{
			cd: &ConditionalDirective{
				Conditions: []Section{
					&FilepathCondition{acceptedFilepaths: []string{"src/main/test", "src/main/java"}, logical: And},
					&BranchCondition{acceptedBranches: []string{`fix\/.*`, `master`, `develop`}, logical: Or},
				},
			},
			shouldPass: true,
		},
		{
			cd: &ConditionalDirective{
				Conditions: []Section{
					&FilepathCondition{acceptedFilepaths: []string{"src/main/test", "src/main/java"}, logical: And},
					&BranchCondition{acceptedBranches: []string{`fix\/.*`, `master`, `develop`}, logical: Or},
					&TextCondition{acceptedTexts: []string{"schema_changed"}},
				},
			},
			shouldPass: true,
		},
		{
			cd: &ConditionalDirective{
				Conditions: []Section{
					&FilepathCondition{acceptedFilepaths: []string{"src/main/test", "src/main/java"}, logical: And},
					&BranchCondition{acceptedBranches: []string{`fix\/.*`, `master`, `develop`}, logical: Or},
					&TextCondition{acceptedTexts: []string{"schema_changed", "trigger_build"}, logical: And},
				},
				Logical: And,
			},
			shouldPass: false,
		},
		{
			cd: &ConditionalDirective{
				Conditions: []Section{
					&FilepathCondition{acceptedFilepaths: []string{"src/main/test", "src/main/java"}, logical: And},
					&BranchCondition{acceptedBranches: []string{`fix\/.*`, `master`, `develop`}, logical: Or},
					&TextCondition{acceptedTexts: []string{"schema_changed", "trigger_build"}, logical: Or},
				},
				Logical: And,
			},
			shouldPass: true,
		},

	}

	for ind, tc := range directives {
		t.Run(fmt.Sprintf("%d", ind), func(t *testing.T){
			if didPass := tc.cd.IsFulfilled(td); didPass != tc.shouldPass {
				t.Errorf("should pass is %v, got %v", tc.shouldPass, didPass)
			}
		})
	}

}