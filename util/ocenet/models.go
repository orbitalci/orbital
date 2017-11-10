package ocenet

//HttpError is the format that errors should be returned from all REST api's - see usage in admin.go
type HttpError struct {
	Status int
	Error  string
	ErrorDetail	string
}
