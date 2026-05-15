package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"dada/check"
	"dada/gen"
	"dada/lex"
	"dada/parse"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	outPath := flag.String("o", "", "output JavaScript file")
	flag.Parse()

	var input io.Reader

	switch flag.NArg() {
	case 0:
		input = os.Stdin

	case 1:
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			return err
		}
		defer f.Close()

		input = f

	default:
		return fmt.Errorf("usage: dada [-o output.js] [file]")
	}

	prog, err := parseInput(input)
	if err != nil {
		return err
	}

	if err := check.Check(prog); err != nil {
		return fmt.Errorf("type error: %s", err)
	}

	var output io.Writer = os.Stdout

	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			return err
		}
		defer f.Close()

		output = f
	}

	return gen.Generate(output, prog)
}

func parseInput(r io.Reader) (prog *parse.Program, err error) {
	defer func() {
		if v := recover(); v != nil {
			switch e := v.(type) {
			case lex.Error:
				err = fmt.Errorf("parse error: %s", e)

			case lex.EOF:
				err = fmt.Errorf("unexpected EOF: %s", e)

			case error:
				err = e

			default:
				panic(v)
			}

			prog = nil
		}
	}()

	p := parse.NewParser(bufio.NewReader(r))
	return p.ParseProgram(), nil
}
