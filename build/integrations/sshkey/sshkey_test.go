package sshkey

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/models/pb"
)

var sshKeys = []pb.OcyCredder{
	&pb.SSHKeyWrapper{
		AcctName: "level11orwhatever",
		PrivateKey: []byte(`-------thisisprivatekeydoyouhearme----
aklfj;osdfiulx,cmnfg
asdf;aseiurqawD83mnvc8aed
asdifu3nazlxci7ensk
AALIW3UYBCUAW6129394
-- END PRIVATE KEY OR WHATEVER -- 
`),
		SubType:    pb.SubCredType_SSHKEY,
		Identifier: "id1OCELOTTEST",
	},
	&pb.SSHKeyWrapper{
		AcctName: "level11orwhatever",
		PrivateKey: []byte(`-------thisisprivatekeydoyouhearme----
aklfj;osdf5468572iulx,cmnfg
asdfd8a3b7;aseiurqawD83mnvc8aed
asdifu3nazlxci7ensk
AALIW3UYBCUAW6129394
-- END PRIVATE KEas6d4f7eY OR WHATEVER -- 
`),
		SubType:    pb.SubCredType_SSHKEY,
		Identifier: "id2OCELOTTEST",
	},
}

func TestSSHKeyInt_GetEnv(t *testing.T) {
	sshInt := &SSHKeyInt{}
	_, err := sshInt.GenerateIntegrationString(sshKeys)
	if err != nil {
		t.Error(err)
	}
	envs := sshInt.GetEnv()
	var expectedEnvs = []string{
		`id1OCELOTTEST=-------thisisprivatekeydoyouhearme----
aklfj;osdfiulx,cmnfg
asdf;aseiurqawD83mnvc8aed
asdifu3nazlxci7ensk
AALIW3UYBCUAW6129394
-- END PRIVATE KEY OR WHATEVER -- 
`,
		`id2OCELOTTEST=-------thisisprivatekeydoyouhearme----
aklfj;osdf5468572iulx,cmnfg
asdfd8a3b7;aseiurqawD83mnvc8aed
asdifu3nazlxci7ensk
AALIW3UYBCUAW6129394
-- END PRIVATE KEas6d4f7eY OR WHATEVER -- 
`,
	}
	// have to do this iterate bs because getEnv uses a map, and that isn't ordered
	for _, env := range expectedEnvs {
		var found bool
		for _, livenv := range envs {
			if env == livenv {
				found = true
			}
		}
		if found == false {
			t.Errorf("could not find env var\n %s\n in list of live envs", env)
		}
	}
}

func TestSSHKeyInt_MakeBashable(t *testing.T) {
	// we don't want to add to every testers' strict host key checking
	sshInt := &SSHKeyInt{strictHostKey: "echo testing"}
	_, err := sshInt.GenerateIntegrationString(sshKeys)
	if err != nil {
		t.Error(err)
	}
	cmds := sshInt.MakeBashable("")
	var stderr, stdout bytes.Buffer
	cmd := exec.Command(cmds[0], cmds[1], cmds[2])
	cmd.Env = sshInt.GetEnv()
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	t.Log(stdout.String())
	if err := cmd.Run(); err != nil {
		t.Log(stderr.String())
		t.Error(err)
	}
	defer func() {
		t.Log("getting rid of test rendered files")
		os.Remove(os.ExpandEnv("$HOME/.ssh/id2OCELOTTEST"))
		os.Remove(os.ExpandEnv("$HOME/.ssh/id1OCELOTTEST"))
	}()
	if id2, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.ssh/id2OCELOTTEST")); err != nil {
		t.Error(err)
	} else {
		expected := sshKeys[1].GetClientSecret() + "\n"
		live := string(id2)
		if diff := deep.Equal(live, expected); diff != nil {
			t.Error(diff)
		}
	}

}

func TestSSHKeyInt_staticstuff(t *testing.T) {
	ssh := Create()
	if ssh.String() != "ssh keyfile integration" {
		t.Error("string() of ssh should be 'ssh keyfile integration'")
	}
	if ssh.SubType() != pb.SubCredType_SSHKEY {
		t.Error("subcredtype of ssh is SSHKEY")
	}
	ssher := ssh.(*SSHKeyInt)
	if ssher.strictHostKey != "mkdir -p ~/.ssh && echo \"StrictHostKeyChecking no\" >> ~/.ssh/config && chmod 400 ~/.ssh/config" {
		t.Error(test.StrFormatErrors("first hostkey script", "mkdir -p ~/.ssh && echo \"StrictHostKeyChecking no\" >> ~/.ssh/config && chmod 400 ~/.ssh/config", ssher.strictHostKey))
	}
	if !ssh.IsRelevant(&pb.BuildConfig{}) {
		t.Error("ssh key integration is always relevant.")
	}
}
