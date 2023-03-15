package repl

import (
	"bufio"
	"fmt"
	"io"
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/parser"
	"strings"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	buf := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, PROMPT)
		newLine := buf.Scan()
		if !newLine {
			return
		}

		line := buf.Text()
		var a *ast.Database
		var err error
		if len(line) > 8 && strings.ToLower(line[:8]) == "traceon;" {
			a, err = parser.Parse([]byte(line[8:]), parser.WithTraceOn())
		} else {
			a, err = parser.Parse([]byte(line))
		}
		if err != nil {
			fmt.Fprintf(out, "ERROR: %s\n", err)
			continue
		}
		r := string(a.Generate())
		fmt.Fprintf(out, "<< \n%s\n", r)
	}
}
