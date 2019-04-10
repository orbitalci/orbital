package stringbuilder

func NCErr(msg string) *NoCreds {
	return &NoCreds{msg: msg}
}

type NoCreds struct {
	msg string
}

func (n *NoCreds) Error() string {
	return n.msg
}
