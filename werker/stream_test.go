package main

import (
	"io"
	"io/ioutil"
	"testing"
	"strings"
)

func TestGetPipe(t *testing.T) {
	// setup
	testHash := "test"
	newHash := "123123123123"
	buildContents := "onomatopoeia"
	reader := strings.NewReader(buildContents)
	readCloser := ioutil.NopCloser(reader)
	readCache := make(map[string] io.ReadCloser)
	readCache[testHash] = readCloser
	a := &appContext{readerCache: readCache,}
	//
	read, write := a.getPipe(testHash)
	// todo: figure out how to get cache to not copy the reader when it puts it in the map. should just be the first
	// memory location...
	if write != nil {
		t.Error("should have found readCloser from map, should return nil writer")
	}  //else if &read != &readCloser {
	//	t.Errorf("should have found ReadCloser from map, should have same location in memory. %s vs %s", &read, &readCloser)
	//}
	read, write = a.getPipe(newHash)
	if write == nil {
		t.Error("should have created new writecloser because hash not in cache")
	}
	if read == nil {
		t.Errorf("should have created new readCloser because hash not in cache")
	}
	// should be in cache now
	_, writecache := a.getPipe(newHash)
	if writecache != nil {
		t.Error("new hash should now be in cache, did not create write closer")
	}
	//if &readcache != &read {
	//	t.Error("new reader should have same location in memory as the first time hash was put in cache")
	//}
}