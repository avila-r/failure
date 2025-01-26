package failure

import (
	"fmt"
	"strconv"
)

type ErrorBuilder struct {
	class       *ErrorClass
	message     string
	cause       error
	mode        ErrorBuildStackMode
	transparent bool
}

type ErrorBuildStackMode int

const (
	ErrorBuildStackModeTraceCollect ErrorBuildStackMode = iota + 1
	ErrorBuildStackModeTraceCollectTraceBorrowOrCollect
	ErrorBuildStackModeTraceBorrowOnly
	ErrorBuildStackModeTraceEnhance
	ErrorBuildStackModeTraceOmit
)

func Builder(c *ErrorClass) ErrorBuilder {
	return ErrorBuilder{
		class: c,
		mode: func() ErrorBuildStackMode {
			if c.Modifiers.CollectStackTrace() {
				return ErrorBuildStackModeTraceCollect
			} else {
				return ErrorBuildStackModeTraceOmit
			}
		}(),
		transparent: c.Modifiers.Transparent(),
	}
}

func (b ErrorBuilder) Cause(err error) ErrorBuilder {
	b.cause = err
	if Cast(err) != nil {
		if b.class.Modifiers.CollectStackTrace() {
			b.mode = ErrorBuildStackModeTraceCollectTraceBorrowOrCollect
		} else {
			b.mode = ErrorBuildStackModeTraceBorrowOnly
		}
	}
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
		b.mode = ErrorBuildStackModeTraceEnhance
	} else {
		b.mode = ErrorBuildStackModeTraceCollect
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
		Message:     b.message,
		Cause:       b.cause,
		Transparent: b.transparent,
		StackTrace: func() *StackTrace {
			switch b.mode {
			case ErrorBuildStackModeTraceCollect:
				return getStackTrace()
			case ErrorBuildStackModeTraceBorrowOnly:
				if casted := Cast(b.cause); casted != nil {
					return casted.StackTrace
				} else {
					return nil
				}
			case ErrorBuildStackModeTraceCollectTraceBorrowOrCollect:
				if casted := Cast(b.cause); casted != nil && casted.StackTrace != nil {
					return casted.StackTrace
				} else {
					return getStackTrace()
				}
			case ErrorBuildStackModeTraceEnhance:
				current := getStackTrace()
				if casted := Cast(b.cause); casted != nil && casted.StackTrace != nil {
					current.Cause = casted.StackTrace
				}
				return current
			case ErrorBuildStackModeTraceOmit:
				return nil
			default:
				panic("unknown mode " + strconv.Itoa(int(b.mode)))
			}
		}(),
	}
}
