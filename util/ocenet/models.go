package ocenet

//HttpError is the format that errors should be returned from all REST api's - see usage in admin.go
type ApiHttpError struct {
	ErrorDescription string
	Error  string
}

