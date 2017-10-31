package main

import (
    "net/http"
    
    "github.com/julienschmidt/httprouter"
    // for pretty printing objects:
    // "github.com/davecgh/go-spew/spew"
    "github.com/shankj3/ocelot/configure"
    "github.com/shankj3/ocelot/work"
)

func main() {
    run_conf := configure.GetRunConfig("/Users/jesseshank/go/src/github.com/shankj3/go-build/configure/test/test.yml")
    run.RunStage(run_conf.Build)
}