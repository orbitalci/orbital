package du

import (
	"testing"
)

func TestSpace(t *testing.T) {
	badFp := "/asdfjkhasdf23flhksd7sbz111xcjv732987w64312na9xzxcfr2w"
	_, _, err := Space(badFp)
	if err == nil {
		t.Error(badFp, "does not exist, this should error")
	}
	total, free, err := Space("/")
	t.Log(free)
	if err != nil {
		t.Error(err)
	}
	if total == 0 {
		t.Error("total disk space would not be 0, this can't be right.")
	}

}
