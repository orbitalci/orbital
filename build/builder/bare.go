package builder

import (
	"bufio"
	"context"
	"io"
	"os/exec"

	cred"bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

type Host struct {

}

func (h *Host) Setup(ctx context.Context, logout chan []byte, dockerIdChan chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (*pb.Result, string) {
	cmd := exec.CommandContext(ctx, "sh", "-c", "some long runnig task")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	//https://stackoverflow.com/questions/45922528/how-to-force-golang-to-close-opened-pipe
	go streamFromPipe(logout, stdout)
	go streamFromPipe(logout, stderr)
	err := cmd.Wait()
	if err != nil {
		return &pb.Result {
			Stage:  "testy testy",
			Status: pb.StageResultVal_FAIL,
			Error: err.Error(),
		}, ""
	}
	return &pb.Result{Stage: "suuup", Status: pb.StageResultVal_PASS, Error:""}, ""
}

func streamFromPipe(logout chan []byte, pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		logout <- scanner.Bytes()
	}
}