package builder

import (
	"context"
	"os/exec"

	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/build"
)


func VagrantUp(ctx context.Context, vagrantDir string, infoChan chan[]byte, stage *build.StageUtil) error {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "vagrant up")
	cmd.Dir = vagrantDir
	if err := runCommandLogToChan(cmd, infoChan, stage); err != nil {
		return err
	}
	go handleCancel(ctx, vagrantDir, infoChan)
	return nil
}

func VagrantDown(ctx context.Context, vagrantDir string, infoChan chan[]byte) error {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "vagrant destroy -f")
	cmd.Dir = vagrantDir
	if err := runCommandLogToChan(cmd, infoChan, build.InitStageUtil("vagrant down")); err != nil {
		return err
	}
	return nil
}

func handleCancel(ctx context.Context, vagrantDir string, infoChan chan[]byte) error {
	select {
	case <-ctx.Done():
		err := VagrantDown(ctx, vagrantDir, infoChan)
		if err != nil {
			log.IncludeErrField(err).Error("unable to kill vagrant")
		}
		return err
	}
}