package common

import "strings"

// helper
func GetAcctRepo(fullName string) (acct string, repo string, err error) {
	list := strings.Split(fullName, "/")
	if len(list) != 2 {
		return "","", BadFormat("acctRepo needs to be in format acct/repo")
	}
	acct = list[0]
	repo = list[1]
	return
}


func BadFormat(msg string) *FormatError {
	return &FormatError{err:msg}
}

type FormatError struct {
	err string
}

func (f *FormatError) Error() string {
	return f.err
}
