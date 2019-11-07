package commandhelper

import (
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/orbitalci/common/testutil"
	"github.com/level11consulting/orbitalci/models/pb"

	"testing"
)

// todo: do we know of any other type of url....?
var goodgiturls = []struct {
	name     string
	url      []byte
	acctRepo string
	vcsType  pb.SubCredType
}{
	{"bitbucket ssh", []byte("git@bitbucket.org:level11consulting/ocelot.git"), "level11consulting/ocelot", pb.SubCredType_BITBUCKET},
	{"bitbucket ssh no .git", []byte("git@bitbucket.org:level11consulting/ocelot"), "level11consulting/ocelot", pb.SubCredType_BITBUCKET},
	{"bitbucket https", []byte("https://jessishank@bitbucket.org/level11consulting/ocelot.git"), "level11consulting/ocelot", pb.SubCredType_BITBUCKET},
	{"bitbucket https no .git", []byte("https://jessishank@bitbucket.org/level11consulting/ocelot"), "level11consulting/ocelot", pb.SubCredType_BITBUCKET},
	{"github https", []byte("https://github.com/kubernetes/charts.git"), "kubernetes/charts", pb.SubCredType_GITHUB},
	{"github https no .git", []byte("https://github.com/kubernetes/charts"), "kubernetes/charts", pb.SubCredType_GITHUB},
	{"github ssh", []byte("git@github.com:kubernetes/charts.git"), "kubernetes/charts", pb.SubCredType_GITHUB},
	{"github ssh no .git", []byte("git@github.com:kubernetes/charts"), "kubernetes/charts", pb.SubCredType_GITHUB},
}

func TestFindAcctRepo(t *testing.T) {
	testutil.BuildServerHack(t)
	acctRepo, vcst, err := FindAcctRepo()
	if err != nil {
		t.Fatal(err)
	}
	if vcst != pb.SubCredType_GITHUB {
		t.Error("should detect github")
	}
	if acctRepo != "level11consulting/ocelot" {
		t.Error(test.StrFormatErrors("detected acct/repo", "level11consulting/ocelot", acctRepo))
	}
}

func Test_matchThis(t *testing.T) {
	for _, tt := range goodgiturls {
		t.Run(tt.name, func(t *testing.T) {
			acctRepo, _, err := matchThis(tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if acctRepo != tt.acctRepo {
				t.Error(test.StrFormatErrors("parsed account repo", tt.acctRepo, acctRepo))
			}
		})
	}
}
