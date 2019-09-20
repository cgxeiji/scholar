package scholar

import (
	"bytes"
	"fmt"
)

type tnfError struct {
	requested string
}

func (e *tnfError) Error() string {
	return fmt.Sprintf("not found: type %s", e.requested)
}

// TypeNotFoundError is deprecated.
var TypeNotFoundError *tnfError

type errorOp string

type errorType uint8

const (
	// ErrNotDefined represents a not defined error.
	errNotDefined errorType = iota
	// ErrTypeNotFound represents an entry type not found error.
	ErrTypeNotFound
	// ErrFieldNotFound represents a field not found error.
	ErrFieldNotFound
)

// String implements the Stringer interface.
func (e errorType) String() string {
	switch e {
	case errNotDefined:
		return "undefined error"
	case ErrTypeNotFound:
		return "entry type not found error"
	case ErrFieldNotFound:
		return "field not found error"
	}

	return "unknown error"
}

// Err represents a custom error handler.
type Err struct {
	op    errorOp
	eType errorType
	extra string
	err   error
}

func errPad(b *bytes.Buffer, s string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(s)
}

// Error implements the error interface.
func (e *Err) Error() string {
	const sep = ": "
	const indent = ":\n\t"

	b := new(bytes.Buffer)
	if e.op != "" {
		errPad(b, sep)
		b.WriteString("[" + string(e.op) + "]")
	}
	if e.eType != errNotDefined {
		errPad(b, sep)
		b.WriteString(e.eType.String())
	}
	if e.extra != "" {
		errPad(b, indent)
		b.WriteString(" > " + e.extra)
	}
	if e.err != nil {
		errPad(b, indent)
		b.WriteString(e.err.Error())
	}

	if b.Len() == 0 {
		return "no error"
	}

	return b.String()
}

// info adds extra information to an error.
func (e *Err) info(text string) *Err {
	e.extra = text
	return e
}

func getError(op errorOp, eType errorType, err error) *Err {
	e := &Err{
		op:    op,
		eType: eType,
		err:   err,
	}

	prev, ok := e.err.(*Err)
	if !ok {
		return e
	}

	if prev.eType == e.eType {
		prev.eType = errNotDefined
	}
	if e.eType == errNotDefined {
		e.eType = prev.eType
		prev.eType = errNotDefined
	}

	if prev.op == e.op {
		prev.op = ""
	}

	return e
}

// IsError checks if err is an error of the given type.
func IsError(eType errorType, err error) bool {
	e, ok := err.(*Err)
	if !ok {
		return false
	}
	if e.eType != errNotDefined {
		return e.eType == eType
	}
	if e.err != nil {
		return IsError(eType, e.err)
	}
	return false
}
