package buildruntime

import (
	"bitbucket.org/level11consulting/go-til/test"
	"testing"
)

func Test_generateBuildPath(t *testing.T) {
	werkerId := "1234-8974-3818-asdf"
	gitHash := "123nd9xz"
	expectedPath := "ci/builds/1234-8974-3818-asdf/123nd9xz"
	path := MakeBuildPath(werkerId, gitHash)
	if path != expectedPath {
		test.StrFormatErrors("build runtime path", expectedPath, path)
	}
}


func Test_generateBuildMapPath(t *testing.T) {
	gitHash := "123123"
	expectedPath := "ci/werker_build_map/123123"
	path := MakeBuildMapPath(gitHash)
	if path != expectedPath {
		test.StrFormatErrors("werker build map path", expectedPath, path)
	}
}

func Test_generateWerkerLocPath(t *testing.T) {
	werkerId := "werkeridnum"
	expectedPath := "ci/werker_location/werkeridnum"
	path := MakeWerkerLocPath(werkerId)
	if expectedPath != path {
		test.StrFormatErrors("werker location path", expectedPath, path)
	}
}

func Test_parseGenericBuildPath(t *testing.T) {
	expectWerkerId := "1234-8974-3818-asdf"
	expectGitHash := "123nd9xz"
	herepathbe := "ci/builds/1234-8974-3818-asdf/123nd9xz/docker_uuid"
	werkerId, gitHash, _ := parseGenericBuildPath(herepathbe)
	if werkerId != expectWerkerId {
		test.StrFormatErrors("werker id ", expectWerkerId, werkerId)
	}
	if gitHash != expectGitHash {
		test.StrFormatErrors("git hash", expectGitHash, gitHash)
	}
	herepathbe2 := "ci/builds/1234-8974-3818-asdf/123nd9xz/docker_uuid/"
	werkerId, gitHash, _ = parseGenericBuildPath(herepathbe2)
	if werkerId != expectWerkerId {
		test.StrFormatErrors("werker id ", expectWerkerId, werkerId)
	}
	if gitHash != expectGitHash {
		test.StrFormatErrors("git hash", expectGitHash, gitHash)
	}
	herepathbe3 := "ci/builds/1234-8974-3818-asdf/123nd9xz"
	werkerId, gitHash, _ = parseGenericBuildPath(herepathbe3)
	if werkerId != expectWerkerId {
		test.StrFormatErrors("werker id ", expectWerkerId, werkerId)
	}
	if gitHash, _ != expectGitHash {
		test.StrFormatErrors("git hash", expectGitHash, gitHash)
	}
}

func Test_parseWerkerLocPath(t *testing.T) {
	expWerkerId := "<werkerId>"
	path := "ci/werker_location/<werkerId>/werker_ip"
	werkerId := parseWerkerLocPath(path)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
	path2 := path + "/"
	werkerId = parseWerkerLocPath(path2)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
	path3 := "ci/werker_location/<werkerId>"
	werkerId = parseWerkerLocPath(path3)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
	path4 := "ci/werker_location/<werkerId>/"
	werkerId = parseWerkerLocPath(path4)
	if expWerkerId != werkerId {
		test.StrFormatErrors("werker id", expWerkerId, werkerId)
	}
}

func Test_parseBuildMapPath(t *testing.T) {
	path := "ci/werker_build_map/<hash>"
	expHash := "<hash>"
	hash := parseBuildMapPath(path)
	if expHash != hash {
		test.StrFormatErrors("hash", expHash, hash)
	}
}
