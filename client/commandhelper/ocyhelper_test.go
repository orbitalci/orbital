package commandhelper

import (
	"flag"
	"fmt"
	"testing"

	"github.com/shankj3/go-til/test"
)

func TestOcyHelper_SetGitHelperFlags(t *testing.T) {
	oh := &OcyHelper{}
	flg := flag.NewFlagSet("h", flag.ExitOnError)
	oh.SetGitHelperFlags(flg, true, true ,true)
	shouldBeSet(t, flg, "acct-repo")
	shouldBeSet(t, flg, "hash")
	shouldBeSet(t, flg, "vcs-type")
	if err := flg.Parse([]string{"-hash=12345678"}); err != nil {
		t.Error("should be playing nice... ", err.Error())
	}
	shouldEqual(t, flg, "hash", "12345678")
	flg = flag.NewFlagSet("e", flag.ExitOnError)
	oh.SetGitHelperFlags(flg, false, false, true)
	shouldNotBeSet(t, flg, "acct-repo")
	shouldNotBeSet(t, flg, "hash")
	shouldBeSet(t, flg, "vcs-type")
	if err := flg.Parse([]string{"-vcs-type=bitbucket"}); err != nil {
		t.Error("should be playing nice... ", err.Error())
	}
	shouldEqual(t, flg, "vcs-type", "bitbucket")
	flg = flag.NewFlagSet("l", flag.ExitOnError)
	oh.SetGitHelperFlags(flg, true, false, true)
	shouldBeSet(t, flg, "acct-repo")
	shouldNotBeSet(t, flg, "hash")
	shouldBeSet(t, flg, "vcs-type")
	if err := flg.Parse([]string{"-acct-repo=shankj3/go-til", "-vcs-type=github"}); err != nil {
		t.Error("should be playing nice... ", err.Error())
	}
	shouldEqual(t, flg, "vcs-type", "github")
	flg = flag.NewFlagSet("p", flag.ExitOnError)
	oh.SetGitHelperFlags(flg, false, false, false)
	shouldNotBeSet(t, flg, "acct-repo")
	shouldNotBeSet(t, flg, "hash")
	shouldNotBeSet(t, flg, "vcs-type")
}

func shouldBeSet(t *testing.T, flagSet *flag.FlagSet, flagName string) {
	if flaggy := flagSet.Lookup(flagName); flaggy == nil {
		t.Error(fmt.Sprintf("-%s should have been set as a flag", flagName))
	}
}

func shouldNotBeSet(t *testing.T, flagSet *flag.FlagSet, flagName string) {
	if flaggy := flagSet.Lookup(flagName); flaggy != nil {
		t.Error(fmt.Sprintf("-%s should not have been set as a flag", flagName))
	}
}

func shouldEqual(t *testing.T, flagSet *flag.FlagSet, flagName, flagValue string) {
	flaggy := flagSet.Lookup(flagName)
	if flaggy == nil {
		t.Error(fmt.Sprintf("-%s should have been set as a flag", flagName))
	}
	live := flaggy.Value.String()
	if live != flagValue {
		t.Error(test.StrFormatErrors(flagName, flagValue, live))
	}
}
