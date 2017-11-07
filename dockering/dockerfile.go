/*
Type / Methods for Dockerfile for converting struct to Dockerfile.
The output file will include the package manager that corresponds with the o.s.
*/

package dockering

import (
    "fmt"
    "io/ioutil"
    "path"
    "strings"
)

type OS int

// enum of operating systems (`type OS int`)
const (
    Debian OS = iota
    Fedora
    Alpine
)

// Object for dockerfile data / metadata
type Dockerfile struct {
    image       string
    packageList []string
    fileName    string
    os          OS
}

/* return string contents of simple docker file that
uses df.image as base and installs the packageList using
the package manager that corresponds to OS */
func (df *Dockerfile) DockerString() string {
    packages := strings.Join(df.packageList, " ")
    var package_install string
    switch df.os {
    case Debian:
        package_install = "apt-get update && apt-get install --no-install-recommends -y"
    case Fedora:
        package_install = "yum update && yum install -y"
    case Alpine:
        package_install = "apk add --no-cache"
    }
    return fmt.Sprintf("FROM %s\nRUN %s %s", df.image, package_install, packages)
}

func (df *Dockerfile) WriteDockerFile() error {
    dockerFileLocation := "/tmp/dockering" // todo: parametrizable? eh?
    tmpfile, err := ioutil.TempFile(dockerFileLocation, "Dockerfile")
    if err != nil {
        return err
    }
    defer tmpfile.Close()
    if _, err = tmpfile.WriteString(df.DockerString()); err != nil {
        return err
    }
    df.fileName = path.Join(dockerFileLocation, tmpfile.Name())
    return nil
}
