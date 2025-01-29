package stacktrace

import (
	"sync"
	"sync/atomic"
)

type FilePathTransformer func(string) string

var Transformer = func() (t struct {
	Mu          *sync.Mutex
	Transform   *atomic.Value
	Initialized bool
}) {
	t.Mu = &sync.Mutex{}
	t.Transform = &atomic.Value{}

	var value FilePathTransformer = func(line string) string {
		return line
	}

	t.Transform.Store(value)

	return t
}()
