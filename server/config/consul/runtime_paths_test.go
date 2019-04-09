package consul

import (
	"testing"

	"github.com/shankj3/go-til/test"
)

var idAndHashTests = []struct {
	field     string
	werkerId  string
	gitHash   string
	buildFunc func(string, string) string
	expected  string
}{
	{"build path", "1234-8974-3818-asdf", "123nd9xz", MakeBuildPath, "ci/builds/1234-8974-3818-asdf/123nd9xz"},
	{"build summary", "123123123123", "hashhashhash", MakeBuildSummaryIdPath, "ci/builds/123123123123/hashhashhash/build_id"},
	{"build stage", "123123123123", "hashhashhash", MakeBuildStagePath, "ci/builds/123123123123/hashhashhash/current_stage"},
	{"build start time", "123123123123", "hashhashhash", MakeBuildStartpath, "ci/builds/123123123123/hashhashhash/start_time"},
	{"build docker uuid", "123123123123", "hashhashhash", MakeDockerUuidPath, "ci/builds/123123123123/hashhashhash/docker_uuid"},
}

var singleFieldTests = []struct {
	field      string
	fieldvalue string
	buildFunc  func(string) string
	expected   string
}{
	{"werker id path", "werkerwerks", MakeBuildWerkerIdPath, "ci/builds/werkerwerks"},
	{"build map path", "hashyhashy", MakeBuildMapPath, "ci/werker_build_map/hashyhashy"},
	{"werker loc path", "werkerid", MakeWerkerLocPath, "ci/werker_location/werkerid"},
	{"werker ip path", "werkerid", MakeWerkerIpPath, "ci/werker_location/werkerid/werker_ip"},
	{"werker grpc path", "werkerid", MakeWerkerGrpcPath, "ci/werker_location/werkerid/werker_grpc_port"},
	{"werker ws path", "werkerid", MakeWerkerWsPath, "ci/werker_location/werkerid/werker_ws_port"},
}

func Test_IdAndHashBuilders(t *testing.T) {
	for _, tst := range idAndHashTests {
		path := tst.buildFunc(tst.werkerId, tst.gitHash)
		if path != tst.expected {
			t.Error(test.StrFormatErrors(tst.field, tst.expected, path))
		}
	}
}

func Test_singleFieldTests(t *testing.T) {
	for _, tst := range singleFieldTests {
		path := tst.buildFunc(tst.fieldvalue)
		if path != tst.expected {
			t.Error(test.StrFormatErrors(tst.field, tst.expected, path))
		}
	}
}

func Test_MakeBuildPath(t *testing.T) {
	werkerId := "1234-8974-3818-asdf"
	gitHash := "123nd9xz"
	expectedPath := "ci/builds/1234-8974-3818-asdf/123nd9xz"
	path := MakeBuildPath(werkerId, gitHash)
	if path != expectedPath {
		t.Error(test.StrFormatErrors("build runtime path", expectedPath, path))
	}
}

func Test_MakeBuildMapPath(t *testing.T) {
	gitHash := "123123"
	expectedPath := "ci/werker_build_map/123123"
	path := MakeBuildMapPath(gitHash)
	if path != expectedPath {
		t.Error(test.StrFormatErrors("werker build map path", expectedPath, path))
	}
}

func Test_MakeWerkerLocPath(t *testing.T) {
	werkerId := "werkeridnum"
	expectedPath := "ci/werker_location/werkeridnum"
	path := MakeWerkerLocPath(werkerId)
	if expectedPath != path {
		t.Error(test.StrFormatErrors("werker location path", expectedPath, path))
	}
}

func Test_parseGenericBuildPath(t *testing.T) {
	expectWerkerId := "1234-8974-3818-asdf"
	expectGitHash := "123nd9xz"
	herepathbe := "ci/builds/1234-8974-3818-asdf/123nd9xz/docker_uuid"
	werkerId, gitHash, _ := ParseGenericBuildPath(herepathbe)
	if werkerId != expectWerkerId {
		test.StrFormatErrors("werker id ", expectWerkerId, werkerId)
	}
	if gitHash != expectGitHash {
		test.StrFormatErrors("git hash", expectGitHash, gitHash)
	}
	herepathbe2 := "ci/builds/1234-8974-3818-asdf/123nd9xz/docker_uuid/"
	werkerId, gitHash, _ = ParseGenericBuildPath(herepathbe2)
	if werkerId != expectWerkerId {
		test.StrFormatErrors("werker id ", expectWerkerId, werkerId)
	}
	if gitHash != expectGitHash {
		test.StrFormatErrors("git hash", expectGitHash, gitHash)
	}
	herepathbe3 := "ci/builds/1234-8974-3818-asdf/123nd9xz/SPECIAL_KEY"
	werkerId, gitHash, _ = ParseGenericBuildPath(herepathbe3)
	if werkerId != expectWerkerId {
		test.StrFormatErrors("werker id ", expectWerkerId, werkerId)
	}
	if gitHash != expectGitHash {
		test.StrFormatErrors("git hash", expectGitHash, gitHash)
	}
}

func Test_parseWerkerLocPath(t *testing.T) {
	expWerkerId := "<werkerId>"
	path := "ci/werker_location/<werkerId>/werker_ip"
	werkerId := ParseWerkerLocPath(path)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
	path2 := path + "/"
	werkerId = ParseWerkerLocPath(path2)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
	path3 := "ci/werker_location/<werkerId>"
	werkerId = ParseWerkerLocPath(path3)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
	path4 := "ci/werker_location/<werkerId>/"
	werkerId = ParseWerkerLocPath(path4)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
}

func Test_parseBuildMapPath(t *testing.T) {
	path := "ci/werker_build_map/<hash>"
	expHash := "<hash>"
	hash := ParseBuildMapPath(path)
	if expHash != hash {
		test.StrFormatErrors("hash", expHash, hash)
	}
}
