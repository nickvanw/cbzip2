package cbzip2

import "fmt"

// BzipError represents an error returned during operation
// of bzlib. It contains a message about the attempted action
// as well as the bzlib return code.
type BzipError struct {
	ReturnCode int
	Message    string
}

func (e BzipError) Error() string {
	return fmt.Sprintf("%s: %d", e.Message, e.ReturnCode)
}
