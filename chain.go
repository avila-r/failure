package failure

import (
	"time"

	"github.com/avila-r/failure/tags"
)

type ErrorChain struct {
	err *Error
}

func (c *ErrorChain) Done() *Error {
	return c.err
}

func (c *ErrorChain) Owner(owner string) *ErrorChain {
	c.err.owner = owner
	return c
}

func (c *ErrorChain) Public(public string) *ErrorChain {
	c.err.public = public
	return c
}

func (c *ErrorChain) Hint(hint string) *ErrorChain {
	c.err.hint = hint
	return c
}

func (c *ErrorChain) Span(span string) *ErrorChain {
	c.err.span = span
	return c
}

func (c *ErrorChain) Trace(trace string) *ErrorChain {
	c.err.trace = trace
	return c
}

func (c *ErrorChain) Tags(tags tags.Tags) *ErrorChain {
	c.err.tags = tags
	return c
}

func (c *ErrorChain) In(domain string) *ErrorChain {
	c.err.domain = domain
	return c
}

func (c *ErrorChain) Duration(duration time.Duration) *ErrorChain {
	c.err.duration = duration
	return c
}

func (c *ErrorChain) Since(t time.Time) *ErrorChain {
	c.err.duration = time.Since(t)
	return c
}

func (c *ErrorChain) Time(time time.Time) *ErrorChain {
	c.err.time = time
	return c
}

func (c *ErrorChain) Assert(condition bool, message ...any) *ErrorChain {
	msg, args := "assertion failed on error's constructor", []any{}
	if len(message) > 0 {
		first, ok := message[0].(string)
		if ok {
			msg = first

			if len(message) > 1 {
				args = append(args, message[1:]...)
			}
		}
	}

	if !condition {
		AssertionFailed.
			New(msg, args...).
			Panic()
	}

	return c
}

func (c *ErrorChain) Recover(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = c.err.Also(e)
			} else {
				err = c.err.Also(Err("%v", r))
			}
		}
	}()

	f()
	return
}
