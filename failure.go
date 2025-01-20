package failure

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/avila-r/failure/trait"
)

func Err(message string, v ...any) error {
	if len(v) > 0 {
		return fmt.Errorf(message, v...)
	}

	return errors.New(message)
}

func Of(message string, v ...any) *Error {
	return New(message, v...)
}

func New(message string, v ...any) *Error {
	return nil
}

func From(err error) *Error {
	return Cast(err)
}

// Cast attempts to cast an error to errorx Type, returns nil if cast has failed.
func Cast(err error) *Error {
	if e, ok := err.(*Error); ok && e != nil {
		return e
	}

	return nil
}

func Decorate(err error, message string, v ...any) *Error {
	return Builder(transparentWrapper).
		Message(message, v...).
		Cause(err).
		Build()
}

func EnhanceStackTrace(err error, message string, v ...any) *Error {
	return Builder(transparentWrapper).
		Message(message, v...).
		Cause(err).
		EnhanceStackTrace().
		Build()
}

func EnsureStackTrace(err error) *Error {
	if casted := Cast(err); casted != nil && casted.StackTrace != nil {
		return casted
	}

	return Builder(stackTraceWrapper).
		Message("").
		Cause(err).
		EnhanceStackTrace().
		Build()
}

func Extends(err error, c *ErrorClass) bool {
	casted := func() *Error {
		raw := err
		for raw != nil {
			typed := Cast(raw)
			if typed != nil {
				return typed
			}
			raw = errors.Unwrap(raw)
		}

		return nil
	}()

	return casted != nil && casted.Extends(c)
}

func Has(err error, trait trait.Trait) bool {
	if casted := Cast(err); casted != nil {
		return casted.Has(trait)
	}

	return false
}

func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	for {
		if reflect.TypeOf(target).Comparable() && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

func As(err error, target any) bool {
	if target == nil || err == nil {
		return false
	}

	val := reflect.ValueOf(target)
	typ := val.Type()

	// target must be a non-nil pointer
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}

	// *target must be interface or implement error
	if e := typ.Elem(); e.Kind() != reflect.Interface && !e.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}

	for {
		typeof := reflect.TypeOf(err)
		if typeof != reflect.TypeOf(&Error{}) && reflect.TypeOf(err).AssignableTo(typ.Elem()) {
			val.Elem().Set(reflect.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(any) bool }); ok && x.As(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

func Cause(err error) error {
	for {
		previous := Unwrap(err)
		if previous == nil {
			return err
		}
		err = previous
	}
}

func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})

	if !ok {
		return nil
	}

	return u.Unwrap()
}

func Trait(label string) trait.Trait {
	return trait.New(label)
}
