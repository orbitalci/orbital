package serde

import (
	"encoding/base64"
)

func BitzToBase64(bits []byte) string {
	return base64.StdEncoding.EncodeToString(bits)
}

func StrToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64ToBitz(b64string string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(b64string)
}
