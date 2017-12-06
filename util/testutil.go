package util

import (
	"bytes"
	"fmt"
	"testing"
)

func StrFormatErrors(testValue string, expected string, actual string) string {
	return fmt.Sprintf("expected %v to be %v, got %v", testValue, expected, actual)
}

func IntFormatErrors(testValue string, expected int, actual int) string {
	return fmt.Sprintf("expected %v to be %d, got %d", testValue, expected, actual)
}

func GenericStrFormatErrors(testValue string, expected interface{}, actual interface{}) string {
	return fmt.Sprintf("expected %v to be %v, got %v", testValue, expected, actual)
}

func NotNull(t *testing.T, theThing interface{}) {
	if theThing == nil {
		t.Error(GenericStrFormatErrors(theThing.(string), nil, theThing))
	}
}

// todo: write string function for printing out errors
func CompareByteArrays(a [][]byte, b [][]byte) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for index, elem := range a {
		if !bytes.Equal(elem, b[index]) {
			return false
		}
	}
	return true
}

func CompareStringArrays(a []string, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for ind, elem := range a {
		if elem != b[ind] {
			return false
		}
	}
	return true
}