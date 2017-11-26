package util

import "fmt"

func StrFormatErrors(testValue string, expected string, actual string) string {
	return fmt.Sprintf("expected %v to be %v, got %v", testValue, expected, actual)
}

func IntFormatErrors(testValue string, expected int, actual int) string {
	return fmt.Sprintf("expected %v to be %d, got %d", testValue, expected, actual)
}

func GenericStrFormatErrors(testValue string, expected interface{}, actual interface{}) string {
	return fmt.Sprintf("expected %v to be %v, got %v", testValue, expected, actual)
}