package main

import (
	"github.com/shankj3/ocelot/util/ocelog"
	"io"
	"sync"
)

type ReaderCache struct {
	dict map[string]*io.PipeReader
	mux sync.Mutex
}

func (c *ReaderCache) CarefulPut(k string, v *io.PipeReader) error{
	ocelog.Log().Debugf("put reader with address %v and hash %s", v, k)
	c.mux.Lock()
	c.dict[k] = v
	c.mux.Unlock()
	return nil
}

// CarefulValue returns a copy of the io.PipeReader
func (c *ReaderCache) CarefulValue(k string) (pipeCopy io.PipeReader, ok bool){
	c.mux.Lock()
	defer c.mux.Unlock()
	elem, ok := c.dict[k]
	if ok {
		ocelog.Log().Debugf("got reader with address %v and hash %s", elem, k)
		pipeCopy = *elem
		return
	}
	return
}

func (c *ReaderCache) CarefulRm(k string){
	c.mux.Lock()
	delete(c.dict, k)
	c.mux.Unlock()

}

func NewReaderCache() *ReaderCache {
	return &ReaderCache{
		dict: make(map[string]*io.PipeReader),
	}
}