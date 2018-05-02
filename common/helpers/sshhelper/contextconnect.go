package sshhelper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/crypto/ssh"
)

type ContextConnection struct {
	session       *ssh.Session
	client		  *ssh.Client
	globalEnvVars [][2]string
	*connectionVars
}

type connectionVars struct {
	privKey   string
	user      string
	host      string
	password  string
	port      int
}


func (c *connectionVars) getSSHCli(conf *ssh.ClientConfig) (*ssh.Client, error) {
	// todo: not sure if this connection should be persisted long term? idk; keep an eye on leaking connections i guess.
	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port), conf)
	if err != nil {
		return nil, errors.New("Unable to get connection, error is: " + err.Error())
	}
	return connection, nil
}

func (c *connectionVars) getClientConfig() (*ssh.ClientConfig, error){
	var auth []ssh.AuthMethod
	switch {
	case c.password != "":
		auth = append(auth, ssh.Password(c.password))
	case c.privKey != "":
		buffer, err := ioutil.ReadFile(c.privKey)
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
		User: c.user,
		Auth: auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return sshConfig, nil
}

// InitContextConnect will instantiate a new instance of ContextConnection. Connect() will still have to be called on the object
// to generate a new session for remote command execution
func InitContextConnect(keyFp, password, user, host string, port int) *ContextConnection {
	return &ContextConnection{connectionVars: &connectionVars{privKey: keyFp, password: password, user: user, host:host, port: port}}
}

// Create will attempt to create an ssh client with the parameters provided here, and it will also kick off a
//   goroutine that will handle the context cancellation
func (c *ContextConnection) Connect(ctx context.Context) error {
	cliConf, err := c.getClientConfig()
	if err != nil {
		return err
	}
	sshCli, err := c.getSSHCli(cliConf)
	if err != nil {
		return err
	}
	c.client = sshCli
	go c.handleCtx(ctx)
	return nil
}

func (c *ContextConnection) SetGlobals(envs []string) {
	splitUp := splitEnvs(envs)
	c.globalEnvVars = splitUp
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


func (c *ContextConnection) handleCtx(ctx context.Context) {
	select {
	case <-ctx.Done():
		if c.session != nil {
			fmt.Println("CONTEXT IS DONE! KILLING SESSION SIGNAL!!!")
			err := c.session.Signal(ssh.SIGKILL)
			fmt.Println(err)
			err = c.session.Close()
			fmt.Println(err)
		}
		c.Close()
	}
}

// CheckConnection will create a session and attempt to run an echo command. Will return any errors; if error is
//   nil you can assume everything is a-ok for running commands.
func (c *ContextConnection) CheckConnection() error {
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
func (c *ContextConnection) Setenvs(extraEnvs ...string) error {
	var err error
	if c.session == nil {
		return errors.New("woah there bud, can't set an env variable if there is no session to attach it to")
	}

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

// Pipehandler is the function that will read off the Reader and transform it, then write it to the logout channel.
// close the done channel when the function has finished processing, it synchronizes the ssh command execution.
type PipeHandler func(r io.Reader, logout chan[]byte, done chan int)

// RunAndLog runs a given command remotely via the ContextConnection's ssh client. The stdout and stderr will be processed
// by the given PipeHandler function and written to logout. the function will wait for the pipehandler function to
// close its done channel on both the stdout processing and the stderr processing.
func (c *ContextConnection) RunAndLog(cmd string, envs []string, logout chan []byte,  pipeHandler PipeHandler) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	c.session = session
	// reset the session attribute. maybe later i will realize that we should just persist the session for the
	// entire build, but idk right now
	defer func(){c.session = nil}()
	defer c.session.Close()
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
	outchan := make(chan int)
	errchan := make(chan int)
	go pipeHandler(out, logout, outchan)
	go pipeHandler(stderr, logout, errchan)
	err = session.Wait()
	<- outchan
	<- errchan
	return err
}


func (c *ContextConnection) Close() error {
	return c.client.Close()
}

