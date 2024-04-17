package testkit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

var inputOutputSeparator = strings.Repeat("-", 6)

type lineScanner struct {
	*bufio.Scanner
	line int
}

func NewLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{
		Scanner: bufio.NewScanner(r),
		line:    0,
	}
}

func (l *lineScanner) Scan() bool {
	ok := l.Scanner.Scan()
	if ok {
		l.line++
	}
	return ok
}

// Testcase represents a single test case.
//
// A typical test case should look like this:
// TestNamexxx
// explain
// SELECT * FROM users
// ------
// XXXXX(output of the explain)
//
// Order of the fields is important.
type Testcase struct {
	Pos      string // file:line
	CaseName string
	Cmd      string
	Sql      string
	Expected string
}

// testDataReader handle(read/rewrite) test cases from a file.
type testDataReader struct {
	path       string
	file       *os.File
	scanner    *lineScanner
	rewriteBuf *bytes.Buffer

	Data Testcase
}

// NewTestDataReader returns a new testDataReader.
func NewTestDataReader(path string, rewrite bool) (*testDataReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var rewriteBuf *bytes.Buffer
	if rewrite {
		rewriteBuf = &bytes.Buffer{}
	}

	return &testDataReader{
		path:       path,
		file:       file,
		scanner:    NewLineScanner(file),
		rewriteBuf: rewriteBuf,
	}, nil
}

func (r *testDataReader) Close() error {
	return r.file.Close()
}

func (r *testDataReader) Rewrite() error {
	if r.rewriteBuf != nil {
		data := r.rewriteBuf.Bytes()

		return os.WriteFile(r.path, data, 0644)
	}

	return nil
}

func (r *testDataReader) Record(test string) {
	if r.rewriteBuf != nil {
		r.rewriteBuf.WriteString(test + "\n")
	}
}

// readCaseLine reads the case line from the file.
// It returns false on EOF.
func (r *testDataReader) readCaseLine() bool {
	caseLine := r.scanner.Text()
	r.Record(caseLine)

	// case line
	fields := strings.Fields(caseLine) // empty line
	if len(fields) == 0 {
		return false
	}
	r.Data.Pos = fmt.Sprintf("%s:%d", r.path, r.scanner.line)
	r.Data.CaseName = strings.TrimSpace(caseLine)
	return true
}

// readCmdLine reads the cmd line from the file.
// It returns false on EOF.
func (r *testDataReader) readCmdLine() bool {
	cmdLine := r.scanner.Text()
	r.Record(cmdLine)

	fields := strings.Fields(cmdLine) // empty line
	if len(fields) == 0 {
		return false
	}
	cmd := fields[0]

	r.Data.Cmd = cmd
	return true
}

// Next reads the next testcase from the file and returns true if successful.
func (r *testDataReader) Next() bool {
	r.Data = Testcase{}

	for {
		// case line
		if !r.scanner.Scan() {
			break
		}

		if !r.readCaseLine() {
			continue
		}

		// cmd line
		if !r.scanner.Scan() {
			break
		}

		if !r.readCmdLine() {
			continue
		}

		// sql line, stop at inputOutputSeparator
		var expectOutputStart bool
		var buf bytes.Buffer
		for r.scanner.Scan() {
			line := r.scanner.Text()
			if strings.TrimSpace(line) == "" {
				break
			}
			r.Record(line)
			if line == inputOutputSeparator {
				expectOutputStart = true
				break
			}
			buf.WriteString(line)
		}
		r.Data.Sql = strings.TrimSpace(buf.String())

		// expected line
		if expectOutputStart {
			buf.Reset()
			for r.scanner.Scan() {
				line := r.scanner.Text()
				//r.Record(line)
				if strings.TrimSpace(line) == "" {
					break
				}
				fmt.Fprintln(&buf, line)
			}
			r.Data.Expected = buf.String()
		}

		return true
	}

	return false
}
