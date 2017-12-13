package builder

//TODO: think about how deployment to nexus fits in
type Builder interface {
	Setup(logout chan []byte, image string) *Result
	Before(logout chan []byte) *Result
	Build(logout chan []byte) *Result
	After(logout chan []byte) *Result
	Test(logout chan []byte) *Result
	Deploy(logout chan []byte) *Result
}

//TODO: could return even less
type Result struct {
	Status string
	Error  error
}

//TODO: move db shit out and write interface
type SQLiteDB struct{}

func (s *SQLiteDB) Connect() {}
