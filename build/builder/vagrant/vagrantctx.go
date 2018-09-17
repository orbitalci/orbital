package vagrant

import (
	"context"
	"os/exec"

	"github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
)

// VagrantUp will run the command "vagrant up" from the vagrantDir specified. It will also save the output
// of vagrant ssh-config to the vagrantDir as the file vagrant-ssh for easy config of ssh clients
func VagrantUp(ctx context.Context, vagrantDir string, infoChan chan []byte, stage *build.StageUtil) error {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "vagrant up")
	cmd.Dir = vagrantDir
	if err := runCommandLogToChan(cmd, infoChan, stage); err != nil {
		return err
	}
	go handleCancel(ctx, vagrantDir, infoChan)
	cmd = exec.CommandContext(ctx, "/bin/sh", "-c", "vagrant ssh-config > vagrant-ssh")
	cmd.Dir = vagrantDir
	return runCommandLogToChan(cmd, infoChan, stage)
}

func VagrantDown(ctx context.Context, vagrantDir string, infoChan chan []byte) error {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "vagrant destroy -f")
	cmd.Dir = vagrantDir
	if err := runCommandLogToChan(cmd, infoChan, build.InitStageUtil("vagrant down")); err != nil {
		return err
	}
	return nil
}

func handleCancel(ctx context.Context, vagrantDir string, infoChan chan []byte) error {
	select {
	case <-ctx.Done():
		log.Log().Info("KILLING VAGRANT")
		err := VagrantDown(ctx, vagrantDir, infoChan)
		if err != nil {
			log.IncludeErrField(err).Error("unable to kill vagrant")
		}
		return err
	}
}
