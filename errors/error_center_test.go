package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestErrorCenter(t *testing.T) {
	ec := New()
	ec.Push(nil)
	ec.Push(fmt.Errorf("error1"))
	ec.Push(fmt.Errorf("error2"))

	if ec.Size() != 2 {
		t.Errorf("ec.Size() did not match, wont %d, got %d", 2, ec.Size())
	}

	if ec.IsEmpty() {
		t.Errorf("ec did not empty, should one error")
	}

	if strings.TrimSpace(ec.Error()) != `error1
error2` {
		t.Errorf("ec.Error() did not match, wont error1\nerror2, got %s", ec.Error())
	}
}

func TestErrorCenter2(t *testing.T) {
	ec1 := New()
	ec1.Push(fmt.Errorf("error1"))
	ec1.Push(fmt.Errorf("error2"))
	ec2 := New()
	ec2.Push(fmt.Errorf("error3"))
	ec2.Push(fmt.Errorf("error4"))
	ec1.Push(ec2)

	if ec1.Size() != 4 {
		t.Errorf("ec1.Size() did not match wont %d, got %d", 4, ec1.Size())
	}

	if strings.TrimSpace(ec1.Error()) != `error1
error2
error3
error4` {
		t.Errorf("ec1.Error() did not match, wont error1\nerror2\nerror3\nerror4, got %s", ec1.Error())
	}
}
