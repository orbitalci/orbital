package build

import (
	"fmt"
	"testing"

	"github.com/shankj3/go-til/test"
)

var branchTests = []struct{
	branch string
	allowedBranches []string
	regexOk bool
}{
	{"lets-go-to-the-movies", []string{"lets-go", "movietime", "lets-go-torp-the-movies"}, false},
	{"here/we/go", []string{"gotime", `here\/.*\/go`}, true},
	{"feature_1", []string{"feature_1", "master"}, true},
	{"feature.1", []string{"feature.1", "master"}, true},
	{"v1.0.0", []string{`v[0-9]+\.[0-9]+\.[0-9a-zA-Z]+`}, true},
	{"rc-v1.0.0", []string{`rc-v.*`}, true},
	{"rc-v1.0.0", []string{`ALL`}, true},

}

func Test_BranchRegexOk(t *testing.T) {
	//BranchRegexOk(branch string, buildBranches []string)
	for ind, tt := range branchTests {
		t.Run(fmt.Sprintf("%d", ind), func(t *testing.T) {
			ok, err := BranchRegexOk(tt.branch, tt.allowedBranches)
			if err != nil {
				t.Error(err)
			}
			if ok != tt.regexOk {
				t.Error(test.GenericStrFormatErrors("ok", tt.regexOk, ok))
			}
		})
	}
}
