package kdone

import "testing"

func TestClosable(t *testing.T) {
	var c int
	dtor := DestructorFunc(func() error {
		c++
		return nil
	})
	closer := Closable(dtor)
	if err := closer.Close(); err != nil {
		t.Fail()
		return
	}
	if c != 1 {
		t.Fail()
		return
	}
}
