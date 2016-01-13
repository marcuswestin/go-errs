// Package errs improves on the standard `error` by encapsulating stack traces, timestamps,
// optional internal information, and optional public user-facing messages.
//
// Usage
//
// Create an empty error with stack trace and timestamp:
//
//  err := errs.New(nil)
//  err.Time() // time.Time at time of creation
//  err.Stack() // output from debug.Stack() at time of creation
//
// Create an error with associated internal info and a user-facing message:
//
//  userEmail := "user@example.com"
//  emailExists := checkIfEmailExists(userEmail)
//  if emailExists {
//    err := errs.New(errs.Info{ "Email":userEmail }, userEmail, "is already taken. Try another!")
//    return err
//  }
//  ...
//  err.Info("Email") // "user@example.com"
//  err.PublicMsg() // "user@example.com is already taken. Try another!"
//
// Wrap a standard error:
//
//  stdErr := errors.New("An error")
//  err := errs.Wrap(stdErr, errs.Info{"Foo": "Bar"}, "User message")
//  ...
//  err.StdError() // stdErr
package errs

import (
	"fmt"
	"runtime/debug"
	"time"
)

type Err interface {
	// Error is an alias for LogString.
	// (errs.Err implements the error interface).
	Error() string

	// Stack returns the result of debug.Stack() from the time when this Err was created.
	Stack() []byte

	// Time returns the time.Time at which this Err was created.
	Time() time.Time

	// If errs.Wrap was used then StdError return the wrapped standard error.
	StdError() error

	// If errs.Wrap or errs.New was called with any publicMsg values
	// then PublicMsg returns a string representation of those values.
	// This is useful for bubbling up user-facing message strings,
	// e.g `errs.New(nil, userEmail, "is already taken. Try another!")`
	PublicMsg() string

	// If errs.Wrap or errs.New was called with an errs.Info object
	// then Info("Foo") return the value of errs.Info{"Foo":...}
	// This is useful for bubbling up internal-facing info,
	// e.g `errs.Wrap(sqlError, { "SqlString":sqlStr, "SqlArgs":sqlArgs })`
	Info(name string) interface{}

	// AllInfo returns all info key-value-pairs passed through errs.New or errs.Wrap
	AllInfo() Info

	// LogString returns a string suitable for logging
	LogString() string
}

// New creates a new Err with the given Info and optional public message
func New(info Info, publicMsg ...interface{}) Err {
	return newErr(nil, info, publicMsg)
}

// Wrap creates a new Err with the given standard error, Info, and optional public message.
// If stdErr is nil, Wrap returns nil.
func Wrap(stdErr error, info Info, publicMsg ...interface{}) Err {
	if stdErr == nil {
		return nil
	}
	if errsErr, isErr := IsErr(stdErr); isErr {
		if errs_err, isErrsErr := errsErr.(*err); isErrsErr {
			fmt.Println("ASDASDASD", errs_err.info)
			errs_err.mergeIn(info, publicMsg)
			return errs_err
		}
		return errsErr
	}
	return newErr(stdErr, info, publicMsg)
}

// Info allows for associating internal info with an error,
// e.g `errs.Wrap(sqlError, { "SqlString":sqlStr, "SqlArgs":sqlArgs })`
type Info map[string]interface{}

// IsErr checks if stdErr is an Err, and return it as an Err if it is.
// This is equivalent to stdErr.(errs.Err)
func IsErr(stdErr error) (err Err, isErr bool) {
	err, isErr = stdErr.(Err)
	return
}

// Internal
///////////

// err implements Err
type err struct {
	stack     []byte
	time      time.Time
	stdErr    error
	info      Info
	publicMsg string
}

func newErr(stdErr error, info Info, publicMsgParts []interface{}) Err {
	publicMsg := fmt.Sprint(publicMsgParts...)
	stack := debug.Stack()
	return &err{stack, time.Now(), stdErr, info, publicMsg}
}

// Implements Err
func (e *err) Stack() []byte     { return e.stack }
func (e *err) Time() time.Time   { return e.time }
func (e *err) StdError() error   { return e.stdErr }
func (e *err) PublicMsg() string { return e.publicMsg }
func (e *err) Error() string     { return e.LogString() }
func (e *err) String() string    { return e.LogString() }
func (e *err) AllInfo() Info     { return e.info }

// Implements Err
func (e *err) Info(key string) interface{} {
	if e.info == nil {
		return nil
	}
	return e.info[key]
}

// Implements Err
func (e *err) LogString() string {
	return fmt.Sprint("Error",
		"| Time:", e.time,
		"| Stack:", string(e.stack),
		"| StdError:", e.stdErrStr(),
		"| Info:["+fmt.Sprint(e.info)+"]",
		"| PublicMsg:", e.publicMsg,
	)
}

// Merge in the given info and public message parts into this error
func (e *err) mergeIn(info Info, publicMsgParts []interface{}) {
	for key, val := range info {
		for e.info[key] != nil {
			key = key + "_duplicate"
		}
		e.info[key] = val
	}
	e.publicMsg = fmt.Sprint(publicMsgParts...) + " - " + e.publicMsg
}

// Get the string representation of the stdErr, or an empty string if stdErr is nil
func (e *err) stdErrStr() string {
	if e == nil {
		return ""
	}
	if e.stdErr == nil {
		return ""
	}
	return e.stdErr.Error()
}
