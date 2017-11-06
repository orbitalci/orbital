package run

import (
	"github.com/shankj3/ocelot/configure"
	"log"
	"os"
	"os/exec"
)

func RunStage(stage configure.Stage) {
	for _, command := range stage.Script {
		cmd := exec.Command(command)
		stdout, _ := cmd.StdoutPipe()
		cmd.Env = append(os.Environ(), stage.Env...)
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		} else {
			log.Print(stdout)
		}
	}
}
