package failure

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

type StackTrace struct {
	Counter []uintptr
	Cause   *StackTrace
}

func getStackTrace() *StackTrace {
	pc := [128]uintptr{}
	return &StackTrace{
		Counter: pc[:runtime.Callers(6, pc[:])],
	}
}

type stackTraceTransformer struct {
	mu          *sync.Mutex
	transform   *atomic.Value
	initialized bool
}

var transformer = func() *stackTraceTransformer {
	t := &stackTraceTransformer{
		&sync.Mutex{},
		&atomic.Value{},
		false,
	}

	var value StackTraceFilePathTransformer = func(line string) string {
		return line
	}

	t.transform.Store(value)

	return t
}()

type StackTraceFilePathTransformer func(string) string

func InitializeStackTraceTransformer(subtransformer StackTraceFilePathTransformer) (StackTraceFilePathTransformer, error) {
	transformer.mu.Lock()
	defer transformer.mu.Unlock()

	old := transformer.transform.Load().(StackTraceFilePathTransformer)
	transformer.transform.Store(subtransformer)

	if transformer.initialized {
		return old, InitializationFailed.New("stack trace transformer was already set up: %#v", old)
	}

	transformer.initialized = true
	return nil, nil
}

var _ fmt.Formatter = (*StackTrace)(nil)

// Format implements fmt.Formatter.
func (s *StackTrace) Format(state fmt.State, verb rune) {
	if s == nil {
		return
	}

	switch verb {
	case 'v', 's':
		transformLine := transformer.transform.Load().(StackTraceFilePathTransformer)

		pc, cropped, subpc := s.Counter, 0, s.Cause.Counter
		if s.Cause != nil {
			found := false
			for i := 1; i <= len(pc) && i <= len(subpc); i++ {
				if pc[len(pc)-i] != subpc[len(subpc)-i] {
					pc, cropped, found = pc[:len(pc)-i], i-1, true
					break
				}
			}

			if !found {
				pc, cropped = nil, len(pc)
			}
		}

		if len(pc) == 0 {
			return
		}

		var (
			frames    = make([]*runtime.Frame, 0, len(pc))
			subframes = runtime.CallersFrames(pc[:])
			next      = true
			raw       runtime.Frame
		)

		for next {
			raw, next = subframes.Next()
			copy := raw
			frames = append(frames, &copy)
		}

		for _, frame := range frames {
			io.WriteString(state, "\n at ")
			io.WriteString(state, frame.Function)
			io.WriteString(state, "()\n\t")
			io.WriteString(state, transformLine(frame.File))
			io.WriteString(state, ":")
			io.WriteString(state, strconv.Itoa(frame.Line))
		}

		if cropped > 0 {
			io.WriteString(state, "\n ...\n (")
			io.WriteString(state, strconv.Itoa(cropped))
			io.WriteString(state, " duplicated frames)")
		}

		if s.Cause != nil {
			io.WriteString(state, "\n ---------------------------------- ")
			s.Cause.Format(state, verb)
		}
	}
}
