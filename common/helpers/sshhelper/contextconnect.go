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
	port      int
}

func InitContextConnect(keyFp, user, host string, port int) *ContextConnection {
	return &ContextConnection{connectionVars: &connectionVars{privKey: keyFp, user: user, host:host, port: port}}
}

// Create will attempt to create an ssh client with the parameters provided here, and it will also kick off a
//   goroutine that will handle the context cancellation
func (c *ContextConnection) Connect(ctx context.Context) error {
	cliConf, err := getClientConfig(c.user, c.privKey)
	if err != nil {
		return err
	}
	sshCli, err := GetSSHCli(c.host, fmt.Sprintf("%d", c.port), cliConf)
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
		envArray := strings.Split(env, "=")
		if len(envArray) != 2 {
			// todo: there is a better way to handle this
			panic(fmt.Sprintf("env defined as %s has more than one '='", env))
		}
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
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	err = session.Run("echo 'hi'")
	return err
}

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


func (c *ContextConnection) RunAndLog(cmd string, envs []string, logout chan []byte,  pipeHandler func(rc io.Reader, logout chan []byte)) error {
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
	session.Start(cmd)
	go pipeHandler(out, logout)
	go pipeHandler(stderr, logout)
	err = session.Wait()
	return err
}


func (c *ContextConnection) Close() error {
	return c.client.Close()
}

func getClientConfig(user string, file string) (*ssh.ClientConfig, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return sshConfig, nil
}

func GetSSHCli(host, port string, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	// todo: not sure if this connection should be persisted long term? idk; keep an eye on leaking connections i guess.
	connection, err := ssh.Dial("tcp", host+":"+port, sshConfig)
	if err != nil {
		return nil, errors.New("Unable to get connection, error is: " + err.Error())
	}
	return connection, nil
}
