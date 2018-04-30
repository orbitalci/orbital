package commandhelper

import (
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/common/testutil"
	"testing"
)

// todo: do we know of any other type of url....?
var goodgiturls = []struct {
	name     string
	url      []byte
	acctRepo string
}{
	{"bitbucket ssh", []byte("git@bitbucket.org:level11consulting/ocelot.git"), "level11consulting/ocelot"},
	{"bitbucket https", []byte("https://jessishank@github.com/shankj3/ocelot.git"), "level11consulting/ocelot"},
	{"github https", []byte("https://github.com/kubernetes/charts.git"), "kubernetes/charts"},
	{"github ssh", []byte("git@github.com:kubernetes/charts.git"), "kubernetes/charts"},
}

func TestFindAcctRepo(t *testing.T) {
	testutil.BuildServerHack(t)
	acctRepo, err := FindAcctRepo()
	if err != nil {
		t.Fatal(err)
	}
	if acctRepo != "level11consulting/ocelot" {
		t.Error(test.StrFormatErrors("detected acct/repo", "level11consulting/ocelot", acctRepo))
	}
}

func Test_matchThis(t *testing.T) {
	for _, tt := range goodgiturls {
		t.Run(tt.name, func(t *testing.T) {
			acctRepo, err := matchThis(tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if acctRepo != tt.acctRepo {
				t.Error(test.StrFormatErrors("parsed account repo", tt.acctRepo, acctRepo))
			}
		})
	}
}
