package stacktrace

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
)

type StackTrace struct {
	pc    []uintptr
	cause *StackTrace
}

func (s *StackTrace) Cause(cause *StackTrace) {
	s.cause = cause
}

func Collect() *StackTrace {
	pc := [128]uintptr{}
	return &StackTrace{
		pc: pc[:runtime.Callers(5, pc[:])],
	}
}

var _ fmt.Formatter = (*StackTrace)(nil)

// Format implements fmt.Formatter.
func (s *StackTrace) Format(state fmt.State, verb rune) {
	if s == nil {
		return
	}

	switch verb {
	case 'v', 's':
		transformLine := Transformer.Transform.Load().(FilePathTransformer)

		var (
			pc      []uintptr
			cropped int
			subpc   []uintptr
		)

		pc, cropped = s.pc, 0
		if s.cause != nil {
			subpc = s.cause.pc
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

		if s.cause != nil {
			io.WriteString(state, "\n ---------------------------------- ")
			s.cause.Format(state, verb)
		}
	}
}
