package check

import "dada/parse"

type Error string

func (e Error) Error() string {
	return string(e)
}

type Checker struct {
}

func newChecker() *Checker {
	return &Checker{}
}

func Check(program *parse.Program) error {
	if program == nil {
		return Error("nil program")
	}

	c := newChecker()
	return c.checkProgram(program)
}

func (c *Checker) checkProgram(program *parse.Program) error {
	return nil
}
