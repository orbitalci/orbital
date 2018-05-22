package exechelper

import (
	"io"
	"os/exec"
	"sync"
)

// StreamingFunc is the function that will process the input coming off of either stdoutpipe() or stderrpipe() and write it to the stream channel. this function is expected to call wg.Done() upon finish
type StreamingFunc func(input io.ReadCloser, streamChan chan []byte, wg *sync.WaitGroup, inputDescription string)

func RunAndStreamCmd(cmd *exec.Cmd, logout chan[]byte, procFunc StreamingFunc) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	var wg = new(sync.WaitGroup)
	if err = cmd.Start(); err != nil {
		return err
	}
	wg.Add(2)
	go procFunc(stdout, logout, wg, "std out")
	go procFunc(stderr, logout, wg, "std err")
	err = cmd.Wait()
	wg.Wait()
	return err
}

