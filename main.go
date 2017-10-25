package main

import (
    "net/http"
    
    "github.com/julienschmidt/httprouter"
    // for pretty printing objects:
    // "github.com/davecgh/go-spew/spew"
    "github.com/shankj3/go-build/configure"
    "github.com/shankj3/go-build/work"
)

func main() {
    run_conf := configure.GetRunConfig("/Users/jesseshank/go/src/github.com/shankj3/go-build/configure/test/test.yml")
    run.RunStage(run_conf.Build)
}