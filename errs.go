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
//  err := errors.New("An error")
//  err = errs.Wrap(err, errs.Info{"Foo": "Bar"}, "User message")
//  ...
//  err.WrappedErr().Error() == "An error"
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

	// If errs.Wrap was used then WrappedError returns the wrapped error.
	WrappedError() error

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

// Wrap the given error in an errs.Err. If err is nil, Wrap returns nil.
// Use Err.WrappedError for direct access to the wrapped error.
func Wrap(wrapErr error, info Info, publicMsg ...interface{}) Err {
	if wrapErr == nil {
		return nil
	}
	if errsErr, isErr := IsErr(wrapErr); isErr {
		if errs_err, isErrsErr := errsErr.(*err); isErrsErr {
			errs_err.mergeIn(info, publicMsg)
			return errs_err
		}
		return errsErr
	}
	return newErr(wrapErr, info, publicMsg)
}

// Info allows for associating key-value-pair info with an error for debugging,
// e.g `errs.Wrap(sqlError, { "SqlString":sqlStr, "SqlArgs":sqlArgs })`
type Info map[string]interface{}

// IsErr checks if err is an errs.Err, and return it as an errs.Err if it is.
// This is equivalent to err.(errs.Err)
func IsErr(err error) (Err, bool) {
	errsErr, isErr := err.(Err)
	return errsErr, isErr
}

// Internal
///////////

// err implements Err
type err struct {
	stack      []byte
	time       time.Time
	wrappedErr error
	info       Info
	publicMsg  string
}

func newErr(wrappedErr error, info Info, publicMsgParts []interface{}) Err {
	publicMsg := concatArgs(publicMsgParts...)
	stack := debug.Stack()
	return &err{stack, time.Now(), wrappedErr, info, publicMsg}
}

// Implements Err
func (e *err) Stack() []byte       { return e.stack }
func (e *err) Time() time.Time     { return e.time }
func (e *err) WrappedError() error { return e.wrappedErr }
func (e *err) PublicMsg() string   { return e.publicMsg }
func (e *err) Error() string       { return e.LogString() }
func (e *err) String() string      { return e.LogString() }
func (e *err) AllInfo() Info       { return e.info }

// Implements Err
func (e *err) Info(key string) interface{} {
	if e.info == nil {
		return nil
	}
	return e.info[key]
}

// Implements Err
func (e *err) LogString() string {
	return concatArgs("Error",
		"| Time:", e.time,
		"| Stack:", string(e.stack),
		"| StdError:", e.wrappedErrStr(),
		"| Info:["+concatArgs(e.info)+"]",
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
	publicMsgPrefix := concatArgs(publicMsgParts...)
	if publicMsgPrefix == "" {
		// do nothing
	} else if e.publicMsg == "" {
		e.publicMsg = publicMsgPrefix
	} else {
		e.publicMsg = publicMsgPrefix + " - " + e.publicMsg
	}
}

// Get the string representation of the wrapper error,
// or an empty string if wrappedErr is nil
func (e *err) wrappedErrStr() string {
	if e == nil {
		return ""
	}
	if e.wrappedErr == nil {
		return ""
	}
	return e.wrappedErr.Error()
}

// Helper to concatenate arguments into a string,
// with spaces between the arguments
func concatArgs(args ...interface{}) string {
	res := fmt.Sprintln(args...)
	return res[0 : len(res)-1] // Remove newline at the end
}
