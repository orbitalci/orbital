package basher

import (
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
	"testing"
)

func TestBasher_DownloadCodebaseDefault(t *testing.T) {
	b := &Basher{}
	wt := &pb.WerkerTask{
		VcsToken:     "something",
		FullName:     "marianne",
		CheckoutHash: "123",
		VcsType:      pb.SubCredType_BITBUCKET,
	}
	defaultResult := b.DownloadCodebase(wt)
	if len(defaultResult) != 3 {
		t.Error(test.GenericStrFormatErrors("default download length", 3, len(defaultResult)))
	}
	if defaultResult[0] != "/bin/sh" {
		t.Error(test.GenericStrFormatErrors("download first param", "/bin/sh", defaultResult[0]))
	}
	if defaultResult[1] != "-c" {
		t.Error(test.GenericStrFormatErrors("download second param", "-c", defaultResult[1]))
	}
	if defaultResult[2] != "/.ocelot/bb_download.sh something https://x-token-auth:something@bitbucket.org/marianne.git 123 /.ocelot/123" {
		t.Error(test.GenericStrFormatErrors("download third param", "/.ocelot/bb_download.sh something https://x-token-auth:something@bitbucket.org/marianne.git 123 /.ocelot/123", defaultResult[2]))
	}
}

func TestBasher_DownloadCodebaseNotDefault(t *testing.T) {
	b := &Basher{}
	b.SetBbDownloadURL("https://localhost:9090/marianne/is/number/one")
	wt := &pb.WerkerTask{
		VcsToken:     "",
		FullName:     "marianne",
		CheckoutHash: "123",
		VcsType:      pb.SubCredType_BITBUCKET,
	}
	defaultResult := b.DownloadCodebase(wt)
	if len(defaultResult) != 3 {
		t.Error(test.GenericStrFormatErrors("default download length", 3, len(defaultResult)))
	}
	if defaultResult[0] != "/bin/sh" {
		t.Error(test.GenericStrFormatErrors("download first param", "/bin/sh", defaultResult[0]))
	}
	if defaultResult[1] != "-c" {
		t.Error(test.GenericStrFormatErrors("download second param", "-c", defaultResult[1]))
	}
	if defaultResult[2] != "/.ocelot/bb_download.sh  https://localhost:9090/marianne/is/number/one 123 /.ocelot/123" {
		t.Error(test.GenericStrFormatErrors("download first param", "/.ocelot/bb_download.sh  https://localhost:9090/marianne/is/number/one 123 /.ocelot/123", defaultResult[2]))
	}
}
