package main

import (
	"fmt"
	"strings"
)

var log = true

// Log prints debug statements.
func Log(strs ...string) {
	if log {
		fmt.Printf("[DEBUG] %s\n", strings.Join(strs, " "))
	}
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

func init() {
	initParser()
}

func initParser() {
	declaredProcs = make(map[string]Element)
	undeclaredProcs = []Element{}
	procParams = make(map[string][]string)
}
