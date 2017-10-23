package main

import (
    "github.com/davecgh/go-spew/spew"
    "github.com/shankj3/go-build/configure"
)


func main() {
    run_conf := configure.GetRunConfig("/Users/jesseshank/go/src/github.com/shankj3/go-build/configure/test/test.yml")
    spew.Dump(run_conf)
}