package poller

import (
	"bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"os"
	"path/filepath"
)

type MsgHandler struct {
	Topic string
	Store storage.PollTable
}

func (m *MsgHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	log.Log().Debug("unmarshaling and processing a poll update msg")
	pollMsg := &pb.PollRequest{}
	if err := proto.Unmarshal(msg, pollMsg); err != nil {
		log.IncludeErrField(err).Error("unmarshal error for poll msg")
		return err
	}
	err := WriteCronFile(pollMsg)
	if err != nil {
		// even if we can't write cron tab, should register that it was requested
		log.IncludeErrField(err).Error("UNABLE TO WRITE CRON TAB")
	}
	// this happens on admin side
	//if pollMsg.IsUpdate {
	//	err = m.Store.UpdatePoll(pollMsg.Account, pollMsg.Repo, true, pollMsg.Cron, pollMsg.Branches)
	//} else {
	//	err = m.Store.InsertPoll(pollMsg.Account, pollMsg.Repo, true, pollMsg.Cron, pollMsg.Branches)
	//}
	//if err != nil {
	//	log.IncludeErrField(err).Error("unable to save cron to database!")
	//} else {
	//	log.Log().WithField("account", pollMsg.Account).WithField("repo", pollMsg.Repo).WithField("cron", pollMsg.Cron).Info("successfully received poll message, wrote cron job, and saved to db")
	//}
	done <- 1
	return err
}

func WriteCronFile(event *pb.PollRequest) error {
	cron := fmt.Sprintf("%s root /bin/run_changecheck.sh %s/%s %s\n", event.Cron, event.Account, event.Repo, event.Branches)
	basePath := "/etc/cron.d"
	fullPath := filepath.Join(basePath, event.Account + "_" + event.Repo)
	isfile, err := exists(fullPath)
	if err != nil {
		return err
	}
	if isfile {
		os.Remove(fullPath)
	}
	log.Log().Info("writing cron tab to ", fullPath)
	err = ioutil.WriteFile(fullPath, []byte(cron), 0644)
	return err
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}