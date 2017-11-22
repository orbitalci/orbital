package main

import (
	"errors"
	"sync"
)

type CD struct {
	dict map[string]chan []byte
	mux sync.Mutex
}

func (c *CD) CarefulPut(k string, v chan []byte) error{
	_, ok := c.CarefulValue(k)
	if ok == true {
		return errors.New("there is already a channel for the git hash " + k)
	}
	c.mux.Lock()
	c.dict[k] = v
	c.mux.Unlock()
	return nil
}

func (c *CD) CarefulValue(k string)  (chan []byte, bool){
	c.mux.Lock()
	defer c.mux.Unlock()
	elem, ok := c.dict[k]
	return elem, ok
}

func NewCD() *CD {
	return &CD{
		dict: make(map[string]chan []byte),
	}
}
