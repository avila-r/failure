package trail

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type (
	Trail struct {
		Span   string
		Frames []step
	}
)

const (
	depth = 10
	this  = "github.com/avila-r/failure"
)

func New(span string) *Trail {
	frames := []step{}

	for i := 1; len(frames) < depth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		file = func(path string) string {
			dirs := filepath.SplitList(os.Getenv("GOPATH"))
			sort.Stable(paths(filepath.SplitList(os.Getenv("GOPATH"))))
			for _, dir := range dirs {
				srcdir := filepath.Join(dir, "src")
				rel, err := filepath.Rel(srcdir, path)
				if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					return rel
				}
			}
			return path
		}(file)

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		function := shorten(f)

		if !(len(runtime.GOROOT()) > 0 &&
			strings.Contains(file, runtime.GOROOT())) &&
			(!(strings.Contains(file, this)) ||
				(strings.Contains(file, this+"/examples/")) ||
				(strings.Contains(file, "_test.go"))) {
			frames = append(frames, step{
				pc:       pc,
				file:     file,
				function: function,
				line:     line,
			})
		}
	}

	return &Trail{
		Span:   span,
		Frames: frames,
	}
}

type (
	step struct {
		pc       uintptr
		file     string
		function string
		line     int
	}

	paths []string
)

func (t *Trail) Error() string {
	return t.String("")
}

func (s *step) String() string {
	current := fmt.Sprintf("%v:%v", s.file, s.line)
	if s.function != "" {
		current = fmt.Sprintf("%v:%v %v()", s.file, s.line, s.function)
	}

	return current
}

func (t *Trail) String(deepest string) string {
	str := ""

	new := func() {
		if str != "" && !strings.HasSuffix(str, "\n") {
			str += "\n"
		}
	}

	for _, frame := range t.Frames {
		if frame.file != "" {
			current := frame.String()
			if current == deepest {
				break
			}

			new()
			str += "  --- at " + current
		}
	}

	return str
}

func (t *Trail) Source() (string, []string) {
	if len(t.Frames) == 0 {
		return "", []string{}
	}

	first := t.Frames[0]
	header := first.String()
	body := from(first)

	return header, body
}

func (p paths) Len() int {
	return len(p)
}

func (p paths) Less(i, j int) bool {
	return len(p[i]) > len(p[j])
}

func (p paths) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func shorten(f *runtime.Func) string {
	longName := f.Name()

	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1)
	shortName = strings.Replace(shortName, "*", "", 1)
	shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}

func from(frame step) []string {
	const (
		NumberLinesBefore = 5
		NumberLinesAfter  = 5
	)

	lines, ok := read(frame.file)
	if !ok {
		return []string{}
	}

	if len(lines) < frame.line {
		return []string{}
	}

	current, output := frame.line-1, []string{}
	start, end :=
		func(a, b int) int {
			if a > b {
				return a
			} else {
				return b
			}
		}(0, current-NumberLinesBefore),
		func(a, b int) int {
			if a > b {
				return b
			} else {
				return a
			}
		}(len(lines)-1, current+NumberLinesAfter)

	for i := start; i <= end; i++ {
		if i < 0 || i >= len(lines) {
			continue
		}

		line := lines[i]
		message := fmt.Sprintf("%d\t%s", i+1, line)
		output = append(output, message)

		convert := func(count int, predicate func(index int) byte) string {
			result := make([]byte, 0, count)
			for i := 0; i < count; i++ {
				result = append(result, predicate(i))
			}
			return string(result)
		}

		if i == current {
			length := len(strings.TrimLeft(line, " \t"))
			leading := len(line) - length
			tabs := strings.Count(line[0:leading], "\t")
			first := leading + (8-1)*tabs // 8 chars per tab
			prefix := convert(first, func(_ int) byte {
				return ' '
			})
			subline := convert(length, func(_ int) byte {
				return '^'
			})
			output = append(output, "\t"+prefix+subline)
		}
	}

	return output
}

var (
	mutex sync.RWMutex
	cache map[string]paths
)

func read(path string) ([]string, bool) {
	mutex.RLock()
	lines, ok := cache[path]
	mutex.RUnlock()

	if ok {
		return lines, true
	}

	if !strings.HasSuffix(path, ".go") {
		return nil, false
	}

	// bearer:disable go_gosec_filesystem_filereadtaint
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	lines = strings.Split(string(b), "\n")

	mutex.Lock()
	cache[path] = lines
	mutex.Unlock()

	return lines, true
}
