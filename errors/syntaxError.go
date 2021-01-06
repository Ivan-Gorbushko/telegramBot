package TBErrors

import "fmt"

type SyntaxError struct {
	Line int
	Col  int
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%d:%d: syntax error", e.Line, e.Col)
}
