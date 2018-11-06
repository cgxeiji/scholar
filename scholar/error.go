package scholar

import "fmt"

type tnfError struct {
	requested string
}

func (e *tnfError) Error() string {
	return fmt.Sprintf("not found: type %s", e.requested)
}

var TypeNotFoundError *tnfError
