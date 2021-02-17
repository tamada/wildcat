package wildcat

import (
	"fmt"
	"io"
	"strings"
)

// ErrorCenter collects errors.
// This type can treat as error.
type ErrorCenter struct {
	errs []error
}

// NewErrorCenter creates a new instance of ErrorCenter
func NewErrorCenter() *ErrorCenter {
	return &ErrorCenter{errs: []error{}}
}

// Push puts the given error into the receiver error center instance.
func (ec *ErrorCenter) Push(err error) bool {
	if err != nil {
		ec.errs = append(ec.errs, err)
	}
	return err != nil
}

// IsEmpty confirms the errors in the receiver error center instance is zero.
func (ec *ErrorCenter) IsEmpty() bool {
	return len(ec.errs) == 0
}

// Error returns the error messages in the receiver error center instance.
func (ec *ErrorCenter) Error() string {
	dest := new(strings.Builder)
	ec.Println(dest)
	return strings.TrimSpace(dest.String())
}

// Println prints the error messages in the receiver error center instance to the given destination.
func (ec *ErrorCenter) Println(dest io.Writer) {
	for _, err := range ec.errs {
		fmt.Fprintln(dest, err.Error())
	}
}
