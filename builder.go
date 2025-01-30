package failure

import (
	"fmt"
	"strconv"

	"github.com/avila-r/failure/stacktrace"
)

type ErrorBuilder struct {
	class       *ErrorClass
	message     string
	cause       error
	mode        stacktrace.BuildStackMode
	transparent bool
}

func Builder(c *ErrorClass) ErrorBuilder {
	return ErrorBuilder{
		class: c,
		mode: func() stacktrace.BuildStackMode {
			if c.Modifiers.CollectStackTrace() {
				return stacktrace.TraceCollect
			} else {
				return stacktrace.TraceOmit
			}
		}(),
		transparent: c.Modifiers.Transparent(),
	}
}

func (b ErrorBuilder) Cause(err error) ErrorBuilder {
	b.cause = err
	if Cast(err) != nil {
		if b.class.Modifiers.CollectStackTrace() {
			b.mode = stacktrace.TraceBorrowOrCollect
		} else {
			b.mode = stacktrace.TraceBorrowOnly
		}
	}
	return b
}

func (b ErrorBuilder) StackTrace() ErrorBuilder {
	b.mode = stacktrace.TraceTrimmed
	return b
}

func (b ErrorBuilder) Transparent() ErrorBuilder {
	if b.cause == nil {
		panic("wrong builder usage: wrap modifier without non-nil cause")
	}
	b.transparent = true
	return b
}

func (b ErrorBuilder) EnhanceStackTrace() ErrorBuilder {
	if b.cause == nil {
		panic("wrong builder usage: wrap modifier without non-nil cause")
	}
	if Cast(b.cause) != nil {
		b.mode = stacktrace.TraceEnhance
	} else {
		b.mode = stacktrace.TraceCollect
	}
	return b
}

func (b ErrorBuilder) Message(message string, v ...any) ErrorBuilder {
	if len(v) == 0 {
		b.message = message
	} else {
		b.message = fmt.Sprintf(message, v...)
	}
	return b
}

func (b ErrorBuilder) Build() *Error {
	return &Error{
		class:       b.class,
		message:     b.message,
		cause:       b.cause,
		transparent: b.transparent,
		stacktrace:  b.SetupStackTrace(),
	}
}

func (b ErrorBuilder) SetupStackTrace() *stacktrace.StackTrace {
	switch b.mode {
	case stacktrace.TraceCollect:
		return stacktrace.Collect()

	case stacktrace.TraceBorrowOrCollect:
		return b.collect(b.cause)
	case stacktrace.TraceBorrowOnly:
		if st := b.collect(b.cause); st != nil {
			return st
		}

		return stacktrace.Collect()
	case stacktrace.TraceEnhance:
		current, initial := stacktrace.Collect(), b.collect(b.cause)
		if initial != nil {
			current.Cause(initial)
		}
		return current
	case stacktrace.TraceOmit:
		return nil
	case stacktrace.TraceTrimmed:
		return stacktrace.
			Collect().
			Trimmed()
	default:
		panic("unknown mode " + strconv.Itoa(int(b.mode)))
	}
}

func (b ErrorBuilder) collect(cause error) *stacktrace.StackTrace {
	if casted := Cast(cause); casted != nil {
		return casted.stacktrace
	}

	return nil
}
