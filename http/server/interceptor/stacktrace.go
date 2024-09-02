package interceptor

import (
	"encoding/json"
	"fmt"
	"go/build"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type (
	StackFields struct {
		Function string `json:"func"`
		Path     string `json:"path"`
		Line     string `json:"line"`
	}
	Stack []StackFields
)

var (
	dir, _ = os.Getwd()
	nr     = strings.NewReplacer(dir, "")
)

func (sf Stack) Error() string {
	var lines string
	for i := range sf {
		lines += fmt.Sprintf("func: %s, path: %s, line: %s\n", sf[i].Function, sf[i].Path, sf[i].Line)
	}
	return lines
}

func (sf Stack) Print() (result []map[string]string) {
	bytes, _ := json.Marshal(sf)
	if err := json.Unmarshal(bytes, &result); err == nil {
		return
	}
	return nil
}

func GetStacktrace() Stack {
	callers := poolGet()
	n := runtime.Callers(3, *callers)
	frames := runtime.CallersFrames(*callers)
	value := make(Stack, 0, n) // Fixed slice initialization to avoid empty elements

	for {
		frame, more := frames.Next()

		if strings.HasPrefix(frame.File, build.Default.GOROOT) {
			if !more {
				break
			}
			continue
		}

		value = append(value, StackFields{
			Function: unescape(frame.Function),
			Path:     nr.Replace(frame.File),
			Line:     strconv.Itoa(frame.Line),
		})

		if !more {
			break
		}
	}
	poolPut(callers)

	return value
}

var (
	pool = new(sync.Pool)
)

func poolGet() *[]uintptr {
	value := pool.Get() // Fixed the incorrect pool used here

	if value != nil {
		callers, typed := value.(*[]uintptr)

		if typed {
			return callers
		}
	}

	callers := make([]uintptr, 256)
	return &callers
}

func poolPut(callers *[]uintptr) {
	pool.Put(callers) // Fixed the incorrect pool used here
}

func unescape(source string) string {
	// in the previous one there was decoder that doesn't return error so we should ignore it there
	unescaped, _ := url.QueryUnescape(source)
	return unescaped
}
