package errs_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/marcuswestin/go-errs"
)

func TestNew(t *testing.T) {
	err := errs.New(nil)
	assert(t, err.Time().Nanosecond() != 0, "Expected non-0 time")
	assert(t, err.Stack() != nil, "Expected non-nil time")
	assert(t, err.StdError() == nil, "Expected nil stdError")
	assert(t, err.PublicMsg() == "", "Expected no public message")
}

func TestInfo(t *testing.T) {
	err := errs.New(errs.Info{"Foo": "Bar"})
	assert(t, err.Info("Foo") == "Bar", "Expected info Foo to be Bar")
}

func TestWrap(t *testing.T) {
	stdErr := errors.New("It broke!")
	err := errs.Wrap(stdErr, nil)
	assert(t, err.StdError() != nil, "Expected a standard error")
	assert(t, err.StdError().Error() == "It broke!", "Expected standard error message to be It broke!")
}

func TestWrapNil(t *testing.T) {
	err := errs.Wrap(nil, nil)
	assert(t, err == nil, "Expected nil-wrapped err to be nil")
}

func TestNilInfo(t *testing.T) {
	err := errs.New(nil)
	assert(t, err.Info("Foo") == nil)
}

func TestMultiWrap(t *testing.T) {
	publicMsg := "publicMsg"
	err := errs.New(errs.Info{"Key": "First"}, publicMsg)
	err = errs.Wrap(err, errs.Info{"Key": "Second"}, publicMsg)
	err = errs.Wrap(err, errs.Info{"Key": "Third"}, publicMsg)
	assert(t, err.Info("Key") == "First")
	assert(t, err.Info("Key_duplicate") == "Second")
	assert(t, err.Info("Key_duplicate_duplicate") == "Third")
	fmt.Println("ASDASD", err.PublicMsg())
	assert(t, err.PublicMsg() == strings.Join([]string{publicMsg, publicMsg, publicMsg}, " - "))
}

func assert(t *testing.T, ok bool, msg ...interface{}) {
	if !ok {
		panic(msg)
		t.Fatal(msg...)
	}
}
