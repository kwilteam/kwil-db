package dbml

import (
	"io"
	"os"
	"strings"
)

func ParseFile(filePath string) (*DBML, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func ParseString(str string) (*DBML, error) {
	return Parse(strings.NewReader(str))
}

func Parse(rd io.Reader) (*DBML, error) {
	return NewParser(NewScanner(rd)).Parse()
}
