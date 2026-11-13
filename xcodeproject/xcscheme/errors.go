package xcscheme

import "fmt"

type NotFoundError struct {
	Scheme    string
	Container string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("scheme %s not found in %s", e.Scheme, e.Container)
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(NotFoundError)
	return ok
}
