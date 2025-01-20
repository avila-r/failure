package failure

import "errors"

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

func (c *ErrorClass) Of(message string, v ...any) *Error {
	return New(message, v...)
}

func (c *ErrorClass) New(message string, v ...any) *Error {
	return Builder(c).
		Message(message, v...).
		Build()
}

func (c *ErrorClass) Blank() *Error {
	return Builder(c).
		Build()
}

func (c *ErrorClass) Wrap(err error, message string, v ...any) *Error {
	return Builder(c).
		Message(message, v...).
		Cause(err).
		Build()
}

func (c *ErrorClass) From(err error) *Error {
	return Builder(c).
		Cause(err).
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
