package wildcat

import (
	"fmt"
	"io"
	"strings"
)

type ErrorCenter struct {
	errs []error
}

func NewErrorCenter() *ErrorCenter {
	return &ErrorCenter{errs: []error{}}
}

func (ec *ErrorCenter) Push(err error) bool {
	if err != nil {
		ec.errs = append(ec.errs, err)
	}
	return err != nil
}

func (ec *ErrorCenter) IsEmpty() bool {
	return len(ec.errs) == 0
}

func (ec *ErrorCenter) Error() string {
	dest := new(strings.Builder)
	ec.Println(dest)
	return dest.String()
}

func (ec *ErrorCenter) Println(dest io.Writer) {
	for _, err := range ec.errs {
		fmt.Fprintln(dest, err.Error())
	}
}
