package zip

type Error struct {
	err error
}

func NewError(err error) Error {
	return Error{err: err}
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) Unwrap() error {
	return e.err
}
