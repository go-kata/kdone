package kdone

// CloserFunc represents a functional implementation of the io.Closer interface.
type CloserFunc func() error

// Close implements the io.Closer interface.
func (f CloserFunc) Close() error {
	return f()
}
