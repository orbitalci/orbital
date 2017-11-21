package main

import "errors"

type CD struct {
	dict map[string]chan string
}

func (c *CD) CarefulPut(k string, v chan string) error{
	_, ok := c.dict[k]
	if ok == true {
		return errors.New("there is already a channel for the git hash " + k)
	}
	c.dict[k] = v
	return nil
}

func NewCD() *CD {
	return &CD{
		dict: make(map[string]chan string),
	}
}
