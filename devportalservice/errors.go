package devportalservice

import "fmt"

type NetworkError struct {
	Status int
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("network request failed with status %d", e.Status)
}
