package build

import (
    "fmt"
    "github.com/shankj3/ocelot/getgit"
)



func main(){
    dir, err := getgit.ShallowCloneRepo("https://github.com/kubernetes/kubernetes-anywhere.git")
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(dir)
}