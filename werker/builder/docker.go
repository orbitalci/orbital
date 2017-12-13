package builder

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"bufio"
)

//TODO: create something using composition that runs run's each stage's bash commands
//TODO: for you
type Docker struct {}

func NewDockerBuilder() Builder {
	return &Docker{}
}

func (d *Docker) Setup(logout chan []byte, image string) *Result {
	ocelog.Log().Debug("doing setup")

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imageName := image

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	defer out.Close()

	if err != nil {
		panic(err)
	}

	//TODO: figure out what exactly happens here
	bufReader := bufio.NewReader(out)
	go d.writeToInfo(bufReader, logout)
	return &Result{}
}

func (d *Docker) Before(logout chan []byte)	*Result {
	return &Result{}
}

func (d *Docker) Build(logout chan []byte)	*Result {
	return &Result{}
}

func (d *Docker) After(logout chan []byte)	*Result {
	return &Result{}
}

func (d *Docker) Test(logout chan []byte)	*Result {
	return &Result{}
}

func (d *Docker) Deploy(logout chan []byte) *Result {
	return &Result{}
}

func (d *Docker) writeToInfo(rd *bufio.Reader, infochan chan []byte) {
	for {
		str, err := rd.ReadString('\n')
		//TODO:
		if err != nil {
			ocelog.Log().Warn("Read Error:", err)
			return
		}
		infochan <- []byte(str)
	}
}

