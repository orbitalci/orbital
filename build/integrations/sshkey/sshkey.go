package sshkey

import (
	"fmt"
	"strings"

	"bitbucket.org/level11consulting/ocelot/build/integrations"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

type SSHKeyInt struct {
	sshKeys map[string]string
}

func (n *SSHKeyInt) String() string {
	return "ssh keyfile integration"
}

func (n *SSHKeyInt) SubType() pb.SubCredType {
	return pb.SubCredType_SSHKEY
}

func Create() integrations.StringIntegrator {
	return &SSHKeyInt{}
}

func (n *SSHKeyInt) GetEnv() []string {
	var envs []string
	for name, value := range n.sshKeys {
		env := fmt.Sprintf("%s=%s", name, value)
		envs = append(envs, env)
	}
	return envs
}

func (n *SSHKeyInt) GenerateIntegrationString(credz []pb.OcyCredder) (string, error) {
	//var sshCreds []*pb.SSHKeyWrapper
	var sshkeys = make(map[string]string)
	for _, credi := range credz {
		sshkeys[credi.GetIdentifier()] = credi.GetClientSecret()
	}
	n.sshKeys = sshkeys
	return "", nil
}

func (n *SSHKeyInt) MakeBashable(str string) []string {
	var cmds []string
	for identifier, _ := range n.sshKeys {
		cmd := fmt.Sprintf("echo ${%s} > ~/.ssh/%s && chmod 600 ~/.ssh/%s", identifier, identifier, identifier)
		cmds = append(cmds, cmd)
	}
	return []string{"/bin/sh", "-c", strings.Join(cmds, " && ")}
}

func (n *SSHKeyInt) IsRelevant(wc *pb.BuildConfig) bool {
	return true
}