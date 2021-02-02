package kdone

import (
	"testing"

	"github.com/go-kata/kerror"
)

type testObjectInitializer = func(t *testing.T, c *int, value int) (int, Destructor, error)

func newTestObject(t *testing.T, c *int, value int) (int, Destructor, error) {
	t.Logf("test object (%d) was initialized", value)
	return value, DestructorFunc(func() error {
		*c -= value
		t.Logf("test object (%d) was finalized", value)
		return nil
	}), nil
}

func newTestObjectWithError(t *testing.T, c *int, value int) (int, Destructor, error) {
	t.Logf("test object (%d) was initialized (but will not be finalized...)", value)
	return value, DestructorFunc(func() error {
		*c -= value
		t.Logf("test object (%d) was finalized", value)
		return nil
	}), kerror.New(kerror.ECustom, "test error")
}

func newTestObjectWithPanic(t *testing.T, c *int, value int) (int, Destructor, error) {
	t.Logf("test object (%d) was initialized (but will not be finalized...)", value)
	kerror.Panic("keep calm, this is a test panic")
	return value, DestructorFunc(func() error {
		*c -= value
		t.Logf("test object (%d) was finalized (and panic now, boo!)", value)
		return nil
	}), nil
}

func newTestObjectWithPanicInDestructor(t *testing.T, c *int, value int) (int, Destructor, error) {
	t.Logf("test object (%d) was initialized", value)
	return value, DestructorFunc(func() error {
		*c -= value
		t.Logf("test object (%d) was finalized (and panic now, boo!)", value)
		kerror.Panic("keep calm, this is a test panic")
		return nil
	}), nil
}

func newTestCompositeObject(t *testing.T, e testObjectInitializer, c *int, values ...int) (int, Destructor, error) {
	reaper := NewReaper()
	defer reaper.MustFinalize()
	sum := 0
	for _, value := range values {
		obj, dtor, err := newTestObject(t, c, value)
		if err != nil {
			return 0, nil, err
		}
		reaper.MustAssume(dtor)
		sum += obj
	}
	if e != nil {
		obj, dtor, err := e(t, c, 1000)
		if err != nil {
			return 0, nil, err
		}
		reaper.MustAssume(dtor)
		sum += obj
	}
	return sum, reaper.MustRelease(), nil
}

func TestReaper(t *testing.T) {
	var c int
	defer func() {
		if c != -6 {
			t.Fail()
			return
		}
	}()
	obj, dtor, err := newTestCompositeObject(t, nil, &c, 1, 2, 3)
	if err != nil {
		t.Fail()
		return
	}
	defer dtor.MustDestroy()
	if c != 0 {
		t.Fail()
		return
	}
	if obj != 6 {
		t.Fail()
		return
	}
}

func TestReaperWithError(t *testing.T) {
	var c int
	_, _, err := newTestCompositeObject(t, newTestObjectWithError, &c, 1, 2, 3)
	if err == nil {
		t.Fail()
		return
	}
	if c != -6 {
		t.Fail()
		return
	}
	t.Logf("\n%+v\n", err)
}

func TestReaperWithPanic(t *testing.T) {
	var c int
	defer func() {
		if c != -6 {
			t.Fail()
			return
		}
		v := recover()
		if v == nil {
			t.Fail()
			return
		}
		t.Logf("\n%+v\n", v)
	}()
	_, _, _ = newTestCompositeObject(t, newTestObjectWithPanic, &c, 1, 2, 3)
}

func TestReaperWithPanicInDestructor(t *testing.T) {
	var c int
	defer func() {
		if c != -1006 {
			t.Fail()
			return
		}
	}()
	obj, dtor, err := newTestCompositeObject(t, newTestObjectWithPanicInDestructor, &c, 1, 2, 3)
	if err != nil {
		t.Fail()
		return
	}
	defer func() {
		dtorErr := dtor.Destroy()
		if dtorErr == nil {
			t.Fail()
			return
		}
		t.Logf("\n%+v\n", dtorErr)
	}()
	if c != 0 {
		t.Fail()
		return
	}
	if obj != 1006 {
		t.Fail()
		return
	}
}

func TestReaper_AssumeNilDestructor(t *testing.T) {
	reaper := NewReaper()
	err := reaper.Assume(nil)
	t.Logf("%+v", err)
	if kerror.ClassOf(err) != kerror.EInvalid {
		t.Fail()
		return
	}
}

func TestReaper_AssumeWhenReleased(t *testing.T) {
	reaper := NewReaper()
	reaper.MustRelease()
	err := reaper.Assume(Noop)
	t.Logf("%+v", err)
	if kerror.ClassOf(err) != kerror.EIllegal {
		t.Fail()
		return
	}
}

func TestReaper_AssumeWhenFinalizer(t *testing.T) {
	reaper := NewReaper()
	reaper.MustFinalize()
	err := reaper.Assume(Noop)
	t.Logf("%+v", err)
	if kerror.ClassOf(err) != kerror.EIllegal {
		t.Fail()
		return
	}
}

func TestReaper_ReleaseWhenReleased(t *testing.T) {
	reaper := NewReaper()
	reaper.MustRelease()
	_, err := reaper.Release()
	t.Logf("%+v", err)
	if kerror.ClassOf(err) != kerror.EIllegal {
		t.Fail()
		return
	}
}

func TestReaper_ReleaseWhenFinalized(t *testing.T) {
	reaper := NewReaper()
	reaper.MustFinalize()
	_, err := reaper.Release()
	t.Logf("%+v", err)
	if kerror.ClassOf(err) != kerror.EIllegal {
		t.Fail()
		return
	}
}

func TestReaper_FinalizeWhenReleased(t *testing.T) {
	var c int
	reaper := NewReaper()
	reaper.MustAssume(DestructorFunc(func() error {
		c--
		return nil
	}))
	reaper.MustRelease()
	reaper.MustFinalize()
	if c != 0 {
		t.Fail()
		return
	}
}

func TestReaper_FinalizeWhenFinalized(t *testing.T) {
	reaper := NewReaper()
	reaper.MustFinalize()
	err := reaper.Finalize()
	t.Logf("%+v", err)
	if kerror.ClassOf(err) != kerror.EIllegal {
		t.Fail()
		return
	}
}

func TestNilReaper_Assume(t *testing.T) {
	defer func() {
		v := recover()
		t.Logf("%+v", v)
		if v == nil {
			t.Fail()
			return
		}
	}()
	_ = (*Reaper)(nil).Assume(Noop)
}
