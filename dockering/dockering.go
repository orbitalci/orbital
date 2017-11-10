package dockering

import (
    "archive/tar"
    "bytes"
    "context"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
    "github.com/shankj3/ocelot/util/ocelog"
    "io/ioutil"
    "os"
    "regexp"
)

var DockerRegex = regexp.MustCompile("Successfully built ([0-9a-z0-9]+)")
var ctx = context.Background()

/*
Write simple Dockerfile with only::
```
FROM <base image>
RUN <update packages and install package list>
```

for example:
```
FROM debian:jessie
RUN apt-get update && apt-get install -y curl
```

Will return filepath of Dockerfile or any error generated
*/

// Build Docker image from a dockerfile filepath.
// will return byte array of the id of image built or an error.
func BuildImageFromDockerFile(filePath string) ([]byte, error) {
    cli, err := client.NewEnvClient()
    if err != nil {
        ocelog.LogErrField(err).Warn("unable to init docker client")
        return nil, err
    }
    buf := new(bytes.Buffer)
    tw := tar.NewWriter(buf)
    defer tw.Close()

    dockerFile := "Dockerfile"
    dockerFileReader, err := os.Open(filePath)
    dockerfileContents, readerr := ioutil.ReadAll(dockerFileReader)
    if err != nil || readerr != nil {
        ocelog.LogErrField(err).WithField("filePath", filePath).Warn("unable to open or read Dockerfile")
        return nil, err
    }
    tarHeader := &tar.Header{
        Name: dockerFile,
        Size: int64(len(dockerfileContents)),
    }
    if err = tw.WriteHeader(tarHeader); err != nil {
        ocelog.LogErrField(err).Warn("Unable to write tar header")
        return nil, err
    }
    if _, err = tw.Write(dockerfileContents); err != nil {
        ocelog.LogErrField(err).Warn("unable to write tar body")
        return nil, err
    }
    dockerFileTarReader := bytes.NewReader(buf.Bytes())

    imageBuildResponse, err := cli.ImageBuild(
        ctx,
        dockerFileTarReader,
        types.ImageBuildOptions{
            Context:    dockerFileTarReader,
            Dockerfile: dockerFile,
            Remove:     true,
        },
    )
    if err != nil {
        ocelog.LogErrField(err).Warn("unable to build docker image")
        return nil, err
    }
    defer imageBuildResponse.Body.Close()
    responseLines, _ := ioutil.ReadAll(imageBuildResponse.Body)
    imageIdMatch := DockerRegex.FindSubmatch(responseLines)
    return imageIdMatch[1], nil
}
