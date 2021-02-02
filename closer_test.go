package kdone

import (
	"testing"

	"github.com/go-kata/kerror"
)

func TestCloserFunc(t *testing.T) {
	var c int
	if err := CloserFunc(func() error {
		c--
		return nil
	}).Close(); err != nil {
		t.Logf("%+v", err)
		t.Fail()
		return
	}
	if c != -1 {
		t.Fail()
		return
	}
}

func TestCloserFuncWithError(t *testing.T) {
	err := CloserFunc(func() error {
		return kerror.New(nil, "test error")
	}).Close()
	t.Logf("%+v", err)
	if err == nil {
		t.Fail()
		return
	}
}
