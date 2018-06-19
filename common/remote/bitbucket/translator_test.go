package bitbucket

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shankj3/go-til/test"
)

func TestBBTranslate_TranslatePR(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	file, err := os.Open(filepath.Join(pwd, "test-fixtures", "1528734656-pr-bb.json"))
	defer file.Close()
	trans := GetTranslator()
	pull, err := trans.TranslatePR(file)
	if err != nil {
		t.Error(err)
		return
	}
	if pull.Source.Hash != "dc128f78cd34" {
		t.Error(test.StrFormatErrors("source hash", "dc128f78cd34", pull.Source.Hash))
	}
	if pull.Source.Branch != "newbranch" {
		t.Error(test.StrFormatErrors("source branch", "newbranch", pull.Source.Branch))
	}
	if pull.Id != 1 {
		t.Error(test.GenericStrFormatErrors("pr id", 1, pull.Id))
	}
	commits := "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/1/commits"
	if pull.Urls.Commits != commits {
		t.Error(test.StrFormatErrors("commits url", commits, pull.Urls.Commits))
	}
	comments := "https://api.bitbucket.org/2.0/repositories/jessishank/mytestocy/pullrequests/1/comments"
	if pull.Urls.Comments != comments {
		t.Error(test.StrFormatErrors("comments url", comments, pull.Urls.Comments))
	}
	if pull.Title != "Newbranch" {
		t.Error(test.StrFormatErrors("titile", "Newbranch", pull.Title))
	}
	if pull.Destination.Branch != "master" {
		t.Error(test.StrFormatErrors("dest branch", "master", pull.Destination.Branch))
	}
	if pull.Destination.Hash != "32ed49560d10" {
		t.Error(test.StrFormatErrors("dest hash", "32ed49560d10", pull.Destination.Hash))
	}
}

