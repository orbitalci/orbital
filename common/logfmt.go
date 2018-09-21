package common

import "regexp"

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
const gradle =`.* \| <[=|-]+> [0-9]+% (CONFIGURING|EXECUTING|INITIALIZING|WAITING) \[([0-9]+m )?[0-9]+s\]>.*\n`


var re = regexp.MustCompile(ansi)
var regrad = regexp.MustCompile(gradle)

func MaybeStrip(output []byte, stripAnsi bool) []byte {
	if stripAnsi {
		return regrad.ReplaceAll(re.ReplaceAll(output, []byte("")), []byte(""))
		//return regrad.ReplaceAllString(re.ReplaceAllString(str, ""), "")
		//return regrad.ReplaceAllString(str, "")
		//return str
	}
	return output
}
