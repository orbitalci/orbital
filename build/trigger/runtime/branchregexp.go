package runtime

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	regexFailure = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ocelot_regex_failures",
		Help: "failures by regex parser in validating branches",
	})
)

func init() {
	prometheus.MustRegister(regexFailure)
}

// BranchRegexOk will attempt to do a regex match on each of the build branches. it will return true if any entry is 'ALL' or if there was a successful regex match.
//   An error will be returned if one of the build branches fails to be compiled into a regex expression
func BranchRegexOk(branch string, buildBranches []string) (bool, error) {
	for _, goodBranchRe := range buildBranches {
		if goodBranchRe == "ALL" {
			return true, nil
		}
		re, err := regexp.Compile("^" + goodBranchRe + "$")
		if err != nil {
			regexFailure.Inc()
			return false, errors.New(fmt.Sprintf("unable to parse acceptable branch item (%s) into regex expression. error is: %s", goodBranchRe, err.Error()))
		}
		if re.MatchString(branch) {
			return true, nil
		}
	}
	return false, nil
}
