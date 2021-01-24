package kdone

// Destructor represents an object destructor.
//
// The usual identifier for variables of this type is dtor.
type Destructor interface {
	// DestroyStack destroys an associated object.
	Destroy() error
	// MustDestroy is a variant of the Destroy that panics on error.
	MustDestroy()
}

// DestructorFunc represents a functional object destructor.
type DestructorFunc func() error

// Destroy implements the Destructor interface.
func (f DestructorFunc) Destroy() error {
	return f()
}

// MustDestroy implements the Destructor interface.
func (f DestructorFunc) MustDestroy() {
	if err := f(); err != nil {
		panic(err)
	}
}

// Noop specifies the destructor which does nothing.
var Noop = DestructorFunc(func() error {
	return nil
})
