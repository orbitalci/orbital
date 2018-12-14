package sshhelper

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/level11consulting/orbitalci/models"
	"golang.org/x/crypto/ssh"
)

func getClient(facts *models.SSHFacts) (*ssh.Client, error) {
	var auth []ssh.AuthMethod
	switch {
	case facts.Password != "":
		auth = append(auth, ssh.Password(facts.Password))
	case facts.KeyFP != "":
		buffer, err := ioutil.ReadFile(facts.KeyFP)
		if err != nil {
			return nil, err
		}

		key, err := ssh.ParsePrivateKey(buffer)
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(key))
	default:
		return nil, errors.New("must have either ssh password or path to private key")
	}
	sshConfig := &ssh.ClientConfig{
		User:            facts.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", facts.Host, facts.Port), sshConfig)
	if err != nil {
		return nil, errors.New("Unable to get connection, error is: " + err.Error())
	}
	return client, nil

}

// CreateSSHChannel will use the werker's configured ssh facts to create an SSH client. It will error at this point
//   if the client cannot connect to the remote sshd. It will also kick off the handleCtx method that will kill the active session and attempt to stop any associated processes
func CreateSSHChannel(ctx context.Context, facts *models.SSHFacts, hash string) (*Channel, error) {
	client, err := getClient(facts)
	if err != nil {
		return nil, err
	}
	channel := &Channel{client: client, hash: hash, ctx: ctx}
	go channel.handleCtx()
	return channel, nil
}

type Channel struct {
	client        *ssh.Client
	session       *ssh.Session
	ctx           context.Context
	globalEnvVars [][2]string
	hash          string
}

func (c *Channel) SetGlobals(envs []string) {
	c.globalEnvVars = splitEnvs(envs)
}

func (c *Channel) AppendGlobals(envs []string) {
	c.globalEnvVars = append(c.globalEnvVars, splitEnvs(envs)...)
}

func splitEnvs(envs []string) (split [][2]string) {
	for _, env := range envs {
		envArray := strings.SplitN(env, "=", 2)
		var fixed [2]string
		copy(fixed[:], envArray[:])
		split = append(split, fixed)
	}
	return split
}

func (c *Channel) handleCtx() {
	select {
	case <-c.ctx.Done():
		if c.session != nil {
			fmt.Println("CONTEXT IS DONE! KILLING SESSION SIGNAL!!!")
			//https://bugzilla.mindrot.org/show_bug.cgi?id=1424
			// right now, according to the bug report^^^, openssh doesn't support sending kill signals
			// soo... yeah
			// find a better way, i guess?
			//err := c.session.Signal(ssh.SIGKILL)
			command := fmt.Sprintf("kill $(ps aux | grep %s | grep -v grep  | awk '{print $2}')", c.hash)
			err := c.JustRun(command, []string{})
			if err != nil {
				fmt.Println("kill failed!!! error: ", err.Error())
			}
			c.Close()
		}
	}
}

// CheckConnection will create a session and attempt to run an echo command. Will return any errors; if error is
//   nil you can assume everything is a-ok for running commands.
func (c *Channel) CheckConnection() error {
	if c.client == nil {
		return errors.New("client has not been initialized, need to call Connect() first")
	}
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	err = session.Run("echo 'hi'")
	return err
}

// Setenvs sets the globalEnvVars and any extra environment variables to  be set on the ssh session
func (c *Channel) Setenvs(extraEnvs ...string) error {
	var err error
	if c.session == nil {
		return errors.New("woah there bud, can't set an env variable if there is no session to attach it to")
	}
	//c.session.Setenv("PATH", "$PATH:/usr/local/bin")
	for _, splitEnv := range c.globalEnvVars {
		if err = c.session.Setenv(splitEnv[0], splitEnv[1]); err != nil {
			return err
		}

	}
	for _, splitExtra := range splitEnvs(extraEnvs) {
		if err = c.session.Setenv(splitExtra[0], splitExtra[1]); err != nil {
			return err
		}
	}
	return nil
}

// StreamingFunc is the function that will read off the Reader and transform it, then write it to the logout channel.
// call wg.Done(), it synchronizes the ssh command execution. See BasicPipeHandler for implementation
type StreamingFunc func(r io.Reader, logout chan []byte, wg *sync.WaitGroup)

// BasicPipeHandler is a simple implementation of the StreamingFunc function type. It does not transform the data coming in,
// it just writes it directly to the logout channel.
func BasicPipeHandler(r io.Reader, logout chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logout <- scanner.Bytes()
	}
	if scanner.Err() != nil {
		logout <- []byte("An error has occured!")
		logout <- []byte(scanner.Err().Error())
	}
}

// RunAndLog runs a given command remotely via the ContextConnection's ssh client. The stdout and stderr will be processed
// by the given StreamingFunc function and written to logout. the function will wait for the StreamingFunc function to
// close its done channel on both the stdout processing and the stderr processing.
func (c *Channel) RunAndLog(cmd string, envs []string, logout chan []byte, streamingFunc StreamingFunc) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	c.session = session
	// reset the session attribute. maybe later i will realize that we should just persist the session for the
	// entire build, but idk right now
	defer func() { c.session = nil }()
	defer session.Close()
	// set environment variables
	err = c.Setenvs(envs...)
	if err != nil {
		return err
	}
	out, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	err = session.Start(cmd)
	if err != nil {
		return err
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go streamingFunc(out, logout, wg)
	go streamingFunc(stderr, logout, wg)
	err = session.Wait()
	wg.Wait()
	return err
}

func (c *Channel) JustRun(cmd string, envs []string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	c.session = session
	defer func() { c.session = nil }()
	defer session.Close()
	err = c.Setenvs(envs...)
	return session.Run(cmd)
}

func (c *Channel) Close() error {
	return c.client.Close()
}
