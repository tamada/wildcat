package wildcat

import (
	"fmt"
	"io"
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

func (ec *ErrorCenter) Println(dest io.Writer) {
	for _, err := range ec.errs {
		fmt.Fprintln(dest, err.Error())
	}
}
