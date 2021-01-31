package kdone

import "io"

// CloserFunc represents a functional implementation of the io.Closer interface.
type CloserFunc func() error

// Close implements the io.Closer interface.
func (f CloserFunc) Close() error {
	return f()
}

// Closable casts the given destructor to the io.Closer interface.
func Closable(dtor Destructor) io.Closer {
	return CloserFunc(dtor.Destroy)
}
