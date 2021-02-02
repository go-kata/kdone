package kdone

import (
	"testing"

	"github.com/go-kata/kerror"
)

func TestDestructorFunc(t *testing.T) {
	var c int
	if err := DestructorFunc(func() error {
		c--
		return nil
	}).Destroy(); err != nil {
		t.Logf("%+v", err)
		t.Fail()
		return
	}
	if c != -1 {
		t.Fail()
		return
	}
}

func TestDestructorFunc_DestroyWithError(t *testing.T) {
	err := DestructorFunc(func() error {
		return kerror.New(nil, "test error")
	}).Destroy()
	t.Logf("%+v", err)
	if err == nil {
		t.Fail()
		return
	}
}

func TestDestructorFunc_MustDestroyWithError(t *testing.T) {
	defer func() {
		v := recover()
		t.Logf("%+v", v)
		if v == nil {
			t.Fail()
			return
		}
	}()
	DestructorFunc(func() error {
		return kerror.New(nil, "test error")
	}).MustDestroy()
}

func TestNoop(t *testing.T) {
	if err := Noop.Destroy(); err != nil {
		t.Logf("%+v", err)
		t.Fail()
		return
	}
}
