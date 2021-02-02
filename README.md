# KDone

[![Go Reference](https://pkg.go.dev/badge/github.com/go-kata/kdone.svg)](https://pkg.go.dev/github.com/go-kata/kdone)
[![codecov](https://codecov.io/gh/go-kata/kdone/branch/master/graph/badge.svg?token=C8RHRTAB2Y)](https://codecov.io/gh/go-kata/kdone)

## Installation

`go get github.com/go-kata/kdone`

## Status

**This is a beta version.** API is not stabilized for now.

## Versioning

*Till the first major release* minor version (`v0.x.0`) must be treated as a major version
and patch version (`v0.0.x`) must be treated as a minor version.

For example, changing of version from `v0.1.0` to `v0.1.1` indicates compatible changes,
but when version changes `v0.1.1` to `v0.2.0` this means that the last version breaks the API.

## How to use

This library is designed to simplify code of constructors and provide guaranteed finalization even on panic.

Here is an example code of application lifecycle *(read comments, they are informative)*:

```go
type Application struct {
	
	// Let's say that we don't use the database directly.
	database *Database

	Logger   *Logger
	Consumer *Consumer
}

func NewApplication() (*Application, error) {
	
	// We may use DI here instead, but let's consider this constructor
	// as topmost and encapsulating the whole application initialization.
	
	logger, err := NewLogger()
	if err != nil {
		return nil, err
	}
	// We can't use defer here - if we do this all initialized resources 
	// will be finalized at the end of constructor and resulting application 
	// instance will be broken.
	
	database, err := NewDatabase(logger)
	if err != nil {
		
		// It won't happen in case of panic.
		_ = logger.Close()
		
		return nil, err
	}
	
	consumer, err := NewConsumer(logger, database)
	if err != nil {
		
		// It won't happen in case of panic.
		_ = database.Close()
		_ = logger.Close()
		
		return nil, err
	}
	
	return &Application{database, logger, consumer}, nil
}

func (app *Application) Close() error {
	var errs []error
	
	if err := app.Consumer.Close(); err != nil {
		errs = append(errs, err)
	}
	
	// It won't happen in case of panic.
	if err := app.database.Close(); err != nil {
		errs = append(errs, err)
	}
	
	// It won't happen in case of panic.
	if err := app.Logger.Close(); err != nil {
		errs = append(errs, err)
	}
	
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

type VerboseApplication struct {
	*Application
}

func NewVerboseApplication() (*VerboseApplication, error) {
	app, err := NewApplication()
	if err != nil {
		return nil, err
	}
	app.Logger.Print("application is up")
	return &VerboseApplication{app}, nil
}

func (app *VerboseApplication) Close() error {
	app.Logger.Print("application is down")
	if err := app.Close(); err != nil {
		return err
	}
}

// ...

app, err := NewApplication()
if err != nil {
	HandleError(err)
}
defer CloseWithErrorHandling(app)
```

Of course, there are some assumptions, e.g. logger will be closed even when database
that depends on it was not successfully closed. But it is enough for a simple example.

This code may be rewritten using the library as follows *(read comments, they are informative)*:

```go
type Application struct {
	Logger   *Logger
	Consumer *Consumer
}

func NewApplication() (_ *Application, _ kdone.Destructor, err error) {
	
	// Just to transform panic to error. May be omitted.
	defer kerror.Catch(&err)
	
	// Reaper will call all destructors at the end
	// if wasn't released from this responsibility.
	reaper := kdone.NewReaper()
	defer reaper.MustFinalize()
	
	logger, dtor, err := NewLogger()
	if err != nil {
		return nil, nil, err
	}
	// Destructor will be called anyway - even in case of panic
	// on other initialization steps or in other destructors.
	//
	// We don't loose errors returned from destructors - all of them
	// will be aggregated into one using kerror.Collector.
	//
	// Panic in destructor will be transformed to error.
	reaper.MustAssume(dtor)
	
	// Let's say that Database implements the io.Closer interface
	// instead of returning a dedicated destructor.
	database, err := NewDatabase(logger)
	if err != nil {
		return nil, nil, err
	}
	reaper.MustAssume(kdone.DestructorFunc(database.Close))
	
	consumer, dtor, err := NewConsumer(logger, database)
	if err != nil {
		return nil, nil, err
	}
	reaper.MustAssume(dtor)
	
	// Now an external code is responsible for calling destructors -
	// reaper is released from this responsibility.
	return &Application{logger, consumer}, reaper.MustRelease(), nil
}

func NewVerboseApplication() (*Application, kdone.Destructor, error) {
	app, dtor, err := NewApplication()
	if err != nil {
		return nil, err
	}
	app.Logger.Print("application is up")
	return app, kdone.DestructorFunc(func() error {
		app.Logger.Print("application is down")
		return dtor.Destroy()
	}), err
}

// ...

app, dtor, err := NewApplication()
if err != nil {
	HandleError(err)
}
// Here CloseWithErrorHandling expects io.Closer.
defer CloseWithErrorHandling(kdone.CloserFunc(dtor.Destroy))

// We can't forget to use the dtor variable - it's the compile-time error.
```

This implementation is shorter (even despite lengthy comments), contains fewer entities and
*gives more guarantees of successful finalization*.

Destructor may be easily converted to or from `io.Closer` thanks to `kdone.CloserFunc` and
`kdone.DestructorFunc` helpers. If you need an idiomatic resource with the `Close` method
as its part you may write something like this:

```go
type ClosableResource struct {
	*Resource
	io.Closer
}

res, dtor, _ := NewResource()
resWithClose := ClosableResource{res, kdone.CloserFunc(dtor.Destroy)}
```

## References

**[KError](https://github.com/go-kata/kerror)** is the library that provides tools for handling errors.
