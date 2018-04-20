package builder

import (
	"bufio"
	"context"
	"io"
	"os/exec"

	"bitbucket.org/level11consulting/ocelot/build/basher"
	"bitbucket.org/level11consulting/ocelot/build/valet"
	cred"bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

type Host struct {
	*basher.Basher
	// host needs buildValet because it's gonna need to update the proc file at every damn turn
	bv *valet.Valet
}
/*
what do we need out of bare metal builds? this is pretty much *exclusively* for any ios builds, because you can't run those in a docker container or on kubernetes.

*/

//type Builder interface {
//	basher.OcyBash
//	Setup(ctx context.Context, logout chan []byte, dockerId chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (res *pb.Result, uuid string)
//	Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result
//	ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *StageUtil, logout chan[]byte) *pb.Result
//}

func runCommandLogToChan(command *exec.Cmd, logout chan []byte) error{
	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()
	command.Start()
	//https://stackoverflow.com/questions/45922528/how-to-force-golang-to-close-opened-pipe
	go streamFromPipe(logout, stdout)
	go streamFromPipe(logout, stderr)
	err := command.Wait()
	return err
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