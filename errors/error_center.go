package errors

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Center collects errors.
// This type can treat as error.
type Center struct {
	errs []error
}

// New creates a new instance of ErrorCenter
func New() *Center {
	return &Center{errs: []error{}}
}

// Size returns the size of errors.
func (ec *Center) Size() int {
	return len(ec.errs)
}

// Push puts the given error into the receiver error center instance.
func (ec *Center) Push(err error) bool {
	var otherCenter *Center
	if err != nil {
		if errors.As(err, &otherCenter) {
			ec.errs = append(ec.errs, otherCenter.errs...)
		} else {
			ec.errs = append(ec.errs, err)
		}
	}
	return err != nil
}

// IsEmpty confirms the errors in the receiver error center instance is zero.
func (ec *Center) IsEmpty() bool {
	return len(ec.errs) == 0
}

// Error returns the error messages in the receiver error center instance.
func (ec *Center) Error() string {
	dest := new(strings.Builder)
	ec.Println(dest)
	return strings.TrimSpace(dest.String())
}

// Println prints the error messages in the receiver error center instance to the given destination.
func (ec *Center) Println(dest io.Writer) {
	for _, err := range ec.errs {
		fmt.Fprintln(dest, err.Error())
	}
}
