package kdone

import "github.com/go-kata/kerror"

// Reaper represents a mechanism of deferred call of destructors.
type Reaper struct {
	// destructors specifies destructors the responsibility of calling which was assumed by this reaper.
	destructors []Destructor
	// released specifies whether was this reaper released
	// from the responsibility for calling destructors.
	released bool
	// finalized specifies whether did this reapers call destructors.
	finalized bool
}

// NewReaper returns a new reaper.
func NewReaper() *Reaper {
	return &Reaper{}
}

// Assume passes the responsibility for calling the given destructor to this reaper.
func (r *Reaper) Assume(dtor Destructor) error {
	if r == nil {
		kerror.NPE()
		return nil
	}
	if r.released {
		return kerror.New(kerror.EIllegal,
			"reaper was already released from responsibility for calling destructors")
	}
	if r.finalized {
		return kerror.New(kerror.EIllegal, "reaper has already called destructors")
	}
	if dtor == nil {
		return kerror.New(kerror.EInvalid, "reaper cannot assume responsibility for calling nil destructor")
	}
	r.destructors = append(r.destructors, dtor)
	return nil
}

// MustAssume is a variant of the Assume that panics on error.
func (r *Reaper) MustAssume(dtor Destructor) {
	if err := r.Assume(dtor); err != nil {
		panic(err)
	}
}

// Release releases this reaper from the responsibility for calling destructors
// and pass it to caller by return a composite destructor that calls destructors
// in the backward order.
func (r *Reaper) Release() (Destructor, error) {
	if r == nil {
		return Noop, nil
	}
	if r.released {
		return nil, kerror.New(kerror.EIllegal,
			"reaper was already released from responsibility for calling destructors")
	}
	if r.finalized {
		return nil, kerror.New(kerror.EIllegal, "reaper has already called destructors")
	}
	destructors := r.destructors
	dtor := DestructorFunc(func() error {
		return reap(destructors...)
	})
	r.released = true
	return dtor, nil
}

// MustRelease is a variant of the Release that panics on error.
func (r *Reaper) MustRelease() Destructor {
	dtor, err := r.Release()
	if err != nil {
		panic(err)
	}
	return dtor
}

// Finalize calls destructors in the backward order
// if this reaper was not released from this responsibility yet.
func (r *Reaper) Finalize() error {
	if r == nil || r.released {
		return nil
	}
	if r.finalized {
		return kerror.New(kerror.EIllegal, "reaper has already called destructors")
	}
	err := reap(r.destructors...)
	r.finalized = true
	return err
}

// MustFinalize is a variant of the Finalize that panics on error.
func (r *Reaper) MustFinalize() {
	if err := r.Finalize(); err != nil {
		panic(err)
	}
}

// reap calls given destructors in the backward order.
func reap(destructors ...Destructor) error {
	coerr := kerror.NewCollector()
	for i := len(destructors) - 1; i >= 0; i-- {
		coerr.Collect(kerror.Try(func() error {
			return destructors[i].Destroy()
		}))
	}
	return coerr.Error()
}
