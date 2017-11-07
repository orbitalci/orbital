package ocenet

//HttpError is the format that errors should be returned from all REST api's
type HttpError struct {
	Status int
	Error  string
}
