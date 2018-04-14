package poll

import (
	"bitbucket.org/level11consulting/go-til/test"
	pb "bitbucket.org/level11consulting/ocelot/models/pb"
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestWriteCronFile_DeleteCronFile(t *testing.T) {
	cronDir = "./test-fixtures"
	event := &pb.PollRequest{Account: "accOmns7f", Repo: "d8sfasdnc3", Cron: "* * * * *", Branches: "test,master,queue"}
	WriteCronFile(event)
	expected := []byte("* * * * * root /bin/run_changecheck.sh accOmns7f/d8sfasdnc3 test,master,queue\n")
	bytez, err := ioutil.ReadFile("./test-fixtures/accOmns7f_d8sfasdnc3")
	if err != nil {
		t.Log(os.Getwd())
		files, err := ioutil.ReadDir("./test-fixtures")
		if err != nil {
			t.Error(err)
		}

		for _, file := range files {
			t.Log(file.Name())
		}
		t.Error(err)
		return
	}
	if !bytes.Equal(bytez, expected) {
		t.Error(test.StrFormatErrors("cron file contents", string(expected), string(bytez)))
	}
	DeleteCronFile(event)
	ok, err := exists("./test-fixtures/accOmns7f_d8sfasdnc3")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("cron file should be deleted. it isnt")
	}

}

func TestMsgHandler_UnmarshalAndProcess(t *testing.T) {

}
