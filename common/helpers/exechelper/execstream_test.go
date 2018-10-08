package exechelper

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/shankj3/go-til/test"
)

var dat string

type tester struct {
	t *testing.T
}

func (t *tester) streamer(input io.ReadCloser, streamChan chan []byte, wg *sync.WaitGroup, desc string) {
	defer wg.Done()
	defer input.Close()
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		dat += desc + " ||| " + string(scanner.Bytes()) + "\n"
	}
	if err := scanner.Err(); err != nil {
		t.t.Error(err)
	}
}

func TestRunAndStreamCmd(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	cmd1 := exec.CommandContext(ctx, "/bin/bash", "-c", ">&2 echo \"error\" && sleep 0.1 && echo \"stdout\"")
	logout := make(chan []byte, 100)
	te := &tester{t: t}
	err := RunAndStreamCmd(cmd1, logout, te.streamer)
	if err != nil {
		t.Error(err)
	}
	t.Log(dat)
	expected := `std err ||| error
std out ||| stdout
`
	if dat != expected {
		t.Error(test.StrFormatErrors("output", expected, dat))
	}
	dat = ""
	ctx, cancel := context.WithCancel(context.Background())
	cmd2 := exec.CommandContext(ctx, "/bin/sh", "-c", "echo \"why\" && sleep 10")
	go RunAndStreamCmd(cmd2, logout, te.streamer)
	time.Sleep(1 * time.Second)
	cancel()
	expected = `std out ||| why
`
	if dat != expected {
		t.Error(test.StrFormatErrors("output", expected, dat))
	}
	t.Log(dat)
	fmt.Println("ctx err " + ctx.Err().Error())
}
