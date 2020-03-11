package pifra

import (
	"fmt"
	"strings"
)

type DeclaredProcess struct {
	Process    Element
	Parameters []string
}

// DeclaredProcs is a map of name -> (process, parameters).
var DeclaredProcs map[string]DeclaredProcess

var log = false

// InitProgram parses the byte array and returns the root undeclared process.
func InitProgram(program []byte) (Element, error) {
	initParser()
	boundNameIndex = 0
	lex := newLexer(program)
	if code := yyParse(lex); code != 0 {
		return nil, fmt.Errorf(parseError)
	}
	if len(undeclaredProcs) == 0 {
		return nil, fmt.Errorf("a process must be undeclared to initialise the program")
	}
	if len(undeclaredProcs) > 1 {
		return nil, fmt.Errorf("there cannot be more than one undeclared processes")
	}
	root := InitRootAst(undeclaredProcs[0])
	return root, nil
}

// Log prints debug statements.
func Log(strs ...string) {
	if log {
		fmt.Printf("[DEBUG] %s\n", strings.Join(strs, " "))
	}
}

func initParser() {
	DeclaredProcs = make(map[string]DeclaredProcess)
	undeclaredProcs = []Element{}
}

func popParStack() Element {
	var elem Element
	elem, parStack = parStack[len(parStack)-1], parStack[:len(parStack)-1]
	return elem
}

func popSumStack() Element {
	var elem Element
	elem, sumStack = sumStack[len(sumStack)-1], sumStack[:len(sumStack)-1]
	return elem
}

func pop(stack []int) (int, []int) {
	var val int
	val, stack = stack[len(stack)-1], stack[:len(stack)-1]
	return val, stack
}
