package TBErrors

import "fmt"

type InternalError struct {
	Path string
}
func (e *InternalError) Error() string {
	return fmt.Sprintf("parse %v: internal error", e.Path)
}