package builder

import (
	"testing"
	"bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/go-til/test"
)

func TestBasher_DownloadCodebaseDefault(t *testing.T) {
	b := &Basher{}
	wt := &protos.WerkerTask{
		VcsToken: "something",
		FullName: "marianne",
		CheckoutHash: "123",
		VcsType: "bitbucket",
	}
	defaultResult := b.DownloadCodebase(wt)
	if len(defaultResult) != 4 {
		t.Error(test.GenericStrFormatErrors("default download length", 4, len(defaultResult)))
	}
	if defaultResult[0] != ".ocelot/bb_download.sh" {
		t.Error(test.GenericStrFormatErrors("download first param", ".ocelot/bb_download.sh", defaultResult[0]))
	}
	if defaultResult[1] != "something" {
		t.Error(test.GenericStrFormatErrors("download second param(token)", "something", defaultResult[1]))
	}
	if defaultResult[2] != "https://bitbucket.org/marianne/get" {
		t.Error(test.GenericStrFormatErrors("download third param(url)", "https://bitbucket.org/marianne/get", defaultResult[2]))
	}
	if defaultResult[3] != "123" {
		t.Error(test.GenericStrFormatErrors("download fourth param(git hash)", "123", defaultResult[3]))
	}
}

func TestBasher_DownloadCodebaseNotDefault(t *testing.T) {
	b := &Basher{}
	b.SetBbDownloadURL("https://localhost:9090/marianne/is/number/one")
	wt := &protos.WerkerTask{
		VcsToken: "",
		FullName: "marianne",
		CheckoutHash: "123",
		VcsType: "bitbucket",
	}
	defaultResult := b.DownloadCodebase(wt)
	if len(defaultResult) != 4 {
		t.Error(test.GenericStrFormatErrors("default download length", 4, len(defaultResult)))
	}
	if defaultResult[0] != ".ocelot/bb_download.sh" {
		t.Error(test.GenericStrFormatErrors("download first param", ".ocelot/bb_download.sh", defaultResult[0]))
	}
	if defaultResult[1] != "something" {
		t.Error(test.GenericStrFormatErrors("download second param(token)", "something", defaultResult[1]))
	}
	if defaultResult[2] != "https://bitbucket.org/marianne/get" {
		t.Error(test.GenericStrFormatErrors("download third param(url)", "https://bitbucket.org/marianne/get", defaultResult[2]))
	}
	if defaultResult[3] != "123" {
		t.Error(test.GenericStrFormatErrors("download fourth param(git hash)", "123", defaultResult[3]))
	}
}

