package poll

import (
	"bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/old/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"os"
	"path/filepath"
)

var cronDir = "/etc/cron.d"

type MsgHandler struct {
	Topic string
	Store storage.PollTable
}

func (m *MsgHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	log.Log().Debug("unmarshaling and processing a poll update msg")
	pollMsg := &pb.PollRequest{}
	defer func(){
		done <- 1
	}()
	if err := proto.Unmarshal(msg, pollMsg); err != nil {
		log.IncludeErrField(err).Error("unmarshal error for poll msg")
		return err
	}
	var err error
	switch m.Topic {
	case "poll_please":
		err = WriteCronFile(pollMsg)
		if err != nil {
			// even if we can't write cron tab, should register that it was requested
			log.IncludeErrField(err).Error("UNABLE TO WRITE CRON TAB")
		}
	case "no_poll_please":
		log.Log().Info("recieved a request for no_poll_please")
		err = DeleteCronFile(pollMsg)
		if err != nil {
			log.IncludeErrField(err).Error("UNABLE TO DELETE CRON TAB")
		} else {
			log.Log().Info("successfully deleted cron tab")
		}
	default:
		err = errors.New("only supported topics are poll_please and no_poll_please")
		log.IncludeErrField(err).Error()
	}
	return err
}

func DeleteCronFile(event *pb.PollRequest) error {
	fullPath := filepath.Join(cronDir, event.Account + "_" + event.Repo)
	err := os.Remove(fullPath)
	return err
}

func WriteCronFile(event *pb.PollRequest) error {
	cron := fmt.Sprintf("%s root /bin/run_changecheck.sh %s/%s %s\n", event.Cron, event.Account, event.Repo, event.Branches)
	fullPath := filepath.Join(cronDir, event.Account + "_" + event.Repo)
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