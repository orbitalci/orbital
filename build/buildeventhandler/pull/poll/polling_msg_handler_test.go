package poll

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/go-til/test"
	pb "github.com/level11consulting/orbitalci/models/pb"

	"io/ioutil"
	"os"
	"testing"
)

func TestWriteCronFile_DeleteCronFile(t *testing.T) {
	event := &pb.PollRequest{Account: "accOmns7f", Repo: "d8sfasdnc3", Cron: "* * * * *", Branches: "test,master,queue"}
	WriteCronFile(event, "./test-fixtures")
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
	DeleteCronFile(event, "./test-fixtures")
	ok, err := exists("./test-fixtures/accOmns7f_d8sfasdnc3")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("cron file should be deleted. it isnt")
	}

}

func Test_exists(t *testing.T) {
	there, _ := exists("./test-fixtures/.keepme")
	if !there {
		t.Error("./test-fixtures/.keepme exists")
	}
	there, _ = exists("./test-fixtures/nada")
	if there {
		t.Error("nada doesn't exist")
	}
}

func TestMsgHandler_UnmarshalAndProcess_pollplz(t *testing.T) {
	var pollRequest = &pb.PollRequest{
		Account:  "acct",
		Repo:     "repo",
		Cron:     "* * * * *",
		Branches: "ALL",
	}
	defer os.Remove("./test-fixtures/acct_repo")
	nopoll := &MsgHandler{Topic: "poll_please", cronDir: "./test-fixtures"}
	pollBytes, err := proto.Marshal(pollRequest)
	if err != nil {
		t.Error(err)
		return
	}
	done, finish := make(chan int, 1), make(chan int, 1)
	err = nopoll.UnmarshalAndProcess(pollBytes, done, finish)
	if err != nil {
		t.Error("this should work fine")
	}
	if isthere, _ := exists("./test-fixtures/acct_repo"); !isthere {
		t.Error("a cron file should have been written ")
	}

}

func TestMsgHandler_UnmarshalAndProcess_pollplz_fail(t *testing.T) {
	var pollRequest = &pb.PollRequest{
		Account:  "acct",
		Repo:     "repo",
		Cron:     "* * * * *",
		Branches: "ALL",
	}
	nopoll := &MsgHandler{Topic: "poll_please", cronDir: "./test-fixtures/dlajkfklsdjfklajsdfkl;ajsdflk;jads"}
	pollBytes, err := proto.Marshal(pollRequest)
	if err != nil {
		t.Error(err)
		return
	}
	done, finish := make(chan int, 1), make(chan int, 1)
	err = nopoll.UnmarshalAndProcess(pollBytes, done, finish)
	if err == nil {
		t.Error("dlajkfklsdjfklajsdfkl;ajsdflk;jads is not a valid directory, this should fail")
	}
}

func TestMsgHandler_UnmarshalAndProcess_nopollplz(t *testing.T) {
	var pollRequest = &pb.PollRequest{
		Account:  "acct2",
		Repo:     "repo2",
		Cron:     "* * * * *",
		Branches: "ALL",
	}
	pollBytes, err := proto.Marshal(pollRequest)
	done, finish := make(chan int, 1), make(chan int, 1)
	poll := &MsgHandler{Topic: "poll_please", cronDir: "./test-fixtures"}
	err = poll.UnmarshalAndProcess(pollBytes, done, finish)
	if err != nil {
		t.Error(err)
	}
	done, finish = make(chan int, 1), make(chan int, 1)
	nopoll := &MsgHandler{Topic: "no_poll_please", cronDir: "./test-fixtures"}
	if err != nil {
		t.Error(err)
		return
	}
	err = nopoll.UnmarshalAndProcess(pollBytes, done, finish)
	if err != nil {
		t.Error("should have successfully deleted cron tab")
	}
	if there, _ := exists("./test-fixtures/acct2_repo2"); there {
		t.Error("./test-fixtures/acct2_repo2 shouldn't be there naymore")
	}
}

func TestMsgHandler_UnmarshalAndProcess_default(t *testing.T) {
	var pollRequest = &pb.PollRequest{
		Account:  "acct2",
		Repo:     "repo2",
		Cron:     "* * * * *",
		Branches: "ALL",
	}
	pollBytes, err := proto.Marshal(pollRequest)
	done, finish := make(chan int, 1), make(chan int, 1)
	poll := &MsgHandler{Topic: "asdf", cronDir: "./test-fixtures"}
	err = poll.UnmarshalAndProcess(pollBytes, done, finish)
	if err == nil {
		t.Error("bad topic name, this should not succesed")
	}
	if err.Error() != "only supported topics are poll_please and no_poll_please" {
		t.Error(test.StrFormatErrors("err msg", "only supported topics are poll_please and no_poll_please", err.Error()))
	}
}

func TestMsgHandler_UnmarshalAndProcess_baddata(t *testing.T) {
	pollBytes := []byte("!!!!hummunu")
	poll := &MsgHandler{Topic: "poll_please", cronDir: "./test-fixtures"}
	done, finish := make(chan int, 1), make(chan int, 1)
	err := poll.UnmarshalAndProcess(pollBytes, done, finish)
	if err == nil {
		t.Error("this is a bad poll proto message that was sent. this should return an error.")
	}
	if err.Error() != "proto: can't skip unknown wire type 6" {
		t.Error("error should have come from proto unmarshaling, error instead is: " + err.Error())
	}
}

func TestNewMsgHandler(t *testing.T) {
	handle := NewMsgHandler("topic1")
	if handle.cronDir != "/etc/cron.d" {
		t.Error(test.StrFormatErrors("cron direc", "/etc/cron.d", handle.cronDir))
	}
}
