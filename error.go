package failure

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/avila-r/failure/property"
	"github.com/avila-r/failure/stacktrace"
	"github.com/avila-r/failure/trait"
)

type Error struct {
	class *ErrorClass

	Message string
	Cause   error

	stacktrace             *stacktrace.StackTrace
	properties             *property.List
	transparent            bool
	hasUnderlying          bool
	printablePropertyCount uint8
}

var (
	_ error         = (*Error)(nil)
	_ fmt.Formatter = (*Error)(nil)
)

func (e *Error) Belongs(err error) bool {
	typed := Cast(err)

	return typed != nil && Extends(e, typed.Class())
}

func (e *Error) Is(err error) bool {
	if err, ok := err.(*Error); ok {
		return e.Message == err.Message
	}

	return e.Message == err.Error()
}

func (e *Error) As(target any) bool {
	t := reflect.Indirect(reflect.ValueOf(target)).Interface()

	if err, ok := t.(*Error); ok {
		if e.Message == err.Message {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(err))
			return true
		}
	}
	return false
}

func (e *Error) Has(trait trait.Trait) bool {
	cause := e
	for cause != nil {
		if !cause.transparent {
			return cause.class.Has(trait)
		}
		cause = Cast(cause.Cause)
	}

	return false
}

func (e *Error) Extends(c *ErrorClass) bool {
	cause := e
	for cause != nil {
		if !cause.transparent {
			return cause.Class().Is(c)
		}

		cause = func() *Error {
			raw := e.Cause
			for raw != nil {
				typed := Cast(raw)
				if typed != nil {
					return typed
				}

				raw = errors.Unwrap(raw)
			}

			return nil
		}()
	}

	return false
}

func (e *Error) Attribute(key string) property.Result {
	return e.Property(key)
}

func (e *Error) Field(key string) property.Result {
	return e.Property(key)
}

func (e *Error) Value(key string) property.Result {
	return e.Property(key)
}

func (e *Error) Property(key string) property.Result {
	cause := e
	for cause != nil {
		value, ok := cause.properties.Get(key)
		if ok {
			return property.Result{
				Value: value,
				Ok:    true,
			}
		}

		if !cause.transparent {
			break
		}

		cause = Cast(cause.Cause)
	}

	return property.Result{
		Value: nil,
		Ok:    false,
	}
}

func (e *Error) With(key string, value any) *Error {
	copy := *e
	copy.properties = copy.properties.Set(key, value)
	if copy.printablePropertyCount < 255 {
		copy.printablePropertyCount++
	}
	return &copy
}

func (e *Error) Also(errs ...error) *Error {
	var (
		underlying = e.underlying()
		new        = underlying
	)

	for _, err := range errs {
		if err == nil {
			continue
		}
		new = append(new, err)
	}

	if len(new) == len(underlying) {
		return e
	}

	l := len(new)
	copy := e.With(property.Underlying, new[:l:l])
	copy.hasUnderlying = true
	return copy
}

func (e *Error) Unwrap() error {
	if e != nil && e.Cause != nil && e.transparent {
		return e.Cause
	} else {
		return nil
	}
}

func (e *Error) Class() *ErrorClass {
	cause := e
	for cause != nil {
		if !cause.transparent {
			return cause.class
		}

		cause = Cast(cause.Cause)
	}

	return foreignClass
}

func (e *Error) Summary() string {
	var join = func(delimiter string, parts ...string) string {
		switch len(parts) {
		case 0:
			return ""
		case 1:
			return parts[0]
		case 2:
			if len(parts[0]) == 0 {
				return parts[1]
			} else if len(parts[1]) == 0 {
				return parts[0]
			} else {
				return parts[0] + delimiter + parts[1]
			}
		default:
			filtered := make([]string, 0, len(parts))
			for _, part := range parts {
				if len(part) > 0 {
					filtered = append(filtered, part)
				}
			}

			return strings.Join(filtered, delimiter)
		}
	}

	properties := ""
	if e.properties != nil && e.printablePropertyCount != 0 {
		var (
			uniq = make(map[string]struct{}, e.printablePropertyCount)
			strs = make([]string, 0, e.printablePropertyCount)
		)

		for m := e.properties; m != nil; m = m.Next {
			if _, ok := uniq[m.Key]; ok {
				continue
			}
			uniq[m.Key] = struct{}{}
			strs = append(strs, fmt.Sprintf("%s: %v", m.Key, m.Value))
		}

		properties = "{" + strings.Join(strs, ", ") + "}"
	}

	text := join(" ", e.Message, properties)
	if cause := e.Cause; cause != nil {
		text = join(", cause: ", text, cause.Error())
	}

	underlying := ""
	if e.hasUnderlying {
		details := make([]string, 0, len(e.underlying()))
		for _, err := range e.underlying() {
			details = append(details, err.Error())
		}
		underlying = fmt.Sprintf("(hidden: %s)", join(", ", details...))
	}

	if transparent := join(" ", text, underlying); e.transparent {
		return transparent
	} else {
		return join(": ", e.class.Name, transparent)
	}
}

func (e *Error) underlying() []error {
	if !e.hasUnderlying {
		return nil
	}
	u, _ := e.properties.Get(property.Underlying)
	return u.([]error)
}

// Error implements the error interface.
// A result is the same as with %s formatter and does not contain a stack trace.
func (e *Error) Error() string {
	return e.Message
}

// Format implements the Formatter interface.
//
// Supported verbs:
//
//	%s		simple message output
//	%v		simple message output
//	%+v		full output complete with a stack trace
//
// In is nearly always preferable to use %+v format.
// If a stack trace is not required, it should be omitted
// at the moment of creation rather in formatting.
func (e *Error) Format(state fmt.State, verb rune) {
	switch message := e.Summary(); verb {
	case 'v':
		_, _ = io.WriteString(state, message)
		if state.Flag('+') && e.stacktrace != nil {
			e.stacktrace.Format(state, verb)
		}
	case 's':
		_, _ = io.WriteString(state, message)
	}
}
